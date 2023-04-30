package model

import (
	"encoding/json"
)

// Router provides information about a router.
type Router interface {
	ID() string
	Position() LonLat

	// Prefix returns ping server prefix, excluding "/ping" suffix.
	// Return empty string if pingserver is unavailable.
	Prefix() string

	// ConnectString returns a connection string for the given transport.
	//  - UDP: host:port.
	//  - WebSocket: URI.
	//  - HTTP3: URI.
	// Return empty string if transport or IP family is not supported.
	ConnectString(tf TransportIPFamily) string

	// Neighbor returns a map of neighbor ID and link cost.
	Neighbors() map[string]int
}

// RouterAvail contains router availability information.
type RouterAvail struct {
	Router
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
		ID        string              `json:"id"`
		Position  LonLat              `json:"position"`
		Prefix    string              `json:"prefix,omitempty"`
		Neighbors map[string]int      `json:"neighbors"`
		Available []TransportIPFamily `json:"available"`
	}{
		ID:        r.Router.ID(),
		Position:  r.Router.Position(),
		Prefix:    r.Router.Prefix(),
		Neighbors: r.Router.Neighbors(),
		Available: []TransportIPFamily{},
	}
	for tf, ok := range r.Available {
		if ok {
			s.Available = append(s.Available, tf)
		}
	}
	return json.Marshal(s)
}
