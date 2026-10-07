// Package update downloads and installs qc release binaries.
package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const DefaultRepository = "edram/qc"

type ReplaceFunc func(source, destination string, mode fs.FileMode) error

type Service struct {
	HTTPClient *http.Client
	Repository string
	APIBaseURL string
	Replace    ReplaceFunc
}

type Result struct {
	CurrentVersion string
	LatestVersion  string
	Updated        bool
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

func (s Service) Update(ctx context.Context, currentVersion, executablePath string) (Result, error) {
	if strings.TrimSpace(executablePath) == "" {
		return Result{}, errors.New("executable path cannot be empty")
	}

	release, err := s.latestRelease(ctx)
	if err != nil {
		return Result{}, err
	}
	latestVersion := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	if latestVersion == "" {
		return Result{}, errors.New("latest release has no version")
	}
	if _, err := parseVersion(latestVersion); err != nil {
		return Result{}, fmt.Errorf("invalid latest release version %q: %w", latestVersion, err)
	}
	result := Result{CurrentVersion: currentVersion, LatestVersion: latestVersion}
	if !needsUpdate(currentVersion, latestVersion) {
		return result, nil
	}

	archiveName := artifactName(latestVersion)
	archiveURL := s.assetURL(release, archiveName)
	if archiveURL == "" {
		return Result{}, fmt.Errorf("latest release does not contain %s", archiveName)
	}
	checksumsURL := s.assetURL(release, "checksums.txt")
	if checksumsURL == "" {
		return Result{}, errors.New("latest release does not contain checksums.txt")
	}

	destinationDir := filepath.Dir(executablePath)
	workDir, err := os.MkdirTemp(destinationDir, ".qc-update-")
	if err != nil {
		return Result{}, fmt.Errorf("create update workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	archivePath := filepath.Join(workDir, archiveName)
	if err := s.download(ctx, archiveURL, archivePath); err != nil {
		return Result{}, fmt.Errorf("download %s: %w", archiveName, err)
	}
	checksums, err := s.downloadBytes(ctx, checksumsURL)
	if err != nil {
		return Result{}, fmt.Errorf("download checksums.txt: %w", err)
	}
	expected, err := checksumFor(string(checksums), archiveName)
	if err != nil {
		return Result{}, err
	}
	if err := verifyChecksum(archivePath, expected); err != nil {
		return Result{}, err
	}

	binaryPath, mode, err := extractBinary(archivePath, workDir)
	if err != nil {
		return Result{}, err
	}
	if info, statErr := os.Stat(executablePath); statErr == nil && info.Mode().Perm() != 0 {
		mode = info.Mode().Perm()
	}
	replace := s.Replace
	if replace == nil {
		replace = replaceBinary
	}
	if err := replace(binaryPath, executablePath, mode); err != nil {
		return Result{}, fmt.Errorf("replace executable: %w", err)
	}
	result.Updated = true
	return result, nil
}

func (s Service) latestRelease(ctx context.Context) (release, error) {
	repository := s.Repository
	if repository == "" {
		repository = os.Getenv("QC_REPOSITORY")
	}
	if repository == "" {
		repository = DefaultRepository
	}
	if !validRepository(repository) {
		return release{}, fmt.Errorf("invalid repository %q", repository)
	}
	baseURL := strings.TrimRight(s.APIBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	body, err := s.downloadBytes(ctx, baseURL+"/repos/"+repository+"/releases/latest")
	if err != nil {
		return release{}, fmt.Errorf("fetch latest release: %w", err)
	}
	var result release
	if err := json.Unmarshal(body, &result); err != nil {
		return release{}, fmt.Errorf("decode latest release: %w", err)
	}
	return result, nil
}

func (s Service) assetURL(release release, name string) string {
	for _, asset := range release.Assets {
		if asset.Name == name && asset.DownloadURL != "" {
			return asset.DownloadURL
		}
	}
	if release.TagName == "" {
		return ""
	}
	repository := s.Repository
	if repository == "" {
		repository = os.Getenv("QC_REPOSITORY")
	}
	if repository == "" {
		repository = DefaultRepository
	}
	return "https://github.com/" + repository + "/releases/download/" + release.TagName + "/" + name
}

func (s Service) downloadBytes(ctx context.Context, url string) ([]byte, error) {
	response, err := s.get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func (s Service) download(ctx context.Context, url, destination string) error {
	response, err := s.get(ctx, url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, response.Body); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func (s Service) get(ctx context.Context, url string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "qc-update")
	client := s.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4<<10))
		return nil, fmt.Errorf("unexpected HTTP status %s", response.Status)
	}
	return response, nil
}

func checksumFor(contents, name string) (string, error) {
	for _, line := range strings.Split(contents, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.TrimPrefix(fields[1], "*") != name {
			continue
		}
		checksum := strings.ToLower(fields[0])
		if len(checksum) != sha256.Size*2 {
			return "", fmt.Errorf("invalid checksum for %s", name)
		}
		if _, err := hex.DecodeString(checksum); err != nil {
			return "", fmt.Errorf("invalid checksum for %s: %w", name, err)
		}
		return checksum, nil
	}
	return "", fmt.Errorf("checksums.txt does not contain %s", name)
}

func verifyChecksum(filename, expected string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open archive for checksum: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash archive: %w", err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != strings.ToLower(expected) {
		return fmt.Errorf("checksum verification failed for %s", filepath.Base(filename))
	}
	return nil
}

func extractBinary(archivePath, destinationDir string) (string, fs.FileMode, error) {
	if runtime.GOOS == "windows" {
		return extractZip(archivePath, destinationDir)
	}
	return extractTarGz(archivePath, destinationDir)
}

func extractTarGz(archivePath, destinationDir string) (string, fs.FileMode, error) {
	archive, err := os.Open(archivePath)
	if err != nil {
		return "", 0, fmt.Errorf("open archive: %w", err)
	}
	defer archive.Close()
	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return "", 0, fmt.Errorf("read gzip archive: %w", err)
	}
	defer gzipReader.Close()
	return extractTar(tar.NewReader(gzipReader), destinationDir)
}

