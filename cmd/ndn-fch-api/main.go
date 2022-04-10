// Command ndn-fch-api runs NDN-FCH 2021 API service.
package main

import (
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/11th-ndn-hackathon/ndn-fch/availlist"
	"github.com/11th-ndn-hackathon/ndn-fch/health"
	"github.com/11th-ndn-hackathon/ndn-fch/routerlist"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name: "ndn-fch",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "listen",
			Usage: "HTTP server address",
			Value: ":5000",
		},
		&cli.DurationFlag{
			Name:        "interval",
			Usage:       "refresh interval",
			Destination: &availlist.RefreshInterval,
			Value:       availlist.RefreshInterval,
		},
		&cli.IntFlag{
			Name:        "max-names",
			Usage:       "maximum destination names",
			Destination: &availlist.MaxNames,
			Value:       availlist.MaxNames,
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
	},
	Before: func(c *cli.Context) (e error) {
		if availlist.ProbeService, e = health.NewHTTPDispatcher(c.String("probe"), c.String("probe3")); e != nil {
			return cli.Exit(e, 1)
		}
		return nil
	},
	Action: func(c *cli.Context) (e error) {
		routerlist.Load()
		go availlist.RefreshLoop(c.Context)
		return cli.Exit(http.ListenAndServe(c.String("listen"), nil), 1)
	},
}

func main() {
	rand.Seed(time.Now().UnixNano())
	app.Run(os.Args)
}
