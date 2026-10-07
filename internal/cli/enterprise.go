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
	View   EnterpriseViewCmd `kong:"cmd,help='View one enterprise.'"`
	qcc    enterpriseDetailSearch
	output io.Writer
}

func newEnterpriseCmd(profile, userAgent string) EnterpriseCmd {
	return EnterpriseCmd{
		qcc:    newSearchQCC(profile, userAgent),
		output: io.Discard,
	}
}

// EnterpriseViewCmd prints a normalized enterprise from a QCC firm page.
type EnterpriseViewCmd struct {
	Reference string `arg:"" required:"" name:"qc/id" help:"QCC enterprise ID, optionally prefixed with qc/."`
}

func (cmd *EnterpriseViewCmd) Run(enterpriseCmd *EnterpriseCmd) error {
	id, err := qccEnterpriseReference(cmd.Reference)
	if err != nil {
		return err
	}
	detail, err := enterpriseCmd.qcc.GetEnterpriseDetail(context.Background(), id)
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
