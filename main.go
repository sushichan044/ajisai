package main

import (
	"fmt"
	"os"

	"github.com/sushichan044/ajisai/cmd/ajisai"
)

var (
	// noglobals error is suppressed by golangci-lint.
	revision = "dev"
)

func main() {
	if err := ajisai.Run(os.Args, revision); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
