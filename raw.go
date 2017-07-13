// +build !windows

package termutil

import (
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
