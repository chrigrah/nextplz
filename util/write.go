package util

import (
	"github.com/nsf/termbox-go"
)

func WriteString(x, y int, width int, fg, bg termbox.Attribute, str string) bool {
	return WriteString_FillWithChar(x, y, width, fg, bg, str, ' ')
}

func WriteString_FillWithChar(x, y int, width int, fg, bg termbox.Attribute, str string, fill rune) bool {
	if len(str) > width {
		str = str[:width]
	}

	for at, char := range str {
		termbox.SetCell(x + at, y, rune(char), fg, bg)
	}

	RepeatCharX(x+len(str), x + width, y, fill, fg, bg)

	return true
}

func FillLineTo(startx, line int, stopx int, color termbox.Attribute) {
	RepeatCharX(startx, stopx, line, ' ', color, color)
}

func RepeatCharX(startx, stopx, y int, c rune, fg, bg termbox.Attribute) {
	for x := startx; x < stopx; x++ {
		termbox.SetCell(x, y, c, fg, bg)
	}
}

func RepeatCharY(starty, stopy, x int, c rune, fg, bg termbox.Attribute) {
	for y := starty; y < stopy; y++ {
		termbox.SetCell(x, y, c, fg, bg)
	}
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
