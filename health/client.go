package health

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/11th-ndn-hackathon/ndn-fch/model"
)

// HTTPClient implements a Service that probes router health via a backend HTTP service.
type HTTPClient struct {
	probeUri string
}

var _ Service = &HTTPClient{}

// Probe implements Service interface.
func (c *HTTPClient) Probe(ctx context.Context, req ProbeRequest) (res ProbeResponse, e error) {
	form := url.Values{}
	form.Set("transport", model.TransportIPFamily{
		TransportType: req.Transport,
		IPFamily:      req.IPFamily,
	}.String())
	form.Set("name", req.Name)
	if req.Suffix {
		form.Set("suffix", "1")
	}

	reqMap := make(map[string]string)
	for _, router := range req.Routers {
		switch req.IPFamily {
		case model.IPv4:
			if !router.IPv4 {
				continue
			}
		case model.IPv6:
			if !router.IPv6 {
				continue
			}
		}

		s := router.ConnectString(req.Transport, false)
		if s == "" {
			continue
		}

		reqMap[s] = router.ID
		form.Add("router", s)
	}

	res = make(ProbeResponse)
	if len(reqMap) == 0 {
		return res, nil
	}

	hReq, e := http.NewRequestWithContext(ctx, http.MethodPost, c.probeUri, strings.NewReader(form.Encode()))
	if e != nil {
		return nil, e
	}
	hReq.Header.Set("content-type", "application/x-www-form-urlencoded")

	hRes, e := http.DefaultClient.Do(hReq)
	if e != nil {
		return nil, e
	}
	if hRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", hRes.StatusCode)
	}
	body, e := io.ReadAll(hRes.Body)
	if e != nil {
		return nil, e
	}

	var resMap map[string]ProbeRouterResult
	e = json.Unmarshal(body, &resMap)
	if e != nil {
		return nil, e
	}

	for s, r := range resMap {
		res[reqMap[s]] = r
	}
	return res, nil
}

// NewHTTPClient creates a Client from base URI.
func NewHTTPClient(uri string) (c *HTTPClient, e error) {
	u, e := url.Parse(uri)
	if e != nil {
		return nil, e
	}
	u.Path = path.Join(u.Path, "probe")
	c = &HTTPClient{
		probeUri: u.String(),
	}
	return c, nil
}
