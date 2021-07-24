package model

import (
	"net"
	"net/url"
)

// Default ports.
const (
	DefaultUDPPort       = "6363"
	DefaultWebSocketPort = "443"
)

// IPFamily indicates an IP address family.
type IPFamily int

const (
	IPv4 IPFamily = 4
	IPv6 IPFamily = 6
)

// IPFamilies is a list of known IPFamily values.
var IPFamilies = []IPFamily{IPv4, IPv6}

// TransportType represents a type of transport.
type TransportType string

const (
	TransportUDP       TransportType = "udp"
	TransportWebSocket TransportType = "wss"
	TransportH3        TransportType = "http3"
)

// TransportTypes is a list of known TransportType values.
var TransportTypes = []TransportType{
	TransportUDP,
	TransportWebSocket,
	TransportH3,
}

// MakeLegacyConnectString converts a connect string to legacy syntax.
func MakeLegacyConnectString(tr TransportType, connect string) string {
	switch tr {
	case TransportUDP:
		host, port, _ := net.SplitHostPort(connect)
		if port == DefaultUDPPort {
			return host
		}
	case TransportWebSocket:
		u, _ := url.Parse(connect)
		if u.Scheme == "wss" && u.Port() == "" && u.Path == "/ws/" {
			return u.Hostname()
		}
	}
	return connect
}

// TransportIPFamily is a combination of TransportType and IPFamily.
type TransportIPFamily struct {
	Transport TransportType `json:"transport"`
	Family    IPFamily      `json:"family"`
}

// TransportTypes is a list of known TransportIPFamily values.
var TransportIPFamilies = func() (list []TransportIPFamily) {
	for _, tr := range TransportTypes {
		for _, af := range IPFamilies {
			list = append(list, TransportIPFamily{tr, af})
		}
	}
	return list
}()
