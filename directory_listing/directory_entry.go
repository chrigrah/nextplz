package directory_listing

import (
	"os"
	"strings"
	"path/filepath"
	"container/list"
)

type file_entry struct {
	Name string
	AbsPath string
	Contents list.List
	contents_read bool
	IsDir, IsAccessible bool
	Parent *file_entry
	ElementInParent *list.Element
}

func CreateDirEntry(abspath string) (entry *file_entry, err error) {
	var new_entry file_entry = file_entry{
		AbsPath: abspath,
		contents_read: false,
		IsDir: true,
		IsAccessible: true,
	}

	new_entry.ValidateContents()

	return &new_entry, nil
}

func (fe *file_entry) ValidateContents() {
	if fe.IsDir && !fe.contents_read {
		directoryFile, err := os.Open(fe.AbsPath)
		if err != nil {
			panic(err)
		}

		files, err := directoryFile.Readdirnames(0)
		if err != nil {
			panic(err)
		}

		for _, filename := range files {
			var abspath = filepath.Join(fe.AbsPath, filename)
			var new_file = file_entry{
				Name: filename,
				AbsPath: abspath,
				IsDir: is_dir(abspath),
				IsAccessible: is_accessible(abspath),
				Parent: fe,
			}

			new_file.ElementInParent = fe.Contents.PushBack(&new_file)
		}
		fe.contents_read = true
	}
}

func (fe *file_entry) GetElementInParent() (eip *list.Element) {
	fe.ValidateParent()
	return fe.ElementInParent
}

func (fe *file_entry) GetParent() (parent *file_entry) {
	fe.ValidateParent()
	return fe.Parent
}

func (fe *file_entry) ValidateParent() {
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
				if e.Value.(*file_entry).Name == fe.Name {
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

func (fe *file_entry) is_root() bool {
	return fe.AbsPath == strings.Join([]string{filepath.VolumeName(fe.AbsPath), string(filepath.Separator)}, "")
}
