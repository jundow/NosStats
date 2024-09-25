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
	var followlist map[int](map[int]int)
	var followerlist map[int](map[int]int)

	pubkeys = make(map[int]string)
	sorted_pubkeys = make(map[string]int)
	followlist = make(map[int]map[int]int)
	followerlist = make(map[int]map[int]int)

	var histgram_num_followers [10000]int
	var histgram_num_follows [10000]int

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
		if len(items) >= 2 {
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
		if len(items) >= 2 {
			i, ierr := strconv.Atoi(items[1])
			if ierr != nil {
				log.Println(ierr)
				return
			}
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

	for i := 0; i < len(histgram_num_followers); i++ {
		histgram_num_followers[i] = 0
		histgram_num_follows[i] = 0
	}

	for _, follow := range followlist {
		num := len(follow)
		if num < 10000 {
			histgram_num_follows[num]++
		}
	}
	//fmt.Println(histgram_num_follows)

	for _, follower := range followerlist {
		num := len(follower)
		if num < 10000 {
			histgram_num_followers[num]++
		}
	}
	//fmt.Println(histgram_num_followers)

	ffollowsstats, ffollowsstatserr := os.Create("followstats.txt")
	if ffollowsstatserr != nil {
		fmt.Println(ffollowsstatserr)
		log.Println(ffollowsstatserr)
		return
	}
	defer ffollowsstats.Close()

	for i := 0; i < len(histgram_num_follows); i++ {
		ffollowsstats.WriteString(fmt.Sprintf("%d\t%d\n", i, histgram_num_follows[i]))
	}

	ffollowersstats, ffollowersstatserr := os.Create("followerstats.txt")
	if ffollowersstatserr != nil {
		fmt.Println(ffollowersstatserr)
		log.Println(ffollowersstatserr)
		return
	}
	defer ffollowersstats.Close()

	for i := 0; i < len(histgram_num_follows); i++ {
		ffollowersstats.WriteString(fmt.Sprintf("%d\t%d\n", i, histgram_num_followers[i]))
	}

	fxyplots, fxyplotserr := os.Create("xyplots.txt")
	if fxyplotserr != nil {
		fmt.Println(fxyplots)
		log.Println(fxyplots)
		return
	}
	defer fxyplots.Close()

	for pk, index := range sorted_pubkeys {
		fxyplots.WriteString(fmt.Sprintf("%s\t%d\t%d\n", pk, len(followlist[index]), len(followerlist[index])))
	}

}
