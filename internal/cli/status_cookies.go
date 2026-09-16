package cli

import (
	"context"
	"errors"

	"github.com/edram/qi/internal/qcc"
)

type statusCookies struct {
	qcc *qcc.Client
}

func newStatusCookies(qccClient *qcc.Client) statusCookiesProvider {
	return &statusCookies{qcc: qccClient}
}

func (s *statusCookies) Cookies(ctx context.Context) (statusCookiesResult, error) {
	source := statusCookieSource{Name: "QCC"}
	cookies, err := s.qcc.CookieInfo(ctx)
	if err != nil {
		if errors.Is(err, qcc.ErrNoCookies) {
			return statusCookiesResult{Sources: []statusCookieSource{source}}, nil
		}
		return statusCookiesResult{}, err
	}
	source.Cookies = make([]statusCookie, 0, len(cookies))
	for _, cookie := range cookies {
		source.Cookies = append(source.Cookies, statusCookie{
			Name:     cookie.Name,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  cookie.Expires,
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HTTPOnly,
		})
	}
	return statusCookiesResult{Sources: []statusCookieSource{source}}, nil
}
