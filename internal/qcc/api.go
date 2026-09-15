package qcc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL           = "https://www.qcc.com"
	defaultHTTPClientTimeout = 10 * time.Second
	defaultUserAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/153.0.0.0 Safari/537.36"
)

// Options configures a client. BaseURL defaults to the QCC host.
// TID enables QCC request signing, PID is sent as the X-Pid header, and
// CookieSource supplies cookies for outgoing requests.
type Options struct {
	BaseURL      string
	HTTPClient   *http.Client
	Timeout      time.Duration
	TID          string
	PID          string
	CookieSource CookieSource
}

// Client sends requests to the configured API host.
type Client struct {
	baseURL      string
	httpClient   *http.Client
	tid          string
	pid          string
	cookieSource CookieSource
}

// New returns a client configured for the specified API host.
func New(options Options) *Client {
	baseURL := options.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		timeout := options.Timeout
		if timeout == 0 {
			timeout = defaultHTTPClientTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		httpClient:   httpClient,
		tid:          options.TID,
		pid:          options.PID,
		cookieSource: options.CookieSource,
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

	var payload []byte
	signedBody := c.tid != "" && body != nil
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
	if c.pid != "" {
		request.Header.Set("X-Pid", c.pid)
	}
	request.Header.Set("User-Agent", defaultUserAgent)
	if c.cookieSource != nil {
		cookies, err := c.cookieSource.Cookies(ctx)
		if err != nil {
			return nil, err
		}
		for _, cookie := range cookies {
			request.AddCookie(cookie)
		}
	}
	if c.tid != "" {
		sign := GenerateSign(request.URL.Path, string(payload), c.tid)
		request.Header.Set(sign.HeaderName, sign.HeaderValue)
	}
	if signedBody {
		request.Header.Set("Content-Type", "application/json")
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
