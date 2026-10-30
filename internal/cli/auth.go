package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/edram/qi/internal/aiqicha"
	"github.com/edram/qi/internal/qcc"
	"github.com/steipete/sweetcookie"
)

type AuthCmd struct {
	Import AuthImportCmd `kong:"cmd,help='Import browser cookies.'"`
}

type AuthImportCmd struct {
	Provider string `name:"provider" enum:"qcc,aiqicha" default:"qcc" help:"Cookie provider (${enum})."`
	Browser  string `help:"Browser name (chrome|brave|edge|firefox|safari)."`
	// BrowserProfile is distinct from the global qc profile: it selects a browser-local profile to read.
	// Based on: https://github.com/openclaw/spogo/blob/3d4edc230f5ece848b81642e95f3682f3ed3e8d7/internal/cli/auth.go
	BrowserProfile string `name:"browser-profile" help:"Browser profile name."`

	profile string
	output  io.Writer
}

func newAuthCmd(profile string) AuthCmd {
	return AuthCmd{
		Import: AuthImportCmd{
			profile: profile,
			output:  io.Discard,
		},
	}
}

func (cmd *AuthImportCmd) Run() error {
	browser := strings.ToLower(strings.TrimSpace(cmd.Browser))
	if browser == "" {
		browser = string(sweetcookie.BrowserChrome)
	}
	importCookies := qcc.ImportBrowserCookies
	if cmd.Provider == "aiqicha" {
		importCookies = aiqicha.ImportBrowserCookies
	}
	count, path, err := importCookies(
		context.Background(),
		cmd.profile,
		sweetcookie.Browser(browser),
		strings.TrimSpace(cmd.BrowserProfile),
	)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.output, "Saved %d cookies to %s\n", count, path)
	return err
}
