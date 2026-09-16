package qcc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestClientReturnsCookieInfoAndAuthInfo(t *testing.T) {
	expires := time.Date(2026, time.September, 17, 8, 30, 0, 0, time.UTC)
	cookieSource := cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
		return []*http.Cookie{
			{Name: "qcc_did", Value: "device-secret", Domain: ".qcc.com", Path: "/", Expires: expires, Secure: true},
			{Name: "QCCSESSID", Value: "session-secret", Domain: ".qcc.com", Path: "/", HttpOnly: true},
		}, nil
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/userCenter/getAuthInfo" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if cookie, err := r.Cookie("QCCSESSID"); err != nil || cookie.Value != "session-secret" {
			t.Fatalf("QCCSESSID = %v, %v", cookie, err)
		}
		_, _ = w.Write([]byte(`{"data":{"userAvatar":"https://image.qcc.com/avatar.png","userEmail":"test@example.com","userNickname":"测试用户","userPhone":"13800138000","userPhone_prefix":"+86"}}`))
	}))
	defer server.Close()

	client := New(Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    testUserAgent,
		CookieSource: cookieSource,
	})
	cookies, err := client.CookieInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	wantCookies := []CookieInfo{
		{Name: "qcc_did", Domain: ".qcc.com", Path: "/", Expires: expires, Secure: true},
		{Name: "QCCSESSID", Domain: ".qcc.com", Path: "/", HTTPOnly: true},
	}
	if !reflect.DeepEqual(cookies, wantCookies) {
		t.Fatalf("CookieInfo() = %#v, want %#v", cookies, wantCookies)
	}

	auth, err := client.AuthInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	wantAuth := AuthInfo{
		Nickname:    "测试用户",
		PhonePrefix: "+86",
		Phone:       "13800138000",
		Email:       "test@example.com",
		AvatarURL:   "https://image.qcc.com/avatar.png",
	}
	if !reflect.DeepEqual(auth, wantAuth) {
		t.Fatalf("AuthInfo() = %#v, want %#v", auth, wantAuth)
	}
}

func TestClientAuthInfoRejectsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "login required", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := New(Options{
		BaseURL:   server.URL,
		PID:       "pid",
		TID:       "tid",
		UserAgent: testUserAgent,
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return []*http.Cookie{{Name: "QCCSESSID", Value: "expired"}}, nil
		}),
	})
	_, err := client.AuthInfo(context.Background())
	if err == nil || !strings.Contains(err.Error(), "401 Unauthorized") {
		t.Fatalf("AuthInfo() error = %v, want HTTP status", err)
	}
}

func TestClientAuthInfoRejectsMissingAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	client := New(Options{
		BaseURL:   server.URL,
		PID:       "pid",
		TID:       "tid",
		UserAgent: testUserAgent,
		CookieSource: cookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return []*http.Cookie{{Name: "QCCSESSID", Value: "expired"}}, nil
		}),
	})
	_, err := client.AuthInfo(context.Background())
	if err == nil || !strings.Contains(err.Error(), "not authenticated") {
		t.Fatalf("AuthInfo() error = %v, want missing account error", err)
	}
}
