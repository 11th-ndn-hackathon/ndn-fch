package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/11th-ndn-hackathon/ndn-fch/health"
	"github.com/11th-ndn-hackathon/ndn-fch/model"
	"github.com/11th-ndn-hackathon/ndn-fch/routerlist"
	"go.uber.org/zap"
)

var (
	refreshInterval time.Duration
	probe           health.Service
	nProbes         int
	apiUpdateUri    string
)

func probeRouters(ctx context.Context) (avail []model.UpdaterObject) {
	routers := routerlist.List()
	collect := make(chan model.UpdaterObject)
	go func() {
		for o := range collect {
			avail = append(avail, o)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(len(model.TransportTypes))
	for _, transport := range model.TransportTypes {
		go func(transport model.TransportType) {
			request := health.ProbeRequest{
				Transport: transport,
				Routers:   routers,
			}

			responses := []health.ProbeResponse{}
			for i := 0; i < nProbes; i++ {
				logEntry := logger.With(
					zap.String("transport", string(transport)),
					zap.Int("i", i),
				)

				request = request.RandomPingServer()
				response, e := probe.Probe(ctx, request)
				if e != nil {
					logEntry.Warn("probe error", zap.Error(e))
					continue
				}

				if ce := logEntry.Check(zap.DebugLevel, "probe response"); ce != nil {
					nSuccess, nFailure := response.Count()
					ce.Write(zap.Int("success-count", nSuccess), zap.Int("failure-count", nFailure))
				}
				responses = append(responses, response)
			}

			best := health.MergeProbeResponse(responses...)
			for _, router := range routers {
				res, ok := best[router.ID]
				if !ok || !res.OK {
					continue
				}
				collect <- transport.UpdaterObject(router)
			}
			wg.Done()
		}(transport)
	}
	wg.Wait()
	close(collect)

	return avail
}

func refresh(ctx context.Context) error {
	avail := probeRouters(ctx)
	if len(avail) == 0 {
		logger.Info("probe result empty, not updating")
		return nil
	}

	if ce := logger.Check(zap.InfoLevel, "probe result"); ce != nil {
		ce.Write(zap.Reflect("avail", avail))
	}

	j, _ := json.Marshal(avail)
	hReq, e := http.NewRequestWithContext(ctx, http.MethodPut, apiUpdateUri, bytes.NewReader(j))
	hReq.Header.Set("content-type", "application/json")
	if e != nil {
		return e
	}

	hRes, e := http.DefaultClient.Do(hReq)
	if e != nil {
		return e
	}
	if hRes.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", hRes.StatusCode)
	}
	return nil
}
