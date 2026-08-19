package delivery

import (
	"context"
	"net/http"
	"time"
)

type Client struct{ HTTP *http.Client }

func New() *Client { return &Client{HTTP: &http.Client{Timeout: 30 * time.Second}} }
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Clone preserves the request fields while replacing only the context that
	// the transport observes. This keeps cancellation, deadlines, and tracing
	// values tied to the caller's request lifecycle.
	return c.HTTP.Do(req.Clone(ctx))
}
