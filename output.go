package termutil

import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ClearLine(sx, y int) {
	for i := 0; i < sx; i++ {
		termbox.SetCell(i, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
}

func Runewidth(ru rune) int {
	rw := runewidth.RuneWidth(ru)
	if rw <= 0 {
		return 1
	} else {
		return rw
	}
}

func RunewidthStr(s string) int {
	ret := 0
	for _, ru := range s {
		ret += Runewidth(ru)
	}
	return ret
}

func PrintRune(x, y int, ru rune, col termbox.Attribute) {
	if unicode.IsControl(ru) || !utf8.ValidRune(ru) {
		sym := '?'
		if ru <= rune(26) {
			sym = '@' + ru
		}
		termbox.SetCell(x, y, sym, termbox.AttrReverse, termbox.ColorDefault)
	} else {
		termbox.SetCell(x, y, ru, col, termbox.ColorDefault)
	}
}

func Printstring(s string, x, y int) {
	i := 0
	for _, ru := range s {
		PrintRune(x+i, y, ru, termbox.ColorDefault)
		i += Runewidth(ru)
	}
}

func pauseForAnyKey(currentRow int) {
	Printstring("<More>", 0, currentRow)
	termbox.Flush()
	ev := termbox.PollEvent()
	for ev.Type != termbox.EventKey {
		ev = termbox.PollEvent()
	}
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()
}

func DisplayScreenMessage(messages ...string) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	sx, sy := termbox.Size()
	currentRow := 0
	for _, msg := range messages {
		for _, s := range strings.Split(msg, "\n") {
			buf := s
			for len(buf) > sx {
				Printstring(buf[:sx], 0, currentRow)
				buf = buf[sx:]
				currentRow++
				if currentRow > sy-2 {
					pauseForAnyKey(currentRow)
					currentRow = 0
				}
			}
			Printstring(buf, 0, currentRow)
			currentRow++
			if currentRow > sy-2 {
				pauseForAnyKey(currentRow)
				currentRow = 0
			}
		}
	}
	termbox.Flush()
}
