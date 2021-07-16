package availlist

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
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
)

type availInfo struct {
	tf model.TransportIPFamily
	id string
	ok bool
}

func refresh(ctx context.Context) {
	routers := routerlist.List()
	oldAvail, _ := List()
	var destinations []string
	for _, router := range oldAvail {
		if router.CountAvail() == 0 {
			continue
		}
		destinations = append(destinations, router.Prefix)
	}
	if len(destinations) == 0 {
		for _, router := range routers {
			destinations = append(destinations, router.Prefix)
		}
	}

	availMap := make(map[string]*model.RouterAvail)
	for _, router := range routers {
		router := router
		availMap[router.ID] = &model.RouterAvail{
			Router:    &router,
			Available: map[model.TransportIPFamily]bool{},
		}
	}
	for _, router := range oldAvail {
		newRouter := availMap[router.ID]
		if newRouter == nil {
			continue
		}
		for tf, ok := range router.Available {
			newRouter.Available[tf] = ok
		}
	}

	collect := make(chan availInfo)
	go func() {
		for ai := range collect {
			availMap[ai.id].Available[ai.tf] = ai.ok
		}
	}()

	var wg sync.WaitGroup
	for _, router := range routers {
		for _, tf := range model.TransportIPFamilies {
			connect := router.ConnectString(tf.Transport, false)
			if !router.HasIPFamily(tf.Family) || connect == "" {
				continue
			}
			time.Sleep(10 * time.Millisecond)
			wg.Add(1)
			go func(router model.Router, tf model.TransportIPFamily, connect string) {
				defer wg.Done()
				request := health.ProbeRequest{
					TransportIPFamily: tf,
					Router:            connect,
				}
				for _, dest := range destinations {
					request.Names = append(request.Names, fmt.Sprintf("%s/ping/ndn-fch-2021/%d", dest, rand.Int()))
				}

				logEntry := logger.With(
					zap.String("transport", string(tf.Transport)),
					zap.Int("ip-family", int(tf.Family)),
					zap.String("router", connect),
					zap.Int("name-count", len(request.Names)),
				)

				response, e := ProbeService.Probe(ctx, request)
				if e != nil {
					logEntry.Warn("probe error", zap.Error(e))
					return
				}

				if !response.Connected {
					logEntry.Debug("probe response",
						zap.Bool("connected", response.Connected),
						zap.String("connect-error", response.ConnectError),
					)
					collect <- availInfo{id: router.ID, tf: tf, ok: false}
					return
				}

				nSuccess, nFailure := response.Count()
				verdict := nSuccess*2 > nFailure
				logEntry.Debug("probe response",
					zap.Int("success-count", nSuccess),
					zap.Int("failure-count", nFailure),
					zap.Bool("verdict", verdict),
				)
				collect <- availInfo{id: router.ID, tf: tf, ok: verdict}
			}(router, tf, connect)
		}
	}
	wg.Wait()
	close(collect)

	listLock.Lock()
	defer listLock.Unlock()
	list = nil
	for _, router := range availMap {
		list = append(list, *router)
	}
	listUpdated = time.Now().UTC()
	logger.Info("updating", zap.Any("avail", list))
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
