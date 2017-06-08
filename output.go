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
	PrintstringColored(termbox.ColorDefault, s, x, y)
}

func PrintstringColored(color termbox.Attribute, s string, x, y int) {
	i := 0
	for _, ru := range s {
		PrintRune(x+i, y, ru, color)
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

type lessRow struct {
	data string
	len  int
}

func lessDrawRows(sx, sy, cy int, rows []lessRow, numrows int) {
	for i := 0; i < sy-1; i++ {
		ri := cy + i
		if ri >= 0 && ri < numrows {
			Printstring(rows[ri].data, 0, i)
		}
	}
	for i := 0; i < sx; i++ {
		PrintRune(i, sy-1, ' ', termbox.AttrReverse)
	}
	PrintstringColored(termbox.AttrReverse, "^C, ^G, q to quit. Arrow keys/Vi keys/Emacs keys to move.", 0, sy-1)
	termbox.Flush()
}

func DisplayScreenMessage(messages ...string) {
	rows := make([]lessRow, 0)
	for _, msg := range messages {
		for _, s := range strings.Split(msg, "\n") {
			renderstring := strings.Replace(s, "\t", "        ", -1)
			rows = append(rows, lessRow{renderstring, len(renderstring)})
		}
	}
	numrows := len(rows)
	cy := 0
	done := false
	for !done {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		sx, sy := termbox.Size()
		if sy > numrows {
			cy = 0
		}
		lessDrawRows(sx, sy, cy, rows, numrows)

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ParseTermboxEvent(ev) {
			case "q", "C-c", "C-g":
				done = true
			case "DOWN", "j", "C-n":
				if cy < numrows-sy {
					cy++
				}
			case "UP", "k", "C-p":
				if cy > 0 {
					cy--
				}
			case "next", "C-v":
				cy += sy
				if cy > numrows-sy {
					cy = numrows - sy
				}
			case "prior", "M-v":
				cy -= sy
				if cy < 0 {
					cy = 0
				}
			case "g", "M-<":
				cy = 0
			case "G", "M->":
				cy = numrows - sy
			}
		}
	}
}
