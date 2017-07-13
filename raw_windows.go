package termutil

import (
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
