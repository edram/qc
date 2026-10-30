package aiqicha

import (
	"context"
	"io"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/edram/qi/internal/cookies"
	"github.com/edram/qi/internal/http"
)

const (
	defaultHTTPClientTimeout = 10 * time.Second
	defaultBaseURL           = "https://www.aiqicha.com"
)

// Options configures the Aiqicha origin and browser session. UserAgent is an
// optional header override. CookieSource overrides the per-profile browser-backed cache.
type Options struct {
	BaseURL      string
	HTTPClient   *nethttp.Client
	Timeout      time.Duration
	Profile      string
	UserAgent    string
	CookieSource cookies.Source
}

// Client sends requests to the configured API host.
type Client struct {
	baseURL      string
	httpClient   *http.Client
	userAgent    string
	cookieSource cookies.Source
}

// New returns a client configured for the specified API host.
func New(options Options) *Client {
	if options.BaseURL == "" {
		options.BaseURL = defaultBaseURL
	}
	timeout := options.Timeout
	if timeout == 0 && options.HTTPClient != nil {
		timeout = options.HTTPClient.Timeout
	}
	if timeout == 0 {
		timeout = defaultHTTPClientTimeout
	}
	source := options.CookieSource
	if source == nil {
		source = defaultCookieSource{profile: options.Profile}
	}

	client := &Client{
		baseURL: strings.TrimRight(options.BaseURL, "/"),
		httpClient: http.New(
			http.WithBaseURL(options.BaseURL),
			http.WithHTTPClient(options.HTTPClient),
			http.WithTimeout(timeout),
			http.WithoutRedirects(),
		),
		userAgent:    strings.TrimSpace(options.UserAgent),
		cookieSource: source,
	}
	client.httpClient.AddMiddleware(client.sessionMiddleware)
	client.httpClient.AddMiddleware(client.responseMiddleware)
	return client
}

// GetJSON decodes the response's data using Aiqicha's ddw digit substitution.
func (c *Client) GetJSON(ctx context.Context, endpoint string, result any, options ...http.RequestOption) error {
	response, err := c.Get(ctx, endpoint, options...)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return decodeResponse(response.Body, result)
}

// Get sends a GET request. The caller must close the response body on success.
func (c *Client) Get(ctx context.Context, endpoint string, options ...http.RequestOption) (*nethttp.Response, error) {
	return c.request(ctx, nethttp.MethodGet, endpoint, nil, options...)
}

// Post sends a POST request with body. The caller must close the response body on success.
func (c *Client) Post(ctx context.Context, endpoint string, body io.Reader, options ...http.RequestOption) (*nethttp.Response, error) {
	return c.request(ctx, nethttp.MethodPost, endpoint, body, options...)
}

func (c *Client) request(ctx context.Context, method, endpoint string, body io.Reader, options ...http.RequestOption) (*nethttp.Response, error) {
	return c.httpClient.WithContext(ctx).Request(method, endpoint, body, options...)
}
