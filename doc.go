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
objects will primarily be composed of bools, ints, strings, Timers (see Timers
section below), and Stacks (see the Components section, below).

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
collection of factories that produce Moves that may be used. The GameManager
maintains a set of all of the different types of move factories that may be
used in this game type.

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

Moves have a number of required methods, and most of them will be no-ops in
many cases. DefaultMove is an optional convenience struct that is designed to
be embedded in your own Moves that implements a bit of the boilerplate
automatically. Moves also should generally use the autoreader codegen tool to
generate their reader methods.

You should make your moves granular enough that any semantically-relevant in-
betweeen state happens between moves, because a move is a bit of a black box
itself, because players can only see the result after the move was made. Think
about it as yielding to the event loop in a UI-driven application.

This means that in some games you'll have LOTS of granular fixUp moves. A
common pattern is to have a chain of FixUp moves that apply one after another
without fail, and are primarily broken into separate moves just to be granular
enough semantically. This creates a lot of cruft--a lot of extra FixUp moves
on manager, and also generally requires some awkward and error-prone signaling
in GameState about SubPhases, and writing finicky Legal() methods for the
later FixUp moves in the chain that only trigger in the precise right
condition.

As an advanced optimization, your move can define an ImmediateFixUp() that
takes a state and returns a Move. After a move is a applied, if the Move has
an ImmediateFixUp, it will be immediately applied BEFORE delegate.ProposeFixUp
is called. Importantly, the moves returned from this method do not need to be
registered on GameManager, because somewhere in their ancestor chain must have
been registered in order to have successfully been Proposed in the first
place. This allows games with many long fix-up chains to be a bit cleaner and
not have to have error-prone Legal logic signaling.

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
precisely one Stack at every State in your game. The design of Stack methods
and Delegate's DistributeComponentsToStartingStacks is carefully configured so
that it is impossible to not satisfy this invariant (as long as you do not
call stack.UnsafeInsertNextComponent outside of testing context).

PlayerIndex

A common task in States and Moves is to keep track of an index of a Player.
PlayerIndex is a special PropertyType that helps verify that these stay in
legal bounds, and be explicit about where such legal indexes are expected.

After a move's Apply() method is called, if any of the PlayerIndexes in the
resulting State are invalid, the move will fail to be applied.

There are two special PlayerIndex values, AdminPlayerIndex, and
ObserverPlayerIndex. These sentinels communicate when a special PlayerIndex
semantic is used in a given context.

PlayerIndex.Next() and .Previous() are convenient ways to increment and
decrement without going out of bounds.

Timers

One type of property that may be set on your States is a Timer. A timer is
used to represent when the passage of time has semantic meaning in the rules
of the game. For example, in Memory, once both cards are revealed, the cards
need to be hidden within 3 seconds, and if they aren't, they should be hidden
automatically. As another example, at the beginning of Galaxy Trucker the
first round of play proceeds for a specific number of minutes, with all
players moving simultaneously until the timer is up. Timers are in contrast to
time-based things that are purely presentational and non- semantic, like an
animation of a card moving from one stack to another in the client.

Timers function by queuing up a move to be automatically proposed (via
proposeMove) after a certain amount of time has elapsed. Activate a timer by
calling Start() on it and passing the amount of time to elapse and the move to
propose. Timers must exist as pre-defined properties in one of your State
objects (GameState, PlayerState, or DynamicComponentValues). The amount of
time for any given timer may be changed dynamically when you Start() it.

When a timer is active, you can inspect its TimeLeft property to see how much
time it has until it triggers. This can be useful to, for example, render a
progress bar in the client tha shows how much time is left. Note that although
you generally call timer.Start in the Apply() method of a Move, the timer does
not actually start counting down until the move is fully applied and saved in
the database.

A timer can be Cancel'd before it has been triggered. Canceling an already-
fired or unstarted timer is a safe no-op. If you call Start() on a timer that
is already Active(), the previously-active Timer will first be canceled.

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
For basic types (e.g. int, string, bool), these are the only two policies. For
those property types, any Policy other than PolicyVisible or PolicyRandom
behaves like PolicyHidden.

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

Sanitization Policies by default control whether the value and identity of a
given component can be known at any given time. However, in many cases the
identity of a given component can be tracked, even when its value is not
known. For example, a player grabs the top card from the draw deck and places
it face down in front of himself. A couple of turns later he reveals that the
card is a blue card. That card is then flipped back face down. Later, it is
transferred to the hand of another player.

In that example, the specific value of the card is only known at a single
point in time. However, it should be possible to keep track of the fact that
it is the same card through that entire series of steps. An astute observer
should be able to deduce that it is the blue card all along--which might allow
the player to deduce other relevant information, like that an earlier secret
play must have been a non-blue card.

