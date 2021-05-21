package health

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HTTPClient implements a Service that probes router health via a backend HTTP service.
type HTTPClient struct {
	probeUri string
}

var _ Service = &HTTPClient{}

// Probe implements Service interface.
func (c *HTTPClient) Probe(ctx context.Context, req ProbeRequest) (res ProbeResponse, e error) {
	form := url.Values{}
	form.Set("transport", string(req.Transport))
	form.Set("name", req.Name)
	if req.Suffix {
		form.Set("suffix", "1")
	}

	reqMap := make(map[string]string)
	for _, router := range req.Routers {
		s := req.Transport.RouterString(router)
		if s == "" {
			continue
		}
		reqMap[s] = router.ID
		form.Add("router", s)
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
	body, e := io.ReadAll(hRes.Body)
	if e != nil {
		return nil, e
	}

	var resMap map[string]ProbeRouterResult
	e = json.Unmarshal(body, &resMap)
	if e != nil {
		return nil, e
	}

	res = make(ProbeResponse)
	for s, r := range resMap {
		res[reqMap[s]] = r
	}
	return res, nil
}

// NewClient creates a Client from base URI.
func NewClient(uri string) (c *HTTPClient, e error) {
	u, e := url.Parse(uri)
	if e != nil {
		return nil, e
	}

	u.Path = "/probe"
	c = &HTTPClient{
		probeUri: u.String(),
	}
	return c, nil
}
