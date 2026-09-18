package cli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/edram/qi/internal/qcc"
)

type statusCookieSourceFunc func(context.Context) ([]*http.Cookie, error)

func (f statusCookieSourceFunc) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	return f(ctx)
}

func TestStatusShowsMissingCookiesWithoutRequestingAuthInfo(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	authRequested := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		authRequested = true
	}))
	defer server.Close()

	api := qcc.New(qcc.Options{
		BaseURL:   server.URL,
		PID:       "pid",
		TID:       "tid",
		UserAgent: qccTestUserAgent,
		CookieSource: statusCookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return nil, qcc.ErrNoCookies
		}),
	})
	var output bytes.Buffer
	command := StatusCmd{
		profile: "work",
		cookies: &statusCookies{qcc: api},
		qcc:     &statusQCC{api: api},
		aiqicha: statusAiqicha{},
		output:  &output,
	}
	if err := command.Run(); err != nil {
		t.Fatal(err)
	}
	if authRequested {
		t.Fatal("status requested auth info without cookies")
	}
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	appPath := filepath.Join(userConfigDir, "qc")
	if got, want := output.String(), "Overview\n  Profile: work\n  App Path: "+appPath+"\n\nCookies\n  QCC: 0\n\nAccounts\n  QCC\n    Status: not authenticated\n  AIQICHA\n    Status: not implemented\n"; got != want {
		t.Fatalf("status output = %q, want %q", got, want)
	}
}

func TestStatusShowsCookieMetadataAndCurrentUser(t *testing.T) {
	expires := time.Date(2026, time.September, 17, 8, 30, 0, 0, time.UTC)
	cookieSource := statusCookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
		return []*http.Cookie{
			{Name: "qcc_did", Value: "device-secret", Domain: ".qcc.com", Path: "/", Expires: expires, Secure: true},
			{Name: "QCCSESSID", Value: "session-secret", Domain: ".qcc.com", Path: "/", HttpOnly: true},
		}, nil
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"userEmail":"test@example.com","userNickname":"测试用户","userPhone":"13800138000","userPhone_prefix":"+86"}}`))
	}))
	defer server.Close()

	api := qcc.New(qcc.Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    qccTestUserAgent,
		CookieSource: cookieSource,
	})
	var output bytes.Buffer
	command := StatusCmd{
		cookies: &statusCookies{qcc: api},
		qcc:     &statusQCC{api: api},
		aiqicha: statusAiqicha{},
		output:  &output,
	}
	if err := command.Run(); err != nil {
		t.Fatal(err)
	}

	got := output.String()
	for _, want := range []string{
		"Cookies\n",
		"QCC: 2",
		"Accounts\n  QCC\n",
		"qcc_did: domain=.qcc.com path=/ expires=2026-09-17T08:30:00Z secure=true http_only=false",
		"QCCSESSID: domain=.qcc.com path=/ expires=session secure=false http_only=true",
		"Status: authenticated",
		"User: 测试用户",
		"Phone: +86 138****8000",
		"Email: t***@example.com",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("status output does not contain %q:\n%s", want, got)
		}
	}
	for _, secret := range []string{"device-secret", "session-secret", "13800138000", "test@example.com"} {
		if strings.Contains(got, secret) {
			t.Errorf("status output leaks %q:\n%s", secret, got)
		}
	}
}

func TestStatusShowsCookiesWhenAuthenticationFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "login required", http.StatusUnauthorized)
	}))
	defer server.Close()

	api := qcc.New(qcc.Options{
		BaseURL:   server.URL,
		PID:       "pid",
		TID:       "tid",
		UserAgent: qccTestUserAgent,
		CookieSource: statusCookieSourceFunc(func(context.Context) ([]*http.Cookie, error) {
			return []*http.Cookie{{Name: "QCCSESSID", Value: "expired", Domain: ".qcc.com", Path: "/"}}, nil
		}),
	})
	var output bytes.Buffer
	command := StatusCmd{
		cookies: &statusCookies{qcc: api},
		qcc:     &statusQCC{api: api},
		aiqicha: statusAiqicha{},
		output:  &output,
	}
	if err := command.Run(); err == nil {
		t.Fatal("Run() error = nil, want authentication error")
	}
	got := output.String()
	for _, want := range []string{"QCC: 1", "QCCSESSID", "Status: not authenticated"} {
		if !strings.Contains(got, want) {
			t.Errorf("status output does not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "expired") {
		t.Fatalf("status output leaks cookie value:\n%s", got)
	}
}
