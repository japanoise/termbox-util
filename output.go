package termutil

import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"strings"
	"unicode"
	"unicode/utf8"
)

//Indicate whether the given rune is a word character
func WordCharacter(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c == '_') || c > 127
}

//Pass the screenwidth and a line number; this function will clear the given line.
func ClearLine(sx, y int) {
	for i := 0; i < sx; i++ {
		termbox.SetCell(i, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
}

//Returns how many cells wide the given rune is.
func Runewidth(ru rune) int {
	rw := runewidth.RuneWidth(ru)
	if rw <= 0 {
		return 1
	} else {
		return rw
	}
}

//Returns how many cells wide the given string is
func RunewidthStr(s string) int {
	ret := 0
	for _, ru := range s {
		ret += Runewidth(ru)
	}
	return ret
}

//Prints the rune given on the screen. Uses reverse colors for unprintable
//characters.
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

//Prints the string given on the screen. Uses the above functions to choose how it
//appears.
func Printstring(s string, x, y int) {
	PrintstringColored(termbox.ColorDefault, s, x, y)
}

//Same as Printstring, but passes a color to PrintRune.
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

func lessDrawRows(sx, sy, cx, cy int, rows []lessRow, numrows int) {
	for i := 0; i < sy-1; i++ {
		ri := cy + i
		if ri >= 0 && ri < numrows {
			if cx < len(rows[ri].data) {
				ts, _ := trimString(rows[ri].data, cx)
				Printstring(ts, 0, i)
			}
		}
	}
	for i := 0; i < sx; i++ {
		PrintRune(i, sy-1, ' ', termbox.AttrReverse)
	}
	PrintstringColored(termbox.AttrReverse, "^C, ^G, q to quit. Arrow keys/Vi keys/Emacs keys to move.", 0, sy-1)
	termbox.Flush()
}

//Prints all strings given to the screen, and allows the user to scroll through,
//rather like less(1).
func DisplayScreenMessage(messages ...string) {
	termbox.HideCursor()
	rows := make([]lessRow, 0)
	for _, msg := range messages {
		for _, s := range strings.Split(msg, "\n") {
			renderstring := strings.Replace(s, "\t", "        ", -1)
			rows = append(rows, lessRow{renderstring, len(renderstring)})
		}
	}
	numrows := len(rows)
	cy := 0
	cx := 0
	done := false
	for !done {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		sx, sy := termbox.Size()
		if sy > numrows {
			cy = 0
		}
		lessDrawRows(sx, sy, cx, cy, rows, numrows)

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ParseTermboxEvent(ev) {
			case "q", "C-c", "C-g":
				done = true
			case "DOWN", "j", "C-n":
				if cy < numrows+1-sy {
					cy++
				}
			case "UP", "k", "C-p":
				if cy > 0 {
					cy--
				}
			case "Home", "C-a":
				cx = 0
			case "LEFT", "h", "C-b":
				if cx > 0 {
					cx--
				}
			case "RIGHT", "l", "C-f":
				cx++
			case "next", "C-v":
				cy += sy - 2
				if cy > numrows+1-sy {
					cy = numrows + 1 - sy
				}
			case "prior", "M-v":
				cy -= sy - 2
				if cy < 0 {
					cy = 0
				}
			case "g", "M-<":
				cy = 0
			case "G", "M->":
				cy = numrows + 1 - sy
			case "/", "C-s":
				search := Prompt("Search", func(ssx, ssy int) {
					lessDrawRows(ssx, ssy, cx, cy, rows, numrows)
				})
				termbox.HideCursor()
				for offset, row := range rows[cy:] {
					if strings.Contains(row.data, search) {
						cy += offset
						break
					}
				}
			}
		}
	}
}

func trimString(s string, coloff int) (string, int) {
	if coloff == 0 {
		return s, 0
	}
	sr := []rune(s)
	if coloff < len(sr) {
		ret := string(sr[coloff:])
		return ret, strings.Index(s, ret)
	} else {
		return "", 0
	}
}
