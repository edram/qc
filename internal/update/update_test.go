package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestServiceUpdateDownloadsAndReplacesBinary(t *testing.T) {
	archiveName := testArtifactName("1.2.0")
	archiveData := buildArchive(t, []byte("new qc binary"))
	checksum := sha256.Sum256(archiveData)
	checksumData := []byte(fmt.Sprintf("%s  %s\n", hex.EncodeToString(checksum[:]), archiveName))

	baseURL := "https://updates.example.test"
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var body []byte
		switch req.URL.Path {
		case "/repos/test/qc/releases/latest":
			body = []byte(fmt.Sprintf(`{"tag_name":"v1.2.0","assets":[{"name":%q,"browser_download_url":%q},{"name":"checksums.txt","browser_download_url":%q}]}`, archiveName, baseURL+"/archive", baseURL+"/checksums"))
		case "/archive":
			body = archiveData
		case "/checksums":
			body = checksumData
		default:
			return nil, fmt.Errorf("unexpected URL %s", req.URL.String())
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header)}, nil
	})}

	executable := filepath.Join(t.TempDir(), "qc")
	if err := os.WriteFile(executable, []byte("old qc binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	var replaced []byte
	service := Service{
		HTTPClient: client,
		Repository: "test/qc",
		APIBaseURL: baseURL,
		Replace: func(source, destination string, _ fs.FileMode) error {
			var err error
			replaced, err = os.ReadFile(source)
			return err
		},
	}

	result, err := service.Update(context.Background(), "1.0.0", executable)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Updated || result.LatestVersion != "1.2.0" {
		t.Fatalf("result = %#v, want an update to 1.2.0", result)
	}
	if !bytes.Equal(replaced, []byte("new qc binary")) {
		t.Fatalf("replaced binary = %q", replaced)
	}
}

func TestServiceUpdateSkipsWhenCurrentVersionIsLatest(t *testing.T) {
	var requests int
	service := Service{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			body := []byte(`{"tag_name":"v1.2.0"}`)
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header)}, nil
		})},
		Repository: "test/qc",
		APIBaseURL: "https://updates.example.test",
	}

	result, err := service.Update(context.Background(), "1.2.0", filepath.Join(t.TempDir(), "qc"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Updated || result.LatestVersion != "1.2.0" {
		t.Fatalf("result = %#v, want no update", result)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}

func TestServiceUpdateRejectsInvalidChecksum(t *testing.T) {
	archiveName := testArtifactName("1.2.0")
	archiveData := buildArchive(t, []byte("new qc binary"))
	baseURL := "https://updates.example.test"
	service := Service{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			var body []byte
			switch req.URL.Path {
			case "/repos/test/qc/releases/latest":
				body = []byte(fmt.Sprintf(`{"tag_name":"v1.2.0","assets":[{"name":%q,"browser_download_url":%q},{"name":"checksums.txt","browser_download_url":%q}]}`, archiveName, baseURL+"/archive", baseURL+"/checksums"))
			case "/archive":
				body = archiveData
			case "/checksums":
				body = []byte("0000000000000000000000000000000000000000000000000000000000000000  " + archiveName + "\n")
			default:
				return nil, fmt.Errorf("unexpected URL %s", req.URL.String())
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header)}, nil
		})},
		Repository: "test/qc",
		APIBaseURL: baseURL,
	}

	_, err := service.Update(context.Background(), "1.0.0", filepath.Join(t.TempDir(), "qc"))
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("error = %v, want checksum error", err)
	}
}

func TestReplaceBinaryReplacesAtomically(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("running executable replacement has different Windows locking semantics")
	}
	directory := t.TempDir()
	source := filepath.Join(directory, "source")
	destination := filepath.Join(directory, "qc")
	if err := os.WriteFile(source, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := replaceBinary(source, destination, 0o755); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "new" {
		t.Fatalf("destination = %q, want %q", contents, "new")
	}
}

func testArtifactName(version string) string {
	suffix := "tar.gz"
	if runtime.GOOS == "windows" {
		suffix = "zip"
	}
	return fmt.Sprintf("qc_%s_%s_%s.%s", version, runtime.GOOS, runtime.GOARCH, suffix)
}

func buildArchive(t *testing.T, content []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	if runtime.GOOS == "windows" {
		writer := zip.NewWriter(&output)
		entry, err := writer.Create("qc.exe")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return output.Bytes()
	}
	gzipWriter := gzip.NewWriter(&output)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: "qc", Mode: 0o755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
