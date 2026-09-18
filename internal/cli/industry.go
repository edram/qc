package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/edram/qi/internal/industries"
)

type IndustryCmd struct {
	List IndustryListCmd `kong:"cmd,help='List or search local industries.'"`
}

type IndustryListCmd struct {
	Search   string `name:"search" help:"Search industry name."`
	Provider string `name:"provider" enum:"qcc,aiqicha" default:"qcc" help:"Industry provider."`
	Limit    int    `name:"limit" default:"20" help:"Maximum number of results."`
	output   io.Writer
}

func newIndustryCmd() IndustryCmd {
	return IndustryCmd{}
}

func (cmd *IndustryListCmd) Run() error {
	if cmd.Limit < 0 {
		return fmt.Errorf("industry limit must not be negative")
	}
	catalog, err := industries.Load(industries.Provider(cmd.Provider))
	if err != nil {
		return err
	}
	result := catalog.TopLevel(cmd.Limit)
	if cmd.Search != "" {
		result = catalog.Search(cmd.Search, cmd.Limit)
	}
	output := make([]industryListResult, 0, len(result))
	for _, industry := range result {
		output = append(output, industryListResult{
			Name: industry.Name,
			Code: industry.ProviderCode,
		})
	}
	return json.NewEncoder(cmd.output).Encode(output)
}

type industryListResult struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
