package routerlist

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"inet.af/netaddr"
)

const (
	testbedNodesURI = "https://ndndemo.arl.wustl.edu/testbed-nodes.json"
)

var (
	testbedRouters     []Router
	testbedRoutersLock sync.RWMutex

	testbedRoutersFile = flag.String("testbed-routers-file", "/tmp/testbed-routers.json", "testbed routers list cache file")
)

type testbedNode struct {
	ShortName    string    `json:"shortname"`
	Site         string    `json:"site"`
	IPAddresses  []string  `json:"ip_addresses"`
	Position     []float64 `json:"position"`
	RealPosition []float64 `json:"_real_position"`
	Prefix       string    `json:"prefix"`
	FchEnabled   bool      `json:"fch-enabled"`
}

func (n testbedNode) Router() (r *Router) {
	if !n.FchEnabled {
		return nil
	}

	r = &Router{
		ID:            n.ShortName,
		Prefix:        strings.TrimPrefix(n.Prefix, "ndn:"),
		UDPPort:       6363,
		WebSocketPort: 443,
	}

	u, e := url.Parse(n.Site)
	if e != nil {
		return nil
	}
	r.Host = u.Hostname()

	switch {
	case len(n.RealPosition) == 2:
		copy(r.Position[:], n.RealPosition)
	case len(n.Position) == 2:
		copy(r.Position[:], n.Position)
	default:
		return nil
	}

	for _, ipStr := range n.IPAddresses {
		ip, e := netaddr.ParseIP(ipStr)
		if e != nil {
			continue
		}
		r.IPv4 = r.IPv4 || ip.Is4()
		r.IPv6 = r.IPv6 || ip.Is6()
	}
	if !r.IPv4 && !r.IPv6 {
		return nil
	}

	return r
}

func fetchTestbedRouters() (routers []Router, e error) {
	response, e := http.Get(testbedNodesURI)
	if e != nil {
		return nil, e
	}
	body, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return nil, e
	}

	var m map[string]testbedNode
	if e := json.Unmarshal(body, &m); e != nil {
		return nil, e
	}

	for _, n := range m {
		r := n.Router()
		if r != nil {
			routers = append(routers, *r)
		}
	}
	return routers, nil
}

func updateTestbedRouters() {
	time.AfterFunc(time.Duration(600+rand.Intn(60))*time.Second, updateTestbedRouters)

	routers, e := fetchTestbedRouters()
	if e != nil {
		log.Println("fetchTestbedRouters", e)
		return
	}
	if e := saveTestbedRouters(routers); e != nil {
		log.Println("saveTestbedRouters", e)
	}

	testbedRoutersLock.Lock()
	defer testbedRoutersLock.Unlock()
	testbedRouters = routers
}

func loadTestbedRouters() error {
	file, e := os.Open(*testbedRoutersFile)
	if e != nil {
		return e
	}
	defer file.Close()

	body, e := io.ReadAll(file)
	if e != nil {
		return e
	}

	return json.Unmarshal(body, &testbedRouters)
}

func saveTestbedRouters(routers []Router) error {
	j, e := json.Marshal(routers)
	if e != nil {
		return e
	}

	tmpFile, e := os.CreateTemp("", "")
	if e != nil {
		return e
	}
	tmpName := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(tmpName)
	}()

	if _, e := tmpFile.Write(j); e != nil {
		return e
	}

	tmpFile.Close()
	return os.Rename(tmpName, *testbedRoutersFile)
}

func init() {
	if e := loadTestbedRouters(); e != nil {
		log.Println("loadTestbedRouters", e)
	}

	go updateTestbedRouters()
}
