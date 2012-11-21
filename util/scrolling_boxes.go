package util

import (
	"time"
	"sync"
	"github.com/nsf/termbox-go"
)

type write_entry struct {
	text string
	fg, bg termbox.Attribute
	box_width int
	scrolling bool
}

type line_entry struct {
	replacement_entry write_entry
	active_entry write_entry
	scroll_at int
	edge_pause int
	Fresh bool
}

const (
	pause_ticks = 3
)

func (le *line_entry) Write(x, y uint16) {
	if len(le.active_entry.text) < le.active_entry.box_width || !le.active_entry.scrolling {
		WriteString(int(x), int(y), le.active_entry.box_width, le.active_entry.fg, le.active_entry.bg,
					le.active_entry.text)
	} else {
		WriteString(int(x), int(y), le.active_entry.box_width, le.active_entry.fg, le.active_entry.bg,
					string([]byte(le.active_entry.text)[le.scroll_at:]))
	}
}

func (le *line_entry) Tick(x, y uint16) {
	if len(le.active_entry.text) <= le.active_entry.box_width || !le.active_entry.scrolling {
		return
	}

	if le.scroll_at == 0 {
		le.edge_pause++
		if le.edge_pause > pause_ticks {
			le.edge_pause = 0
			le.scroll_at = 1
		}
	} else {
		end_diff := le.scroll_at + le.active_entry.box_width - len(le.active_entry.text)
		if end_diff == 0 {
			le.edge_pause++
			if le.edge_pause > pause_ticks {
				le.edge_pause = 0
				le.scroll_at = 0
			}
		} else {
			le.scroll_at++
		}
	}
}

type ScrollingBoxes struct {
	lines map[uint32] *line_entry
	lock sync.Mutex
	UpdateChan chan int
}

var DefaultSL ScrollingBoxes

func (sl *ScrollingBoxes) StartTicker() {
	sl.lines = make(map[uint32] *line_entry)
	tick_len, _ := time.ParseDuration("250ms")
	tick_chan := time.Tick(tick_len)

	for {
		<-tick_chan
		sl.ScrollTick()
		sl.UpdateChan <- 1
	}
}

func (sl *ScrollingBoxes) WriteString(x, y uint16, width int, fg, bg termbox.Attribute, str string, scrolling bool) bool {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	index := uint32(x) << 16
	index = index + uint32(y)

	entry := sl.lines[index]

	if entry != nil {
		// If there is already a line_entry at that index then the write_entry struct replacement_entry is filled out 
		// and the new information will be used only if it is the last WriteString call for the given x,y combination
		// before the next Update call, at which point the replacement_entry becomes the active_entry

		entry.replacement_entry.text = str
		entry.replacement_entry.fg = fg
		entry.replacement_entry.bg = bg
		entry.replacement_entry.scrolling = scrolling
		entry.replacement_entry.box_width = width
		entry.Fresh = true
		return true
	}

	sl.lines[index] = &line_entry {
		active_entry: write_entry{
				text: str,
				fg: fg,
				bg: bg,
				box_width: width,
				scrolling: scrolling,
			},
		scroll_at: 0,
		edge_pause: 0,
		Fresh: true,
	}

	return true
}

func (sl *ScrollingBoxes) ScrollTick() {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	for index, value := range sl.lines {
		y := uint16(index % (1<<16))
		x := uint16(index >> 16)
		value.Tick(x, y)
		//value.Write(x, y)
	}
}

func (sl *ScrollingBoxes) Draw() {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	for index, value := range sl.lines {
		y := uint16(index % (1<<16))
		x := uint16(index >> 16)
		if !value.Fresh {
			delete(sl.lines, index)
			continue
		}

		if value.replacement_entry.text != value.active_entry.text {
			value.scroll_at = 0
			value.edge_pause = 0
		}
		value.active_entry = value.replacement_entry
		value.Write(x, y)
		value.Fresh = false
	}
}
