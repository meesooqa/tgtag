package fs

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

type Finder struct {
	log *slog.Logger
}

func NewFinder(log *slog.Logger) *Finder {
	return &Finder{log: log}
}

func (f *Finder) FindFiles(fileOrDirPath string, filesChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(filesChan)

	i, e := os.Stat(fileOrDirPath)
	if e != nil {
		f.log.Error("Can't stat file", "fileOrDirPath", fileOrDirPath, "err", e.Error())
		return
	}

	if i.IsDir() {
		f.findFilesInDir(fileOrDirPath, filesChan)
	} else {
		// is file
		filesChan <- fileOrDirPath
	}
}

func (f *Finder) findFilesInDir(root string, filesChan chan<- string) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			f.log.Error("error while walking", "path", path, "err", err)
		}

		if !info.IsDir() {
			filesChan <- path
		}
		return nil
	})
	if err != nil {
		f.log.Error("directory walk error", "err", err)
	}
}
