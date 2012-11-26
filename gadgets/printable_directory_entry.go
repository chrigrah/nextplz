package gadgets

import (
	"github.com/chrigrah/nextplz/backend"
	"path/filepath"
	"strings"
)

type PLDirectoryEntry struct {
	dl        *backend.DirectoryEntry
	list_type uint
}

const (
	PLDE_NORMAL = iota
	PLDE_RECURSIVE
)

func (plde *PLDirectoryEntry) DisplayValue() string {
	switch plde.list_type {
	case PLDE_NORMAL:
		return dl.Name
	case PLDE_RECURSIVE:
		if !strings.HasSuffix(dl.Name, ".rar") {
			return dl.Name
		} else {
			dirs := filepath.SplitList(AbsPath)
			if len(dirs) <= 2 {
				return dl.Name
			}
			nearest_dir := dirs[len(dirs)-2]
			return filepath.Join(nearest_dir, dl.Name)
		}
	}
}

func (plde *PLDirectoryEntry) FilterValue() string {
	return plde.DisplayValue()
}

func (dl *PLDirectoryEntry) Path() string {
	return dl.AbsPath
}

func (dl *PLDirectoryEntry) IsAccessible() bool {
	return dl.IsAccessible
}

func (dl *PLDirectoryEntry) IsVideo() bool {
	return dl.IsVideo
}

func (dl *PLDirectoryEntry) IsDir() bool {
	return dl.IsDir
}
