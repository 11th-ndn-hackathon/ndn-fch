package model

import "fmt"

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

// TransportTypes is a list of known TransportType values.
var TransportTypes = []TransportType{
	TransportUDP4,
	TransportUDP6,
	TransportWebSocket4,
	TransportWebSocket6,
	TransportH3IPv4,
	TransportH3IPv6,
}

// RouterString returns a connection string for the router using the given transport.
// For UDP this is host:port. For WebSockets or HTTP/3 this is URI.
// Returns empty string if router does not support this transport type.
func (transport TransportType) RouterString(router Router) string {
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

// UpdaterObject returns an object for updating web-api component.
func (transport TransportType) UpdaterObject(router Router) (o UpdaterObject) {
	o = UpdaterObject{
		"id":       fmt.Sprintf("%s:%s", router.ID, transport),
		"position": router.Position,
	}

	switch transport {
	case TransportUDP4, TransportWebSocket4, TransportH3IPv4:
		o["ipv4"] = true
	case TransportUDP6, TransportWebSocket6, TransportH3IPv6:
		o["ipv6"] = true
	}

	switch transport {
	case TransportUDP4, TransportUDP6:
		if router.UDPPort == 6363 {
			o["udp"] = router.Host
		} else {
			o["udp"] = router.UDPHostPort()
		}
	case TransportWebSocket4, TransportWebSocket6:
		if router.WebSocketPort == 443 {
			o["wss"] = router.Host
		} else {
			o["wss"] = router.WebSocketHostPort()
		}
	case TransportH3IPv4, TransportH3IPv6:
		o["http3"] = router.HTTP3URI()
	}

	return o
}

// UpdaterObject is a JSON object sent to API service.
type UpdaterObject map[string]interface{}
