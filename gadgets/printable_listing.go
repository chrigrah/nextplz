package gadgets

import (
	"fmt"
	"strings"
	"regexp"
	"container/list"
	"github.com/chrigrah/nextplz/util"
	"github.com/chrigrah/nextplz/backend"
	"github.com/nsf/termbox-go"
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
	highlighted_index int

	startx, starty int
	width, height int
	column_width int
	rows, cols int
	col_at int
	filter_nomatch bool

	sb *util.ScrollingBoxes
}

func (pl *PrintableListing) PrintListing() int {
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
	//num_files := util.Min(pl.items.Len(), util.Min(max_cols, pl.cols) * pl.rows)
	start_at := pl.col_at * pl.rows

	var i int = 0
	for e := pl.items.Front(); i-start_at < available_boxes; {
		if i < start_at {
			i++
			e = e.Next()
			continue // Fast forward
		}

		effective_index := i - start_at

		if (e != nil) {
			var file *backend.FileEntry = e.Value.(*backend.FileEntry)

			pl.print_entry(
				effective_index % pl.rows, effective_index / pl.rows,
				file, e == pl.highlighted_element);

			e = e.Next()
		} else {
			col := effective_index / pl.rows
			row := effective_index % pl.rows
			str_width := util.Min(pl.column_width, pl.width - (col * pl.column_width))
			pl.sb.WriteString(
					uint16(col * pl.column_width), uint16(row + 1), str_width,
					termbox.ColorBlack, termbox.ColorBlack,	"", false);
		}
		i++
	}
	return -1
}

func (pl *PrintableListing) print_entry(row, col int, entry *backend.FileEntry, is_highlighted bool) {
	var fg, bg termbox.Attribute
	if is_highlighted {
		bg = termbox.ColorMagenta
	} else {
		bg = termbox.ColorBlack
	}

	if !entry.IsAccessible {
		fg = termbox.ColorRed
	} else if entry.IsVideo {
		fg = termbox.ColorGreen
	} else if entry.IsDir {
		fg = termbox.ColorCyan
	} else {
		fg = termbox.ColorWhite
	}

	room_left := pl.width - (col * pl.column_width)
	str_width := pl.column_width - 1
	if str_width < room_left {
		pl.sb.WriteString(
				uint16(col * pl.column_width), uint16(row + 1), str_width,
				fg, bg,	entry.Name, is_highlighted);
		
		// Fill in the blank spot between columns
		termbox.SetCell((col + 1) * pl.column_width - 1, row + 1, ' ', termbox.ColorBlack, termbox.ColorBlack)
	} else {
		pl.sb.WriteString(
				uint16(col * pl.column_width), uint16(row + 1), room_left,
				fg, bg,	entry.Name, is_highlighted);
	}
}

func (pl *PrintableListing) calc_start_column() (r int) {
	var num_visible_cols int
	var highlight_at_col int

	if pl.rows == 0 {
		return 0
	}

	num_visible_cols = int(pl.width / pl.column_width) - 1 // Fully visible columns
	if (pl.items.Len() / pl.rows) <= num_visible_cols {
		return 0
	}
	highlight_at_col = int(pl.highlighted_index / pl.rows)
	if highlight_at_col < pl.col_at {
		return highlight_at_col
	} else if highlight_at_col > pl.col_at + num_visible_cols {
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
		seen_old_highlight = seen_old_highlight || e.Value.(*backend.FileEntry) == old_selection
		matched := re.MatchString(e.Value.(*backend.FileEntry).Name)
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
		if element.Value.(*backend.FileEntry) == highlighted_entry {
			pl.highlighted_element = element
		}
	}
}

func (pl *PrintableListing) get_highlighted_entry(superset *list.List) (result *backend.FileEntry) {
	if superset.Len() == 0 {
		result = nil
	} else if pl.highlighted_element != nil {
		result = pl.highlighted_element.Value.(*backend.FileEntry)
	} else {
		result = superset.Front().Value.(*backend.FileEntry)
	}
	return
}

func (pl *PrintableListing) GetSelected() (abs_path string, ok bool) {
	return pl.highlighted_element.Value.(*backend.FileEntry).AbsPath, true
}

func (pl *PrintableListing) MoveCursorDown() {
	if pl.highlighted_element == nil { return; }
	next_element := pl.highlighted_element.Next()
	if next_element != nil {
		pl.highlighted_element = next_element
	}
}

func (pl *PrintableListing) MoveCursorUp() {
	if pl.highlighted_element == nil { return; }
	prev_element := pl.highlighted_element.Prev()
	if prev_element != nil {
		pl.highlighted_element = prev_element
	}
}

func (pl *PrintableListing) MoveCursorLeft() {
	if pl.highlighted_element == nil { return; }
	var element = pl.highlighted_element.Prev()

	for i := 0; i < pl.rows && element != nil; i++ {
		pl.highlighted_element = element
		element = element.Prev()
	}
}

func (pl *PrintableListing) MoveCursorRight() {
	if pl.highlighted_element == nil { return; }
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
