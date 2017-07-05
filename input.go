//These are some nice functions torn out of Gomacs which I think are better
//suited to be out of the project for reuse. It's imported as termutil.
package termutil

import (
	"errors"
	"fmt"
	"github.com/nsf/termbox-go"
	"unicode/utf8"
)

//Get a string from the user. They can use typical emacs-ish editing commands,
//or press C-c or C-g to cancel.
func Prompt(prompt string, refresh func(int, int)) string {
	return PromptWithCallback(prompt, refresh, nil)
}

//As prompt, but calls a function after every keystroke.
func PromptWithCallback(prompt string, refresh func(int, int), callback func(string, string)) string {
	if callback == nil {
		return DynamicPromptWithCallback(prompt, refresh, nil)
	} else {
		return DynamicPromptWithCallback(prompt, refresh, func(a, b string) string {
			callback(a, b)
			return a
		})
	}
}

//As prompt, but calls a function after every keystroke that can modify the query.
func DynamicPromptWithCallback(prompt string, refresh func(int, int), callback func(string, string) string) string {
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
				result := callback(buffer, key)
				if result != buffer {
					buffer, buflen, bufpos, cursor = recalcBuffer(result)
				}
			}
			return ""
		case "RET":
			if callback != nil {
				result := callback(buffer, key)
				if result != buffer {
					buffer, buflen, bufpos, cursor = recalcBuffer(result)
				}
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
					result := callback(buffer, key)
					if result != buffer {
						buffer, buflen, bufpos, cursor = recalcBuffer(result)
					}
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
			result := callback(buffer, key)
			if result != buffer {
				buffer, buflen, bufpos, cursor = recalcBuffer(result)
			}
		}
	}
}

func recalcBuffer(result string) (string, int, int, int) {
	rlen := len(result)
	return result, rlen, rlen, RunewidthStr(result)
}

//Allows the user to select one of many choices displayed on-screen.
//Takes a title, choices, and default selection. Returns an index into the choices
//array; or def (default)
func ChoiceIndex(title string, choices []string, def int) int {
	selection := def
	nc := len(choices) - 1
	if selection < 0 || selection > nc {
		selection = 0
	}
	offset := 0
	cx := 0
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
			ts, _ := trimString(s, cx)
			Printstring(ts, 3, i+1)
			if cx > 0 {
				Printstring("←", 2, i+1)
			}
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
		case "LEFT":
			if cx > 0 {
				cx--
			}
		case "RIGHT":
			cx++
		case "RET":
			return selection
		}
	}
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

//Displays the prompt p and asks the user to say y or n. Returns true if y; false
//if no.
func YesNo(p string, refresh func(int, int)) bool {
	ret, _ := yesNoChoice(p, false, refresh)
	return ret
}

//Same as YesNo, but will return a non-nil error if the user presses C-g.
func YesNoCancel(p string, refresh func(int, int)) (bool, error) {
	return yesNoChoice(p, true, refresh)
}

// Asks the user to press one of a set of keys. Returns the one which they pressed.
func PressKey(p string, refresh func(int, int), keys ...string) string {
	var plen int
	pm := p + " ("
	for i, key := range keys {
		if i != 0 {
			pm += "/"
		}
		pm += key
	}
	pm += ")"
	plen = utf8.RuneCountInString(pm) + 1
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
			pressedkey := ParseTermboxEvent(ev)
			for _, key := range keys {
				if key == pressedkey {
					return key
				}
			}
		}
	}
}

func yesNoChoice(p string, allowcancel bool, refresh func(int, int)) (bool, error) {
	if allowcancel {
		key := PressKey(p, refresh, "y", "n", "C-g")
		switch key {
		case "y":
			return true, nil
		case "n":
			return false, nil
		case "C-g", "C-c":
			return false, errors.New("User cancelled")
		}
	}
	key := PressKey(p, refresh, "y", "n")
	return key == "y", nil
}
