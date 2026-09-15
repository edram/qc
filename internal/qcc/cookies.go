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

// BrowserCookieSource reads QCC cookies from a local browser profile.
type BrowserCookieSource struct {
	Browser sweetcookie.Browser
	Profile string
}

// Cookies reads the cookies that match the QCC web origin.
func (s BrowserCookieSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	browser := s.Browser
	if browser == "" {
		browser = sweetcookie.BrowserChrome
	}

	options := sweetcookie.Options{
		URL:      defaultBaseURL,
		Browsers: []sweetcookie.Browser{browser},
		Mode:     sweetcookie.ModeFirst,
		Timeout:  5 * time.Second,
	}
	if s.Profile != "" {
		options.Profiles = map[sweetcookie.Browser]string{browser: s.Profile}
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
