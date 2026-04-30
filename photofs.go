package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/fs"
)

var jpgExts = map[string]bool{
	".jpg": true, ".jpeg": true,
}

var rawExts = map[string]bool{
	".cr2": true, ".cr3": true, ".nef": true, ".arw": true,
	".dng": true, ".raf": true, ".orf": true, ".rw2": true,
}

type exifEntry struct {
	SourceFile       string `json:"SourceFile"`
	DateTimeOriginal string `json:"DateTimeOriginal"`
}

func newRoot(path string) *fs.FS {
	out, err := exec.Command("exiftool", "-r", "-json", "-DateTimeOriginal", path).Output()
	if err != nil {
		log.Printf("exiftool: %v", err)
	}

	var jpgEntries, rawEntries []exifEntry
	var undatedFiles []string

	if out != nil {
		var entries []exifEntry
		if err := json.Unmarshal(out, &entries); err != nil {
			log.Printf("exiftool parse: %v", err)
		}
		for _, e := range entries {
			ext := strings.ToLower(filepath.Ext(e.SourceFile))
			if !jpgExts[ext] && !rawExts[ext] {
				continue
			}
			if e.DateTimeOriginal == "" {
				undatedFiles = append(undatedFiles, e.SourceFile)
				continue
			}
			if jpgExts[ext] {
				jpgEntries = append(jpgEntries, e)
			} else {
				rawEntries = append(rawEntries, e)
			}
		}
	}

	gofs, root := fs.NewFS("none", "none", 0555)
	root.AddChild(newMonthsDir(gofs, "jpg", jpgEntries))
	root.AddChild(newMonthsDir(gofs, "raw", rawEntries))
	root.AddChild(newPhotoDir(gofs, "undated", undatedFiles))
	return gofs
}

func main() {
	addr := flag.String("addr", "localhost:8000", "listen address")
	path := flag.String("path", "", "path to sd card")
	flag.Parse()

	if *path == "" {
		log.Fatal("--path is required")
	}
	if _, err := os.Stat(*path); err != nil {
		log.Fatalf("--path: %v", err)
	}

	gofs := newRoot(*path)

	log.Printf("serving 9p on %s", *addr)
	if err := go9p.Serve(*addr, gofs.Server()); err != nil {
		log.Fatal(err)
	}
}
