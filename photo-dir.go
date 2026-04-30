package main

import (
	"slices"
	"syscall"

	"github.com/hugelgupf/p9/fsimpl/templatefs"
	"github.com/hugelgupf/p9/p9"
)

func newPhotoFile(path string, qid p9.QID) *OSFile {
	return &OSFile{path: path, qid: qid}
}

func newPhotoDir(base string, files []string, qid p9.QID) *PhotoDir {
	return &PhotoDir{base: base, files: files, qid: qid, osfiles: make(map[string]*OSFile)}
}

type PhotoDir struct {
	templatefs.NoopFile
	qid     p9.QID
	base    string
	files   []string
	osfiles map[string]*OSFile
}

func (p *PhotoDir) Walk(names []string) ([]p9.QID, p9.File, error) {
	if len(names) == 0 {
		return nil, p, nil
	}

	if !slices.Contains(p.files, names[0]) {
		return nil, nil, syscall.ENOENT
	}

	if p.osfiles[names[0]] == nil {
		p.osfiles[names[0]] = newPhotoFile(p.base+names[0], p9.QID{Type: p9.TypeRegular})
	}

	orig := p.osfiles[names[0]]
	file := &OSFile{path: orig.path, qid: orig.qid}

	return []p9.QID{file.qid}, file, nil
}

func (d *PhotoDir) Readdir(offset uint64, count uint32) (p9.Dirents, error) {
	end := min(offset+uint64(count), uint64(len(d.files)))
	slice := d.files[offset:end]

	if len(slice) == 0 {
		return nil, nil
	}

	dirs := make(p9.Dirents, len(slice))

	for i, v := range slice {
		if d.osfiles[v] == nil {
			d.osfiles[v] = newPhotoFile(d.base+v, p9.QID{Type: p9.TypeRegular})
		}

		qid := d.osfiles[v].qid
		dirs[i] = p9.Dirent{QID: qid, Offset: offset + uint64(i), Type: qid.Type, Name: v}
	}


	return dirs, nil
}
