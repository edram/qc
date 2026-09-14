package cli

import (
	"context"

	"github.com/edram/qc/internal/aiqicha"
	"github.com/edram/qc/internal/models"
)

const searchSourceAiqicha searchSourceName = "aiqicha"

type searchAiqicha struct {
	api *aiqicha.Client
}

func newSearchAiqicha() search {
	return &searchAiqicha{api: aiqicha.New()}
}

func (s *searchAiqicha) SearchEnterprises(ctx context.Context, query string) ([]models.Enterprise, error) {
	return s.api.SearchEnterprises(ctx, query)
}

func (s *searchAiqicha) SearchPeople(ctx context.Context, query string) ([]models.Person, error) {
	return s.api.SearchPeople(ctx, query)
}
