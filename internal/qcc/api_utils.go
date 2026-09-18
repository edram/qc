package qcc

import (
	"context"
	"errors"
	"io"
	"regexp"
)

var (
	pidPattern = regexp.MustCompile(`window\.pid\s*=\s*['"]([^'"]+)['"]`)
	tidPattern = regexp.MustCompile(`window\.tid\s*=\s*['"]([^'"]+)['"]`)
)

func (c *Client) getPIDAndTID(ctx context.Context) (string, string, error) {
	c.identifierMu.Lock()
	defer c.identifierMu.Unlock()

	if c.pid != "" && c.tid != "" {
		return c.pid, c.tid, nil
	}

	pid, tid, err := c.FetchPIDAndTID(ctx)
	if err != nil {
		return "", "", err
	}
	if c.pid == "" {
		c.pid = pid
	}
	if c.tid == "" {
		c.tid = tid
	}
	return c.pid, c.tid, nil
}

func (c *Client) identifiers() (string, string) {
	c.identifierMu.Lock()
	defer c.identifierMu.Unlock()
	return c.pid, c.tid
}

// FetchPIDAndTID reads the request identifiers embedded in the configured base page.
func (c *Client) FetchPIDAndTID(ctx context.Context) (string, string, error) {
	response, err := c.httpClient.WithContext(ctx).Get("")
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", err
	}

	pid := identifier(pidPattern, body)
	if pid == "" {
		return "", "", errors.New("qcc: pid not found in base page")
	}
	tid := identifier(tidPattern, body)
	if tid == "" {
		return "", "", errors.New("qcc: tid not found in base page")
	}
	return pid, tid, nil
}

func identifier(pattern *regexp.Regexp, body []byte) string {
	match := pattern.FindSubmatch(body)
	if len(match) != 2 {
		return ""
	}
	return string(match[1])
}
