package main

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

func newPhotoFile(gofs *fs.FS, path string) fs.File {
	stat := gofs.NewStat(filepath.Base(path), "none", "none", 0444)
	if info, err := os.Stat(path); err == nil {
		stat.Length = uint64(info.Size())
		stat.Mtime = uint32(info.ModTime().Unix())
		stat.Atime = uint32(info.ModTime().Unix())
	}

	var mu sync.Mutex
	fids := map[uint64]*os.File{}

	return &fs.WrappedFile{
		File: fs.NewBaseFile(stat),
		OpenF: func(fid uint64, _ proto.Mode) error {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			mu.Lock()
			fids[fid] = f
			mu.Unlock()
			return nil
		},
		ReadF: func(fid uint64, offset, count uint64) ([]byte, error) {
			mu.Lock()
			f := fids[fid]
			mu.Unlock()
			buf := make([]byte, count)
			n, err := f.ReadAt(buf, int64(offset))
			if err == io.EOF {
				err = nil
			}
			return buf[:n], err
		},
		CloseF: func(fid uint64) error {
			mu.Lock()
			f := fids[fid]
			delete(fids, fid)
			mu.Unlock()
			return f.Close()
		},
	}
}
