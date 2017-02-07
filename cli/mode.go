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
	overlayContent() *overlayContent
	//Which cell in overlayContent is currently selected. -1, -1 is "none"
	overlaySelectedCell() (row, col int)

	//Returns the title for overlay
	overlayTitle() string
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
	content     *overlayContent
	currentLine int
}

type modeEditMove struct {
	modeBase
	content     *overlayContent
	move        boardgame.Move
	currentLine int
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

//ColumnWidths verifies that evertyhing is aligned and then returns the column sizes
func (o *overlayContent) ColumnWidths() []int {
	if !o.Aligned() {
		o.DefaultPad()
	}

	row := (*o)[0]

	result := make([]int, len(row))

	for c := 0; c < len(row); c++ {
		result[c] = len(row[c])
	}

	return result
}

//DefaultPad goes through and makes sure that every column has the same
//lenght. Defaults to left-aligned for all.
func (o *overlayContent) DefaultPad() {
	if o.Aligned() {
		return
	}

	numColumns := len((*o)[0])

	alignments := make([]columnAlignment, numColumns)

	for i := 0; i < numColumns; i++ {
		alignments[i] = alignLeft
	}

	o.PadWithAlignment(alignments...)
}

//Pad verifies that each column in the overlayContent is the same size, with
//each column left or right aligned.
func (o *overlayContent) PadWithAlignment(alignments ...columnAlignment) {
	//TODO: implement
	if o.Aligned() {
		return
	}

	//Make sure the length of alignments lines up with Aligned.
	if len((*o)[0]) != len(alignments) {
		return
	}

	for c, alignment := range alignments {

		maxColLength := len((*o)[0][c])
		for r := 1; r < len((*o)); r++ {
			if len((*o)[r][c]) > maxColLength {
				maxColLength = len((*o)[r][c])
			}
		}

		for r, line := range *o {
			length := len(line[c])

			lengthToPad := maxColLength - length

			if alignment == alignLeft {

				(*o)[r][c] = (*o)[r][c] + strings.Repeat(" ", lengthToPad)

			} else if alignment == alignRight {
				(*o)[r][c] = strings.Repeat(" ", lengthToPad) + (*o)[r][c]
			}
		}
	}

}

//String returns overlay content rendered as a simple string with lines separated by \n
func (o *overlayContent) String() string {

	if !o.Aligned() {
		o.DefaultPad()
	}

	lines := make([]string, len(*o))

	for i, line := range *o {
		lines[i] = strings.Join(line, "")
	}

	return strings.Join(lines, "\n")

}

func newModeEditMove(c *Controller, move boardgame.Move) *modeEditMove {
	return &modeEditMove{
		modeBase{
			c,
		},
		nil,
		move,
		0,
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

func (m *modeBase) overlayContent() *overlayContent {
	return nil
}

func (m *modeBase) overlayTitle() string {
	return ""
}

func (m *modeBase) overlaySelectedCell() (row, col int) {
	return -1, -1
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
		nil,
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

func (m *modePickMove) overlayContent() *overlayContent {

	if m.content == nil {

		moves := m.c.renderMoves()

		result := make(overlayContent, len(moves))

		for i := 0; i < len(moves); i++ {
			result[i] = []string{moves[i]}
		}

		m.content = &result

	}

	return m.content

}

func (m *modePickMove) overlayTitle() string {
	return "Pick Move To Propose"
}

func (m *modePickMove) overlaySelectedCell() (row, col int) {
	return m.currentLine, 0
}

func (m *modePickMove) MoveSelectionUp() {

	if m.currentLine == 0 {
		return
	}
	m.currentLine--
}

func (m *modePickMove) MoveSelectionDown() {

	if m.content == nil {
		return
	}

	if m.currentLine+1 >= len(*m.content) {
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
	case gocui.KeyArrowUp:
		//TODO: KeyArrowUp and KeyArrowDown should just be handled in a sub-mode
		m.MoveSelectionUp()
		handled = true
	case gocui.KeyArrowDown:
		m.MoveSelectionDown()
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

func (m *modeEditMove) overlayContent() *overlayContent {

	if m.content == nil {

		var result overlayContent

		s := reflect.ValueOf(m.move).Elem()
		typeOfT := s.Type()
		for i := 0; i < s.NumField(); i++ {
			f := s.Field(i)
			fieldName := typeOfT.Field(i).Name
			if !moveFieldNameShouldBeIncluded(fieldName) {
				continue
			}
			result = append(result, []string{fmt.Sprintf("%s (%s)", fieldName, f.Type()), ":", fmt.Sprintf("%v", f.Interface())})
		}

		if len(result) == 0 {
			//No fields!
			result = overlayContent{[]string{"No fields to modify"}}
		}

		m.content = &result

		m.content.PadWithAlignment(alignRight, alignLeft, alignLeft)
	}

	return m.content

}

//TODO: pop these out into a base class
func (m *modeEditMove) MoveSelectionUp() {

	if m.currentLine == 0 {
		return
	}
	m.currentLine--
}

func (m *modeEditMove) MoveSelectionDown() {

	if m.content == nil {
		return
	}

	if m.currentLine+1 >= len(*m.content) {
		return
	}
	m.currentLine++
}

func (m *modeEditMove) overlaySelectedCell() (row, col int) {
	return m.currentLine, 2
}

func (m *modeEditMove) overlayTitle() string {
	return "Editing Move"
}
