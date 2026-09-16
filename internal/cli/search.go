package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

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
	output  io.Writer
}

func newSearchCmd(profile string) SearchCmd {
	return SearchCmd{
		qcc:     newSearchQCC(profile),
		aiqicha: newSearchAiqicha(),
		output:  io.Discard,
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
	enterprises := make([]models.Enterprise, 0)
	for _, name := range searchCmd.Sources {
		searcher, err := searchCmd.resolveSource(name)
		if err != nil {
			return err
		}
		found, err := searcher.SearchEnterprises(context.Background(), cmd.Query)
		if err != nil {
			return err
		}
		enterprises = append(enterprises, found...)
	}
	return json.NewEncoder(searchCmd.output).Encode(enterprises)
}

type SearchPersCmd struct{ SearchArgs }

func (cmd *SearchPersCmd) Run(searchCmd *SearchCmd) error {
	people := make([]models.Person, 0)
	for _, name := range searchCmd.Sources {
		searcher, err := searchCmd.resolveSource(name)
		if err != nil {
			return err
		}
		found, err := searcher.SearchPeople(context.Background(), cmd.Query)
		if err != nil {
			return err
		}
		people = append(people, found...)
	}
	return json.NewEncoder(searchCmd.output).Encode(people)
}
