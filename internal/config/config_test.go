package config

import (
	"path/filepath"
	"testing"
)

func TestUserAgentIsStoredByProfile(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("HOME", configDir)
	t.Setenv("USERPROFILE", configDir)
	t.Setenv("QC_HOME", "")

	path, err := SetUserAgent("work", "work-user-agent")
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("SetUserAgent() path is empty")
	}
	wantPath := filepath.Join(configDir, ".qc", "config.work.json")
	if path != wantPath {
		t.Fatalf("SetUserAgent() path = %q, want %q", path, wantPath)
	}
	got, err := UserAgent("work")
	if err != nil {
		t.Fatal(err)
	}
	if got != "work-user-agent" {
		t.Fatalf("UserAgent(work) = %q, want %q", got, "work-user-agent")
	}
	if _, err := SetUserAgent("default", "default-user-agent"); err != nil {
		t.Fatal(err)
	}
	got, err = UserAgent("work")
	if err != nil {
		t.Fatal(err)
	}
	if got != "work-user-agent" {
		t.Fatalf("UserAgent(work) after default update = %q, want %q", got, "work-user-agent")
	}
	got, err = UserAgent("default")
	if err != nil {
		t.Fatal(err)
	}
	if got != "default-user-agent" {
		t.Fatalf("UserAgent(default) = %q, want %q", got, "default-user-agent")
	}
}

func TestUserAgentInheritsDefaultProfile(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("HOME", configDir)
	t.Setenv("USERPROFILE", configDir)
	t.Setenv("QC_HOME", "")

	if _, err := SetUserAgent("default", "default-user-agent"); err != nil {
		t.Fatal(err)
	}
	got, err := UserAgent("work")
	if err != nil {
		t.Fatal(err)
	}
	if got != "default-user-agent" {
		t.Fatalf("UserAgent(work) = %q, want inherited %q", got, "default-user-agent")
	}
}
