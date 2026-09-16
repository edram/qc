package cli

import "context"

type statusAiqicha struct{}

func newStatusAiqicha() statusAccountProvider {
	return statusAiqicha{}
}

func (statusAiqicha) Account(context.Context) (statusAccountResult, error) {
	return statusAccountResult{Name: "AIQICHA", NotImplemented: true}, nil
}
