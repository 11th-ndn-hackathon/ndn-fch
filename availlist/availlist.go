package availlist

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/11th-ndn-hackathon/ndn-fch/health"
	"github.com/11th-ndn-hackathon/ndn-fch/logging"
	"github.com/11th-ndn-hackathon/ndn-fch/model"
	"github.com/11th-ndn-hackathon/ndn-fch/routerlist"
	"github.com/pkg/math"
	"go.uber.org/zap"
)

var (
	list        []model.RouterAvail
	listUpdated time.Time
	listLock    sync.RWMutex
)

// List returns available router list.
func List() (l []model.RouterAvail, updated time.Time) {
	listLock.RLock()
	defer listLock.RUnlock()
	return list, listUpdated
}

var logger = logging.New("availlist")

var (
	RefreshInterval time.Duration
	ProbeService    health.Service
	ProbeCount      int
)

type availInfo struct {
	tf model.TransportIPFamily
	id string
}

func refresh(ctx context.Context) {
	routers := routerlist.List()
	availMap := make(map[string]*model.RouterAvail)
	for _, router := range routers {
		router := router
		availMap[router.ID] = &model.RouterAvail{
			Router:    &router,
			Available: map[model.TransportIPFamily]bool{},
		}
	}
	collect := make(chan availInfo)
	go func() {
		for ai := range collect {
			availMap[ai.id].Available[ai.tf] = true
		}
	}()

	var wg sync.WaitGroup
	wg.Add(len(model.TransportIPFamilies))
	var probeErrors int32
	for _, tf := range model.TransportIPFamilies {
		go func(tf model.TransportIPFamily) {
			defer wg.Done()
			request := health.ProbeRequest{
				Transport: tf.TransportType,
				IPFamily:  tf.IPFamily,
				Routers:   routers,
			}

			responses := []health.ProbeResponse{}
			for i := 0; i < ProbeCount; i++ {
				request = request.RandomPingServer()

				logEntry := logger.With(
					zap.String("transport", string(tf.TransportType)),
					zap.Int("ip-family", int(tf.IPFamily)),
					zap.Int("i", i),
					zap.String("target-name", request.Name),
				)

				response, e := ProbeService.Probe(ctx, request)
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
			if nSuccess, nFailure := best.Count(); nSuccess+nFailure == 0 {
				atomic.AddInt32(&probeErrors, 1)
				return
			}

			for _, router := range routers {
				res, ok := best[router.ID]
				if !ok || !res.OK {
					continue
				}
				collect <- availInfo{id: router.ID, tf: tf}
			}
		}(tf)
	}
	wg.Wait()
	close(collect)

	if probeErrors > 0 {
		logger.Warn("some probes failed, not updating")
		return
	}

	listLock.Lock()
	defer listLock.Unlock()
	list = nil
	for _, router := range availMap {
		list = append(list, *router)
	}
	listUpdated = time.Now().UTC()

	if ce := logger.Check(zap.InfoLevel, "updating"); ce != nil {
		ce.Write(zap.Reflect("avail", list))
	}
}

// RefreshLoop refreshes availList periodically.
func RefreshLoop(ctx context.Context) {
	RefreshInterval = time.Duration(math.MaxInt64(int64(RefreshInterval), int64(time.Minute)))

	refreshOnce := func() {
		ctx, cancel := context.WithTimeout(ctx, RefreshInterval*9/10)
		defer cancel()

		t0 := time.Now()
		refresh(ctx)
		logger.Debug("refresh", zap.Duration("duration", time.Since(t0)))
	}

	time.Sleep(time.Second)
	refreshOnce()
	for range time.Tick(RefreshInterval) {
		time.Sleep(time.Duration(rand.Intn(10)) * RefreshInterval / 100)
		refreshOnce()
	}
}
