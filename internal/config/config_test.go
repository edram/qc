package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserAgentIsStoredByProfile(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	path, err := SetUserAgent("work", "work-user-agent")
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("SetUserAgent() path is empty")
	}
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filepath.Join(userConfigDir, "qc", "config.work.json")
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
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

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
