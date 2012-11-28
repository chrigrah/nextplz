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
	max_length := cl.Length - len(cl.Prefix) - 1
	cmd_length := util.Min(len(cl.Cmd), max_length)

	var cmd string = string(cl.Cmd[:cmd_length])
	util.WriteString(cl.X, cl.Y, cl.Length, cl.FG, cl.BG, cl.Prefix)
	util.WriteString_FillWithChar(cl.X+len(cl.Prefix), cl.Y, cl.Length, cl.FG, cl.BG, cmd, cl.FillRune)
	if draw_cursor {
		termbox.SetCursor(cl.X+len(cl.Prefix)+cl.cursor_at, cl.Y)
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
		cl.cursor_at++
	} else {
		switch event.Key {
		case termbox.KeyBackspace:
			if cl.cursor_at > 0 {
				copy(cl.Cmd[cl.cursor_at-1:], cl.Cmd[cl.cursor_at:])
				cl.Cmd = cl.Cmd[:len(cl.Cmd)-1]
				if cl.cursor_at > 0 {
					cl.cursor_at--
				}
			}
		case termbox.KeyDelete:
			if len(cl.Cmd) > cl.cursor_at {
				copy(cl.Cmd[cl.cursor_at:], cl.Cmd[cl.cursor_at+1:])
				cl.Cmd = cl.Cmd[:len(cl.Cmd)-1]
			}
		case termbox.KeySpace:
			cl.append(' ')
			cl.cursor_at++
		case termbox.KeyArrowLeft:
			if cl.cursor_at > 0 {
				cl.cursor_at--
			}
		case termbox.KeyArrowRight:
			if cl.cursor_at < len(cl.Cmd) {
				cl.cursor_at++
			}
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
