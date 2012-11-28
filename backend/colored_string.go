package backend

import (
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
)

const (
	pause_ticks = 3
)

type ColoredScrollingString struct {
	colors  []termbox.Attribute
	strings []string

	total_length int

	scroll_at          int
	edge_pause         int
	last_print_tick_id uint
}

func (cs *ColoredScrollingString) AppendString(str string, color termbox.Attribute) {
	cs.strings = append(cs.strings, str)
	cs.colors = append(cs.colors, color)
	cs.total_length += len(str)
}

func (cs *ColoredScrollingString) Print(x, y int, width int, scrolling, is_highlighted bool, tick_id uint) {
	if tick_id == cs.last_print_tick_id+1 && scrolling {
		cs.tick_scrolling(width)
	} else if tick_id != cs.last_print_tick_id || !scrolling {
		cs.reset()
	}
	cs.last_print_tick_id = tick_id

	bg := get_bg_color(is_highlighted)

	var_scroll := cs.scroll_at
	var i int = 0
	for ; i < len(cs.strings) && var_scroll > len(cs.strings[i]); i++ {
		var_scroll -= len(cs.strings[i])
	}
	var x_at = x
	if i < len(cs.strings) {
		first_string := cs.strings[i][var_scroll:]
		util.WriteString(x_at, y, width, cs.colors[i], bg, first_string)
		x_at += len(first_string)
		width -= len(first_string)
		i++
	}
	for ; i < len(cs.strings) && width > 0; i++ {
		util.WriteString(x_at, y, width, cs.colors[i], bg, cs.strings[i])
		x_at += len(cs.strings[i])
		width -= len(cs.strings[i])
	}
}

func (cs *ColoredScrollingString) reset() {
	cs.scroll_at = 0
	cs.edge_pause = 0
}

func (cs *ColoredScrollingString) tick_scrolling(width int) {
	if cs.total_length <= width {
		return
	}

	if cs.scroll_at == 0 {
		cs.edge_pause++
		if cs.edge_pause > pause_ticks {
			cs.edge_pause = 0
			cs.scroll_at = 1
		}
	} else {
		end_diff := cs.scroll_at + width - cs.total_length
		if end_diff == 0 {
			cs.edge_pause++
			if cs.edge_pause > pause_ticks {
				cs.edge_pause = 0
				cs.scroll_at = 0
			}
		} else {
			cs.scroll_at++
		}
	}
}

func get_bg_color(is_highlighted bool) (bg termbox.Attribute) {
	if is_highlighted {
		bg = termbox.ColorMagenta
	} else {
		bg = termbox.ColorBlack
	}
	return
}
