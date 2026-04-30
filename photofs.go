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
	baseDir
	jpgDir     *MonthsDir
	rawDir     *MonthsDir
	undatedDir *PhotoDir
}

func newRoot(path string) *Root {
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

	return &Root{
		baseDir:    newBaseDir(),
		jpgDir:     newMonthsDir(jpgEntries),
		rawDir:     newMonthsDir(rawEntries),
		undatedDir: newPhotoDir(undatedFiles),
	}
}

func (r *Root) Attach() (p9.File, error) {
	return r, nil
}

func (r *Root) Walk(names []string) ([]p9.QID, p9.File, error) {
	if len(names) == 0 {
		return []p9.QID{r.qid}, r, nil
	}

	switch names[0] {
	case "jpg":
		return []p9.QID{r.jpgDir.qid}, r.jpgDir, nil
	case "raw":
		return []p9.QID{r.rawDir.qid}, r.rawDir, nil
	case "undated":
		return []p9.QID{r.undatedDir.qid}, r.undatedDir, nil
	}

	return nil, nil, syscall.ENOENT
}

func (r *Root) Readdir(offset uint64, count uint32) (p9.Dirents, error) {
	all := []struct {
		name string
		qid  p9.QID
	}{
		{"jpg", r.jpgDir.qid},
		{"raw", r.rawDir.qid},
		{"undated", r.undatedDir.qid},
	}

	end := min(offset+uint64(count), uint64(len(all)))
	slice := all[offset:end]

	if len(slice) == 0 {
		return nil, nil
	}

	dirs := make(p9.Dirents, len(slice))
	for i, v := range slice {
		dirs[i] = p9.Dirent{
			QID:    v.qid,
			Offset: offset + uint64(i) + 1,
			Name:   v.name,
			Type:   p9.TypeDir,
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
