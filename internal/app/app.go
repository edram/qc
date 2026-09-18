// Package app owns application-wide identity and filesystem conventions.
package app

import (
	"os"
	"path/filepath"
)

// Name is the stable application identifier used in user-scoped paths.
const Name = "qc"

// Dir returns the application's directory under the current user's configuration directory.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, Name), nil
}

// Path returns a path within the application directory.
func Path(elements ...string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	pathElements := append([]string{dir}, elements...)
	return filepath.Join(pathElements...), nil
}
