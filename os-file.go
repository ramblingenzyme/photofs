package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

type OsFile struct{
	fs.BaseFile

	path string
	fids map[uint64]*os.File
}

func (p *OsFile) Open(fid uint64, _ proto.Mode) error {
	f, err := os.Open(p.path)
	if err != nil {
		return err
	}

	p.Lock()
	p.fids[fid] = f
	p.Unlock()

	return nil
}

func (p *OsFile) Read(fid uint64, offset, count uint64) ([]byte, error) {
	p.RLock()
	f := p.fids[fid]
	p.Unlock()

	buf := make([]byte, count)
	n, err := f.ReadAt(buf, int64(offset))
	if err == io.EOF {
		err = nil
	}

	return buf[:n], err
}

func (p *OsFile) Close(fid uint64) error {
	p.Lock()
	f := p.fids[fid]
	delete(p.fids, fid)
	p.Unlock()

	return f.Close()
}

func newOsFile(gofs *fs.FS, path string) fs.File {
	stat := gofs.NewStat(filepath.Base(path), "none", "none", 0444)
	if info, err := os.Stat(path); err == nil {
		stat.Length = uint64(info.Size())
		stat.Mtime = uint32(info.ModTime().Unix())
		stat.Atime = uint32(info.ModTime().Unix())
	}

	file := &OsFile{
		BaseFile: *fs.NewBaseFile(stat),
		path: path,
	}

	return file
}
