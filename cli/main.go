/*

cli is a simple package to allow a game to render to the screen and be
interacted with. It's intended primarily as a tool to diagnose and play around
with a game while its moves and logic are being defined.

*/
package cli

import (
	"github.com/jkomoros/boardgame"
	"github.com/nsf/termbox-go"
	"strings"
)

//Controller is the primary type of the package.
type Controller struct {
	game     *boardgame.Game
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

//Once the controller is set up, call MainLoop. It will block until it is time
//to exit.
func (c *Controller) MainLoop() {

	termbox.Init()

	defer termbox.Close()

	c.draw()

	for {
		evt := termbox.PollEvent()

		if c.mode.handleInput(c, evt) {
			return
		}

		c.draw()

	}

}

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

func (c *Controller) ToggleRender() {
	c.render = !c.render
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
