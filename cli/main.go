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
	game        *boardgame.Game
	gui         *gocui.Gui
	renderer    RendererFunc
	render      renderType
	mode        inputMode
	mainView    *gocui.View
	overlayView *gocui.View
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

//keyPressed is the func we use to "edit" the main view. gocui's keybinding is
//a bit verbose for us (especially to register handlers for many keys), so
//instead we have the main view ALWAYS be current, set it to have this as an
//editor, and then route all input through it. This is a perversion of gocui's
//model, and basically circumvents half the value of the package. Oh well, the
//abstractions around windows are still useful for rendering. :-/
func (c *Controller) keyPressed(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	c.mode.handleInput(key, ch, mod)
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
		v.Editable = true
		v.Editor = gocui.EditorFunc(c.keyPressed)
		c.gui.SetCurrentView("main")
		c.mainView = v
	}

	if v, err := g.SetView("status", 0, maxY-2, maxX-1, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorBlack
		v.BgColor = gocui.ColorWhite
		v.Frame = false
	}

	if c.mode.showOverlay() {
		if v, err := g.SetView("overlay", 0, maxY-30, maxX-1, maxY-2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Frame = true
			v.Title = c.mode.overlayTitle()
			v.Highlight = true
			v.SelFgColor = gocui.ColorBlack
			v.SelBgColor = gocui.ColorWhite

			g.SetViewOnTop("overlay")

			c.overlayView = v

			//We'll render the content below
		}
	} else {
		//Delete the view, if it exists
		if err := g.DeleteView("overlay"); err != nil {
			//It's OK if it's ErrUnknownView because that just means it wasn't
			//in there and was a no op.
			if err != gocui.ErrUnknownView {
				return err
			}
		}
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

	if view, err := g.View("overlay"); err == nil {

		view.Clear()

		fmt.Fprint(view, strings.Join(c.mode.overlayContent(), "\n"))

		highlightedLine := c.mode.overlayHighlightedLine()

		if highlightedLine >= 0 {
			view.SetCursor(0, highlightedLine)
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

func (c *Controller) PickMoveToEdit(move boardgame.Move) {
	c.EnterMode(newModeEditMove(c, move))
}

func (c *Controller) ToggleRender() {
	c.render++
	if c.render > renderChest {
		c.render = 0
	}
}

func (c *Controller) StartProposingMove() {
	c.EnterMode(newModePickMove(c))
}

//Cancels any mode that we're in by going back to normal mode.
func (c *Controller) CancelMode() {
	c.EnterMode(newModeNormal(c))
}

func (c *Controller) EnterMode(m inputMode) {
	g := c.gui

	//Clear out all keybindings; mode. enterMode will reestablish them.
	g.DeleteKeybindings("")
	for _, view := range g.Views() {
		g.DeleteKeybindings(view.Name())
	}

	m.enterMode()
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

	c.EnterMode(newModeNormal(c))

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}

}
