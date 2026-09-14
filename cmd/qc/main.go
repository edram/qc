package main

import (
	"os"

	"github.com/edram/qc/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
