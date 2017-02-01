package cli

import (
	"github.com/jroimartin/gocui"
)

var (
	modeNormal = &normalMode{}
)

type inputMode interface {
	//enterMode enters the specified mode. All of the keybindings will have
	//been cleared before this is called, so the main point of order is to
	//establish whatever key bindings are valid in this mode.
	enterMode(c *Controller)
	//statusLine returns the text that should be displayed in the status line.
	statusLine() string
}

type baseMode struct{}

type normalMode struct {
	baseMode
}

//TODO: a modeProposingMove which has different key bindings.

func (b *baseMode) enterMode(c *Controller) {
	//Establish the keybindings that exist in every mode.

	g := c.gui

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err)
	}

}

func (b *baseMode) statusLine() string {
	return "Type 't' to toggle json or render output, 'm' to propose a move, Ctrl-C to quit"
}

func (n *normalMode) enterMode(c *Controller) {
	n.baseMode.enterMode(c)

	g := c.gui

	if err := g.SetKeybinding("", 't', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ToggleRender()
		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ScrollUp()
		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", gocui.MouseWheelUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ScrollUp()
		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ScrollDown()
		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", gocui.MouseWheelDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ScrollDown()
		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", 'm', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.StartProposingMove()
		return nil
	}); err != nil {
		panic(err)
	}

}
