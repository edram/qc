package cli

import (
	"context"
	"fmt"

	"github.com/edram/qi/internal/models"
)

type searchSourceName string

// search is the provider-independent behavior used by the search commands.
type search interface {
	SearchEnterprises(context.Context, string) ([]models.Enterprise, error)
	SearchPeople(context.Context, string) ([]models.Person, error)
}

type SearchCmd struct {
	Sources []searchSourceName `name:"source" enum:"qcc,aiqicha" default:"qcc,aiqicha" help:"Search source (${enum}); repeat to select multiple."`
	Ents    SearchEntsCmd      `kong:"cmd,help='Search enterprises.'"`
	Pers    SearchPersCmd      `kong:"cmd,help='Search people.'"`

	qcc     search
	aiqicha search
}

func newSearchCmd() SearchCmd {
	return SearchCmd{
		qcc:     newSearchQCC(),
		aiqicha: newSearchAiqicha(),
	}
}

func (cmd *SearchCmd) resolveSource(name searchSourceName) (search, error) {
	switch name {
	case searchSourceQCC:
		return cmd.qcc, nil
	case searchSourceAiqicha:
		return cmd.aiqicha, nil
	default:
		return nil, fmt.Errorf("unknown search source %q", name)
	}
}

type SearchArgs struct {
	Query string `arg:"" required:"" name:"query" help:"Name to search."`
}

type SearchEntsCmd struct{ SearchArgs }

func (cmd *SearchEntsCmd) Run(searchCmd *SearchCmd) error {
	for _, name := range searchCmd.Sources {
		searcher, err := searchCmd.resolveSource(name)
		if err != nil {
			return err
		}
		if _, err := searcher.SearchEnterprises(context.Background(), cmd.Query); err != nil {
			return err
		}
	}
	return nil
}

type SearchPersCmd struct{ SearchArgs }

func (cmd *SearchPersCmd) Run(searchCmd *SearchCmd) error {
	for _, name := range searchCmd.Sources {
		searcher, err := searchCmd.resolveSource(name)
		if err != nil {
			return err
		}
		if _, err := searcher.SearchPeople(context.Background(), cmd.Query); err != nil {
			return err
		}
	}
	return nil
}
