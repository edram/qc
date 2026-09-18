package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/edram/qi/internal/app"
)

type statusCookiesProvider interface {
	Cookies(context.Context) (statusCookiesResult, error)
}

type statusAccountProvider interface {
	Account(context.Context) (statusAccountResult, error)
}

type statusCookiesResult struct {
	Sources []statusCookieSource
}

type statusCookieSource struct {
	Name    string
	Cookies []statusCookie
}

type statusAccountResult struct {
	Name           string
	NotImplemented bool
	User           *statusUser
}

type statusCookie struct {
	Name     string
	Domain   string
	Path     string
	Expires  time.Time
	Secure   bool
	HTTPOnly bool
}

type statusUser struct {
	Nickname    string
	PhonePrefix string
	Phone       string
	Email       string
}

type StatusCmd struct {
	profile string
	cookies statusCookiesProvider
	qcc     statusAccountProvider
	aiqicha statusAccountProvider
	output  io.Writer
}

func newStatusCmd(profile, userAgent string) StatusCmd {
	qccClient := newStatusQCCClient(profile, userAgent)
	return StatusCmd{
		profile: profile,
		cookies: newStatusCookies(qccClient),
		qcc:     newStatusQCC(qccClient),
		aiqicha: newStatusAiqicha(),
		output:  io.Discard,
	}
}

func (cmd *StatusCmd) configureProfile(profile, userAgent string) {
	qccClient := newStatusQCCClient(profile, userAgent)
	cmd.profile = profile
	cmd.cookies = newStatusCookies(qccClient)
	cmd.qcc = newStatusQCC(qccClient)
}

func (cmd *StatusCmd) Run() error {
	ctx := context.Background()
	appPath, err := app.Dir()
	if err != nil {
		return fmt.Errorf("resolve app path: %w", err)
	}
	_, _ = fmt.Fprintln(cmd.output, "Overview")
	_, _ = fmt.Fprintf(cmd.output, "  Profile: %s\n", cmd.profile)
	_, _ = fmt.Fprintf(cmd.output, "  App Path: %s\n\n", appPath)
	cookies, err := cmd.cookies.Cookies(ctx)
	if err != nil {
		return err
	}
	cmd.writeCookies(cookies)
	_, _ = fmt.Fprintln(cmd.output)
	_, _ = fmt.Fprintln(cmd.output, "Accounts")
	for _, provider := range []statusAccountProvider{cmd.qcc, cmd.aiqicha} {
		account, err := provider.Account(ctx)
		cmd.writeAccount(account)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cmd *StatusCmd) writeCookies(status statusCookiesResult) {
	_, _ = fmt.Fprintln(cmd.output, "Cookies")
	for _, source := range status.Sources {
		_, _ = fmt.Fprintf(cmd.output, "  %s: %d\n", source.Name, len(source.Cookies))
		for _, cookie := range source.Cookies {
			expires := "session"
			if !cookie.Expires.IsZero() {
				expires = cookie.Expires.Format(time.RFC3339)
			}
			_, _ = fmt.Fprintf(
				cmd.output,
				"    - %s: domain=%s path=%s expires=%s secure=%t http_only=%t\n",
				cookie.Name,
				cookie.Domain,
				cookie.Path,
				expires,
				cookie.Secure,
				cookie.HTTPOnly,
			)
		}
	}
}

func (cmd *StatusCmd) writeAccount(status statusAccountResult) {
	_, _ = fmt.Fprintf(cmd.output, "  %s\n", status.Name)
	if status.NotImplemented {
		_, _ = fmt.Fprintln(cmd.output, "    Status: not implemented")
		return
	}
	if status.User == nil {
		_, _ = fmt.Fprintln(cmd.output, "    Status: not authenticated")
		return
	}
	_, _ = fmt.Fprintln(cmd.output, "    Status: authenticated")
	_, _ = fmt.Fprintf(cmd.output, "    User: %s\n", status.User.Nickname)
	if status.User.Phone != "" {
		_, _ = fmt.Fprintf(cmd.output, "    Phone: %s %s\n", status.User.PhonePrefix, maskPhone(status.User.Phone))
	}
	if status.User.Email != "" {
		_, _ = fmt.Fprintf(cmd.output, "    Email: %s\n", maskEmail(status.User.Email))
	}
}

func maskPhone(phone string) string {
	if len(phone) <= 7 {
		return strings.Repeat("*", len(phone))
	}
	return phone[:3] + strings.Repeat("*", len(phone)-7) + phone[len(phone)-4:]
}

func maskEmail(email string) string {
	local, domain, found := strings.Cut(email, "@")
	if !found || local == "" {
		return strings.Repeat("*", len(email))
	}
	return local[:1] + "***@" + domain
}
