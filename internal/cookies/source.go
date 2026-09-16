package cookies

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/steipete/sweetcookie"
)

// ErrNoCookies is returned when a browser has no cookies for the requested URL.
var ErrNoCookies = errors.New("no browser cookies found")

var readBrowserCookies = sweetcookie.Get

// Source supplies cookies to an HTTP client.
type Source interface {
	Cookies(context.Context) ([]*http.Cookie, error)
}

// BrowserSource reads cookies for URL from local browser profiles.
// An empty Browser uses sweetcookie's default browser discovery order.
type BrowserSource struct {
	URL     string
	Browser sweetcookie.Browser
	Profile string // Profile applies only when Browser is set.
}

func (s BrowserSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	options := sweetcookie.Options{
		URL:     s.URL,
		Mode:    sweetcookie.ModeFirst,
		Timeout: 5 * time.Second,
	}
	if s.Browser != "" {
		options.Browsers = []sweetcookie.Browser{s.Browser}
		if s.Profile != "" {
			options.Profiles = map[sweetcookie.Browser]string{s.Browser: s.Profile}
		}
	}
	result, err := readBrowserCookies(ctx, options)
	if err != nil {
		return nil, err
	}
	if len(result.Cookies) == 0 {
		// Preserve sweetcookie diagnostics so users can distinguish missing profiles and decryption failures.
		// Based on: https://github.com/openclaw/spogo/blob/3d4edc230f5ece848b81642e95f3682f3ed3e8d7/internal/cookies/source.go
		if len(result.Warnings) > 0 {
			return nil, fmt.Errorf("%w; %s", ErrNoCookies, strings.Join(result.Warnings, "; "))
		}
		return nil, ErrNoCookies
	}

	cookies := make([]*http.Cookie, 0, len(result.Cookies))
	for _, cookie := range result.Cookies {
		httpCookie := &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HTTPOnly,
		}
		if cookie.Expires != nil {
			httpCookie.Expires = *cookie.Expires
		}
		cookies = append(cookies, httpCookie)
	}
	return cookies, nil
}

// SetReadCookies replaces the browser reader and returns a restore function.
// It is intended for tests in packages that compose BrowserSource.
func SetReadCookies(read func(context.Context, sweetcookie.Options) (sweetcookie.Result, error)) func() {
	previous := readBrowserCookies
	readBrowserCookies = read
	return func() { readBrowserCookies = previous }
}
