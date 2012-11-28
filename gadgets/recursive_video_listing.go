package gadgets

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/chrigrah/nextplz/backend"
	MP "github.com/chrigrah/nextplz/media_player"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	//"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
	"regexp"
)

var (
	RecursiveListingIsOpen bool = false
	EnableFoldersForRars   bool = true

	cd_re, _ = regexp.Compile("(?i)(cd[0-9])")
)

type RecursiveListing struct {
	pl          PrintableListing
	video_files list.List
	update_chan chan int
	lock        sync.Mutex
	CL          CommandLine

	tick_id uint

	current_coloredstrings map[*backend.FileEntry]*backend.ColoredScrollingString
}

func InitRecursiveFromDirectory(dl *DirectoryListing, update_chan chan int) *RecursiveListing {
	var rl RecursiveListing
	rl.pl = PrintableListing{
		column_width: 80,
		startx:       dl.pl.startx,
		starty:       dl.pl.starty,
		width:        dl.pl.width,
		height:       dl.pl.height,
	}
	rl.current_coloredstrings = make(map[*backend.FileEntry]*backend.ColoredScrollingString)
	rl.pl.ElementToFilterValue = rl_elementtofiltervalue_func()
	rl.pl.ElementPrintValue = rl_elementprintvalue_func(&rl)
	rl.pl.header = fmt.Sprintf("Recursive listing of %s", dl.current_dir.AbsPath)
	rl.update_chan = update_chan

	rl.CL.X = rl.pl.startx
	rl.CL.Y = rl.pl.starty + rl.pl.height
	rl.CL.Length = rl.pl.width
	rl.CL.FG = termbox.ColorWhite
	rl.CL.BG = termbox.ColorBlack
	rl.CL.Cmd = make([]rune, 0, 8)
	rl.CL.FillRune = ' '
	rl.CL.Prefix = "> "

	rl.pl.UpdateFilter(&rl.video_files, string(rl.CL.Cmd))

	// Start dat funky recursion
	go func() {
		filepath.Walk(dl.current_dir.AbsPath, rl.get_walk_func())
	}()

	go func() {
		ticker := time.Tick(250 * time.Millisecond)
		for _ = range ticker {
			rl.lock.Lock()
			rl.tick_id++
			rl.lock.Unlock()
			rl.update_chan <- 1
		}
	}()

	RecursiveListingIsOpen = true
	return &rl
}

func rl_elementtofiltervalue_func() func(element interface{}) string {
	return func(element interface{}) string {
		entry := element.(*backend.FileEntry)
		top_folder := rl_fe_get_top_folder(entry)
		return fmt.Sprintf("(%s) %s", top_folder, entry.Name)
	}
}

func rl_elementprintvalue_func(rl *RecursiveListing) func(interface{}, int, int, int, bool) {
	return func(element interface{}, x, y int, width int, is_highlighted bool) {
		entry := element.(*backend.FileEntry)

		if cs, ok := rl.current_coloredstrings[entry]; ok {
			cs.Print(x, y, width, is_highlighted, is_highlighted, rl.tick_id)
		} else {
			new_cs := rl_fe_to_coloredstring(entry)
			new_cs.Print(x, y, width, is_highlighted, is_highlighted, rl.tick_id)
			rl.current_coloredstrings[entry] = new_cs
		}
	}
}

func rl_fe_to_coloredstring(entry *backend.FileEntry) (cs *backend.ColoredScrollingString) {
	cs = &backend.ColoredScrollingString{}
	cs.AppendString(entry.Name, termbox.ColorGreen)

	if EnableFoldersForRars && strings.HasSuffix(entry.Name, ".rar") {
		cs.AppendString(" (", termbox.ColorWhite)
		top_folder := rl_fe_get_top_folder(entry)
		cs.AppendString(top_folder, termbox.ColorCyan)
		cs.AppendString(")", termbox.ColorWhite)
	}
	return
}

func rl_fe_get_top_folder(entry *backend.FileEntry) (top_folder string) {
	parent_folders := strings.Split(entry.AbsPath, string(os.PathSeparator))
	top_folder = parent_folders[len(parent_folders)-2]
	if cd_re.MatchString(top_folder) && len(parent_folders) > 2 {
		top_folder = parent_folders[len(parent_folders)-3]
	}
	return
}

func (rl *RecursiveListing) Input(event termbox.Event) (err error) {
	switch event.Key {
	case termbox.KeyCtrlH:
		fallthrough
	case termbox.KeyArrowLeft:
		rl.pl.MoveCursorLeft()
	case termbox.KeyCtrlJ:
		fallthrough
	case termbox.KeyArrowDown:
		rl.pl.MoveCursorDown()
	case termbox.KeyCtrlK:
		fallthrough
	case termbox.KeyArrowUp:
		rl.pl.MoveCursorUp()
	case termbox.KeyCtrlL:
		fallthrough
	case termbox.KeyArrowRight:
		rl.pl.MoveCursorRight()
	case termbox.KeyCtrlB:
		file, ok := rl.pl.GetSelected()
		if ok {
			file_str := file.(*backend.FileEntry).AbsPath
			MP.GlobalMediaPlayer.PlayFile(file_str)
		} else {
			err = errors.New(fmt.Sprintf("Could not play file: Invalid selection"))
		}
	default:
		err = rl.CL.Input(event)
	}

	rl.pl.UpdateFilter(&rl.video_files, string(rl.CL.Cmd))

	return
}

func (rl *RecursiveListing) Finalize() IRStatus {
	return IRStatus{false, nil}
}

func (rl *RecursiveListing) HandleEscape() bool {
	if len(rl.CL.Cmd) > 0 {
		rl.CL.Clear()
		return true
	}
	return false
}

func (rl *RecursiveListing) Deactivate() error {
	RecursiveListingIsOpen = false
	// TODO: Shut down threads and that.
	return nil
}

func (rl *RecursiveListing) Draw(is_focused bool) error {
	rl.lock.Lock()
	defer rl.lock.Unlock()

	rl.pl.UpdateFilter(&rl.video_files, string(rl.CL.Cmd))
	rl.pl.PrintListing()

	rl.CL.Draw(is_focused)

	return nil
}

func (rl *RecursiveListing) SetFinalizeCallback(callback func(string) error) {
	// Doesn't finalize
}

func (rl *RecursiveListing) GetPrintableListing() *PrintableListing {
	return &rl.pl
}

func (rl *RecursiveListing) get_walk_func() filepath.WalkFunc {
	return filepath.WalkFunc(
		func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && backend.IsVideo(filepath.Base(path)) {
				rl.lock.Lock()

				var new_file = backend.FileEntry{
					Name:         filepath.Base(path),
					AbsPath:      path,
					IsDir:        false,
					IsAccessible: true,
					IsVideo:      true,
				}

				rl.video_files.PushBack(&new_file)
				rl.lock.Unlock()
				rl.update_chan <- 1
			}
			return nil
		})
}
