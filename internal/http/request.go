package http

import (
	"context"
	"io"
	stdhttp "net/http"
	"net/url"
	"strings"
)

// Get sends a GET request to endpoint.
func (c *Client) Get(ctx context.Context, endpoint string) (*stdhttp.Response, error) {
	return c.Request(ctx, stdhttp.MethodGet, endpoint, nil)
}

// Post sends a POST request with body to endpoint.
func (c *Client) Post(ctx context.Context, endpoint string, body io.Reader) (*stdhttp.Response, error) {
	return c.Request(ctx, stdhttp.MethodPost, endpoint, body)
}

// Request builds and sends an HTTP request.
func (c *Client) Request(ctx context.Context, method, endpoint string, body io.Reader) (*stdhttp.Response, error) {
	requestURL, err := c.resolveURL(endpoint)
	if err != nil {
		return nil, err
	}
	request, err := stdhttp.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, err
	}
	return c.Do(request)
}

func (c *Client) resolveURL(endpoint string) (string, error) {
	if c.prefix != "" {
		endpoint = strings.TrimRight(c.prefix, "/") + "/" + strings.TrimLeft(endpoint, "/")
	}
	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil || parsedEndpoint.IsAbs() || c.baseURL == "" {
		return endpoint, err
	}
	parsedBaseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	return parsedBaseURL.ResolveReference(parsedEndpoint).String(), nil
}

// IsBaseRequest reports whether request targets the client's configured base URL.
func (c *Client) IsBaseRequest(request *stdhttp.Request) bool {
	baseURL, err := c.resolveURL("")
	return err == nil && request.URL.String() == baseURL
}
