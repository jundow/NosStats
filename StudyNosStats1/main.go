package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"sort"
	"time"

	"github.com/gorilla/websocket"
)

type event_struct struct {
	Created_at string `json:"created_at"`
	Pubkey     string `json:"pubkey"`
	Kind       string `json:"kind"`
	Content    string `json:"content"`
}

func read_events(relay string, reqstring string, conn *websocket.Conn, v *[]any, done chan struct{}) {
	var dat any
	defer close(done)

	wrerr := conn.WriteMessage(websocket.TextMessage, []byte(reqstring))

	if wrerr != nil {
		log.Println(wrerr)
		return
	}

	for {
		_, message, rerr := conn.ReadMessage()
		if rerr != nil {
			log.Printf("%s ", relay)
			log.Println(rerr)
			return
		}
		jerr := json.Unmarshal(message, &dat)
		if jerr != nil {
			log.Printf("%s ", relay)
			log.Println(jerr)
			return
		}

		if msgtyp := (dat.([]any))[0]; msgtyp == "EOSE" {
			log.Printf("%s ", relay)
			log.Printf("EOSE: %s\n", dat)
			/*
				wrerr = conn.WriteMessage(websocket.TextMessage, []byte(reqstring))
				if wrerr != nil {
					log.Println(wrerr)
					return
				}
			*/
		} else if msgtyp == "NOTICE" {
			log.Printf("%s ", relay)
			log.Printf("NOTICE: %s\n", dat)
			return
		} else {
			*v = append(*v, dat)
		}
	}
}

func read_profile(relay string, conn *websocket.Conn, v *[]any, done chan struct{}) {
	var dat any
	defer close(done)

	for {
		_, message, rerr := conn.ReadMessage()
		if rerr != nil {
			log.Printf("%s ", relay)
			log.Println(rerr)
			return
		}

		jerr := json.Unmarshal(message, &dat)
		if jerr != nil {
			log.Printf("%s ", relay)
			log.Println(jerr)
			return
		}

		if msgtyp := (dat.([]any))[0]; msgtyp == "EOSE" {
		} else if msgtyp == "NOTICE" {
			log.Printf("%s ", relay)
			log.Printf("notice: %s\n", dat)
		} else {
			*v = append(*v, dat)
		}
	}
}

func main() {

	flog, flerr := os.Create("log.txt")
	if flerr != nil {
		fmt.Println(flerr)
		return
	}
	defer flog.Close()

	log.SetOutput(flog)

	relays := []string{"nos.lol",
		/*"relay.nostr.band",*/
		"relay.snort.social",
		"nostr.fmt.wiz.biz",
		"nostr-pub.wellorder.net",
		"nostr.mom",
		"nostr.oxtr.dev",
		"nostr.semisol.dev",
		"relay.damus.io",
		"relay.nostr.bg",
		"soloco.nl",
		"nostr.bitcoiner.social",
		"nostr.einundzwanzig.space"}

	var conns []*websocket.Conn
	var interrupt chan os.Signal
	var done_chs []chan struct{}
	var notes [](*[]any)

	interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer close(interrupt)

	reqid := fmt.Sprint(time.Now().UnixMilli())

	date1, errd1 := time.Parse("20060102", "20240801")
	date2, errd2 := time.Parse("20060102", "20240831")
	fmt.Println(date1)
	fmt.Println(date2)

	if errd1 != nil || errd2 != nil {
		fmt.Println(errd1)
		fmt.Println(errd2)
		return
	}

	req := "[\"REQ\", \"" + fmt.Sprint(reqid) + "\", {\"kinds\":[1]," +
		"\"since\":" + fmt.Sprint(date1.Unix()) + "," +
		"\"until\":" + fmt.Sprint(date2.Unix()) + "," +
		"\"limit\":1000000}]"

	//req := "[\"REQ\", \"" + fmt.Sprint(reqid) + "\", {\"kinds\":[1]," + "\"limit\":1000000}]"
	fmt.Println(req)

	for _, relay := range relays {
		u := url.URL{Scheme: "wss", Host: relay, Path: "/"}
		conn, _, werr := websocket.DefaultDialer.Dial(u.String(), nil)
		if werr != nil {
			log.Print(relay + " ")
			log.Println(werr)
			return
		}
		defer conn.Close()
		conns = append(conns, conn)

		done := make(chan struct{})
		done_chs = append(done_chs, done)

		var v []any
		notes = append(notes, &v)
		go read_events(relay, req, conn, &v, done)

		/*
			wrerr := conn.WriteMessage(websocket.TextMessage, []byte(req))
			if wrerr != nil {
				log.Println(wrerr)
				return
			}
		*/
	}

	fmt.Println("--- Return to stop ---")
	var s string
	fmt.Scanln(&s)

	for i := range notes {
		fp, ferr := os.Create(relays[i] + ".txt")
		if ferr != nil {
			log.Println(ferr)
			return
		}
		for _, note := range *notes[i] {
			//utime := int64((note.([]any))[2].(map[string]any)["created_at"].(float64))
			//ti := time.Unix(utime, 0)
			//fp.WriteString(fmt.Sprint(ti) + " ")
			fp.WriteString(fmt.Sprintln(note))
		}
		defer fp.Close()
	}

	var pubkeies []string
	//var users []user

	for i := range notes {
		for j := range *notes[i] {
			note := (*notes[i])[j].([]any)
			if note[0] == "EVENT" {
				data := note[2].(map[string]any)
				if data["pubkey"] != nil {
					if !slices.Contains(pubkeies, data["pubkey"].(string)) {
						pubkeies = append(pubkeies, data["pubkey"].(string))
					}
				}
			}
		}
	}

	sort.Slice(pubkeies, func(i, j int) bool {
		return pubkeies[i] < pubkeies[j]
	})

	fpp, fperr := os.Create("pubkeys.txt")
	if fperr != nil {
		log.Println(fperr)
		return
	}

	for _, pubk := range pubkeies {
		_, werr := fpp.WriteString(pubk + "\n")
		if werr != nil {
			log.Println(werr)
			return
		}
	}
	defer fpp.Close()

	for i := range relays {
		closeerr := conns[i].WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if closeerr != nil {
			log.Println(closeerr)
			return
		}
	}
}
