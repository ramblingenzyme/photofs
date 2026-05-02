package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type exifEntry struct {
	SourceFile       string `json:"SourceFile"`
	DateTimeOriginal string `json:"DateTimeOriginal"`
}

type State struct {
	mu         sync.RWMutex
	path       string
	log        strings.Builder
	generation uint
	files      map[string][]exifEntry
}

func NewState() *State {
	return &State{files: make(map[string][]exifEntry)}
}


var jpgExts = map[string]bool{
	".jpg": true, ".jpeg": true,
}

var rawExts = map[string]bool{
	".cr2": true, ".cr3": true, ".nef": true, ".arw": true,
	".dng": true, ".raf": true, ".orf": true, ".rw2": true,
}

func (s *State) Sync() error {
	s.mu.RLock()
	path := s.path
	s.mu.RUnlock()

	out, err := exec.Command("exiftool", "-r", "-json", "-DateTimeOriginal", path).Output()
	if err != nil {
		return err
	}

	if out == nil {
		return nil
	}

	var entries []exifEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return err
	}

	var jpgEntries, rawEntries, undatedEntries []exifEntry

	for _, e := range entries {
		ext := strings.ToLower(filepath.Ext(e.SourceFile))
		if !jpgExts[ext] && !rawExts[ext] {
			continue
		}
		if e.DateTimeOriginal == "" {
			undatedEntries = append(undatedEntries, e)
			continue
		}
		if jpgExts[ext] {
			jpgEntries = append(jpgEntries, e)
		} else {
			rawEntries = append(rawEntries, e)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.files = map[string][]exifEntry{
		"jpg":     jpgEntries,
		"raw":     rawEntries,
		"undated": undatedEntries,
	}
	s.generation++

	return nil
}

func (s *State) Path(p string) error {
	s.mu.Lock()
	s.path = p
	s.mu.Unlock()
	return s.Sync()
}

func (s *State) Status() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.path == "" {
		return "uninitialised"
	}
	return "ok"
}

func (s *State) Files(k string) ([]exifEntry, uint) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.files[k], s.generation
}
