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

func (r Router) TransportString(tr TransportType, legacySyntax bool) string {
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

// MarshalJSON implements json.Marshaler interface.
func (r RouterAvail) MarshalJSON() (j []byte, e error) {
	s := struct {
		*Router
		Available []string `json:"available"`
	}{r.Router, nil}
	for tf, ok := range r.Available {
		if ok {
			s.Available = append(s.Available, tf.String())
		}
	}
	return json.Marshal(s)
}
