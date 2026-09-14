package cli

import "github.com/alecthomas/kong"

type CLI struct {
	Version kong.VersionFlag `help:"Print version."`
	Search  SearchCmd        `kong:"cmd,help='Search enterprise and person records.'"`
}

func New() *CLI {
	return &CLI{
		Search: newSearchCmd(),
	}
}
