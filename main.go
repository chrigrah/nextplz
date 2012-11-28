package main

import (
	"container/list"
	"flag"
	"github.com/chrigrah/nextplz/backend"
	"github.com/chrigrah/nextplz/gadgets"
	MP "github.com/chrigrah/nextplz/media_player"
	"github.com/nsf/termbox-go"
	"strings"
)

var (
	width, height int
	dl            *gadgets.DirectoryListing
	sl            gadgets.StatusLine
	at_state      int
	events        chan termbox.Event = make(chan termbox.Event, 10)
	update_chan   chan int           = make(chan int, 10)
	focus_stack   *list.List

	media_extensions string
)

func main() {
	var err error
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.Clear(termbox.ColorBlack, termbox.ColorBlack)

	initialize_globals()
	defer MP.GlobalMediaPlayer.Disconnect()

	go feed_events()

	update()

	for {
		select {
		case event := <-events:
			if event.Type != termbox.EventKey {
				continue
			}
			switch event.Key {
			case termbox.KeyEsc:
				handled := focus_stack.Front().Value.(gadgets.InputReceiver).HandleEscape()
				if !handled {
					err = focus_stack.Front().Value.(gadgets.InputReceiver).Deactivate()
					focus_stack.Remove(focus_stack.Front())

					if focus_stack.Len() <= 0 {
						return
					}
				}
			case termbox.KeyEnter:
				status := focus_stack.Front().Value.(gadgets.InputReceiver).Finalize()

				if status.Done {
					err = focus_stack.Front().Value.(gadgets.InputReceiver).Deactivate()
					focus_stack.Remove(focus_stack.Front())
				}
				if status.Chain != nil {
					err = status.Chain
				}
			case termbox.KeyF1:
				if !gadgets.TextBoxIsOpen {
					tb, err := gadgets.CreateTextBox("Change directory:", width, height)
					if err != nil {
						display_error(err)
						continue
					}
					tb.X = width/2 - tb.Width/2
					tb.Y = height/2 - tb.Height/2
					tb.FinalizeCallback = func(dir string) error { return dl.ChangeDirectory(dir) }
					focus_stack.PushFront(tb)
				}
			case termbox.KeyF4:
				rl := gadgets.InitRecursiveFromDirectory(dl, update_chan)
				focus_stack.PushFront(rl)
			case termbox.KeyCtrlSpace:
				err = MP.GlobalMediaPlayer.(*MP.VLC).Pause()
			}

			err = focus_stack.Front().Value.(gadgets.InputReceiver).Input(event)

			display_error(err)
			update()

		case <-update_chan:
			update()
		}
	}
}

func feed_events() {
	for {
		events <- termbox.PollEvent()
	}
}

func initialize_globals() {
	mp_info := MP.InitMediaPlayerFlagParser()
	flag.StringVar(&media_extensions, "extensions", ".avi,.mkv,.mpg,.wmv",
		"Comma separated list of file extensions that should be considered video files.")
	flag.IntVar(&gadgets.LS_COL_WIDTH, "cw", 50, "Column width for directory listing.")
	flag.BoolVar(&backend.FilterSubs, "filter-subs", true,
		"If set to true, rar files matching [.-]subs[.-] will be filtered out from recursive listings.")
	flag.BoolVar(&backend.FilterSamples, "filter-samples", true,
		"If set to true, video files matching [.-]sample[.-] will be filtered out from recursive listings.")
	flag.BoolVar(&gadgets.EnableFoldersForRars, "rar-folders", true,
		"If set to true rar files will also be filtered by folder in recursive listings")
	flag.Parse()

	width, height = termbox.Size()
	dl = gadgets.NewListing(0, 0, width, height-1, update_chan)

	backend.VideoExtensions = strings.Split(media_extensions, ",")
	var err error
	MP.GlobalMediaPlayer, err = mp_info.CreateMediaPlayer()
	if err != nil {
		panic(err)
	}

	sl.X = 0
	sl.Y = height - 1
	sl.Length = width

	focus_stack = list.New()
	focus_stack.PushFront(dl)
}

func update() {
	//termbox.Clear(termbox.ColorBlack, termbox.ColorBlack)

	for drawable := focus_stack.Back(); drawable != nil; drawable = drawable.Prev() {
		is_focused := drawable.Prev() == nil
		err := drawable.Value.(gadgets.InputReceiver).Draw(is_focused)
		display_error(err)
	}

	termbox.Flush()
}

func display_error(err error) {
	if err != nil {
		sl.ShowError(err)
	}
}
