package model

import (
	"cmp"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

// Query represents an API query.
type Query struct {
	Count     int
	Transport TransportType
	IPv4      bool
	IPv6      bool
	Position  LonLat
	Network   string
}

func (q Query) matchTransport(router RouterAvail) bool {
	switch {
	case q.IPv4 && router.Available[TransportIPFamily{q.Transport, IPv4}]:
		return true
	case q.IPv6 && router.Available[TransportIPFamily{q.Transport, IPv6}]:
		return true
	}
	return false
}

func (q Query) matchNetwork(router RouterAvail) bool {
	return q.Network == "" || strings.HasPrefix(router.Prefix(), q.Network)
}

// Execute executes a query.
func (q Query) Execute(avail []RouterAvail) (res []RouterAvail) {
	for _, router := range avail {
		if q.matchTransport(router) && q.matchNetwork(router) {
			res = append(res, router)
		}
	}
	slices.SortFunc(res, func(a, b RouterAvail) int {
		return cmp.Compare(Distance(q.Position, a.Position()), Distance(q.Position, b.Position()))
	})

	if len(res) > q.Count {
		res = res[:q.Count]
	}
	return res
}

// ParseQueries constructs a list of Query from URL query string.
func ParseQueries(qs string) (list []Query) {
	v, _ := url.ParseQuery(qs)

	q := Query{
		Count:     1,
		Transport: TransportUDP,
		IPv4:      v.Get("ipv4") != "0",
		IPv6:      v.Get("ipv6") != "0",
	}
	q.Position[0], _ = strconv.ParseFloat(v.Get("lon"), 64)
	q.Position[1], _ = strconv.ParseFloat(v.Get("lat"), 64)
	if network := strings.Trim(v.Get("network"), "/"); network != "" {
		q.Network = "/" + network + "/"
	}

	counts := []int{}
	for _, n := range v["k"] {
		k, _ := strconv.ParseUint(n, 10, 32)
		counts = append(counts, max(1, int(k)))
	}
	if len(counts) == 0 {
		counts = append(counts, 1)
	}

	for i, tr := range v["cap"] {
		q.Count = counts[i%len(counts)]
		q.Transport = TransportType(tr)
		list = append(list, q)
	}
	if len(list) == 0 {
		list = append(list, q)
	}
	return list
}

// QueryResponse represents an API response.
type QueryResponse struct {
	Updated int64 `json:"updated"` // last update time, milliseconds since epoch

	Routers []QueryResponseRouter `json:"routers"`
}

// QueryResponseRouter is part of QueryResponse.
type QueryResponseRouter struct {
	Transport TransportType `json:"transport"`
	Connect   string        `json:"connect"`
	Prefix    string        `json:"prefix,omitempty"`
}
