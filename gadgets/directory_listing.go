package gadgets

import (
	"errors"
	"fmt"
	"github.com/chrigrah/nextplz/backend"
	MP "github.com/chrigrah/nextplz/media_player"
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
	"os"
)

var (
	LS_COL_WIDTH int = 50 // Available as a command line flag - then it's const :)
)

type DirectoryListing struct {
	current_dir *backend.FileEntry

	pl PrintableListing
	CL CommandLine
	sb util.ScrollingBoxes

	FinalizeCallback func(string) error
	Debug_message    string
}

func NewListing(startx, starty int, width, height int, update_chan chan int) *DirectoryListing {
	directoryName, err := os.Getwd()
	panic_perhaps(err)
	cwd, err := backend.CreateDirEntry(directoryName)
	panic_perhaps(err)

	dl := &DirectoryListing{
		current_dir: cwd,
		pl: PrintableListing{
			column_width: LS_COL_WIDTH,
			startx:       startx,
			starty:       starty,
			width:        width,
			height:       height - 1,
		},
	}
	dl.pl.sb = &dl.sb
	dl.FinalizeCallback = func(_ string) error {
		dl.CdHighlighted()
		dl.CL.Cmd = dl.CL.Cmd[0:0]
		return nil
	}

	dl.CL.X = startx
	dl.CL.Y = starty + height - 1
	dl.CL.Length = width
	dl.CL.FG = termbox.ColorWhite
	dl.CL.BG = termbox.ColorBlack
	dl.CL.Cmd = make([]byte, 0, 8)
	dl.CL.FillRune = ' '
	dl.CL.Prefix = "> "

	dl.pl.UpdateFilter(&dl.current_dir.Contents, string(dl.CL.Cmd))
	dl.sb.UpdateChan = update_chan
	go dl.sb.StartTicker()

	return dl
}

func (dl *DirectoryListing) Input(event termbox.Event) (err error) {
	switch event.Key {
	case termbox.KeyF5:
		dl.ChangeDirectory(dl.current_dir.AbsPath)
	case termbox.KeyBackspace2:
		err = dl.CdUp()
		dl.CL.Cmd = dl.CL.Cmd[0:0]
	case termbox.KeyCtrlH:
		fallthrough
	case termbox.KeyArrowLeft:
		dl.pl.MoveCursorLeft()
	case termbox.KeyCtrlJ:
		fallthrough
	case termbox.KeyArrowDown:
		dl.pl.MoveCursorDown()
	case termbox.KeyCtrlK:
		fallthrough
	case termbox.KeyArrowUp:
		dl.pl.MoveCursorUp()
	case termbox.KeyCtrlL:
		fallthrough
	case termbox.KeyArrowRight:
		dl.pl.MoveCursorRight()
	case termbox.KeyCtrlN:
		err = dl.NextDirectory()
	case termbox.KeyCtrlP:
		err = dl.PrevDirectory()
	case termbox.KeyCtrlB:
		file, ok := dl.pl.GetSelected()
		if ok {
			MP.GlobalMediaPlayer.PlayFile(file)
		} else {
			err = errors.New(fmt.Sprintf("Could not play file: Invalid selection"))
		}
	default:
		err = dl.CL.Input(event)
	}

	dl.pl.UpdateFilter(&dl.current_dir.Contents, string(dl.CL.Cmd))

	return
}

func (dl *DirectoryListing) SetFinalizeCallback(callback func(string) error) {
	dl.FinalizeCallback = callback
}

func (dl *DirectoryListing) Finalize() IRStatus {
	return IRStatus{false, dl.FinalizeCallback(string(dl.CL.Cmd))}
}

func (dl *DirectoryListing) HandleEscape() bool {
	if len(dl.CL.Cmd) > 0 {
		dl.CL.Cmd = dl.CL.Cmd[0:0]
		return true
	}
	return false
}

func (dl *DirectoryListing) Deactivate() error {
	return nil
}

func (dl *DirectoryListing) ChangeDirectory(dir string) error {
	cwd, err := backend.CreateDirEntry(dir)
	if err != nil {
		return err
	}

	dl.current_dir = cwd
	dl.pl = PrintableListing{
		column_width: LS_COL_WIDTH,
		startx:       dl.pl.startx,
		starty:       dl.pl.starty,
		width:        dl.pl.width,
		height:       dl.pl.height,
		sb:           dl.pl.sb,
	}
	return nil
}

func (dl *DirectoryListing) Draw(is_focused bool) error {
	if dl.Debug_message != "" {
		dl.pl.header = dl.Debug_message
	} else {
		dl.pl.header = dl.current_dir.AbsPath
	}

	dl.CL.Draw(is_focused)
	dl.pl.PrintListing()

	dl.sb.Draw()

	return nil
}

func (dl *DirectoryListing) GetPrintableListing() *PrintableListing {
	return &dl.pl
}

func (dl *DirectoryListing) CdHighlighted() error {
	if dl.pl.highlighted_element == nil {
		return errors.New("No entry is highlighted. Is this an empty folder? Helloooooo....")
	}
	highlighted_entry := dl.pl.highlighted_element.Value.(*backend.FileEntry)
	if !highlighted_entry.IsDir {
		return errors.New("Highlighted entry is not a directory")
	}
	return dl.ChangeDir(highlighted_entry)
}

func (dl *DirectoryListing) CdUp() error {
	return dl.ChangeDir(dl.current_dir.GetParent())
}

func (dl *DirectoryListing) ChangeDir(dir *backend.FileEntry) error {
	err := dir.ValidateContents()
	if err != nil {
		dir.IsAccessible = false
		return err
	}
	dl.current_dir = dir
	dl.pl.highlighted_element = nil
	return nil
}

func (dl *DirectoryListing) PrevDirectory() error {
	for element := dl.current_dir.GetElementInParent().Prev(); element != nil; element = element.Prev() {
		at_entry := element.Value.(*backend.FileEntry)
		if !at_entry.IsDir || !at_entry.IsAccessible {
			continue
		}

		return dl.ChangeDir(element.Value.(*backend.FileEntry))
	}
	return errors.New("No previous directory")
}

func (dl *DirectoryListing) NextDirectory() error {
	for element := dl.current_dir.GetElementInParent().Next(); element != nil; element = element.Next() {
		at_entry := element.Value.(*backend.FileEntry)
		if !at_entry.IsDir || !at_entry.IsAccessible {
			continue
		}

		return dl.ChangeDir(element.Value.(*backend.FileEntry))
	}
	return nil
}

func panic_perhaps(err error) {
	if err != nil {
		panic(err)
	}
}
