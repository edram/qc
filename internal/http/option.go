package http

import (
	stdhttp "net/http"
	"time"
)

type clientOptions struct {
	baseURL         string
	prefix          string
	httpClient      *stdhttp.Client
	timeout         time.Duration
	timeoutSet      bool
	withoutRedirect bool
}

// Option configures a Client.
type Option func(options *clientOptions)

// WithBaseURL sets the base URL used to resolve endpoints with standard URL semantics.
// A trailing slash makes the final path segment a directory.
func WithBaseURL(baseURL string) Option {
	return func(options *clientOptions) {
		options.baseURL = baseURL
	}
}

// WithPrefix prepends a path or URL to request endpoints before BaseURL resolution.
func WithPrefix(prefix string) Option {
	return func(options *clientOptions) {
		options.prefix = prefix
	}
}

// WithHTTPClient sets the underlying HTTP client without modifying the provided instance.
func WithHTTPClient(httpClient *stdhttp.Client) Option {
	return func(options *clientOptions) {
		if httpClient == nil {
			return
		}
		options.httpClient = httpClient
	}
}

// WithTimeout sets the total timeout for a request.
func WithTimeout(timeout time.Duration) Option {
	return func(options *clientOptions) {
		options.timeout = timeout
		options.timeoutSet = true
	}
}

// WithoutRedirects prevents the client from following HTTP redirects.
func WithoutRedirects() Option {
	return func(options *clientOptions) {
		options.withoutRedirect = true
	}
}
