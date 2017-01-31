/*

cli is a simple package to allow a game to render to the screen and be
interacted with. It's intended primarily as a tool to diagnose and play around
with a game while its moves and logic are being defined.

*/
package cli

import (
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jroimartin/gocui"
)

//Controller is the primary type of the package.
type Controller struct {
	game     *boardgame.Game
	gui      *gocui.Gui
	renderer RendererFunc
	//Whether or not we should render JSON (false) or the RendererFunc (true)
	render bool
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

func makeLayoutFunc(c *Controller) func(g *gocui.Gui) error {
	//Used to create a closure that captures 'c'
	return func(g *gocui.Gui) error {
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

		//Update the json field of view

		if view, err := g.View("main"); err == nil {
			view.Clear()
			if c.render {
				//Print renderered view
				fmt.Fprint(view, c.renderer(c.game.State.Payload))
			} else {
				//Print JSON view
				fmt.Fprint(view, string(boardgame.Serialize(c.game.State.JSON())))
			}
		}

		if view, err := g.View("status"); err == nil {
			view.Clear()
			fmt.Fprint(view, "Type 't' to toggle json or render output, Ctrl-C to quit")
		}

		return nil
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func makeToggleRenderFunc(c *Controller) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		c.ToggleRender()
		return nil
	}
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

	//manager has to be set before setting keybindings, because it clears all
	//keybindings when set.
	g.SetManagerFunc(makeLayoutFunc(c))

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", 't', gocui.ModNone, makeToggleRenderFunc(c)); err != nil {
		panic(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}

}

func (c *Controller) ToggleRender() {
	c.render = !c.render
}
