package main

import (
	"fmt"
	"os"

	"github.com/gentle-ai/ff15-openspec-agents/internal/ff15"
)

func main() {
	if err := ff15.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
