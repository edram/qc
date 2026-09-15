package cli

import (
	"context"

	"github.com/edram/qi/internal/models"
	"github.com/edram/qi/internal/qcc"
)

const searchSourceQCC searchSourceName = "qcc"

type searchQCC struct {
	api *qcc.Client
}

func newSearchQCC() search {
	return &searchQCC{api: qcc.New(qcc.Options{})}
}

func (*searchQCC) SearchEnterprises(context.Context, string) ([]models.Enterprise, error) {
	return nil, nil
}

func (*searchQCC) SearchPeople(context.Context, string) ([]models.Person, error) {
	return nil, nil
}
