package http

import (
	"context"
	"io"
	stdhttp "net/http"
	"net/url"
	"strings"
)

// ScopedClient sends requests through a Client with one request context.
type ScopedClient struct {
	client *Client
	ctx    context.Context
}

// WithContext returns a request scope without mutating the shared client.
func (c *Client) WithContext(ctx context.Context) *ScopedClient {
	if ctx == nil {
		panic("http: nil context")
	}
	return &ScopedClient{client: c, ctx: ctx}
}

// Get sends a GET request to endpoint.
func (c *Client) Get(endpoint string) (*stdhttp.Response, error) {
	return c.Request(stdhttp.MethodGet, endpoint, nil)
}

// Get sends a GET request to endpoint with the scoped context.
func (c *ScopedClient) Get(endpoint string) (*stdhttp.Response, error) {
	return c.Request(stdhttp.MethodGet, endpoint, nil)
}

// Post sends a POST request with body to endpoint.
func (c *Client) Post(endpoint string, body io.Reader) (*stdhttp.Response, error) {
	return c.Request(stdhttp.MethodPost, endpoint, body)
}

// Post sends a POST request with body to endpoint with the scoped context.
func (c *ScopedClient) Post(endpoint string, body io.Reader) (*stdhttp.Response, error) {
	return c.Request(stdhttp.MethodPost, endpoint, body)
}

// Request builds and sends an HTTP request.
func (c *Client) Request(method, endpoint string, body io.Reader) (*stdhttp.Response, error) {
	return c.request(context.Background(), method, endpoint, body)
}

// Request builds and sends an HTTP request with the scoped context.
func (c *ScopedClient) Request(method, endpoint string, body io.Reader) (*stdhttp.Response, error) {
	return c.client.request(c.ctx, method, endpoint, body)
}

func (c *Client) request(ctx context.Context, method, endpoint string, body io.Reader) (*stdhttp.Response, error) {
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
