package cli

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
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

func TestSearchSources(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []searchSourceName
	}{
		{
			name: "default",
			args: []string{"search", "ents", "百度"},
			want: []searchSourceName{searchSourceQCC, searchSourceAiqicha},
		},
		{
			name: "single source",
			args: []string{"search", "ents", "百度", "--source", "qcc"},
			want: []searchSourceName{searchSourceQCC},
		},
		{
			name: "repeated flag",
			args: []string{"search", "ents", "百度", "--source", "qcc", "--source", "aiqicha"},
			want: []searchSourceName{searchSourceQCC, searchSourceAiqicha},
		},
		{
			name: "comma separated",
			args: []string{"search", "pers", "李彦宏", "--source", "qcc,aiqicha"},
			want: []searchSourceName{searchSourceQCC, searchSourceAiqicha},
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
			if got := command.Search.Sources; !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Sources = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchRejectsUnknownSource(t *testing.T) {
	command := New()
	parser, err := kong.New(command, kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := parser.Parse([]string{"search", "ents", "百度", "--source", "unknown"}); err == nil {
		t.Fatal("Parse() error = nil, want an invalid source error")
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
