package delivery

import (
	"context"
	"net/http"
	"time"
)

type Client struct{ HTTP *http.Client }

func New() *Client { return &Client{HTTP: &http.Client{Timeout: 30 * time.Second}} }
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// req.WithContext returns a shallow copy preserving the request's method,
	// header, and body while binding the caller's ctx to the outgoing request,
	// so the ctx's deadline, cancellation, and trace semantics reach transport.
	return c.HTTP.Do(req.WithContext(ctx))
}
