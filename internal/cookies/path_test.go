package cookies

import (
	"testing"

	"github.com/edram/qi/internal/app"
)

func TestPathSeparatesProfiles(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	path, err := Path("qcc", "default")
	if err != nil {
		t.Fatal(err)
	}
	want, err := app.Path("cookies", "qcc.default.json")
	if err != nil {
		t.Fatal(err)
	}
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}

	path, err = Path("qcc", "work")
	if err != nil {
		t.Fatal(err)
	}
	want, err = app.Path("cookies", "qcc.work.json")
	if err != nil {
		t.Fatal(err)
	}
	if path != want {
		t.Fatalf("profile path = %q, want %q", path, want)
	}
}

func TestPathRejectsUnsafeProfiles(t *testing.T) {
	for _, profile := range []string{".", "..", "../work", "work/personal", `work\personal`} {
		if _, err := Path("qcc", profile); err == nil {
			t.Errorf("Path() error = nil for profile %q", profile)
		}
	}
}
