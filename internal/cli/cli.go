package cli

import (
	"strings"

	"github.com/alecthomas/kong"
	appconfig "github.com/edram/qi/internal/config"
)

const defaultProfileName = "default"

type CLI struct {
	Version   kong.VersionFlag `help:"Print version."`
	Profile   string           `help:"Authentication profile." env:"QC_PROFILE" default:"default"`
	UserAgent string           `help:"Browser User-Agent associated with QCC cookies."`
	Auth      AuthCmd          `kong:"cmd,help='Authentication and cookies.'"`
	Config    ConfigCmd        `kong:"cmd,help='Profile configuration.'"`
	Search    SearchCmd        `kong:"cmd,help='Search enterprise and person records.'"`
	Area      AreaCmd          `kong:"cmd,help='Search the local area catalog.'"`
	Industry  IndustryCmd      `kong:"cmd,help='Search the local industry catalog.'"`
	Status    StatusCmd        `kong:"cmd,help='Show cookie and account status.'"`
}

func New() *CLI {
	return &CLI{
		Profile:  defaultProfileName,
		Auth:     newAuthCmd(defaultProfileName),
		Config:   newConfigCmd(defaultProfileName),
		Search:   newSearchCmd(defaultProfileName, ""),
		Area:     newAreaCmd(),
		Industry: newIndustryCmd(),
		Status:   newStatusCmd(defaultProfileName, ""),
	}
}

func (c *CLI) loadProfileConfig() error {
	if strings.TrimSpace(c.UserAgent) != "" {
		return nil
	}
	userAgent, err := appconfig.UserAgent(c.Profile)
	if err != nil {
		return err
	}
	c.UserAgent = userAgent
	return nil
}

func (c *CLI) loadProfileConfigFor(ctx *kong.Context) error {
	if selected := ctx.Selected(); selected != nil {
		path := selected.FullPath()
		if strings.HasSuffix(path, " area list") || strings.HasSuffix(path, " industry list") {
			return nil
		}
	}
	return c.loadProfileConfig()
}

func (c *CLI) configureProfile() {
	c.Auth.Import.profile = c.Profile
	c.Config.Set.UserAgent.profile = c.Profile
	c.Search.qcc = newSearchQCC(c.Profile, c.UserAgent)
	c.Status.configureProfile(c.Profile, c.UserAgent)
}
