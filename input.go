package termutil

import (
	"errors"
	"fmt"
	"github.com/nsf/termbox-go"
	"unicode/utf8"
)

func Prompt(prompt string, refresh func(int, int)) string {
	return PromptWithCallback(prompt, refresh, nil)
}

func PromptWithCallback(prompt string, refresh func(int, int), callback func(string, string)) string {
	buffer := ""
	bufpos := 0
	cursor := 0
	for {
		buflen := len(buffer)
		x, y := termbox.Size()
		if refresh != nil {
			refresh(x, y)
		}
		ClearLine(x, y-1)
		Printstring(prompt+": "+buffer, 0, y-1)
		termbox.SetCursor(utf8.RuneCountInString(prompt)+2+cursor, y-1)
		termbox.Flush()
		ev := termbox.PollEvent()
		if ev.Type != termbox.EventKey {
			continue
		}
		key := ParseTermboxEvent(ev)
		switch key {
		case "LEFT":
			if bufpos > 0 {
				r, rs := utf8.DecodeLastRuneInString(buffer[:bufpos])
				bufpos -= rs
				cursor -= Runewidth(r)
			}
		case "RIGHT":
			if bufpos < buflen {
				r, rs := utf8.DecodeRuneInString(buffer[bufpos:])
				bufpos += rs
				cursor += Runewidth(r)
			}
		case "C-a":
			fallthrough
		case "Home":
			bufpos = 0
			cursor = 0
		case "C-e":
			fallthrough
		case "End":
			bufpos = buflen
			cursor = RunewidthStr(buffer)
		case "C-c":
			fallthrough
		case "C-g":
			if callback != nil {
				callback(buffer, key)
			}
			return ""
		case "RET":
			if callback != nil {
				callback(buffer, key)
			}
			return buffer
		case "C-d":
			fallthrough
		case "deletechar":
			if bufpos < buflen {
				r, rs := utf8.DecodeRuneInString(buffer[bufpos:])
				bufpos += rs
				cursor += Runewidth(r)
			} else {
				if callback != nil {
					callback(buffer, key)
				}
				continue
			}
			fallthrough
		case "DEL":
			if buflen > 0 {
				if bufpos == buflen {
					r, rs := utf8.DecodeLastRuneInString(buffer)
					buffer = buffer[0 : buflen-rs]
					bufpos -= rs
					cursor -= Runewidth(r)
				} else if bufpos > 0 {
					r, rs := utf8.DecodeLastRuneInString(buffer[:bufpos])
					buffer = buffer[:bufpos-rs] + buffer[bufpos:]
					bufpos -= rs
					cursor -= Runewidth(r)
				}
			}
		default:
			if utf8.RuneCountInString(key) == 1 {
				r, _ := utf8.DecodeLastRuneInString(buffer)
				buffer = buffer[:bufpos] + key + buffer[bufpos:]
				bufpos += len(key)
				cursor += Runewidth(r)
			}
		}
		if callback != nil {
			callback(buffer, key)
		}
	}
}

func ChoiceIndex(title string, choices []string, def int) int {
	selection := def
	nc := len(choices) - 1
	if selection < 0 || selection > nc {
		selection = 0
	}
	offset := 0
	for {
		_, sy := termbox.Size()
		termbox.HideCursor()
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		Printstring(title, 0, 0)
		for selection < offset {
			offset -= 5
			if offset < 0 {
				offset = 0
			}
		}
		for selection-offset >= sy-1 {
			offset += 5
			if offset >= nc {
				offset = nc
			}
		}
		for i, s := range choices[offset:] {
			Printstring(s, 3, i+1)
		}
		Printstring(">", 1, (selection+1)-offset)
		termbox.Flush()
		ev := termbox.PollEvent()
		if ev.Type != termbox.EventKey {
			continue
		}
		key := ParseTermboxEvent(ev)
		switch key {
		case "C-v":
			fallthrough
		case "next":
			selection += sy - 5
			if selection >= len(choices) {
				selection = len(choices) - 1
			}
		case "M-v":
			fallthrough
		case "prior":
			selection -= sy - 5
			if selection < 0 {
				selection = 0
			}
		case "C-c":
			fallthrough
		case "C-g":
			return def
		case "UP":
			if selection > 0 {
				selection--
			}
		case "DOWN":
			if selection < len(choices)-1 {
				selection++
			}
		case "RET":
			return selection
		}
	}
}

func ParseTermboxEvent(ev termbox.Event) string {
	if ev.Ch == 0 {
		prefix := ""
		if ev.Mod == termbox.ModAlt {
			prefix = "M-"
		}
		switch ev.Key {
		case termbox.KeyBackspace:
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

func YesNo(p string, refresh func(int, int)) bool {
	ret, _ := yesNoChoice(p, false, refresh)
	return ret
}

func YesNoCancel(p string, refresh func(int, int)) (bool, error) {
	return yesNoChoice(p, true, refresh)
}

func yesNoChoice(p string, allowcancel bool, refresh func(int, int)) (bool, error) {
	var plen int
	var pm string
	if allowcancel {
		pm = p + " (y/n/C-g)"
		plen = utf8.RuneCountInString(pm) + 1
	} else {
		pm = p + " (y/n)"
		plen = utf8.RuneCountInString(pm) + 1
	}
	x, y := termbox.Size()
	if refresh != nil {
		refresh(x, y)
	}
	ClearLine(x, y-1)
	Printstring(pm, 0, y-1)
	termbox.SetCursor(plen, y-1)
	termbox.Flush()
	for {
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventResize {
			x, y = termbox.Size()
			if refresh != nil {
				refresh(x, y)
			}
			ClearLine(x, y-1)
			Printstring(pm, 0, y-1)
			termbox.SetCursor(plen, y-1)
			termbox.Flush()
		} else if ev.Type == termbox.EventKey {
			if ev.Key == termbox.KeyCtrlG && allowcancel {
				return false, errors.New("User cancelled")
			} else if ev.Ch == 'y' {
				return true, nil
			} else if ev.Ch == 'n' {
				return false, nil
			}
		}
	}
}