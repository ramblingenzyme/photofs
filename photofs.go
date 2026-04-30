package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/hugelgupf/p9/fsimpl/templatefs"
	"github.com/hugelgupf/p9/p9"
)

var qidCounter uint64

func nextQID(t p9.QIDType) p9.QID {
	// Static version number since the FS is readonly
	return p9.QID{Type: t, Version: 1, Path: atomic.AddUint64(&qidCounter, 1)}
}

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

type Root struct {
	templatefs.NoopFile
	jpgDir *MonthsDir
	rawDir *MonthsDir
	qid p9.QID
}

func newRoot(path string) *Root {
	out, err := exec.Command("exiftool", "-r", "-json", "-DateTimeOriginal", path).Output()
	if err != nil {
		log.Printf("exiftool: %v", err)
	}

	var jpgEntries, rawEntries []exifEntry
	if out != nil {
		var entries []exifEntry
		if err := json.Unmarshal(out, &entries); err != nil {
			log.Printf("exiftool parse: %v", err)
		}
		for _, e := range entries {
			ext := strings.ToLower(filepath.Ext(e.SourceFile))
			if jpgExts[ext] {
				jpgEntries = append(jpgEntries, e)
			} else if rawExts[ext] {
				rawEntries = append(rawEntries, e)
			}
		}
	}

	return &Root{
		qid: nextQID(p9.TypeDir),
		jpgDir: newMonthsDir(jpgEntries),
		rawDir: newMonthsDir(rawEntries),
	}
}

func (r *Root) GetAttr(req p9.AttrMask) (p9.QID, p9.AttrMask, p9.Attr, error) {
	return r.qid,
		p9.AttrMask{Mode: true},
		p9.Attr{Mode: p9.ModeDirectory | 0555},
		nil
}

func (r *Root) Open(mode p9.OpenFlags) (p9.QID, uint32, error) {
	return r.qid, 4096, nil
}

func (r *Root) Attach() (p9.File, error) {
	return r, nil
}

func (r *Root) Walk(names []string) ([]p9.QID, p9.File, error) {
	if len(names) == 0 {
		return []p9.QID{r.qid}, r, nil
	}

	var child *MonthsDir

	switch names[0] {
	case "jpg":
		child = r.jpgDir
	case "raw":
		child = r.rawDir
	}

	if child == nil {
		return nil, nil, syscall.ENOENT
	}

	return []p9.QID{child.qid}, child, nil
}

func (r *Root) Readdir(offset uint64, count uint32) (p9.Dirents, error) {
	all := []string{"jpg","raw"}
	end := min(offset+uint64(count), uint64(len(all)))
	slice := all[offset:end]

	if len(slice) == 0 {
		return nil, nil
	}

	dirs := make(p9.Dirents, len(slice))
	for i, v := range slice {
		dirs[i] = p9.Dirent{
			Offset: offset + uint64(i) + 1,
			Name: v,
			Type: p9.TypeDir,
		}

		switch v {
			case "jpg":
				dirs[i].QID = r.jpgDir.qid
			case "raw":
				dirs[i].QID = r.rawDir.qid
		}
	}

	return dirs, nil
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

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	root := newRoot(*path)

	log.Printf("serving 9p on %s", *addr)

	if err := p9.NewServer(root).Serve(ln); err != nil {
		log.Fatal(err)
	}
}
