package cli

import (
	"context"
	"errors"

	"github.com/edram/qi/internal/aiqicha"
	"github.com/edram/qi/internal/models"
)

const searchSourceAiqicha searchSourceName = "aiqicha"

var errAiqichaSearchNotImplemented = errors.New("aiqicha API is not implemented")

type searchAiqicha struct {
	api *aiqicha.Client
}

func newSearchAiqicha() search {
	return &searchAiqicha{api: aiqicha.New(aiqicha.Options{})}
}

func (*searchAiqicha) SearchEnterprises(context.Context, string) ([]models.Enterprise, error) {
	return nil, errAiqichaSearchNotImplemented
}

func (*searchAiqicha) SearchPeople(context.Context, string) ([]models.Person, error) {
	return nil, errAiqichaSearchNotImplemented
}
