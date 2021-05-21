package health

import (
	"context"
	"math/rand"
	"path"

	"github.com/11th-ndn-hackathon/ndn-fch-control/routerlist"
)

// TransportType represents a type of transport.
type TransportType string

const (
	TransportUDP4       TransportType = "udp4"
	TransportUDP6       TransportType = "udp6"
	TransportWebSocket4 TransportType = "wss-ipv4"
	TransportWebSocket6 TransportType = "wss-ipv6"
	TransportH3IPv4     TransportType = "http3-ipv4"
	TransportH3IPv6     TransportType = "http3-ipv6"
)

// RouterString returns a connection string for the router using the given transport.
// For UDP this is host:port. For WebSockets or HTTP/3 this is URI.
// Returns empty string if router does not support this transport type.
func (transport TransportType) RouterString(router routerlist.Router) string {
	switch transport {
	case TransportUDP4, TransportWebSocket4, TransportH3IPv4:
		if !router.IPv4 {
			return ""
		}
	case TransportUDP6, TransportWebSocket6, TransportH3IPv6:
		if !router.IPv6 {
			return ""
		}
	}

	switch transport {
	case TransportUDP4, TransportUDP6:
		return router.UDPHostPort()
	case TransportWebSocket4, TransportWebSocket6:
		return router.WebSocketURI()
	case TransportH3IPv4, TransportH3IPv6:
		return router.HTTP3URI()
	}

	return ""
}

// ProbeRequest contains a request for health monitoring probe.
type ProbeRequest struct {
	Transport TransportType
	Routers   []routerlist.Router
	Name      string
	Suffix    bool
}

// RandomPingServer randomly selects a ping server.
func (req ProbeRequest) RandomPingServer() ProbeRequest {
	i := rand.Intn(len(req.Routers))
	req.Name = path.Join(req.Routers[i].Prefix, "ping")
	req.Suffix = true
	return req
}

// ProbeResponse is a map from router.ID to router probe result.
type ProbeResponse map[string]ProbeRouterResult

// ProbeRouterResult contains per-router probe result.
type ProbeRouterResult struct {
	OK    bool    `json:"ok"`
	RTT   float64 `json:"rtt,omitempty"` // milliseconds
	Error string  `json:"error,omitempty"`
}

// Service represents a service that can probe router health.
type Service interface {
	Probe(ctx context.Context, req ProbeRequest) (res ProbeResponse, e error)
}
