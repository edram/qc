package qcc

import (
	"bytes"
	"errors"
	"fmt"
	stdhtml "html"
	"io"
	nethttp "net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/edram/qi/internal/http"
)

// QCC error pages are HTML rather than API payloads, so their title is the stable user-facing message.
var htmlTitlePattern = regexp.MustCompile(`(?is)<title(?:\s[^>]*)?>(.*?)</title\s*>`)

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

func (c *Client) responseMiddleware(ctx *http.Context, next http.Next) error {
	if err := next(); err != nil {
		return err
	}
	if ctx.Response == nil || ctx.Response.StatusCode == nethttp.StatusOK {
		return nil
	}

	response := ctx.Response
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	ctx.Response = nil

	if err := c.redirectResponseError(response); err != nil {
		return err
	}
	statusCode := response.StatusCode
	title := responseTitle(body)
	if title == "" {
		return fmt.Errorf("qcc: HTTP %d", statusCode)
	}
	return fmt.Errorf("qcc: HTTP %d: %s", statusCode, title)
}

func (c *Client) redirectResponseError(response *nethttp.Response) error {
	if response.StatusCode < nethttp.StatusMultipleChoices || response.StatusCode >= nethttp.StatusBadRequest {
		return nil
	}
	locationHeader := response.Header.Get("Location")
	location, err := url.Parse(locationHeader)
	if err == nil && location.Path == "/weblogin" && location.Query().Has("back") {
		return errors.New("qcc: login required; log in to QCC again")
	}
	if err == nil && location.Scheme == "https" && strings.EqualFold(location.Host, "www.qcc.com") && location.Path == "/405.html" {
		if source, ok := c.cookieSource.(cookieCacheClearer); ok {
			if err := source.clearCache(); err != nil {
				return fmt.Errorf("qcc: request rejected; User-Agent may not match the imported cookies; delete cached cookies before running auth import again: %w", err)
			}
		}
		return errors.New("qcc: request rejected; User-Agent may not match the imported cookies; cached cookies were deleted; run auth import again")
	}
	if locationHeader != "" {
		return fmt.Errorf("qcc: redirect to %s", locationHeader)
	}
	return nil
}

func responseTitle(body []byte) string {
	match := htmlTitlePattern.FindSubmatch(body)
	if len(match) != 2 {
		return ""
	}
	return strings.TrimSpace(stdhtml.UnescapeString(string(match[1])))
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
