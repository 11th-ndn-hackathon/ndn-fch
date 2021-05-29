package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/11th-ndn-hackathon/ndn-fch/availlist"
	"github.com/11th-ndn-hackathon/ndn-fch/model"
	"github.com/11th-ndn-hackathon/ndn-fch/routerlist"
	"github.com/elnormous/contenttype"
)

func init() {
	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User-Agent: *\nDisallow: /\n"))
	})

	http.HandleFunc("/routerlist.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		j, _ := json.Marshal(routerlist.List())
		w.Write(j)
	})

	http.HandleFunc("/availlist.json", func(w http.ResponseWriter, r *http.Request) {
		list, updated := availlist.List()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Last-Modified", updated.Format(http.TimeFormat))
		j, _ := json.Marshal(list)
		w.Write(j)
	})

	http.HandleFunc("/", handleQuery)
}

const (
	mimeJSON = "application/json"
	mimeText = "text/plain"
)

var (
	queryAccepts = []contenttype.MediaType{
		contenttype.NewMediaType(mimeText),
		contenttype.NewMediaType(mimeJSON),
	}
)

type queryResponse struct {
	Updated int64                            `json:"updated"`
	Routers map[model.TransportType][]string `json:"routers"`
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	avail, updated := availlist.List()
	queries := model.ParseQueries(r.URL.RawQuery)
	response := queryResponse{
		Updated: updated.UnixNano() / int64(time.Millisecond),
		Routers: map[model.TransportType][]string{},
	}

	contentType := mimeText
	if accept, _, e := contenttype.GetAcceptableMediaType(r, queryAccepts); e == nil {
		contentType = accept.String()
	}

	preferLegacySyntax := contentType == mimeText
	for _, q := range queries {
		res := q.Execute(avail)
		routers := []string{}
		for _, r := range res {
			routers = append(routers, r.TransportString(q.Transport, preferLegacySyntax))
		}
		response.Routers[q.Transport] = routers
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Last-Modified", updated.Format(http.TimeFormat))
	switch contentType {
	case mimeJSON:
		j, _ := json.Marshal(response)
		w.Write(j)
	case mimeText:
		delim := []byte{}
		for _, routers := range response.Routers {
			for _, router := range routers {
				w.Write(delim)
				delim = []byte(",")
				w.Write([]byte(router))
			}
		}
	}
}
