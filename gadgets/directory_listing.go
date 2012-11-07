package gadgets

import (
	"os"
	"strings"
	"regexp"
	"fmt"
	"errors"
	"github.com/chrigrah/nextplz/backend"
	"github.com/chrigrah/nextplz/util"
	"github.com/nsf/termbox-go"
	"container/list"
)

const (
	COL_WIDTH int = 50
)

type Listing struct {
	current_dir *backend.FileEntry
	selection list.List
	highlighted_element *list.Element
	highlighted_index int

	startx, starty int
	width, height int
	rows, cols int
	col_at int
	filter_nomatch bool

	sb *util.ScrollingBoxes

	Debug_message string
}

func NewListing(startx, starty int, width, height int, sb *util.ScrollingBoxes) *Listing {
	directoryName, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cwd, err := backend.CreateDirEntry(directoryName)
	panic_perhaps(err)

	return &Listing{
		current_dir: cwd,
		startx: startx,
		starty: starty,
		width:  width,
		height: height,
		sb: sb,
	}
}

func (ls *Listing) ChangeDirectory(dir string) error {
	cwd, err := backend.CreateDirEntry(dir)
	if err != nil {
		return err
	}

	*ls = Listing{
		current_dir: cwd,
		startx: ls.startx,
		starty: ls.starty,
		width:  ls.width,
		height: ls.height,
		sb: ls.sb,
	}
	return nil
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

	util.WriteString(0, 0, ls.width, dir_header_fg, termbox.ColorBlue, header_text)
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
		var file *backend.FileEntry = e.Value.(*backend.FileEntry)

		ls.print_entry(
			effective_index % ls.rows, effective_index / ls.rows,
			file, e == ls.highlighted_element);
		i++
	}
	return -1
}

func (ls *Listing) print_entry(row, col int, entry *backend.FileEntry, is_highlighted bool) {
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

	str_width := util.Min(COL_WIDTH - 1, ls.width - (col * COL_WIDTH))
	ls.sb.WriteString(uint16(col * COL_WIDTH), uint16(row + 1), str_width, fg, bg, entry.Name, is_highlighted)
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
	if ls.current_dir.Contents.Len() == 0 {
		return // Special case
	}

	pattern := create_pattern_from_input(input)
	regexp, err := regexp.Compile(pattern)
	panic_perhaps(err)

	highlighted_entry := ls.get_highlighted_entry()
	ls.selection, ls.highlighted_element, ls.highlighted_index = select_and_highlight(&ls.current_dir.Contents, highlighted_entry, regexp)

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

func select_and_highlight(all *list.List, old_selection *backend.FileEntry, re *regexp.Regexp) (selection list.List, highlight *list.Element, highlight_i int) {
	var finalized_highlight bool = false
	var seen_old_highlight bool = false
	var i int = 0
	selection = list.List{}

	for e := all.Front(); e != nil; e = e.Next() {
		seen_old_highlight = seen_old_highlight || e.Value.(*backend.FileEntry) == old_selection
		matched := re.MatchString(e.Value.(*backend.FileEntry).Name)
		if matched {
			new_select_element := selection.PushBack(e.Value)

			if !seen_old_highlight {
				highlight = new_select_element
				highlight_i = i
			} else if !finalized_highlight {
				highlight = new_select_element
				highlight_i = i
				finalized_highlight = true
			}
		}
		i++
	}
	return
}

func (ls *Listing) select_all() {
	highlighted_entry := ls.get_highlighted_entry()
	for e := ls.current_dir.Contents.Front(); e != nil; e = e.Next() {
		element := ls.selection.PushBack(e.Value) // Important that element is that of ls.selection
		if element.Value.(*backend.FileEntry) == highlighted_entry {
			ls.highlighted_element = element
		}
	}
}

func (ls *Listing) get_highlighted_entry() (result *backend.FileEntry) {
	if ls.current_dir.Contents.Len() == 0 {
		result = nil
	} else if ls.highlighted_element != nil {
		result = ls.highlighted_element.Value.(*backend.FileEntry)
	} else {
		result = ls.current_dir.Contents.Front().Value.(*backend.FileEntry)
	}
	return
}

func (ls *Listing) GetSelected() (abs_path string, ok bool) {
	return ls.highlighted_element.Value.(*backend.FileEntry).AbsPath, true
}

func (ls *Listing) CdHighlighted() error {
	if ls.highlighted_element == nil {
		return errors.New("No entry is highlighted. Is this an empty folder? Helloooooo....")
	}
	highlighted_entry := ls.highlighted_element.Value.(*backend.FileEntry)
	if !highlighted_entry.IsDir {
		return errors.New("Highlighted entry is not a directory")
	}
	return ls.ChangeDir(highlighted_entry)
}

func (ls *Listing) CdUp() error {
	return ls.ChangeDir(ls.current_dir.GetParent())
}

func (ls *Listing) ChangeDir(dir *backend.FileEntry) error {
	err := dir.ValidateContents()
	if err != nil {
		dir.IsAccessible = false
		return err
	}
	ls.current_dir = dir
	ls.highlighted_element = nil
	return nil
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

func (ls *Listing) PrevDirectory() error {
	for element := ls.current_dir.GetElementInParent().Prev(); element != nil; element = element.Prev() {
		at_entry := element.Value.(*backend.FileEntry)
		if !at_entry.IsDir || !at_entry.IsAccessible {
			continue
		}

		return ls.ChangeDir(element.Value.(*backend.FileEntry))
	}
	return errors.New("No previous directory")
}

func (ls *Listing) NextDirectory() error {
	for element := ls.current_dir.GetElementInParent().Next(); element != nil; element = element.Next() {
		at_entry := element.Value.(*backend.FileEntry)
		if !at_entry.IsDir || !at_entry.IsAccessible {
			continue
		}

		return ls.ChangeDir(element.Value.(*backend.FileEntry))
	}
	return nil
}

func panic_perhaps(err error) {
	if err != nil {
		panic(err)
	}
}
