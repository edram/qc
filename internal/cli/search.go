package cli

import "errors"

type SearchCmd struct {
	Ents SearchEntsCmd `kong:"cmd,help='Search enterprises.'"`
	Pers SearchPersCmd `kong:"cmd,help='Search people.'"`
}

type SearchArgs struct {
	Query string `arg:"" required:"" name:"query" help:"Name to search."`
}

type SearchEntsCmd struct{ SearchArgs }

func (cmd *SearchEntsCmd) Run() error {
	return errors.New("qc search ents is not implemented yet")
}

type SearchPersCmd struct{ SearchArgs }

func (cmd *SearchPersCmd) Run() error {
	return errors.New("qc search pers is not implemented yet")
}
