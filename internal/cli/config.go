package cli

import (
	"fmt"
	"io"

	appconfig "github.com/edram/qi/internal/config"
)

type ConfigCmd struct {
	Set ConfigSetCmd `kong:"cmd,help='Set profile configuration.'"`
}

type ConfigSetCmd struct {
	UserAgent ConfigSetUserAgentCmd `kong:"cmd,name='user-agent',help='Set the browser User-Agent for the current profile.'"`
}

type ConfigSetUserAgentCmd struct {
	Value string `arg:"" name:"value" help:"Browser User-Agent value."`

	profile string
	output  io.Writer
}

func newConfigCmd(profile string) ConfigCmd {
	return ConfigCmd{
		Set: ConfigSetCmd{
			UserAgent: ConfigSetUserAgentCmd{
				profile: profile,
				output:  io.Discard,
			},
		},
	}
}

func (cmd *ConfigSetUserAgentCmd) Run() error {
	path, err := appconfig.SetUserAgent(cmd.profile, cmd.Value)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.output, "Saved User-Agent for profile %s to %s\n", cmd.profile, path)
	return err
}
