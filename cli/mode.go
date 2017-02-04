package cli

import (
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jroimartin/gocui"
	"reflect"
	"strings"
	"unicode"
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
	//TODO: overlayContent should return an *overlayContent once all of those
	//methods are set.

	//Returns the title for overlay
	overlayTitle() string
	//Which line in the overlay to highlight. -1 is "none"
	overlayHighlightedLine() int
}

type overlayContent [][]string

type columnAlignment int

const (
	alignLeft columnAlignment = iota
	alignRight
)

type modeBase struct {
	c *Controller
}

type modeNormal struct {
	modeBase
}

type modePickMove struct {
	modeBase
	numLines    int
	currentLine int
}

type modeEditMove struct {
	modeBase
	move boardgame.Move
}

//Valid returns true if each row has the same number of columns
func (o overlayContent) Valid() bool {
	if len(o) < 1 {
		return true
	}
	numColumns := len(o[0])
	for _, row := range o {
		if len(row) != numColumns {
			return false
		}
	}
	return true
}

//Aligned returns true if all cells in a column have the same length
func (o overlayContent) Aligned() bool {
	if !o.Valid() {
		return false
	}
	if len(o) == 0 {
		return true
	}
	numColumns := len(o[0])
	for c := 0; c < numColumns; c++ {
		colLength := len(o[0][c])
		for _, row := range o {
			if len(row[c]) != colLength {
				return false
			}
		}
	}
	return true
}

//DefaultPad goes through and makes sure that every column has the same
//lenght. Defaults to left-aligned for all.
func (o *overlayContent) DefaultPad() {
	//TODO: implement
}

//Pad verifies that each column in the overlayContent is the same size, with
//each column left or right aligned.
func (o *overlayContent) PadWithAlignment(alignments ...columnAlignment) {
	//TODO: implement
}

func newModeEditMove(c *Controller, move boardgame.Move) *modeEditMove {
	return &modeEditMove{
		modeBase{
			c,
		},
		move,
	}
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

func (m *modeBase) overlayHighlightedLine() int {
	return -1
}

func newModeNormal(c *Controller) *modeNormal {
	return &modeNormal{modeBase{c}}
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

func newModePickMove(c *Controller) *modePickMove {
	return &modePickMove{
		modeBase{
			c,
		},
		0,
		0,
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

func (m *modePickMove) overlayHighlightedLine() int {
	return m.currentLine
}

func (m *modePickMove) MoveSelectionUp() {

	if m.currentLine == 0 {
		return
	}
	m.currentLine--
}

func (m *modePickMove) MoveSelectionDown() {

	if m.currentLine+1 >= m.numLines {
		return
	}
	m.currentLine++
}

func (m *modePickMove) PickCurrentlySelectedMoveToEdit() {

	move := m.c.game.Moves()[m.currentLine]

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

func moveFieldNameShouldBeIncluded(name string) bool {
	if len(name) < 1 {
		return false
	}

	firstChar := []rune(name)[0]

	if firstChar != unicode.ToUpper(firstChar) {
		//It was not upper case, thus private, thus should not be included.
		return false
	}

	return true
}

func (m *modeEditMove) overlayContent() []string {
	//TODO; return real content

	var lines []string

	s := reflect.ValueOf(m.move).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fieldName := typeOfT.Field(i).Name
		if !moveFieldNameShouldBeIncluded(fieldName) {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s (%s): %v", fieldName, f.Type(), f.Interface()))
	}

	result := make([]string, len(lines))

	//Make sure all of the field types for the size are set the same size
	maxLineLength := 0

	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts[0]) > maxLineLength {
			maxLineLength = len(parts[0])
		}
	}

	for i, line := range lines {
		parts := strings.Split(line, ":")
		result[i] = strings.Repeat(" ", maxLineLength-len(parts[0])) + line
	}

	if len(result) == 0 {
		//No fields!
		return []string{"No fields to modify"}
	}

	return result

}

func (m *modeEditMove) overlayTitle() string {
	return "Editing Move"
}
