package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/11th-ndn-hackathon/ndn-fch-control/health"
	"github.com/11th-ndn-hackathon/ndn-fch-control/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockRouters = []model.Router{
	{ID: "dual", Host: "dual.example.net", IPv4: true, IPv6: true,
		UDPPort: 1201, WebSocketPort: 1202, HTTP3Port: 1203},
	{ID: "only4", Host: "only4.example.net", IPv4: true, IPv6: false,
		UDPPort: 1301, WebSocketPort: 1302, HTTP3Port: 1303},
	{ID: "only6", Host: "only6.example.net", IPv4: false, IPv6: true,
		UDPPort: 1301, WebSocketPort: 1302, HTTP3Port: 1303},
	{ID: "udp", Host: "udp.example.net", IPv4: true, IPv6: true, UDPPort: 1401},
	{ID: "wss", Host: "wss.example.net", IPv4: true, IPv6: true, WebSocketPort: 1402},
	{ID: "h3", Host: "h3.example.net", IPv4: true, IPv6: true, HTTP3Port: 1403},
}

func TestClient(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	var form url.Values
	var tm http.ServeMux
	tm.HandleFunc("/probe", func(w http.ResponseWriter, req *http.Request) {
		assert.NoError(req.ParseForm())
		form = req.PostForm

		res := make(health.ProbeResponse)
		for _, router := range form["router"] {
			if strings.Contains(router, "dual") {
				res[router] = health.ProbeRouterResult{OK: false, Error: "timeout"}
			} else {
				res[router] = health.ProbeRouterResult{OK: true, RTT: 10}
			}
		}

		w.Header().Set("content-type", "application/json")
		j, _ := json.Marshal(res)
		w.Write(j)
	})
	ts := httptest.NewServer(&tm)
	defer ts.Close()

	c, e := health.NewHTTPClient(ts.URL)
	require.NoError(e)

	tests := []struct {
		tr       model.TransportType
		expected map[string]bool
	}{
		{model.TransportUDP4, map[string]bool{"dual": false, "only4": true, "udp": true}},
		{model.TransportUDP6, map[string]bool{"dual": false, "only6": true, "udp": true}},
		{model.TransportWebSocket4, map[string]bool{"dual": false, "only4": true, "wss": true}},
		{model.TransportWebSocket6, map[string]bool{"dual": false, "only6": true, "wss": true}},
		{model.TransportH3IPv4, map[string]bool{"dual": false, "only4": true, "h3": true}},
		{model.TransportH3IPv6, map[string]bool{"dual": false, "only6": true, "h3": true}},
	}

	for _, tt := range tests {
		res, e := c.Probe(context.TODO(), health.ProbeRequest{
			Transport: tt.tr,
			Routers:   mockRouters,
			Name:      "/n",
			Suffix:    true,
		})
		if assert.NoError(e) {
			assert.Len(form["router"], len(tt.expected))

			actual := make(map[string]bool)
			for key, value := range res {
				actual[key] = value.OK
			}
			assert.Equal(tt.expected, actual)
		}
	}
}
