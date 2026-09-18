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

const testUserAgent = "Mozilla/5.0 test browser"

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
		if got := r.Header.Get("User-Agent"); got != testUserAgent {
			t.Fatalf("User-Agent = %q, want %q", got, testUserAgent)
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
		BaseURL:   server.URL,
		TID:       "tid-123",
		PID:       "pid-123",
		UserAgent: testUserAgent,
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

func TestClientDoesNotFollowRedirects(t *testing.T) {
	redirectedRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirected" {
			redirectedRequests++
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/redirected", http.StatusFound)
	}))
	defer server.Close()

	client := New(Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    testUserAgent,
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) { return nil, nil }),
	})
	response, err := client.Post(context.Background(), "/api/search/searchMulti", strings.NewReader(`{}`))
	if response != nil {
		response.Body.Close()
		t.Fatal("Post() response is non-nil for redirect error")
	}
	if err == nil || err.Error() != "qcc: redirect to /redirected" {
		t.Fatalf("Post() error = %v, want redirect location", err)
	}
	if redirectedRequests != 0 {
		t.Fatalf("redirected requests = %d, want 0", redirectedRequests)
	}
}

func TestClientReportsLoginRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/weblogin?back=%2Fcompanies", http.StatusFound)
	}))
	defer server.Close()

	client := New(Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    testUserAgent,
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) { return nil, nil }),
	})
	response, err := client.Get(context.Background(), "/companies")
	if response != nil {
		response.Body.Close()
		t.Fatal("Get() response is non-nil for login redirect")
	}
	if err == nil || err.Error() != "qcc: login required; log in to QCC again" {
		t.Fatalf("Get() error = %v, want login guidance", err)
	}
}

func TestClientClearsCookieCacheForUserAgentMismatchRedirect(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(userConfigDir, "qc", "cookies", "qcc.work.json")
	if err := sharedcookies.Write(cachePath, []*http.Cookie{{Name: "QCCSESSID", Value: "stale"}}); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://www.qcc.com/405.html", http.StatusFound)
	}))
	defer server.Close()

	client := New(Options{
		BaseURL:   server.URL,
		PID:       "pid",
		TID:       "tid",
		Profile:   "work",
		UserAgent: testUserAgent,
	})
	response, err := client.Get(context.Background(), "/companies")
	if response != nil {
		response.Body.Close()
		t.Fatal("Get() response is non-nil for User-Agent mismatch redirect")
	}
	if err == nil {
		t.Fatal("Get() error = nil, want User-Agent mismatch guidance")
	}
	for _, want := range []string{"User-Agent may not match", "auth import", "cached cookies were deleted"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Get() error = %q, want %q", err, want)
		}
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("cookie cache still exists at %q: %v", cachePath, err)
	}
}

func TestClientRejectsNonOKResponseWithHTMLTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>访问超频 - 企查查</title>
</head>
</html>`))
	}))
	defer server.Close()

	client := New(Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    testUserAgent,
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) { return nil, nil }),
	})
	response, err := client.Get(context.Background(), "/companies")
	if response != nil {
		response.Body.Close()
		t.Fatal("Get() response is non-nil for HTTP error")
	}
	if err == nil || err.Error() != "qcc: HTTP 429: 访问超频 - 企查查" {
		t.Fatalf("Get() error = %v, want status and HTML title", err)
	}
}

func TestClientRequiresUserAgentBeforeRequest(t *testing.T) {
	requested := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requested = true
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
		PID:     "pid",
		TID:     "tid",
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return nil, nil
		}),
	})
	response, err := client.Get(context.Background(), "/companies")
	if response != nil {
		response.Body.Close()
	}
	if err == nil {
		t.Fatal("Get() error = nil, want missing User-Agent error")
	}
	for _, want := range []string{"QCC may invalidate cookies", "--user-agent", "--profile <name>", "config set user-agent"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Get() error = %q, want %q", err, want)
		}
	}
	if requested {
		t.Fatal("QCC request sent without configured User-Agent")
	}
}

type cookieSourceFunc func(context.Context) ([]*http.Cookie, error)

func (f cookieSourceFunc) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	return f(ctx)
}

func TestNewUsesDefaultBaseURL(t *testing.T) {
	client := New(Options{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if got := request.URL.String(); got != "https://www.qcc.com/companies" {
				t.Fatalf("URL = %q, want %q", got, "https://www.qcc.com/companies")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("ok")),
				Request:    request,
			}, nil
		})},
		PID:       "pid",
		TID:       "tid",
		UserAgent: testUserAgent,
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return nil, nil
		}),
	})
	response, err := client.Get(context.Background(), "/companies")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
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
		response, err := New(Options{BaseURL: server.URL, PID: "pid", TID: "tid", UserAgent: testUserAgent}).Get(context.Background(), "/companies")
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

	response, err := New(Options{BaseURL: server.URL, PID: "pid", TID: "tid", UserAgent: testUserAgent}).Get(context.Background(), "/companies")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if browserReads != 1 {
		t.Fatalf("browser reads = %d, want 1", browserReads)
	}
}

func TestClientUsesSelectedCookieProfile(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(userConfigDir, "qc", "cookies", "qcc.work.json")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cachePath, []byte(`[{"name":"session","value":"work-cookie","domain":".qcc.com","path":"/"}]`), 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "work-cookie" {
			t.Errorf("session cookie = %v, %v", cookie, err)
		}
	}))
	defer server.Close()

	response, err := New(Options{
		BaseURL:   server.URL,
		PID:       "pid",
		TID:       "tid",
		Profile:   "work",
		UserAgent: testUserAgent,
	}).Get(context.Background(), "/companies")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
}
