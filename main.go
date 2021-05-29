package main

import (
	"context"
	"math/rand"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/11th-ndn-hackathon/ndn-fch/health"
	"github.com/11th-ndn-hackathon/ndn-fch/logging"
	"github.com/pkg/math"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var logger = logging.New("main")

var app = &cli.App{
	Name: "ndn-fch-control",
	Flags: []cli.Flag{
		&cli.DurationFlag{
			Name:        "interval",
			Usage:       "refresh interval",
			Destination: &refreshInterval,
			Value:       5 * time.Minute,
		},
		&cli.StringFlag{
			Name:     "probe",
			Usage:    "UDP/WebSockets health probe URI",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "probe3",
			Usage:    "HTTP3 health probe URI",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "api",
			Usage:    "frontend API URI",
			Required: true,
		},
		&cli.IntFlag{
			Name:        "probe-count",
			Usage:       "number of probe attempts",
			Destination: &nProbes,
			Value:       3,
		},
	},
	Before: func(c *cli.Context) (e error) {
		refreshInterval = time.Duration(math.MaxInt64(int64(refreshInterval), int64(time.Minute)))

		if probe, e = health.NewHTTPDispatcher(c.String("probe"), c.String("probe3")); e != nil {
			return cli.Exit(e, 1)
		}

		apiUri, e := url.Parse(c.String("api"))
		if e != nil {
			return cli.Exit(e, 1)
		}
		apiUri.Path = path.Join(apiUri.Path, "routers")
		apiUpdateUri = apiUri.String()
		return nil
	},
	Action: func(c *cli.Context) (e error) {
		refreshOnce := func() {
			ctx, cancel := context.WithTimeout(c.Context, refreshInterval*9/10)
			defer cancel()

			t0 := time.Now()
			if e := refresh(ctx); e != nil {
				logger.Error("refresh error", zap.Error(e))
			} else {
				logger.Debug("refresh", zap.Duration("duration", time.Since(t0)))
			}
		}

		time.Sleep(time.Second)
		refreshOnce()
		for range time.Tick(refreshInterval) {
			time.Sleep(time.Duration(rand.Intn(10)) * refreshInterval / 100)
			refreshOnce()
		}
		return nil
	},
}

func main() {
	rand.Seed(time.Now().UnixNano())
	app.Run(os.Args)
}
