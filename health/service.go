package health

import (
	"context"
	"math"
	"math/rand"
	"path"

	"github.com/11th-ndn-hackathon/ndn-fch-control/model"
)

// ProbeRequest contains a request for health monitoring probe.
type ProbeRequest struct {
	Transport model.TransportType
	Routers   []model.Router
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

// Count returns number of routers with OK==true and OK==false.
func (response ProbeResponse) Count() (nSuccess, nFailure int) {
	for _, res := range response {
		if res.OK {
			nSuccess++
		} else {
			nFailure++
		}
	}
	return
}

// MergeProbeResponse merges ProbeResponse to the best result for each router.
func MergeProbeResponse(responses ...ProbeResponse) (best ProbeResponse) {
	best = make(ProbeResponse)
	for _, m := range responses {
		for id, res := range m {
			prev, ok := best[id]
			if ok {
				better := ProbeRouterResult{
					OK: prev.OK || res.OK,
				}
				if better.OK {
					better.RTT = math.Min(prev.RTT, res.RTT)
				} else {
					better.Error = prev.Error
				}
				best[id] = better
			} else {
				best[id] = res
			}
		}
	}
	return best
}

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
