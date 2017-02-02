package cli

import (
	"github.com/jroimartin/gocui"
)

type inputMode interface {
	//enterMode enters the specified mode. All of the keybindings will have
	//been cleared before this is called, so the main point of order is to
	//establish whatever key bindings are valid in this mode.
	enterMode()
	//statusLine returns the text that should be displayed in the status line.
	statusLine() string
	//Whether or not the overlay should be visible
	showOverlay() bool
	//Returns overlay content
	overlayContent() []string
	//Returns the title for overlay
	overlayTitle() string
}

type modeBase struct {
	c *Controller
}

type modeNormal struct {
	modeBase
}

type modePickMove struct {
	modeBase
}

type modeEditMove struct {
	modeBase
}

func (m *modeBase) enterMode() {
	//Establish the keybindings that exist in every mode.

	g := m.c.gui

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err)
	}

}

func (m *modeBase) statusLine() string {
	return "Type 't' to toggle json or render output, 'm' to propose a move, Ctrl-C to quit"
}

func (m *modeBase) showOverlay() bool {
	return false
}

func (m *modeBase) overlayContent() []string {
	return nil
}

func (m *modeBase) overlayTitle() string {
	return ""
}

func (m *modeNormal) enterMode() {

	c := m.c

	m.modeBase.enterMode()

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

func (m *modePickMove) enterMode() {

	m.modeBase.enterMode()

	c := m.c

	g := c.gui

	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.CancelMode()
		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("overlay", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ScrollMoveSelectionUp(v)
		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("overlay", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ScrollMoveSelectionDown(v)
		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("overlay", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.PickCurrentlySelectedMoveToEdit(v)
		return nil
	}); err != nil {
		panic(err)
	}
}

func (m *modePickMove) statusLine() string {
	return "'Enter' to pick a move to edit. 'Esc' to cancel"
}

func (m *modePickMove) showOverlay() bool {
	return true
}

func (m *modePickMove) overlayContent() []string {
	//TODO: memoize this
	moves := m.c.renderMoves()

	//TODO: this is VERY weird that we're using a side-effect to set this
	//piece of state in controller.
	m.c.numLines = len(moves)

	return moves
}

func (m *modePickMove) overlayTitle() string {
	return "Pick Move To Propose"
}

func (m *modeEditMove) enterMode() {
	m.modeBase.enterMode()

	c := m.c

	g := c.gui

	//TODO:should the esc handler just be in baseMode?
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.CancelMode()
		return nil
	}); err != nil {
		panic(err)
	}
}

func (m *modeEditMove) statusLine() string {
	return "this is where you edit the move, I guess. Or 'Esc' to cancel. I don't care."
}

func (m *modeEditMove) showOverlay() bool {
	return true
}

func (m *modeEditMove) overlayContent() []string {
	//TODO; return real content
	return []string{"This is where the move will have its fields enumerated for editing"}

}

func (m *modeEditMove) overlayTitle() string {
	return "Editing Move"
}
