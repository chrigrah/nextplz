package backend

import (
	"os"
	"strings"
	"path/filepath"
	"container/list"
)

var (
	VideoExtensions []string
)

type FileEntry struct {
	Name string
	AbsPath string
	Contents list.List
	contents_read bool
	IsDir, IsAccessible, IsVideo bool
	Parent *FileEntry
	ElementInParent *list.Element
}

func CreateDirEntry(abspath string) (*FileEntry, error) {
	var new_entry FileEntry = FileEntry{
		AbsPath: abspath,
		contents_read: false,
		IsDir: true,
		IsAccessible: true,
	}

	err := new_entry.ValidateContents()
	if err != nil {
		return nil, err
	}

	return &new_entry, nil
}

func (fe *FileEntry) ValidateContents() error {
	if fe.IsDir && !fe.contents_read {
		err := filepath.Walk(fe.AbsPath, fe.walk_func())
		if err != nil {
			return err
		}

		fe.contents_read = true
	}
	return nil
}

func (file_entry *FileEntry) walk_func() filepath.WalkFunc {
	fe := file_entry
	return func(dir string, fi os.FileInfo, err error) error {
		if dir == fe.AbsPath {
			return err
		}

		var new_file = FileEntry{
			Name: filepath.Base(dir),
			AbsPath: dir,
			IsDir: fi.IsDir(),
			IsAccessible: err == nil,
			IsVideo: is_video(dir),
			Parent: fe,
		}
		new_file.ElementInParent = fe.Contents.PushBack(&new_file)
		if fi.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}
}

func (fe *FileEntry) GetElementInParent() (eip *list.Element) {
	fe.ValidateParent()
	return fe.ElementInParent
}

func (fe *FileEntry) GetParent() (parent *FileEntry) {
	fe.ValidateParent()
	return fe.Parent
}

func (fe *FileEntry) ValidateParent() {
	if fe.Parent == nil {
		if fe.is_root() {
			fe.Parent = fe
			fe.ElementInParent = &list.Element{ Value: fe }
		} else {
			parent, err := CreateDirEntry(filepath.Join(fe.AbsPath, ".."))
			if err != nil {
				panic(err) // TODO: improve
			}
			fe.Parent = parent
			for e := fe.Parent.Contents.Front(); e != nil; e = e.Next() {
				if e.Value.(*FileEntry).Name == fe.Name {
					fe.ElementInParent = e
					e.Value = fe
					break
				}
			}
		}
	}
}

func is_dir(path string) bool {
	file_info, err := os.Stat(path)
	if err != nil || !file_info.IsDir() {
		return false
	}
	return true
}

func is_accessible(path string) bool {
	_, err := os.Open(path)
	if err != nil {
		return false
	}
	return true
}

func is_video(path string) bool {
	for _, ext := range VideoExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func (fe *FileEntry) is_root() bool {
	return fe.AbsPath == strings.Join([]string{filepath.VolumeName(fe.AbsPath), string(filepath.Separator)}, "")
}
