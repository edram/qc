package cli

import "github.com/alecthomas/kong"

const defaultProfileName = "default"

type CLI struct {
	Version kong.VersionFlag `help:"Print version."`
	Profile string           `help:"Authentication profile." env:"QC_PROFILE" default:"default"`
	Search  SearchCmd        `kong:"cmd,help='Search enterprise and person records.'"`
	Status  StatusCmd        `kong:"cmd,help='Show cookie and account status.'"`
}

func New() *CLI {
	return &CLI{
		Profile: defaultProfileName,
		Search:  newSearchCmd(defaultProfileName),
		Status:  newStatusCmd(defaultProfileName),
	}
}

func (c *CLI) configureProfile() {
	c.Search.qcc = newSearchQCC(c.Profile)
	c.Status.configureProfile(c.Profile)
}
