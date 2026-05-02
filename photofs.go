package main

import (
	"flag"
	"log"
	"os"

	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/fs"
)

func newRoot(path string) *fs.FS {
	state := NewState()
	err := state.Path(path)
	if err != nil {
		log.Printf("exiftool: %v", err)
	}

	undatedEntries, _ := state.Files("undated")

	gofs, root := fs.NewFS("none", "none", 0555)
	root.AddChild(newMonthsDir(gofs, "jpg", state))
	root.AddChild(newMonthsDir(gofs, "raw", state))

	undatedFiles := make([]string, len(undatedEntries))
	for i, e := range undatedEntries {
		undatedFiles[i] = e.SourceFile
	}

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