In addition, in general, keeping track of the identity of cards, in order to
animate them from moving from one stack to another, or show that a given card
is the same card when it is flipped over and its value is revealed, requires
being able to keep track of the card's identity.

In order to do this, every component has a semi-stable Id. This Id is
calculated based on a hash of the component, deck, deckIndex, gameId, and also
a secret salt for the game. This way, the same component in different games
will have different Ids, and if you have never observed the value of the
component in a given game, it is impossible to guess it. However, it is
possible to keep track of the component as it moves between different stacks
within a game.

Every stack has an ordered list of Ids representing the Id for each component.
Components can also be queried for their Id.

Stacks also have an unordered set of IdsLastSeen, which tracks the last time
the Id was affirmitively seen in a stack. The basic time this happens is when
a component is first inserted into a stack. (See below for additional times
when this hapepns)

Different Sanitization Policies will do different things to Ids and
IdsLastSeen, according to the following table:

	| Policy         | Values Behavior                                                  | Ids()        | IdsLastSeen() | Notes                                                                                                 |
	|----------------|------------------------------------------------------------------|--------------|---------------|-------------------------------------------------------------------------------------------------------|
	| PolicyVisible  | All values visible                                               | Present      | Present       | Visible is effectively no transformation                                                              |
	| PolicyOrder    | All values replaced by generic component                         | Present      | Present       | PolicyOrder is similar to PolicyLen, but the order of components is observable                        |
	| PolicyLen      | All values replaced by generic component                         | Random Order | Present       | PolicyLen makes it so it's only possible to see the length of a stack, not its order.                 |
	| PolicyNonEmpty | Values will be either 0 components or a single generic component | Absent       | Present       | PolicyNonEmpty makes it so it's only possible to tell if a stack had 0 items in it or more than zero. |
	| PolicyHidden   | Values are completely empty                                      | Absent       | Absent        | PolicyHidden is the most restrictive; stacks look entirely empty.                                     |


However, in some cases it is not possible to keep track of the precise order
of components, even with perfect observation. The canonical example is when a
stack is shuffled. Another example would be when a card is inserted at an
unknown location in a deck.

For this reason, a component's Id is only semi-stable. When one of these
secret moves has occurred, the Ids is randomized. However, in order to be able
to keep track of where the component is, the component is "seen" in
IdsLastSeen immediately before having its Id scrambled, and immediately after.
This procedure is referred to as "scrambling" the Ids.

stack.Shuffle() automatically scrambles the ids of all items in the stack.
SecretMoveComponent, which is similar to the normal MoveComponent, moves the
component to the target stack and then scrambles the Ids of ALL components in
that stack as described above. This is because if only the new item's id
changed, it would be trivial to observe that the new Id is equivalent to the
old Id.

Computed Properties

In some cases there are properties on your State whose value is
deterministically defined by other properties in your state. Storing the
derived value as well as the underlying values is redundant and has the
possibility for the State to accidentally get out of sync. For example, in a
game of Blackjack you might have a method for each player's hand
that returns the effective value of their hand--which is a tricky calculation
because Aces can be worth either 1 or 11.

ComputedProperties are the way to model these. They are formally defined, and
are part of the default JSON serialization of your state so are available
automatically to your client-side renderer.

There are two types of ComputedProperties: those that operate on the whole
state ("Global"), and output a top-level value, and those that operate on a
single player state at a time and output a value specific to that player state
("Player"). An example of the former would be "PlayerWithLongestTrain" in
Ticket to Ride. An example of the latter would be the effective blackjack hand
value situation described immediately above. The former is called a
ComputedProperty and the latter is a ComputedPlayerProperty.

Both types of computed properties exhaustively identify the other state
properties that they rely on. This helps the engine only recalculate state
properties when the underlying values have changed, and other optimizations.
When the value is computed, a Sanitized() State is passed instead of a normal
state. This sanitized state object only has the properties defined that were
explicitly listed as dependencies, with other properties obscured with
PolicyRandom. This allows us to verify that your dependency list is a superset
of the properties you actually rely on. This means that a common pattern is to
define your computed properties as methods on your concrete GameStates and
PlayerState objects. Then, your computed properties simply cast the input to
the concrete underlying type you know it is and return the value of calling
that method on it. As long as your dependencies are comprehensive that the
method actually relies on, the result should be the same. This way, you can
define a method you use directly on the server, but also have the same
property available computed on the client.

