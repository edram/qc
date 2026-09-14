package aiqicha

import (
	"context"
	"errors"

	"github.com/edram/qc/internal/models"
)

var ErrNotImplemented = errors.New("aiqicha API is not implemented")

type Client struct{}

func New() *Client {
	return &Client{}
}

func (*Client) SearchEnterprises(context.Context, string) ([]models.Enterprise, error) {
	return nil, ErrNotImplemented
}

func (*Client) SearchPeople(context.Context, string) ([]models.Person, error) {
	return nil, ErrNotImplemented
}
