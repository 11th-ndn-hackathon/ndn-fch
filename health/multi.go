package health

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/11th-ndn-hackathon/ndn-fch-control/model"
	"go.uber.org/multierr"
)

// Dispatcher dispatches router health probes to underlying Services based on TransportType.
type Dispatcher map[model.TransportType]Service

var _ Service = Dispatcher{}

// Probe implements Service interface.
func (m Dispatcher) Probe(ctx context.Context, req ProbeRequest) (res ProbeResponse, e error) {
	s := m[req.Transport]
	if s == nil {
		return nil, fmt.Errorf("no service for %s", req.Transport)
	}
	return s.Probe(ctx, req)
}

// NewHTTPDispatcher creates a Multi of HTTPClients from base URI.
func NewHTTPDispatcher(uri string) (m Dispatcher, e error) {
	u, e := url.Parse(uri)
	if e != nil {
		return nil, e
	}

	u0 := *u
	u0.Path = path.Join(u.Path, "health")
	c0, e0 := NewHTTPClient(u0.String())

	u3 := *u
	u3.Path = path.Join(u.Path, "health-http3")
	c3, e3 := NewHTTPClient(u3.String())

	if e := multierr.Combine(e0, e3); e != nil {
		return nil, e
	}

	return Dispatcher{
		model.TransportUDP4:       c0,
		model.TransportUDP6:       c0,
		model.TransportWebSocket4: c0,
		model.TransportWebSocket6: c0,
		model.TransportH3IPv4:     c3,
		model.TransportH3IPv6:     c3,
	}, nil
}
