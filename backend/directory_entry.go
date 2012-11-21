package backend

import (
	"os"
	"strings"
	"regexp"
	"path/filepath"
	"container/list"
)

var (
	VideoExtensions []string
	CoddleRars bool = true
	FilterSubs bool = true
	FilterSamples bool = true
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
			IsVideo: IsVideo(filepath.Base(dir)),
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

var sample_regexp *regexp.Regexp
var rar_part_regexp *regexp.Regexp
var rar_sub_regexp *regexp.Regexp

func IsVideo(path string) bool {
	if FilterSamples && sample_regexp == nil {
		var err error
		sample_regexp, err = regexp.Compile("(?:^|[.-])(?i)(?:sample)[.-]")
		if err != nil { panic(err); }
	}
	for _, ext := range VideoExtensions {
		if strings.HasSuffix(path, ext) && (!FilterSamples || !sample_regexp.MatchString(path)) {
			return true
		}
	}
	if CoddleRars && strings.HasSuffix(path, ".rar") {
		if rar_part_regexp == nil {
			var err error
			rar_part_regexp, err = regexp.Compile("\\.part([0-9]{2})\\.rar$")
			if err != nil { panic(err); }
		}
		if FilterSubs && rar_sub_regexp == nil {
			var err error
			rar_sub_regexp, err = regexp.Compile("(?:^|[.-])subs[.-]")
			if err != nil { panic(err); }
		}
		if matches := rar_part_regexp.FindStringSubmatch(path); len(matches) == 2 && matches[1] != "01" {
			return false
		}
		if FilterSubs && rar_sub_regexp.MatchString(path) {
			return false
		}

		return true
	}
	return false
}

func (fe *FileEntry) is_root() bool {
	return fe.AbsPath == strings.Join([]string{filepath.VolumeName(fe.AbsPath), string(filepath.Separator)}, "")
}
