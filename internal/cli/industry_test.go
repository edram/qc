package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

func TestIndustryListCommandPath(t *testing.T) {
	parser, err := kong.New(New(), kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := parser.Parse([]string{"industry", "list", "--search", "软件"})
	if err != nil {
		t.Fatal(err)
	}
	if got := ctx.Command(); got != "industry list" {
		t.Fatalf("Command() = %q, want %q", got, "industry list")
	}
}

func TestIndustryListWritesJSON(t *testing.T) {
	command := New()
	parser, err := kong.New(command, kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := parser.Parse([]string{"industry", "list", "--search", "软件", "--limit", "5"})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command.Industry.List.output = &output
	if err := ctx.Run(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"name":"软件和信息技术服务业"`) {
		t.Fatalf("output = %q", output.String())
	}
	if !strings.Contains(output.String(), `"code"`) {
		t.Fatalf("output does not contain provider code: %q", output.String())
	}
	if strings.Contains(output.String(), `"id"`) || strings.Contains(output.String(), `"path"`) {
		t.Fatalf("output contains unnecessary metadata: %q", output.String())
	}
}
