package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func read_profile(relay string, pubks []string, v *[]any, done chan struct{}) {
	var dat any
	defer close(done)

	var pubk_lists []([]string)

	n := 0
	m := 0
	pubk_lists = append(pubk_lists, []string{})
	for _, pubk := range pubks {
		if n == 100 {
			n = 0
			m++
			pubk_lists = append(pubk_lists, []string{})
		}
		pubk_lists[m] = append(pubk_lists[m], pubk)
		n++
	}

	for _, pubk_list := range pubk_lists {

		array_pubk := ""
		for _, pubk := range pubk_list {
			array_pubk += "\"" + pubk + "\","
		}

		array_pubk = "[" + array_pubk[0:len(array_pubk)-1] + "]"

		u := url.URL{Scheme: "wss", Host: relay, Path: "/"}
		conn, _, werr := websocket.DefaultDialer.Dial(u.String(), nil)
		if werr != nil {
			log.Print(relay + " 1 ")
			log.Println(werr)
			return
		}
		defer conn.Close()

		reqid := fmt.Sprint(time.Now().UnixMilli())

		req_profile := "[\"REQ\", \"" + reqid + "\", {\"kinds\":[3],\"authors\":" + array_pubk + "}]"
		wrerr := conn.WriteMessage(websocket.TextMessage, []byte(req_profile))
		if wrerr != nil {
			log.Println(wrerr)
			return
		}

		for {
			_, message, rerr := conn.ReadMessage()
			if rerr != nil {
				log.Printf("%s 2 ", relay)
				log.Println(rerr)
				return
			}

			jerr := json.Unmarshal(message, &dat)
			if jerr != nil {
				log.Println(relay + ": " + req_profile)
				log.Printf("%s 3 ", relay)
				log.Println(jerr)
				return
			}

			fmt.Println(req_profile)
			fmt.Println(dat)

			if msgtyp := (dat.([]any))[0]; msgtyp == "EOSE" {
				log.Printf(relay + "EOSE : " + req_profile)
				log.Printf(" %s \n", dat)
				break
			} else if msgtyp == "NOTICE" {
				log.Printf(relay + "NOTICE : " + req_profile)
				log.Printf(" %s \n", dat)
				return
			} else if msgtyp == "EVENT" {
				log.Printf(relay + "EVENT : " + req_profile)
				*v = append(*v, dat)
				fmt.Println(dat)
			}
		}

		closeerr := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if closeerr != nil {
			log.Println(" aa  " + fmt.Sprint(closeerr))
		}
	}
}

func main() {
	var pubkeys []string
	var sorted_pubkeys map[string]int
	var followlist map[int](map[int]int)
	var followerlist map[int](map[int]int)

	sorted_pubkeys = make(map[string]int)
	followlist = make(map[int]map[int]int)
	followerlist = make(map[int]map[int]int)

	flog, flerr := os.Create("log.txt")
	if flerr != nil {
		fmt.Println(flerr)
		return
	}
	defer flog.Close()
	log.SetOutput(flog)

	fpubk, fpubkerr := os.OpenFile("pubkeys.txt", os.O_RDONLY, 0666)
	pubkscanner := bufio.NewScanner(fpubk)

	if fpubkerr != nil {
		fmt.Println(fpubkerr)
		log.Println(fpubkerr)
	}

	for pubkscanner.Scan() {
		pubk := strings.TrimSpace(pubkscanner.Text())
		if len(pubk) > 1 {
			pubkeys = append(pubkeys, pubk)
			index := len(pubkeys) - 1
			sorted_pubkeys[pubk] = index
			followlist[index] = make(map[int]int)
			followerlist[index] = make(map[int]int)
		}
	}

	relays := []string{
		"nos.lol",
		/*"relay.nostr.band",*/
		//"relay.snort.social",
		//"nostr.fmt.wiz.biz",
		"nostr-pub.wellorder.net",
		//"nostr.mom",
		//"nostr.oxtr.dev",
		//"nostr.semisol.dev",
		//"relay.damus.io",
		//"relay.nostr.bg",
		//"soloco.nl",
		//"nostr.bitcoiner.social",
		"nostr.einundzwanzig.space",
	}

	var interrupt chan os.Signal
	var profiles [](*[]any)

	interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer close(interrupt)

	for _, relay := range relays {
		fmt.Println("Reading from " + relay)
		done := make(chan struct{})
		var v []any
		profiles = append(profiles, &v)
		read_profile(relay, pubkeys, &v, done)
	}

	/*
		fmt.Println("--- Return to stop ---")
		var s string
		fmt.Scanln(&s)
	*/
	for i := range profiles {
		for j := range *profiles[i] {
			profile := (*profiles[i])[j].([]any)
			pk := fmt.Sprint((profile[2].(map[string]any))["pubkey"])
			index := sorted_pubkeys[pk]
			for _, item := range (profile[2].(map[string]any))["tags"].([]any) {
				if len(item.([]any)) > 1 {
					tag := fmt.Sprint((item.([]any))[0])
					if tag == "p" {
						fpubk := fmt.Sprint((item.([]any))[1])
						findex, ok := sorted_pubkeys[fpubk]
						if ok {
							followlist[index][findex] = findex
							followerlist[findex][index] = index
						}
					}
				}
			}
		}
	}

	ffollows, ffollowserr := os.Create("follows.txt")
	if ffollowserr != nil {
		log.Println(ffollowserr)
		return
	}
	defer ffollows.Close()

	for i, user := range followlist {
		ffollows.WriteString(fmt.Sprintf("%s %d ", pubkeys[i], i))
		for _, j := range user {
			ffollows.WriteString(fmt.Sprintf("%d ", j))
		}
		ffollows.WriteString(fmt.Sprintln())
	}

	ffollowers, ffollowerserr := os.Create("followers.txt")
	if ffollowerserr != nil {
		log.Println(ffollowerserr)
		return
	}
	defer ffollowers.Close()

	for i, user := range followerlist {
		ffollowers.WriteString(fmt.Sprintf("%s %d ", pubkeys[i], i))
		for _, j := range user {
			ffollowers.WriteString(fmt.Sprintf("%d ", j))
		}
		ffollowers.WriteString(fmt.Sprintln())
	}
}
