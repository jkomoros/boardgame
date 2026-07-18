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

# When to Create a Behavior

Not every piece of state deserves to be a behavior. A simple int or bool field
is fine on its own. Behaviors earn their keep when they serve as Schelling
points -- standard names and interfaces that multiple parts of the system
coordinate around. Specifically, a behavior is appropriate when it provides one
or more of the following:

Companion Interface: The behavior satisfies an interface in [moves/interfaces]
that moves discover via type assertion. This is the primary value of most
behaviors. [CurrentPlayerBehavior] satisfies [interfaces.CurrentPlayerSetter],
which [moves.FinishTurn] casts for. [InactivePlayer] satisfies
[interfaces.PlayerInactiver], which [moves.ActivateInactivePlayer] casts for.
Without the interface, each game would need to tell each move where to find
the relevant state. With it, moves auto-discover the state.

Framework Integration: The behavior plugs into framework machinery beyond
moves. [InactivePlayer] integrates with [boardgame.PlayerIndex.Next] (skipping
inactive players), [base.GameDelegate.CheckGameFinished] (excluding inactive
players from winners), and [base.GameDelegate.PlayerMayBeActive].
[PlayerOrderBehavior] integrates with [boardgame.PlayerIndex.Next] for custom
turn order. [ScoreBehavior] integrates with [base.GameDelegate.PlayerScore] and
[base.GameDelegate.CheckGameFinished] for automatic winner determination.

Non-Trivial Logic: The behavior has methods with real logic, not just getters
and setters. [LocationBehavior] provides graph-based pathfinding
([LocationBehavior.ShortestPathTo], [LocationBehavior.Neighbors],
[LocationBehavior.DistanceTo]). [PlayerColor] provides component ownership
tracking ([PlayerColor.OwnsToken], [PlayerColor.Token]).
[Seat.SetSeatFilled] enforces a constraint (filling a seat also closes it).
[PlayerOrderBehavior.SetPlayerOrder] validates that the order is a valid
permutation.

Schelling Point: The behavior standardizes a name that multiple systems
coordinate around. [CurrentPlayerBehavior] standardizes the field name
CurrentPlayer, which [base.GameDelegate.CurrentPlayerIndex] looks for.
[PhaseBehavior] standardizes Phase. [PlayerRole] standardizes Role, which
[base.GameDelegate.GroupMembership] recognizes. Even when a behavior has minimal
logic, the naming convention it establishes enables framework features.

When NOT to create a behavior: if a piece of state is only used by game-specific
logic and no move or framework system needs to discover it via type assertion, a
plain field is simpler and more readable. A behavior that is just a field with
trivial getters adds indirection without value. The bar for a new behavior is: at
least one move or framework system will look for its companion interface, or the
logic it encapsulates would be error-prone to reimplement in each game.

# Connectable Behaviors

Behaviors often require access to the struct they're embedded within. These
types of behaviors are called [Connectable]. The framework automatically
detects embedded Connectable fields and calls ConnectBehavior on them during
state setup, before FinishStateSetUp is called. You do not need to call
ConnectBehavior yourself.

# Struct Tag Configuration

Some behaviors support declarative configuration via struct tags on the
embedding site. For example, [LocationBehavior] can be configured with a
"location" tag that names a SizedStack field on the same struct:

	type playerState struct {
		base.SubState
		behaviors.LocationBehavior `location:"Spaces"`
		Spaces boardgame.SizedStack `sizedstack:"tokens,24"`
	}

This eliminates the need for a FinishStateSetUp override in the common case.
Behaviors that support tag-based configuration implement the
[boardgame.TagConfigurable] interface, and the framework calls
ConfigureFromTags automatically after ConnectBehavior but before
FinishStateSetUp.

FinishStateSetUp is still available for wiring that cannot be expressed as
struct tags. For example, [LocationBehavior] optionally accepts a graph that
must be connected imperatively:

	func (p *playerState) FinishStateSetUp() {
		p.LocationBehavior.ConnectGraph(myConnectivityGraph)
	}

