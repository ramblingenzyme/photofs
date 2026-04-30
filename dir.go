package main

import (
	"github.com/hugelgupf/p9/fsimpl/templatefs"
	"github.com/hugelgupf/p9/p9"
)

type baseDir struct {
	templatefs.NoopFile
	qid p9.QID
}

func newBaseDir() baseDir {
	return baseDir{qid: nextQID(p9.TypeDir)}
}

func (d *baseDir) Open(mode p9.OpenFlags) (p9.QID, uint32, error) {
	return d.qid, 4096, nil
}

func (d *baseDir) GetAttr(req p9.AttrMask) (p9.QID, p9.AttrMask, p9.Attr, error) {
	return d.qid,
		p9.AttrMask{Mode: true},
		p9.Attr{Mode: p9.ModeDirectory | 0555},
		nil
}
