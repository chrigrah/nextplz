package gadgets

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
	"regexp"
	"strings"
)

type Listing interface {
	UpdateFilter(input string)
	PrintListing() int
	GetPrintableListing() *PrintableListing
}

type PrintableListing struct {
	header string

	items list.List

	highlighted_element *list.Element
	highlighted_index   int

	startx, starty int
	width, height  int
	column_width   int
	rows, cols     int
	col_at         int
	filter_nomatch bool

	ElementToFilterValue func(interface{}) string
	ElementPrintValue    func(element interface{}, x, y int, width int, is_highlighted bool)
}

func (pl *PrintableListing) PrintListing() {
	var dir_header_fg termbox.Attribute
	if pl.filter_nomatch {
		dir_header_fg = termbox.ColorRed
	} else {
		dir_header_fg = termbox.ColorWhite
	}

	pl.col_at = pl.calc_start_column()
	util.WriteString(0, 0, pl.width, dir_header_fg, termbox.ColorBlue, pl.header)
	pl.rows = pl.height - 1
	pl.cols = (pl.items.Len() / pl.height) + 1
	max_cols := (pl.width / pl.column_width) + 1
	available_boxes := max_cols * pl.rows
	start_at := pl.col_at * pl.rows

	var i int = 0
	for e := pl.items.Front(); i-start_at < available_boxes; {
		if i < start_at {
			i++
			if e == nil {
				panic(errors.New(fmt.Sprintf("col_at: %d", pl.col_at)))
			}
			e = e.Next()
			continue // Fast forward
		}

		effective_index := i - start_at

		if e != nil {
			pl.print_entry(
				effective_index%pl.rows, effective_index/pl.rows,
				e, e == pl.highlighted_element)

			e = e.Next()
		} else {
			col := effective_index / pl.rows
			row := effective_index % pl.rows
			str_width := util.Min(pl.column_width, pl.width-(col*pl.column_width))
			util.WriteString(
				col*pl.column_width, row+1, str_width,
				termbox.ColorBlack, termbox.ColorBlack, "")
		}
		i++
	}
}

func (pl *PrintableListing) print_entry(row, col int, entry *list.Element, is_highlighted bool) {
	room_left := pl.width - (col * pl.column_width)
	if pl.column_width < room_left {
		pl.ElementPrintValue(entry.Value, col*pl.column_width, row+1, pl.column_width-1, is_highlighted)
		termbox.SetCell((col+1)*pl.column_width-1, row+1, ' ', termbox.ColorBlack, termbox.ColorBlack)
	} else {
		pl.ElementPrintValue(entry.Value, col*pl.column_width, row+1, room_left, is_highlighted)
	}

}

func (pl *PrintableListing) calc_start_column() (r int) {
	var num_visible_cols int
	var highlight_at_col int

	if pl.rows == 0 {
		return 0
	}

	num_visible_cols = util.Max(int(pl.width/pl.column_width)-1, 0) // Fully visible columns, at least one
	if (pl.items.Len() / pl.rows) <= num_visible_cols {
		return 0
	}
	highlight_at_col = int(pl.highlighted_index / pl.rows)
	if highlight_at_col < pl.col_at {
		return highlight_at_col
	} else if highlight_at_col > pl.col_at+num_visible_cols {
		return highlight_at_col - num_visible_cols
	}
	return pl.col_at
}

func (pl *PrintableListing) UpdateFilter(superset *list.List, input string) {
	if superset.Len() == 0 {
		return // Special case
	}

	pattern := create_pattern_from_input(input)
	regexp, err := regexp.Compile(pattern)
	panic_perhaps(err)

	pl.select_and_highlight(superset, regexp)

	if pl.items.Len() == 0 {
		pl.select_all(superset)
		pl.filter_nomatch = true
	} else {
		pl.filter_nomatch = false
	}
}

func (pl *PrintableListing) select_and_highlight(superset *list.List, re *regexp.Regexp) {
	var finalized_highlight bool = false
	var seen_old_highlight bool = false
	var i int = 0
	pl.items = list.List{}

	old_selection := pl.get_highlighted_entry(superset)

	for e := superset.Front(); e != nil; e = e.Next() {
		seen_old_highlight = seen_old_highlight || e.Value == old_selection
		matched := re.MatchString(pl.ElementToFilterValue(e.Value))
		if matched {
			new_select_element := pl.items.PushBack(e.Value)

			if !seen_old_highlight {
				pl.highlighted_element = new_select_element
				pl.highlighted_index = i
			} else if !finalized_highlight {
				pl.highlighted_element = new_select_element
				pl.highlighted_index = i
				finalized_highlight = true
			}
			i++
		}
	}
	return
}

func (pl *PrintableListing) select_all(superset *list.List) {
	highlighted_entry := pl.get_highlighted_entry(superset)
	for e := superset.Front(); e != nil; e = e.Next() {
		element := pl.items.PushBack(e.Value) // Important that element is that of ls.selection
		if element.Value == highlighted_entry {
			pl.highlighted_element = element
		}
	}
}

func (pl *PrintableListing) get_highlighted_entry(superset *list.List) (result interface{}) {
	if superset.Len() == 0 {
		result = nil
	} else if pl.highlighted_element != nil {
		result = pl.highlighted_element.Value
	} else {
		result = superset.Front().Value
	}
	return
}

func (pl *PrintableListing) GetSelected() (selected interface{}, ok bool) {
	return pl.highlighted_element.Value, true
}

func (pl *PrintableListing) MoveCursorDown() {
	if pl.highlighted_element == nil {
		return
	}
	next_element := pl.highlighted_element.Next()
	if next_element != nil {
		pl.highlighted_element = next_element
	}
}

func (pl *PrintableListing) MoveCursorUp() {
	if pl.highlighted_element == nil {
		return
	}
	prev_element := pl.highlighted_element.Prev()
	if prev_element != nil {
		pl.highlighted_element = prev_element
	}
}

func (pl *PrintableListing) MoveCursorLeft() {
	if pl.highlighted_element == nil {
		return
	}
	var element = pl.highlighted_element.Prev()

	for i := 0; i < pl.rows && element != nil; i++ {
		pl.highlighted_element = element
		element = element.Prev()
	}
}

func (pl *PrintableListing) MoveCursorRight() {
	if pl.highlighted_element == nil {
		return
	}
	var element = pl.highlighted_element.Next()

	for i := 0; i < pl.rows && element != nil; i++ {
		pl.highlighted_element = element
		element = element.Next()
	}
}

func create_pattern_from_input(input string) string {
	input = regexp.QuoteMeta(input)
	input = strings.Replace(input, " ", "(.*)", -1)
	input = fmt.Sprintf("(?i)(%s)", input)
	return input
}
