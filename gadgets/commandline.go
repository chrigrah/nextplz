package gadgets

import (
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
)

type CommandLine struct {
	X, Y int
	Length int
	FG, BG termbox.Attribute
	Cmd []byte
	Prefix string
	FillRune rune
}

func (cl *CommandLine) Draw(draw_cursor bool) {
	max_length := cl.Length - len(cl.Prefix) - 1
	cmd_length := util.Min(len(cl.Cmd), max_length)

	var cmd string = string(cl.Cmd[:cmd_length])
	util.WriteString(cl.X, cl.Y, cl.Length, cl.FG, cl.BG, cl.Prefix)
	util.WriteString_FillWithChar(cl.X+len(cl.Prefix), cl.Y, cl.Length, cl.FG, cl.BG, cmd, cl.FillRune)
	if draw_cursor {
		termbox.SetCursor(cl.X + len(cl.Prefix) + cmd_length, cl.Y)
	}
}

func (cl *CommandLine) append(char byte) {
	at := len(cl.Cmd)
	if cap(cl.Cmd) < at + 1 {	// Reallocate buffer with doubled capacity
		tmpArr := make([]byte, at + 1, at*2)
		copy(tmpArr, cl.Cmd)
		cl.Cmd = tmpArr
	} else {			// Increase length of slice to accomodate char
		cl.Cmd = cl.Cmd[:at+1]
	}
	cl.Cmd[at] = char
}

func (cl *CommandLine) Input(event termbox.Event) error {
	if event.Type == termbox.EventKey && event.Ch != 0 {
		cl.append(byte(event.Ch))
	} else {
		switch event.Key {
		case termbox.KeyBackspace:
			if len(cl.Cmd) > 0 {
				cl.Cmd = cl.Cmd[:len(cl.Cmd)-1]
			}
		case termbox.KeySpace:
			cl.append(' ')
		}
	}
	return nil
}
