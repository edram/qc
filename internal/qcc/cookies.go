package qcc

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/steipete/sweetcookie"
)

// ErrNoCookies is returned when the configured browser has no QCC cookies.
var ErrNoCookies = errors.New("qcc: no browser cookies found")

var readBrowserCookies = sweetcookie.Get

// CookieSource supplies cookies for an outgoing QCC request.
type CookieSource interface {
	Cookies(context.Context) ([]*http.Cookie, error)
}

// BrowserCookieSource reads QCC cookies from local browser profiles.
// An empty Browser uses sweetcookie's default browser discovery order.
type BrowserCookieSource struct {
	Browser sweetcookie.Browser
	Profile string // Profile applies only when Browser is set.
}

// Cookies reads the cookies that match the QCC web origin.
func (s BrowserCookieSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	options := sweetcookie.Options{
		URL:     defaultBaseURL,
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
