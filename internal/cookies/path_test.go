package cookies

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPathSeparatesProfiles(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	path, err := DefaultPath("qc", "qcc", "default")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(userConfigDir, "qc", "cookies", "qcc.default.json")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}

	path, err = DefaultPath("qc", "qcc", "work")
	if err != nil {
		t.Fatal(err)
	}
	want = filepath.Join(userConfigDir, "qc", "cookies", "qcc.work.json")
	if path != want {
		t.Fatalf("profile path = %q, want %q", path, want)
	}
}

func TestDefaultPathRejectsUnsafeProfiles(t *testing.T) {
	for _, profile := range []string{".", "..", "../work", "work/personal", `work\personal`} {
		if _, err := DefaultPath("qc", "qcc", profile); err == nil {
			t.Errorf("DefaultPath() error = nil for profile %q", profile)
		}
	}
}
