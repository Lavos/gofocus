package gofocus

import (
	"github.com/nsf/termbox-go"
	"strings"
	"fmt"
	"log"
)

type Terminal struct {

}

func (t *Terminal) ColorHline(y int, bg termbox.Attribute) {
	var width, _ = termbox.Size()

	for c := 0; c < width; c++ {
		termbox.SetCell(c, y, '\000', termbox.ColorDefault, bg)
	}
}

func (t *Terminal) PrintWordwrap(msg string, width, y int) (int, int) {
	working := []byte(msg)
	lines := make([][]byte, 0)

work:
	for len(working) != 0 {

		// if the width is less than a line, no problem!
		if len(working) <= width {
			lines = append(lines, working)
			working = working[len(working):]

			continue
		}

		// find the space closest to the end, and split there
		for xw := width; xw != 0; xw-- {
			if working[xw] == 0x20 {
				lines = append(lines, working[:xw])
				working = working[xw+1:]

				continue work
			}
		}

		// didn't find any spaces to split on, so just split at width
		lines = append(lines, working[:width])
		working = working[width:]
	}

	var lastx int

	for index, line := range lines {
		lastx = t.PrintLine(1, index+y, termbox.ColorWhite, termbox.ColorDefault, string(line))
	}

	return lastx, len(lines)+y -1
}

func (t *Terminal) PrintLine(x, y int, fg, bg termbox.Attribute, msg string) int {
	var clean string = strings.Replace(msg, "\n", "", -1)

	for _, c := range clean {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}

	return x
}

func (t *Terminal) DrawScreen(remaining int, tweet *MicroTweet, value []rune) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	// title line
	t.ColorHline(0, termbox.ColorWhite)
	t.PrintLine(1, 0, termbox.ColorBlack, termbox.ColorWhite, "Go Focus")

	var bar_color termbox.Attribute

	if remaining == 0 {
		bar_color = termbox.ColorWhite
	} else {
		bar_color = termbox.ColorYellow
	}

	// status line
	t.ColorHline(1, bar_color)
	t.PrintLine(1, 1, termbox.ColorBlack, bar_color, fmt.Sprintf("%v tweets remaining.", remaining))

	// tweet
	t.PrintLine(1, 2, termbox.ColorMagenta, termbox.ColorBlack, tweet.ScreenName)
	t.PrintLine(21, 2, termbox.ColorGreen, termbox.ColorBlack, tweet.UserName)
	t.PrintWordwrap(tweet.Text, 50, 3)

	// compose line
	t.ColorHline(9, termbox.ColorBlue)
	t.PrintLine(1, 9, termbox.ColorBlack, termbox.ColorBlue, fmt.Sprintf("%v characters.", len(value)))

	if len(value) > 0 {
		lastx, lasty := t.PrintWordwrap(string(value), 50, 10)
		termbox.SetCursor(lastx, lasty)
	} else {
		termbox.SetCursor(1, 10)
	}
}

func (t *Terminal) Run(key chan termbox.Event){
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			key <- ev
		}
	}
}

func NewTerminal() *Terminal {
	tb_err := termbox.Init()
	if tb_err != nil {
		log.Panic(tb_err)
	}
	termbox.SetInputMode(termbox.InputEsc)

	return &Terminal{}
}
