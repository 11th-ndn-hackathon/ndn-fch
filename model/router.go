package model

import (
	"encoding/json"
	"net"
	"net/url"
	"strconv"
)

// Router contains information about a router.
type Router struct {
	ID       string `json:"id"`
	Position LonLat `json:"position"`
	Prefix   string `json:"prefix"`

	Host          string `json:"host"`
	IPv4          bool   `json:"ipv4"`
	IPv6          bool   `json:"ipv6"`
	UDPPort       uint16 `json:"udp-port,omitempty"`
	WebSocketPort uint16 `json:"wss-port,omitempty"`
	HTTP3Port     uint16 `json:"http3-port,omitempty"`
}

// HasIPFamily determines whether router supports given IPFamily.
func (r Router) HasIPFamily(family IPFamily) bool {
	switch family {
	case IPv4:
		return r.IPv4
	case IPv6:
		return r.IPv6
	}
	return false
}

// ConnectString returns a connection string for the given transport.
//  - UDP: host:port; if legacySyntax is true and port is default, host.
//  - WebSocket: URI; if legacySyntax is true and port is default, host.
//  - HTTP3: URI.
// Return empty string if transport is not supported.
func (r Router) ConnectString(tr TransportType, legacySyntax bool) string {
	switch tr {
	case TransportUDP:
		if r.UDPPort == 0 {
			return ""
		}
		if r.UDPPort == DefaultUDPPort && legacySyntax {
			return r.Host
		}
		return net.JoinHostPort(r.Host, strconv.FormatUint(uint64(r.UDPPort), 10))
	case TransportWebSocket:
		if r.WebSocketPort == 0 {
			return ""
		}
		if r.WebSocketPort == DefaultWebSocketPort && legacySyntax {
			return r.Host
		}
		return (&url.URL{
			Scheme: "wss",
			Host:   net.JoinHostPort(r.Host, strconv.FormatUint(uint64(r.WebSocketPort), 10)),
			Path:   "/ws/",
		}).String()
	case TransportH3:
		if r.HTTP3Port == 0 {
			return ""
		}
		return (&url.URL{
			Scheme: "https",
			Host:   net.JoinHostPort(r.Host, strconv.FormatUint(uint64(r.HTTP3Port), 10)),
			Path:   "/ndn",
		}).String()
	}
	return ""
}

// RouterAvail contains router availability information.
type RouterAvail struct {
	*Router
	Available map[TransportIPFamily]bool
}

// CountAvail returns number of available TransportIPFamily combinations.
func (r RouterAvail) CountAvail() (n int) {
	for _, ok := range r.Available {
		if ok {
			n++
		}
	}
	return n
}

// MarshalJSON implements json.Marshaler interface.
func (r RouterAvail) MarshalJSON() (j []byte, e error) {
	s := struct {
		*Router
		Available []TransportIPFamily `json:"available"`
	}{r.Router, nil}
	for tf, ok := range r.Available {
		if ok {
			s.Available = append(s.Available, tf)
		}
	}
	return json.Marshal(s)
}
