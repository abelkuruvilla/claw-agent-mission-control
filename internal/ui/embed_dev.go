//go:build dev

package ui

import (
	"io/fs"
	"os"
)

func Assets() (fs.FS, error) {
	return os.DirFS("ui/out"), nil
}