func extractTar(reader *tar.Reader, destinationDir string) (string, fs.FileMode, error) {
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", 0, fmt.Errorf("read tar archive: %w", err)
		}
		if !isBinaryEntry(header.Name, "qc") || !isRegularTarEntry(header.Typeflag) {
			continue
		}
		mode := fs.FileMode(header.Mode).Perm()
		if mode == 0 {
			mode = 0o755
		}
		binaryPath, err := writeExtracted(reader, destinationDir, "qc", mode)
		return binaryPath, mode, err
	}
	return "", 0, errors.New("release archive does not contain qc")
}

func extractZip(archivePath, destinationDir string) (string, fs.FileMode, error) {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", 0, fmt.Errorf("read zip archive: %w", err)
	}
	defer archive.Close()
	for _, entry := range archive.File {
		if !isBinaryEntry(entry.Name, "qc.exe") || !entry.FileInfo().Mode().IsRegular() {
			continue
		}
		reader, err := entry.Open()
		if err != nil {
			return "", 0, fmt.Errorf("open qc in archive: %w", err)
		}
		mode := fs.FileMode(0o755)
		binaryPath, writeErr := writeExtracted(reader, destinationDir, "qc.exe", mode)
		closeErr := reader.Close()
		if writeErr != nil {
			return "", 0, writeErr
		}
		if closeErr != nil {
			return "", 0, fmt.Errorf("close qc in archive: %w", closeErr)
		}
		return binaryPath, mode, nil
	}
	return "", 0, errors.New("release archive does not contain qc.exe")
}

func writeExtracted(reader io.Reader, destinationDir string, name string, mode fs.FileMode) (string, error) {
	destination := filepath.Join(destinationDir, name)
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode.Perm())
	if err != nil {
		return "", fmt.Errorf("create extracted binary: %w", err)
	}
	if _, err := io.Copy(file, reader); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("extract binary: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close extracted binary: %w", err)
	}
	return destination, nil
}

func replaceBinary(source, destination string, mode fs.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(destination), ".qc-update-binary-")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	sourceFile, err := os.Open(source)
	if err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := io.Copy(temporary, sourceFile); err != nil {
		_ = sourceFile.Close()
		_ = temporary.Close()
		return err
	}
	if err := sourceFile.Close(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Chmod(mode.Perm()); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		return err
	}
	return nil
}

func isBinaryEntry(name, binaryName string) bool {
	name = strings.TrimPrefix(name, "./")
	return name == binaryName && path.Clean(name) == binaryName
}

func isRegularTarEntry(typeflag byte) bool {
	return typeflag == tar.TypeReg || typeflag == tar.TypeRegA
}

func artifactName(version string) string {
	extension := "tar.gz"
	if runtime.GOOS == "windows" {
		extension = "zip"
	}
	return "qc_" + version + "_" + runtime.GOOS + "_" + runtime.GOARCH + "." + extension
}

func validRepository(repository string) bool {
	parts := strings.Split(repository, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != "" && !strings.ContainsAny(repository, "?# \t\r\n")
}

func needsUpdate(current, latest string) bool {
	if strings.TrimSpace(current) == "" || strings.TrimSpace(current) == "dev" {
		return true
	}
	comparison, err := compareVersions(current, latest)
	return err != nil || comparison < 0
}

func compareVersions(left, right string) (int, error) {
	leftParts, err := parseVersion(left)
	if err != nil {
		return 0, err
	}
	rightParts, err := parseVersion(right)
	if err != nil {
		return 0, err
	}
	for i := range leftParts.core {
		if leftParts.core[i] != rightParts.core[i] {
			if leftParts.core[i] < rightParts.core[i] {
				return -1, nil
			}
			return 1, nil
		}
	}
	if leftParts.pre == rightParts.pre {
		return 0, nil
	}
	if leftParts.pre == "" {
		return 1, nil
	}
	if rightParts.pre == "" {
		return -1, nil
	}
	return strings.Compare(leftParts.pre, rightParts.pre), nil
}

type parsedVersion struct {
	core [3]int
	pre  string
}

func parseVersion(value string) (parsedVersion, error) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	value = strings.SplitN(value, "+", 2)[0]
	parts := strings.SplitN(value, "-", 2)
	coreParts := strings.Split(parts[0], ".")
	if len(coreParts) > 3 || coreParts[0] == "" {
		return parsedVersion{}, fmt.Errorf("invalid version %q", value)
	}
	var result parsedVersion
	for i, part := range coreParts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return parsedVersion{}, fmt.Errorf("invalid version %q", value)
		}
		result.core[i] = number
	}
	if len(parts) == 2 {
		result.pre = parts[1]
	}
	return result, nil
}
