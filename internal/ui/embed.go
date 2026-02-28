//go:build !dev

package ui

import (
	"embed"
	"io/fs"
)

// Embed all files from the Next.js build output
// The build process copies ui/out to internal/ui/assets before building
// Use all: prefix to include all files including those starting with _ or .
//
//go:embed all:assets
var embeddedUI embed.FS

func Assets() (fs.FS, error) {
	return fs.Sub(embeddedUI, "assets")
}
