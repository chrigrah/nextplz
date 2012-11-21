package gadgets

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/chrigrah/nextplz/backend"
	MP "github.com/chrigrah/nextplz/media_player"
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
	"os"
	"path/filepath"
	"sync"
)

var (
	RecursiveListingIsOpen bool = false
)

type RecursiveListing struct {
	pl          PrintableListing
	video_files list.List
	update_chan chan int
	lock        sync.Mutex
	CL          CommandLine
	sb          util.ScrollingBoxes
}

func InitRecursiveFromDirectory(dl *DirectoryListing, update_chan chan int) *RecursiveListing {
	var rl RecursiveListing
	rl.pl = PrintableListing{
		column_width: 80,
		startx:       dl.pl.startx,
		starty:       dl.pl.starty,
		width:        dl.pl.width,
		height:       dl.pl.height,
		sb:           &rl.sb,
	}
	rl.pl.header = fmt.Sprintf("Recursive listing of %s", dl.current_dir.AbsPath)
	rl.update_chan = update_chan

	rl.CL.X = rl.pl.startx
	rl.CL.Y = rl.pl.starty + rl.pl.height
	rl.CL.Length = rl.pl.width
	rl.CL.FG = termbox.ColorWhite
	rl.CL.BG = termbox.ColorBlack
	rl.CL.Cmd = make([]byte, 0, 8)
	rl.CL.FillRune = ' '
	rl.CL.Prefix = "> "

	rl.pl.UpdateFilter(&rl.video_files, string(rl.CL.Cmd))

	// Start dat funky recursion
	go func() {
		filepath.Walk(dl.current_dir.AbsPath, rl.get_walk_func())
	}()

	rl.sb.UpdateChan = update_chan
	go rl.sb.StartTicker()

	RecursiveListingIsOpen = true
	return &rl
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
			MP.GlobalMediaPlayer.PlayFile(file)
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
		rl.CL.Cmd = rl.CL.Cmd[0:0]
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

	rl.sb.Draw()
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
