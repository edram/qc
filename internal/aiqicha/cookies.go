package aiqicha

import (
	"context"
	"net/http"

	"github.com/edram/qi/internal/cookies"
	"github.com/steipete/sweetcookie"
)

type defaultCookieSource struct {
	profile string
}

func (source defaultCookieSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	path, err := cookies.Path("aiqicha", source.profile)
	if err != nil {
		return nil, err
	}
	return (cookies.CachedSource{Path: path, Source: cookies.BrowserSource{URL: defaultBaseURL}}).Cookies(ctx)
}

// ImportBrowserCookies replaces only the Aiqicha cache for the selected qc profile.
func ImportBrowserCookies(ctx context.Context, profile string, browser sweetcookie.Browser, browserProfile string) (int, string, error) {
	path, err := cookies.Path("aiqicha", profile)
	if err != nil {
		return 0, "", err
	}
	items, err := (cookies.BrowserSource{URL: defaultBaseURL, Browser: browser, Profile: browserProfile}).Cookies(ctx)
	if err != nil {
		return 0, "", err
	}
	if err := cookies.Write(path, items); err != nil {
		return 0, "", err
	}
	return len(items), path, nil
}
