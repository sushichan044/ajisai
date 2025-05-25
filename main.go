package main

import (
	"fmt"
	"os"

	"github.com/sushichan044/ajisai/cmd/ajisai"
)

func main() {
	if err := ajisai.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
