package cookies

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type storedCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires"`
	Secure   bool      `json:"secure"`
	HTTPOnly bool      `json:"http_only"`
}

func readFile(path string) ([]*http.Cookie, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var stored []storedCookie
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}
	cookies := make([]*http.Cookie, 0, len(stored))
	now := time.Now()
	for _, cookie := range stored {
		if !cookie.Expires.IsZero() && !cookie.Expires.After(now) {
			continue
		}
		cookies = append(cookies, &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  cookie.Expires,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HTTPOnly,
		})
	}
	return cookies, nil
}

// Write stores cookies in the JSON cache format used by CachedSource.
func Write(path string, cookies []*http.Cookie) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	stored := make([]storedCookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		stored = append(stored, storedCookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  jsonSafeExpiry(cookie.Expires),
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HttpOnly,
		})
	}
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func jsonSafeExpiry(expires time.Time) time.Time {
	if expires.IsZero() {
		return time.Time{}
	}
	if _, err := expires.MarshalJSON(); err != nil {
		return time.Time{}
	}
	return expires
}
