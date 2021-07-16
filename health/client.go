package health

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// HTTPClient implements a Service that probes router health via a backend HTTP service.
type HTTPClient struct {
	probeUri string
}

var _ Service = &HTTPClient{}

// Probe implements Service interface.
func (c *HTTPClient) Probe(ctx context.Context, req ProbeRequest) (res ProbeResponse, e error) {
	jReq, _ := json.Marshal(req)
	hReq, e := http.NewRequestWithContext(ctx, http.MethodPost, c.probeUri, bytes.NewReader(jReq))
	if e != nil {
		return res, e
	}
	hReq.Header.Set("content-type", "application/json")

	hRes, e := http.DefaultClient.Do(hReq)
	if e != nil {
		return res, e
	}
	if hRes.StatusCode != http.StatusOK {
		return res, fmt.Errorf("HTTP %d", hRes.StatusCode)
	}
	jRes, e := io.ReadAll(hRes.Body)
	if e != nil {
		return res, e
	}

	e = json.Unmarshal(jRes, &res)
	return res, e
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
