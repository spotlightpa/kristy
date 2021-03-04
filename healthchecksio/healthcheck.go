// Package healthchecksio is a wrapper for API calls to HealthChecks.io
package healthchecksio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// Client is a convenient way to ping HealthChecks.io
type Client struct {
	c    *http.Client
	uuid string
}

// New returns a configured client. If c is nil, http.DefaultClient is used.
func New(uuid string, c *http.Client) *Client {
	if c == nil {
		c = http.DefaultClient
	}
	return &Client{c, uuid}
}

func (cl *Client) req(ctx context.Context, url string, body []byte) (err error) {
	var r io.Reader
	if len(body) > 0 {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, r)
	if err != nil {
		return err
	}
	res, err := cl.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return StatusErr(res.StatusCode)
	}
	// HealthChecks.io shouldn't be giving us anything to discard,
	// but if they do, read some of it to try to reuse connections
	const maxDiscardSize = 640 * 1 << 10
	if _, err = io.CopyN(io.Discard, res.Body, maxDiscardSize); err == io.EOF {
		err = nil
	}
	return err
}

// Start calls the start HealthChecks.io endpoint
func (cl *Client) Start(ctx context.Context) error {
	url := fmt.Sprintf("https://hc-ping.com/%s/start", cl.uuid)
	return maybeNote(cl.req(ctx, url, nil), "problem sending start signal to Healthchecks.io")
}

// Status calls the HealthChecks.io status endpoint
func (cl *Client) Status(ctx context.Context, code int, msg []byte) error {
	url := fmt.Sprintf("https://hc-ping.com/%s/%d", cl.uuid, code)
	return maybeNote(cl.req(ctx, url, msg), "problem sending status to Healthchecks.io")
}
