package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

func main() {

	var pubkeys map[int]string
	var sorted_pubkeys map[string]int
	var followerlist map[int](map[int]int)

	pubkeys = make(map[int]string)
	sorted_pubkeys = make(map[string]int)
	followerlist = make(map[int]map[int]int)

	g := simple.NewDirectedGraph()

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

	for index, _ := range followerlist {
		node := simple.Node(index)
		g.AddNode(node)
	}

	for index, item := range followerlist {
		from := simple.Node(index)
		for findex := range item {
			to := simple.Node(findex)
			edge := g.NewEdge(from, to)
			g.SetEdge(edge)
		}
	}

	fresult, fresuterr := os.Create("result.txt")
	if fresuterr != nil {
		log.Println(fresuterr)
		fmt.Println(fresuterr)
	}
	defer fresult.Close()

	fresult_st, fresut_sterr := os.Create("result_stats.txt")
	if fresuterr != nil {
		log.Println(fresut_sterr)
		fmt.Println(fresut_sterr)
	}
	defer fresult_st.Close()
	var distance_distribution [10000]int

	i := 0

	for start, pks := range pubkeys {
		i++
		fmt.Println(i, " ", pks, " ", start)
		shortest := path.DijkstraFrom(simple.Node(start), g)
		for end := range pubkeys {
			if start != end {
				p, w := shortest.To(int64(end))
				if !math.IsInf(w, 0) {
					fresult.WriteString(fmt.Sprintf("%d\t%d\t%d\t", start, end, int(w)))
					for _, node := range p {
						fresult.WriteString(fmt.Sprintf("%d\t", node))
					}
					fresult.WriteString(fmt.Sprintln())
					distance_distribution[int(w)]++
				} else {
					distance_distribution[0]++
				}
			}
		}
	}

	for j := 0; j < 10000; j++ {
		fresult_st.WriteString(fmt.Sprintf("%d\t%d\n", j, distance_distribution[j]))
	}
}
