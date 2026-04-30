package main

import (
	"log"
	"sort"
	"time"

	"github.com/knusbaum/go9p/fs"
)

func newMonthsDir(gofs *fs.FS, name string, entries []exifEntry) *fs.StaticDir {
	photoMap := make(map[string][]string)
	var months []string

	for _, e := range entries {
		if e.DateTimeOriginal == "" {
			continue
		}
		t, err := time.Parse("2006:01:02 15:04:05", e.DateTimeOriginal)
		if err != nil {
			log.Printf("skipping %s: %v", e.SourceFile, err)
			continue
		}
		month := t.Format("2006-01")
		if _, ok := photoMap[month]; !ok {
			months = append(months, month)
		}
		photoMap[month] = append(photoMap[month], e.SourceFile)
	}
	sort.Strings(months)

	dir := fs.NewStaticDir(gofs.NewStat(name, "none", "none", 0555))
	for _, month := range months {
		dir.AddChild(newPhotoDir(gofs, month, photoMap[month]))
	}
	return dir
}
