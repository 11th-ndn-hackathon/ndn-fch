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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/11th-ndn-hackathon/ndn-fch/logging"
	"github.com/11th-ndn-hackathon/ndn-fch/model"
	"github.com/caitlinelfring/go-env-default"
	"go.uber.org/zap"
)

var (
	testbedLogger    = logging.New("routerlist.testbed")
	testbedNodesURI  = env.GetDefault("FCH_ROUTERLIST_TESTBED_URI", "https://testbed-status.named-data.net/testbed-nodes.json")
	testbedNodesFile = env.GetDefault("FCH_ROUTERLIST_TESTBED_NODES", "./fch-testbed-nodes.json")
	testbedBadList   = func() []string {
		if s := os.Getenv("FCH_ROUTERLIST_TESTBED_BAD"); s != "" {
			return strings.Split(s, ",")
		}
		return []string{}
	}()

	testbedRouters     []model.Router
	testbedRoutersLock sync.RWMutex
)

type testbedRouter struct {
	node      testbedNode
	host      string
	hasIPv4   bool
	hasIPv6   bool
	neighbors map[string]int
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

func (r testbedRouter) ConnectString(tf model.TransportIPFamily) string {
	switch tf.Family {
	case model.IPv4:
		if !r.hasIPv4 {
			return ""
		}
	case model.IPv6:
		if !r.hasIPv6 {
			return ""
		}
	}

	switch tf.Transport {
	case model.TransportUDP:
		return net.JoinHostPort(r.host, model.DefaultUDPPort)
	case model.TransportWebSocket:
		return (&url.URL{
			Scheme: "wss",
			Host:   r.host,
			Path:   "/ws/",
		}).String()
	}
	return ""
}

func (r testbedRouter) Neighbors() map[string]int {
	return r.neighbors
}

type testbedNode struct {
	ShortName    string    `json:"shortname"`
	Site         string    `json:"site"`
	IPAddresses  []string  `json:"ip_addresses"`
	Position     []float64 `json:"position"`
	RealPosition []float64 `json:"_real_position"`
	Prefix       string    `json:"prefix"`
	Neighbors    []string  `json:"neighbors"`
}

func (n testbedNode) Router() (r *testbedRouter) {
	r = &testbedRouter{}

	u, e := url.Parse(n.Site)
	if e != nil {
		return nil
	}
	r.host = u.Hostname()
	if r.host == "0.0.0.0" || slices.Contains(testbedBadList, n.ShortName) {
		return nil
	}

	n.Prefix = strings.TrimPrefix(n.Prefix, "ndn:")

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

	r.neighbors = map[string]int{}
	for _, neighbor := range n.Neighbors {
		if !slices.Contains(testbedBadList, neighbor) {
			r.neighbors[neighbor] = -1
		}
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
