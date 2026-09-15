package qcc

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultHTTPClientTimeout = 10 * time.Second

// Options configures a client. BaseURL is optional when callers use absolute URLs.
type Options struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// Client sends requests to the configured API host.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New returns a client configured for the specified API host.
func New(options Options) *Client {
	httpClient := options.HTTPClient
	if httpClient == nil {
		timeout := options.Timeout
		if timeout == 0 {
			timeout = defaultHTTPClientTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		baseURL:    strings.TrimRight(options.BaseURL, "/"),
		httpClient: httpClient,
	}
}

// Get sends a GET request. The caller must close the response body.
func (c *Client) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.request(ctx, http.MethodGet, endpoint, nil)
}

// Post sends a POST request with body. The caller must close the response body.
func (c *Client) Post(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	return c.request(ctx, http.MethodPost, endpoint, body)
}

func (c *Client) request(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	requestURL, err := c.requestURL(endpoint)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(request)
}

func (c *Client) requestURL(endpoint string) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.IsAbs() || c.baseURL == "" {
		return endpoint, err
	}

	return c.baseURL + "/" + strings.TrimLeft(endpoint, "/"), nil
}
