package routerlist

import (
	"net"
	"net/url"
	"strconv"
)

// LonLat is a tuple of longitute,latitude in GeoJSON format.
// https://datatracker.ietf.org/doc/html/rfc7946#section-3.1.1
type LonLat [2]float64

// List returns a list of known routers.
func List() (routers []Router) {
	testbedRoutersLock.RLock()
	defer testbedRoutersLock.RUnlock()
	routers = append(routers, testbedRouters...)
	routers = append(routers, yoursunnyRouters...)
	return routers
}

// Router contains information about a router.
type Router struct {
	ID       string
	Position [2]float64
	Prefix   string

	Host          string
	IPv4          bool
	IPv6          bool
	UDPPort       uint16
	WebSocketPort uint16
	HTTP3Port     uint16
}

func (r Router) UDPHostPort() string {
	if r.UDPPort == 0 {
		return ""
	}
	return net.JoinHostPort(r.Host, strconv.FormatUint(uint64(r.UDPPort), 10))
}

func (r Router) WebSocketHostPort() string {
	if r.WebSocketPort == 0 {
		return ""
	}
	return net.JoinHostPort(r.Host, strconv.FormatUint(uint64(r.WebSocketPort), 10))
}

func (r Router) WebSocketURI() string {
	hostport := r.WebSocketHostPort()
	if hostport == "" {
		return ""
	}
	return (&url.URL{
		Scheme: "wss",
		Host:   hostport,
		Path:   "/ws",
	}).String()
}

func (r Router) HTTP3URI() string {
	if r.HTTP3Port == 0 {
		return ""
	}
	return (&url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(r.Host, strconv.FormatUint(uint64(r.HTTP3Port), 10)),
		Path:   "/ndn",
	}).String()
}
