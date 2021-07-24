package routerlist

import (
	"os"

	"github.com/11th-ndn-hackathon/ndn-fch/logging"
	"github.com/11th-ndn-hackathon/ndn-fch/model"
	"go.uber.org/zap"
)

var (
	ndn6Logger       = logging.New("routerlist.ndn6")
	ndn6TopoFilename = os.Getenv("FCH_ROUTERLIST_NDN6_TOPO")
	ndn6Routers      []model.Router
)

type ndn6Topo struct {
	Nodes map[string]*ndn6Node `json:"nodes"`
	Links []ndn6Link           `json:"links"`
}

type ndn6Node struct {
	id       string
	allLinks map[string]int

	PositionV model.LonLat `json:"position"`

	Public struct {
		IPv4 bool   `json:"ipv4,omitempty"`
		IPv6 bool   `json:"ipv6,omitempty"`
		WSS  string `json:"wss,omitempty"`
		H3   string `json:"h3,omitempty"`
	} `json:"public"`

	Links []struct {
		ID   string `json:"remote_id"`
		Cost int    `json:"cost"`
	} `json:"links"`
}

var _ model.Router = ndn6Node{}

func (n ndn6Node) ID() string {
	return n.id
}

func (n ndn6Node) Position() (pos model.LonLat) {
	return n.PositionV
}

func (r ndn6Node) Prefix() string {
	return "/yoursunny/_/" + r.id
}

func (r ndn6Node) HasIPFamily(family model.IPFamily) bool {
	switch family {
	case model.IPv4:
		return r.Public.IPv4
	case model.IPv6:
		return r.Public.IPv6
	}
	return false
}

func (r ndn6Node) ConnectString(tr model.TransportType) string {
	switch tr {
	case model.TransportWebSocket:
		return r.Public.WSS
	case model.TransportH3:
		return r.Public.H3
	}
	return ""
}

func (r ndn6Node) Neighbors() (links map[string]int) {
	return r.allLinks
}

type ndn6Link struct {
	Src  string `json:"src"`
	Dst  string `json:"dst"`
	Cost int    `json:"cost"`
}

func loadNDN6Topo() {
	var topo ndn6Topo
	if e := loadJSONFile(ndn6TopoFilename, &topo); e != nil {
		ndn6Logger.Error("load error", zap.Error(e))
		return
	}

	for id, node := range topo.Nodes {
		node.id = id
		node.allLinks = map[string]int{}
		for _, link := range node.Links {
			node.allLinks[link.ID] = link.Cost
		}
	}
	for _, link := range topo.Links {
		nodeA, nodeB := topo.Nodes[link.Src], topo.Nodes[link.Dst]
		if nodeA == nil || nodeB == nil {
			continue
		}
		nodeA.allLinks[nodeB.id] = link.Cost
		nodeB.allLinks[nodeA.id] = link.Cost
	}

	ndn6Routers = nil
	for _, node := range topo.Nodes {
		ndn6Routers = append(ndn6Routers, node)
	}

	ndn6Logger.Info("load success", zap.Int("count", len(ndn6Routers)))
}
