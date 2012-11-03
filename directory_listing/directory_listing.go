package directory_listing

import (
	"os"
	"strings"
	"regexp"
	"fmt"
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
	"container/list"
)

const (
	COL_WIDTH int = 50
)

type Listing struct {
	current_dir *file_entry
	selection list.List
	highlighted_element *list.Element
	highlighted_index int

	startx, starty int
	width, height int
	rows, cols int
	col_at int
	filter_nomatch bool

	Debug_message string
}

func NewListing(startx, starty int, width, height int) *Listing {
	directoryName, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cwd, err := CreateDirEntry(directoryName)
	panic_perhaps(err)

	return &Listing{
		current_dir: cwd,
		startx: startx,
		starty: starty,
		width:  width,
		height: height,
	}
}

func (ls *Listing) PrintDirectoryListing() int {
	var dir_header_fg termbox.Attribute
	if ls.filter_nomatch {
		dir_header_fg = termbox.ColorRed
	} else {
		dir_header_fg = termbox.ColorWhite
	}

	ls.col_at = ls.calc_start_column()

	var header_text string
	if ls.Debug_message != "" {
		header_text = ls.Debug_message
	} else {
		header_text = ls.current_dir.AbsPath
	}

	util.WriteString(0, 0, ls.width, ls.width, dir_header_fg, termbox.ColorBlue, header_text)
	ls.rows = ls.height - 1
	ls.cols = (ls.selection.Len() / ls.height) + 1
	max_cols := (ls.width / COL_WIDTH) + 1
	num_files := util.Min(ls.selection.Len(), util.Min(max_cols, ls.cols) * ls.rows)
	start_at := ls.col_at * ls.rows

	var i int = 0
	for e := ls.selection.Front(); e != nil && i-start_at < num_files; e = e.Next() {
		if i < start_at {
			i++
			continue // Fast forward
		}

		effective_index := i - start_at
		var file *file_entry = e.Value.(*file_entry)

		ls.print_entry(
			effective_index % ls.rows, effective_index / ls.rows,
			file, e == ls.highlighted_element);
		i++
	}
	return -1
}

func (ls *Listing) print_entry(row, col int, entry *file_entry, is_highlighted bool) {
	var fg, bg termbox.Attribute
	if is_highlighted {
		bg = termbox.ColorMagenta
	} else {
		bg = termbox.ColorBlack
	}

	if !entry.IsAccessible {
		fg = termbox.ColorRed
	} else if entry.IsDir {
		fg = termbox.ColorCyan
	} else {
		fg = termbox.ColorWhite
	}

	util.WriteString(col * COL_WIDTH, row + 1, COL_WIDTH - 1, ls.width, fg, bg, entry.Name)
}

func (ls *Listing) calc_start_column() (r int) {
	var num_visible_cols int
	var highlight_at_col int

	if ls.rows == 0 {
		return 0
	}

	num_visible_cols = int(ls.width / COL_WIDTH) - 1 // Fully visible columns
	if (ls.selection.Len() / ls.rows) <= num_visible_cols {
		return 0
	}
	highlight_at_col = int(ls.highlighted_index / ls.rows)
	if highlight_at_col < ls.col_at {
		return highlight_at_col
	} else if highlight_at_col > ls.col_at + num_visible_cols {
		return highlight_at_col - num_visible_cols
	}
	return ls.col_at
}

func (ls *Listing) UpdateFilter(input string) {
	ls.selection.Init() // more like Reset()

	if ls.current_dir.Contents.Len() == 0 {
		return // Special case
	}

	pattern := create_pattern_from_input(input)
	regexp, err := regexp.Compile(pattern)
	panic_perhaps(err)

	ls.select_and_highlight(regexp)

	if ls.selection.Len() == 0 {
		ls.select_all()
		ls.filter_nomatch = true
	} else {
		ls.filter_nomatch = false
	}
}

func create_pattern_from_input(input string) string {
	input = regexp.QuoteMeta(input)
	input = strings.Replace(input, " ", "(.*)", -1)
	input = fmt.Sprintf("(?i)(%s)", input)
	return input
}

