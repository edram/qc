package qcc

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchPIDAndTID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<script>
window.pid = 'pid-from-base-url';
window.tid = 'tid-from-base-url';
</script>`))
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return nil, nil
		}),
	})
	pid, tid, err := client.FetchPIDAndTID(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if pid != "pid-from-base-url" || tid != "tid-from-base-url" {
		t.Fatalf("pid, tid = %q, %q", pid, tid)
	}
}

func TestClientFetchesPIDAndTIDWhenMissing(t *testing.T) {
	baseRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			baseRequests++
			if r.Method != http.MethodGet {
				t.Errorf("method = %s, want GET", r.Method)
			}
			if got := r.Header.Get("User-Agent"); got != defaultUserAgent {
				t.Errorf("User-Agent = %q, want %q", got, defaultUserAgent)
			}
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value != "browser-cookie" {
				t.Errorf("session cookie = %v, %v", cookie, err)
			}
			_, _ = w.Write([]byte(`<script>
window.pid = '2431b49cb68c42989a96d790fe21bb7f';
window.tid = '60b78b88dde384dbb6a24cb6c09c4656';
</script>`))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		sign := GenerateSign(r.URL.Path, string(body), "60b78b88dde384dbb6a24cb6c09c4656")
		if got := r.Header.Get("X-Pid"); got != "2431b49cb68c42989a96d790fe21bb7f" {
			t.Errorf("X-Pid = %q", got)
		}
		if got := r.Header.Get(sign.HeaderName); got != sign.HeaderValue {
			t.Errorf("%s = %q, want %q", sign.HeaderName, got, sign.HeaderValue)
		}
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return []*http.Cookie{{Name: "session", Value: "browser-cookie"}}, nil
		}),
	})

	for range 2 {
		response, err := client.Post(context.Background(), "/api/search", strings.NewReader(`{"query":"test"}`))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
	}
	if baseRequests != 1 {
		t.Fatalf("base requests = %d, want 1", baseRequests)
	}
}

func TestClientRequiresBothIdentifiers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Errorf("unexpected API request to %s", r.URL.Path)
			return
		}
		_, _ = w.Write([]byte(`<script>window.pid = 'pid-only';</script>`))
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return nil, nil
		}),
	})

	response, err := client.Post(context.Background(), "/api/search", strings.NewReader(`{"query":"test"}`))
	if response != nil {
		response.Body.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "tid") {
		t.Fatalf("error = %v, want missing tid error", err)
	}
}
