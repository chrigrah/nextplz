package util

import (
	"github.com/nsf/termbox-go"
)

func FillLineTo(startx, line int, stopx int, color termbox.Attribute) (err error) {
	for i := startx; i < stopx; i++ {
		termbox.SetCell(i, line, ' ', color, color)
	}
	return
}

func WriteString(x, y int, box_width, maxX int, fg, bg termbox.Attribute, str string) bool {
	var effective_box_width int = Min(box_width, maxX-x)

	if len(str) > effective_box_width {
		str = str[:effective_box_width]
	}

	for at, char := range str {
		termbox.SetCell(x + at, y, rune(char), fg, bg)
	}

	FillLineTo(x+len(str), y, x + effective_box_width, bg)

	return true
}

func Max(x, y int) (r int) {
	if x > y {
		r = x
	} else {
		r = y
	}
	return
}

func Min(i, j int) (r int) {
	if i < j {
		r = i
	} else {
		r = j
	}
	return
}
