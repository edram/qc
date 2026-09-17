package qcc

import (
	"bytes"
	"io"

	"github.com/edram/qi/internal/http"
)

func (c *Client) sessionMiddleware(ctx *http.Context, next http.Next) error {
	if c.userAgent == "" {
		return ErrUserAgentRequired
	}
	if shouldSignRequest(ctx) {
		pid, _ := c.identifiers()
		ctx.Request.Header.Set("X-Pid", pid)
	}
	ctx.Request.Header.Set("User-Agent", c.userAgent)

	if c.cookieSource != nil {
		cookies, err := c.cookieSource.Cookies(ctx.Request.Context())
		if err != nil {
			return err
		}
		for _, cookie := range cookies {
			ctx.Request.AddCookie(cookie)
		}
	}
	return next()
}

func (c *Client) signMiddleware(ctx *http.Context, next http.Next) error {
	if !shouldSignRequest(ctx) {
		return next()
	}
	payload, hasBody, err := readRequestBody(ctx)
	if err != nil {
		return err
	}
	_, tid := c.identifiers()
	sign := GenerateSign(ctx.Request.URL.Path, string(payload), tid)
	ctx.Request.Header.Set(sign.HeaderName, sign.HeaderValue)
	if hasBody {
		ctx.Request.Header.Set("Content-Type", "application/json")
	}
	return next()
}

func readRequestBody(ctx *http.Context) ([]byte, bool, error) {
	if ctx.Request.Body == nil {
		return nil, false, nil
	}
	payload, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, true, err
	}
	_ = ctx.Request.Body.Close()
	ctx.Request.Body = io.NopCloser(bytes.NewReader(payload))
	return payload, true, nil
}

func shouldSignRequest(ctx *http.Context) bool {
	return !ctx.Client.IsBaseRequest(ctx.Request)
}
