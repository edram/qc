package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/edram/qi/internal/regions"
)

type AreaCmd struct {
	List AreaListCmd `kong:"cmd,help='List or search local areas.'"`
}

type AreaListCmd struct {
	Search   string `name:"search" help:"Search area name."`
	Provider string `name:"provider" enum:"qcc" default:"qcc" help:"Area provider."`
	Limit    int    `name:"limit" default:"20" help:"Maximum number of results."`
	output   io.Writer
}

func newAreaCmd() AreaCmd {
	return AreaCmd{}
}

func (cmd *AreaListCmd) Run() error {
	if cmd.Limit < 0 {
		return fmt.Errorf("area limit must not be negative")
	}
	catalog, err := regions.Load(regions.Provider(cmd.Provider))
	if err != nil {
		return err
	}
	result := catalog.TopLevel(cmd.Limit)
	if cmd.Search != "" {
		result = catalog.Search(cmd.Search, cmd.Limit)
	}
	output := make([]areaListResult, 0, len(result))
	for _, area := range result {
		output = append(output, areaListResult{
			Name: area.Name,
			Code: area.ProviderCode,
		})
	}
	return json.NewEncoder(cmd.output).Encode(output)
}

type areaListResult struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
