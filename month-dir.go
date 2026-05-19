package main

import (
	"log"
	"time"

	"github.com/knusbaum/go9p/fs"
)

type MonthDir struct {
	fs.StaticDir
	fsys       *fs.FS
	built      bool
	generation uint
	monthsMap  map[string]fs.FSNode
	name       string
	state      *State
}

func newMonthsDir(gofs *fs.FS, name string, state *State) *MonthDir {
	_, generation := state.Files(name)

	return &MonthDir{
		StaticDir:        *fs.NewStaticDir(gofs.NewStat(name, "none", "none", 0555)),
		fsys:       gofs,
		state:      state,
		generation: generation,
		monthsMap:  make(map[string]fs.FSNode),
		name:       name,
	}
}

func (m *MonthDir) Build() {
	photoMap := make(map[string][]string)
	files, generation := m.state.Files(m.name)

	for _, e := range files {
		t, err := time.Parse("2006:01:02 15:04:05", e.DateTimeOriginal)
		if err != nil {
			log.Printf("skipping %s: %v", e.SourceFile, err)
			continue
		}
		month := t.Format("2006-01")
		photoMap[month] = append(photoMap[month], e.SourceFile)
	}

	m.monthsMap = make(map[string]fs.FSNode)
	for mo, f := range photoMap {
		m.monthsMap[mo] = newPhotoDir(m.fsys, mo, f)
	}
	m.built = true
	m.generation = generation
}

func (m *MonthDir) Children() map[string]fs.FSNode {
	m.Lock()
	defer m.Unlock()

	_, curGen := m.state.Files(m.name)

	if curGen == 0 {
		return make(map[string]fs.FSNode)
	}

	if !m.built || m.generation != curGen {
		m.Build()
	}

	return m.monthsMap
}
