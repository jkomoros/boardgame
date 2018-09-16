/*

moves is a convenience package that implements composable Moves to make it
easy to implement common logic. The Base move type is a very simple move that
implements the basic stubs necessary for your straightforward moves to have
minimal boilerplate. Although it's technically optional, a lot of the magic
features throughout the framework depend on some if its base logic, so it's
recommended to always embed it anonymously in your move struct (or embed a
struct that embeds it).

You interact with and configure various move types by implementing interfaces.
Those interfaes are defined in the interfaces subpackage, to make this
package's design more clear.

There are many move types defined. Some are designed to be used directly with
minimal modification; others are powerful move types that are designed to be
sub-classed.

Automatic MoveConfig Generation

Creating MoveConfig's is a necessary part of installing moves on your
GameManager, but it's verbose and error-prone. You need to create a lot of
extra structs, and then remember to provide the right properties in your
config. And to use many of the powerful moves in the moves package, you
need to write a lot of boilerplate methods to integrate correctly.
Finally, you end up repeating yourself often--which makes it a pain if you
change the name of a move.

Take this example:

	//boardgame:codegen
	type MoveDealInitialCards struct {
		moves.DealComponentsUntilPlayerCountReached
	}

	var moveDealInitialCardsConfig = boardgame.MoveConfig {
		Name: "Deal Initial Cards",
		Constructor: func() boardgame.Move {
			return new(MoveDealInitialCards)
		},
	}

	func (m *MoveDealInitialCards) GameStack(gState boardgame.MutableSubState) boardgame.MutableStack {
		return gState.(*gameState).DrawStack
	}

	func (m *MoveDealInitialCards) PlayerStack(pState boardgame.MutablePlayerState) boardgame.MutableStack {
		return pState.(*playerState).Hand
	}

	func (m *MoveDealInitialCards) TargetCount() int {
		return 2
	}

	func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
		return moves.Add(
			&moveDealInitialCardsConfig,
		)
	}

auto.Config (and its panic-y sibling auto.MustConfig) help reduce this
signficantly:

	func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

		auto := moves.NewAutoConfigurer(g)

		return moves.Add(
			auto.MustConfig(
				new(moves.DealComponentsUntilPlayerCountReached),
				moves.WithGameStack("DrawStack"),
				moves.WithPlayerStack("Hand"),
				moves.WithTargetCount(2),
			)
		)
	}

Basic Usage

AutoConfigurer takes an example struct representing your move, and then a
list of 0 to n interfaces.CustomConfigurationOption. These options are
given a boardgame.PropertyCollection and then add specific properties to
it, and then stash that on the CustomConfiguration property of the
returned MoveTypeConfig. Different move methods will then reach into that
configuration to alter the behavior of moves of that type.

The moves/with package defines a large collection of
CustomConfigurationOption for use with the moves in the moves package.

Moves that are used with AutoConfigurer must satisfy the AutoConfigurableMove
interface, which adds one method: DeriveName() string. AutoConfigurer.Config()
primarily consists of some set up and then using those return values as
fields on the returned MoveConfig. These methods are implemented in
moves.Base, which means that any move structs that embed moves.Base
(directly or indirectly) can be used with AutoConfigurer.

moves.Base does a fair bit of magic in these methods to implement much of
the logic of AutoConfigurer. In general, if you pass a configuration option
(via WithMoveName, for example) then that option will be used for that
method. moves.Base.DeriveName() also will use reflection to automatically
set a struct name like "MoveDealInitialCards" to "Deal Initial Cards". All
of the moves in the moves package will also automatically return
reasonable names for DeriveName(), so in many cases you can use those
structs directly without having to pass WithMoveName().

Other moves in the moves package, like DealCountComponents, will use
configuration, like WithGameStack(), to power their default GameStack()
method.

All moves in the moves package are designed to return an error from
ValidConfiguration(), which means that if you forgot to pass a required
configuration property (e.g. you don't override GameStack and also don't
provide WithGameStack), when you try to create NewGameManager() and all
moves' ValidConfiguration() is checked, you'll get an error. This helps
catch mis-configurations during boot time.

Refer to the documentation of the various methods in that package for
their precise behavior and how to configure them.

Idiomatic Move Definition and Installation

AutoConfigurer is at the core of idiomatic definition and installation of
moves, and typically is used for every move you install in your game. The
following paragraphs describe the high-level idioms to follow.

Never create your own MoveConfig objects--it's just another global
variable that clutters up your code and makes it harder to change.
Instead, use AutoConfigurer. There are some rare cases where you do want to
refer to the move by name (and not rely on finicky string-based lookup),
such as when you want an Agent to propose a speciifc type of move. In
those cases use AutoConfigurer to create the move type config, then save the
resulting config's Name to a global variable that you use elsewhere, and
then pass the created config to moves.Add() (and its cousins)

In general, you should only create a bespoke Move struct in your game if
it is not possible to use one of the off-the-shelf moves from the moves
package, combined with configuarion options, to do what you want. In
practice this means that only if you need to override a method on one of
the base moves do you need to create a bespoke struct. This typically
allows you to drastically reduce the number of bespoke move structs your
game defines, saving thousands of lines of code (each bespoke struct also
has hundreds of lines of auto-generated PropertyReader code).

If you do create a bespoke struct, name it like this: "MoveNameOfMyMove",
so that moves.Base's default DeriveName() will give it a reasonable name
automatically (in this example, "Name Of My Move").

In many cases if you subclass powerful moves like DealCountComponents the
default HelpText() value is sufficient (especially if it's a FixUp
move that won't ever be seen by players). In other cases, WithHelpText()
is often the only config option you will pass to AutoConfigurer.

If your move will be a FixUp move that doesn't sublcass one of the more
advanced fix up moves (like RoundRobin or DealCountComponents), embed
moves.FixUp into your struct. That will cause IsFixUp to return the right
value even without using WithIsFixUp--because WithIsFixUp is easy to
forget given that it's often in a different file. In almost all cases if
you use WithIsFixUp you should simply embed moves.FixUp instead.

AutoConfigurer.MustConfig is like AutoConfigurer.Config, but instead of returning a MoveConfig
and an error, it simply returns a MoveConfig--and panics if it would have
returned an error. Since your GameDelegate's ConfigureMoves() is typically
called during the boot-up sequence of your game, it is safe to use
AutoConfigurer.MustConfig exclusively, which saves many lines of boilerplate error
checking.

Configure Move Helpers

Your Game Delegate's ConfigureMoves() []boardgame.MoveConfig is where the
action happens for installing moves. In practice you can do whatever you want
in there as long as you return a list of MoveConfigs. In practice you often
use AutoConfigurer (see section above). If you have a very simple game type you
might not need to do anythign special.

If, however, your game type is complicated enough to need the notion of
phases, then you'll probably want to use some of the convenience methods for
installing moves: Combine, Add, AddForPhase, and AddOrderedForPhase. These
methods make sure that enough information is stored for the Legal() methods of
those moves to know when the move is legal. Technically they're just
convenience wrappers (each describes the straightforward things it's doing),
but in practice they're the best way to do it. See the tutorial in the main
package for more.

Base Move

Implementing a Move requires a lot of stub methods to implement the
boardgame.Move interface, but also a lot of logic, especially to support
Phases correctly. moves.Base is a move that all moves should embed somewhere
in their hierarchy. It is very important to always call your superclasse's
Legal(), because moves.Base.Legal contains important logic to implement phases
and ordered moves within phases.

Base also includes a number of methods needed for moves to work well with
AutoConfigurer.

FixUp Move

FixUp moves are simple embedding of move.Base, but they default to having
IsFixUp generated by AutoConfigurer be true instead of false. This is useful so
you don't forget to pass WithIsFixUp(true) yourself in AutoConfigurer.Config.

FixUpMulti is the same as FixUp, but also has a AllowMultipleInProgression
that returns true, meaning that the ordered move logic within phases will
allow multiple of this move type to apply in a row.

Default Component Move

DefaultComponent is a move type that, in DefaultsForState, searches through
all of the components in the stack provided with WithSourceStack, and testing
the Legal() method of each component. It sets the first one that returns nil
to m.ComponentIndex. Its Legal() returns whether there is a valid component
specified, and what its Legal returns. You provide your own Apply().

It's useful for fixup moves that need to apply actions to components in a
given stack when certain conditions are met--for example, crowning a token
that makes it to the opposite end of a board in checkers.

The componentValues.Legal() takes a legalType. This is the way you can use
multiple DefaultComponent moves for the same type of component. If you only
have one then you can skip passing WithLegalType, and just default to 0. If
you do have multiple legalTypes, the idiomatic way is to have those be members
of an Enum for that purpose.

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

In practice you'll often use WithTargetCount, WithGameStack, and friends as
configuration to AutoConfigurer.Config instead of overriding those yourself. In fact, in
many cases configuartion options are powerful enough to allow you to use these
moves types on their own directly in your game. See the documentation in
the sections above for more examples.

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
player, resetting state as appropriate. In practice you often can use this
move directly in your game without even passing any WithOPTION configuration
to AutoConfigurer.Config.

StartPhase

The StartPhase move is designed to set your game's phase to the next phase.
It's generally used as the last move in an ordered phase, for example, the
last move in your game's SetUp phase. This move can also generally be used
directly in your game, by using the WithPhaseToStart configuration option in
AutoConfigurer.Config.

ShuffleStack

Shuffle stack is a simple move that just shuffles the stack denoted by
SourceStack.

*/
package moves
