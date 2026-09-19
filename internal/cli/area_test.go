package cli

import (
	"bytes"
	"testing"
)

func TestAreaListSearchesLocalCatalog(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"--profile", "../invalid", "area", "list", "--search", "深圳", "--limit", "5"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	want := "[{\"name\":\"深圳市\",\"code\":\"440300\"}]\n"
	if got := stdout.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestAreaListDefaultsToTopLevelAreas(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"--profile", "../invalid", "area", "list", "--limit", "2"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	want := "[{\"name\":\"北京市\",\"code\":\"BJ\"},{\"name\":\"天津市\",\"code\":\"TJ\"}]\n"
	if got := stdout.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestAreaListRejectsNegativeLimit(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"--profile", "../invalid", "area", "list", "--limit=-1"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("Execute() code = %d, want 1", code)
	}
	if got := stderr.String(); got != "area limit must not be negative\n" {
		t.Fatalf("stderr = %q", got)
	}
}

func TestAreaListRejectsUnsupportedProvider(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"--profile", "../invalid", "area", "list", "--provider", "aiqicha"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Execute() code = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected an invalid provider error")
	}
}
