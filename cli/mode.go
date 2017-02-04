package cli

import (
	"github.com/jkomoros/boardgame"
	"github.com/jroimartin/gocui"
)

type inputMode interface {
	//enterMode enters the specified mode, doing any init necessary. Do not
	//register keybindings (unless you're doing mouse)
	enterMode()

	//handleInput is where all input is routed.
	handleInput(key gocui.Key, ch rune, mode gocui.Modifier)
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
	numLines int
}

type modeEditMove struct {
	modeBase
	mode boardgame.Move
}

func (m *modeBase) enterMode() {

	m.c.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	})

}

func (m *modeBase) handleInput(key gocui.Key, ch rune, mode gocui.Modifier) {
	//Currently there arent any base handlers. Quitting via Ctrl-C is handled
	//via gocui key bindings so we can exit cleanly.
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

func (m *modeNormal) handleInput(key gocui.Key, ch rune, mode gocui.Modifier) {
	handled := false
	switch key {
	case gocui.KeyArrowUp:
		m.c.ScrollUp()
		handled = true
	case gocui.KeyArrowDown:
		m.c.ScrollDown()
		handled = true
	}
	if handled {
		return
	}
	switch ch {
	case 't':
		m.c.ToggleRender()
		handled = true
	case 'm':
		m.c.StartProposingMove()
		handled = true
	}
	if handled {
		return
	}
	m.modeBase.handleInput(key, ch, mode)
}

func (m *modeNormal) enterMode() {

	c := m.c

	m.modeBase.enterMode()

	g := c.gui

	//TODO: can we skip setting key bindings for mouse?

	if err := g.SetKeybinding("", gocui.MouseWheelUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ScrollUp()
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

}

func (m *modePickMove) handleInput(key gocui.Key, ch rune, mode gocui.Modifier) {
	handled := false
	switch key {
	case gocui.KeyArrowUp:
		m.MoveSelectionUp()
		handled = true
	case gocui.KeyArrowDown:
		m.MoveSelectionDown()
		handled = true
	case gocui.KeyEsc:
		m.c.CancelMode()
		handled = true
	case gocui.KeyEnter:
		m.PickCurrentlySelectedMoveToEdit()
	}
	if handled {
		return
	}
	m.modeBase.handleInput(key, ch, mode)
}

func (m *modePickMove) enterMode() {

	m.modeBase.enterMode()

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
	m.numLines = len(moves)

	return moves
}

func (m *modePickMove) overlayTitle() string {
	return "Pick Move To Propose"
}

func (m *modePickMove) MoveSelectionUp() {
	_, y := m.c.overlayView.Cursor()

	if y == 0 {
		return
	}

	m.c.overlayView.MoveCursor(0, -1, false)
}

func (m *modePickMove) MoveSelectionDown() {
	_, y := m.c.overlayView.Cursor()

	if y+1 >= m.numLines {
		return
	}

	m.c.overlayView.MoveCursor(0, +1, false)
}

func (m *modePickMove) PickCurrentlySelectedMoveToEdit() {
	//TODO: store this state not in cursor but within this view
	_, index := m.c.overlayView.Cursor()

	move := m.c.game.Moves()[index]

	m.c.PickMoveToEdit(move)
}

func (m *modeEditMove) handleInput(key gocui.Key, ch rune, mode gocui.Modifier) {
	handled := false
	switch key {
	case gocui.KeyEsc:
		//TODO: should this esc handler just be in baseMode?
		m.c.CancelMode()
		handled = true
	}
	if handled {
		return
	}
	m.modeBase.handleInput(key, ch, mode)
}

func (m *modeEditMove) enterMode() {
	m.modeBase.enterMode()

	//TODO:should the esc handler just be in baseMode?

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
