/*

moves is a convenience package that implements composable Moves to make it
easy to implement common logic. The Base move type is a very simple move that
implements the basic stubs necessary for your straightforward moves to have
minimal boilerplate.

You interact with and configure various move types by implementing interfaces.
Those interfaes are defined in the moveinterfaces subpackage, to make this
package's design more clear.

There are many move types defined. Some are designed to be used directly with
minimal modification; others are powerful move types that are designed to be
sub-classed.

DefaultConfig

In a number of cases you'll just use moves in this package will minimal
overriding. In those cases it's a pain to write up MoveTypeConfigs because
it's mostly boilerplate.

DefaultConfig is a package-level constructor that takes a manager, and an
example move struct that embeds moves from this pacakge, and returns a default
config object with reaonsable defaults so you don't have to. It uses
move.MoveTypeName and move.MoveTypeHelpText to generate strings, which moves
in this package normally do reasonable things with. Example use:

	type myMove struct {
		moves.DealCountComponents
	}

	func (m *myMove) GameStack(gState boardgame.MutableSubState) boardgame.MutableStack {
		return gState.(*gameState).DrawStack
	}

	func (m *myMove) PlayerStack(pState boardgame.MutablePlayerState) boardgame.MutableStack {
		return pState.(*playerState).Hand
	}

	func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {

		//...
		).AddMoves(
			//Name, HelpText, MoveConstructor, and IsFixUp will be set reasonably.
			moves.MustDefaultConfig(g.Manager(), new(myMove)),
		)
		//...

	}

StartPhase is special because often you want to override just one small part,
so we provide NewStartPhaseConfig.

Base Move

Implementing a Move requires a lot of stub methods to implement the
boardgame.Move interface, but also a lot of logic, especially to support
Phases correctly. moves.Base is a move that all moves should embed somewhere
in their hierarchy. It is very important to always call your superclasse's
Legal(), because moves.Base.Legal contains important logic to implement phases
and ordered moves within phases.

Current Player Move

These moves are for moves that are only legal to be made by the current
player. Their Legal() will verify that it is the proposer's turn.

Move Deal and Collect Component Moves

Generally when moving components from one place to another it makes sense to
move one component at a time, so that each component is animated separately.
However, this is a pain to implement, because it requires implementing a move
that knows how many times to apply itself in a row, which is fincky and error-
prone.

There is a collection of 9 moves that all do basically the same thing for
moving components, one at a time, from stack to stack. Move-type moves move
components between two specific stacks, often both on your GameState. Deal and
Collect type moves move components between a stack in GameState and a stack in
each Player's PlayerState. Deal-type moves move components from the game stack
to the player stack, and Collect-type moves move components from each player
to the GameState.

All of these moves define a way to define the source and destination stack.
For Move-type moves, you define SourceStack() and DestinationStack(). For Deal
and Collect-type moves, you implement GameStack() and PlayerStack().

All moves in this collection implement TargetCount() int, and all of them
default to 1. Override this if you want a different number of components
checked for in the end condition.

Each of Move, Deal, and Collect have three variants based on the end
condition. Note that Move-type moves have only two stacks, but Deal and
Collect type moves operate on n pairs of stacks, where n is the number of
players in the game. In general for Deal and Collect type moves, the condition
is met when all pairs of stacks meet the end condition.

{Move,Deal,Collect}CountComponents simply apply that many moves without regard
to the number of components in the source or destination stacks. Move names
that end in CountReached operate until the destination stacks all have
TargetCount or more items. Move names that end in CountLeft operate until the
source stacks all have TargetCount or fewer items in them.

ApplyUntil ApplyUntilCount and ApplyCountTimes

These moves are what the MoveComponents moves are based off of and are
designed to be subclassed. They apply the move in question until some
condition is reached.

RoundRobin and RoundRobinNumRounds

Round Robin moves are like ApplyUntilCount and friends, except they go around
and operate on each player in succession. RoundRobinNumRounds goes around each
player until NumRounds() cycles have completed. The base RoundRobin goes
around until the PlayerCondition has been met for each player. These are the
most complicated moves in the set; if you subclass one directly you're most
likely to subclass RoundRobinNumRounds.

FinishTurn

FinishTurn is a move that is designed to be used as a fix-up move during
normal phases of play in your game. It checks whether the current player's
turn is done (based on criteria you specify) and if so advances to the next
player, resetting state as appropriate.

StartPhase

The StartPhase move is designed to set your game's phase to the next phase.
It's generally used as the last move in an ordered phase, for example, the
last move in your game's SetUp phase. In general you don't need to subclass
this move directly at all; just use NewStartPhaseConfig to get a
MoveTyepConfig that does what you want.

ShuffleStack

Shuffle stack is a simple move that just shuffles the stack denoted by
SourceStack.

*/
package moves