func (ls *Listing) select_and_highlight(re *regexp.Regexp) {
	highlighted_entry := ls.get_highlighted_entry()
	var finalized_highlight bool = false
	var seen_old_highlight bool = false
	var i int = 0

	for e := ls.current_dir.Contents.Front(); e != nil; e = e.Next() {
		seen_old_highlight = seen_old_highlight || e.Value.(*file_entry) == highlighted_entry

		matched := re.MatchString(e.Value.(*file_entry).Name)
		if matched {
			new_select_element := ls.selection.PushBack(e.Value)

			if !seen_old_highlight {
				ls.highlight_element(new_select_element, i) // Inefficient. Meh...
			} else if !finalized_highlight {
				ls.highlight_element(new_select_element, i)
				finalized_highlight = true
			}
		}
		i++
	}
}

func (ls *Listing) highlight_element(e *list.Element, index_in_selection int) {
	ls.highlighted_element = e
	ls.highlighted_index = index_in_selection
}

func (ls *Listing) select_all() {
	highlighted_entry := ls.get_highlighted_entry()
	for e := ls.current_dir.Contents.Front(); e != nil; e = e.Next() {
		element := ls.selection.PushBack(e.Value) // Important that element is that of ls.selection
		if element.Value.(*file_entry) == highlighted_entry {
			ls.highlighted_element = element
		}
	}
}

func (ls *Listing) get_highlighted_entry() (result *file_entry) {
	if ls.current_dir.Contents.Len() == 0 {
		result = nil
	} else if ls.highlighted_element != nil {
		result = ls.highlighted_element.Value.(*file_entry)
	} else {
		result = ls.current_dir.Contents.Front().Value.(*file_entry)
	}
	return
}

func (ls *Listing) GetSelected() (abs_path string, ok bool) {
	return ls.highlighted_element.Value.(*file_entry).AbsPath, true
}

func (ls *Listing) CdHighlighted() {
	highlighted_entry := ls.highlighted_element.Value.(*file_entry)
	if !highlighted_entry.IsDir {
		return
	}
	ls.ChangeDir(highlighted_entry)
}

func (ls *Listing) CdUp() {
	ls.ChangeDir(ls.current_dir.GetParent())
}

func (ls *Listing) ChangeDir(dir *file_entry) {
	dir.ValidateContents()
	ls.current_dir = dir
	ls.highlighted_element = nil
}

func (ls *Listing) MoveCursorDown() {
	next_element := ls.highlighted_element.Next()
	if next_element != nil {
		ls.highlighted_element = next_element
	}
}

func (ls *Listing) MoveCursorUp() {
	prev_element := ls.highlighted_element.Prev()
	if prev_element != nil {
		ls.highlighted_element = prev_element
	}
}

func (ls *Listing) MoveCursorLeft() {
	var element = ls.highlighted_element.Prev()

	for i := 0; i < ls.rows && element != nil; i++ {
		ls.highlighted_element = element
		element = element.Prev()
	}
}

func (ls *Listing) MoveCursorRight() {
	var element = ls.highlighted_element.Next()

	for i := 0; i < ls.rows && element != nil; i++ {
		ls.highlighted_element = element
		element = element.Next()
	}
}

func (ls *Listing) PrevDirectory() {
	for element := ls.current_dir.GetElementInParent().Prev(); element != nil; element = element.Prev() {
		at_entry := element.Value.(*file_entry)
		if !at_entry.IsDir || !at_entry.IsAccessible {
			continue
		}

		ls.ChangeDir(element.Value.(*file_entry))
		break
	}
}

func (ls *Listing) NextDirectory() {
	for element := ls.current_dir.GetElementInParent().Next(); element != nil; element = element.Next() {
		at_entry := element.Value.(*file_entry)
		if !at_entry.IsDir || !at_entry.IsAccessible {
			continue
		}

		ls.ChangeDir(element.Value.(*file_entry))
		break
	}
}

func panic_perhaps(err error) {
	if err != nil {
		panic(err)
	}
}
