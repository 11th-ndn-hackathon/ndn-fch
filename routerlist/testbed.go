package routerlist

import (
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/11th-ndn-hackathon/ndn-fch/logging"
	"github.com/11th-ndn-hackathon/ndn-fch/model"
	"go.uber.org/zap"
)

const (
	testbedNodesURI = "https://ndndemo.arl.wustl.edu/testbed-nodes.json"
)

var (
	testbedLogger      = logging.New("routerlist.testbed")
	testbedNodesFile   = os.Getenv("FCH_ROUTERLIST_TESTBED_NODES")
	testbedRouters     []model.Router
	testbedRoutersLock sync.RWMutex
)

type testbedRouter struct {
	node    testbedNode
	host    string
	hasIPv4 bool
	hasIPv6 bool
}

var _ model.Router = testbedRouter{}

func (r testbedRouter) ID() string {
	return r.node.ShortName
}

func (r testbedRouter) Position() (pos model.LonLat) {
	pos[1], pos[0] = r.node.Position[0], r.node.Position[1]
	return
}

func (r testbedRouter) Prefix() string {
	return r.node.Prefix
}

func (r testbedRouter) HasIPFamily(family model.IPFamily) bool {
	switch family {
	case model.IPv4:
		return r.hasIPv4
	case model.IPv6:
		return r.hasIPv6
	}
	return false
}

func (r testbedRouter) ConnectString(tr model.TransportType) string {
	switch tr {
	case model.TransportUDP:
		return net.JoinHostPort(r.host, model.DefaultUDPPort)
	case model.TransportWebSocket:
		if !r.node.WsTls {
			return ""
		}
		return (&url.URL{
			Scheme: "wss",
			Host:   r.host,
			Path:   "/ws/",
		}).String()
	}
	return ""
}

func (r testbedRouter) Neighbors() (links map[string]int) {
	links = map[string]int{}
	for _, neighbor := range r.node.Neighbors {
		links[neighbor] = -1
	}
	return links
}

type testbedNode struct {
	ShortName    string    `json:"shortname"`
	Site         string    `json:"site"`
	IPAddresses  []string  `json:"ip_addresses"`
	Position     []float64 `json:"position"`
	RealPosition []float64 `json:"_real_position"`
	Prefix       string    `json:"prefix"`
	NdnUp        bool      `json:"ndn-up"`
	WsTls        bool      `json:"ws-tls"`
	Neighbors    []string  `json:"neighbors"`
}

func (n testbedNode) Router() (r *testbedRouter) {
	r = &testbedRouter{}

	u, e := url.Parse(n.Site)
	if e != nil {
		return nil
	}
	r.host = u.Hostname()
	if r.host == "0.0.0.0" {
		return nil
	}

	if n.NdnUp {
		n.Prefix = strings.TrimPrefix(n.Prefix, "ndn:")
	} else {
		n.Prefix = ""
	}

	switch {
	case len(n.RealPosition) == 2:
		n.Position = n.RealPosition
	case len(n.Position) == 2:
	default:
		return nil
	}

	for _, ipStr := range n.IPAddresses {
		ip, e := netip.ParseAddr(ipStr)
		if e != nil {
			continue
		}
		r.hasIPv4 = r.hasIPv4 || ip.Is4()
		r.hasIPv6 = r.hasIPv6 || ip.Is6()
	}
	if !r.hasIPv4 && !r.hasIPv6 {
		return nil
	}

	r.node = n
	return r
}

func fetchTestbedNodes() (m map[string]testbedNode) {
	response, e := http.Get(testbedNodesURI)
	if e != nil {
		testbedLogger.Warn("fetch HTTP", zap.Error(e))
		return nil
	}
	if response.StatusCode != http.StatusOK {
		testbedLogger.Warn("fetch HTTP", zap.String("status", response.Status))
		return nil
	}

	body, e := io.ReadAll(response.Body)
	if e != nil {
		testbedLogger.Warn("fetch read", zap.Error(e))
		return nil
	}

	if e := json.Unmarshal(body, &m); e != nil {
		testbedLogger.Warn("fetch decode", zap.Error(e))
		return nil
	}
	return m
}

func updateTestbedRouters() {
	time.AfterFunc(time.Duration(600+rand.Intn(60))*time.Second, updateTestbedRouters)

	nodes := fetchTestbedNodes()
	if len(nodes) == 0 {
		if e := loadJSONFile(testbedNodesFile, &nodes); e != nil {
			testbedLogger.Error("load cached", zap.Error(e))
			return
		}
	} else {
		if e := saveJSONFile(testbedNodesFile, nodes); e != nil {
			testbedLogger.Warn("save cached", zap.Error(e))
		}
	}

	routers := []model.Router{}
	for _, n := range nodes {
		r := n.Router()
		if r != nil {
			routers = append(routers, *r)
		}
	}

	testbedRoutersLock.Lock()
	defer testbedRoutersLock.Unlock()
	testbedLogger.Debug("update",
		zap.Int("old-len", len(testbedRouters)),
		zap.Int("new-len", len(routers)),
	)
	testbedRouters = routers
}
