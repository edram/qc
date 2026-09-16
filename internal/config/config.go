package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultProfile = "default"

type file struct {
	UserAgent string `json:"user_agent,omitempty"`
}

// UserAgent returns the persisted User-Agent for profile.
func UserAgent(profileName string) (string, error) {
	settings, err := load(profileName)
	if err != nil {
		return "", err
	}
	return settings.UserAgent, nil
}

// SetUserAgent persists a User-Agent for profile and returns the config path.
func SetUserAgent(profileName, userAgent string) (string, error) {
	path, err := defaultPath(profileName)
	if err != nil {
		return "", err
	}
	settings, err := read(path)
	if err != nil {
		return "", err
	}
	settings.UserAgent = strings.TrimSpace(userAgent)
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func defaultPath(profileName string) (string, error) {
	profileName = normalizeProfile(profileName)
	if profileName == "." || profileName == ".." || strings.ContainsAny(profileName, `/\`) {
		return "", fmt.Errorf("invalid config profile %q", profileName)
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "qc", "config."+profileName+".json"), nil
}

func read(path string) (file, error) {
	var settings file
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return file{}, err
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return file{}, err
	}
	return settings, nil
}

func load(profileName string) (file, error) {
	defaultConfigPath, err := defaultPath(defaultProfile)
	if err != nil {
		return file{}, err
	}
	settings, err := read(defaultConfigPath)
	if err != nil {
		return file{}, err
	}
	if normalizeProfile(profileName) == defaultProfile {
		return settings, nil
	}
	profilePath, err := defaultPath(profileName)
	if err != nil {
		return file{}, err
	}
	profile, err := read(profilePath)
	if err != nil {
		return file{}, err
	}
	return merge(settings, profile), nil
}

func merge(base, override file) file {
	if override.UserAgent != "" {
		base.UserAgent = override.UserAgent
	}
	return base
}

func normalizeProfile(profileName string) string {
	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		return defaultProfile
	}
	return profileName
}
