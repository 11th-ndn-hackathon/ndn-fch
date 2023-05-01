package main

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"time"

	"github.com/11th-ndn-hackathon/ndn-fch/availlist"
	"github.com/11th-ndn-hackathon/ndn-fch/model"
	"github.com/elnormous/contenttype"
)

func init() {
	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User-Agent: *\nDisallow: /\n"))
	})

	http.HandleFunc("/routers.json", func(w http.ResponseWriter, r *http.Request) {
		list, updated := availlist.List()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Last-Modified", updated.Format(http.TimeFormat))
		j, _ := json.Marshal(list)
		w.Write(j)
	})

	http.HandleFunc("/", handleQuery)
}

const (
	mimeText = "text/plain"
	mimeJSON = "application/json"
)

var (
	queryAccepts = []contenttype.MediaType{
		contenttype.NewMediaType(mimeText),
		contenttype.NewMediaType(mimeJSON),
	}
)

func handleQuery(w http.ResponseWriter, r *http.Request) {
	avail, updated := availlist.List()
	if len(avail) == 0 {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	queries := model.ParseQueries(r.URL.RawQuery)
	response := model.QueryResponse{
		Updated: updated.UnixNano() / int64(time.Millisecond),
		Routers: []model.QueryResponseRouter{},
	}

	contentType := mimeText
	if accept, _, e := contenttype.GetAcceptableMediaType(r, queryAccepts); e == nil {
		contentType = accept.String()
	}

	preferLegacySyntax := contentType != mimeJSON
	for _, q := range queries {
		for _, r := range q.Execute(avail) {
			connect := r.ConnectString(model.TransportIPFamily{Transport: q.Transport, Family: 4})
			if connect == "" {
				connect = r.ConnectString(model.TransportIPFamily{Transport: q.Transport, Family: 6})
			}
			if connect == "" {
				continue
			}
			if preferLegacySyntax {
				connect = model.MakeLegacyConnectString(q.Transport, connect)
			}
			response.Routers = append(response.Routers, model.QueryResponseRouter{
				Transport: q.Transport,
				Connect:   connect,
				Prefix:    r.Prefix(),
			})
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Last-Modified", updated.Format(http.TimeFormat))
	switch contentType {
	case mimeJSON:
		j, _ := json.Marshal(response)
		w.Write(j)
	case mimeText:
		connectStrings := make([]string, 0, len(response.Routers))
		for _, router := range response.Routers {
			connectStrings = append(connectStrings, router.Connect)
		}
		cw := csv.NewWriter(w)
		cw.Write(connectStrings)
		cw.Flush()
	}
}
