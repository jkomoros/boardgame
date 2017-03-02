/*

boardgame is a package that makes it possible to build boardgames with minimial fuss.

boardgame/server implements a progressive web app based on your game with just
a few lines of configuration.

Games

Games are the funamdental object. They represent a specific game with a
certain number of players and a versioned history of States. Once created,
games can only be modified by Moves.

Each game is associated with one GameManager. The GameManager manages state
that is shared across multiple games (like components, moves, delegates, etc)
and interacts with the storage layer.

The majority of a game's state is stored as a State--a JSON-able object that
represents the entirety of the semantic state for the game, in a manner
particular to this type of game. Your State's GameState and PlayerState
objects will primarily be composed of bools, ints, and Stacks (see the
Components section, below).

Each game has a Version() that monotonically increases as Moves are
successfully applied to the game to modify it. Each version has precisely one
State associated with it. Once a state is created, it can never be modified.

Games have a Modifiable property. When a game is first created it is
Modifiable, but most other games are not modifiable--when they are created
they are set to a snapshot of what the storage layer has for that game. This
reflects that multiple Game objects might represent the same notional game as
far as the storage layer is concerned. As long as you use the same manager,
only one Modifiable version of a given notional Game will ever be in
existence.

NewGames are empty, and must be SetUp() before moves can be applied to them.
Only Modifiable games may actually have a move applied to them. More in the
next section.

Moves

Moves are the only way to modify a game's state. A given type of game has a
collection of Moves that may be used. The GameManager maintains a set of all
of the different types of moves that may be used in this game type.

Moves can be serialized as JSON and contain a set of properties that together
define all of the information necessary to fully describe the Move.

Moves have a Legal method that is given a State object and returns an error if
the Move is not legal to make at the given game state.

Moves have an Apply method that is given a new state object to modify in
accordance with the game's semantics.

Games have a ProposeMove method that takes a Move object and queues it up to
be Applied to the Game. If the given Game is not modifiable, the move will be
dispatched, via the GameManager, to a Game object for this notional Game that
is.

Moves that are proposed are considered in order. Their Legal method is called
with the CurrentState of the Game (it may have changed if there were other
Moves ahead of this move in the queue). If the move is still legal, it will be
applied. The game's version number will increment, and a new State will be
addded to the history that reflects the output of that Move's Apply method.

There are two types of Moves: PlayerMoves, and FixUpMoves. PlayerMoves are
moves that can be proposed by Players. FixUp moves are moves that are never
legal to be applied by players. These FixUp moves are "meta" moves that help
keep the game in a consistent, playable state. For example, a typical FixUp
Move is AdvanceCurrentPlayer, which notes that the current player has no more
valid actions left and then advances so it is the next player's turn.

After each Move is Apply'd, the Game's Delegate (see below) is given an
opportunity to examine the new State of the game and decide if a FixUp move
should be applied.

After each move, and when there are no more FixUp moves to apply, the Game
checks to see if the game is now over by asking its Delegate (see below). If
so, the game is marked as Finished, and the winners are noted. At that point
no more moves may be applied.

Game Delegates

Each GameManager has a reference to a GameDelegate that is specific to this
game type. The GameDelegate is where you can configure precise behaviors that
happen at key points in the lifecycle of your particular game type.

For example, Delegates are consulted in the following points:

1) After every Move is applied, to decide if a FixUp move should not be
applied before the next PlayerMove in the queue is applied. For example, you
might have a ShuffleDiscardStackToDrawStack that is Legal whenever the
DrawStack is empty.

2) To initialize the State. When your game is first created, your Delegate
will provide the first concrete State object. Future state objects will be
created by Copying and modifying this state object.

3) Deserializing State objects from storage. Your storage objects will have
been serialized as JSON and must be reinflated into concrete types.

4) CheckGameFinished(), called after every move, checks the game's
CurrentState to see if the game is now finished, and if so, who won.

In some cases, your delegate doesn't need to do much special. For example,
Delegate.ProposeFixUp() is often the same for many games: iterate through each
FixUpMove that has been configured on the manager, and return the first one
that is Legal(). For those reasons, this package defines a DefaultGameDelegate
that is designed to be anonymously embedded into your own struct, so you only
need to modify the behavior of the methods whose behavior is actually special
to your game.

Components

Your game has a set of Components, which is every object that could be moved
around in the game. In practice, it includes dice, cards, meeples, resource
tokens, and anything else that a real-world board game would enumerate in its
Components section for players trying to verify that they still had all of the
necessary pieces. Components have a set of immutable properties. Different
types of components in your game might have different types of properties.

There is one global set of Components that are used in each type of game. This
set is called a Component Chest. After it is created, its shape is frozen and
associated with your GameManager.

The Chest consists of 0 to n Decks. A Deck is a collection of components, all
of the same basic type. For example, in Ticket To Ride your Chest might have a
Deck of Contract Cards, and a Deck of train cards. The terminology for Decks
makes the most sense for cards, but it applies for any components. For
example, if your game included multiple dice, you might have them in a Deck
called "Dice".

Your State object will contain a collection of Stacks. Stacks are mutable
ordered collections of Components of a specific type. For example, you might
have a Stack for the Draw pile, a Stack for Discard pile, and a stack
representing each player's hand. Every component in the Chest must live in
precisely one Stack at every State in your game. During Game set-up, your
delegate's DistributeComponentToStarterStack will be called for each component
in the chest in turn, which helps you conform to this important invariant from
the very beginning, and then make sure to maintain it in each Move's Apply
method.

Sanitization

The server canonically knows all state in a game. However, there are certain
bits of state that should not be known by specific players. For example, in
poker,the other players should not know the two hidden cards in your hand.

boardgame handles this with a notion of sanitization. When preparing a state
object to be sent to a client, it is possible to get a sanitized version of
the state with GameManager.SanitizedStateForPlayer(index). This will sanitize
certain fields according to a policy that your Delegate defines in
StateSanitizationPolicy. The result is a copy of the input state, with the
various fields obscured, and which will have Sanitized() return true. All of
the fields will always have the same "shape" as before (e.g. GrowableStacks
will not be reduced to an int), but will have key properties changed so that
less information can be recovered.

The policy for a game will never change during the course of the game; it is
tied to which player the state is being prepared for, which key we are
considering, and which groups the various players are in. The same policy will
be applied to each PlayerState in the State; use Groups to change the behavior.

boardgame has no notion of who is who; it will generate a SanitizedState for
whomever you request. Other packages, like Server, keep track of which person
is which via mechanisms like cookies.

There are a number of policies that can be applied to each key, of type
Policy. PolicyVisible is the default; if there is no effective policy in
place, it defaults to PolicyVisible. It leaves the property unchanged.
PolicyHidden is the most restrictive; it sets the property to its zero value.
For basic types (e.g. int, string, bool), these are the only two policies. Any
Policy other than PolicyVisible behaves like PolicyHidden.

Groups (e.g. SizedStacks and GrowableStacks) have a few extra policies.
PolicyLen will obscure the group so that the number of items is clear, but all
elements will be replaced by the Deck's GenericComponent. PolicyNonEmpty is
similar to PolicyLen, but if the real Stack has 1 or more components, the
output result will have a single GenericComponent. This allows you to observe
whether the stack was empty or not, but not anything about how many components
it had. PolicyOrder replaces each Component with a stable but obscured
ShadowComponent, so that observes can keep track of the lenght, and when
components swithc orders in the stack, but not what the underlying components
are.

To compute the effective policy for a given property, we have to consider the
Groups. Conceptually there are a number of groups, which define which players
are in or out of each one. In the future there will be a way to define group
membership that can be modified just like any other part of the state. At this
point there are three special groups.  Every player is a member of GroupAll.
GroupSelf is the group that only the player who the state is being prepared
for is in. GroupOther contains all players who the state is not being prepared
for.

Policies contain GroupPolicies for each key in Game and State. GroupPolicies
are a map of Group ID to the effective policy. When preparing a sanitized
state for a given property, we to through each group/policy pair in the
GroupPolicy. We collect each policy where the player that the state is being
prepared for is in. Then the effective policy is the *least* restrictive
policy that applies. In practice this means that policies like
GroupAll:PolicyLen, GroupSelf:PolicyVisible make sense to do.

Implementing Your Own Game

When you are implementing your own game, at a high level you must do the
following things:

1) Define a GameState/PlayerState implementation that fully captures all of
the semantic state of the game at all times. In practice this will likely
include state that is central to the game, as well as state specific to each
user. It often includes more things than you might first think. For example,
your state should include how many of each type of action the current player
can still do in their turn, so that your game can decide when to advance to
the next player. Ensure that your GameState and PlayerState objects serialize
with all necessary state when marshaled as JSON, and that they survive a
round-trip through GameManager.StateFromBlob. All of the properties you want
serialized or available through PropReader should be public fields on the
struct. For convenience, it's good to implement a concreteStates(state) that
returns the concrete *gameState and []*playerState. That way in the top of
your specific methods that accept a State, you can quickly get workable, type-
checked structs with only a single conversion leap of faith at the top.

2) Define the complete set of Components that exist in your game. Every item
that could be manipulated or moved, including cards, meeples, resource tokens,
dice, and much more, should be enumerated in your Component Chest.

3) Define a GameDelegate that overrides various game level logic at key points
in a Game's lifecycle. For example, the delegate is consulted to provide a
starting state for a new game, to decide if a game is now finished, whether
any fixup moves should be applied, and much more. A substantial portion of the
logical "meat" of your implementation will be here.

4) Define a set of Moves that fully define all of the possible modifcations
that could ever occur in your game, both for Players and FixUp moves. Each
Move needs a Legal() and Apply() method, and should be fully serialized by
json.Marshal(). This is where the majority of the logical "meat" of your game
definition will live.

5) Often the end result of your game will be a Progressive Web App. You'll
need to do a few more things, as described in the boardgame/server package, to
complete the web app.

Reflection and Properties

The aim of the boardgame package is to make it as easy as possible for you to
implement your own boardgames, focusing only on the central semantic logic of
the game. The package tries to hit a sweet spot between concrete types whose
behavior is modified by delegates, and interfaces that you must implement.

States, Moves, and Components, in particular, will have a set of properties
that is very specific to your particular game and taht particular object type.
The package itself tries to rely on reflection only rarely, and only when
instructed to. In practice, inside of your Move.Apply, Move.Legal, and
Delegate methods, you will often immediately cast the provided generic Move,
State, or Component to the underlying type you know it is.

Every so often the package has to interact with objects whose shape it does
not know. It relies on the PropertyReader and PropertyReadSetter interfaces to
do this manipulation. Implementing these methods can be a pain, which is why
this package provides a set of implementation methods that rely on reflection
to satisfy these interfaces.

*/
package boardgame
