# termbox-util

[![GoDoc](https://godoc.org/github.com/japanoise/termbox-util?status.svg)](https://godoc.org/github.com/japanoise/termbox-util)

These are some nice functions torn out of [Gomacs](https://github.com/japanoise/gomacs)
which I think are better suited to be out of the project for reuse. It's imported
as `termutil`.

## Output Functions

~~~
	ClearLine(sx, y int)

Pass the screenwidth and a line number; this function will clear the given line.

	Runewidth(ru rune) int

Returns how many cells wide the given rune is.

	RunewidthStr(s string) int

Returns how many cells wide the given string is.

	PrintRune(x, y int, ru rune, col termbox.Attribute)

Prints the rune given on the screen. Uses reverse colors for unprintable
characters.

	Printstring(s string, x, y int)

Prints the string given on the screen. Uses the above functions to choose how it
appears.

	PrintstringColored(c termbox.Attribute, s string, x, y int)

Same as Printstring, but passes a color to PrintRune.

	DisplayScreenMessage(messages ...string)

Prints all strings given to the screen, and allows the user to scroll through,
rather like less(1).
~~~

## Input Functions

Most functions take a function they can use to refresh the screen; they pass it
the x and y size of the screen (calculated using termbox.Size)

~~~
	Prompt(prompt string, refresh func(int, int)) string

Get a string from the user. They can use typical emacs-ish editing commands,
or press C-c or C-g to cancel.

	PromptWithCallback(prompt string, refresh func(int, int), callback func(string, string)) string

As prompt, but calls a function after every keystroke.

	DynamicPromptWithCallback(prompt string, refresh func(int, int), callback func(string, string) string) string

As prompt, but calls a function after every keystroke that can modify the query.

	ChoiceIndex(title string, choices []string, def int) int

Allows the user to select one of many choices displayed on-screen.
Takes a title, choices, and default selection. Returns an index into the choices
array; or def (default)

	ChoiceIndexCallback(title string, choices []string, def int, f func(sel int, sx int, sy int)) int

As ChoiceIndex, but calls a function after drawing the interface, passing it the
currently selected choice, screen width, and screen height.

   ParseTermboxEvent(ev termbox.Event) string

Parses a termbox.EventKey event and returns it as an emacs-ish keybinding string
(e.g. "C-c", "LEFT", "TAB", etc.)

	YesNo(p string, refresh func(int, int)) bool {

Displays the prompt p and asks the user to say y or n. Returns true if y; false
if no.

	YesNoCancel(p string, refresh func(int, int)) (bool, error)

As above, but will return a non-nil error if the user presses C-g.
~~~
