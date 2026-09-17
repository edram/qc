package qcc

import (
	"context"
	"errors"
	"io"
	nethttp "net/http"
	"strings"
	"sync"

	"github.com/edram/qi/internal/http"
)

const (
	defaultBaseURL = "https://www.qcc.com"
	// defaultUserAgent remains the generic fallback; QCC requests require Options.UserAgent.
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/153.0.0.0 Safari/537.36"
)

// ErrUserAgentRequired prevents QCC from receiving cookies under a different browser identity.
var ErrUserAgentRequired = errors.New("qcc: User-Agent is required because QCC may invalidate cookies when it differs from the login browser; use --user-agent for one command or run 'qc [--profile <name>] config set user-agent <value>' to save it")

// Options configures a client. BaseURL defaults to the QCC host.
// PID and TID override the identifiers discovered from BaseURL.
// Profile selects the browser-backed cookie cache; CookieSource overrides it.
// UserAgent is required for network requests and must match the login browser.
type Options struct {
	BaseURL      string
	HTTPClient   *nethttp.Client
	TID          string
	PID          string
	Profile      string
	UserAgent    string
	CookieSource CookieSource
}

// Client sends requests to the configured API host.
type Client struct {
	httpClient   *http.Client
	tid          string
	pid          string
	userAgent    string
	cookieSource CookieSource
	identifierMu sync.Mutex
}

// New returns a client configured for the specified API host.
func New(options Options) *Client {
	baseURL := options.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	cookieSource := options.CookieSource
	if cookieSource == nil {
		cookieSource = newCookieSource(options.Profile)
	}

	client := &Client{
		httpClient: http.New(
			http.WithBaseURL(baseURL),
			http.WithHTTPClient(options.HTTPClient),
			http.WithoutRedirects(),
		),
		tid:          options.TID,
		pid:          options.PID,
		userAgent:    strings.TrimSpace(options.UserAgent),
		cookieSource: cookieSource,
	}
	client.httpClient.AddMiddleware(client.sessionMiddleware)
	client.httpClient.AddMiddleware(client.signMiddleware)
	return client
}

// Get sends a GET request. The caller must close the response body.
func (c *Client) Get(ctx context.Context, endpoint string) (*nethttp.Response, error) {
	return c.request(ctx, nethttp.MethodGet, endpoint, nil)
}

// Post sends a POST request with body. The caller must close the response body.
func (c *Client) Post(ctx context.Context, endpoint string, body io.Reader) (*nethttp.Response, error) {
	return c.request(ctx, nethttp.MethodPost, endpoint, body)
}

func (c *Client) request(ctx context.Context, method, endpoint string, body io.Reader) (*nethttp.Response, error) {
	if _, _, err := c.getPIDAndTID(ctx); err != nil {
		return nil, err
	}
	return c.httpClient.Request(ctx, method, endpoint, body)
}
