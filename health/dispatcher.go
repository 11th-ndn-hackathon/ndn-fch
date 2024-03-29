package health

import (
	"context"
	"errors"
	"fmt"

	"github.com/11th-ndn-hackathon/ndn-fch/model"
)

// Dispatcher dispatches router health probes to underlying Services based on TransportType.
type Dispatcher map[model.TransportType]Service

var _ Service = Dispatcher{}

// Probe implements Service interface.
func (m Dispatcher) Probe(ctx context.Context, req ProbeRequest) (res ProbeResponse, e error) {
	s := m[req.Transport]
	if s == nil {
		return res, fmt.Errorf("no service for %s", req.Transport)
	}
	return s.Probe(ctx, req)
}

// NewHTTPDispatcher creates a Dispatcher of HTTPClients from base URI.
//
//	uri: base URI for UDP and WebSockets probe.
//	uri3: base URI for HTTP/3 probe.
func NewHTTPDispatcher(uri, uri3 string) (m Dispatcher, e error) {
	c0, e0 := NewHTTPClient(uri)
	c3, e3 := NewHTTPClient(uri3)
	if e := errors.Join(e0, e3); e != nil {
		return nil, e
	}

	return Dispatcher{
		model.TransportUDP:       c0,
		model.TransportWebSocket: c0,
		model.TransportH3:        c3,
	}, nil
}
