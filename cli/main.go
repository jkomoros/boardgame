/*

cli is a simple package to allow a game to render to the screen and be
interacted with. It's intended primarily as a tool to diagnose and play around
with a game while its moves and logic are being defined.

*/
package cli

import (
	"encoding/json"
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jroimartin/gocui"
	"strings"
)

type renderType int

const (
	renderJSON renderType = iota
	renderRender
	renderChest
)

//Controller is the primary type of the package.
type Controller struct {
	game     *boardgame.Game
	gui      *gocui.Gui
	renderer RendererFunc
	render   renderType
	mode     inputMode

	//TODO: having these here feels bad. Shouldn't these be in a mode or
	//something?
	numLines    int
	currentMove boardgame.Move
}

//RenderrerFunc takes a state and outputs a list of strings that should be
//printed to screen to depict it.
type RendererFunc func(boardgame.StatePayload) string

func NewController(game *boardgame.Game, renderer RendererFunc) *Controller {
	return &Controller{
		game:     game,
		renderer: renderer,
	}
}

//Implement the gocui.Manager interface
func (c *Controller) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("main", 0, 0, maxX-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "JSON"
		v.Frame = true
	}

	if v, err := g.SetView("status", 0, maxY-2, maxX-1, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorBlack
		v.BgColor = gocui.ColorWhite
		v.Frame = false
	}

	switch c.mode {
	case modePickMove, modeEditMove:
		if v, err := g.SetView("move", 0, maxY-30, maxX-1, maxY-2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Frame = true
			v.Title = "Proposing Move"
			v.Highlight = true
			v.SelFgColor = gocui.ColorBlack
			v.SelBgColor = gocui.ColorWhite

			v.SetCursor(0, 0)

			g.SetViewOnTop("move")
			g.SetCurrentView("move")

			//We'll render the content below
		}
	case modeNormal:
		//Delete the view, if it exists
		if err := g.DeleteView("move"); err != nil {
			//It's OK if it's ErrUnknownView because that just means it wasn't
			//in there and was a no op.
			if err != gocui.ErrUnknownView {
				return err
			}
		}
	default:
		//Presumably a nil mode. Meh.
	}

	//Update the json field of view

	if view, err := g.View("main"); err == nil {
		view.Clear()
		switch c.render {
		case renderRender:
			//Print renderered view
			fmt.Fprint(view, c.renderRendered())
			view.Title = "Rendered"
		case renderJSON:
			//Print JSON view
			fmt.Fprint(view, c.renderJSON())
			view.Title = "JSON"
		case renderChest:
			fmt.Fprint(view, c.renderChest())
			view.Title = "Chest"
		}

	}

	if view, err := g.View("move"); err == nil {

		view.Clear()
		if c.mode == modePickMove {
			//TODO: this regenerates the move information every time, which seems funny...
			moves := c.renderMoves()

			c.numLines = len(moves)

			fmt.Fprint(view, strings.Join(moves, "\n"))
		} else {
			fmt.Fprint(view, "This is where the move will have its fields enumerated for editing")
		}
	}

	if view, err := g.View("status"); err == nil {
		view.Clear()
		fmt.Fprint(view, c.mode.statusLine())
	}

	return nil
}

func (c *Controller) renderMoves() []string {

	var result []string

	for _, move := range c.game.Moves() {
		result = append(result, move.Name()+" : "+move.Description())
	}

	if len(result) == 0 {
		result = []string{"No moves configured for this game."}
	}

	return result

}

func (c *Controller) renderRendered() string {
	return c.renderer(c.game.State.Payload)
}

func (c *Controller) renderJSON() string {
	return string(boardgame.Serialize(c.game.State.JSON()))
}

func (c *Controller) renderChest() string {

	deck := make(map[string][]interface{})

	for _, name := range c.game.Chest().DeckNames() {

		components := c.game.Chest().Deck(name).Components()

		values := make([]interface{}, len(components))

		for i, component := range components {
			values[i] = struct {
				Index  int
				Values interface{}
			}{
				i,
				component.Values,
			}
		}

		deck[name] = values
	}

	json, err := json.MarshalIndent(deck, "", "  ")

	if err != nil {
		panic(err)
	}

	return string(json)

}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (c *Controller) ScrollUp() {
	var view *gocui.View
	var err error
	if view, err = c.gui.View("main"); err != nil {
		return
	}
	x, y := view.Origin()
	view.SetOrigin(x, y-1)
}

func (c *Controller) ScrollDown() {
	var view *gocui.View
	var err error
	if view, err = c.gui.View("main"); err != nil {
		return
	}
	x, y := view.Origin()
	view.SetOrigin(x, y+1)
}

func (c *Controller) ScrollMoveSelectionUp(v *gocui.View) {
	_, y := v.Cursor()

	if y == 0 {
		return
	}

	v.MoveCursor(0, -1, false)
}

func (c *Controller) ScrollMoveSelectionDown(v *gocui.View) {
	_, y := v.Cursor()

	if y+1 >= c.numLines {
		return
	}

	v.MoveCursor(0, +1, false)
}

func (c *Controller) PickCurrentlySelectedMoveToEdit(v *gocui.View) {
	_, index := v.Cursor()

	move := c.game.Moves()[index]

	c.PickMoveToEdit(move)
}

func (c *Controller) PickMoveToEdit(move boardgame.Move) {
	c.currentMove = move

	c.EnterMode(modeEditMove)
}

func (c *Controller) ToggleRender() {
	c.render++
	if c.render > renderChest {
		c.render = 0
	}
}

func (c *Controller) StartProposingMove() {
	c.EnterMode(modePickMove)
}

//Cancels any mode that we're in by going back to normal mode.
func (c *Controller) CancelMode() {
	c.EnterMode(modeNormal)
}

func (c *Controller) EnterMode(m inputMode) {
	g := c.gui

	//Clear out all keybindings; mode. enterMode will reestablish them.
	g.DeleteKeybindings("")
	for _, view := range g.Views() {
		g.DeleteKeybindings(view.Name())
	}

	m.enterMode(c)
	c.mode = m
}

//Once the controller is set up, call Start. It will block until it is time
//to exit.
func (c *Controller) Start() {

	g, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		panic("Couldn't create gui:" + err.Error())
	}

	defer g.Close()

	c.gui = g

	g.InputEsc = true

	//manager has to be set before setting keybindings, because it clears all
	//keybindings when set.
	g.SetManager(c)

	c.EnterMode(modeNormal)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}

}
