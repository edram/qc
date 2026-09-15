package cookies

import (
	"context"
	"errors"
	"net/http"
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
