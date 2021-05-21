package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/11th-ndn-hackathon/ndn-fch-control/routerlist"
)

var (
	listenFlag = flag.String("listen", "127.0.0.1:6324", "HTTP listen address")
)

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr, r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.DefaultServeMux.ServeHTTP(w, r)
})

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	log.Fatalln(http.ListenAndServe(*listenFlag, handler))
}

func init() {
	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("User-Agent: *\nDisallow: /\n"))
	})

	http.HandleFunc("/nodes.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		j, _ := json.Marshal(routerlist.List())
		w.Write(j)
	})
}
