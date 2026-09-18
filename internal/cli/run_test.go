package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	appconfig "github.com/edram/qi/internal/config"
)

func TestHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "search") {
		t.Fatalf("help output does not contain search command:\n%s", stdout.String())
	}
}

func TestVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"--version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != "dev" {
		t.Fatalf("version output = %q, want %q", got, "dev")
	}
}

func TestStatusCommandPath(t *testing.T) {
	parser, err := kong.New(New(), kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := parser.Parse([]string{"status"})
	if err != nil {
		t.Fatal(err)
	}
	if got := ctx.Command(); got != "status" {
		t.Fatalf("Command() = %q, want %q", got, "status")
	}
}

func TestProfileFlag(t *testing.T) {
	command := New()
	parser, err := kong.New(command, kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse([]string{"--profile", "work", "status"}); err != nil {
		t.Fatal(err)
	}
	if command.Profile != "work" {
		t.Fatalf("Profile = %q, want %q", command.Profile, "work")
	}
}

func TestUserAgentFlag(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	if _, err := appconfig.SetUserAgent("default", "persisted-user-agent"); err != nil {
		t.Fatal(err)
	}

	command := New()
	parser, err := kong.New(command, kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse([]string{"--user-agent", "browser-user-agent", "status"}); err != nil {
		t.Fatal(err)
	}
	if err := command.loadProfileConfig(); err != nil {
		t.Fatal(err)
	}
	if command.UserAgent != "browser-user-agent" {
		t.Fatalf("UserAgent = %q, want %q", command.UserAgent, "browser-user-agent")
	}
}

func TestConfigSetUserAgentForProfile(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"--profile", "work", "config", "set", "user-agent", "browser-user-agent"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	userAgent, err := appconfig.UserAgent("work")
	if err != nil {
		t.Fatal(err)
	}
	if userAgent != "browser-user-agent" {
		t.Fatalf("UserAgent(work) = %q, want %q", userAgent, "browser-user-agent")
	}
	if !strings.Contains(stdout.String(), "profile work") {
		t.Fatalf("stdout = %q, want profile confirmation", stdout.String())
	}
}

func TestCLIUsesPersistedProfileUserAgent(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	if _, err := appconfig.SetUserAgent("work", "persisted-user-agent"); err != nil {
		t.Fatal(err)
	}

	command := New()
	command.Profile = "work"
	if err := command.loadProfileConfig(); err != nil {
		t.Fatal(err)
	}
	if command.UserAgent != "persisted-user-agent" {
		t.Fatalf("UserAgent = %q, want %q", command.UserAgent, "persisted-user-agent")
	}
}

func TestCLIConfiguresSelectedProfile(t *testing.T) {
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

	command := New()
	parser, err := kong.New(command, kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse([]string{"--profile", "work", "status"}); err != nil {
		t.Fatal(err)
	}
	command.configureProfile()

	if command.Status.profile != "work" {
		t.Fatalf("status profile = %q, want %q", command.Status.profile, "work")
	}
	statusCookies, err := command.Status.cookies.Cookies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(statusCookies.Sources) != 1 || len(statusCookies.Sources[0].Cookies) != 1 || statusCookies.Sources[0].Cookies[0].Name != "session" {
		t.Fatalf("status cookies = %#v", statusCookies)
	}
	searcher, ok := command.Search.qcc.(*searchQCC)
	if !ok {
		t.Fatalf("QCC searcher = %T", command.Search.qcc)
	}
	searchCookies, err := searcher.api.CookieInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(searchCookies) != 1 || searchCookies[0].Name != "session" {
		t.Fatalf("search cookies = %#v", searchCookies)
	}
}

func TestSearchHelpListsSubcommands(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"search", "--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	for _, subcommand := range []string{"ents", "pers"} {
		if !strings.Contains(stdout.String(), subcommand) {
			t.Errorf("search help does not contain %q:\n%s", subcommand, stdout.String())
		}
	}
}

func TestSearchCommandPaths(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"search", "ents", "百度"}, want: "search ents <query>"},
		{args: []string{"search", "pers", "李彦宏"}, want: "search pers <query>"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			parser, err := kong.New(New(), kong.Name("qc"), kong.Exit(func(int) {}))
			if err != nil {
				t.Fatal(err)
			}
			ctx, err := parser.Parse(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if got := ctx.Command(); got != tt.want {
				t.Fatalf("Command() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSearchProviders(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []searchProviderName
	}{
		{
			name: "default",
			args: []string{"search", "ents", "百度"},
			want: []searchProviderName{searchProviderQCC, searchProviderAiqicha},
		},
		{
			name: "single provider",
			args: []string{"search", "ents", "百度", "--provider", "qcc"},
			want: []searchProviderName{searchProviderQCC},
		},
		{
			name: "repeated flag",
			args: []string{"search", "ents", "百度", "--provider", "qcc", "--provider", "aiqicha"},
			want: []searchProviderName{searchProviderQCC, searchProviderAiqicha},
		},
		{
			name: "comma separated",
			args: []string{"search", "pers", "李彦宏", "--provider", "qcc,aiqicha"},
			want: []searchProviderName{searchProviderQCC, searchProviderAiqicha},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := New()
			parser, err := kong.New(command, kong.Name("qc"), kong.Exit(func(int) {}))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parser.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if got := command.Search.Providers; !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Providers = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchRejectsUnknownProvider(t *testing.T) {
	command := New()
	parser, err := kong.New(command, kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := parser.Parse([]string{"search", "ents", "百度", "--provider", "unknown"}); err == nil {
		t.Fatal("Parse() error = nil, want an invalid provider error")
	}
}

func TestSearchRequiresSubcommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"search", "百度"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Execute() code = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected a parse error")
	}
}

func TestSearchQueryIsRequired(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"search", "ents"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Execute() code = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected a missing query error")
	}
}
