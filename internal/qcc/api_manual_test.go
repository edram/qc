package qcc

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestManualSearchMulti(t *testing.T) {
	ctx := context.Background()
	client := New(Options{
		PID: "3ff8f7c8c0c2daa435bc6307fb87af7c",
		TID: "fd8fc3911321c8c4e3ce9b0144b4e600",
	})
	response, err := client.Post(
		ctx,
		"/api/search/searchMulti",
		strings.NewReader(`{"searchKey":"91320594088140947F","pageIndex":1,"pageSize":20}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) > 512 {
		body = append(body[:512], "..."...)
	}
	t.Logf("status=%s body=%s", response.Status, body)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("status = %s", response.Status)
	}
}
