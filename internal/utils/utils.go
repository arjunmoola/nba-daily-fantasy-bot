package utils

import (
	//"fmt"
	"io/fs"
	"path"
)

func FindGoProjectRootFs(fsys fs.FS, currentDir string) (string, bool, error) {
	var found bool

Loop:
	for currentDir != "." {
		entries, err := fs.ReadDir(fsys, currentDir)

		if err != nil {
			return "", false, err
		}

		for _, entry := range entries {
			if entry.Name() == "go.mod" {
				found = true
				break Loop
			}
		}

		currentDir = path.Dir(currentDir)
	}

	return currentDir, found, nil
}
