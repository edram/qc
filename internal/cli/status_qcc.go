package cli

import (
	"context"
	"errors"

	"github.com/edram/qi/internal/qcc"
)

type statusQCC struct {
	api *qcc.Client
}

func newStatusQCCClient(profile string) *qcc.Client {
	return qcc.New(qcc.Options{Profile: profile})
}

func newStatusQCC(api *qcc.Client) statusAccountProvider {
	return &statusQCC{api: api}
}

func (s *statusQCC) Account(ctx context.Context) (statusAccountResult, error) {
	result := statusAccountResult{Name: "QCC"}
	auth, err := s.api.AuthInfo(ctx)
	if err != nil {
		if errors.Is(err, qcc.ErrNoCookies) {
			return result, nil
		}
		return result, err
	}
	result.User = &statusUser{
		Nickname:    auth.Nickname,
		PhonePrefix: auth.PhonePrefix,
		Phone:       auth.Phone,
		Email:       auth.Email,
	}
	return result, nil
}
