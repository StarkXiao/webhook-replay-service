package delivery

import (
	"context"
	"net/http"
)

type Client struct{ HTTP *http.Client }

func New() *Client { return &Client{} }
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.HTTP.Do(req.WithContext(ctx))
}
