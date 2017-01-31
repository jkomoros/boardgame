/*

cli is a simple package to allow a game to render to the screen and be
interacted with. It's intended primarily as a tool to diagnose and play around
with a game while its moves and logic are being defined.

*/
package cli

import (
	"github.com/jkomoros/boardgame"
	"github.com/nsf/termbox-go"
)

//Controller is the primary type of the package.
type Controller struct {
	game *boardgame.Game
}

func NewController(game *boardgame.Game) *Controller {
	return &Controller{
		game: game,
	}
}

//Once the controller is set up, call MainLoop. It will block until it is time
//to exit.
func (c *Controller) MainLoop() {

	c.draw()

	for {
		evt := termbox.PollEvent()

		switch evt.Type {
		case termbox.EventKey:
			//TODO: do something special for keys
		}

		c.draw()

	}

}

//draw draws the entire app to the screen
func (c *Controller) draw() {

	clearScreen()

	//TODO: do real drawing
	x := 0
	y := 0
	for _, ch := range "Hello World!" {
		termbox.SetCell(x, y, ch, termbox.ColorWhite, termbox.ColorBlack)
		x++
	}
}

func clearScreen() {
	width, height := termbox.Size()
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}
