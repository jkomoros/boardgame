package cli

import (
	"github.com/nsf/termbox-go"
)

var (
	modeDefault = &defaultMode{}
)

type inputMode interface {
	handleInput(c *Controller, evt termbox.Event) (doQuit bool)
	statusLine() string
}

type baseMode struct{}

type defaultMode struct {
	baseMode
}

func (b *baseMode) handleInput(c *Controller, evt termbox.Event) bool {
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyCtrlC:
			return true
		}
	}

	return false
}

func (b *baseMode) statusLine() string {
	return "Ctrl-C to exit"
}

func (d *defaultMode) handleInput(c *Controller, evt termbox.Event) bool {

	//TODO: do our own event handling.
	return d.baseMode.handleInput(c, evt)

}
