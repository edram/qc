package aiqicha

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/edram/qi/internal/cookies"
	clienthttp "github.com/edram/qi/internal/http"
)

func TestClientSignsRequestsWithEffectiveUserAgent(t *testing.T) {
	for _, userAgent := range []string{"", "test-browser"} {
		t.Run(userAgent, func(t *testing.T) {
			before := time.Now().UnixMilli()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				signature, err := base64.StdEncoding.DecodeString(r.Header.Get("aqc_ssr"))
				if err != nil {
					t.Error(err)
					return
				}
				stamp, ua, ok := strings.Cut(string(signature), "cal110")
				if !ok || ua != r.UserAgent() || ua == "" {
					t.Errorf("signature User-Agent = %q, request User-Agent = %q", ua, r.UserAgent())
				}
				milliseconds, err := strconv.ParseInt(stamp, 10, 64)
				if err != nil || milliseconds < before || milliseconds > time.Now().UnixMilli() {
					t.Errorf("signature timestamp = %q: %v", stamp, err)
				}
				if userAgent != "" && r.UserAgent() != userAgent {
					t.Errorf("User-Agent = %q", r.UserAgent())
				}
				fmt.Fprint(w, `{"status":0,"data":{}}`)
			}))
			defer server.Close()
			client := New(Options{BaseURL: server.URL, UserAgent: userAgent, CookieSource: testCookieSource{}})
			var response any
			if err := client.GetJSON(context.Background(), "/companies", &response); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestClientGetAndPostWithoutUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Referer(); got != "" {
			t.Errorf("generic request inherited a search Referer: %q", got)
		}
		if r.Method == http.MethodGet {
			w.Write([]byte("get"))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if r.Method != http.MethodPost || string(body) != "request body" {
			t.Fatalf("request = %s %q, want POST %q", r.Method, body, "request body")
		}
		w.Write([]byte("post"))
	}))
	defer server.Close()

	client := New(Options{BaseURL: server.URL, CookieSource: testCookieSource{}})
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

type testCookieSource struct{}

func (testCookieSource) Cookies(context.Context) ([]*http.Cookie, error) {
	return []*http.Cookie{{Name: "BDUSS", Value: "test", Path: "/"}}, nil
}

func TestClientUsesProfileSessionAndRequestHeaders(t *testing.T) {
	t.Setenv("QC_HOME", t.TempDir())
	path, err := cookies.Path("aiqicha", "work")
	if err != nil {
		t.Fatal(err)
	}
	if err := cookies.Write(path, []*http.Cookie{{Name: "BDUSS", Value: "work-session", Path: "/"}}); err != nil {
		t.Fatal(err)
	}
	var origin string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "matching-browser" {
			t.Errorf("User-Agent = %q", got)
		}
		if cookie, err := r.Cookie("BDUSS"); err != nil || cookie.Value != "work-session" {
			t.Errorf("session cookie missing: %v", err)
		}
		if got := r.Header.Get("X-Requested-With"); got != "XMLHttpRequest" {
			t.Errorf("X-Requested-With = %q", got)
		}
		if got := r.Header.Get("Accept"); !strings.Contains(got, "application/json") {
			t.Errorf("Accept = %q", got)
		}
		if got := r.Referer(); got != origin+"/companies" {
			t.Errorf("Referer = %q", got)
		}
		fmt.Fprint(w, `{"status":0,"data":{"totalNum":0,"resultList":[]}}`)
	}))
	defer server.Close()
	origin = server.URL
	client := New(Options{BaseURL: origin, Profile: "work", UserAgent: "matching-browser"})
	var response any
	if err := client.GetJSON(context.Background(), "/companies", &response, clienthttp.WithRequestHeader("Referer", origin+"/companies")); err != nil {
		t.Fatal(err)
	}
}

func TestClientKeepsSessionWithinOrigin(t *testing.T) {
	forwarded := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded = true
	}))
	defer target.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer origin.Close()
	provided := &http.Client{}
	client := New(Options{BaseURL: origin.URL, HTTPClient: provided, UserAgent: "test-browser", CookieSource: testCookieSource{}})
	var result any
	if err := client.GetJSON(context.Background(), "/s/advanceFilterAjax", &result); err == nil || !strings.Contains(err.Error(), "HTTP 302") {
		t.Fatalf("redirect error = %v", err)
	}
	if err := client.GetJSON(context.Background(), target.URL, &result); err == nil || !strings.Contains(err.Error(), "outside configured origin") {
		t.Fatalf("foreign-origin error = %v", err)
	}
	if forwarded {
		t.Fatal("session request reached a different origin")
	}
	if provided.CheckRedirect != nil || provided.Timeout != 0 {
		t.Fatal("caller-owned client was modified")
	}
}

func TestClientRequiresMatchingCookies(t *testing.T) {
	for _, tt := range []struct {
		name    string
		cookies []*http.Cookie
		want    string
	}{
		{"cookies", nil, "no matching cookies"},
		{"foreign cookies", []*http.Cookie{{Name: "BDUSS", Value: "test", Domain: "example.test", Path: "/"}}, "no matching cookies"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("request sent without a matching browser session")
			}))
			defer server.Close()
			client := New(Options{BaseURL: server.URL, CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) { return tt.cookies, nil })})
			var result any
			if err := client.GetJSON(context.Background(), "/s/advanceFilterAjax", &result); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %s", err, tt.want)
			}
		})
	}
}

type cookieSourceFunc func(context.Context) ([]*http.Cookie, error)

func (source cookieSourceFunc) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	return source(ctx)
}

func TestClientRejectsHTTPFailuresAndClosesBody(t *testing.T) {
	for _, method := range []string{"GET", "POST", "GetJSON"} {
		t.Run(method, func(t *testing.T) {
			body := &trackedResponseBody{Reader: strings.NewReader("forbidden")}
			client := New(Options{
				UserAgent:    "test-browser",
				CookieSource: testCookieSource{},
				HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusForbidden, Body: body, Request: request}, nil
				})},
			})
			var response *http.Response
			var err error
			switch method {
			case "GET":
				response, err = client.Get(context.Background(), "/companies")
			case "POST":
				response, err = client.Post(context.Background(), "/companies", strings.NewReader("request body"))
			case "GetJSON":
				var result any
				err = client.GetJSON(context.Background(), "/companies", &result)
			}
			if err == nil || !strings.Contains(err.Error(), "aiqicha: HTTP 403") {
				t.Errorf("error = %v, want HTTP 403", err)
			}
			if !body.closed {
				t.Error("HTTP error response body was not closed")
			}
			if response != nil {
				response.Body.Close()
				t.Error("HTTP failure returned a response instead of an error")
			}
		})
	}
}

type trackedResponseBody struct {
	io.Reader
	closed bool
}

func (body *trackedResponseBody) Close() error {
	body.closed = true
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}
