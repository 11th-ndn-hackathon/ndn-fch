package health

import (
	"context"

	"github.com/11th-ndn-hackathon/ndn-fch/model"
)

// ProbeRequest contains a request for health monitoring probe.
type ProbeRequest struct {
	model.TransportIPFamily
	Router string   `json:"router"`
	Names  []string `json:"names"`
}

// ProbeNameResult contains per-name probe result.
type ProbeNameResult struct {
	OK    bool    `json:"ok"`
	RTT   float64 `json:"rtt,omitempty"` // milliseconds
	Error string  `json:"error,omitempty"`
}

// ProbeResponse contains router probe result.
type ProbeResponse struct {
	Connected    bool              `json:"connected"`
	ConnectError string            `json:"connectError,omitempty"`
	Probes       []ProbeNameResult `json:"probes,omitempty"`
}

// Count returns number of probes with OK==true and OK==false.
func (response ProbeResponse) Count() (nSuccess, nFailure int) {
	for _, res := range response.Probes {
		if res.OK {
			nSuccess++
		} else {
			nFailure++
		}
	}
	return
}

// Service represents a service that can probe router health.
type Service interface {
	Probe(ctx context.Context, req ProbeRequest) (res ProbeResponse, e error)
}
