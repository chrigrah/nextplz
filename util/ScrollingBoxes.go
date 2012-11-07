package util

import (
	"time"
	"sync"
	"github.com/nsf/termbox-go"
)

type line_entry struct {
	Text string
	box_width int
	scrolling bool
	fg, bg termbox.Attribute
	scroll_at int
	edge_pause int
	Fresh bool
}

const (
	pause_ticks = 3
)

func (le *line_entry) Write(x, y uint16) {
	if len(le.Text) < le.box_width || !le.scrolling {
		WriteString(int(x), int(y), le.box_width, le.fg, le.bg, le.Text)
	} else {
		WriteString(int(x), int(y), le.box_width, le.fg, le.bg, string([]byte(le.Text)[le.scroll_at:]))
	}
}

func (le *line_entry) Tick(x, y uint16) {
	if len(le.Text) <= le.box_width || !le.scrolling {
		return
	}

	if le.scroll_at == 0 {
		le.edge_pause++
		if le.edge_pause > pause_ticks {
			le.edge_pause = 0
			le.scroll_at = 1
		}
	} else {
		end_diff := le.scroll_at + le.box_width - len(le.Text)
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
	UpdateTicks chan int
}

var DefaultSL ScrollingBoxes

func (sl *ScrollingBoxes) StartTicker() {
	sl.lines = make(map[uint32] *line_entry)
	tick_len, _ := time.ParseDuration("250ms")
	tick_chan := time.Tick(tick_len)
	sl.UpdateTicks = make(chan int, 3)

	for {
		<-tick_chan
		sl.TickAndWriteAll()
		sl.UpdateTicks <- 1
	}
}

func (sl *ScrollingBoxes) WriteString(x, y uint16, width int, fg, bg termbox.Attribute, str string, scrolling bool) bool {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	index := uint32(x) << 16
	index = index + uint32(y)

	entry := sl.lines[index]

	if entry != nil {
		if entry.Text == str {
			entry.Fresh = true
			entry.fg = fg
			entry.bg = bg
			if entry.scrolling != scrolling {
				entry.scroll_at = 0
				entry.edge_pause = 0
				entry.scrolling = scrolling
			}
			return true
		} else {
			delete(sl.lines, index)
		}
	}

	sl.lines[index] = &line_entry {
		Text: str,
		box_width: width,
		scrolling: scrolling,
		fg: fg,
		bg: bg,
		scroll_at: 0,
		edge_pause: 0,
		Fresh: true,
	}

	return true
}

func (sl *ScrollingBoxes) TickAndWriteAll() {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	for index, value := range sl.lines {
		y := uint16(index % (1<<16))
		x := uint16(index >> 16)
		value.Tick(x, y)
		value.Write(x, y)
	}
}

func (sl *ScrollingBoxes) WriteAll() {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	for index, value := range sl.lines {
		y := uint16(index % (1<<16))
		x := uint16(index >> 16)
		if !value.Fresh {
			delete(sl.lines, index)
			continue
		}
		value.Write(x, y)
		value.Fresh = false
	}
}
