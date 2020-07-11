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

Connectable Behaviors

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

Seats, Inactivity, and Players

This package defines two behaviors, InactivePlayer and Seat, whose use isn't
necessarily obvious. This section describes why they're useful.

The core engine has a notion of players, but--with the exception of players who
are configured to be an agent--it doesn't have a sense of who is playing on
behalf of any player. That logic is handled at different layers, most notably in
the server package. That package is the one that keeps track of the actual users
and which ones are tied to which players in the core game logic.

Crucially, when a game is created, a number of the player slots might be
unfilled, as we wait for other users to be invited to the game and attach to it.
The core game engine has no notion that this is happening, because it doesn't
know anything about users in the first place, let alone which ones are actually
attached to the game. Your custom game logic also doesn't by default know
anything about the players--which ones are actually there, which ones are
currently unfilled, etc. For some games, you want to wait until every player is
configured before starting. For other games it's possible to get the game
rolling and have other players join the next round. But that's not possible to
express if your logic doesn't know which players have real users behind them.

For that reason, this package introduces the notion of Seats and
InactivePlayers. Instead of thinking of the core engine's notion of a player as
a literal player, think of it as a seat, that may or may not be occupied. A seat
can be denoted as having a player sitting in it (that it is "Filled"). It can
also express that even though it is not filled, it is no longer open for anyone
to sit in it (that it is "Closed").

If the server logic sees that your game logic includes the Seat behavior, then
it knows that it should seek to communicate to your game logic when a user joins
the game, and listen for your game logic to communicate which open seats should
no longer be filled, even if there are new users. The only way to modify state
in your game logic is by making a move, so that's how the server package tells
your game logic that a seat is filled. If it is sees that your game logic has a
legal move type that is moves.SeatPlayer (or a move that embeds that move
struct), then it will propose that move whenever there is a player to seat.

You can control when the server tries to seat a player by controlling when that
move type is legal. For example, if it always OK to seat a player at any point,
you'd configure it so that move is legal in any phase. If you wanted to only
seat players in a certain phase, you'd use AddMovesForPhase. You can even
control the precise logic of when it is legal by using AddOrderedMovesForPhase.

Note that when a user seeks to join the game, they aren't actually finally added
to the game until they're seated. This means that your game logic can't really
inspect whether there are any players waiting to be seated. It also means that
if that user leaves the window before they're seated they might not be seated.
For that reason there's another, related behavior called InactivePlayer.

By default, every player slot is considered active--that is, they should be
treated as a real, normal player. When the turn order gets to them, the logic of
the game waits for them to make their move before continuing on to the other
player. But sometimes that seat is empty, and we want to get a move on without
waiting for any more players to join. Or sometimes we want to seat a player
immediately, but finish out the current round of play without them, only dealing
them in next round. The way to do that is to embed InactivePlayer behavior,
which contains a flag for whether the player is Inactive.

If you mark the player as inactive, then boardgame.PlayerIndex.Next (and Prev)
will skip that player, acting as though they don't exist. This means that nearly
all of your game logic (that doesn't just count the number of playerStates
naively) will operate as though they don't exist. By default all players are
considered active--even seats that are empty. That's because in general the
safest default is to wait to start the game until everyone is there.

But if you embed both Seat and InactivePlayer in your playerStates, then the
SeatPlayer move will immediately mark any player that is seated to be inactive.
This is again a safe default; it's safest to assume that a recently seated
player needs to be 'dealt in,' likely before the next round starts, before
they're active.

There are other moves that are designed to work with this system.
moves.ActivateInactivePlayer will go through and activate any inactive players.
This is typically included in a phase progression just before a round starts.
moves.CloseEmptySeat will mark as closed any seats that are currently empty,
which effectively says "even though there are more seats, no more people may be
seated". moves.InactivateEmptySeat marks any empty seat as inactive, which
effectively communicates "until I say otherwise, just pretend like the empty
seats aren't there".

Because of these concepts, when you want to know the number of logical players
in your game at any moment, Game.NumPlayers() is often not what you want.
Instead, see boardgame/base.GameDelegate.NumSeatedActivePlayers.

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
