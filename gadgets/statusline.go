package gadgets

import (
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
)

type StatusLine struct {
	Status string
	X, Y   int
	Length int
	FG, BG termbox.Attribute
}

func (sl *StatusLine) Draw() {
	util.WriteString(sl.X, sl.Y, sl.Length, sl.FG, sl.BG, sl.Status)
}

func (sl *StatusLine) ShowUpdate(update string) {
	sl.Status = update
	sl.FG = termbox.ColorWhite
	sl.BG = termbox.ColorBlue
}

func (sl *StatusLine) ShowError(err error) {
	sl.Status = err.Error()
	sl.FG = termbox.ColorWhite
	sl.BG = termbox.ColorRed
}
