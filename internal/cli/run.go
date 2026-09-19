package cli

import (
	"fmt"
	"io"

	"github.com/alecthomas/kong"
)

func Execute(args []string, stdout, stderr io.Writer) int {
	command := New()
	command.Auth.Import.output = stdout
	command.Config.Set.UserAgent.output = stdout
	command.Search.output = stdout
	command.Area.List.output = stdout
	command.Industry.List.output = stdout
	command.Status.output = stdout
	exitCode := -1
	parser, err := kong.New(
		command,
		kong.Name("qc"),
		kong.Description("Query Chinese enterprise information from the command line."),
		kong.UsageOnError(),
		kong.Writers(stdout, stderr),
		kong.Vars(versionVars()),
		kong.Exit(func(code int) {
			exitCode = code
		}),
	)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}

	ctx, err := parser.Parse(args)
	if exitCode >= 0 {
		return exitCode
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	if err := command.loadProfileConfigFor(ctx); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	command.configureProfile()
	if err := ctx.Run(); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
