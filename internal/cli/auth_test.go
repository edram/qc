package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	sharedcookies "github.com/edram/qi/internal/cookies"
	"github.com/edram/qi/internal/qcc"
	"github.com/steipete/sweetcookie"
)

func TestAuthImportSeparatesTargetAndBrowserProfiles(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("HOME", configDir)
	t.Setenv("USERPROFILE", configDir)
	t.Setenv("QC_HOME", "")
	cachePath := filepath.Join(configDir, ".qc", "cookies", "qcc.work.json")
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
	t.Setenv("HOME", configDir)
	t.Setenv("USERPROFILE", configDir)
	t.Setenv("QC_HOME", "")

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

func TestAuthImportAiqichaKeepsProviderAndProfileIsolated(t *testing.T) {
	t.Setenv("QC_HOME", t.TempDir())
	qccPath, err := sharedcookies.Path("qcc", "work")
	if err != nil {
		t.Fatal(err)
	}
	if err := sharedcookies.Write(qccPath, []*http.Cookie{{Name: "QCCSESSID", Value: "existing"}}); err != nil {
		t.Fatal(err)
	}
	restore := sharedcookies.SetReadCookies(func(_ context.Context, options sweetcookie.Options) (sweetcookie.Result, error) {
		if options.URL != "https://www.aiqicha.com" || len(options.Browsers) != 1 || options.Browsers[0] != sweetcookie.BrowserEdge || options.Profiles[sweetcookie.BrowserEdge] != "Profile 2" {
			t.Errorf("browser options = %#v", options)
		}
		return sweetcookie.Result{Cookies: []sweetcookie.Cookie{{Name: "BDUSS", Value: "test-secret", Domain: ".aiqicha.com", Path: "/"}}}, nil
	})
	t.Cleanup(restore)
	var stdout, stderr bytes.Buffer
	if code := Execute([]string{"--profile", "work", "auth", "import", "--provider", "aiqicha", "--browser", "edge", "--browser-profile", "Profile 2"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Execute() = %d: %s", code, stderr.String())
	}
	path, err := sharedcookies.Path("aiqicha", "work")
	if err != nil {
		t.Fatal(err)
	}
	got, err := (sharedcookies.CachedSource{Path: path}).Cookies(context.Background())
	if err != nil || len(got) != 1 || got[0].Name != "BDUSS" || got[0].Value != "test-secret" {
		t.Fatalf("aiqicha cache not populated: count %d, error %v", len(got), err)
	}
	previous, err := (sharedcookies.CachedSource{Path: qccPath}).Cookies(context.Background())
	if err != nil || len(previous) != 1 || previous[0].Value != "existing" {
		t.Fatalf("qcc cache changed: %v", err)
	}
	if strings.Contains(stdout.String()+stderr.String(), "test-secret") {
		t.Fatal("cookie value leaked to output")
	}
}

func TestAuthImportFailureKeepsExistingCookiesAndReportsWarning(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("HOME", configDir)
	t.Setenv("USERPROFILE", configDir)
	t.Setenv("QC_HOME", "")
	cachePath := filepath.Join(configDir, ".qc", "cookies", "qcc.default.json")
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
