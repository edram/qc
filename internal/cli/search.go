package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/edram/qi/internal/models"
)

type searchProviderName string

// search is the provider-independent behavior used by the search commands.
type search interface {
	SearchEnterprises(context.Context, string, enterpriseSearchFilter) ([]models.Enterprise, error)
	SearchPeople(context.Context, string, personSearchFilter) ([]models.Person, error)
}

type enterpriseSearchField string

const (
	enterpriseSearchFieldName                enterpriseSearchField = "name"
	enterpriseSearchFieldScope               enterpriseSearchField = "scope"
	enterpriseSearchFieldIntroduction        enterpriseSearchField = "introduction"
	enterpriseSearchFieldAddress             enterpriseSearchField = "address"
	enterpriseSearchFieldBrand               enterpriseSearchField = "brand"
	enterpriseSearchFieldLegalRepresentative enterpriseSearchField = "legal-representative"
	enterpriseSearchFieldPatent              enterpriseSearchField = "patent"
	enterpriseSearchFieldTrademark           enterpriseSearchField = "trademark"
	enterpriseSearchFieldShareholder         enterpriseSearchField = "shareholder"
	enterpriseSearchFieldKeyPersonnel        enterpriseSearchField = "key-personnel"
)

type enterpriseSearchStatus string

const (
	enterpriseSearchStatusActive       enterpriseSearchStatus = "active"
	enterpriseSearchStatusMoved        enterpriseSearchStatus = "moved"
	enterpriseSearchStatusEstablishing enterpriseSearchStatus = "establishing"
	enterpriseSearchStatusCancelled    enterpriseSearchStatus = "cancelled"
	enterpriseSearchStatusRevoked      enterpriseSearchStatus = "revoked"
)

type enterpriseSearchFilter struct {
	Fields     []enterpriseSearchField
	Areas      []string
	Industries []string
	Statuses   []enterpriseSearchStatus
}

type personSearchFilter struct {
	Area     string
	Industry string
}

type SearchCmd struct {
	Providers []searchProviderName `name:"provider" enum:"qcc,aiqicha" default:"qcc,aiqicha" help:"Search provider (${enum}); repeat to select multiple."`
	Ents      SearchEntsCmd        `kong:"cmd,help='Search enterprises.'"`
	Pers      SearchPersCmd        `kong:"cmd,help='Search people.'"`

	qcc     search
	aiqicha search
	output  io.Writer
}

func newSearchCmd(profile, userAgent string) SearchCmd {
	return SearchCmd{
		qcc:     newSearchQCC(profile, userAgent),
		aiqicha: newSearchAiqicha(),
		output:  io.Discard,
	}
}

func (cmd *SearchCmd) resolveProvider(name searchProviderName) (search, error) {
	switch name {
	case searchProviderQCC:
		return cmd.qcc, nil
	case searchProviderAiqicha:
		return cmd.aiqicha, nil
	default:
		return nil, fmt.Errorf("unknown search provider %q", name)
	}
}

type SearchArgs struct {
	Query string `arg:"" required:"" name:"query" help:"Name to search."`
}

type SearchEntsCmd struct {
	SearchArgs
	Fields     []enterpriseSearchField  `name:"match" enum:"name,scope,introduction,address,brand,legal-representative,patent,trademark,shareholder,key-personnel" help:"Fields to match (${enum}); repeat to select multiple."`
	Areas      []string                 `name:"area" help:"Area name, full path, or QCC code; repeat to select multiple."`
	Industries []string                 `name:"industry" help:"Industry name; repeat to select multiple."`
	Statuses   []enterpriseSearchStatus `name:"status" enum:"active,moved,establishing,cancelled,revoked" help:"Registration status (${enum}); repeat to select multiple."`
}

func (cmd *SearchEntsCmd) Run(searchCmd *SearchCmd) error {
	enterprises := make([]models.Enterprise, 0)
	for _, name := range searchCmd.Providers {
		searcher, err := searchCmd.resolveProvider(name)
		if err != nil {
			return err
		}
		found, err := searcher.SearchEnterprises(context.Background(), cmd.Query, enterpriseSearchFilter{
			Fields:     cmd.Fields,
			Areas:      cmd.Areas,
			Industries: cmd.Industries,
			Statuses:   cmd.Statuses,
		})
		if err != nil {
			return err
		}
		enterprises = append(enterprises, found...)
	}
	return json.NewEncoder(searchCmd.output).Encode(enterprises)
}

type SearchPersCmd struct {
	SearchArgs
	Area     string `help:"Area path as displayed by QCC, for example '广东省 深圳市'."`
	Industry string `help:"Industry name."`
}

func (cmd *SearchPersCmd) Run(searchCmd *SearchCmd) error {
	people := make([]models.Person, 0)
	for _, name := range searchCmd.Providers {
		searcher, err := searchCmd.resolveProvider(name)
		if err != nil {
			return err
		}
		found, err := searcher.SearchPeople(context.Background(), cmd.Query, personSearchFilter{
			Area:     cmd.Area,
			Industry: cmd.Industry,
		})
		if err != nil {
			return err
		}
		people = append(people, found...)
	}
	return json.NewEncoder(searchCmd.output).Encode(people)
}
