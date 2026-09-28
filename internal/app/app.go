// Package app owns application-wide identity and filesystem conventions.
package app

import (
	"os"
	"path/filepath"
)

// Name is the stable application identifier used in user-scoped paths.
const Name = "qc"

// Dir returns the application's directory. QC_HOME overrides the default ~/.qc path.
func Dir() (string, error) {
	if override := os.Getenv("QC_HOME"); override != "" {
		return filepath.Abs(override)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "."+Name), nil
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
