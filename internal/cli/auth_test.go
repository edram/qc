package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sharedcookies "github.com/edram/qi/internal/cookies"
	"github.com/edram/qi/internal/qcc"
	"github.com/steipete/sweetcookie"
)

func TestAuthImportSeparatesTargetAndBrowserProfiles(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(userConfigDir, "qc", "cookies", "qcc.work.json")
	if err := sharedcookies.Write(cachePath, []*http.Cookie{{Name: "stale_cookie", Value: "stale"}}); err != nil {
		t.Fatal(err)
	}

	var options sweetcookie.Options
	restore := sharedcookies.SetReadCookies(func(_ context.Context, got sweetcookie.Options) (sweetcookie.Result, error) {
		options = got
		return sweetcookie.Result{Cookies: []sweetcookie.Cookie{{
			Name:   "QCCSESSID",
			Value:  "session-secret",
			Domain: ".qcc.com",
			Path:   "/",
		}}}, nil
	})
	t.Cleanup(restore)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{
		"--profile", "work",
		"auth", "import",
		"--browser", "chrome",
		"--browser-profile", "Profile 1",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if len(options.Browsers) != 1 || options.Browsers[0] != sweetcookie.BrowserChrome {
		t.Fatalf("Browsers = %v, want [chrome]", options.Browsers)
	}
	if got := options.Profiles[sweetcookie.BrowserChrome]; got != "Profile 1" {
		t.Fatalf("browser profile = %q, want %q", got, "Profile 1")
	}

	if got, want := stdout.String(), fmt.Sprintf("Saved 1 cookies to %s\n", cachePath); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}

	client := qcc.New(qcc.Options{Profile: "work"})
	cookies, err := client.CookieInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cookies) != 1 || cookies[0].Name != "QCCSESSID" {
		t.Fatalf("imported cookies = %#v", cookies)
	}
}

func TestAuthImportDefaultsToChrome(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	var options sweetcookie.Options
	restore := sharedcookies.SetReadCookies(func(_ context.Context, got sweetcookie.Options) (sweetcookie.Result, error) {
		options = got
		return sweetcookie.Result{Cookies: []sweetcookie.Cookie{{Name: "QCCSESSID"}}}, nil
	})
	t.Cleanup(restore)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Execute([]string{"auth", "import"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Execute() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if len(options.Browsers) != 1 || options.Browsers[0] != sweetcookie.BrowserChrome {
		t.Fatalf("Browsers = %v, want [chrome]", options.Browsers)
	}
}

func TestAuthImportFailureKeepsExistingCookiesAndReportsWarning(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(userConfigDir, "qc", "cookies", "qcc.default.json")
	if err := sharedcookies.Write(cachePath, []*http.Cookie{{Name: "QCCSESSID", Value: "existing"}}); err != nil {
		t.Fatal(err)
	}

	restore := sharedcookies.SetReadCookies(func(context.Context, sweetcookie.Options) (sweetcookie.Result, error) {
		return sweetcookie.Result{Warnings: []string{"chrome profile not found"}}, nil
	})
	t.Cleanup(restore)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Execute([]string{"auth", "import", "--browser", "chrome"}, &stdout, &stderr); code != 1 {
		t.Fatalf("Execute() code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "chrome profile not found") {
		t.Fatalf("stderr = %q, want browser warning", stderr.String())
	}

	cookies, err := qcc.New(qcc.Options{}).CookieInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cookies) != 1 || cookies[0].Name != "QCCSESSID" {
		t.Fatalf("cached cookies = %#v, want existing cookie", cookies)
	}
}
