package qcc

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sharedcookies "github.com/edram/qi/internal/cookies"
	"github.com/steipete/sweetcookie"
)

func TestClientGetAndPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		sign := GenerateSign(r.URL.Path, string(body), "tid-123")
		if got := r.Header.Get("X-Pid"); got != "pid-123" {
			t.Fatalf("X-Pid = %q, want %q", got, "pid-123")
		}
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "browser-cookie" {
			t.Fatalf("session cookie = %v, %v", cookie, err)
		}
		if got := r.Header.Get("User-Agent"); got != defaultUserAgent {
			t.Fatalf("User-Agent = %q, want %q", got, defaultUserAgent)
		}
		if got := r.Header.Get(sign.HeaderName); got != sign.HeaderValue {
			t.Fatalf("%s = %q, want %q", sign.HeaderName, got, sign.HeaderValue)
		}

		if r.Method == http.MethodGet {
			w.Write([]byte("get"))
			return
		}

		if r.Method != http.MethodPost || string(body) != "request body" {
			t.Fatalf("request = %s %q, want POST %q", r.Method, body, "request body")
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want %q", got, "application/json")
		}
		w.Write([]byte("post"))
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
		TID:     "tid-123",
		PID:     "pid-123",
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return []*http.Cookie{{Name: "session", Value: "browser-cookie"}}, nil
		}),
	})
	for _, tt := range []struct {
		name string
		do   func() (*http.Response, error)
		want string
	}{
		{name: "GET", do: func() (*http.Response, error) { return client.Get(context.Background(), "/companies") }, want: "get"},
		{name: "POST", do: func() (*http.Response, error) {
			return client.Post(context.Background(), "/companies", strings.NewReader("request body"))
		}, want: "post"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response, err := tt.do()
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(body) != tt.want {
				t.Fatalf("response body = %q, want %q", body, tt.want)
			}
		})
	}
}

type cookieSourceFunc func(context.Context) ([]*http.Cookie, error)

func (f cookieSourceFunc) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	return f(ctx)
}

func TestNewUsesDefaultBaseURL(t *testing.T) {
	if got := New(Options{}).baseURL; got != "https://www.qcc.com" {
		t.Fatalf("baseURL = %q, want %q", got, "https://www.qcc.com")
	}
}

func TestClientCachesBrowserCookiesByDefault(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	browserReads := 0
	restore := sharedcookies.SetReadCookies(func(context.Context, sweetcookie.Options) (sweetcookie.Result, error) {
		browserReads++
		return sweetcookie.Result{Cookies: []sweetcookie.Cookie{{
			Name:   "session",
			Value:  "browser-cookie",
			Domain: ".qcc.com",
			Path:   "/",
		}}}, nil
	})
	t.Cleanup(restore)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "browser-cookie" {
			t.Errorf("session cookie = %v, %v", cookie, err)
		}
	}))
	defer server.Close()

	for range 2 {
		response, err := New(Options{BaseURL: server.URL, PID: "pid", TID: "tid"}).Get(context.Background(), "/companies")
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
	}

	if browserReads != 1 {
		t.Fatalf("browser reads = %d, want 1", browserReads)
	}
	cachePath := filepath.Join(userConfigDir, "qc", "cookies", "qcc.default.json")
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("cookie cache %q: %v", cachePath, err)
	}
}

func TestClientRefreshesExpiredCookieCache(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(userConfigDir, "qc", "cookies", "qcc.default.json")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatal(err)
	}
	expiredData := []byte(`[{"name":"session","value":"expired-cookie","expires":"` + time.Now().Add(-time.Hour).Format(time.RFC3339Nano) + `"}]`)
	if err := os.WriteFile(cachePath, expiredData, 0o600); err != nil {
		t.Fatal(err)
	}

	browserReads := 0
	restore := sharedcookies.SetReadCookies(func(context.Context, sweetcookie.Options) (sweetcookie.Result, error) {
		browserReads++
		return sweetcookie.Result{Cookies: []sweetcookie.Cookie{{
			Name:   "session",
			Value:  "fresh-cookie",
			Domain: ".qcc.com",
			Path:   "/",
		}}}, nil
	})
	t.Cleanup(restore)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "fresh-cookie" {
			t.Errorf("session cookie = %v, %v", cookie, err)
		}
	}))
	defer server.Close()

	response, err := New(Options{BaseURL: server.URL, PID: "pid", TID: "tid"}).Get(context.Background(), "/companies")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if browserReads != 1 {
		t.Fatalf("browser reads = %d, want 1", browserReads)
	}
}
