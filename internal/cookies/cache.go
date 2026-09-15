package cookies

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// CachedSource reads a JSON cache before falling back to Source and updating it.
type CachedSource struct {
	Path   string
	Source Source
}

func (s CachedSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	cookies, err := readFile(s.Path)
	if err == nil && len(cookies) > 0 {
		return cookies, nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read cookie cache: %w", err)
	}
	if s.Source == nil {
		return nil, errors.New("cookie source required")
	}

	cookies, err = s.Source.Cookies(ctx)
	if err != nil {
		return nil, err
	}
	if err := writeFile(s.Path, cookies); err != nil {
		return nil, fmt.Errorf("write cookie cache: %w", err)
	}
	return cookies, nil
}

// DefaultPath returns a per-source, per-profile cookie cache under an application's config directory.
func DefaultPath(application, source, profile string) (string, error) {
	if profile == "" {
		profile = "default"
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, application, "cookies", source+"."+profile+".json"), nil
}
