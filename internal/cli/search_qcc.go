package cli

import (
	"context"
	"errors"

	"github.com/edram/qi/internal/models"
	"github.com/edram/qi/internal/qcc"
)

const searchSourceQCC searchSourceName = "qcc"

var errQCCSearchNotImplemented = errors.New("qcc API is not implemented")

type searchQCC struct {
	api *qcc.Client
}

func newSearchQCC() search {
	return &searchQCC{api: qcc.New(qcc.Options{})}
}

func (*searchQCC) SearchEnterprises(context.Context, string) ([]models.Enterprise, error) {
	return nil, errQCCSearchNotImplemented
}

func (*searchQCC) SearchPeople(context.Context, string) ([]models.Person, error) {
	return nil, errQCCSearchNotImplemented
}
