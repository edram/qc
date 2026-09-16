package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	appconfig "github.com/edram/qi/internal/config"
	"github.com/edram/qi/internal/qcc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	userAgent, err := appconfig.UserAgent("default")
	if err != nil {
		return err
	}
	client := qcc.New(qcc.Options{
		PID:       "3ff8f7c8c0c2daa435bc6307fb87af7c",
		TID:       "fd8fc3911321c8c4e3ce9b0144b4e600",
		UserAgent: userAgent,
	})
	response, err := client.Post(
		context.Background(),
		"/api/search/searchMulti",
		strings.NewReader(`{"searchKey":"91320594088140947F","pageIndex":1,"pageSize":20}`),
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if len(body) > 512 {
		body = append(body[:512], "..."...)
	}
	fmt.Printf("status=%s body=%s\n", response.Status, body)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", response.Status)
	}
	return nil
}
