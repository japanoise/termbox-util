package termutil

import (
	"fmt"
	"github.com/nsf/termbox-go"
)

// Returns a rough approximation of the raw character the user presses.
// It ain't big, it ain't clever, but it should work well enough and get the
// library compiling on Windows.
func GetRawChar(refresh func(int, int)) string {
	ev := func() termbox.Event {
		for {
			ret := termbox.PollEvent()
			if ret.Type == termbox.EventKey {
				return ret
			} else if ret.Type == termbox.EventResize && refresh != nil {
				refresh(ret.Width, ret.Height)
			}
		}
	}()
	prefix := ""
	if ev.Mod == termbox.ModAlt {
		prefix = "\x1b"
	}
	if ev.Ch == 0 {
		if ev.Key <= 0x1F {
			return prefix + string(ev.Key)
		}
		switch ev.Key {
		case termbox.KeyBackspace2, termbox.KeyBackspace, termbox.KeyDelete:
			return prefix + "\x7f"
		case termbox.KeyTab:
			return prefix + "\x09"
		case termbox.KeyEnter:
			return prefix + "\x0d"
		case termbox.KeyEsc:
			return prefix + "\x1b"
		case termbox.KeyCtrlUnderscore:
			return prefix + "\x1f"
		case termbox.KeyCtrlSpace:
			return prefix + "\x00"
		case termbox.KeySpace:
			return prefix + " "
		}
	}
	return prefix + string(ev.Ch)
}

//Parses a termbox.EventKey event and returns it as an emacs-ish keybinding string
//(e.g. "C-c", "LEFT", "TAB", etc.)
//The windows version interprets a C-h as a DEL
func ParseTermboxEvent(ev termbox.Event) string {
	if ev.Ch == 0 {
		prefix := ""
		if ev.Mod == termbox.ModAlt {
			prefix = "M-"
		}
		switch ev.Key {
		case termbox.KeyBackspace2, termbox.KeyBackspace:
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
