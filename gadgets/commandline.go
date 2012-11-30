package gadgets

import (
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
)

type CommandLine struct {
	X, Y     int
	Length   int
	FG, BG   termbox.Attribute
	Cmd      []rune
	Prefix   string
	FillRune rune

	cursor_at int
}

func (cl *CommandLine) Draw(draw_cursor bool) {
	cmd_length := cl.get_visible_length()
	cmd := string(cl.Cmd[:cmd_length])
	util.WriteString(cl.X, cl.Y, cl.Length, cl.FG, cl.BG, cl.Prefix)
	util.WriteString_FillWithChar(cl.X+len(cl.Prefix), cl.Y, cl.Length, cl.FG, cl.BG, cmd, cl.FillRune)
	if draw_cursor {
		termbox.SetCursor(cl.X+len(cl.Prefix)+util.Min(cl.cursor_at, cmd_length), cl.Y)
	}
}

func (cl *CommandLine) append(character rune) {
	at := len(cl.Cmd)
	if cap(cl.Cmd) < at+1 {
		// Reallocate buffer with doubled capacity
		tmpArr := make([]rune, at+1, at*2)
		copy(tmpArr, cl.Cmd)
		cl.Cmd = tmpArr
	} else {
		// Increase length of slice to accomodate char
		cl.Cmd = cl.Cmd[:at+1]
	}
	copy(cl.Cmd[cl.cursor_at+1:], cl.Cmd[cl.cursor_at:])
	cl.Cmd[cl.cursor_at] = character
}

func (cl *CommandLine) Input(event termbox.Event) error {
	if event.Type == termbox.EventKey && event.Ch != 0 {
		cl.append(event.Ch)
		cl.step_cursor_right()
	} else {
		switch event.Key {
		case termbox.KeyBackspace2:
			fallthrough
		case termbox.KeyBackspace:
			if cl.cursor_at > 0 {
				copy(cl.Cmd[cl.cursor_at-1:], cl.Cmd[cl.cursor_at:])
				cl.Cmd = cl.Cmd[:len(cl.Cmd)-1]
				cl.step_cursor_left()
			}
		case termbox.KeyDelete:
			if len(cl.Cmd) > cl.cursor_at {
				copy(cl.Cmd[cl.cursor_at:], cl.Cmd[cl.cursor_at+1:])
				cl.Cmd = cl.Cmd[:len(cl.Cmd)-1]
			}
		case termbox.KeySpace:
			cl.append(' ')
			cl.step_cursor_right()
		case termbox.KeyArrowLeft:
			cl.step_cursor_left()
		case termbox.KeyArrowRight:
			cl.step_cursor_right()
		case termbox.KeyHome:
			cl.cursor_at = 0
		case termbox.KeyEnd:
			cl.cursor_at = len(cl.Cmd)
		}
	}
	return nil
}

func (cl *CommandLine) Clear() {
	cl.Cmd = cl.Cmd[0:0]
	cl.cursor_at = 0
}

func (cl *CommandLine) get_visible_length() int {
	max_length := cl.Length - len(cl.Prefix) - 1
	return util.Min(len(cl.Cmd), max_length)
}

func (cl *CommandLine) step_cursor_right() {
	if cl.cursor_at < len(cl.Cmd) {
		cl.cursor_at++
	}
}

func (cl *CommandLine) step_cursor_left() {
	if cl.cursor_at > 0 {
		cl.cursor_at--
	}
}
