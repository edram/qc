package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/edram/qi/internal/models"
)

type enterpriseDetailSearch interface {
	GetEnterpriseDetail(context.Context, string) (models.EnterpriseDetail, error)
}

// EnterpriseCmd contains commands for viewing one enterprise.
type EnterpriseCmd struct {
	View    EnterpriseViewCmd `kong:"cmd,help='View one enterprise.'"`
	qcc     enterpriseDetailSearch
	aiqicha enterpriseDetailSearch
	output  io.Writer
}

func newEnterpriseCmd(profile, userAgent string) EnterpriseCmd {
	return EnterpriseCmd{
		qcc:     newSearchQCC(profile, userAgent),
		aiqicha: newSearchAiqicha(profile, userAgent),
		output:  io.Discard,
	}
}

// EnterpriseViewCmd prints one normalized enterprise from the selected provider.
type EnterpriseViewCmd struct {
	Reference string             `arg:"" required:"" name:"id" help:"Enterprise ID returned by the selected provider (QCC also accepts qc/id)."`
	Provider  searchProviderName `name:"provider" enum:"qcc,aiqicha" default:"qcc" help:"Enterprise data provider (${enum})."`
}

func (cmd *EnterpriseViewCmd) Run(enterpriseCmd *EnterpriseCmd) error {
	id := strings.TrimSpace(cmd.Reference)
	var provider enterpriseDetailSearch
	switch cmd.Provider {
	case "", searchProviderQCC:
		var err error
		id, err = qccEnterpriseReference(id)
		if err != nil {
			return err
		}
		provider = enterpriseCmd.qcc
	case searchProviderAiqicha:
		provider = enterpriseCmd.aiqicha
	default:
		return fmt.Errorf("enterprise view: unsupported provider %q", cmd.Provider)
	}
	detail, err := provider.GetEnterpriseDetail(context.Background(), id)
	if err != nil {
		return err
	}
	return encodeJSON(enterpriseCmd.output, detail)
}

func qccEnterpriseReference(reference string) (string, error) {
	reference = strings.TrimSpace(reference)
	if strings.HasPrefix(reference, "qc/") {
		reference = strings.TrimPrefix(reference, "qc/")
	}
	if reference == "" {
		return "", fmt.Errorf("enterprise view: enterprise ID is required")
	}
	if strings.ContainsAny(reference, "/\\?#") {
		return "", fmt.Errorf("enterprise view: invalid QCC enterprise ID %q", reference)
	}
	return reference, nil
}
