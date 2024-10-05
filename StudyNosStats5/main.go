package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {

	var pubkeys map[int]string
	var sorted_pubkeys map[string]int
	var followerlist map[int](map[int]int)
	var followlist map[int](map[int]int)
	//var clustering_coeff map[int]float64
	//var clustering_coeff_cyclic map[int]float64

	pubkeys = make(map[int]string)
	sorted_pubkeys = make(map[string]int)
	followerlist = make(map[int]map[int]int)
	followlist = make(map[int]map[int]int)
	//clustering_coeff = map[int]float64{}
	//clustering_coeff_cyclic = map[int]float64{}

	flog, flerr := os.Create("log.txt")
	if flerr != nil {
		fmt.Println(flerr)
		return
	}
	defer flog.Close()
	log.SetOutput(flog)

	ffollowers, ffollowerserr := os.OpenFile("followers.txt", os.O_RDONLY, 0666)
	followeresscanner := bufio.NewScanner(ffollowers)
	defer ffollowers.Close()
	if ffollowerserr != nil {
		fmt.Println(ffollowerserr)
		log.Println(ffollowerserr)
	}

	for followeresscanner.Scan() {
		line := strings.TrimSpace(followeresscanner.Text())
		items := strings.Split(line, " ")
		//Gather pubkuies have at least one follower
		if len(items) >= 3 {
			i, ierr := strconv.Atoi(items[1])
			if ierr != nil {
				log.Println(ierr)
				return
			}
			sorted_pubkeys[items[0]] = i
			pubkeys[i] = items[0]
			followerlist[i] = make(map[int]int)
			for j := 2; j < len(items); j++ {
				k, kerr := strconv.Atoi(items[j])
				if kerr != nil {
					log.Println(kerr)
					return
				}
				if i != k {
					followerlist[i][k] = k
				}
			}
		}
	}

	ffollows, ffollowserr := os.OpenFile("follows.txt", os.O_RDONLY, 0666)
	followsscanner := bufio.NewScanner(ffollows)
	defer ffollows.Close()
	if ffollowserr != nil {
		fmt.Println(ffollowserr)
		log.Println(ffollowserr)
	}

	for followsscanner.Scan() {
		line := strings.TrimSpace(followsscanner.Text())
		items := strings.Split(line, " ")
		//Gather pubkuies have at least one follow
		if len(items) >= 3 {
			i, ierr := strconv.Atoi(items[1])
			if ierr != nil {
				log.Println(ierr)
				return
			}
			sorted_pubkeys[items[0]] = i
			pubkeys[i] = items[0]
			followlist[i] = make(map[int]int)
			for j := 2; j < len(items); j++ {
				k, kerr := strconv.Atoi(items[j])
				if kerr != nil {
					log.Println(kerr)
					return
				}
				if i != k {
					followlist[i][k] = k
				}
			}
		}
	}

	fresult, fresuterr := os.Create("result.txt")
	if fresuterr != nil {
		log.Println(fresuterr)
		fmt.Println(fresuterr)
	}
	defer fresult.Close()

	/*
		fresult_st, fresut_sterr := os.Create("result_stats.txt")
		if fresuterr != nil {
			log.Println(fresut_sterr)
			fmt.Println(fresut_sterr)
		}
		defer fresult_st.Close()
	*/

	cluster_information := make(map[int][]any)

	for index1, followers := range followerlist {
		cluster_information[index1] = make([]any, 3)
		cluster_information[index1][0] = int(0)
		cluster_information[index1][1] = int(len(followers) * (len(followers) - 1))
		cluster_information[index1][2] = float64(0)

		if cluster_information[index1][1].(int) != 0 {
			for index2 := range followers {
				//fmt.Println(pubkeys[index1], "", index1, " ", index2)
				if index2 != index1 {
					for index3 := range followers {
						if index3 != index2 && index3 != index1 {
							_, ok := followerlist[index2][index3]
							if ok {
								cluster_information[index1][0] = cluster_information[index1][0].(int) + 1
							}
						}
					}
				}
			}
			cluster_information[index1][2] = float64(cluster_information[index1][0].(int)) / float64(cluster_information[index1][1].(int))
		}
	}

	for index, item := range cluster_information {
		fresult.WriteString(fmt.Sprintf("%d\t%s\t%d\t%d\t%d\t%f\n", index, pubkeys[index], len(followerlist[index]), item[0].(int), item[1].(int), item[2].(float64)))
	}
}
