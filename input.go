//These are some nice functions torn out of Gomacs which I think are better
//suited to be out of the project for reuse. It's imported as termutil.
package termutil

import (
	"errors"
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
	offset := 0
	iw := RunewidthStr(prompt + ": ")
	for {
		buflen := len(buffer)
		x, y := termbox.Size()
		if refresh != nil {
			refresh(x, y)
		}
		ClearLine(x, y-1)
		for iw+cursor >= x {
			offset++
			cursor--
		}
		for iw+cursor < iw {
			offset--
			cursor++
		}
		t, _ := trimString(buffer, offset)
		Printstring(prompt+": "+t, 0, y-1)
		termbox.SetCursor(iw+cursor, y-1)
		termbox.Flush()
		ev := termbox.PollEvent()
		if ev.Type != termbox.EventKey {
			continue
		}
		key := ParseTermboxEvent(ev)
		switch key {
		case "LEFT", "C-b":
			if bufpos > 0 {
				r, rs := utf8.DecodeLastRuneInString(buffer[:bufpos])
				bufpos -= rs
				cursor -= Runewidth(r)
			}
		case "RIGHT", "C-f":
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
			offset = 0
		case "C-e":
			fallthrough
		case "End":
			bufpos = buflen
			if RunewidthStr(buffer) > x {
				cursor = x - 1
				offset = buflen - x
			} else {
				offset = 0
				cursor = RunewidthStr(buffer)
			}
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
		case "DEL", "C-h":
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
		case "UP", "C-p":
			if selection > 0 {
				selection--
			}
		case "DOWN", "C-n":
			if selection < len(choices)-1 {
				selection++
			}
		case "LEFT", "C-b":
			if cx > 0 {
				cx--
			}
		case "RIGHT", "C-f":
			cx++
		case "C-a", "Home":
			cx = 0
		case "M-<":
			selection = 0
		case "M->":
			selection = len(choices) - 1
		case "RET":
			return selection
		}
	}
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
