package cli

import (
	"bytes"
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

func TestSearchCommandsAreWired(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "enterprises", args: []string{"search", "ents", "百度"}, want: "qc search ents is not implemented yet"},
		{name: "people", args: []string{"search", "pers", "李彦宏"}, want: "qc search pers is not implemented yet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := Execute(tt.args, &stdout, &stderr)

			if code != 1 {
				t.Fatalf("Execute() code = %d, want 1", code)
			}
			if !strings.Contains(stderr.String(), tt.want) {
				t.Fatalf("stderr = %q, want it to contain %q", stderr.String(), tt.want)
			}
		})
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
