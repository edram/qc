package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirReturnsApplicationDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("QC_HOME", "")
	want := filepath.Join(home, ".qc")

	got, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Dir() = %q, want %q", got, want)
	}
}

func TestDirUsesQCHome(t *testing.T) {
	customDir := filepath.Join(t.TempDir(), "custom-qc")
	t.Setenv("QC_HOME", customDir)

	got, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if got != customDir {
		t.Fatalf("Dir() = %q, want %q", got, customDir)
	}
}

func TestDirResolvesRelativeQCHome(t *testing.T) {
	t.Setenv("QC_HOME", "custom-qc")
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	got, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(workingDir, "custom-qc")
	if got != want {
		t.Fatalf("Dir() = %q, want %q", got, want)
	}
}

func TestPathJoinsApplicationRelativeElements(t *testing.T) {
	appDir := filepath.Join(t.TempDir(), "custom-qc")
	t.Setenv("QC_HOME", appDir)
	want := filepath.Join(appDir, "cookies", "qcc.work.json")

	got, err := Path("cookies", "qcc.work.json")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}
