package main

import (
	"path/filepath"
	"syscall"

	"github.com/hugelgupf/p9/p9"
)

func newPhotoFile(path string) *OSFile {
	return &OSFile{path: path, qid: nextQID(p9.TypeRegular)}
}

func newPhotoDir(files []string) *PhotoDir {
	return &PhotoDir{baseDir: newBaseDir(), files: files, osfiles: make(map[string]*OSFile)}
}

type PhotoDir struct {
	baseDir
	files   []string           // full disk paths
	osfiles map[string]*OSFile // keyed by basename
}

func (p *PhotoDir) Walk(names []string) ([]p9.QID, p9.File, error) {
	if len(names) == 0 {
		return []p9.QID{p.qid}, p, nil
	}

	osf, ok := p.osfiles[names[0]]
	if !ok {
		var fullPath string
		for _, f := range p.files {
			if filepath.Base(f) == names[0] {
				fullPath = f
				break
			}
		}
		if fullPath == "" {
			return nil, nil, syscall.ENOENT
		}
		osf = newPhotoFile(fullPath)
		p.osfiles[names[0]] = osf
	}

	file := &OSFile{path: osf.path, qid: osf.qid}
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
		name := filepath.Base(v)
		if d.osfiles[name] == nil {
			d.osfiles[name] = newPhotoFile(v)
		}
		qid := d.osfiles[name].qid
		dirs[i] = p9.Dirent{QID: qid, Offset: offset + uint64(i) + 1, Type: qid.Type, Name: name}
	}

	return dirs, nil
}