ComputedProperties are defined based on the config object your GameDelegate
returns from the ComputedPropertiesConfig method. This method should always
return the same value. These objects are normally heavy and large (they often
include method definitions), so it is recommended to define a single Config at
init() time and always pass a reference to that config.

ComputedPlayerProperties come in two flavors: Compute and GlobalCompute,
depending on which of those properties is non-nil. Compute takes as its only
argument just the specific PlayerState corresponding to the player whose value
is being computed. But in some cases that might not be enough information; for
example, sometimes there are properties per player that depend on the location
of other players, too. In those cases, you shoul define a GlobalCompute
instead. It is passed the whole state object and the target PlayerIndex. In
that case it is up to you to return only the value associated with the target
PlayerIndex, but you have all other values available to you.

For convenience you can have your delegate return objects that implement
ComputedPropertyCollections to an underlying concrete struct, that use
DefaultReadSetter(). This allows you to only have to cast once to get a type-
checked property accessor. If those delegate methods return nil, a generic,
flexible PropertyReadSetter will be used instead. When you call
manager.SetUp(), the framework will verify that the concrete collections you
return (if you do) are shaped to accept the properties that are configged.

Importantly, your Compute() methods for your computed properties should be
resilient to states that have sanitized. In cases where we are generating a
SanitizedState, your compute method will be called and pass in the sanitized
values. In many cases the naive approach will do what you want, but in cases
of Stacks with things like PolicyLen it can get trickier.

Your Compute() methods should also be deterministic based on their inputs.
This is important so that the system can be conservative about when to
calculate them.

Dynamic Component Values

Components are always read-only and have the same values across every game.
However, there are some cases where a given component might have a specific
dynamic value that can change over the life of a specific game. A canonical
example is a die: the faces that it has are fixed across all games, but which
face is "selected" can change with each roll--and in some cases its value may
even be hidden. A more complex example is in Evolution: Climate, where a given
player may have 1 to n Species cards, each with a body size, population size,
stack of fed food, and up to four slots for trait cards.

These types of situations can be modeled with Dynamic Component Values. Every
component in a given deck can have dynamic properties of the same shape. These
properties are stored in the "Components" section of the JSON output of a
given GameState. You can access them server-side by getting a reference to the
associated component and calling DynamicValues. Client-side, the dynamic
values are automatically exapanded in the expandedState.

By default a deck does not have any dynamic component state for its values. To
override this, your GameDelegate should return a non-nil value when
EmptyDynamicComponentValues is called for the specified deck. This method
should return the same shape of underlying object every time it is called for
the same deck.

The visibility of Dynamic Component Values in a Sanitized state is somewhat
more complex than normal state. The dynamic component values associated with
any one component will only be visible if the component is in a Stack whose
effective policy is PolicyVisible--if its containing stack is anything else,
then every property on that dynamic component value will be set to
PolicyHidden (or PolicyRandom when Computed Property dependencies are being
calculated). Then, the policy for each property on the DynamicComponentValue,
as configured on Policy.DynamicComponentValues, is applied to each property to
achieve the final dynamic component values visbility.

Note that this visibility of components is transitive: it is possible for the
Dynamic Component Values of a given component to have its own Stack containing
components that themselves have dynamic component values. In that case, the
visibility of the "parent" component is expanded to apply to the "children"
components. That is, if the parent component is visible, and the stack
property of the component is also visible, the child component will also be
considered visible.

Agents

Agents are artificial intelligent agents who play as a particular player in a
given game.

Managers must have agents congfigured on them when they are created, and at
the Game creation it is specified in SetUp which players (if any) should
actually be represented as agents.

Agents have the option to have extra state (over and beyond what is
represented in the Game's state). For example, in a game of deduction, an
agent might keep track of what cards it thinks the other player has in their
hand. The agent is responsible for serializing and deserializing this state to
a slice of bytes. The Game engine will store and retrieve these state blobs.
Each time it is called the agent has the opportunity to return a new state; if
it does, that state blob will be stored. Each time an Agent is called, it is
passed the most recent state it had saved.

After each time a Player has made a move (and any resulting FixUp moves have
been applied), agents are woken up and given a chance to propose a move. If
they return a move, it is Proposed, via ProposeMove, to the game.

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

Note that when we make a copy of your PlayerState or GameStates, we only
enumerate and copy allowed property types that are visible via the
PropertyReader interface. So do not rely on other hidden properties, because
they will not be copied over. (This should go without saying as when those are
marshaled to JSON they are not included anyway).

The cmd autoreader makes it easy to automatically generate Reader and
ReadSetter implementations with just a few comment lines and go:generate. See
that packages's documentation for how to use it.

*/
package boardgame