[Connectable] behaviors that are not connected will error when their
ValidConfiguration is called, and the main library will notice that while
NewGameManager is being executed, which will fail with a descriptive error.

# LocationBehavior

[LocationBehavior] tracks the position of a token within a SizedStack (where
each slot represents a space on the board). Embed it in a playerState or
gameState to gain LocationIndex(), MoveTo(), and graph-based helpers like
Neighbors(), ShortestPathTo(), and DistanceTo(). It is a [Connectable] behavior
(auto-connected by the framework). Use the "location" struct tag to point it at
a SizedStack field, or call ConnectLocationStack in FinishStateSetUp. Optionally
call ConnectGraph in FinishStateSetUp to enable graph-based queries.
[LocationBehavior] also stores a LocRemainingPath field used by the
[moves.HopAlongPath] FixUp for animated multi-hop movement. See the
[LocationBehavior] type documentation and the spatial game API section of the
tutorial for more.

# Seats, Inactivity, and Players

This package defines two behaviors, [InactivePlayer] and [Seat], whose use isn't
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

For that reason, this package introduces the notion of [Seat] and
[InactivePlayer]. Instead of thinking of the core engine's notion of a player as
a literal player, think of it as a seat, that may or may not be occupied. A seat
can be denoted as having a player sitting in it (that it is "Filled"). It can
also express that even though it is not filled, it is no longer open for anyone
to sit in it (that it is "Closed").

