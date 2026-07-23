/*
Package moves is a convenience package that implements composable Moves to make
it easy to implement common logic. The [base.Move] type is a very simple move that
implements the basic stubs necessary for your straightforward moves to have
minimal boilerplate. Although it's technically optional, a lot of the magic
features throughout the framework depend on some if its base logic, so it's
recommended to always embed it anonymously in your move struct (or embed a
struct that embeds it).

You interact with and configure various move types by implementing interfaces.
Those interfaes are defined in the interfaces subpackage, to make this package's
design more clear.

There are many move types defined. Some are designed to be used directly with
minimal modification; others are powerful move types that are designed to be
sub-classed.

# Automatic MoveConfig Generation

Creating MoveConfig's is a necessary part of installing moves on your
GameManager, but it's verbose and error-prone. You need to create a lot of extra
structs, and then remember to provide the right properties in your config. And
to use many of the powerful moves in the moves package, you need to write a lot
of boilerplate methods to integrate correctly. Finally, you end up repeating
yourself often--which makes it a pain if you change the name of a move.

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

	func (m *MoveDealInitialCards) TargetCount(state boardgame.ImmutableState) int {
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
	            moves.WithGameProperty("DrawStack"),
	            moves.WithPlayerProperty("Hand"),
	            moves.WithTargetCount(2),
	        )
	    )
	}

# Basic Usage

[AutoConfigurer] takes an example struct representing your move, and then a list
of 0 to n [CustomConfigurationOption]. These options are given a
[boardgame.PropertyCollection] and then add specific properties to it, and then
stash that on the CustomConfiguration property of the returned MoveTypeConfig.
Different move methods will then reach into that configuration to alter the
behavior of moves of that type.

Moves that are used with [AutoConfigurer] must satisfy the AutoConfigurableMove
interface, which adds one method: DeriveName() string. [AutoConfigurer.Config]
primarily consists of some set up and then using those return values as fields
on the returned [boardgame.MoveConfig]. These methods are implemented in [Default],
which means that any move structs that embed [Default] (directly or
indirectly) can be used with [AutoConfigurer].

[Default] does a fair bit of magic in these methods to implement much of the
logic of [AutoConfigurer]. In general, if you pass a configuration option (via
[WithMoveName], for example) then that option will be used for that method.
[Default.DeriveName] also will use reflection to automatically set a
struct name like "MoveDealInitialCards" to "Deal Initial Cards". All of the
moves in the moves package will also automatically return reasonable names for
DeriveName(), so in many cases you can use those structs directly without having
to pass [WithMoveName].

Other moves in the moves package, like [DealCountComponents], will use
configuration, like [WithGameProperty], to power their default GameStack()
method.

All moves in the moves package are designed to return an error from
ValidConfiguration(), which means that if you forgot to pass a required
configuration property (e.g. you don't override GameStack and also don't provide
[WithGameProperty]), when you try to create NewGameManager() and all moves'
ValidConfiguration() is checked, you'll get an error. This helps catch
mis-configurations during boot time.

Refer to the documentation of the various methods in that package for their
precise behavior and how to configure them.

# Idiomatic Move Definition and Installation

AutoConfigurer is at the core of idiomatic definition and installation of moves,
and typically is used for every move you install in your game. The following
paragraphs describe the high-level idioms to follow.

Never create your own [boardgame.MoveConfig] objects--it's just another global variable that
clutters up your code and makes it harder to change. Instead, use
[AutoConfigurer]. There are some rare cases where you do want to refer to the move
by name (and not rely on finicky string-based lookup), such as when you want an
Agent to propose a speciifc type of move. In those cases use [AutoConfigurer] to
create the move type config, then save the resulting config's Name to a global
variable that you use elsewhere, and then pass the created config to [Add]
(and its cousins)

In general, you should only create a bespoke Move struct in your game if it is
not possible to use one of the off-the-shelf moves from the moves package,
combined with configuarion options, to do what you want. In practice this means
that only if you need to override a method on one of the base moves do you need
to create a bespoke struct. This typically allows you to drastically reduce the
number of bespoke move structs your game defines, saving thousands of lines of
code (each bespoke struct also has hundreds of lines of auto-generated
PropertyReader code).

If you do create a bespoke struct, name it like this: "MoveNameOfMyMove", so
that [Default.DeriveName] will give it a reasonable name
automatically (in this example, "Name Of My Move").

Anonymous embedding requires no binding boilerplate. The engine records the
final constructed move in its [boardgame.MoveInfo], so framework methods on an
embedded move can discover capabilities and invoke overrides implemented by
the outer move. This is automatic for game authors. Authors of reusable move
frameworks may use move.Info().ConcreteMove() when they deliberately need this
final-dispatch behavior; it should not be used as a general service locator.
Framework tests and tools may use [boardgame.NewOrphanMove] when they need this
affiliation without a game or manager.

In many cases if you extend powerful moves like [DealCountComponents] the
default HelpText() value is sufficient (especially if it's a [FixUp] move that
won't ever be seen by players). In other cases, [WithHelpText] is often the only
config option you will pass to [AutoConfigurer].

If your move will be a FixUp move that doesn't embed one of the more advanced
fix up moves (like [RoundRobin] or [DealCountComponents]), embed [FixUp] into
your struct. That will cause IsFixUp to return the right value even without
using [WithIsFixUp]--because [WithIsFixUp] is easy to forget given that it's often
in a different file. In almost all cases if you use [WithIsFixUp] you should
simply embed [FixUp] instead.

[AutoConfigurer.MustConfig] is like [AutoConfigurer.Config], but instead of
returning a [boardgame.MoveConfig] and an error, it simply returns a [boardgame.MoveConfig]--and panics
if it would have returned an error. Since your GameDelegate's ConfigureMoves()
is typically called during the boot-up sequence of your game, it is safe to use
[AutoConfigurer.MustConfig] exclusively, which saves many lines of boilerplate
error checking.

# Configure Move Helpers

Your Game Delegate's ConfigureMoves() []boardgame.MoveConfig is where the action
happens for installing moves. In practice you can do whatever you want in there
as long as you return a list of MoveConfigs. In practice you often use
AutoConfigurer (see section above). If you have a very simple game type you
might not need to do anythign special.

If, however, your game type is complicated enough to need the notion of phases,
then you'll probably want to use some of the convenience methods for installing
moves: [Combine], [Add], [AddForPhase], and [AddOrderedForPhase]. These methods make
sure that enough information is stored for the Legal() methods of those moves to
know when the move is legal. Technically they're just convenience wrappers (each
describes the straightforward things it's doing), but in practice they're the
best way to do it. See the tutorial in the main package for more.

# Declarative Legality

[Default], [CurrentPlayer], [FixUp], [FixUpMulti], and [StartPhase] support an additional, optional way to express
a move's legality: instead of (or alongside) overriding Legal(), pass
[WithLegalPreconditions] to auto.Config with one or more legal.Spec values built
from the [legal] package's predicate catalog:

	auto.MustConfig(
	    new(moveRevealCard),
	    moves.WithLegalPreconditions(
	        legal.PropAtLeast("player.CardsLeftToReveal", 1),
	        legal.RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
	        legal.MayMoveToSameSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
	    ),
	)

This is purely sugar: existing imperative behavior is unchanged for any move that doesn't
pass WithLegalPreconditions. A move that does opt in gets its checks assembled
into a plan at NewGameManager (base-derived checks — phase, move-progression,
stack constraints, and for CurrentPlayer the proposer check — first, then the
authored WithLegalPreconditions specs in order); Default.Legal() then evaluates
that plan instead of running the old chain. Memoized state-only checks retain
that same observable order. Calling [WithLegalPreconditions] with no specs
opts in a contributed-only move; implementing boardgame.CustomLegaler opts in
automatically. [WithoutLegalPrecondition] suppresses
one of those base-derived checks by its stable name (pass the exported
constants: [PreconditionInPhase], [PreconditionInProgression],
[PreconditionStackConstraints], [PreconditionProposerIsCurrentPlayer]) for a
move that wants to opt out of something it would otherwise inherit — the
[ForceFinishTurn] "inherit nothing" pattern, now expressible without a
bespoke base type. Suppressions are boot-validated: a name the move doesn't
actually contribute is a NewGameManager error naming the move
(see [WithoutLegalPrecondition]'s own doc). For legality that isn't a simple relation over a property
path — arithmetic, graph walks, anything with real Go logic —
[boardgame.CustomLegaler]'s LegalCustom method runs as imperative residue
after every declared precondition passes; see the [legal] package doc for the
full authoring guide, the catalog's rules of growth, and its honest limits
(the supported seam is [Default], [CurrentPlayer], [FixUp], [FixUpMulti],
and [StartPhase]; other framework move bases remain opaque).
See the tutorial's "Declarative Move Legality" section for a complete worked
example.

# Move Type Hierarchy

The moves in this package are all defined as a hierarchy of structs that
anonymously embed higher level structs, overriding, modifying, and extending the
behavior of the struct they embed.

You can use many of these moves directly, using the configuration options like
[WithSourceProperty] to configure which properties they should operate on.
Alternatively, you can embed them in your own move struct, overriding or
tweaking their behavior, perhaps adding an additional check to their Legal
check.

For convenience, here's the type hierarchy, with a brief description of the diff
each has on the one above it. See the documentation for each struct for more.

	base.Move
	└ Default
	  ├ Done
	  ├ CurrentPlayer
	  │ └ MoveOnGraph
	  ├ AnyPlayer
	  │ ├ AdminPlayer
	  │ ├ SelectTeam
	  │ ├ SelectRole
	  │ └ SelectColor
	  ├ CloseAllSeats
	  ├ SeatPlayer
	  └ FixUp
	    ├ NoOp
	    ├ Increment
	    ├ ShuffleStack
	    ├ StartPhase
	    ├ FinishTurn
	    ├ HopAlongPath
	    ├ AdvanceToken
	    ├ WaitForEnoughPlayers
	    └ FixUpMulti
	      ├ DefaultComponent
	      ├ ActivateInactivePlayer
	      ├ ActivateEmptySeat
	      ├ CloseEmptySeat
	      ├ InactivateEmptySeat
	      └ ApplyUntil
	        └ ApplyUntilCount
	          ├ ApplyCountTimes
	          │ ├ MoveCountComponents
	          │ │ ├ MoveComponentsUntilCountReached
	          │ │ └ MoveComponentsUntilCountLeft
	          │ │   └ MoveAllComponents
	          │ ├ RoundRobin
	          │ │ └ RoundRobinNumRounds
	          │ └ DealCountComponents
	          │   ├ DealComponentsUntilPlayerCountReached
	          │   │ ├ CollectComponentsUntilPlayerCountReached
	          │   │ └ CollectComponentsUntilPlayerCountLeft
	          │   │   └ CollectAllComponents
	          │   ├ CollectComponentsUntilPlayerCountLeft
	          │   ├ DealComponentsUntilGameCountLeft
	          │   │ ├ DealAllComponents
	          │   │ └ CollectComponentsUntilGameCountLeft
	          │   └ CollectCountComponents
	          └ (your custom ApplyUntilCount extensions)

# Gathering Phase Moves

The gathering system provides moves for the "lobby" experience — waiting for
players, team/role/color selection, and starting the game. There is no special
"lobby mode"; a gathering phase is just a normal phase where these moves are
legal. The client auto-renders appropriate UI when it detects these moves.

[AnyPlayer] is the base move for self-selection. Like [CurrentPlayer], it has a
TargetPlayerIndex, but it allows any seated player to propose the move for
themselves (not just the current player). Use it for any move where the proposer
acts on their own behalf during a simultaneous phase.

[AdminPlayer] embeds AnyPlayer and additionally requires that the proposer is
the game administrator (the player whose [behaviors.GameAdministrator].IsAdmin
is true). Use it for host-only actions like starting the game, kicking players,
or changing configuration. If the playerState does not embed
[behaviors.GameAdministrator], AdminPlayer behaves identically to AnyPlayer.
Individual moves can also be configured with [WithRequireAdmin] for the same
effect without changing their base type.

[SelectTeam], [SelectRole], and [SelectColor] embed AnyPlayer and provide
built-in selection for players who have the corresponding behaviors embedded in
their playerState ([behaviors.PlayerTeam], [behaviors.PlayerRole],
[behaviors.PlayerColor]). Each validates the selection against the named enum and
optionally enforces uniqueness ([WithUnique] / [WithAllowDuplicates]).

[GatheringMoves] is a convenience function that auto-detects which selection
behaviors are present and returns the corresponding MoveConfigs. Usage:

	moves.AddForPhase(phaseGathering, moves.GatheringMoves(auto)...)

This registers the selection moves as legal in any order during the gathering
phase. Use [AddForPhase] (not [AddOrderedForPhase]) so players can pick teams,
roles, and colors freely at any time.

The delegate method ReadyToStart(state) error is called by both
[WaitForEnoughPlayers] and [CloseAllSeats] to validate that the game's
configuration is complete before proceeding. The default returns nil. Override
it to add custom validation (e.g., "each team needs at least 2 players").

If your game needs an "unset" sentinel to detect players who haven't made a
selection, define your enum with a sentinel first value:

	const (
	    teamUnset = iota // sentinel: no team selected yet
	    teamRed
	    teamBlue
	)

Then check for the sentinel in your ReadyToStart implementation.

# Move Deal and Collect Component Moves

Generally when moving components from one place to another it makes sense to
move one component at a time, so that each component is animated separately.
However, this is a pain to implement, because it requires implementing a move
that knows how many times to apply itself in a row, which is finicky and error-
prone.

This collection is intentionally different from [boardgame.Stack.MoveCountTo].
MoveCountTo transfers exactly N components atomically within its caller's
enclosing game move; it does not create a move or persistence boundary itself.
The move types below produce a distinct persisted game move and animation
boundary for each component, which is usually preferable when dealing cards or
showing a sequence of token movements.

There is a collection of 9 moves that all do basically the same thing for moving
components, one at a time, from stack to stack. Move-type moves move components
between two specific stacks, often both on your GameState. Deal and Collect type
moves move components between a stack in GameState and a stack in each Player's
PlayerState. Deal-type moves move components from the game stack to the player
stack, and Collect-type moves move components from each player to the GameState.

All of these moves define a way to specify the source and destination stack. For
Move-type moves, you configure these via [WithSourceProperty] and
[WithDestinationProperty] (or override SourceStack() and DestinationStack()
directly). For Deal and Collect-type moves, you use [WithGameProperty] and
[WithPlayerProperty] (or override GameStack() and PlayerStack()).

All moves in this collection implement TargetCount() int, and all of them
default to 1. Override this if you want a different number of components checked
for in the end condition.

In practice you'll often use [WithTargetCount], [WithGameProperty], and friends as
configuration to [AutoConfigurer.Config] instead of overriding those yourself. In
fact, in many cases configuration options are powerful enough to allow you to
use these moves types on their own directly in your game. See the documentation
in the sections above for more examples.

Each of Move, Deal, and Collect have three variants based on the end condition.
Note that Move-type moves have only two stacks, but Deal and Collect type moves
operate on n pairs of stacks, where n is the number of players in the game. In
general for Deal and Collect type moves, the condition is met when all pairs of
stacks meet the end condition.

{Move,Deal,Collect}CountComponents apply that many moves and validate each
scheduled transfer before it occurs. With its default Apply implementation,
MoveCountComponents additionally preflights its complete remaining sequence
before the first transfer. An embedding move that overrides Apply is responsible
for matching its own Legal preflight to that custom mutation. Move names that end
in CountReached operate until the destination stacks all have TargetCount or more
items. Move names that end in CountLeft operate until the source stacks all have
TargetCount or fewer items in them.

The complete-remainder guarantee intentionally performs N+(N-1)+...+1 checks
over an N-component sequence. A constrained destination also requires a
disposable whole-state copy whenever more than one component remains in a
proposal; single-component checks use a direct live-state preflight. For a large
transfer that does not need separate move records and animations, prefer one
[boardgame.Stack.MoveCountTo] call and its matching May preflight.

Since a common configuration of these moves is to use
{Move,Deal,Collect}ComponentsUntil*Reached with a TargetCount of 0, each also
provides a *AllComponents as sugar.

# Groups

Groups allow you to specify a specific set of moves that must occur in a given
order. You pass them to [AddOrderedForPhase]. All of the groups are of type
[MoveProgressionGroup], and this package defines 5: [Serial], [Parallel],
[ParallelCount], [Repeat], and [Optional]. They can be nested as often as you'd like
to express the semantics of your move progression.

They are defined as functions that return anonymous underlying structs so that
when used in configuration you can avoid needing to wrap your children list with
[][MoveProgressionGroup], saving you typing.

	//Example

	//AddOrderedForPhase accepts move configs from auto.Config, or
	//groups.
	moves.AddOrderedForPhase(PhaseNormal,
	    //Top level groups are all joined implicitly into a group.Serial.
	    auto.MustConfig(new(MoveZero)),
	    moves.Serial(
	        auto.MustConfig(new(MoveOne)),
	        moves.Optional(
	            moves.Serial(
	                auto.MustConfig(new(MoveTwo)),
	                auto.MustConfig(new(MoveThree)),
	            ),
	        ),
	        moves.ParallelCount(
	            CountAny(),
	            auto.MustConfig(new(MoveFour)),
	            auto.MustConfig(new(MoveFive)),
	            moves.Repeat(
	                CountAtMost(2),
	                moves.Serial(
	                    auto.MustConfig(new(MoveSix)),
	                    auto.MustConfig(new(MoveSeven)),
	                ),
	            ),
	        ),
	    ),
	)

Move names must be unique, but sometimes you want to use the same underlying
move at multiple points in a progression. [WithMoveNameSuffix] is useful for that
case.

# Spatial Game Moves

This package provides three move types for games with spatial mechanics (boards
with spaces, tokens that move between spaces, NPC patrol routes, etc.).

[MoveOnGraph] is a player-facing move (embeds [CurrentPlayer]) for moving a
player's token to a destination on a graph. The player specifies a
TargetLocation; the framework computes the shortest path via Dijkstra and
stores it on the player's [behaviors.LocationBehavior].LocRemainingPath. The
embedding move must implement interfaces.LocationProvider to identify which
LocationBehavior to use. It may optionally implement [interfaces.SpaceValidator]
(to reject certain spaces), [interfaces.MovementBudgeter] (to limit movement
distance), [interfaces.FreeMovePredicate] (for teleport/card-based movement that
bypasses adjacency), and [interfaces.FreeMoveApplier] (for cleanup after a free
move).

[HopAlongPath] is a [FixUp] that executes one hop of the stored path per
application. Each hop produces a separate game version, giving the client
a distinct animation frame for each step of movement (similar to how
[DealCountComponents] animates one card at a time). Register [HopAlongPath] before
other FixUps in ConfigureMoves() so hops complete before other FixUps fire.

[AdvanceToken] is a [FixUp] for deterministic, non-player-driven token movement
(e.g., an NPC that patrols a route). The embedding move must implement
[interfaces.AdvanceCondition] (to gate when advancement
occurs) and [interfaces.PostAdvanceHandler] (for side effects after the token
moves).

See the spatial game API section of the tutorial for a worked example using
these moves with [behaviors.LocationBehavior] and enum/graph.

# Seats and Inactive Players

For more on the concepts of seats and inactive players, see the package doc of
boardgame/behaviors.
*/
package moves
