package main

import (
	"flag"
	"log"
	"net"
	"syscall"

	"github.com/hugelgupf/p9/fsimpl/templatefs"
	"github.com/hugelgupf/p9/p9"
)

type Root struct {
	templatefs.NoopFile
	base   string
	jpgs   []string
	jpgDir *MonthsDir
	rawDir *MonthsDir
	raws   []string
}

func newRoot(path string) *Root {
	// TODO: lookup all the JPGs & RAWs in the path and populate the paths

	return &Root{base: path}
}

func (r *Root) Walk(names []string) ([]p9.QID, p9.File, error) {
	if len(names) == 0 {
		return nil, r, nil
	}

	var child *MonthsDir

	switch names[0] {
	case "jpg":
		if r.jpgDir == nil {
			r.jpgDir = newMonthsDir(r.base, r.jpgs)
		}
		child = r.jpgDir

	case "raw":
		if r.rawDir == nil {
			r.rawDir = newMonthsDir(r.base, r.raws)
		}
		child = r.rawDir
	}

	if child == nil {
		return nil, nil, syscall.ENOENT
	}

	return []p9.QID{child.qid}, child, nil
}

func (r *Root) Attach() (p9.File, error) {
	return r, nil
}

func main() {
	addr := flag.String("addr", "localhost:8000", "listen address")
	path := flag.String("path", "", "path to sd card")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("serving 9p on %s", *addr)

	if err := p9.NewServer(newRoot(*path)).Serve(ln); err != nil {
		log.Fatal(err)
	}
}