If the server logic sees that your game logic includes the [Seat] behavior, then
it knows that it should seek to communicate to your game logic when a user joins
the game, and listen for your game logic to communicate which open seats should
no longer be filled, even if there are new users. The only way to modify state
in your game logic is by making a move, so that's how the server package tells
your game logic that a seat is filled. If it is sees that your game logic has a
legal move type that is [moves.SeatPlayer] (or a move that embeds that move
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
For that reason there's another, related behavior called [InactivePlayer].

By default, every player slot is considered active--that is, they should be
treated as a real, normal player. When the turn order gets to them, the logic of
the game waits for them to make their move before continuing on to the other
player. But sometimes that seat is empty, and we want to get a move on without
waiting for any more players to join. Or sometimes we want to seat a player
immediately, but finish out the current round of play without them, only dealing
them in next round. The way to do that is to embed [InactivePlayer] behavior,
which contains a flag for whether the player is Inactive.

If you mark the player as inactive, then boardgame.PlayerIndex.Next (and Prev)
will skip that player, acting as though they don't exist. This means that nearly
all of your game logic (that doesn't just count the number of playerStates
naively) will operate as though they don't exist. By default all players are
considered active--even seats that are empty. That's because in general the
safest default is to wait to start the game until everyone is there.

But if you embed both [Seat] and [InactivePlayer] in your playerStates, then the
[moves.SeatPlayer] move will immediately mark any player that is seated to be inactive.
This is again a safe default; it's safest to assume that a recently seated
player needs to be 'dealt in,' likely before the next round starts, before
they're active.

There are other moves that are designed to work with this system.
[moves.ActivateInactivePlayer] will go through and activate any inactive players.
This is typically included in a phase progression just before a round starts.
[moves.CloseEmptySeat] will mark as closed any seats that are currently empty,
which effectively says "even though there are more seats, no more people may be
seated". [moves.InactivateEmptySeat] marks any empty seat as inactive, which
effectively communicates "until I say otherwise, just pretend like the empty
seats aren't there".

Because of these concepts, when you want to know the number of logical players
in your game at any moment, Game.NumPlayers() is often not what you want.
Instead, see boardgame/base.GameDelegate.NumSeatedActivePlayers.

# ScoreBehavior

[ScoreBehavior] tracks the player's game score as a simple int. Embedding it
provides a GameScore() method that satisfies [base.PlayerGameScorer], which means
[base.GameDelegate.CheckGameFinished] automatically uses this score to determine
winners. For games where score is derived from other state (e.g. counting
components in a stack), implement GameScore() directly on your playerState
instead of using this behavior.

# DrawDiscardPair

[DrawDiscardPair] tracks a draw stack and a discard stack on a gameState. It is a
[Connectable], [boardgame.TagConfigurable] behavior. Use the "draw" and "discard"
struct tags to reference existing Stack fields. The companion move
[moves.ShuffleDiscardIntoDraw] automatically reshuffles the discard pile into the
draw pile when it empties. A single anonymous pair is zero-config; named pairs
use [moves.WithDrawDiscardPairField] and receive distinct derived move names.

# PlayerTeam

[PlayerTeam] tracks which team a player belongs to via an enum field. It is a
[Connectable] behavior. It provides [PlayerTeam.TeamMembers] and
[PlayerTeam.Opponents] helpers that iterate players and filter by team,
excluding inactive players. The companion move [moves.SelectTeam] lets players
choose their team during a gathering phase. [HasPlayerTeam] is the interface
that moves and framework code use to discover the behavior via type assertion.

# PlayerRole

[PlayerRole] tracks which role a player has via an enum field named "role". It is
used for asymmetric roles like spymaster vs guesser, or unique characters. The
companion move [moves.SelectRole] lets players choose their role during a
gathering phase. [HasPlayerRole] is the discovery interface. If your role enum is
combined with the group enum, [base.GameDelegate.GroupMembership] picks it up
automatically.

# PlayerColor

[PlayerColor] tracks which color a player has via an enum field named "color". It
is a [Connectable] behavior that provides [PlayerColor.OwnsToken],
[PlayerColor.Token], and [PlayerColor.TokenFromDeck] methods for functional
color ownership (e.g., a player's armies in Risk). The companion move
[moves.SelectColor] lets players choose their color during a gathering phase.
[HasPlayerColor] is the discovery interface. SelectColor enforces uniqueness by
default — use [moves.WithAllowDuplicates] to disable.

# GameAdministrator

[GameAdministrator] tracks whether a player has game-admin authority — the
ability to perform host-like actions such as starting the game, reassigning
teams, or kicking players. The first player seated is automatically marked as
admin by [moves.SeatPlayer]. [HasGameAdministrator] is the discovery interface.
[PlayerIsAdmin] is the convenience function. Moves that should be admin-only can
embed [moves.AdminPlayer] as their base type, or individual moves can be
configured with [moves.WithRequireAdmin].

# FaceUpMarket

[FaceUpMarket] tracks a source deck and a face-up display area on a gameState. It
is a [Connectable], [boardgame.TagConfigurable] behavior. Use the "source" and
"display" struct tags to configure it. A bounded display's live MaxSize is the
default target; use "size" only for an unbounded or intentionally partial
display. The companion move
[moves.ReplenishMarket] automatically fills empty display slots from the source
deck, one component per FixUp. Multiple markets are supported via named fields
and [moves.WithMarketField], with stable distinct derived move names.

# PlayerElimination

[PlayerElimination] tracks whether a player has been eliminated (knocked out) from
play. This is distinct from [InactivePlayer]: elimination is a game concept
(this player has lost), while inactivity controls turn order. The behavior is
scope-agnostic -- the game decides when to set and clear the flag.

# MoveBudget

[MoveBudget] tracks how many actions a player has remaining in their current
turn. The bool-action pattern (has this player made their move?) is handled by a
budget of 1.
*/
package behaviors

import "github.com/jkomoros/boardgame"

// Connectable is the interface that behaviors that are Connectable implements.
// Connectable behaviors need a reference to their containing SubState. The
// framework automatically calls ConnectBehavior on all embedded Connectable
// fields during state setup, before FinishStateSetUp is called. The
// ValidConfiguration method will return an error if they weren't connected,
// which will help diagnose the problem early.
type Connectable interface {
	//ConnectBehavior lets the behavior have a reference to the struct its
	//embedded in, as some behaviors need access to the broader state. Called
	//automatically by the framework for embedded behaviors.
	ConnectBehavior(containgSubState boardgame.SubState)

	//Connectable behaviors should implement ValidConfiguration and return an
	//error if they haven't yet been Connected, which will help the main engine
	//know to fail NewGameManager, allowing the problem to be fixed more quickly.
	boardgame.ConfigurationValidator
}
