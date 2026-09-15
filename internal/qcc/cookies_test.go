package qcc

import (
	"context"
	"testing"

	"github.com/steipete/sweetcookie"
)

func TestBrowserCookieSource(t *testing.T) {
	previous := readBrowserCookies
	readBrowserCookies = func(_ context.Context, options sweetcookie.Options) (sweetcookie.Result, error) {
		if options.URL != "https://www.qcc.com" {
			t.Fatalf("URL = %q", options.URL)
		}
		if len(options.Browsers) != 1 || options.Browsers[0] != sweetcookie.BrowserChrome {
			t.Fatalf("Browsers = %v", options.Browsers)
		}
		if got := options.Profiles[sweetcookie.BrowserChrome]; got != "Default" {
			t.Fatalf("Chrome profile = %q", got)
		}
		return sweetcookie.Result{Cookies: []sweetcookie.Cookie{{
			Name:   "session",
			Value:  "browser-cookie",
			Domain: ".qcc.com",
			Path:   "/",
			Secure: true,
		}}}, nil
	}
	t.Cleanup(func() { readBrowserCookies = previous })

	cookies, err := (BrowserCookieSource{Profile: "Default"}).Cookies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value != "browser-cookie" {
		t.Fatalf("cookies = %#v", cookies)
	}
}
