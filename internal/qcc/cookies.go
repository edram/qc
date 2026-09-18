package qcc

import (
	"context"
	"errors"
	"net/http"
	"os"

	sharedcookies "github.com/edram/qi/internal/cookies"
	"github.com/steipete/sweetcookie"
)

const (
	cookieSourceName = "qcc"
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

type defaultCookieSource struct {
	profile string
}

// cookieCacheClearer is implemented only by sources that own a removable local cache.
type cookieCacheClearer interface {
	clearCache() error
}

func newCookieSource(profile string) CookieSource {
	return defaultCookieSource{profile: profile}
}

func (s defaultCookieSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	path, err := cookieCachePath(s.profile)
	if err != nil {
		return nil, err
	}
	return (sharedcookies.CachedSource{
		Path:   path,
		Source: BrowserCookieSource{},
	}).Cookies(ctx)
}

func (s defaultCookieSource) clearCache() error {
	path, err := cookieCachePath(s.profile)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// ImportBrowserCookies replaces the selected profile's cache with cookies from a browser profile.
func ImportBrowserCookies(ctx context.Context, profile string, browser sweetcookie.Browser, browserProfile string) (int, string, error) {
	cookies, err := (BrowserCookieSource{
		Browser: browser,
		Profile: browserProfile,
	}).Cookies(ctx)
	if err != nil {
		return 0, "", err
	}
	path, err := cookieCachePath(profile)
	if err != nil {
		return 0, "", err
	}
	if err := sharedcookies.Write(path, cookies); err != nil {
		return 0, "", err
	}
	return len(cookies), path, nil
}

func cookieCachePath(profile string) (string, error) {
	return sharedcookies.Path(cookieSourceName, profile)
}
