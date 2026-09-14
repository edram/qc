package cli

import (
	"context"

	"github.com/edram/qc/internal/models"
	"github.com/edram/qc/internal/qcc"
)

const searchSourceQCC searchSourceName = "qcc"

type searchQCC struct {
	api *qcc.Client
}

func newSearchQCC() search {
	return &searchQCC{api: qcc.New()}
}

func (s *searchQCC) SearchEnterprises(ctx context.Context, query string) ([]models.Enterprise, error) {
	return s.api.SearchEnterprises(ctx, query)
}

func (s *searchQCC) SearchPeople(ctx context.Context, query string) ([]models.Person, error) {
	return s.api.SearchPeople(ctx, query)
}
