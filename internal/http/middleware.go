package http

import (
	"errors"
	stdhttp "net/http"
)

// ErrNextCalledMoreThanOnce indicates that a middleware advanced the chain repeatedly.
var ErrNextCalledMoreThanOnce = errors.New("http: middleware next called more than once")

// Next continues the middleware chain and must be called at most once.
type Next func() error

// Middleware can modify a request before next and its response after next.
type Middleware func(ctx *Context, next Next) error

type handler func(ctx *Context) error

// Context carries the request and response through the middleware chain.
type Context struct {
	Client   *Client
	Request  *stdhttp.Request
	Response *stdhttp.Response
}

func (c *Client) middlewareChain() handler {
	var next handler = func(ctx *Context) error {
		response, err := c.httpClient.Do(ctx.Request)
		ctx.Response = response
		return err
	}

	for i := len(c.middlewares) - 1; i >= 0; i-- {
		middleware := c.middlewares[i]
		nextHandler := next
		next = func(ctx *Context) error {
			called := false
			return middleware(ctx, func() error {
				if called {
					return ErrNextCalledMoreThanOnce
				}
				called = true
				return nextHandler(ctx)
			})
		}
	}
	return next
}
