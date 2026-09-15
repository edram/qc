package qcc

import (
	"context"
	"net/http"

	sharedcookies "github.com/edram/qi/internal/cookies"
	"github.com/steipete/sweetcookie"
)

const (
	cookieApplicationName = "qc"
	cookieSourceName      = "qcc"
	cookieCacheProfile    = "default"
)

// CookieSource supplies cookies for an outgoing QCC request.
type CookieSource = sharedcookies.Source

// ErrNoCookies is returned when the configured browser has no QCC cookies.
var ErrNoCookies = sharedcookies.ErrNoCookies

// BrowserCookieSource reads QCC cookies from local browser profiles.
// An empty Browser uses sweetcookie's default browser discovery order.
type BrowserCookieSource struct {
	Browser sweetcookie.Browser
	Profile string // Profile applies only when Browser is set.
}

// Cookies reads the cookies that match the QCC web origin.
func (s BrowserCookieSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	return (sharedcookies.BrowserSource{
		URL:     defaultBaseURL,
		Browser: s.Browser,
		Profile: s.Profile,
	}).Cookies(ctx)
}

type defaultCookieSource struct{}

func newCookieSource() CookieSource {
	return defaultCookieSource{}
}

func (defaultCookieSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	path, err := sharedcookies.DefaultPath(cookieApplicationName, cookieSourceName, cookieCacheProfile)
	if err != nil {
		return nil, err
	}
	return (sharedcookies.CachedSource{
		Path:   path,
		Source: BrowserCookieSource{},
	}).Cookies(ctx)
}
