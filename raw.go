// +build !windows

package termutil

import (
	"fmt"
	"github.com/nsf/termbox-go"
)

//Get a raw character from termbox
func GetRawChar(refresh func(int, int)) string {
	done := false
	chara := ""
	for !done {
		data := make([]byte, 4)
		i := termbox.PollRawEvent(data)
		parsed := termbox.ParseEvent(data)
		if parsed.Type == termbox.EventKey {
			chara = string(data[:i.N])
			done = true
		} else if parsed.Type == termbox.EventResize && refresh != nil {
			refresh(termbox.Size())
		}
	}
	return chara
}

//Parses a termbox.EventKey event and returns it as an emacs-ish keybinding string
//(e.g. "C-c", "LEFT", "TAB", etc.)
func ParseTermboxEvent(ev termbox.Event) string {
	if ev.Ch == 0 {
		prefix := ""
		if ev.Mod == termbox.ModAlt {
			prefix = "M-"
		}
		switch ev.Key {
		case termbox.KeyBackspace2:
			return prefix + "DEL"
		case termbox.KeyTab:
			return prefix + "TAB"
		case termbox.KeyEnter:
			return prefix + "RET"
		case termbox.KeyArrowDown:
			return prefix + "DOWN"
		case termbox.KeyArrowUp:
			return prefix + "UP"
		case termbox.KeyArrowLeft:
			return prefix + "LEFT"
		case termbox.KeyArrowRight:
			return prefix + "RIGHT"
		case termbox.KeyPgdn:
			return prefix + "next"
		case termbox.KeyPgup:
			return prefix + "prior"
		case termbox.KeyHome:
			return prefix + "Home"
		case termbox.KeyEnd:
			return prefix + "End"
		case termbox.KeyDelete:
			return prefix + "deletechar"
		case termbox.KeyInsert:
			return prefix + "insert"
		case termbox.KeyEsc:
			return prefix + "ESC"
		case termbox.KeyCtrlUnderscore:
			if ev.Mod == termbox.ModAlt {
				return "C-M-_"
			} else {
				return "C-_"
			}
		case termbox.KeyCtrlSpace:
			if ev.Mod == termbox.ModAlt {
				return "C-M-@" // ikr, weird. but try: C-h c, C-SPC. it's C-@.
			} else {
				return "C-@"
			}
		case termbox.KeySpace:
			if ev.Mod == termbox.ModAlt {
				return "M-SPC"
			}
			return " "
		}
		if ev.Key <= 0x1A {
			if ev.Mod == termbox.ModAlt {
				return fmt.Sprintf("C-M-%c", 96+ev.Key)
			} else {
				return fmt.Sprintf("C-%c", 96+ev.Key)
			}
		} else if ev.Key <= termbox.KeyF1 && ev.Key >= termbox.KeyF12 {
			if ev.Mod == termbox.ModAlt {
				return fmt.Sprintf("M-f%d", 1+(termbox.KeyF1-ev.Key))
			} else {
				return fmt.Sprintf("f%d", 1+(termbox.KeyF1-ev.Key))
			}
		}
	} else if ev.Mod == termbox.ModAlt {
		return fmt.Sprintf("M-%c", ev.Ch)
	}
	return string(ev.Ch)
}
