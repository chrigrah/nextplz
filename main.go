package main

import (
	"fmt"
	MC "github.com/chrigrah/nextplz/media_commands"
	DL "github.com/chrigrah/nextplz/directory_listing"
	CL "github.com/chrigrah/nextplz/commandline"
	"github.com/nsf/termbox-go"
)

var (
	ls *DL.Listing
	cl CL.CommandLine
	at_state int
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic("Could not init termbox. ")
	}
	defer termbox.Close()

	width, height := termbox.Size()
	ls = DL.NewListing(0, 0, width, height - 1)

	cl.X = 0; cl.Y = height-1;
	cl.Length = width;
	cl.FG = termbox.ColorWhite; cl.BG = termbox.ColorBlack;
	cl.Cmd = make([]byte, 0, 8)
	cl.Prefix = "> "

	update()

	for event := termbox.PollEvent(); true; event = termbox.PollEvent() {
		if event.Type != termbox.EventKey {
			continue
		}
		switch event.Key {
		case termbox.KeyEsc:
			if len(cl.Cmd) > 0 {
				cl.Cmd = cl.Cmd[0:0]
				update()
			} else {
				return
			}
		case termbox.KeyEnter:
			ls.CdHighlighted()
			cl.Cmd = cl.Cmd[0:0]
			update()
		case termbox.KeyBackspace2:
			ls.CdUp()
			cl.Cmd = cl.Cmd[0:0]
			update()
		case termbox.KeyCtrlH: fallthrough
		case termbox.KeyArrowLeft:
			ls.MoveCursorLeft()
			update()
		case termbox.KeyCtrlJ: fallthrough
		case termbox.KeyArrowDown:
			ls.MoveCursorDown()
			update()
		case termbox.KeyCtrlK: fallthrough
		case termbox.KeyArrowUp:
			ls.MoveCursorUp()
			update()
		case termbox.KeyCtrlL: fallthrough
		case termbox.KeyArrowRight:
			ls.MoveCursorRight()
			update()
		case termbox.KeyCtrlN:
			ls.NextDirectory()
			update()
		case termbox.KeyCtrlP:
			ls.PrevDirectory()
			update()
		case termbox.KeyCtrlB:
			file, ok := ls.GetSelected()
			if ok {
				MC.PlayFile(file)
			} else {
				fmt.Println(file)
			}
		}

		if textChanged := cl.Input(event); textChanged {
			update()
		}
	}
}

func update() {
	// There's error checking to be done here... 
	termbox.Clear(termbox.ColorBlack, termbox.ColorBlack)
	ls.UpdateFilter(string(cl.Cmd))
	ls.PrintDirectoryListing()
	cl.Draw()
	termbox.Flush()
}
