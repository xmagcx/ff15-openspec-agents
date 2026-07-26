package ff15

import (
	"embed"
	"io/fs"
)

//go:embed syncassets/AGENTS.md syncassets/agents/*.agent.md syncassets/prompts/*.prompt.md syncassets/docs/*.md
var embeddedSyncAssets embed.FS

func syncAssetsFS() fs.FS {
	sub, err := fs.Sub(embeddedSyncAssets, "syncassets")
	if err != nil {
		panic(err)
	}
	return sub
}
