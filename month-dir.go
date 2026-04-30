package main

import (
	"syscall"

	"github.com/hugelgupf/p9/fsimpl/templatefs"
	"github.com/hugelgupf/p9/p9"
)

type Photos struct {
	files []string
	dir   *PhotoDir
}

type MonthsDir struct {
	templatefs.NoopFile
	qid          p9.QID
	base         string
	filesByMonth map[string]*Photos
	months       []string // ordered list of months
}

func newMonthsDir(base string, files []string) *MonthsDir {
	// TODO: integrate exiftool here to group files by month & populate the map
	photoMap := make(map[string]*Photos)

	return &MonthsDir{
		base:         base,
		filesByMonth: photoMap,
		months:       []string{},
	}
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
		// TODO: add Path to QID
		photos.dir = newPhotoDir(d.base, photos.files, p9.QID{Type: p9.TypeDir})
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
			// TODO: add Path to QID
			photos.dir = newPhotoDir(d.base, photos.files, p9.QID{Type: p9.TypeDir})
		}
		qid := photos.dir.qid
		dirs[i] = p9.Dirent{QID: qid, Offset: offset + uint64(i), Type: qid.Type, Name: v}
	}

	return dirs, nil
}
