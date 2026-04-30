package main

import (
	"log"
	"path/filepath"

	"github.com/knusbaum/go9p/fs"
)

func newPhotoDir(gofs *fs.FS, name string, files []string) *fs.StaticDir {
	dir := fs.NewStaticDir(gofs.NewStat(name, "none", "none", 0555))
	for _, path := range files {
		if err := dir.AddChild(newPhotoFile(gofs, path)); err != nil {
			log.Printf("skipping %s: %v", filepath.Base(path), err)
		}
	}
	return dir
}
