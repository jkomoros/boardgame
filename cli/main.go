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
	mode     inputMode
	renderer RendererFunc
	//Whether or not we should render JSON (false) or the RendererFunc (true)
	render bool
}

//RenderrerFunc takes a state and outputs a list of strings that should be
//printed to screen to depict it.
type RendererFunc func(boardgame.StatePayload) []string

func NewController(game *boardgame.Game, renderer RendererFunc) *Controller {
	return &Controller{
		game:     game,
		mode:     modeDefault,
		renderer: renderer,
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("main", 0, 0, maxX/2-1, maxY/2-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Hey!"
		v.Frame = true

		fmt.Fprintln(v, "Hello, world!")
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	panic("quit called")
	return gocui.ErrQuit
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

	//TODO: key bindings don't appear to be running...
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err)
	}

	g.SetManagerFunc(layout)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}

}

func (c *Controller) ToggleRender() {
	c.render = !c.render
}

/*

//draw draws the entire app to the screen
func (c *Controller) draw() {

	clearScreen()

	if c.render {
		c.drawRender()
	} else {
		c.drawJSON()
	}

	c.drawStatusLine()

	termbox.Flush()
}


func (c *Controller) statusLine() string {
	return c.mode.statusLine()
}

func clearScreen() {
	width, height := termbox.Size()
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

//TODO: these should be global funcs that take a cont
func (c *Controller) drawStatusLine() {
	line := c.statusLine()

	width, height := termbox.Size()

	//Render white background
	y := height - 1

	for x := 0; x < width; x++ {
		termbox.SetCell(x, y, ' ', termbox.ColorBlack, termbox.ColorWhite)
	}

	x := 0

	underlined := false

	var fg termbox.Attribute

	for _, ch := range ">>> " + line {

		if ch == '{' {
			underlined = true
			continue
		} else if ch == '}' {
			underlined = false
			continue
		}

		fg = termbox.ColorBlack

		if underlined {
			fg = fg | termbox.AttrUnderline | termbox.AttrBold
		}

		termbox.SetCell(x, y, ch, fg, termbox.ColorWhite)
		x++
	}
}

func (c *Controller) drawRender() {
	x := 0
	y := 0

	for _, line := range c.renderer(c.game.State.Payload) {
		x = 0

		for _, ch := range line {
			termbox.SetCell(x, y, ch, termbox.ColorWhite, termbox.ColorBlack)
			x++
		}

		y++
	}
}

//Draws the JSON output of the current state to the screen
func (c *Controller) drawJSON() {
	x := 0
	y := 0

	json := string(boardgame.Serialize(c.game.State.JSON()))

	for _, line := range strings.Split(json, "\n") {

		x = 0

		for _, ch := range line {
			termbox.SetCell(x, y, ch, termbox.ColorWhite, termbox.ColorBlack)
			x++
		}

		y++
	}

}

*/
