package health

import (
	"context"
	"math"
	"math/rand"
	"path"

	"github.com/11th-ndn-hackathon/ndn-fch/model"
)

// ProbeRequest contains a request for health monitoring probe.
type ProbeRequest struct {
	Transport model.TransportType
	IPFamily  model.IPFamily
	Routers   []model.Router
	Name      string
	Suffix    bool
}

// RandomPingServer randomly selects a ping server.
func (req ProbeRequest) RandomPingServer() ProbeRequest {
	req.Name = ""
	for attempt := 0; req.Name == "" || attempt < 3; attempt++ {
		i := rand.Intn(len(req.Routers))
		req.Name = req.Routers[i].Prefix
	}
	if req.Name == "" { // no pingserver available
		req.Name = "/localhop"
	}
	req.Name = path.Join(req.Name, "ping")
	req.Suffix = true
	return req
}

// ProbeRouterResult contains per-router probe result.
type ProbeRouterResult struct {
	OK    bool    `json:"ok"`
	RTT   float64 `json:"rtt,omitempty"` // milliseconds
	Error string  `json:"error,omitempty"`
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
			prev, hasPrev := best[id]
			switch {
			case !hasPrev:
				best[id] = res
			case prev.OK && res.OK:
				best[id] = ProbeRouterResult{
					OK:  true,
					RTT: math.Min(prev.RTT, res.RTT),
				}
			case prev.OK:
				best[id] = prev
			default:
				best[id] = res
			}
		}
	}
	return best
}

// Service represents a service that can probe router health.
type Service interface {
	Probe(ctx context.Context, req ProbeRequest) (res ProbeResponse, e error)
}
