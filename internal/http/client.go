// Package http provides a middleware-capable HTTP client.
package http

import (
	stdhttp "net/http"
	"sync"
)

// Client sends HTTP requests through middleware.
type Client struct {
	baseURL      string
	prefix       string
	httpClient   *stdhttp.Client
	middlewares  []Middleware
	handler      handler
	middlewareMu sync.Mutex
}

// New creates a Client from options.
func New(providedOptions ...Option) *Client {
	options := clientOptions{}
	for _, option := range providedOptions {
		if option != nil {
			option(&options)
		}
	}

	return &Client{
		baseURL:    options.baseURL,
		prefix:     options.prefix,
		httpClient: configureHTTPClient(options.httpClient, options),
	}
}

func configureHTTPClient(httpClient *stdhttp.Client, options clientOptions) *stdhttp.Client {
	if httpClient == nil {
		httpClient = &stdhttp.Client{}
	} else {
		httpClientCopy := *httpClient
		httpClient = &httpClientCopy
	}
	if options.timeoutSet {
		httpClient.Timeout = options.timeout
	}
	if options.withoutRedirect {
		httpClient.CheckRedirect = func(*stdhttp.Request, []*stdhttp.Request) error {
			return stdhttp.ErrUseLastResponse
		}
	}
	return httpClient
}

// AddMiddleware appends middleware to the request pipeline.
func (c *Client) AddMiddleware(middleware Middleware) *Client {
	if middleware == nil {
		panic("http: middleware cannot be nil")
	}
	c.middlewareMu.Lock()
	defer c.middlewareMu.Unlock()
	if c.handler != nil {
		panic("http: middleware cannot be added after a request")
	}
	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Do sends an existing request through the middleware chain.
func (c *Client) Do(request *stdhttp.Request) (*stdhttp.Response, error) {
	c.middlewareMu.Lock()
	if c.handler == nil {
		c.handler = c.middlewareChain()
	}
	handler := c.handler
	c.middlewareMu.Unlock()

	ctx := &Context{Client: c, Request: request}
	err := handler(ctx)
	return ctx.Response, err
}
