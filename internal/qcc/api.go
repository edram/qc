package qcc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultBaseURL           = "https://www.qcc.com"
	defaultHTTPClientTimeout = 10 * time.Second
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
	HTTPClient   *http.Client
	Timeout      time.Duration
	TID          string
	PID          string
	Profile      string
	UserAgent    string
	CookieSource CookieSource
}

// Client sends requests to the configured API host.
type Client struct {
	baseURL      string
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

	httpClient := options.HTTPClient
	if httpClient == nil {
		timeout := options.Timeout
		if timeout == 0 {
			timeout = defaultHTTPClientTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	// Keep redirect responses visible to callers instead of issuing another request.
	httpClientCopy := *httpClient
	httpClientCopy.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	httpClient = &httpClientCopy

	return &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		httpClient:   httpClient,
		tid:          options.TID,
		pid:          options.PID,
		userAgent:    strings.TrimSpace(options.UserAgent),
		cookieSource: cookieSource,
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
	if c.userAgent == "" {
		return nil, ErrUserAgentRequired
	}
	requestURL, err := c.requestURL(endpoint)
	if err != nil {
		return nil, err
	}
	pid, tid, err := c.getPIDAndTID(ctx)
	if err != nil {
		return nil, err
	}

	var payload []byte
	signedBody := body != nil
	if signedBody {
		payload, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Pid", pid)
	request.Header.Set("User-Agent", c.userAgent)
	if err := c.addCookies(ctx, request); err != nil {
		return nil, err
	}
	sign := GenerateSign(request.URL.Path, string(payload), tid)
	request.Header.Set(sign.HeaderName, sign.HeaderValue)
	if signedBody {
		request.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(request)
}

func (c *Client) addCookies(ctx context.Context, request *http.Request) error {
	if c.cookieSource == nil {
		return nil
	}
	cookies, err := c.cookieSource.Cookies(ctx)
	if err != nil {
		return err
	}
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	return nil
}

func (c *Client) requestURL(endpoint string) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.IsAbs() || c.baseURL == "" {
		return endpoint, err
	}

	return c.baseURL + "/" + strings.TrimLeft(endpoint, "/"), nil
}
