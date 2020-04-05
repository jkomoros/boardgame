/*

Package moves is a convenience package that implements composable Moves to make
it easy to implement common logic. The Base move type is a very simple move that
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

Automatic MoveConfig Generation

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
                moves.WithGameProperty("DrawStack"),
                moves.WithPlayerProperty("Hand"),
                moves.WithTargetCount(2),
            )
        )
    }

Basic Usage

AutoConfigurer takes an example struct representing your move, and then a list
of 0 to n interfaces.CustomConfigurationOption. These options are given a
boardgame.PropertyCollection and then add specific properties to it, and then
stash that on the CustomConfiguration property of the returned MoveTypeConfig.
Different move methods will then reach into that configuration to alter the
behavior of moves of that type.

The moves/with package defines a large collection of CustomConfigurationOption
for use with the moves in the moves package.

Moves that are used with AutoConfigurer must satisfy the AutoConfigurableMove
interface, which adds one method: DeriveName() string. AutoConfigurer.Config()
primarily consists of some set up and then using those return values as fields
on the returned MoveConfig. These methods are implemented in moves.Default,
which means that any move structs that embed moves.Default (directly or
indirectly) can be used with AutoConfigurer.

moves.Default does a fair bit of magic in these methods to implement much of the
logic of AutoConfigurer. In general, if you pass a configuration option (via
WithMoveName, for example) then that option will be used for that method.
moves.Default.DeriveName() also will use reflection to automatically set a
struct name like "MoveDealInitialCards" to "Deal Initial Cards". All of the
moves in the moves package will also automatically return reasonable names for
DeriveName(), so in many cases you can use those structs directly without having
to pass WithMoveName().

Other moves in the moves package, like DealCountComponents, will use
configuration, like WithGameProperty(), to power their default GameStack()
method.

All moves in the moves package are designed to return an error from
ValidConfiguration(), which means that if you forgot to pass a required
configuration property (e.g. you don't override GameStack and also don't provide
WithGameProperty), when you try to create NewGameManager() and all moves'
ValidConfiguration() is checked, you'll get an error. This helps catch
mis-configurations during boot time.

Refer to the documentation of the various methods in that package for their
precise behavior and how to configure them.

Idiomatic Move Definition and Installation

AutoConfigurer is at the core of idiomatic definition and installation of moves,
and typically is used for every move you install in your game. The following
paragraphs describe the high-level idioms to follow.

Never create your own MoveConfig objects--it's just another global variable that
clutters up your code and makes it harder to change. Instead, use
AutoConfigurer. There are some rare cases where you do want to refer to the move
by name (and not rely on finicky string-based lookup), such as when you want an
Agent to propose a speciifc type of move. In those cases use AutoConfigurer to
create the move type config, then save the resulting config's Name to a global
variable that you use elsewhere, and then pass the created config to moves.Add()
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
that moves.Default's default DeriveName() will give it a reasonable name
automatically (in this example, "Name Of My Move").

In many cases if you subclass powerful moves like DealCountComponents the
default HelpText() value is sufficient (especially if it's a FixUp move that
won't ever be seen by players). In other cases, WithHelpText() is often the only
config option you will pass to AutoConfigurer.

If your move will be a FixUp move that doesn't sublcass one of the more advanced
fix up moves (like RoundRobin or DealCountComponents), embed moves.FixUp into
your struct. That will cause IsFixUp to return the right value even without
using WithIsFixUp--because WithIsFixUp is easy to forget given that it's often
in a different file. In almost all cases if you use WithIsFixUp you should
simply embed moves.FixUp instead.

AutoConfigurer.MustConfig is like AutoConfigurer.Config, but instead of
returning a MoveConfig and an error, it simply returns a MoveConfig--and panics
if it would have returned an error. Since your GameDelegate's ConfigureMoves()
is typically called during the boot-up sequence of your game, it is safe to use
AutoConfigurer.MustConfig exclusively, which saves many lines of boilerplate
error checking.

Configure Move Helpers

Your Game Delegate's ConfigureMoves() []boardgame.MoveConfig is where the action
happens for installing moves. In practice you can do whatever you want in there
as long as you return a list of MoveConfigs. In practice you often use
AutoConfigurer (see section above). If you have a very simple game type you
might not need to do anythign special.

If, however, your game type is complicated enough to need the notion of phases,
then you'll probably want to use some of the convenience methods for installing
moves: Combine, Add, AddForPhase, and AddOrderedForPhase. These methods make
sure that enough information is stored for the Legal() methods of those moves to
know when the move is legal. Technically they're just convenience wrappers (each
describes the straightforward things it's doing), but in practice they're the
best way to do it. See the tutorial in the main package for more.

Move Type Hierarchy

The moves in this package are all defined as a hierarchy of structs that
anonymously embed higher level structs, overriding, modifying, and extending the
behavior of the struct they embed.

You can use many of these moves directly, using the configuration options like
WithSourceProperty() to configure which properties they should operate on.
Alternatively, you can embed them in your own move struct, overriding or
tweaking their behavior, perhaps adding an additional check to their Legal
check.

For convenience, here's the type hierarchy, with a brief description of the diff
each has on the one above it. See the documentation for each struct for more.

    * base.Move - The simplest, unopinonated stub of a move, from the base package.
        * Default - Substantial base logic, including base property overriding for with and especially in Legal() around move progressions and phases.
            * Done - A simple move that does nothing in its Apply and has no extra Legal() logic, meaning it's primarily a non-fix-up move applied by a player to move out of a move progression.
            * CurrentPlayer - Defaults to the GameDelegate.CurrentPlayerIndex, and only lets the move be made if it's on behalf of that player.
            * SeatPlayer - A special move that the server package uses to tell the game logic that a new player has been added to the game.
            * FixUp - Overrides IsFixUp() to always return true, making the move eligible for base.GameDelegate.ProposeFixUpMove.
                * NoOp - A move that does nothing. Useful for specific edge cases of MoveProessionMatching, and also to signal to AddOrderedForPhase that the lack of a StartPhase move was intentional.
                * Increment - Increments the provided SourceProperty by Amount. Useful to run automatically at a given spot in a move progression.
                * ShuffleStack - Shuffles the stack at SourceProperty. Useful to run automatically at a certain time in a MoveProgression.
                * StartPhase - Calls BeforeLeavePhase, then BeforeEnterPhase, then SetCurrentPhase. Generally you have one of these at the end of an AddOrderedForPhase.
                * FinishTurn - Checks if State.CurrentPlayer().TurnDone() is true, and if so increments CurrentPlayerIndex to the next player, calling playerState.ResetForTurnEnd() and then ResetForTurnStart.
                * FixUpMulti - Overrides AllowMultipleInProgression() to true, meaning multiple of the same move are legal to apply in a row according to Deafult.Legal()
                    * DefaultComponent - Looks at each component in SourceStack() and sees which one's method of Legal() returns nil, selecting that component for you to operate on in your own Apply.
                    * ActivateInactivePlayer - Activates any players who are not currently active
                    * CloseEmptySeat - Closes any seats that are still marked as empty, so they won't be filled by the server.
                    * ApplyUntil - Legal() returns nil only once ConditionMet() returns nil.
                        * ApplyUntilCount - Supplies a ConditionMet that returns true when Count() is the same as TargetCount().
                            * ApplyCountTimes - Supplies a Count() that is the number of times this move has been applied in a row.
                                * MoveCountComponents - Moves components from SourceStack to DestinationStack until TargetCount have been moved.
                                    * MoveComponentsUntilCountReached - Overrides Count to be how many components are in DestinatinoStack
                                    * MoveComponentsUntilCountLeft - Overrides Count to be how many components are left in SourceStack
                                        * MoveAllComponents - Overrides TargetCount to be 0
                        * RoundRobin - Applies around and aroudn for each player until PlayerConditionMet returns true for all. You must embed RoundRobinGameStateProperties in your GameState, as all of these moves store state in properties.
                            * RoundRobinNumRounds - Also checks that no more than NumRounds() around have happened
                                * DealCountComponents - Moves a component from GameStack to PlayerSTack() one at a time until each player has been dealt TargetCount() components.
                                    * DealComponentsUntilPlayerCountReached - Instead of a fixed number, done when every player has TargetCount or more components in PlayerStack.
                                        * CollectComponentsUntilPlayerCountReached - Flip so that the components move from PlayerStack to GameStack.
                                            * CollectComponentsUntilPlayerCountLeft - Flips it so the TargetCount is when each PlayerStack has that many items or fewer.
                                                *CollectAllComponents - Overrides TargetCount to 0, colleting all components.
                                        * CollectComponentsUntilPlayerCountLeft - Flip movement to be from PlayerStack to GameStack, and flip TargetCount to be when all PlayerStack have TargetCount or less.
                                    * DealComponentsUntilGameCountLeft - Instead of a fixed number, done when GameStack's count is TargetCount or less.
                                        * DealAllComponents - Overrides TargetCount to be 0
                                        * CollectComponentsUntilGameCountLeft - Flips move so it's from PlayerStack to GameStack
                                    * CollectCountComponents - Flips so components move from PlayerStacks to GameStack

Move Deal and Collect Component Moves

Generally when moving components from one place to another it makes sense to
move one component at a time, so that each component is animated separately.
However, this is a pain to implement, because it requires implementing a move
that knows how many times to apply itself in a row, which is fincky and error-
prone.

There is a collection of 9 moves that all do basically the same thing for moving
components, one at a time, from stack to stack. Move-type moves move components
between two specific stacks, often both on your GameState. Deal and Collect type
moves move components between a stack in GameState and a stack in each Player's
PlayerState. Deal-type moves move components from the game stack to the player
stack, and Collect-type moves move components from each player to the GameState.

All of these moves define a way to define the source and destination stack. For
Move-type moves, you define SourceStack() and DestinationStack(). For Deal and
Collect-type moves, you implement GameStack() and PlayerStack().

All moves in this collection implement TargetCount() int, and all of them
default to 1. Override this if you want a different number of components checked
for in the end condition.

In practice you'll often use WithTargetCount, WithGameProperty, and friends as
configuration to AutoConfigurer.Config instead of overriding those yourself. In
fact, in many cases configuartion options are powerful enough to allow you to
use these moves types on their own directly in your game. See the documentation
in the sections above for more examples.

Each of Move, Deal, and Collect have three variants based on the end condition.
Note that Move-type moves have only two stacks, but Deal and Collect type moves
operate on n pairs of stacks, where n is the number of players in the game. In
general for Deal and Collect type moves, the condition is met when all pairs of
stacks meet the end condition.

{Move,Deal,Collect}CountComponents simply apply that many moves without regard
to the number of components in the source or destination stacks. Move names that
end in CountReached operate until the destination stacks all have TargetCount or
more items. Move names that end in CountLeft operate until the source stacks all
have TargetCount or fewer items in them.

Since a common configuration of these moves is to use
{Move,Deal,Collect}ComponentsUntil*Reached with a TargetCount of 0, each also
provides a *AllComponents as sugar.

Groups

Groups allow you to specify a specific set of moves that must occur in a given
order. You pass them to AddOrderedForPhase. All of the groups are of type
MoveProgressionGroup, and this package defines 5: Serial, Parallel,
ParallelCount, Repeat, and Optional. They can be nested as often as you'd like
to express the semantics of your move progression.

They are defined as functions that return anonymous underlying structs so that
when used in configuration you can avoid needing to wrap your children list with
[]MoveProgressionGroup, saving you typing.

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
move at multiple points in a progression. WithMoveNameSuffix is useful for that
case.
*/
package moves
