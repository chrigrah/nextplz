package gadgets

import (
	"errors"
	"fmt"
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
)

type IRStatus struct {
	Done  bool
	Chain error
}

type InputReceiver interface {
	Input(termbox.Event) error
	Finalize() IRStatus
	HandleEscape() bool

	Draw(is_focused bool) error
	//Insert() error
	Deactivate() error

	SetFinalizeCallback(func(string) error)
}

var (
	TextBoxIsOpen bool = false
)

type TextBox struct {
	question      string
	X, Y          int
	Width, Height int
	cl            *CommandLine

	FinalizeCallback func(string) error
}

const (
	horizontal_overhead int = 4
	vertical_overhead   int = 6 // borders and inputbox
	box_minheight       int = vertical_overhead + 1
	comfortable_width       = 30
)

func (tb *TextBox) SetFinalizeCallback(callback func(string) error) {
	tb.FinalizeCallback = callback
}

func (tb *TextBox) Finalize() IRStatus {
	err := tb.FinalizeCallback(string(tb.cl.Cmd))
	tb.cl.Cmd = tb.cl.Cmd[0:0]

	TextBoxIsOpen = false
	return IRStatus{true, err}
}

func (tb *TextBox) HandleEscape() bool {
	if len(tb.cl.Cmd) > 0 {
		tb.cl.Cmd = tb.cl.Cmd[0:0]
		return true
	}
	return false
}

func (tb *TextBox) Deactivate() error {
	TextBoxIsOpen = false
	return nil
}

func CreateTextBox(question string, maxwidth, maxheight int) (*TextBox, error) {
	if maxheight < box_minheight {
		return nil, errors.New(fmt.Sprintf("TextBoxes need at least %d rows", box_minheight))
	}
	var tb TextBox
	tb.question = question

	startx := tb.X + 2
	stopx := tb.X + tb.Width - 2
	effective_width := stopx - startx
	question_rows := (effective_width / len(tb.question)) + 1
	needed_rows := question_rows + vertical_overhead
	if needed_rows > maxheight {
		return nil, errors.New(fmt.Sprintf(
			"Not enough room provided for textbox. maxwidth:%d maxheight:%d needed_rows:%d",
			maxwidth, maxheight, needed_rows))
	}
	if question_rows > 1 {
		tb.Width = maxwidth
	} else {
		tb.Width = len(tb.question) + horizontal_overhead
		if tb.Width < comfortable_width {
			tb.Width = util.Min(maxwidth, comfortable_width)
		}
	}
	tb.Height = question_rows + vertical_overhead

	tb.cl = &CommandLine{
		Length:   tb.Width - 4,
		FG:       termbox.ColorWhite,
		BG:       termbox.ColorRed,
		Prefix:   "",
		FillRune: '_',
		Cmd:      make([]byte, 0, 8),
	}

	TextBoxIsOpen = true
	return &tb, nil
}

func (tb *TextBox) Input(ev termbox.Event) error {
	return tb.cl.Input(ev)
}

func (tb *TextBox) Draw(is_focused bool) error {
	tb.cl.X = tb.X + 2
	tb.cl.Y = tb.Y + tb.Height - 3

	tb.draw_borders()
	tb.fill()
	tb.draw_question()
	//	tb.draw_input()
	tb.cl.Draw(is_focused)

	return nil
}

func (tb *TextBox) draw_borders() {
	termbox.SetCell(tb.X, tb.Y, '+', termbox.ColorWhite, termbox.ColorBlue)
	termbox.SetCell(tb.X+tb.Width-1, tb.Y, '+', termbox.ColorWhite, termbox.ColorBlue)
	termbox.SetCell(tb.X, tb.Y+tb.Height-1, '+', termbox.ColorWhite, termbox.ColorBlue)
	termbox.SetCell(tb.X+tb.Width-1, tb.Y+tb.Height-1, '+', termbox.ColorWhite, termbox.ColorBlue)
	util.RepeatCharX(tb.X+1, tb.X+tb.Width-1, tb.Y, '-', termbox.ColorWhite, termbox.ColorBlue)
	util.RepeatCharX(tb.X+1, tb.X+tb.Width-1, tb.Y+tb.Height-1, '-', termbox.ColorWhite, termbox.ColorBlue)
	util.RepeatCharY(tb.Y+1, tb.Y+tb.Height-1, tb.X, '|', termbox.ColorWhite, termbox.ColorBlue)
	util.RepeatCharY(tb.Y+1, tb.Y+tb.Height-1, tb.X+tb.Width-1, '|', termbox.ColorWhite, termbox.ColorBlue)
}

func (tb *TextBox) fill() {
	for y := tb.Y + 1; y < tb.Y+tb.Height-1; y++ {
		util.FillLineTo(tb.X+1, y, tb.X+tb.Width-1, termbox.ColorBlue)
	}
}

func (tb *TextBox) draw_question() {
	startx := tb.X + 2
	starty := tb.Y + 2
	stopx := tb.X + tb.Width - 2
	line_width := stopx - startx
	question_barray := []byte(tb.question)

	for i := starty; len(question_barray) > 0; i++ {
		str_len := util.Min(line_width, len(question_barray))
		str := string(question_barray[:str_len])
		util.WriteString(startx, i, line_width, termbox.ColorWhite, termbox.ColorBlue, str)
		question_barray = question_barray[str_len:]
	}
}
