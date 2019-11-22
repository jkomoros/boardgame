/*

Package behaviors defines a handful of convenient behaviors that can be
anonymously embedded into your SubState (e.g. gameState and playerState)
structs.

A behavior is a combination of persistable properties, as well as methods that
mutate those properties. They encapsulate commonly required behavior, like
setting current player or round robin properties. Think of them like lego bricks
you can add to your game and player states.

`boardgame-util codegen` will automatically include the state properties of the
behaviors in the generated PropertyReader.

Behaviors often require access to the struct they're embedded within, so their
ConnectBehavior should always be called within the subState's
FinishStateSetUp, like so:

    func (g *gameState) FinishStateSetUp() {
        g.PlayerColor.ConnectBehavior(g)
    }

*/
package behaviors

import "github.com/jkomoros/boardgame"

//Interface is the interface that all behaviors must implement
type Interface interface {
	//ConnectBehavior lets the behavior have a reference to the struct its
	//embedded in, as some behaviors need access to the broader state.
	ConnectBehavior(containgSubState boardgame.SubState)
}
