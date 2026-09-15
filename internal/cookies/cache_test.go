package cookies

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

type sourceFunc func(context.Context) ([]*http.Cookie, error)

func (f sourceFunc) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	return f(ctx)
}

func TestCachedSourceAcceptsExpiryOutsideJSONRange(t *testing.T) {
	browserReads := 0
	source := CachedSource{
		Path: filepath.Join(t.TempDir(), "cookies.json"),
		Source: sourceFunc(func(context.Context) ([]*http.Cookie, error) {
			browserReads++
			return []*http.Cookie{{
				Name:    "session",
				Value:   "browser-cookie",
				Expires: time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC),
			}}, nil
		}),
	}

	if _, err := source.Cookies(context.Background()); err != nil {
		t.Fatal(err)
	}
	cookies, err := source.Cookies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if browserReads != 1 {
		t.Fatalf("browser reads = %d, want 1", browserReads)
	}
	if len(cookies) != 1 || !cookies[0].Expires.IsZero() {
		t.Fatalf("cached cookies = %#v", cookies)
	}
}
