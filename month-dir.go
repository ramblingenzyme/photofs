package main

import (
	"log"
	"sort"
	"syscall"
	"time"

	"github.com/hugelgupf/p9/fsimpl/templatefs"
	"github.com/hugelgupf/p9/p9"
)

type Photos struct {
	files []string // full disk paths
	dir   *PhotoDir
}

type MonthsDir struct {
	templatefs.NoopFile
	qid          p9.QID
	filesByMonth map[string]*Photos
	months       []string // ordered list of months
}

func newMonthsDir(entries []exifEntry) *MonthsDir {
	photoMap := make(map[string]*Photos)
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
			photoMap[month] = &Photos{}
			months = append(months, month)
		}
		photoMap[month].files = append(photoMap[month].files, e.SourceFile)
	}
	sort.Strings(months)

	return &MonthsDir{
		qid:          nextQID(p9.TypeDir),
		filesByMonth: photoMap,
		months:       months,
	}
}

func (d *MonthsDir) Open(mode p9.OpenFlags) (p9.QID, uint32, error) {
	return d.qid, 4096, nil
}

func (d *MonthsDir) GetAttr(req p9.AttrMask) (p9.QID, p9.AttrMask, p9.Attr, error) {
	return d.qid,
		p9.AttrMask{Mode: true},
		p9.Attr{Mode: p9.ModeDirectory | 0555},
		nil
}

func (d *MonthsDir) Walk(names []string) ([]p9.QID, p9.File, error) {
	if len(names) == 0 {
		return []p9.QID{d.qid}, d, nil
	}

	photos, ok := d.filesByMonth[names[0]]
	if !ok || len(names) > 1 {
		return nil, nil, syscall.ENOENT
	}

	if photos.dir == nil {
		photos.dir = newPhotoDir(photos.files)
	}

	return []p9.QID{photos.dir.qid}, photos.dir, nil
}

func (d *MonthsDir) Readdir(offset uint64, count uint32) (p9.Dirents, error) {
	end := min(offset+uint64(count), uint64(len(d.months)))
	slice := d.months[offset:end]

	if len(slice) == 0 {
		return nil, nil
	}

	dirs := make(p9.Dirents, len(slice))

	for i, v := range slice {
		photos := d.filesByMonth[v]
		if photos.dir == nil {
			photos.dir = newPhotoDir(photos.files)
		}
		qid := photos.dir.qid
		dirs[i] = p9.Dirent{QID: qid, Offset: offset + uint64(i) + 1, Type: qid.Type, Name: v}
	}

	return dirs, nil
}
