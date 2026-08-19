package delivery

import (
	"context"
	"net/http"
	"time"
)

type Client struct{ HTTP *http.Client }

func New() *Client { return &Client{HTTP: &http.Client{Timeout: 30 * time.Second}} }
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.HTTP.Do(req.WithContext(ctx))
}
