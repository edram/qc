package cli

import (
	"context"

	"github.com/edram/qi/internal/aiqicha"
	"github.com/edram/qi/internal/models"
)

const searchSourceAiqicha searchSourceName = "aiqicha"

type searchAiqicha struct {
	api *aiqicha.Client
}

func newSearchAiqicha() search {
	return &searchAiqicha{api: aiqicha.New(aiqicha.Options{})}
}

func (*searchAiqicha) SearchEnterprises(context.Context, string, enterpriseSearchFilter) ([]models.Enterprise, error) {
	return nil, nil
}

func (*searchAiqicha) SearchPeople(context.Context, string, personSearchFilter) ([]models.Person, error) {
	return nil, nil
}
