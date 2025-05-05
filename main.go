package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var (
	// version and revision are set by goreleaser during the build process
	version  = "dev"
	revision = "dev"
)

func main() {
	// THIS IS THE ENTRY POINT. ALREADY SET UP.
	app := getCli()
	app.Run(context.Background(), os.Args)
}

func getCli() *cli.Command {
	cli.VersionPrinter = func(cmd *cli.Command) {
		root := cmd.Root()
		fmt.Printf("%s version %s (revision:%s)\n", root.Name, root.Version, revision)
	}

	cmd := &cli.Command{
		Name:    "ai-rules-manager",
		Usage:   "Manage AI agent configuration presets (rules)",
		Version: version,
	}

	return cmd
}
