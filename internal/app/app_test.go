package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirReturnsApplicationDirectory(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(userConfigDir, "qc")

	got, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Dir() = %q, want %q", got, want)
	}
}

func TestPathJoinsApplicationRelativeElements(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	appDir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(appDir, "cookies", "qcc.work.json")

	got, err := Path("cookies", "qcc.work.json")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}
