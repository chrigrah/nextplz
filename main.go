package main

import (
	"fmt"
	"flag"
	"strings"
	MP "github.com/chrigrah/nextplz/media_player"
	"github.com/chrigrah/nextplz/gadgets"
	"github.com/chrigrah/nextplz/util"
	"github.com/chrigrah/nextplz/backend"
	"github.com/nsf/termbox-go"
	"container/list"
)

var (
	width, height int
	mp MP.MediaPlayer
	ls *gadgets.Listing
	tb *gadgets.TextBox
	sl gadgets.StatusLine
	cl gadgets.CommandLine
	sb util.ScrollingBoxes
	at_state int

	events chan termbox.Event = make(chan termbox.Event, 10)
	focus_stack *list.List

	media_extensions string
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	initialize_globals()
	defer mp.Disconnect()

	go sb.StartTicker()
	go feed_events()

	update()

	for { select {
	case event := <-events:
		if event.Type != termbox.EventKey {
			continue
		}
		switch event.Key {
		case termbox.KeyEsc:
			if tb != nil {
				focus_stack.Remove(focus_stack.Front())
				tb = nil
				update()
			} else if len(cl.Cmd) > 0 {
				cl.Cmd = cl.Cmd[0:0]
				update()
			} else {
				return
			}
		case termbox.KeyEnter:
			err := focus_stack.Front().Value.(gadgets.InputReceiver).Finalize()
			if err != nil {
				display_error(err)
			} else {
				if tb != nil {
					focus_stack.Remove(focus_stack.Front())
					tb = nil
					update()
				}
			}
			update()
		case termbox.KeyBackspace2:
			err := ls.CdUp()
			display_error(err)
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
			err := ls.NextDirectory()
			display_error(err)
			update()
		case termbox.KeyCtrlP:
			err := ls.PrevDirectory()
			display_error(err)
			update()
		case termbox.KeyCtrlB:
			file, ok := ls.GetSelected()
			if ok {
				mp.PlayFile(file)
			} else {
				fmt.Println(file)
			}
		case termbox.KeyF1:
			if tb == nil {
				var err error
				tb, err = gadgets.CreateTextBox("Change directory:", width, height)
				if err != nil {
					display_error(err)
					continue
				}
				tb.X = width / 2 - tb.Width / 2
				tb.Y = height / 2 - tb.Height / 2
				tb.FinalizeCallback = func(dir string) error { return ls.ChangeDirectory(dir); }
				focus_stack.PushFront(tb)
				update()
			}
		case termbox.KeyCtrlSpace:
			err := mp.(*MP.VLC).Pause()
			display_error(err)
		}

		focus_stack.Front().Value.(gadgets.InputReceiver).Input(event)
		update()

	case <-sb.UpdateTicks:
		update()
	}}
}

func feed_events() {
	for {
		events <- termbox.PollEvent()
	}
}

func initialize_globals() {
	mp_info := MP.InitMediaPlayerFlagParser()
	flag.StringVar(&media_extensions, "extensions", ".avi,.mkv,.rar,.mpg,.wmv", "Comma separated list of file extensions that should be considered video files.")
	flag.Parse()

	width, height = termbox.Size()
	ls = gadgets.NewListing(0, 0, width, height - 2, &sb)
	backend.VideoExtensions = strings.Split(media_extensions, ",")
	var err error
	mp, err = mp_info.CreateMediaPlayer()
	if err != nil {
		panic(err)
	}

	sl.X = 0; sl.Y = height-2;
	sl.Length = width

	cl.X = 0; cl.Y = height-1;
	cl.Length = width;
	cl.FG = termbox.ColorWhite; cl.BG = termbox.ColorBlack;
	cl.Cmd = make([]byte, 0, 8)
	cl.FillRune = ' '
	cl.Prefix = "> "
	cl.FinalizeCallback = func(_ string) error { ls.CdHighlighted(); cl.Cmd = cl.Cmd[0:0]; return nil; }

	focus_stack = list.New()
	focus_stack.PushFront(&cl)
}

func update() {
	// There's error checking to be done here... 
	termbox.Clear(termbox.ColorBlack, termbox.ColorBlack)
	ls.UpdateFilter(string(cl.Cmd))
	ls.PrintDirectoryListing()
	sb.WriteAll()
	cl.Draw()
	sl.Draw()
	if tb != nil {
		tb.Draw()
	}
	termbox.Flush()
}

func display_error(err error) {
	if err != nil {
		sl.ShowError(err)
	}
}
