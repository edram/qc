package aiqicha

import (
	"encoding/base64"
	"fmt"
	nethttp "net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/edram/qi/internal/http"
)

func (c *Client) sessionMiddleware(ctx *http.Context, next http.Next) error {
	request := ctx.Request
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	// Browser credentials must stay on the configured origin.
	if request.URL.Scheme != base.Scheme || request.URL.Host != base.Host {
		return fmt.Errorf("aiqicha: request outside configured origin")
	}
	items, err := c.cookieSource.Cookies(request.Context())
	if err != nil {
		return fmt.Errorf("aiqicha cookies: %w", err)
	}
	jar, _ := cookiejar.New(nil)
	for _, cookie := range items {
		if cookie != nil {
			jar.SetCookies(request.URL, []*nethttp.Cookie{cookie})
		}
	}
	matching := jar.Cookies(request.URL)
	if len(matching) == 0 {
		return fmt.Errorf("aiqicha: no matching cookies; run 'qc auth import --provider aiqicha'")
	}
	for _, cookie := range matching {
		request.AddCookie(cookie)
	}
	if c.userAgent != "" {
		request.Header.Set("User-Agent", c.userAgent)
	} else if request.UserAgent() == "" {
		request.Header.Set("User-Agent", "Go-http-client/1.1")
	}
	// The website's request interceptor uses aqc_ssr when its browser signer is
	// unavailable. It binds the timestamp to the User-Agent actually sent.
	// https://xinpub.cdn.bcebos.com/aiqicha/static/1793259024/js/vendor.92cf53d085f.js
	signature := strconv.FormatInt(time.Now().UnixMilli(), 10) + "cal110" + request.UserAgent()
	request.Header.Set("aqc_ssr", base64.StdEncoding.EncodeToString([]byte(signature)))
	request.Header.Set("Accept", "application/json, text/plain, */*")
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	return next()
}

func (c *Client) responseMiddleware(ctx *http.Context, next http.Next) error {
	if err := next(); err != nil {
		return err
	}
	if ctx.Response == nil || ctx.Response.StatusCode == nethttp.StatusOK {
		return nil
	}
	response := ctx.Response
	_ = response.Body.Close()
	ctx.Response = nil
	return fmt.Errorf("aiqicha: HTTP %d", response.StatusCode)
}
