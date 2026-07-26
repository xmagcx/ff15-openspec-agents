//go:build ignore

package main

import (
	"log"

	"github.com/gentle-ai/ff15-openspec-agents/internal/ff15"
)

func main() {
	if err := ff15.GenerateProjectCover("skin/chocobo.png"); err != nil {
		log.Fatal(err)
	}
}
