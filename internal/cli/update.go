package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/edram/qi/internal/update"
)

type UpdateCmd struct {
	service update.Service
	output  io.Writer
}

func newUpdateCmd() UpdateCmd {
	return UpdateCmd{service: update.Service{}, output: io.Discard}
}

func (cmd *UpdateCmd) Run() error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	if resolvedPath, err := filepath.EvalSymlinks(executablePath); err == nil {
		executablePath = resolvedPath
	}
	result, err := cmd.service.Update(context.Background(), version, executablePath)
	if err != nil {
		return err
	}
	if result.Updated {
		_, err = fmt.Fprintf(cmd.output, "Updated qc from %s to %s\n", result.CurrentVersion, result.LatestVersion)
		return err
	}
	_, err = fmt.Fprintf(cmd.output, "qc is already up to date (%s)\n", result.LatestVersion)
	return err
}
