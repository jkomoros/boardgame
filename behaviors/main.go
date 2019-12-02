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

Behaviors often require access to the struct they're embedded within. These
types of behaviors are called Connectable, and if they are this type thentheir
ConnectBehavior should always be called within the subState's FinishStateSetUp,
like so:
	//FinishStateSetUp is called by the main engine when a State is being created.
	//ConnectContainingState will have already been called, so State and StatePropertyRef
	//will be set. The engine doesn't require anything to be done in this method; it's
	//typically used to connect behaviors.
	func (g *gameState) FinishStateSetUp() {
		//PlayerColor is a Connectable, which requires us to call ConnectBehavior
		//on it, passing a reference to ourselvs (the struct it's embeddded in).
		g.PlayerColor.ConnectBehavior(g)
		//If you had ohter Connectable's in this struct, you'd call ConnectBehavior
		//here, too.
	}

Connectable behaviors that are not connected will error when their
ValidConfiguration is called, and the main library will notice that while
NewGameManager is being executed, which will fail with a descriptive error.
*/
package behaviors

import "github.com/jkomoros/boardgame"

//Connectable is the interface that behaviors that are Connectable implements.
//Connectable behaviors are ones that must have their ConnectBehavior called
//within their SubState cdontainer's FinishStateSetUp method. The
//ValidConfiguration method will return an error if they weren't connected,
//which will help diagnose the problem early if you forget.
type Connectable interface {
	//ConnectBehavior lets the behavior have a reference to the struct its
	//embedded in, as some behaviors need access to the broader state.
	ConnectBehavior(containgSubState boardgame.SubState)

	//Connectable behaviors should implement ValidConfiguration and return an
	//error if they haven't yet been Connected, which will help the main engine
	//know to fail NewGameManager, allowing the problem to be fixed more quickly.
	boardgame.ConfigurationValidator
}
