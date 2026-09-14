package qcc

import (
	"context"
	"errors"

	"github.com/edram/qi/internal/models"
)

var ErrNotImplemented = errors.New("qcc API is not implemented")

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
