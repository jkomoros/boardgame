# boardgame

boardgame is a work-in-progress package that aspires to make it easy to define multi-player boardgames that can be easily hosted in a high-quality web app with minimal configuration. It is under active development as a hobby project and different components of it vary in their completeness and polish.

The core of your game logic is constructed using the core library into a *game manager* for each game. The server package makes it easy to take those game managers and install them into a server instance. Each game manager defines a basic view that knows how to render any given state of one of its game for the user.

A number of example games are defined in the examples sub-package to demonstrate how to use many of the key concepts. Real documentation for the core game engine is in the [godoc package docs](https://godoc.org/github.com/jkomoros/boardgame).

## Tutorial

*This tutorial will walk through some concrete examples of how to configure a server and create games, in a way that narratively makes sense but leaves a number of topics unexplored or lightly developed. For more in-depth documentation of the core concepts, check out the core library's package doc, and for more about the server, see `server/README.md`*

## boardgame-util

`boardgame-util` is a multi-purpose tool provided by this package that does a number of things, from automatically generating code, to adminstering a mysql databse, to building and running a server based on a configuration file.

The rest of this tutorial will assume you have it installed. Sitting in the `boardgame` folder, run `go install ./...` to install `boardgame-util` as well as registering all of the sub-packages on the system.

Now ensure it's installed by running:

```sh
boardgame-util
```

You should see a help message describing what boardgame-util can do. The help of the boardgame-util command is very comprehensive, and it's the best way to learn about what it can do. The `help` subcommand to learn more about any given comand or sub-command:

```sh
boardgame-util help
```

The `boardgame-util` command looks for configuration in json files when it runs (looking up the directory hiearchy until it finds one). The boardgame package provides a `config.SAMPLE.json` in its root, which defines reasonable defaults. For all of the commands in this tutorial, it assumes that the current working directory is `$GOPATH/src/github.com/jkomoros/boardgame` or one of its sub-directories.

## Quickstart servers

*Note: this tutorial will walk you through real examples to introduce you to the concepts. If you want to start creating your own game based on a quick-start example you can modify, feel free to skip ahead to the "Creating your own game" section. *

The quickest way to get a server running is via the `boardgame-util serve` command. 

The command requires you have `npm` installed. You can install npm by following the instructions to install node: https://nodejs.org/en/

That's the only JavaScript prerequisite — all other dependencies (including the Vite dev server) are installed automatically via `npm install`.

Now you have the prerequisites installed and can use the `boardgame-util serve` command.

Sitting in the boardgame package, run:

```sh
boardgame-util serve
```

Now you can visit the web app in your browser by navigating to `localhost:8080`

This command automatically uses the default configuration file in `boardgame/config.SAMPLE.json` to identify which games packages to include, then creates a temporary binary that imports and instantiates them, as well as pulling together the necessary static files to serve as well. When you kill the command with `Ctrl-C` those temporary files are deleted.

## Game Managers

Now that you have the server set up, let's dig into how a given game is constructed.

We'll dig into `examples/memory` because it covers many of the core concepts. The memory game is the classic childhood game where there's a deck of cards of symbols, with exactly two cards for each symbol. The cards are arrayed face down on the table and players take turn flipping over two cards. If they get a match, they get to keep the cards.

At the core of every game is the `GameManager`. This is an object that encapsulates all of the logic about a game and can be installed into a server. The `GameManager` is a struct provided by the core package that handles much of the operation of the game engine, but it's a shell that doesn't do much on its own. A `GameDelegate`, which you write for your game and provide when you create a new GameManager, encapsulates the core of the logic central to your game, including its name, what moves can be made, how many players can play, when the game is finished, and much more.

Each game type, fundamentally, is about representing all of the semantics of a given game state in a versioned **State** and then configuring when and how modifications may be made by defining **Moves**.

### State

The state is the complete encapsulation of all semantically relevant information for your game at any point. Every time a move is succesfully applied, a new state is created, with a version number one greater than the previous current state. States may only be modified by applying moves.

Game states are represented by a handful of structs specific to your game type. All of these structs are composed only of certain types of simple properties, which are enumerated in `boardgame.PropertyType`. The two most common structs for your game are `GameState` and `PlayerState`.

`GameState` represents all of the state of the game that is not specific to any player. For example, this is where you might capture who the current player is, and the Draw and Discard decks for a game of cards.

`PlayerState`s represent the state specific to each individual player in the game. For example, this is where each player's current score would be encoded, and also which cards they have in their hand.

Let's dig into concrete examples in memory, in `examples/memory/state.go`.

The core of the states are represented here:

```go
//boardgame:codegen
type gameState struct {
	base.SubState
	CardSet        string
	NumCards       int
	CurrentPlayer  boardgame.PlayerIndex
	HiddenCards    boardgame.SizedStack  `sizedstack:"cards,40" sanitize:"order"`
	VisibleCards   boardgame.SizedStack  `sizedstack:"cards,40"`
	Cards          boardgame.MergedStack `overlap:"VisibleCards,HiddenCards"`
	HideCardsTimer boardgame.Timer
	//Where cards not in use reside most of the time
	UnusedCards boardgame.Stack `stack:"cards"`
}

//boardgame:codegen
type playerState struct {
	base.SubState
	playerIndex       boardgame.PlayerIndex
	CardsLeftToReveal int
	WonCards          boardgame.Stack `stack:"cards"`
}
```

There's a lot going on here, so we'll unpack it piece by piece.

At the core you can see that these objects are simple structs with (mostly) public properties. The game engine will marshal your objects to JSON and back often, so it's important that the properties be public.

It's not explicitly listed, but the only (public) properties on these objects are ones that are
legal according to `boardgame.PropertyType`. Your GameManager would fail to be created if your state structs included illegal property types.

Note the first anonymous field of `base.SubState`. This is a simple struct designed to be anonymously embedded in the substates you define that implements the SetState method that SubStates must define. It's technically optional, but you'll normally just want to anonymously embed it in your gameState and playerStates.

Most of the properties are straightforward. Each player has how many cards they are still allowed to draw this turn, for example.

#### Stacks and Components

As you can see, stacks of cards are represented by type `Stack`, `SizedStack`, or `MergedStack`. These are all different related types of a notion called a Stack.

Stacks contain 0 or more **Components**. Components are anything in a game that can move around: cards, meeples, resource tokens, dice, etc. Each game type defines a complete enumeration of all components included in their game in something called a **ComponentChest**. We'll get back to that later in the tutorial.

By default Stacks can grow to accomodate new components and have no empty spaces in the middle. Adding a new component to a slot in the middle of a stack would simply push components from there onward down a slot, and grow the stack by one.

A SizedStack is a special kind of Stack that has a fixed number of slots, each of which may be empty or contain a single component. The default growable Stacks are useful in most instances, including representing a player's Hand or a Draw or Discard deck. SizedStacks are useful when there's a specific fixed size or where there might be gaps between components. A SizedStack can be used anywhere a normal Stack can.

Each component is organized into exactly one **Deck**. A deck is a collection of components all of the same type. For example, you might have a deck of playing cards, a deck of meeples, and a deck of dice in a game. (The terminology makes most sense for cards, but applies to any group of components in a game.) The ComponentChest is simply an enumeration of all of the Decks for this game type. Memory has only has a single deck of cards, but other games will have significantly more decks.

Each Stack is associated with exactly one deck, and only components that are members of that deck may be inserted into that stack. The deck is the complete enumeration of all components in a given set within the game. In memory you can see struct tags that associate a given stack with a given deck. We'll get into how that works later in the tutorial.

**Each component must be in precisely one stack in every state**. This reflects the notion that components are phsyical objects that are in only one location at any given time, and must exist *somewhere*. Later we will see how the methods available on stacks to move around components help enforce that invariant.

When a memory game starts, most of the cards will be in GameState.HiddenCards. Players can also have cards in a stack in their hand when they win them, in WonCards. You'll note that there are actually three stacks for cards in GameState: HiddenCards, VisibleCards, and Cards. We'll get into why that is later.

#### Stack Constraints

Stacks support **constraints**: functions that are automatically checked before a component is moved into a stack. If a constraint returns an error, the move is rejected and the component remains in its source. Constraints are also checked during Legal() for moves that use WithSourceProperty/WithDestinationProperty, giving early feedback before Apply() is even called.

Constraints are useful for expressing invariants like "this hand can hold at most 5 cards" or "all cards in this pile must be the same suit."

You can add constraints programmatically:

```go
func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
    gs := state.GameState().(*gameState)
    gs.Hand.AddConstraint(constraints.MaxNumComponents(5))
    return nil
}
```

Or via struct tags (the default `base.GameDelegate` already includes all pre-built constraint constructors):

```go
type gameState struct {
    base.SubState
    Hand boardgame.Stack `stack:"cards,5,max(5)"`
}
```

The `constraints` sub-package provides pre-built constraints: `MaxNumComponents`, `Unique`, `Same`, and `MaxDistinctValues`. See the `constraints` package documentation for details on property path syntax and available constraints.

Constraints are **not** checked during initial game setup (when components are distributed via `DistributeComponentToStarterStack`), only during normal gameplay moves.

#### Pre-Validating Moves with MayMoveTo

When writing custom `Legal()` methods, you often need to check whether a component can be moved to a destination stack *before* actually doing it in `Apply()`. The framework provides `MayMoveTo` and friends for this purpose:

```go
func (m *movePlaceToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
        return err
    }

    game, players := concreteStates(state)
    p := players[m.TargetPlayerIndex.EnsureValid(state)]

    first := p.UnusedTokens.ImmutableFirst()
    if first == nil {
        return errors.New("no tokens left to place")
    }

    // MayMoveToSlot checks: same deck, slots remaining, constraints,
    // and that the specific slot is valid and unoccupied.
    return first.MayMoveToSlot(game.Slots, m.Slot)
}
```

This single `MayMoveToSlot` call replaces what would otherwise be several manual checks (bounds checking, slot occupancy, constraint validation, etc.). The key methods are:

- **`component.MayMoveTo(dest)`** — slot-independent check; validates deck match, slots remaining, and all constraints. Covers `MoveToNextSlot`, `MoveToFirstSlot`, `MoveToLastSlot`, and `SecretMoveTo`.
- **`component.MayMoveToSlot(dest, slotIndex)`** — like `MayMoveTo`, plus validates that the specific slot is in range and (for SizedStacks) unoccupied.
- **`stack.MayMoveAllTo(dest)`** — validates that *all* components in the source stack could be moved to the destination.
- **`stack.MaySwapComponents(i, j)`** — validates that a swap would succeed.

If `MayMoveTo` or `MayMoveToSlot` returns nil in `Legal()`, the corresponding `MoveTo` or `MoveToNextSlot` call in `Apply()` is guaranteed to succeed.

The `moves` package (DealCountComponents, MoveCountComponents, etc.) uses `MayMoveTo` internally, so if you use those moves, constraint checking happens automatically. You only need to call `MayMoveTo` explicitly in custom moves.

#### boardgame-util codegen

Both of the State objects also have a cryptic comment above them: `//boardgame:codegen`. These are actually a critical concept to understand about the core engine.

In a number of cases (including your GameState and PlayerState), your specific game package provides the structs to operate on. The core engine doesn't know their shape. In a number of cases, however, it is necessary to interact with specific fields of that struct, or enumerate how many of a certain type of property there are. It's possible to do that via reflection, but that would be slow. In addition, the engine requires that your structs be simple and only have known types of properties, but if general reflection were used it would be harder to detect that.

The core package has a notion of a `PropertyReader` (as well as `PropertyReadSetter` and `PropertyReadSetConfigurer`), which makes it possible to enumerate, read, and set properties on these types of objects. The signature looks something like this:

```go
type PropertyReader interface {
    //Enumerate all properties it is valid to read and set on this object, and their types.
	Props() map[string]PropertyType
    //Retrieve the IntProp with the given name.
	IntProp(name string) (int, error)

	//... Getters for all of the other PropertyTypes, similar to IntProp

    //An untyped getter for the property with that name
	Prop(name string) (interface{}, error)
}

type PropertyReadSetter interface {
	//All PropertyReadSetters have read interfaces
	PropertyReader

	SetIntProp(name string, value int) error
	
	//Setters for all other non-interface types, similar to IntProp

	//For interface types the setter also wants to give access to the mutable
	//underlying value so it can be mutated in place.
	EnumProp(name string) (enum.Val, error)
	StackProp(name string) (Stack, error)
	TimerProp(name string) (Timer, error)

	PropMutable(name string) bool

	SetProp(name string, value interface{}) error
}

type PropertyReadSetConfigurer interface {
	PropertyReadSetter

	ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error
	ConfigureImmutableStackProp(name string, value ImmutableStack) error
	ConfigureImmutableTimerProp(name string, value ImmutableTimer) error

    ConfigureEnumProp(name string, value enum.Val) error
    ConfigureStackProp(name string, value Stack) error
    ConfigureTimerProp(name string, value Timer) error

	ConfigureProp(name string, value interface{}) error
}
```

This known signature is used a lot within the package for the engine to interact with objects specific to a given game type.

For simple types (like bools, ints, and strings) the signature is
straightforward: a getter and a setter. However, there are three types of
supported properties that are special: `Stack`, `Enum`, and `Timer`. These three types are called "Interface types" because they are a container with some configuration, as well as the specific values within that container. The base interface has "Immutable" prepended and they have read-only methods, and variants without "Immutable" add mutator methods to that base interface. (Note that `Stack` also has variants `SizedStack` and `MergedStack`, and `Enum` also has a `RangedEnum` variant, but as far as the Reader interface is concerned they're all just the base type).

A generic Setter for those properties doesn't make sense in a
`PropertyReadSetter` because the configuration of the property doesn't change,
only its value within the container. For those, the PropertyReader getters are for the "Immutable" variants, and the PropertyReadSetters allow access to the mutable versions, which allows mutation, and also have a ConfigureTYPEProp setters, which are used only after the object is freshly-minted in order to configure the container.

Implementing all of those getters and setters for each custom object type you have is a complete pain. That's why there's a command, suitable for use with `go generate`, that automatically creates PropertyReaders for your structs.

First, install the command by running `go install` from within `$GOPATH/github.com/jkomoros/boardgame/boardgame-util`. You only need to do this once.

Somewhere in the package, include:

```go
//go:generate boardgame-util codegen
```

(In the memory package you'll find it near the top of `examples/memory/main.go`)

And then immediately before every struct you want to have a PropertyReader for, include the magic comment:

```go
//boardgame:codegen
type MyStruct struct {
	//....
}
```

Then, every time you change the shape of one of your objects, run `go generate` on the command line. That will create `auto_reader.go`, with generated getters and setters for all of your objects.

One other thing to note: the actual concrete structs that you define, like `gameState` and `playerState`, should almost always include the mutable variant of an interface type (`Stack`, `SizedStack`, `Enum`, `RangeEnum` and `Timer`; not the versions with "Immutable" prepended); the PropertyReader methods will return just the read-only subset of those objects. In general the whole point of having a state object is to represent the state that *changes* which is why you generally want the mutable variant. However, there are couple of cases where you might want the immutable variant: when you have read-only properties on a component, or when you're using Merged Stacks, which are inherently read-only (more on that later). But for the most part just always use the mutable variants in your state objects.

The game engine generally reasons about States as one concrete object made up of one GameState, and **n** PlayerStates (one for each player). (There are other components of State that we'll get into later.) The `State` object is defined in the core package, and the getters for Game and Player states return things that generically implement the interface, although under the covers they are the concrete type specific to your game type. 

Many of the methods you'll implement will be passed `ImmutableState` objects, because you are only allowed to read the properties, not change them. In the vast majority of cases you are not allowed to modify the State object. To help make the intention clear, you will be passed either an `ImmutableState` or `State` object (the latter embedding the `ImmutableState` interface and adding mutation methods) to make the expectation clear.

Many of the methods you implement will accept an ImmutableState object. Of course, it would be a total pain if you had to interact with all of your objects within your own package that way--to say nothing of losing a lot of type safety.

That's why it's convention for each game package to define the following private method in their package:

```go
func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}
```

Whenever the game engine hands you a state object, this one-liner will hand you back the concrete states specific to your game type:

```go
func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {
	game, players := concreteStates(state)
	//do something with game and players, since they are now the concrete types defined in this package
}
```

... Of course, when you pass the ImmutableState or State object through your concreteStates method you'll just get the naked, modifiable, concrete structs back, and there's nothing to prevent you from changing the properties. Don't do that--at best it won't actually make a change that will be persisted, but at worse it could lead to odd inconsitencies later, if the engine for example re-used the same state object.

#### PlayerIndex

gameState has a property named `CurrentPlayer` of type `boardgame.PlayerIndex`. This property, as you might expect, encodes whose turn it currently is.

It would be reasonable to encode that bit of state as a simple `int` (and indeed, that's basically what a `PlayerIndex` property is). However, it's so common to have to encode a `PlayerIndex` (for example, if there's a move to attack another player), and there are enough convenience methods that apply, that the core engine defines the type as a fundamental type.

`PlayerIndex`es make it easy to increment the `PlayerIndex` to the next player (wrapping around at the end). The engine also won't let you save a State with a `PlayerIndex` that is set to an invalid value.

`PlayerIndex`es have three special values: the `AdminPlayerIndex`, the `ObserverPlayerIndex`, and the `AnyPlayerIndex`. The AdminPlayerIndex encodes the special omniscient, all-powerful player who can do everything. Special moves like FixUp Moves (more on those below) are applied by the AdminPlayerIndex. In dev mode it's also possible to turn on Admin mode in the UI, which allows you to make moves on behalf of any player. The ObserverPlayerIndex encodes a run-of-the-mill observer: someone who can only see public state (more on public and private state later) and is not allowed to make any moves. The AnyPlayerIndex is returned by `CurrentPlayerIndex()` during simultaneous phases where any player may act. It behaves like a wildcard in `Equivalent()` checks (like AdminPlayerIndex), but does not grant omniscient access to hidden state.

#### Timer

The last type of property in the states for Memory is the HideCardsTimer, which is of type `boardgame.Timer`. Timers aren't used in most types of games. After a certain amount of time has passed they automatically propose a move. For Memory the timer is used to ensure that the cards that are revealed are re-hidden within 3 seconds by the player who flipped them--and if not, flip them back over automatically.

Timers are rare because they represent parts of the game logic where the time is semantic to the rules of the game. In memory, for example, if players could leave revealed cards showing indefinitely the game would drag on as players competed to exhaustively commit the location of each card to their memory. Contrast that with animations, where the time that passes is merely presentational, to allow the state changes to be visibly demonstrated to players.

### GameDelegate

OK, so we've defined our state objects. How do we tell the engine to actually use them?

The answer to that, and many other questions, is the `GameDelegate`. The `GameManager` is a concrete type of object in the main engine, with many methods and fields. But there are lots of instances where your game type needs to customize the precise behavior. The answer is to define the logic in your `GameDelegate` object. The GameManager will consult your GameDelegate at key points to see if there is behavior it should do specially.

The most basic methods are about the name of your gametype:

```go
type GameDelegate interface {
	Name() string
	DisplayName() string
	Description() string
	//... many more methods follow
}
```

Those methods are how you configure the name of the type of the game (e.g. 'memory' or 'blackjack', or 'pig'), what the game type should be called when presented to users (e.g. "Memory", "Blackjack", or "Pig"), and a short description of the game (e.g. "A card game where players draw cards trying to get as close to 21 as possible without going over")

The GameDelegate interface is long and complex. In many cases you only need to override a handful out of the tens of methods. That's why the base package provides a `base.GameDelegate` struct that has default stubs for each of the methods a `GameDelegate` must implement. That way you can embed a `base.GameDelegate` in your concrete GameDelegate and only implement the methods where you need special behavior.

Most of the methods on GameDelegate are straightforward, like `LegalNumPlayers(num int) bool` which is consulted when a game is created to ensure that it includes a legal number of players.

GameDelegates are also where you have "Constructors" for your core concrete types:

```go
type GameDelegate interface {
	//...
	GameStateConstructor() ConfigurableSubState
	PlayerStateConstructor(player PlayerIndex) ConfigurableSubState
	//...
}
```

ConfigurableSubState is a simple interface that primarily define how to get a `PropertyReader`, `PropertyReadSetter`, and `PropertyReadSetConfigurer` from the object. Many other sub-state values that we'll encounter later have the same shape, which is why the name is generic.

GameStateConstructor and PlayerStateConstructor should return zero-value objects of your concrete types.

In many cases they can just be a single line or two, as you can see for the PlayerStateConstructor and GameStateConstructor in main.go:

```go
func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurableSubState {

	return new(playerState)
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}
```

This is actually very interesting. As mentioned above, Interface properties (like Stacks, Timers, and Enums) need to have their container initalized to a reasonable starting state. For stacks this includes what deck they should be affiliated with, whether they should be a fixed size, and their starting size. For these interface types the zero-value is effectively missing type information.

One way to do that is to initalize them to a reasonable value in the GameStateConstructor:

```go
func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {

	//This sample shows a way to write this that is NOT what memory
	//actually does.

	cards := g.Manager().Chest().Deck(cardsDeckName)

	if cards == nil {
		return nil
	}

	return &gameState{
		HiddenCards:   cards.NewSizedStack(len(cards.Components())),
		VisibleCards: cards.NewSizedStack(len(cards.Components())),
	}
}
```

But that's not what memory does; it simply returns a pointer to a gameState object with all properties at their zero-value. (And that's lucky, it would be kind of a pain to have to do this for all of your interface types)

The answer is in the struct tags in game and playerStates:

```go
//boardgame:codegen
type gameState struct {
	//...
	HiddenCards    boardgame.SizedStack `sizedstack:"cards,40" sanitize:"order"`
	VisibleCards  boardgame.SizedStack `sizedstack:"cards,40"`
	UnusedCards    boardgame.Stack `stack:"cards"`
	//...
}

//boardgame:codegen
type playerState struct {
	//...
	WonCards          boardgame.Stack `stack:"cards"`
}
```

For stacks, you can provide a struct tag that has the name of the deck it's affiliated with. Then you can return a nil value from your constructor for that property, and the system will automatically instantiate a zero-value stack of that shape. (Even cooler, this uses reflection only a single time, at engine start up, so it's fast in normal usage) It's also possible to include the starting size (for default stacks, the max size, and for sized stacks the number of slots). You can also use constants instead of ints for the size. See the section on Constants at the end of this tutorial.

The vast majority of real-world usecases you'll encounter can just use struct tags.

#### Other GameDelegate methods

The GameDelegate has a number of other important methods to override.

One of them is `CheckGameFinished`, which is run after every Move is applied. In it you should check whether the state of the game denotes a game that is finished, and if it is finished, which players (if any) are winners. This allows you to express situations like draws and ties.

Memory's `CheckGameFinished` could look like this:

```go
func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {

	//This is NOT how memory's CheckGameFinished looks

    game, players := concreteStates(state)

    if game.Cards.NumComponents > 0 {
        return false, nil
    }

    //If we get to here, the game is over. Who won?
    maxScore := 0

    for _, player := range players {
        score := player.WonCards.NumComponents()
        if score > maxScore {
            maxScore = score
        }
    }

    for i, player := range players {
        score := player.WonCards.NumComponents()

        if score >= maxScore {
            winners = append(winners, boardgame.PlayerIndex(i))
        }
    }

    return true, winners

}
```

If there are no cards left in the grid, it figures out which player has the most cards, and denotes them the winner.

However, this pattern--check if the game is finished, and if it is return as a winner any player who has the highest score--is so common that the engine makes it easy to implement with a default behavior built into `base.GameDelegate`. Memory uses it, as you can see in `examples/memory/main.go`:

```go
func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	game, _ := concreteStates(state)

	if game.Cards.NumComponents() > 0 {
		return false
	}

	return true
}

func (g *gameDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
	player := pState.(*playerState)

	return player.WonCards.NumComponents()
}
```

Implementing these two methods is sufficient for base.GameDelegate's default CheckGameFinished to do the right thing.

After `CheckGameFinished` returns true, the game is over and no more moves may be applied.

Another method is `CurrentPlayerIndex`. This method should inspect the provided state and return the `PlayerIndex` corresponding to the current player. If any player may make a move (e.g. during a simultaneous phase), you should return `AnyPlayerIndex`, and if no player may make a move, you should return `ObserverPlayerIndex`. `AnyPlayerIndex` acts as a wildcard in `Equivalent()` checks (allowing any player's move to match), but unlike `AdminPlayerIndex` it does not grant omniscient access to hidden state. `AdminPlayerIndex` should be reserved for engine-initiated actions like fix-up moves and timers. This method is consulted for various convenience methods elsewhere. The reason it can't be done fully automatically is because different games might store this value in a differently-named field, have non obvious rules for when it changes (for example, return the value in this field in the first phase of the game, but a value in another field in the second phase of the game), or not have a notion of current player at all.

The convention is simply to store this value in a property on your gameState called `CurrentPlayer`. If you do that, base.GameDelegate's `CurrentPlayerIndex` will just return that.

There are also four methods that start with `Configure`, which are called to set up which decks to use, which enums, and other state. Those are covered later in the guide.

GameDelegate has a number of other methods that are consulted at various key points and drive certain behaviors. Each is documented to describe what they do. In a number of cases the default implementations in `base.GameDelegate` do complex behaviors that are almost always the correct thing, but can theoretically be overriden if necessary. `SanitizationPolicy` is a great example. We'll get to what it does in just a little bit, but although the method is quite generic, `base.GameDelegate`'s implementation encodes the formal convention of using struct-based tags to configure its behavior.

#### Set Up

Once you have a GameManager, you can create individual games from it by calling `NewGame`, passing the number of players and any other optional configuration. This is where the game's state is initalized and made ready for the first moves to be applied. `NewGame` may fail for any number of reasons. For example, if the provided number of players is not legal according to the `GameDelegate`'s `LegalNumPlayers` method, `NewGame` will fail.

The initalization of the state object is handled in three phases that can be customized by the `GameDelegate`: `BeginSetup`, `DistributeComponentToStarterStack` and `FinishSetup`.

`BeginSetup` is called first. It provides the State, which will be everything's zero-value (as returned from the Constructors, with minimal fixup and sanitization applied by the engine). This is the chance to do any modifications to the state before components are distributed. It is also the idiomatic place to call `StatePropertyRef.Validate()` on any global `StatePropertyRef` variables, which will catch misconfigured string-based property names early.

`DistributeComponentToStarterStack` is called repeatedly, once per Compoonent in the game. This is the opportunity to distribute each component to the stack that it will reside in. After this phase is completed, components can only be moved around by calling `SwapComponents`, `MoveComponent`, or `Shuffle` (or their variants). This is how the invariant that each component must reside in precisely one stack at every state version is maintained. Each time that `DistributeComponentToStarterStack` is called, your method should return a reference to the `Stack` that they should be inserted into. If no stack is returned, or if there isn't room in that stack, then the `NewGame` will return an error. Components in this phase are always put into the next space in the stack from front to back. If you desire a different ordering you will fix it up in `FinishSetup`.

`FinishSetup` is the last configurable phase of setting up a game. This is the phase after all components are distributed to their starter stacks. This is where stacks will traditionally be `Shuffle`d or otherwise have their components put into the correct order.

The game returned from `NewGame` is ready for moves to be applied immediately.

### Moves

Up until this point games have existed as a static snapshot of a given state. Outside of the `SetUp` routines, the only modifications to state must be made by `Move`s. 

The bulk of the logic for your game type will be defined as Move structs and then configured onto your GameManager.

The two most important parts of Moves are the methods `Legal` and `Apply`. When a move is proposed on a game, first its `Legal` method will be called. If it returns an error, the move will be rejected. If it returns `nil`, then `Apply` will be called, which should modify the state according to the semantics and configuration of the move. If `Apply` does not return an error, and if the modified state is legal (for example, if all `PlayerIndex` properties are within legal bounds), then the state will be persisted to the database, the `Version` of the game will be incremented, and the game will be ready for the next move.

Moves are proposed on a game by calling `ProposeMove` and providing the Move, along with which player it is being proposed on behalf of. (The server package keeps track of which user corresponds to which player; more on that later.) The moves are appended to a queue. One at a time the engine will remove the first move in the queue, see if it is Legal for the current state, and if it is will Apply it, as described above.

#### Moves and MoveConfigs

There are two types of objects related to Moves: `MoveConfig` and `Move`s.

A `Move` is a specific instantiation of a particular type of Move. It is a concrete struct that you define and that adheres to the `Move` interface:

```go
type Move interface {
    Legal(state ImmutableState, proposer PlayerIndex) error
    Apply(state State) error
    //... Other minor methods follow
}
```

Your moves also must implement the `PropertyReader` interface. Some moves contain no extra fields, but many will encode things like which player the move operates on, and also things like which slot from a stack the player drew the card from. Moves also implement a method called `DefaultsForState` which is provided a state and sets the properties on the Move to reasonable values. For example, a common pattern is for a move to have a property that encodes which player the move should operate on; this is generally set to the `CurrentPlayerIndex` for the given state via `DefaultsForState`.

A `MoveConfig` is a configuration object used to install moves when you are
setting up your `GameManager`. It is what you return from your delegate's
ConfigureMoves() method. It is a simple struct with a name, a constructor for
the move struct, and a bundle of (optional) custom configuration that will be
available on each move of that type's Info.CustomConfiguration(). In practice,
you almost never create your own `MoveTypeConfig`, but rather use
`moves.AutoConfigurer` to generate them automatically for you. More on that
later, too.

#### Player and FixUp Moves

Conceptually there are two types of Moves: Player Moves, and FixUp moves. Player moves are any moves that are legal for normal players to propose at some point in the game. FixUp moves are special moves that are never legal for players to propose, and are instead useful for fixing up a state to ensure it is valid. For example, a common type of FixUp move examines if the DrawStack is empty, and if so moves all cards from the DiscardStack to the DrawStack and then shuffles it. In practice the only thing that distinguishes FixUp moves is that their `Move.IsFixUp()` returns true.

After each move is succesfully applied via ProposeMove, and before the next move in the queue of moves is considered, the engine checks if any FixUp moves should be applied. It does this by consulting the `ProposeFixUpMove` method on the GameDelegate. If that method returns a move, it will be immediately applied, so long as it is legal. This will continue until `ProposeFixUpMove` returns nil, at which point the next player move in the proposed move queue will be considered.

FixUp moves, as a concept, do not exist as a base concept in the base library, except that moves returned from ProposeFixUpMove implicitly are a FixUp move. In practie, the notion of FixUp move is implemented in base.GameDelegate's ProposeFixUpMove, and base.Move.

Technically it is possible to override the behavior of exactly when to apply certain FixUp moves. Realistically, however, the behavior of `ProposeFixUpMove` on `base.GameDelegate` is almost always sufficient. It simply runs through each move configured on the gametype in order, skipping any for whom `.sFixUp()` returns false, setting its values by calling DefaultsForState, and then checking if it is `Legal`. It returns the first fix up move it finds that is legal. This means that it is **important to make sure that your FixUp moves always have well-constructed `Legal` methods**. If a given FixUp move always returns Legal for some state, then the engine will get in an infinite loop. (Technically the engine will detect that it is in an unreasonable state and will panic.)

#### What should be a move?

One of the most important decisions you make when implementing a game is what actions should be broken up into separate Moves. In general each move should represent the *smallest semantically meaningful and coherent modification on the state*. Operations "within" a move are not "visible" to the engine or to observers. In some cases, this means that operations that should have animations in the webapp won't have them because the operations aren't granular enough.

For example, the memory game is broken into the following moves:
- **RevealCard** (Player Move): If the current player's `CardsLeftToReveal` is 1 or greater, reveal the card at the specified index in `HiddenCards`.
- **HideCards** (Player Move): Once two cards are revealed, this move hides them both. It can be applied manually by players, but is also applied automatically when the HideCardsTimer fires.
- **FinishTurn** (FixUp Move): If the current player has done all of their actions and no cards are visible, advances to the next player, and sets the `CardsLeftToReveal` property of the newly selected player to 2.
- **CaptureCards** (FixUp Move): If two cards are visible and they are the same card, move them to the current player's `WonCards` stack.
- **StartHideCardsTimer** (FixUp Move): If two cards are visible, start a countdown timer. If *HideCards* isn't called by the current player before the timer fires, this will propse *HideCards*.

#### common Move Types

There is a fair bit of boilerplate to implement a move, and you'll define a large number of them for your game. There are also patterns that recur often and are tedious and error-prone to implement.

That's why there's a `moves` package that defines three common move types. You embed these moves anonymously in your move struct and then only override the methods you need to change. In some cases you don't even need to implement your own `Legal` or `Apply` because the base ones are sufficent.

##### base.Move and moves.Default

base.Move is the simplest possible base move. It implements stubs for every required method, with the exception of `Apply` and `Legal` which you must implement yourself. This allows you to minimize the boilerplate you have to implement for simple moves. Almost every move you make will embed this move type either directly or indirectly.

In practice though you'll use a base move that adds a little bit more functionality: moves.Default. moves.Default embeds base.Move but adds more meaty Legal logic and the ability to override certain methods via With* constructors in the moves package.

Default includes a lot of base functionality and defaults. The most important is its `Legal()` method, which is where much of the notion of Phases is implemented. More on that in a later section. For now it's important to know that if you embed a move anonymously in your own move struct, it's very important to always call your "super"'s Legal method as well, because non-trivial logic is encoded in it in Default.

Another simple type of move is `FixUp`. It's a simple embedding of `Default`, but
if your move is a FixUp move it's best to embed it so that
`moves.AutoConfigurer` will treat it as a FixUp move automatically.

##### moves.CurrentPlayer

Many Player moves can only be made by the CurrentPlayer. This move encodes which player the move applies to (set automatically in `DefaultsForState`) and also includes the logic to verify that the `proposer` of the move is allowed to make the move, and is modifiying their own state. (This logic is slightly tricky because it needs to accomodate `AdminPlayerIndex` making moves on behalf of any player).

In typical use you embed this struct, and then either declare your move's legality (see "Declarative Move Legality" just below — this is the primary, recommended way today) or, for logic the declarative catalog can't express, override `Legal` and call the embedded type's `Legal` at the top of your own. (memory's `moveHideCards` was once the example here; its checks — how many cards are left to reveal, and whether any card is still showing — turned out to be expressible declaratively, so it has since been migrated. See "Declarative Move Legality" below.)

A check that genuinely *cannot* be declared is a comparison between two components' values. memory's `moveCaptureCards` captures the two face-up cards only if they are the *same type*, comparing the `Type` of one revealed card against the `Type` of another — and the catalog has no "two components' properties equal each other" primitive (nor a Timer-state predicate, which its sibling `moveStartHideCardsTimer` also needs). It embeds `moves.FixUp` rather than `moves.CurrentPlayer`, but the escape-hatch shape is identical — super-call the embedded `Legal` first, then add your own checks:

```go
// examples/memory/moves.go — moveCaptureCards.Legal (abridged)
func (m *moveCaptureCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	// ... gather the two face-up cards into revealedCards ...

	cardOneType := revealedCards[0].Values().(*cardValue).Type
	cardTwoType := revealedCards[1].Values().(*cardValue).Type

	if cardOneType != cardTwoType {
		return errors.New("The two revealed cards are not of the same type")
	}

	return nil
}
```

Similarly, note that if you have your own logic in `DefaultsForState`, you should not forget to also call the embedded `DefaultsForState`.

##### Declarative Move Legality

Everything above — overriding `Legal()`, calling your embedded type's `Legal()` first, returning `errors.New(...)` for each failure condition — still works, forever. But for a large and growing class of legality checks, you don't have to write a `Legal()` method at all. Instead you *declare* the conditions as data, and the framework evaluates them for you.

**This is purely sugar — read this paragraph before anything else in this section.** `Legal(state, proposer) error` remains the ground-truth contract. The imperative chain you just read about remains unchanged for every move that doesn't opt in. Declarative legality is an additional way to write a `Legal()`-equivalent for moves based on `moves.Default`, `moves.CurrentPlayer`, `moves.FixUp`, `moves.FixUpMulti`, or `moves.StartPhase`.

###### A real before/after

Here is memory's `moveRevealCard`, a `moves.CurrentPlayer` move, before this framework feature existed. This is not an invented example — it's the actual `Legal()` body that shipped for years, preserved verbatim in a comment in the current source (`examples/memory/moves.go:39-62`) for exactly this kind of historical reference:

```go
// examples/memory/moves.go:39-62 (historical -- preserved in a comment, no
// longer compiled; see the "after" below for what replaced it)
func (m *moveRevealCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
        return err
    }
    game, players := concreteStates(state)
    p := players[game.CurrentPlayer.EnsureValid(state)]
    if p.CardsLeftToReveal < 1 {
        return errors.New("You have no cards left to reveal this turn")
    }
    c := game.HiddenCards.ImmutableComponentAt(m.CardIndex)
    if c == nil {
        if game.VisibleCards.ImmutableComponentAt(m.CardIndex) == nil {
            return errors.New("there is no card at that index")
        }
        return errors.New("that card has already been revealed")
    }
    return c.MayMoveToSlot(game.VisibleCards, m.CardIndex)
}
```

Three checks, each with its own hand-written error string, plus a `MayMoveToSlot` pre-check to guarantee `Apply()`'s `MoveTo` call can't fail. Here is the *entire* move today — `Legal()` is gone:

```go
// examples/memory/moves.go:19-23 (verbatim)
//boardgame:codegen
type moveRevealCard struct {
    moves.CurrentPlayer
    CardIndex int
}

// DefaultsForState is unchanged -- omitted here, see moves.go.
// Apply is unchanged -- omitted here, see moves.go.
```

The three checks moved to where the move is *installed*, in `main.go`'s `ConfigureMoves`, as data instead of code:

```go
// examples/memory/main.go:309-311,317-322 (verbatim)
revealCardConfig := auto.MustConfig(
    new(moveRevealCard),
    moves.WithHelpText("Reveals the card at the specified location"),
    moves.WithLegalPreconditions(
        legal.PropAtLeast("player.CardsLeftToReveal", 1).WithMessage("reveal.no_cards_left"),
        legal.RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
        legal.MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
    ),
)
```

`moves.CurrentPlayer.Legal` — still called, exactly as before, since `moveRevealCard` no longer overrides `Legal()` at all — now detects the declared plan and evaluates it instead of running a chain. It runs, in order: `CurrentPlayer`'s own contributed proposer check (unchanged), then these three, base-first-then-declaration-order, exactly matching the old chain's checks and their historical first-failure precedence. `RevealableCardAt` is memory's one *purpose-built* predicate — a 12-line hand-written `Evaluate` function living in the catalog (`legal/catalog_purpose.go`) for the two-branch "no card here" vs. "already revealed" disambiguation that isn't a simple relation over one path; see "The escape hatch" below for when to reach for a purpose-built predicate versus `LegalCustom`.

One string needed to be preserved explicitly, since the catalog's generic default text for `PropAtLeast`'s failure doesn't match the old bespoke message:

```go
// examples/memory/main.go:365-369 (verbatim)
func (g *gameDelegate) ConfigureLegalTemplates() map[string]string {
    return map[string]string{
        "reveal.no_cards_left": "You have no cards left to reveal this turn",
    }
}
```

`RevealableCardAt`'s two Fail branches and `MayMoveToSlot`'s pass-through already default to the exact legacy text (`legal.DefaultTemplates()`), so no override was needed for those. See "Templates" below.

###### `WithLegalPreconditions` and the catalog

`moves.WithLegalPreconditions(specs ...legal.Spec)` is a `CustomConfigurationOption`, passed to `auto.Config`/`auto.MustConfig` just like `WithHelpText` or `WithSourceProperty`. Calling it opts a move in, even with zero specs; the zero-argument spelling publishes and evaluates only the move base's contributed checks. Each `legal.Spec` names one predicate from the `legal` package's catalog (`peer to constraints — see package legal`); at boot, `NewGameManager` resolves every spec, validates every path it references, and assembles one ordered plan per move type: the move's *contributed* specs (derived automatically from the supported move base's own configuration — phase, move-progression, stack constraints, and, for `CurrentPlayer`, the proposer check) first, then your authored specs from `WithLegalPreconditions`, in the order you wrote them. Field-independent predicates may be memoized, but memoization never changes evaluation or ledger order. Implementing `LegalCustom` or calling `WithoutLegalPrecondition` also opts the move in automatically, because each is already an unambiguous request to assemble a plan.

For an existing `Legal()` method: confirm its base is supported; remove the base super-call; move fixed path-shaped gates into `WithLegalPreconditions` in their original order; and, if algorithmic residue remains, rename `Legal` to `LegalCustom`. `LegalCustom` automatically opts the move in and runs last. Keep a wholesale `Legal()` only for unsupported bases or logic that cannot safely compose with a plan.

The most common catalog predicates (full list: `legal.DefaultConstructors()`; representative signatures are compile-checked in `examples/memory/tutorial_snippets_test.go`):

| Predicate | Passes when | Facet read |
|---|---|---|
| `legal.PropAtLeast(path string, n int)` | the int property at `path` is `>= n` | values |
| `legal.PropCompare(path, op string, n int)` | the int property at `path` compares to `n` via `op` (`"<"`, `"<="`, `">"`, `">="`, `"=="`, `"!="`) | values |
| `legal.PlayerBool(prop string)` | the bool property `prop` on the relevant player is `true` (sugar for `PlayerBoolIs(prop, true)`) | values |
| `legal.PlayerBoolIs(prop string, want bool)` | the bool property `prop` on the relevant player equals `want` — the negation leaf for player bools (e.g. `PlayerBoolIs("DoneWithPhase", false)`) | values |
| `legal.PlayerBoolAt(player legal.PlayerSelector, prop string, want bool)` | the selected player's bool property equals `want`; use a semantic behavior helper below when one exists | values |
| `legal.PlayerHasSubmitted` / `PlayerHasNotSubmitted` | the selected player's `behaviors.PlayerSubmission` flag has the requested state | values |
| `legal.PlayerIsActive` / `PlayerIsInactive` | the selected player's `behaviors.InactivePlayer` flag has the requested state | values |
| `legal.PlayerSeatIsFilled` / `PlayerSeatIsClosed` | the selected player's `behaviors.Seat` flag is true | values |
| `legal.PlayerIsAdmin` | the selected player's `behaviors.GameAdministrator` flag is true | values |
| `legal.StackCount(path, op string, n int)` | the stack at `path`'s `NumComponents()` compares to `n` via `op` (same op vocabulary as `PropCompare`) | count |
| `legal.StackEmpty(path string)` | the stack at `path` has zero components | non-empty (safe to reveal under `PolicyNonEmpty`) |
| `legal.StackNotEmpty(path string)` | the stack at `path` has at least one component | non-empty |
| `legal.PropEquals(path, value string)` | the property at `path` equals `value` — dispatches on the property's resolved type: int (`value` parses as int), bool (`"true"`/`"false"`), enum (`value` is a value NAME), or `PlayerIndex` (`value` is an int, or `"observer"`/`"admin"`) | values |
| `legal.PropNotEquals(path, value string)` | the exact negation of `PropEquals` (an `Unknown` — unparseable value, unknown enum name — is never flipped to a Pass) | values |
| `legal.ComponentPresentAt(stackPath, idxField string)` | the stack at `stackPath` has a non-nil component at the int index named by `idxField` | occupancy |
| `legal.ComponentAbsentAt(stackPath, idxField string)` | the exact negation of `ComponentPresentAt` — no component at that index | occupancy |
| `legal.ComponentPresentAtKey(stackPath, keyField string)` | like above, but the slot is identified by an enum-valued key (e.g. a board position) | occupancy |
| `legal.MayMoveTo(srcPath, dstPath, idxField string)` | the component at `idxField` in `srcPath` could legally move into `dstPath` (`ImmutableComponentInstance.MayMoveTo`) | values |
| `legal.MayMoveToSlot(srcPath, dstPath, idxField string)` | like above, but into the *same* index slot in `dstPath` (the "mirrored stacks" pattern, e.g. memory's `HiddenCards`/`VisibleCards`) | occupancy (src) + values (dst, idx) |
| `legal.Any(subs ...legal.Spec)` | at least one of `subs` passes (Kleene: Pass beats Unknown beats Fail) | union of children |
| `legal.AllActivePlayers(inner legal.Spec)` | `inner` holds for every active (non-inactive) player; `inner` must be `PlayerBool`/`PlayerBoolIs`, a player-path `PropAtLeast`/`PropCompare`, or an `Any` of those (int/bool-typed inner leaves only — see Limits) | per-player values |
| `legal.RevealableCardAt(hiddenPath, visiblePath, idxField string)` | purpose-built two-branch occupancy check (see above) | occupancy |
| `legal.ComponentPropEqualsCurrentPlayer(stackPath, keyField, prop string)` | the named component property at that slot equals the current player's identity (checkers' color↔player mapping) | values |

`legal.ProposerIsCurrentPlayer()` exists too, but you'll rarely author it directly — `moves.CurrentPlayer` contributes it automatically, matching what `CurrentPlayer.Legal()` always checked. For a move that acts on behalf of a player named by one of its fields without requiring that player to be the current player, author `legal.ProposerIsPlayerFromMove("TargetPlayerIndex")`; admin engine moves bypass that actor check explicitly, while Observer and Any are rejected as actors.

**Path grammar**, for any predicate above that takes a `path`/`stackPath`/`idxField` string argument: `"game.X"`, `"player.X"` (the current player), `"proposer.X"` (the concrete proposing player), `"move.X"` (a field on the move being evaluated), `"players[*].X"` (quantifier-only, inside `AllActivePlayers`), and `"players[move.<Field>].<Prop>"` — the player state of whichever player the move's own `<Field>` names. That field must be `boardgame.PlayerIndex`-typed; a plain int is a boot error because nothing marks it as a player index.

For behavior-aware predicates, prefer typed selectors instead of writing those paths yourself:

```go
moves.WithLegalPreconditions(
    // Correct during simultaneous play, where CurrentPlayer may be AnyPlayerIndex.
    legal.PlayerHasNotSubmitted(legal.Proposer()),
    // Select a player through a PlayerIndex field on the move.
    legal.PlayerSeatIsFilled(legal.PlayerFromMove("TargetPlayerIndex")),
)
```

`legal.CurrentPlayer()` selects `player.X`; `legal.Proposer()` selects `proposer.X`; and `legal.PlayerFromMove("Field")` selects `players[move.Field].X`. `PlayerSelector` is symbolic: all three variants share one player-index-source resolver and never manufacture a player for a sentinel. During simultaneous play `AnyPlayerIndex` is only the current-player wildcard; the proposer remains concrete. Observer and Any are invalid proposers. Semantic behavior helpers used with `Proposer()` explicitly bypass actor eligibility for `AdminPlayerIndex`; current-player and move-selected state requirements still run.

Behavior helpers describe canonical persisted bool fields; they do not call arbitrary behavior methods or inject policy merely because state embeds a behavior. The move explicitly chooses the rule, player, and polarity.

**The catalog's rule of growth:** first prefer a general relation over a property path. Add a purpose-built predicate such as `RevealableCardAt` only for small, reusable branchy logic that cannot be expressed by those relations. Arithmetic, loops, and hidden game-specific business logic belong in computed state or `LegalCustom`, never an inline lambda.

###### Templates

Every declarative failure carries a `Message{Template, Bindings}` — a template *key*, never a pre-baked string — so it can be localized, grepped, and re-rendered anywhere (server logs, the fixup rejection log, someday a TS client). `legal.DefaultTemplates()` ships default English text for every built-in predicate's failure. A game overrides or adds keys via an optional delegate method, validated at boot (an unregistered key referenced anywhere is a boot error naming the move):

```go
// examples/blackjack/main.go:343-347 (verbatim)
func (g *gameDelegate) ConfigureLegalTemplates() map[string]string {
    return map[string]string{
        "cleanup.players_unfinished": "not all active players have finished their turn",
    }
}
```

Use `.WithMessage("your.key")` on any `legal.Spec` to point its failure at a specific key instead of the predicate's default, as memory's `PropAtLeast(...).WithMessage("reveal.no_cards_left")` did above.

###### The escape hatch: `LegalCustom` and `WithoutLegalPrecondition`

Not everything belongs in the catalog, and that's fine — it's not a gap to be worked around, it's the design. Two knobs handle the remaining cases:

**`LegalCustom`** runs imperative code *after* every declared precondition has passed — the residue that doesn't fit a relation-over-a-path. Checkers' `moveMoveToken` combines three declarative gates with one purpose-built predicate, then a graph-walk of legal capture/move destinations that stays fully imperative:

```go
// examples/checkers/moves.go:165-187 (verbatim)
func (m *moveMoveToken) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

    g := state.ImmutableGameState().(*gameState)

    c := g.Spaces.ImmutableComponentAtKey(m.TokenIndexToMove.Value())

    t := c.Values().(*token)

    //If it's one of the legal spaces, great.
    for _, space := range t.FreeNextSpaces(state, m.TokenIndexToMove.Value().Int()) {
        if m.SpaceIndex.Value().Int() == space {
            return nil
        }
    }

    for _, space := range t.LegalCaptureSpaces(state, m.TokenIndexToMove.Value().Int()) {
        if m.SpaceIndex.Value().Int() == space {
            return nil
        }
    }

    return legal.Errorf("checkers.illegal_dest", nil)
}
```

`legal.Errorf(templateKey, bindings)` returns an `error` that carries a structured, template-rendered message, exactly like a declarative Fail — use it instead of `errors.New` inside `LegalCustom` when you want the same explainability the catalog gets for free. A plain `errors.New` still works; it's wrapped as a one-off template. Implementing `LegalCustom` automatically opts a supported move into declarative legality, so a custom-only move needs no ceremonial constructor option. `LegalCustom` on an unsupported base or on a move that also wholesale-overrides `Legal()` is a boot error, since no safe plan seam could reach it.

**`WithoutLegalPrecondition(name moves.PreconditionName)`** suppresses one *contributed* check by its stable name — pass one of the exported constants `moves.PreconditionInPhase`, `moves.PreconditionInProgression`, `moves.PreconditionStackConstraints`, or `moves.PreconditionProposerIsCurrentPlayer` — for a move that wants to opt out of something it would otherwise inherit. Calling it also opts the move in; requiring a separate marker would make an explicit suppression silently inert:

```go
// synthetic example, compile-checked in
// examples/memory/tutorial_snippets_test.go's
// TestTutorialSnippetWithoutLegalPrecondition
auto.Config(
    new(moveCaptureCards), // embeds moves.FixUp
    moves.WithLegalPhases(phaseNormalPlay), // usually via moves.AddForPhase
    moves.WithoutLegalPrecondition(moves.PreconditionInPhase), // legal in ANY phase
)
```

Suppression is validated at boot, so it can never silently do nothing:

- A name that matches **no check the move actually contributes** is a boot error naming the move, the unmatched name, and the move's real contributed names. This catches both a typo (`"inphase"` — another reason to pass the constants, which make typos a compile error) and suppressing a check the move never had (e.g. `moves.PreconditionInProgression` on a move with no move progression — in the example above, dropping the `WithLegalPhases` line would make the `inPhase` suppression itself a boot error).
- A suppression automatically assembles a plan, even when it is the move's only declarative configuration.
- Suppressing `moves.PreconditionProposerIsCurrentPlayer` on a move that embeds `moves.CurrentPlayer` is a **boot error** of its own: the plan can drop the atom, but `CurrentPlayer.Legal()`'s own imperative proposer check still runs — the suppression would be a no-op on actual legality while telling the client the move is proposable by anyone, exactly the client/server divergence the system exists to prevent. If a move genuinely shouldn't be proposer-gated, embed `moves.Default` (or `moves.FixUp`) instead of `moves.CurrentPlayer`.

###### What the client gets for free

None of this requires the client to change anything — `LegalForPlayer`, `LegalForPlayerError`, and `LegalForAnyone` on each move form are unchanged, byte-identical for an opaque (non-declarative) move. But for a move that opted in, the server ships an additional per-predicate ledger alongside them (`server/api/main.go:80-121`; schematic — field names match `server/api/main.go`'s `preconditionEntry` struct tags; not literal output):

```jsonc
"Preconditions": [
  {"name": "proposerIsCurrentPlayer", "verdict": "pass", "evaluable": true, "provisional": true},
  {"name": "propAtLeast", "args": ["player.CardsLeftToReveal", "1"], "verdict": "fail",
   "message": {"template": "reveal.no_cards_left"}, "evaluable": true},
  {"name": "custom", "verdict": "unknown", "evaluable": false}
]
```

Each entry is `pass`/`fail`/`unknown` for one named predicate, plus:

- **`evaluable`** — whether the generic client catalog knows the predicate's semantics, its complete expression is serialized, and every property it reads survives sanitization for the viewer. Game predicates default to false; `LegalCustom` is false; compositors or quantifiers whose child expression is not shipped are false.
- **`provisional`** — this verdict was computed against server-chosen default field values (the same caveat `LegalForPlayer` already carries at the whole-move level), so filling out the move's form differently could change the answer.
- **`message`** — present on `fail`/`unknown`; bindings are stripped entirely when `evaluable` is `false`, so a lower-privileged viewer is never handed data derived from state they can't see, even indirectly through a binding.

Today the *server* is still the one evaluating every predicate (there is no TypeScript evaluator yet — a designed-for follow-up); the ledger's value right now is explainability (a structured reason for every check, not just one flattened error string) and future-proofing the wire format for client-side evaluation later.

###### Limits (read this before relying on the catalog)

The wire catalog is v4. It includes counts, typed equality, move-field-indexed player paths, explicit negation leaves, typed proposer selectors, explicit admin policy, and the stricter client-evaluability contract. Some honest boundaries remain:

- **`player.X` still means current player, not proposer.** Use `legal.Proposer()` for a proposing player's persisted behavior state during simultaneous play, or `legal.PlayerFromMove("Field")` when the move explicitly names the relevant player. Proposer reads are conservatively marked client-evaluable only when every player's sanitization policy exposes that property; this avoids assuming the ledger viewer and proposer are always the same identity.
- **Stack-count/emptiness now has catalog primitives** (`StackCount`/`StackEmpty`/`StackNotEmpty`, above) — the v1 gap that blocked the most real migrations (debuganimations, pass, darwin, valentine all hit it) is closed.
- **Typed equality now has a catalog primitive** (`PropEquals`/`PropNotEquals`, above) — int/bool/enum/`PlayerIndex` compares, including the `"observer"`/`"admin"` specials, no longer force a fallback to `LegalCustom`. One honest wrinkle: an unknown enum value NAME or a mismatched-type value is a `LegalUnknown` at *evaluate* time, not a boot-time construction error (a constructor-time typo guard catches most cases when a chest is available, but the "unknown enum name = boot error" design-doc aspiration isn't fully delivered — see the completeness-round spec's implementation notes).
- **Negation is explicit-leaf-only, not general.** `PlayerBoolIs(prop, false)`, `PropNotEquals`, and `ComponentAbsentAt` cover common single-property negations; `any` remains the only compositor, so `(A∧B)∨(C∧D)` stays `LegalCustom`.
- **`AllActivePlayers`'s inner leaf only accepts int/bool-typed properties** (`PlayerBool`/`PlayerBoolIs`, player-path `PropAtLeast`/`PropCompare`, or an `Any` of those) — a per-player quantifier over an enum- or `PlayerIndex`-typed property (e.g. "every active player has voted", where the vote is a `PlayerIndex`) has no expression yet, even though `PropEquals` supports those types at the top level. Real gap, not yet closed — found migrating a werewolf-shaped game.
- **The composition seam is `moves.Default`, `moves.CurrentPlayer`, `moves.FixUp`, `moves.FixUpMulti`, and `moves.StartPhase`.** `CurrentPlayer` has a plan-aware `Legal` override and returns after its `Default.Legal` super-call evaluates a plan, avoiding duplicate proposer enforcement. Other framework move types remain opaque and reject declarative configuration at boot.
- **`MayMoveTo`/`MayMoveToSlot` take a single index field**, used for both the source lookup and (for `MayMoveToSlot`) the destination slot — there's no variant for a source index and a *different* destination index.
- **`DynamicComponentValues` still have no path grammar equivalent.** A check that reads a component's per-game dynamic values (as opposed to its static chest-defined `Values()`) — a card's current face-up type after a swap, a species' population counter, and the like — has no declarative expression at all; it always needs `LegalCustom`. This is the single most common reason a real move stays partially opaque across every game surveyed so far.

None of these are dead ends: the escape hatch (`LegalCustom`) always works, and every one of these is exactly the kind of gap the catalog's "growth rule" above is designed to fill in, one purpose-built predicate at a time, as real games need it.

##### moves.FinishTurn

Another common pattern is to have a FixUp move that inspects the state to see if the current player's turn is done, and if it is, advances to the next player and resets their properties for turn start.

`moves.FinishTurn` defines two interafaces that your sub-state objects must implement:

```go
type CurrentPlayerSetter interface {
    SetCurrentPlayer(currentPlayer boardgame.PlayerIndex)
}
```

must be implemented by your gameState. Generally this is as simple as setting the CurrentPlayer index to that value, as you can see in the example from memory:

```go
func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
    g.CurrentPlayer = currentPlayer
}
```

The next interface must be implemented by your playerStates:

```go
type PlayerTurnFinisher interface {
    //TurnDone should return nil when the turn is done, or a descriptive error
    //if the turn is not done.
    TurnDone() error
    //ResetForTurnStart will be called when this player begins their turn.
    ResetForTurnStart() error
    //ResetForTurnEnd will be called right before the CurrentPlayer is
    //advanced to the next player.
    ResetForTurnEnd() error
}
```

In most cases, your playerState has enough information to return an answer for each of these. However, some games have more complicated logic that must look at other aspects of the State as well. If that's necessary, you can find the state your playerState is part of by inspecting the state that was passed to it via SetState().

`moves.FinishTurn` uses the GameDelegate's `CurrentPlayerIndex` to figure out who the current player is. It then calls `TurnDone` on the playerState for the player whose turn it is. If the turn is done (that is, `nil` is returned), it calls `ResetForTurnEnd` on the given PlayerState, then advances to the next player by calling gameState.`SetCurrentPlayer` (wrapping around if it's currently the last player's turn), and then calls `ResetForTurnStart` on the player whose turn it now is. This is where you typically configure how many actions of each type the current player has remaining.

Memory's implementation of these methods looks like follows:

```go
func (p *playerState) TurnDone() error {
    if p.CardsLeftToReveal > 0 {
        return errors.New("they still have cards left to reveal")
    }

    game, _ := concreteStates(state)

    if game.VisibleCards.NumComponents() > 0 {
        return errors.New("there are still some cards revealed, which they must hide")
    }

    return nil
}

func (p *playerState) ResetForTurnStart() error {
    p.CardsLeftToReveal = 2
    return nil
}

func (p *playerState) ResetForTurnEnd() error {
    return nil
}
```

As you can see from the way the errors are constructed in `TurnDone`, the error message will be included in a larger error message. In practice it will return messages like "The current player is not done with their turn because they still have cards left to reveal". 

Because most of the logic for moves that embed `moves.FinishTurn` lives in methods on gameState and playerState, it's common to not need to override the `Legal` or `Apply` methods on `moves.FinishTurn` at all. You can see this in practice on memory's `MoveFinishTurn` which simply embeds `moves.FinishTurn`.

##### Other move types

moves.Default, moves.CurrentPlayer, and moves.FinsihTurn are only three types of moves defined in the moves package. There are a number of others that are useful in other contexts. More detail about how to use some of them is covered below in the Phases section.

#### moves.AutoConfigurer

The next section will walk through a fully manually example where you define
your own MoveTypeConfig and configure that on your game, before showing how to
instead do it with `moves.AutoConfigurer`. In practice
`moves.AutoConfigurer()`is almost always used to automatically generate a
MoveTypeConfig based on a move, minimizing boilerplate you have to write. You
can learn more about how to use it, and good idioms to follow for defining and
installing moves, in the `moves` package doc.

#### Worked Move Example

Let's look at a fully-worked example of defining a specific move from memory:

```go
//boardgame:codegen readsetter
type moveHideCards struct {
    moves.CurrentPlayer
}
```

MoveHideCards is a simple concrete struct that embeds a `moves.CurrentPlayer`. This means it is a move that may only be made by the player who turn it is.

MoveHideCards is decorated by the magic codegen comment, which means its ReadSetter will be automatically generated. The `readsetter` at the end of the comment tells `boardgame-util codegen` to only bother creating the `PropertyReadSetter` method and not worry about the `PropertyReader` method. It would work fine (just with a tiny bit more code generated) with that argument omitted.

```go
var moveHideCardsConfig = boardgame.MoveConfig{
    Name:     "Hide Cards",
    Constructor: func() boardgame.Move {
        return new(moveHideCards)
    },
    //We don't include CustomConfiguration, which is optional.
}
```

This is the MoveConfig object. This is what we will actually use to install the move type in the GameManager (more on that later).

The `Name` property is a unique-within-this-game-package, human-readable name for the move. It is the string that will be used to retrieve this move type from within the game manager. (You'll rarely do this yourself, but the server package will do this for example to deserialize `POST`s that propose a move).

The most important aspect is the `Constructor`. Similar to other Constructor methods, this is where your concrete type that implements the interface from the core library will be returned. In almost every case this is a single line method that just `new`'s your concrete Move struct. If you use properties whose zero-value isn't legal (like Enums, which we haven't encountered yet in the tutorial), then as long as you use struct tags, the engine will automatically instantiate them for you, similar to how `GameStateConstructor` works.

`moveHideCards` is legal when two conditions hold: the current player has finished revealing (`CardsLeftToReveal` is no longer positive), and a card is still showing to hide. Rather than write a `Legal` method, it *declares* those conditions on its config — the recommended approach (see "Declarative Move Legality" above):

```go
moves.WithLegalPreconditions(
    legal.PropCompare("player.CardsLeftToReveal", "<=", 0).WithMessage("hide.cards_still_to_reveal"),
    legal.StackNotEmpty("game.VisibleCards").WithMessage("hide.no_cards_to_hide"),
)
```

attached to the move when it is installed (the "Declarative Move Legality" section above shows the mechanics). The framework evaluates these in place of a `Legal` method, and `moves.CurrentPlayer`'s own proposer/current-player check is contributed automatically — so the two verbatim legacy strings above are the only game-specific part.

The imperative alternative still works for any move, and is exactly what the frozen chain runs for moves that don't opt in: override `Legal`, super-call the embedded type's `Legal` first (always — even `moves.Default` contributes important logic), then add your own checks. Use that shape (see the `moves.CurrentPlayer` section above) when a check can't be declared.

```go
func (m *moveHideCards) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)

	//Cancel a timer in case it was still going.
	game.HideCardsTimer.Cancel()

	for i, c := range game.VisibleCards.MutableComponents() {
		if c != nil {
			if err := c.MoveTo(game.HiddenCards, i); err != nil {
				return errors.New("Couldn't move component: " + err.Error())
			}
		}
	}

	return nil
}
```

This is our Apply method. There's not much interesting going on--except to note that calling MoveTo can fail (for example, if the stack we're moving to is already max size), so we check for that and return an error. If your Move's `Apply` method returns an error then the move will not be applied. In general it is best practice in `Legal` to check for any condition that could cause your `Apply` to fail, so that failures in `Apply` are truly unexpected. Use `MayMoveTo` or `MayMoveToSlot` in your `Legal()` method to pre-validate component moves — if they return nil, the corresponding `MoveTo` in `Apply()` is guaranteed to succeed. See the "Pre-Validating Moves with MayMoveTo" section earlier in this tutorial.

### NewDelegate

We've now explored enough concepts to build a game. The last remaining piece is to combine everything into a ready-to-use `GameManager` that we can then pass to a server or use in other contexts. We do this by passing our delegate to `boardgame.NewGameManager()`, which calls various life-cycle methods on the delegate to get things set up.

By convention, each game package has a `NewDelegate` method that returns a `boardgame.GameDelegate`. In general you don't need to do anything special in this method, and can just return an instaniation of your gameDelegate object:

```
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
```

Of course, you could do more in this method, but in practice it's enough to just instantiate a zero-value of your gameDelegate, because its Configure methods will be called when the new GameManager based on it is instantiated.

#### Component structs

Remember that each component is immutable, and lives in precisely one deck in the `ComponentChest` for a game type. Specific instantiations of a Game of this GameType will ensure that each component in the chest lives in exactly one position in one stack at every version. Since the component is immutable, each game's version's stacks have pointers to the same shared components across all games that come from that gametype.

The `Component` struct is a concrete struct defined in the core package. It is immutable, and includes a reference to the deck this component is in, what its index is within that stack, and the `Values` of this Component--the specific properties of this particular component within this game's semantics.

For example, a component that is a card from a traditional American deck of playing cards would have two properties in its Values object; `Rank` and `Suit`. (In fact, American playing cards are so common that for convenience a ready-to-use version of them are defined in `components/playingcards`). The `Values` object will be a concrete struct that you define in your package that adheres to the `CompontentValues` interface, which includes the `Reader` interface. This mean--you guessed it--that the `boardgame-util codegen` tool will be useful.

The components for memory are quite simple:

```go
package memory

import (
	"github.com/jkomoros/boardgame"
)

var generalCards = []string{
	"🚴",
	"✋",
	"💘",
	"🎓",
	"🌍",
	"🏖",
	"🏛",
	"⛺",
	"🚑",
	"🚕",
	"⚓",
	"🕰",
	"🌈",
	"🔥",
	"⛄",
	"🎄",
	"🎁",
	"🏆",
	"⚽",
	"🎳",
}

// Two other sets of cards here

const cardsDeckName = "cards"

//boardgame:codegen reader
type cardValue struct {
	base.ComponentValues
	Type    string
	CardSet string
}

func newDeck() *boardgame.Deck {
	cards := boardgame.NewDeck()

	for _, val := range generalCards {
		cards.AddComponentMulti(&cardValue{
			Type:    val,
			CardSet: cardSetGeneral,
		}, 2)
	}

	//The two other sets of cards are added here

	cards.SetShadowValues(&cardValue{
		Type: "<hidden>",
	})

	return cards
}
```

The file primarily consists of two constants--the icons that we will have on the cards, and tha name that we will refer to the deck of cards as. Decks are canonically refered to within a `ComponentChest` by a string name. It's convention to define a constant for that name to make sure that typos in that name will be caught by the compiler.

And then the concrete struct we will use for `Values` is a trivial struct with a single string property, and the `codegen` magic comment. It also embeds `base.ComponentValues` to automatically implement `ContainingComponent()` and `SetContainingComponent()`.

In more complicated games, your components and their related constants might be much, much more verbose and effectively be a transcription of the values of a large deck of cards.

#### ConfigureMoves

Your GameDelegate implements a method called `ConfigureMoves()
[]boardgame.MoveConfig`. This method will be called during the
creation process for a GameManager and all of the returned MoveConfigs will be installed on the manager (NewGameManager will error if any of the Moves are invalid for any reason).

An example that could be for memory is here:

```go
//Not what memory actually does
func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig{
	return []boardgame.MoveType{
		//moveRevealCardConfig and others would be defined in the same file as the move structs they are associated with.
		&moveRevealCardConfig,
		&moveHideCardsConfig,
		&moveFinishTurnConfig,
		&moveCaptureCardsConfig,
		&moveStartHideCardsTimerConfig,
	}
}
```

In practice, however, memory uses `moves.AutoConfigurer`--just as almost every game will--to automatically generate MoveConfigs.

```go
func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {

	//...some lines elided...

	auto := moves.NewAutoConfigurer(g)

	return moves.Add(
		//... one move type configuration elided ...
		auto.MustConfig(
			new(moveHideCards),
			moves.WithHelpText("After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player."),
		),
		auto.MustConfig(
			new(moves.FinishTurn),
		),
		auto.MustConfig(
			new(moveCaptureCards),
			moves.WithHelpText("If two cards are showing and they are the same type, capture them to the current player's hand."),
		),
		auto.MustConfig(
			new(moveStartHideCardsTimer),
			moves.WithHelpText("If two cards are showing and they are not the same type and the timer is not active, start a timer to automatically hide them."),
		),
	)
}
```

Technically the moves.Add() is fully optional and it would be equivalent to replace it with `[]boardgame.MoveConfig{...}`. However, the moves.Add convenience method is idiomatic for games with phases, as descirbed in the section on Phases, below, so we include it.

`moves.AutoConfigurer` is a very powerful tool. It automatically generates move constructors, and even move names (based on the name of the struct). In this case, you can see that we didn't even need to create a `MoveFinishTurn` in our package--we could simply use `moves.FinishTurn` directly.

You can learn much more about how to use `moves.AutoConfigurer` in the package doc for `moves`.

More complicated games would use more advanced methods, like `moves.AddForPhase` and others. See the section on Phases, below, for more.

#### ConfigureDecks and ConfigureEnums

There are two other methods that are called on your delegate during the game manager set up.

`ConfigureDecks() map[string]*boardgame.Deck` should simply return a map of names of decks to deck objects for your game.

Memory's is very simple:

```go
func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		cardsDeckName: newDeck(),
	}
}
```

`ConfigureEnums() *enum.Set` should return the enum set for your game. If you're using `boardgame-util codegen`, a simple method that returns the Enums for your package will have already been generated for your gameDelegate.

### Property sanitization

So far all of the properties on State are visible to anyone who cares to look at them. But many (most?) games have some kind of hidden state that should only be revealed to particular players in particular circumstances. In many cases, the whole *point* of the game is to deduce what that hidden state is!

One way would just be to never show that state to the user directly and take care to never render it in the UI. But that's effectively security by obscurity--anyone who was curious could poke in DevTools, discover the secret, and then gain an unfair advantage.

For this reason, the core engine introduces the notion of **sanitization**. This also finally explains that last struct tag in the memory example (HiddenCards having `sanitize:"order"`).

The core engine always keeps track of the full, unsanitized state, and all moves operate on that unsanitized state. However, states can be sanitized to be appropriate to show to any given player, for example before the JSON serialization is transmitted to the client. Then, even if a savvy user pokes in DevTools, they'll never be able to discover the hidden information.

Conceptually, every property in your substate objects has a **sanitization policy** (which may vary by player--more on that in a second) that defines how to sanitize that property. The least restrictive is `PolicyVisible`, which doesn't modify the value at all. The most restrictive is `PolicyHidden`, which hides all information. Stacks have many more subtle policies that obscure some or all information (more on those in a bit).

In almost all cases you will define your policy with struct tags. It is possible to override this behavior by re-implementing SanitizationPolicy on your delegate, see the package doc for more. If no sanitization policy is configured for a property, it defaults to PolicyVisible.

The sanitization configuration is a constant and may never change. Policies apply at the granularity of a property, which means that all components in a given stack will have the same policy applied.

This immutability of the policy explains why memory's GameState has two stacks: HiddenCards and VisibleCards. HiddenCards has a policy to never show the value of the cards in that stack (only the presence or abscence of a card in each slot), whereas RevealCard always shows the values of the cards in it. To "flip" a card from hidden to visible, the `RevealCard` move moves the given card from the HiddenCards stack to the same slot in the VisibleCards stack. On the client the two stacks are merged into one logical stack and rendered appropriately (we'll dig into client rendering, and this particular pattern, more later in the tutorial).

Policies are immutable, but different players might see different things for the same property. For example, in a game of poker no player (except an Admin) should ever be able to see the values (or order) of cards in the DrawStack. Similiarly, the only person who should be able to see the values of the cards in a player's poker hand is that particular player (or the admin).

By default, the policy you apply for GameStates and DynamicComponentValues apply to *all* players (except for Admin, who can always see all state). For PlayerStates, the policies by default apply to *other* players. That means that individual players will, by default, always be able to see all of the properties on their *own* PlayerState, but for other PlayerStates the provided policy will apply.

This behavior can be overridden in more detail by being more explicit about which groups the policies apply to and also by defining policies for multiple groups. For more on that, see the package doc. In almost all cases the default behavior is sufficient.

It's also possible to define your own group names and computed group names for property santization. See Advanced Sanitization in this tutorial, below.

As an aside, sanitization is actually a bit more involved than it looks originally, because it must be possible for the client to know which components in two different states are the "same" in order to do animations of items as they move from stack to stack between states--even if the stacks themselves are sanitized. This concept is referrred to as "Ids". In general everything should just work as you expect automatically. If you want to learn more, refer to the Sanitization section of the package doc.

#### Policies in Detail

The following policies are available:

| Policy         | Description                                                                                           |
|----------------|-------------------------------------------------------------------------------------------------------|
| PolicyVisible  | Visible is effectively no transformation                                                              |
| PolicyOrder    | PolicyOrder is similar to PolicyLen, but the order of components is observable                        |
| PolicyLen      | PolicyLen makes it so it's only possible to see the length of a stack, not its order.                 |
| PolicyNonEmpty | PolicyNonEmpty makes it so it's only possible to tell if a stack had 0 items in it or more than zero. |
| PolicyHidden   | PolicyHidden is the most restrictive; stacks look entirely empty.                                     |

Different policies will lead to different animations automatically occurring in
the client. Typically you want PolicyLen for any large stacks, like Draw decks
in a game, and PolicyOrder for shorter stacks, like a player's Hand, where an
astute observer would be able to keep track of how a given player reorganized
their cards in their hand.

When using struct-tag based policies, the string to use is the name of the
Policy, without the Policy keyword, e.g. "visible", "order", "len".

#### Worked example

In most cases, applying a policy is as simple as adding a struct tag to any fields that should not default to PolicyVisible.

Memory's states are defined as follows:

```go
//boardgame:codegen
type gameState struct {
	base.SubState
	CardSet        string
	NumCards       int
	CurrentPlayer  boardgame.PlayerIndex
	HiddenCards    boardgame.SizedStack  `sizedstack:"cards,40" sanitize:"order"`
	VisibleCards   boardgame.SizedStack  `sizedstack:"cards,40"`
	Cards          boardgame.MergedStack `overlap:"VisibleCards,HiddenCards"`
	HideCardsTimer boardgame.Timer
	//Where cards not in use reside most of the time
	UnusedCards boardgame.Stack `stack:"cards"`
}

//boardgame:codegen
type playerState struct {
	base.SubState
	playerIndex       boardgame.PlayerIndex
	CardsLeftToReveal int
	WonCards          boardgame.Stack `stack:"cards"`
}
```

HiddenCards is the only stack that is sanitized; everything else is fully visible.

Now that we know about sanitization, we can finally understand why there are three stacks in game: `HiddenCards`, `VisibleCards`, and `Cards`. 

##### Aside: Merged Stacks

Each stack must be sanitized the same way--if the components are hidden, then **all** of the components are hidden. But in memory, there are cards that are hidden and cards that are revealed in the same area. 

The way we do it is by **merging** two stacks together, so they can be used logically as one read-only stack, both server and client-side. There are two types of merged stacks, and they're both created in a similar way. ``NewOveralappedStack`` returns an overlapped stack, and `NewConcatenatedStack` returns a concatenated stack. An overlapped stack takes the first stack provided and returns those components--unless that slot is empty, in which case whatever is in that location of the second slot is returned. For overlapped stacks, both stacks must be fixed size, and they both must be the same size. Concatenated stacks simply have all of the slots of the first stack followed by all of the slots of the second stack.

We can use tag-based auto-inflation for merged stacks, too. We use either `concatenate` or `overlap` and then pass the property names of the input stacks. Note that because Merged Stacks are fundamentally read only, they must be stored in an immutable stack property in your state object. (One of the rare cases where you want a `MergedStack` or `Stack` property but not a `MutableStack`.) Tag-based auto-inflation requires the source properties to live on the same object. If the source stacks live in different SubStates, expose the small derived value the renderer actually needs (for example, component IDs or a count) as a typed computed property; stack objects inside `Computed` are not expanded by the client.

When you use merged stacks, the convention is to name the hidden stack `HiddenFoo`, the visible stack `VisibleFoo`, and the merged stack that combines them just `Foo`.

That's not a *particularly* interesting example. Here's the states for blackjack:

```go
//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.RoundRobin
	behaviors.CurrentPlayerBehavior
	behaviors.PhaseBehavior
	DiscardStack  boardgame.Stack `stack:"cards" sanitize:"len"`
	DrawStack     boardgame.Stack `stack:"cards" sanitize:"len"`
	UnusedCards   boardgame.Stack `stack:"cards"`
}

//boardgame:codegen
type playerState struct {
	base.SubState
	playerIndex boardgame.PlayerIndex
	HiddenHand  boardgame.Stack       `stack:"cards,1" sanitize:"len"`
	VisibleHand boardgame.Stack       `stack:"cards"`
	Hand        boardgame.MergedStack `concatenate:"HiddenHand,VisibleHand"`
	Busted      bool
	Stood       bool
}
```

As you can see, both the draw stack and the discard stack are hidden (via
PolicyLen), and the hidden portion of each player's hand is also hidden. (Note
that blackjack also uses the same pattern that memory does with a separate
Hidden and Revealed hand, since some of the cards in the hand are hidden.) In
these cases PolicyLen and PolicyOrder are effectively equivalent, because the
order of the cards in those stacks never change anyway.

Note that Blackjack also makes use of Merged Stacks, but with concatenation
instead of overlapping.

That's a whirlwind tour of the core concepts that you'll need to know to
implement just about any game. There are other concepts that are useful in some
cases, but we'll get to those later. For now, we'll turn to how the core logic
of your game is turned into a visible, interactive game within a web app.

### Client Architecture

As mentioned earlier, the web app is split into two: a REST-ful API server where
all of the game logic is conducted (effectively, the logic that we just
described how to define above), and the single-page-app (SPA) webapp that
interacts with that REST endpoint and creates an interactive web app.

The web app itself is very generic and implemented as a collection of web
components. With no additional configuration it makes it possible for users to
create and manage games that are configured on this server instance, treating
them all the same.

When a user visits a URL to view a specific game, the web app fetches the meta-
information for the game (including who is playing in it), and the current
bundle of state. The server then imports the web-component for the renderer for
your specific game (at a known location and name), instantiates it, and passes
the state bundle to it to render.

The client then creates a WebSocket so it will be notified when new versions of
the state are available, at which point it will fetch the state and pass it to
your renderer so it can update its view. It also listens for events that your
renderer emits that instruct the engine to propose a particular move on the
game, which is then forwarded to the server, which decides whether or not it is
legal.

Other features, like the score board, admin controls and debug information (for
users who have admin privileges) and more are all automatically configured.

This means the primary thing you have to implement for the client-side portion
of your game is a web component that takes a state bundle for your game and
stamps out views for it, referred to as a **renderer**.

#### Aside: Users vs Players

The core game engine doesn't keep track of which player is which--it will make
any move on behalf of any player that it is instructed to. It is up to the
server to keep track of who is who and who is allowed to make moves on behalf of
whom.

The server has a notion of **users**. A user is a particular person, who might
be a Player in 0 or more games. Each player in each game the server controls has
a User that is associated. The user is authenticated via their Google identity,
or via a username/password pair specific to your webapp. A user might have a
display name and a picture, which will be displayed in the scoreboard on any
game they're playing.

The server makes sure to authenticate every incoming modification request and
verify that the user has permission to play as that player. (This gets
complicated if the user has admin privileges and wants to make a move on behalf
of another player).

All of this is handled for you automatically. The main thing to know is that the
server contains a significant amount of logic on top of the core game engine to
manage these kinds of concepts and security.

### Renderers

The renderer is a Lit custom element in a known location. It is the primary
client-side object you define. Extend the generated `GameRenderer`; it receives
strict reactive properties including:

* **state**, the deeply readonly, expanded snapshot for the current version.
  Stacks contain visible, hidden, or empty components. Visible component data
  lives under typed `.Values` and optional `.DynamicValues` properties.
* **chest**, the sanitized static component catalogue and enum metadata.
* **diagram**, the result of your `GameDelegate.Diagram()` method, retained as a
  useful fallback.
* **viewingAsPlayer** and **currentPlayerIndex**, including the framework's
  named observer/admin/simultaneous-player sentinels.
* Typed move actions, legality, player presentations, outcome state, animation
  state, and stable timer services through the generated base.

The job of your renderer is to map that snapshot into meaningful Lit markup and
bind typed actions to framework controls. The framework actions own proposal
transport, server-authoritative legality, pending/stale state, animation gates,
and accessible disabled explanations. Renderers do not emit `propose-move`
events or duplicate game rules.

#### location of renderers

The renderer must be in a specific, known location so it can be imported.

Put the ordinary renderer in
`client/boardgame-render-game-GAMENAME.ts`, where `GAMENAME` is what your
`GameDelegate.Name()` returns. Import `GameRenderer` and
`registerGameRenderer` from the generated `client/_game_renderer.js` module;
the decorator owns the exact custom-element tag, so you never hand-type it.

Your game type might be imported into many different servers, so it's best
practice to keep the renderer definition near the package defining your server
code.

The idiomatic way to do this is, within the package that defines your game
type's go code, have a sub-folder structure, as you can see by looking at
memory:

```
memory/
├── client/
│   ├── _game_renderer.ts
│   ├── _move_args.ts
│   ├── _move_names.ts
│   ├── _types.ts
│   ├── boardgame-render-game-memory.ts
│   └── boardgame-render-player-info-memory.ts
├── agent.go
├── agent_test.go
├── auto_reader.go
├── components.go
├── main.go
├── main_test.go
├── moves.go
└── state.go
```

(We'll get to `boardgame-render-player-info-memory.ts` in a bit.)

`boardgame-util serve` assembles configured game clients into the development
package. `boardgame-util build static` does the same for production, and
`boardgame-util check-client` performs the strict isolated compile/freshness
gate without starting a server. During development, `boardgame-util
check-client --fix` first refreshes every generated client file as one
failure-atomic transaction, then runs that same strict gate. Use the read-only
form in CI.

By following this convention, you cleanly keep your client views for a game next
to the server logic, and also make it easy to import the game package into
different servers with a minimum of fuss.

#### Helpful Components

Before we get into a specific worked example, it's important to dig into a
collection of helpful components and what they do. In many cases the components
the framework provides will do most of what you want, and your renderer is
chiefly concerned with databinding the state object into a specific collection of those components.

##### boardgame-card and boardgame-component-stack

Many games make use of cards in different stacks. Implementing styling and
animations (especially animating from one stack to another) is challenging to
get right. Luckily, two key components, `boardgame-card` and `boardgame-component-
stack`, when used together idiomatically, almost always do exactly what
you want using idiomatic CSS layout with things like flexbox and grid to lay them out and then, with minimal configuration, have high-quality, performant animations created.
Their implementation is non-trivial and handles many edge cases and conditions that are not immediately obvious. They use the `Id` machinery briefly described in the Sanitization section above to keep track of which cards--even cards that are hidden--are which in between states and then animate the cards moving from stack to stack appropriately. They even handle cases like cards flipping from visible to hidden--if done naively, the content of the card would disappear immediately before the flip animation plays! In general, it is strongly recommended to use these components.

boardgame-cards are the basic cards. You can instantiate yourself and set their various properties,
but in practice it is best to bind their `item` attribute to each component item in the state.

boardgame-card's size can be affected by two css properties: --component-scale (a float, with 1.0 being default size) and --card-aspect-ratio (a float, defaulting to 0.6666). Cards are always 100px width by default, with scale affecting the amount of space they take up physically in the layout, as well as applying a transform to their contents to get them to be the right size. --card-aspect-ratio changes how long the minor-axis is compared to the first. If the scale and aspect-ratio are set based on the position in the layout, the size will animate smoothly.

It can be finicky to keep card DOM identity stable enough for animation. Define
one renderer-scoped Lit view and give it to each stack that displays that deck.
The generated stack type makes component values strict, and the discriminated
`kind` forces unusual sanitized states to be handled deliberately:

```typescript
import { cardView, html } from '../../src/client.js';
import type { GameState } from './_types.js';

private readonly cards = cardView<GameState['Cards']>({
  render: ({ kind, component }) => kind === 'visible'
    ? html`<div>${component.Values.Type}</div>`
    : null,
  properties: ({ kind }) => ({
    rotated: true,
    faceUp: kind === 'visible',
  }),
});
```

`kind` is `visible`, `hidden`, or `empty`. A visible component has the exact
generated `.Values` and optional `.DynamicValues` types. A hidden component is
intentionally opaque; render no front content and the standard card displays
its back. An empty sized-stack slot becomes a spacer. `cardView` and `tokenView`
type-check standard host properties. `componentView` is the escape hatch for a
custom element extending `BoardgameComponent`.

Create views once as renderer fields, not inside `render()`. The stable recipe
lets the stack retain component hosts across snapshots, which preserves focus,
pooling, and FLIP animation identity. Each factory used with `componentView`
must return a fresh registered component element of one consistent type; invalid
factories fail loudly.

Import the renderer facade (`../../src/client.js`) once; it registers the
curated card, token, stack, board, action, status, layout, and workflow elements.
Do not add side-effect imports from `src/components`. The strict client checker
rejects deep imports for facade-owned elements so a renderer cannot accidentally
depend on an undocumented transitive registration order.

Start an ordinary solo renderer with `boardgame-game-surface`. It supplies the
game's semantic heading, centered responsive bounds, safe narrow-screen
spacing, and named regions without imposing a visual theme or knowing anything
about your state:

```typescript
return html`<boardgame-game-surface heading="Memory">
  <boardgame-game-outcome
    slot="status"
    .finished=${this.gameFinished}
    .animating=${this.animating}
    .winners=${this.gameWinners}>
  </boardgame-game-outcome>

  <!-- The board, zones, and other primary content use the default slot. -->

  <boardgame-action-bar slot="actions" label="Memory actions">
    <!-- Typed action controls. -->
  </boardgame-action-bar>

  <boardgame-turn-status
    slot="status"
    .turn=${this.turnStatus}>
  </boardgame-turn-status>
</boardgame-game-surface>`;
```

The optional `status`, `actions`, and `footer` regions disappear when their
slots are unassigned; `header` content sits beside the required heading. The
heading is visible by default; use `heading-level` only
to fit the surrounding document outline, and `hide-heading` only when another
visible heading already names the game. Style the stable `surface`, `header`,
`heading`, `status`, `content`, `actions`, and `footer` parts, or tune
`--boardgame-game-surface-max-width`, `--boardgame-game-surface-padding`, and
`--boardgame-game-surface-gap`. Drop down to ordinary Lit markup when a game
needs a deliberately unusual shell.

`this.turnStatus` is the generated renderer base's complete turn-presentation
context. The component shows “Your turn” to the acting player, names the current
player for another player or an observer, describes simultaneous turns, and
does not mislabel the admin perspective. It withholds stale announcements while
state animations are running and after the game finishes. Use `.playerLabels`
for display names, or `active-label` / `simultaneous-label` to adjust the two
standard messages. Custom phase, readiness, and multi-step workflow text remains
ordinary game-owned Lit content beside this primitive.

For simultaneous phases where readiness itself is public, use the typed
`boardgame-readiness` building block instead of hand-rolling counts, progress,
and status announcements:

```typescript
const voters = this.state.Players.map((player, playerIndex) => ({
  key: playerIndex,
  label: this.seatPresentations[playerIndex]?.displayName ?? `Player ${playerIndex}`,
  state: player.Eliminated ? 'not-required'
    : player.Vote >= 0 ? 'ready' : 'waiting',
} as const));

html`<boardgame-readiness
  label="Day votes"
  complete-label="All votes cast"
  progress-label="votes cast"
  ready-label="Voted"
  waiting-label="Thinking"
  not-required-label="Eliminated"
  .participants=${voters}>
</boardgame-readiness>`;
```

The `participants` property is a strict array of stable string/number keys,
non-empty labels, and the closed states `ready`, `waiting`, or `not-required`.
The component validates uniqueness and bounds, renders a visible heading,
progress, participant states, and a polite atomic summary, and supports the
closed `list` (default) and `summary` views. Empty required sets are explicitly
“No participants are required,” never misleadingly complete. Theme its exported
parts or `--boardgame-readiness-*` tokens. The optional progress and three state
labels let game language say “votes cast,” “Voted,” or “Eliminated” without
reimplementing the state model.

Only pass readiness already safe for the current viewer. This component does
not infer private choices, make client state secret, or coordinate a reveal.
Submit choices through ordinary typed snapshot-bound actions; model visibility
and synchronized reveal in authoritative game state first.

The facade also exports `ObserverPlayerIndex`, `AdminPlayerIndex`, and
`AnyPlayerIndex` with the same values and names as Go. Prefer these constants and
the `isConcretePlayerIndex()` / `isKnownPlayerIndex()` guards over client-side
magic negative numbers.

Then stamping those components is as simple as binding a sanitized stack to a
`boardgame-component-zone` from your Lit renderer:

```typescript
html`<boardgame-component-zone
  label="Won cards"
  layout="stack"
  messy
  .stack=${this.state?.Players[0]?.WonCards ?? null}
  .componentView=${this.cards}>
</boardgame-component-zone>`
```

The zone supplies a named semantic region, visible heading, occupied-item count,
responsive surface, empty state, CSS parts, and theme tokens. With no actions it
automatically makes every component display-only. Add `.componentActions` and
the exact bound actions control interactivity, so there is no separate disabled
flag to forget. Use `hide-count` or `hide-empty-state` only when those automatic
elements are inappropriate. Use its `heading-actions` slot for small zone-local controls and
its default slot for status or callout content. Drop down to
`boardgame-component-stack` when you need board/spatial geometry or unusual
animation plumbing.

When a renderer has one arbitrary panel per player, let
`boardgame-player-grid` own the collection layout instead of repeating flexbox
breakpoints in the game:

```typescript
html`<boardgame-player-grid>
  ${this.state?.Players.map((player, playerIndex) => html`
    <boardgame-component-zone
      label=${`Player ${playerIndex + 1}'s cards`}
      .stack=${player.Hand}
      .componentView=${this.cards}>
    </boardgame-component-zone>
  `)}
</boardgame-player-grid>`
```

It provides a named Players region, visible heading, useful empty state, and an
auto-fitting grid that collapses to one column in narrow containers. The
children remain ordinary game-owned Lit content, so a panel can be a component
zone, score card, controls, or any custom element. Use `label` for a different
collection name, `hide-heading` when the visible heading would be redundant,
and `--boardgame-player-grid-min-width` / `--boardgame-player-grid-gap` for
layout tuning. Blank labels, invalid heading levels, and blank enabled empty
states fail loudly.

When each player needs more than one zone or value, use
`boardgame-player-panel` as the grid child:

```typescript
html`<boardgame-player-grid>
  ${this.state?.Players.map((player, playerIndex) => html`
    <boardgame-player-panel
        label=${`Player ${playerIndex + 1}`}
        .active=${playerIndex === this.currentPlayerIndex}>
      <div>Score <boardgame-status-text .value=${player.Score}></boardgame-status-text></div>
      <boardgame-component-zone
        label="Hand"
        .stack=${player.Hand}
        .componentView=${this.cards}>
      </boardgame-component-zone>
      <boardgame-action-button slot="actions" .action=${this.move(MoveNames.Pass)}>
        Pass
      </boardgame-action-button>
    </boardgame-player-panel>
  `)}
</boardgame-player-grid>`;
```

The required label, heading, padding, border, content flow, and current-player
badge are automatic. `header`, `status`, `actions`, and `footer` slots keep
panel-local content structured; unassigned optional regions collapse. The
`active` property adds both styling and `aria-current`, while elimination,
selection, roles, and other game-specific states remain your own classes and
content. Use `::part(panel)` and the `--boardgame-player-panel-*` tokens for a
distinctive design, or keep arbitrary markup as a grid child when this semantic
shape does not fit.

The internal stack creates stable card hosts and rerenders their light-DOM content with
Lit whenever their logical slot changes. The view is local to this renderer, so
two games may use the same deck name without a global registration collision.

`layout` is a strict choice of `stack`, `grid`, `fan`, `pile`, `spread`,
`board`, or `spatial`; misspellings fail both TypeScript and at runtime. If a UI
selects a layout dynamically, narrow its string with the exported
`isStackLayout(value)` guard before assignment. Invalid board dimensions, faux
component counts, stagger values, and spatial coordinates also fail loudly
instead of producing partially positioned components.

Prefer the view's typed `properties` callback for component-dependent card/token
presentation. Use `this.cards.withProperties({ rotated: true })` for typed
stack-specific properties; this preserves host identity even when values change.
`components-disabled` is the explicit display-only common case.
`.unsafeComponentAttrs` remains an intentionally named escape hatch for custom
host properties the typed view cannot express. Do not put move names or move
arguments in it. For one typed action per slot, create a target
collection and pass its actions in stack order:

```typescript
const cards = this.state?.Game.Cards ?? null;
const reveals = this.move(MoveNames.RevealCard).targets(
  cards?.Components.map((_card, cardIndex) => cardIndex) ?? [],
  CardIndex => ({ CardIndex }),
);

return html`<boardgame-component-zone
  label="Cards"
  layout="grid"
  .stack=${cards}
  .componentView=${this.cards}
  .componentActions=${reveals.candidates.map(candidate => candidate.action)}>
</boardgame-component-zone>`;
```

That is the complete common-case interaction wiring. The stack owns pointer and
Enter/Space activation, live legality and pending state, `aria-disabled`, focus
semantics, explanations, subscriptions, and cleanup while preserving component
identity and movement animations. Use `null` at a slot that is deliberately not
interactive. The array must contain exactly one entry per stack slot; a mismatch
or an unbound action throws an actionable error instead of silently targeting
the wrong card. Removed proposal keys such as `proposeMove`, `indexAttributes`,
and `data-arg-*` are rejected even through `.unsafeComponentAttrs`; they cannot
make a component look interactive or bypass the typed action path.

For more complex processing, render ordinary Lit content in the view callback.
If the host itself must be custom, use `componentView()` with a factory that
returns a fresh registered element extending `BoardgameComponent`. The framework
checks that the factory never reuses an element or changes host type.

For card art, rule reminders, maps, or other content that deserves a larger
view, compose the same game-owned presentation into the inspector. The common
case needs no event handlers or modal state:

```typescript
html`<boardgame-inspector
  label="Moon vision"
  description="A moonlit path through a forest">
  <img slot="thumbnail" src=${moonThumbnail} alt="">
  <figure slot="detail">
    <img src=${moonArtwork} alt="Moonlit forest path">
    <figcaption>Follow the path beyond the old oak.</figcaption>
  </figure>
</boardgame-inspector>`
```

The framework turns `thumbnail` into a named 44-pixel-minimum trigger and owns
the native modal focus trap, Escape and backdrop dismissal, focus restoration,
scroll containment, phone bottom-sheet sizing, forced colors, and reduced
motion. The dialog's visible `label` is required; `trigger-label` overrides the
default “Inspect …” name. Omit the thumbnail for a useful text trigger. Empty
labels, nested interactive thumbnail controls, or opening without meaningful
`detail` content fail loudly; the thumbnail is presentation inside the
framework-owned button.

Set `.dismissible=${false}` only when accidental Escape/backdrop dismissal would
be harmful; the visible Close control always remains. For route- or
renderer-owned state, bind the boolean `open` property or call `show()` and
`close()`. The typed `inspector-open-changed` event reports `trigger`, `escape`,
`backdrop`, `close-button`, or `programmatic` without making those events game
state. Theme the exported trigger, dialog, panel, header, title, description,
close, and content parts or the `--boardgame-inspector-*` tokens. This primitive
is for presentation/inspection; multi-step trading or configuration workflows
should keep their domain state in a dedicated controller.

##### boardgame-fading-text

In many cases you want to draw attention to values that change as the result of moves. For example, when it's the current player's turn you might want to make that fact obvious. A common way to do that is to have that text expand from that location and fade as it does so, drawing attention to the changed value. `boardgame-fading-text` will do this for you.

The `boardgame-fading-text` element renders a polite live-region callout when
its typed scalar `.trigger` changes. `message="Your Turn"` keeps fixed text;
`auto-message="new"`, `"diff"`, or `"diff-up"` derives text from the new
value. `suppress="falsey"` and `"truthy"` cover conditional callouts. Invalid
policies and non-finite numeric triggers fail loudly. The font size can be
changed with `--message-font-size`; reduced-motion preferences collapse the
effect to 1ms while retaining the announcement.

In many cases there are parts of your UI that show a value in them, and when that value changes you want to draw attention to it. For example, if you have some text that shows the number of cards in a given stack, you might want users to notice when that changes.

Use `boardgame-status-text` when a displayed string or number should call
attention to changes. Bind its typed `.value` property; it displays the current
value, announces changes politely to assistive technology, and uses the
`diff-up` fading strategy by default. Set `.autoMessage=${'diff'}`, `'new'`, or
`'fixed'` when that better describes the change.

```typescript
html`<boardgame-status-text
  .value=${this.state?.Game.Cards.Components.length ?? 0}>
</boardgame-status-text>`
```

Game timers are stable references in renderer state, not clocks that force the
whole game snapshot to change every animation frame. Bind one to the timer
primitive for an accessible label, countdown, smooth progress, idle hiding,
and an expiry announcement:

```typescript
html`<boardgame-timer
  label="Cards hide in"
  .timer=${this.state?.Game.HideCardsTimer ?? null}>
</boardgame-timer>`
```

Use `format="clock"` for `m:ss`, `hide-progress` or `hide-value` when only one
representation is useful, and the `timer`, `header`, `label`, `value`, and
`progress` CSS parts plus `--boardgame-timer-*` tokens for styling. The
generated timer object intentionally exposes only stable `ID`/`IsTimer`
identity; live values come from the route-scoped clock so unrelated renderer
and roster content does not rerender at 60Hz.

For custom UI, construct `new TimerController(this, () => timerReference)` in a
Lit element and render `controller.reading`. The default second cadence updates
only when the displayed second changes; request `{ cadence: 'frame' }` only for
continuous visuals. Controllers unsubscribe on disconnect and fail loudly when
mounted outside a game view or given a malformed reference.

##### boardgame-base-game-renderer

Game renderers should extend the generated `GameRenderer` in
`client/_game_renderer.ts`. It binds `boardgame-base-game-renderer` to the
generated state, component, move-name, native move-input, schema-fingerprint,
and renderer-tag contracts for your game.

**Move Proposal:** Build controls from typed actions. A zero-input move is the
one-liner `this.move(MoveNames.RollDice)`. A move with creator-supplied fields
uses `this.move(MoveNames.PlaceToken).with({ Slot: 3 })`; required-input actions
do not expose `propose()` until the exact generated fields are bound.

Prefer a framework control's `.action` property because it owns activation,
disabled/pending state, accessible explanations, and error presentation:

```typescript
html`<boardgame-action-button .action=${this.move(MoveNames.DoneTurn)}>
  Done
</boardgame-action-button>`

html`<boardgame-die
  .item=${this.state?.Game.Die.Components[0]}
  .action=${this.move(MoveNames.RollDice)}>
</boardgame-die>`
```

For a custom native control, prefer the element-part binding adapter. It owns
disabled, pending, title, and ARIA state while connected and restores the
element's original state when detached. It is proven with Material Web buttons:

```typescript
html`<md-filled-button ${bindMoveAction(this.move(MoveNames.RollDice))}>
  Roll
</md-filled-button>`
```

Pass `{ disabled: true }` as the adapter's second argument to combine a local
application constraint with the action state. Since an arbitrary element-part
cannot create a visible explanation next to itself, prefer
`boardgame-action-button` whenever the framework control fits; it includes a
live visible status. The adapter provides `title` and ARIA explanation only.
Both controls keep transient preview failures activatable: activating again
retries the legality check before proposing.

`boardgame-action-button` requires visible text so controls cannot silently ship
without an accessible name. For an icon-only control, set its typed
`label="Draw a card"`. Invalid or unbound actions throw an actionable authoring
error. Pending submissions automatically show a reduced-motion-safe spinner;
the internal `button`, `label`, `spinner`, and `status` CSS parts plus
`--boardgame-action-background` and `--boardgame-action-color` allow themed
renderers without replacing its interaction behavior.

Group related choices with the facade-registered action bar. It supplies group
semantics, consistent spacing, and switches from a wrapping row to full-width
controls when its own container is narrow:

```typescript
html`<boardgame-action-bar label="Turn actions">
  <boardgame-action-button .action=${this.move(MoveNames.Hit)}>Hit</boardgame-action-button>
  <boardgame-action-button .action=${this.move(MoveNames.Stand)}>Stand</boardgame-action-button>
</boardgame-action-bar>`
```

Set `orientation="horizontal"` to opt out of responsive stacking or
`orientation="vertical"` to stack at every size. The typed `alignment` values
are `start`, `center`, `end`, and `space-between`; invalid values and blank
accessible labels fail loudly. Theme with `--boardgame-action-gap` or the
bar's `bar` CSS part.

Render the server-authoritative verdict with the outcome primitive. It stays out
of the DOM while the final animation is running, then appears and announces the
result only after the board settles:

```typescript
html`<boardgame-game-outcome
  .finished=${this.gameFinished}
  .animating=${this.animating}
  .winners=${this.gameWinners}
  .viewer=${this.viewingAsPlayer >= 0 ? this.viewingAsPlayer : null}>
</boardgame-game-outcome>`
```

A `null` viewer produces a shared/public “Player N wins” verdict; a player index
produces “You won” or “You lost.” Empty winners on a finished game means a draw.
Table renderers may pass `.winnerLabels` in winner order for display names.
Duplicate/invalid winners, premature winners, label mismatches, and sentinel
viewer indexes fail loudly. Theme the `outcome`, `title`, `message`, and `winner`
parts or the `--boardgame-outcome-*` tokens; reduced motion is automatic.

For a board or other set of independent targets, create one typed target action
directly from the unbound move. The mapper receives each native key and must
return the exact generated move input:

```typescript
const slots = this.state?.Game.Slots;
const places = slots
  ? this.move(MoveNames.PlaceToken).targets(
      slots.Components.map((_, slot) => slot),
      slot => ({ Slot: slot }),
    )
  : null;

return html`<boardgame-game-board
  rows="3"
  cols="3"
  .stack=${slots}
  .action=${places}>
</boardgame-game-board>`;
```

`targets()` returns a headless `TargetAction<Key>` whose candidates each carry
the same canonical `BoundMoveAction` used by ordinary controls. It batches all
candidate legality checks into one version-bound request, correlates results by
opaque ID, rejects stale or malformed responses, and preserves the global
single-submission gate. Recreating the same collection during Lit rendering is
cached; it does not refetch or fan out into one request per square. Call
`targets()` on `this.move(name)`, not on an already-bound `.with(...)` action.
If two distinct UI regions intentionally submit the same input, pass
`{ allowDuplicateInputs: true }` as the third argument; duplicate inputs are a
loud error by default because they usually indicate a mapper bug.

`boardgame-game-board` is the row-major presentation adapter for numeric keys.
It validates that dimensions, stack cardinality, target keys, and accessible
labels agree, then supplies native buttons, roving arrow/Home/End navigation,
guarded `aria-disabled` targets, pending state, and visible failures. Illegal
targets remain focusable so keyboard and screen-reader users can discover the
reason. Use `.labelFor=${...}` when its default “B1, occupied” labels are not
specific enough.

For a menu of ordinary labeled choices—players to vote for, cards to name, or
actions scoped by a string key—keep the exact target keys and labels together:

```typescript
const votes = this.move(MoveNames.CastVote).targets(
  eligiblePlayerIndexes,
  VoteTarget => ({ VoteTarget }),
);

return html`<boardgame-target-list
  label="Vote to eliminate"
  .choices=${targetList(votes, playerIndex => displayNameFor(playerIndex))}>
</boardgame-target-list>`;
```

`targetList()` checks the label callback against the exact target-key union and
rejects blank, excessive, or throwing labels. `boardgame-target-list` starts the
single batched preview, renders every choice as a native action button, keeps
illegal choices visible with their reasons, provides list/heading/empty-state
semantics, and supports `layout="grid"` for wider choices. Forged bindings,
blank labels, invalid heading levels, and unknown layouts fail loudly. For rich
game-specific rows such as a card name plus rule text, render
`target.candidates` directly and bind each candidate's `.action`; the headless
`TargetAction` deliberately has no layout assumptions.

For a source-then-destination board (checkers, chess, tactical movement), add a
single Lit reactive controller. It resets selection automatically when the
renderer snapshot changes; the ordinary target action still owns all legality
and submission behavior:

```typescript
private readonly moveToken = new SourceDestinationController<number>(this);

render() {
  const spaces = this.state?.Game.Spaces ?? null;
  const components = spaces?.Components ?? [];
  const playerColor = this.state?.Players[this.proposingAsPlayer]?.Color;
  const interaction = this.moveToken.bind({
    sources: components.flatMap((piece, index) =>
      isVisibleComponent(piece) && piece.Values.Color === playerColor
        ? [index]
        : []),
    destinations: TokenIndexToMove => this.move(MoveNames.MoveToken).targets(
      components.flatMap((piece, SpaceIndex) => piece ? [] : [SpaceIndex]),
      SpaceIndex => ({ TokenIndexToMove, SpaceIndex }),
    ),
  });

  return html`<boardgame-game-board
    rows="8" cols="8"
    .stack=${spaces}
    .sourceDestination=${interaction}>
  </boardgame-game-board>`;
}
```

The board owns re-selection, Escape-to-cancel, accessible selected/source
labels, legal-destination highlighting, and clearing after success. Illegal
destinations keep the source selected and expose the server reason. Source and
destination keys must be disjoint, in range, finite, and unique; ambiguous
configuration fails loudly. Creators do not maintain a parallel selected index,
preview request, disabled-space array, string serialization, or click handler.

For a turn assembled locally before one exact move—word tiles, formation
orders, route planning, or several resource assignments—use a placement draft.
It owns the local item-to-target state and ordinary undo/clear controls while
the final commit remains a generated, server-previewed move action:

```typescript
import { PlacementDraftController } from '../../src/client.js';

private readonly wordDraft = new PlacementDraftController<string, number>(this);

render() {
  const rack = this.state?.Players[this.proposingAsPlayer]?.Rack;
  const tileIDs = rack?.Components.flatMap(tile => tile ? [tile.ID] : []) ?? [];
  const squares = Array.from({ length: 225 }, (_, index) => index);
  const draft = this.wordDraft.bind({
    items: tileIDs,
    targets: squares,
    minPlacements: 1,
    maxPlacements: 7,
    action: placements => this.move(MoveNames.PlayTiles).with({
      // The game chooses its generated wire model; the draft never submits itself.
      Placements: placements.map(({ item, target }) => `${item}:${target}`).join(','),
    }),
  });

  return html`
    <div aria-label="Rack">
      ${tileIDs.map(id => html`<boardgame-placement-item
        .item=${draft.item(id)}
        label=${`Select tile ${id}`}>
        <word-tile .tileID=${id}></word-tile>
      </boardgame-placement-item>`)}
    </div>
    <boardgame-spatial-board
      board-label="Word board"
      .artwork=${wordBoardArtwork}
      .placementDraft=${draft}>
    </boardgame-spatial-board>
    <boardgame-draft-controls
      label="Word placement"
      commit-label="Play word"
      .draft=${draft}>
    </boardgame-draft-controls>`;
}
```

`selectItem()` followed by `place()` is the required keyboard/click path.
`draft.item(id)` gives `boardgame-placement-item` a compile-time-correlated,
44px native selector with pressed, placed, capacity, focus, fallback, and
content-safety behavior. On authored SVG or raster artwork,
`.placementDraft=${draft}` makes the board's exact geometry keys the destination
controls, including the accessible space list and disabled reasons; combining
it with `.action` or legacy `disabledSpaces` fails loudly. For custom target
markup, `draft.target(key)` provides the same `canPlace`, `reason`, occupancy,
and checked `place()` binding. A plain rectangular board uses the identical
composition—`<boardgame-game-board rows="15" cols="15"
.placementDraft=${draft}>`—and verifies that numeric targets cover every cell.
Optional drag handling calls `assign(item, target)` and therefore cannot bypass
the same unknown-item, unknown-target, occupancy, or maximum checks. Placements
are immutable and carry stable IDs; `targetFor()` and `itemAt()` project the
local overlay without mutating the server snapshot. The stock controls provide
responsive Commit, Undo, Clear, count, disabled reasons, polite status, and
rebase notices. Server preview supplies the final legality reason.

The safe rebase default is `clear`: any new version, viewer perspective, game,
or replaced state clears the draft and visibly explains why. Use
`rebase: 'keep-valid'` only when stable item and target IDs make preservation
intentional; unavailable placements are pruned and announced. Undo history is
bounded and never crosses a snapshot boundary. The commit action is created
from the current immutable placement list, so normal schema validation,
animation/submission gates, preview, `ExpectedVersion`, and stale-snapshot
failure all remain in force. Network failures retain the draft for retry;
success consumes the snapshot until the authoritative next state arrives.

When choices do not have destinations—cards to discard, resources to trade,
payment cards, or simultaneous picks—use the sibling `SelectionDraftController`.
It deliberately shares the same stock controls:

```typescript
private readonly payment = new SelectionDraftController<string>(this);

const draft = this.payment.bind({
  candidates: resourceCardIDs,
  minSelected: 2,
  maxSelected: 4,
  action: selected => this.move(MoveNames.Pay).with({
    Cards: selected.join(','),
  }),
});

return html`
  <div aria-label="Payment cards">
    ${resourceCardIDs.map(id => html`<boardgame-selection-option
      .option=${draft.option(id)}
      label=${`Select resource card ${id}`}>
      <resource-card .cardID=${id}></resource-card>
    </boardgame-selection-option>`)}
  </div>
  <boardgame-draft-controls
    label="Payment"
    commit-label="Pay"
    .draft=${draft}>
  </boardgame-draft-controls>`;
```

`toggle`, `select`, and `deselect` all enforce the same typed candidate set and
maximum. Selection order is stable, the arrays are frozen, undo is bounded,
and `clear` remains the snapshot-change default. Explicit `keep-valid` rebasing
retains only keys still offered by the new authoritative snapshot and announces
what was pruned. Both draft controllers expose the one small
`DraftControlsBinding` surface, so Commit/Undo/Clear behavior stays consistent
while the game keeps ownership of card, resource, or board presentation.

`boardgame-selection-option` supplies the accessible shell around each
game-owned visual. It provides a named 44px native toggle, keyboard focus,
`aria-pressed`, and capacity handling; once full, unselected choices are
disabled while selected choices remain available to deselect. It fails loudly
for blank labels, unknown or inconsistent choices, malformed bindings, and
nested interactive content. A visible text fallback appears when nothing is
slotted, and CSS parts and tokens support game-specific styling. Its `disabled`
property is only an extra presentation gate—the draft remains the authority for
selection validity. The single `draft.option(id)` binding deliberately couples
the key to that draft's candidate union, so TypeScript rejects a choice from the
wrong set before the renderer can run.

For a wholly custom interaction, call `activate()` from the user's gesture and
mirror `canActivate`, `reason`, `submission`, and `subscribe()`. `activate()`
retries a transient exact-preview failure before proposing. For headless
one-shot code, `propose()` and `canPropose` deliberately mean “submit only if
everything is ready now”; `propose` is permanently bound and returns a
discriminated promise result (`success`, `server-rejection`,
`network-failure`, `blocked`, or `stale-snapshot`) instead of throwing for an
ordinary move failure. The adapter or framework control is normally safer.
Actions fail closed during animation, schema skew, stale state,
unchecked/failed exact preview, or another pending submission. A
successful submission consumes that snapshot until a newer state arrives; the
server also checks the expected version inside its serialized move loop.

The old `propose-move="MOVENAME"`, `data-arg-*`, and direct `proposeMove()`
renderer APIs have been removed. They bypassed the action's snapshot,
animation, pending, exact-preview, and result contracts. A legacy-looking
attribute inside a game renderer is inert; use `move()`, bind required input
with `with()`, then hand the action to a framework control or call
`activate()` from a custom gesture.

**Move Legality:** The server computes `Legal()` for each non-FixUp move against the current state and ships the result to the client. Your renderer receives this via the `moveLegality` property (set automatically by the framework), and can use two convenience methods:

- `isMoveCurrentlyLegal(moveName)` — returns `true` if the move is legal for the viewing player right now. Typed actions consume this automatically; use the helper only for custom presentation.
- `isMovePossible(moveName)` — returns `true` if the move is structurally legal (legal for any player / admin). Use this to **hide** buttons entirely (e.g. when the move doesn't apply in the current game phase).

The legality info includes three fields per move: `LegalForPlayer` (is it legal for this player?), `LegalForPlayerError` (the error message if not), and `LegalForAnyone` (is it legal for anyone?). These are server-authoritative — no game logic duplication needed in the client.

For a move whose author opted in to declarative legality (see "Declarative Move Legality" above), the move form also carries a `Preconditions` ledger — one entry per declared check, with its own pass/fail/unknown verdict, whether a client could in principle re-derive that verdict itself (`evaluable`), and whether it's provisional (computed against server-default field values). There is no client-side evaluator yet (`isMoveCurrentlyLegal`/`isMovePossible` above remain the whole story for now, and they already reflect the server's declarative evaluation for opted-in moves); the ledger exists today for explainability and to future-proof the wire format for client-side evaluation later.

#### Generated Move Name Constants

When you run `boardgame-util serve` (or `boardgame-util emit-move-names`), the tool generates a `client/_move_names.ts` file for each game package. This file exports typed constants for all player-proposable move names:

```typescript
// Auto-generated — DO NOT EDIT.
export const MoveNames = {
  RevealCard: "Reveal Card",
  HideCards: "Hide Cards",
} as const;

export type MoveName = typeof MoveNames[keyof typeof MoveNames];
```

Import these in your renderer to get type safety and autocomplete instead of error-prone hardcoded strings:

```typescript
import { MoveNames } from './_move_names.js';
import type { MoveName } from './_move_names.js';
```

The generated `GameRenderer` passes `MoveName` and `MoveInputs` to the framework
for you, so `move()`, `isMoveCurrentlyLegal()`, and `isMovePossible()` accept
only generated names without author-written generic parameters.

These files follow the same convention as `auto_reader.go` and `auto_enum.go`: they are regenerated on each serve but should be committed to source control. Only non-FixUp moves (i.e., player-proposable moves) are included.

#### Generated Type Definitions

When you run `boardgame-util serve`, `boardgame-util emit-types`, or
`boardgame-util check-client --fix`, the tool generates the complete client
contract surface as one transaction. A failure in move, board-space, state, or
strict-TypeScript validation leaves the previous complete generation intact.
That surface includes `client/_types.ts` and `client/_game_renderer.ts` for each
game package. The first exports typed state, player, component, computed-value,
constant, and enum contracts. The second binds those types plus move names and
native author inputs into the renderer bases and exact registration decorators:

```typescript
// Auto-generated — DO NOT EDIT.
import type { ExpandedStack, FullGameState } from '../../src/types/boardgame-types.js';

export type PhaseValue = "Setup" | "Playing";

export interface CardsComponentValues {
  Rank: string;
  Suit: string;
}

export interface GameConstants {
  readonly "numCards": 52;
  readonly "friendly": true;
}

export interface GameComputed {
  readonly CurrentPlayerName: string;
}

export interface PlayerComputed {
  readonly Color: string;
  readonly MayBeActive: boolean;
}

export type DynamicComponentValues = Readonly<Record<string, never>>;

export interface GameState {
  readonly CurrentPlayer: number;
  readonly Phase: PhaseValue;
  readonly DrawStack: ExpandedStack<CardsComponentValues>;
  readonly Computed?: GameComputed;
}

export interface PlayerState {
  readonly Hand: ExpandedStack<CardsComponentValues>;
  readonly Score: number;
  readonly Computed?: PlayerComputed;
}

export type State = FullGameState<
  GameState,
  PlayerState,
  GameComputed,
  PlayerComputed,
  DynamicComponentValues
>;
```

Import these types in your renderer to get full type safety and autocomplete on `this.state`:

```typescript
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';

@registerGameRenderer
export class BoardgameRenderGameMyGame extends GameRenderer {
  // this.state?.Game?.DrawStack is now typed as ExpandedStack<CardsComponentValues>
  // this.state?.Players?.[0]?.Score is now typed as number
  // this.chest?.Constants?.numCards is the literal type 52
  // this.chest?.Constants?.numCard is a compile error — no dynamic keys
  // this.isMoveCurrentlyLegal("Bad Name") is a compile error — only valid move names allowed
}
```

The generated registration decorators are also runtime contract boundaries.
They reject a renderer that extends the wrong generated surface base and report
duplicate exact tags immediately, with the game name, renderer class, tag, and
expected base in the error. Use the matching generated decorator for each
surface (`registerGameRenderer`, `registerTableRenderer`,
`registerHandRenderer`, or `registerPlayerInfoRenderer`) instead of calling
`customElements.define()` yourself; this keeps both TypeScript and runtime
diagnostics precise.

**Enum types** are generated as string literal unions. If your game or any imported package (like `playingcards`) defines enums, the corresponding fields will use the union type instead of `string`. For example, if your enum has values "Red" and "Blue", the generated type will be `"Red" | "Blue"`.

**Component values** are generated as interfaces matching the fields on your component value structs. Stack fields in your state are typed as `ExpandedStack<YourComponentValues>`, giving you autocomplete on `component.Values.FieldName`.

**Dynamic component values** are also supported. If a deck has dynamic component values (see [Dynamic Component Values](#dynamic-component-values) below), a separate interface is generated and the stack type gains a second generic parameter:

```typescript
export interface TokensComponentValues {
  Color: ColorValue;
}

export interface TokensDynamicComponentValues {
  Crowned: boolean;
}

export interface GameState {
  Spaces: ExpandedStack<TokensComponentValues, TokensDynamicComponentValues>;
}
```

This gives you type safety on `component.DynamicValues.Crowned` in addition to `component.Values.Color`.

Like `_move_names.ts`, these files are regenerated on each serve but should be committed to source control.

#### Worked Example

In general your renderer stamps state into visual primitives and attaches typed
actions. Pig's complete interaction code is deliberately boring:

```typescript
import { html } from '../../src/client.js';
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { MoveNames } from './_move_names.js';

@registerGameRenderer
export class BoardgameRenderGamePig extends GameRenderer {
  override render() {
    return html`
      <boardgame-die
        .item=${this.state?.Game.Die.Components[0]}
        .action=${this.move(MoveNames.RollDice)}>
      </boardgame-die>
      <boardgame-action-button .action=${this.move(MoveNames.DoneTurn)}>
        Done
      </boardgame-action-button>
    `;
  }
}
```

The generated names prevent typos, generated native inputs prevent missing,
extra, context-owned, or wrong-primitive arguments, and the action carries the
server's legality and explanation. The renderer neither duplicates game rules
nor manually coordinates animation, double-clicks, pending state, stale
versions, transport errors, or accessibility attributes.

#### Player-info

The web app also has a bar along the top of the game that lists each player, their picture, their name, and their player index. It also by default shows whether it's their turn (according to your delegate's `CurrentPlayerIndex`).

You can add information for each player (like their score) by implementing a
`boardgame-render-player-info-GAMETYPE` element. Extend the generated
`PlayerInfoRenderer` from `_game_renderer.ts`; its `state` and `playerIndex`
properties are reactive and strictly typed for your game. Its typed
`playerState` getter is derived from those two inputs, so it cannot become stale
or refer to a different player.

Override the typed `chip` getter to customize the small roster badge. Return a
`text` string, a CSS `color` string, or both; an empty value retains the
framework fallback. The framework observes state changes and publishes the
presentation automatically—do not create reactive chip properties or dispatch
change events:

```typescript
override get chip() {
  return {
    text: this.playerState?.TokenValue ?? '',
    color: 'rebeccapurple',
  };
}
```

Wrong value types, unknown fields, invalid CSS colors, negative indexes, and
indexes outside the current state fail loudly.

Memory's player-info is therefore small:

```typescript
import { html } from '../../src/client.js';
import {
  PlayerInfoRenderer,
  registerPlayerInfoRenderer,
} from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoMemory extends PlayerInfoRenderer {
  override render() {
    return html`Won Cards
      <boardgame-status-text
        .value=${this.playerState?.WonCards.Indexes.length ?? 0}>
      </boardgame-status-text>`;
  }
}
```

The tictactoe example shows how to override the badge/chip color and text.

## Creating your own game

So far we've worked through an example game using a default config. But how do you set up your own game? In this section we'll describe all of the steps to take to get up and running.

First, create a new directory where all of your new games will go. This will be your git repo.

Before we go further we'll want to generate a config.json. In the tutorial to date we've been using the config.SAMPLE.json in the boardgame library.

`boardgame-util` can help us create and modify config files. The rest of the commands in this section assume you're sitting in the root of your new games repo.

```sh
boardgame-util config init
```

This creates config.PUBLIC.json in the current directory, with reasonable starting values.

The default config has mysql as the defaultstoragetype, so we need to get mysql set up for use.

First, install mysql on your system and run it. The rest of the steps assume it's running on port 3306 (default) and has user: `root` and pass: `root`

Now we need to set-up the tables we expect. `boardgame-util` can help us with that, too:

```sh
boardgame-util db setup
```

In the future if we upgrade the library, you can make sure your mysql tables are migrated to the most recent structure by runing `boardgame-util db up`. If they're already migrated it will have no effect.

OK, now we should have mysql set up. Verify everything's working:

```sh
boardgame-util serve
```

When you actually push to production you'll need to set the production mysql config string. You'll run:

```sh
boardgame-util config set --secret --prod storage mysql USERNAME:PASSWORD@unix(CONNECTIONSTRING)/boardgame
```

See storage/mysql/README.md for more on the structure of that property.

OK, so we have the server set up, but we don't have our own game. `boardgame-util` can help us generate a starter game.

```sh
boardgame-util stub examplegame
```

This will start an interactive prompt of a few questions. Feel free to hit [ENTER] to accept the default for each, with the exception of the question that asks if you want tutorial content--accept that. It generates a lot more example code.

(In general if you aren't a beginner you want all of those defaults, but without tutorial content. You can pass `-f` to skip the interactive prompts and accept all of the defaults.)

This made a new directory called examplegame and filled it with lots of starter content to demonstrate how to wire up a complete simple game. 

You still need to add it to your games list, so run:

```sh
boardgame-util config add games github.com/USERNAME/REPONAME/examplegame
```

First refresh the generated contracts and run the fatal gate:

```sh
boardgame-util check-client --fix
```

It transactionally installs a complete generation, checks each configured
client in isolation with the framework's pinned strict TypeScript and Lit
rules, and reports unsafe escape hatches or deep imports. Commit the generated
changes. CI should run plain `boardgame-util check-client`, which is read-only
and reports each stale or orphaned contract as its own clickable diagnostic.
Then run `boardgame-util serve` to play the game.

Remember that as you modify and recompile, you need to run `go generate` every time you modify the defined fields of a struct.

`go test` in that directory will help verify that the game is set up reasonably.

From here on out you can tweak the game and continue running `boardgame-util serve` to play with it!


## Other important concepts

The sections above cover the information you almost always need to know to build a game from start to finish. However, there are other, slightly more complex features and concepts that are optional but sometimes useful for specific types of games. They're described in separate sections below.

### Dynamic Component Values

By default Components are entirely fixed--their values are exactly the same in every game. That works well for things like cards, but isn't sufficiently general. As a simple example, it's not possible to model a Die, because a die has a fixed set of sides that are the same for all games, but also a specific face that is currently face-up. As a much more complex example, the game Evolution allows players to have any number of Species cards in front of them, each with a population size, a body size, consumed food, and up to 4 trait cards.

These use cases are represented by the concept of *Dynamic Component Values*. For decks that have dynamic component values, the values will be stored as an extra section in your State, just like `gameState` and your `playerState`s. On the server, given a state and a component c, you can access the dynamic component values like so:

```go
values := c.DynamicValues(state)
```

On the client, these dynamic component values will be merged in directly on the component objects in the state passed to your renderer. The generated `_types.ts` file (see [Generated Type Definitions](#generated-type-definitions)) will include a `DynamicComponentValues` interface for each deck that has them, and the corresponding `ExpandedStack` type will include both static and dynamic type parameters.

If you look at the JSON output of a state, you'll see that dynamic component values are stored in a section called "Components", with a key for each deck name that has DynamicComponentValues, and then a slot for values associated with each component in that deck. component.DynamicValues is then just a convenience method that fetches the right component values associated with this component.

The way you configure that a given deck has dynamic component values is by the output to `GameDelegate.DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState`. For decks that don't have dynamic values, just return nil. For decks that do have dynamic component values, just return a new concrete struct, just as you would for `GameStateConstructor` and `PlayerStateConstructor`.

If the struct you return from DynamicComponentValuesConstructor also implements the ComponentValues interface, then SetContainingComponent will be called on the struct every time a new one is created, and pass a reference the component it's associated with. This is useful if the dynamicComponentValues needs access to static property of the component it's associated with to do some methods. You can simply anonymously embed base.ComponentValues in your DynamicComponentValues struct to get that reference for free.

When sanitizing dynamic component values, each deck has its own policy. Importantly, though, that policy is only effective if the stack that the component is currently in has a policy of Visible. In most cases it should just behave as you'd naively expect. For more about specific behaviors, see the package doc on Sanitization.

### Computed properties

It's common to define methods on your `gameState` and `playerState` objects to modify the states and also to provide getters for values that can be computed entirely based on the values of specific properties. This works great on the server, but sometimes you want to have those same computed values available on the client in order to do view data-binding more easily.

Declare client-visible computed values in `ConfigureComputedProperties()`. Each
entry couples its name, exact value type, scope, and evaluator, so the runtime
value and generated TypeScript contract cannot drift. Framework values such as
player color remain automatic; unlike the old map-mutation API, there is no base
method to call and no framework map to merge.

Typically this is a simple enumeration of the names of the values and the method calls, like you can see in memory:

```go
func (g *gameDelegate) ConfigureComputedProperties() []boardgame.ComputedProperty {
	return []boardgame.ComputedProperty{
		boardgame.GlobalComputedBool("CurrentPlayerHasCardsToReveal", func(state boardgame.ImmutableState) bool {
			game, _ := concreteStates(state)
			return game.CurrentPlayerHasCardsToReveal()
		}),
	}
}
```

Use the matching `GlobalComputed*` or `PlayerComputed*` constructor for bool,
int, string, player-index, slice, or enum values. Enum constructors also take
the enum itself, which preserves its generated string-literal union. Manager
construction rejects empty or duplicate names, nil callbacks, incompatible
framework overrides, and enums that are not in the game's chest. Configured
keys are always present; use a stable zero/default value instead of conditionally
changing the shape of the client contract.

`boardgame-util emit-types` writes these declarations into that game's exact
`GameComputed` and `PlayerComputed` interfaces. Misspelled or undeclared
computed keys therefore fail TypeScript compilation without declaration
merging or hand-written index signatures.

Note that when this method is called, your state will likely aready have been sanitized, which means that **your computed property methods should return reasonable values for sanitized states**. In most cases you don't have to think much about this, because all sanitization transformations keep the objects of the same "shape". But it is something to keep an eye out for.

Note that although Merged Stacks might *feel* like computed properties, in most cases (as long as the stacks are on the same SubState object), you can simply use tag-based auto-inflation and have the merged stacks live directly on your state objects.

### Enums

There are a number of cases where a given property can be one of a small set of options--what you'd call in other languages an Enum.

Representing those values as an int is OK, but it doesn't allow you to enumerate which values are legal. In addition, you sometimes want to know the string value of the enum value in question.

Boardgame formalizes this notion as an `enum`, which is a valid property type and is defined in `boardgame/enum`. 

You define your named Enums at set up time as part of an `EnumSet`, and list the values that are legal (and their string equivalents). You can retrieve the EnumSet in use from `manager.Chest().Enums()`.

Given an enum, you can create an `enum.Val`, which is a container for a value from that enum. These `enum.Val` and `enum.MutableVal` are legal properties to add to your states and moves, and like stacks can be configured via struct tags, as you can see in blackjack's `state.go`:

```go
//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.RoundRobin
	behaviors.CurrentPlayerBehavior
	behaviors.PhaseBehavior
	DiscardStack  boardgame.Stack `stack:"cards" sanitize:"len"`
	DrawStack     boardgame.Stack `stack:"cards" sanitize:"len"`
	UnusedCards   boardgame.Stack `stack:"cards"`
}
```

Creating an enum is slightly cumbersome and repetitive. You typically create a const block, enumerate all of the values, and then later install each of those values, while passing their string equivalent.

The `boardgame-util codegen` command can also help automate this, as you can see in the blackjack example in `state.go`:

```go
//boardgame:codegen
const (
	phaseInitialDeal = iota
	phaseNormalPlay
)
```

This will automatically create a global `enums` EnumSet, and a global `phaseEnum` that contains the two values, configured with the string values of "Initial Deal" and "Normal Play". You can find much more details on the conventions and how to configure `boardgame-util codegen` in the enums package doc.

Note that the convention is to have your enum constants be package-private (that
is, start with a lowercase letter), although the codegen tool will work either way.

### RangedEnum and Enum Graphs

Sometimes when you're creating a boardgame--especially one with a board and multiple connected spaces--you need to keep track of which spaces are connected to one another.

The enum package also allows you to create a ranged enum. It's just a normal enum, but created with all of the values in the given dimensions:

```go
//returns an enum with 9 items
e := set.MustAddRanged("Spaces", 3, 3)

//returns true
e.IsRange()
```

Under the covers it's just a simple enum with values from 0 to 8, where the string value for 0 is "0,0". But because it was created with AddRange is also has a few additional convience getters to and from the raw index to the multi-dimensional index it represents.

```go
//Returns []int{0,1}
e.ValueToRange(3)

//returns 3
e.RangeToValue(0, 1)
```

Typically to model a board with spaces, you create a RangedEnum of the correct dimensions. Then on your gameState you'd have a SizedStack that is the same size as the RangedEnum. You'd use the Ranged getters to convert a multi-dimensional index into a single-dimensional index into the stack. This set-up works if each space on the board can have only one token; if a given space can host more than one, create a Spaces SizedStack for each player.

```go
const DIM = 8
//TOTAL_DIM is exported as a constant, so it can be used in the tag-based struct inflation.
const TOTAL_DIM = DIM * DIM

chessBoard := set.MustAddRange("Spaces", DIM, DIM)

type gameState struct {
	base.SubState
	Spaces boargame.Stack `sizedstack:"Tokens, TOTAL_DIM"`
}

//retrive the token at space 3,3 in the chessboard
gState.Spaces.ComponentAt(chessBoard.RangeToValue(3,3))
```

`enum/graph` is a package that allows you to create graphs where each value in an enum is a node, and you add edges between nodes. These graphs are useful to test whether indexes in a stack that represents spaces in a game board are adjacent or not.

You can add your own edges between items, but for grid-based boards, NewGridConnectedness() often does what you want. Check out the package doc for more, but here's a quick example:

```go
set := enum.NewSet()
chessBoard := set.MustAddRange("Spaces", 8, 8)

//blackLegalMoves will have moves that are only valid upwards and diagonal.
blackLegalMoves := graph.NewGridConnectedness(chessBoard, DirectionDiagonal, DirectionUp)
redLegalMoves := graph.NewGridConnectedness(chessBoard, DirectionDiagonal, DirectionDown)
```

Graphs also support `ShortestPath(start, end)` and `Distance(start, end)`, which use Dijkstra's algorithm. Edge weights default to 1 when not explicitly set via `SetEdgeWeight`. These are used by the spatial game moves described below.

### Spatial Game API

Many board games have a map or board with spaces that tokens move between. The framework provides reusable building blocks for these "spatial" games: `behaviors.LocationBehavior` tracks a token's position, `enum/graph` provides adjacency and pathfinding, and the `moves` package provides `MoveOnGraph`, `HopAlongPath`, and `AdvanceToken` for common movement patterns.

#### LocationBehavior

`behaviors.LocationBehavior` is a behavior (like `PlayerColor` or `RoundRobin`) designed to be embedded in a `playerState` or `gameState`. It tracks which slot in a `SizedStack` a token occupies.

To use it, embed it in your state struct and use the `location` struct tag to point it at the SizedStack field that tracks position:

```go
type playerState struct {
    base.SubState
    behaviors.LocationBehavior `location:"Location"`
    // Location is a SizedStack with one slot per space on the board.
    // Exactly one slot holds the player's token component.
    Location boardgame.SizedStack `sizedstack:"tokens,NUM_SPACES"`
}
```

The framework automatically calls `ConnectBehavior` and reads the `location` tag for you. If you also need a graph for adjacency and pathfinding, connect it in `FinishStateSetUp`:

```go
func (p *playerState) FinishStateSetUp() {
    p.LocationBehavior.ConnectGraph(myConnectivityGraph)
}
```

You can also skip the tag and wire everything manually in `FinishStateSetUp` if you prefer:

```go
func (p *playerState) FinishStateSetUp() {
    p.LocationBehavior.ConnectLocationStack(p.Location)
    p.LocationBehavior.ConnectGraph(myConnectivityGraph)
}
```

Once wired up, `LocationBehavior` provides:

- `LocationIndex() int` -- returns the index of the slot containing the token
- `MoveTo(targetIndex int) error` -- swaps the token to a different slot
- `Neighbors() []int` -- spaces adjacent to the current position (requires graph)
- `IsConnectedTo(target int) bool` -- whether a space is adjacent (requires graph)
- `ShortestPathTo(target int) ([]int, error)` -- shortest path via Dijkstra (requires graph)
- `DistanceTo(target int) (int, error)` -- distance via Dijkstra (requires graph)

`LocationBehavior` also stores a `LocRemainingPath []int` field that is used internally by `HopAlongPath` for multi-hop animated movement. You generally don't interact with this field directly.

#### MoveOnGraph

`moves.MoveOnGraph` is a player-facing move (embeds `CurrentPlayer`) for moving a token to a destination on the board. The player specifies a `TargetLocation` (an integer index); the framework computes the shortest path, validates it, and stores the path on the player's `LocationBehavior` for animated execution.

Your move struct embeds `MoveOnGraph` and implements the `LocationProvider` interface to tell the framework which `LocationBehavior` to use:

```go
//boardgame:codegen
type MoveMoveToken struct {
    moves.MoveOnGraph
}

func (m *MoveMoveToken) PlayerLocationBehavior(pState boardgame.ImmutableSubState) *behaviors.LocationBehavior {
    return &pState.(*playerState).LocationBehavior
}
```

You may also implement optional interfaces to customize the behavior:

- `SpaceValidator` -- reject certain spaces (e.g., closed or blocked spaces)
- `MovementBudgeter` -- limit how far the player can move per turn
- `FreeMovePredicate` -- allow teleportation to specific spaces (skips adjacency and budget checks)
- `FreeMoveApplier` -- run cleanup after a free move (e.g., discard the card that granted it)

#### HopAlongPath

`moves.HopAlongPath` is a framework-provided FixUp that executes one hop of the stored path per application. Each hop produces a separate game version, giving the client a distinct animation frame for each step of movement (similar to how `DealCountComponents` animates one card at a time).

Register it in your `ConfigureMoves()` before other FixUps that might depend on movement being complete:

```go
auto.MustConfig(
    new(moves.HopAlongPath),
    moves.WithHelpText("Execute one hop of a multi-hop movement path."),
),
```

You don't need to implement any interfaces for `HopAlongPath`. It automatically finds whichever `LocationBehavior` has a remaining path and executes the next hop.

#### AdvanceToken

`moves.AdvanceToken` is a FixUp for deterministic, non-player-driven token movement -- for example, an NPC that patrols a fixed route. The embedding move implements the `TokenAdvancer` interface:

```go
//boardgame:codegen
type MoveMoveNPC struct {
    moves.AdvanceToken
}

func (m *MoveMoveNPC) AdvancableLocation(state boardgame.State) *behaviors.LocationBehavior {
    return &state.GameState().(*gameState).LocationBehavior
}

func (m *MoveMoveNPC) NextAdvanceIndex(state boardgame.ImmutableState, currentIndex enum.ImmutableVal) enum.EnumKey {
    // Simple patrol: advance to the next space, wrapping around
    next := currentIndex.Value() + 1
    if next > maxSpace {
        next = 1
    }
    return next
}
```

Optionally implement `AdvanceCondition` (to gate when the token should advance) and `PostAdvanceHandler` (for side effects after the token moves, such as changing whose turn it is).

#### Client: boardgame-spatial-board

`boardgame-spatial-board` turns authored SVG artwork into the same typed target
interaction used by `boardgame-game-board`. The SVG describes geometry and
labels; `TargetAction` describes legality and activation. Game renderers do not
need a second click-handler/disabled-array proposal system.

Mark the real hit region for every space. Keys are ordinary strings and may
contain punctuation; they are never interpolated into CSS selectors.

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 820 616">
  <path data-board-space="library"
        data-board-label="Library"
        data-board-order="0"
        data-board-group="rooms"
        d="..." />

  <!-- Optional: interaction, keyboard focus, and pieces may use distinct geometry. -->
  <circle data-board-focus-anchor="library" cx="105" cy="88" r="8" />
  <circle data-board-piece-anchor="library" cx="130" cy="110" r="8" />
</svg>
```

String keys are the default even when they look like `"2"`. If the move field
is a numeric enum, opt in explicitly on the root with
`data-board-key-type="number"`; generation then rejects non-canonical or
JavaScript-unsafe integers instead of guessing from their spelling. Every
region needs an accessible label. `data-board-label` is clearest, while an
`aria-label` or direct child `<title>` is also accepted.

Then bind typed targets and an explicit piece projection:

```ts
import { html, piecesFromSizedStacks } from '../../src/client.js';
import { MoveNames } from './_move_names.js';
import { BoardSpaceKeys } from './_board_spaces.js';

const moveToRoom = this.move(MoveNames.MoveToRoom).targets(
  BoardSpaceKeys,
  room => ({ TargetLocation: room }),
);
const pieces = piecesFromSizedStacks(this.positionStacks, BoardSpaceKeys);

return html`<boardgame-spatial-board
  svgUrl="game-src/mygame/board.svg"
  .action=${moveToRoom}
  .pieces=${pieces}>
</boardgame-spatial-board>`;
```

`boardgame-util emit-types` (and dev-server startup) extracts `_board_spaces.ts`
from `board.svg`; `boardgame-util emit-board-spaces` does just this step, and
`--check` verifies freshness. The generated tuple preserves literal keys, so an
SVG rename makes move/projection code fail strict checking. If a sized stack has
an enum sentinel with no physical region, make it explicit in the projection:
`[null, ...BoardSpaceKeys]`.

The component validates duplicate/missing labels, keys, ordering, measurable
regions and anchors, unknown action/piece keys, and stack cardinality. It uses
each anchor's complete SVG transform, so nested transformed groups, a nonzero
`viewBox`, responsive scaling, and letterboxing do not require manual pixel
math. Pointer interaction follows the actual authored region; a compact native
button list provides keyboard and screen-reader access with the same legality
reasons. Loading is abortable and stale-safe, failures are visible and retryable,
and fetched SVG is sanitized before insertion (scripts, event handlers,
`foreignObject`, animation mutation, and external resources are removed).

While authoring, add the boolean `geometry-inspector` attribute to show the
resolved key/order/label, region element, focus and piece coordinates, and
possible bounding-box overlaps. Remove it for the shipped renderer; it is a
diagnostic view, not game state.

For unusual artwork, pass `.geometry=${svg => ...}`. The callback receives the
same sanitized, mounted SVG that is displayed and returns `BoardGeometry`, so
custom regions and distinct anchors can use real measurable elements without
re-fetching or reaching into the component's shadow root. Pointer activation
uses those returned regions, not a hidden dependency on data attributes.

Set `board-label` (default: `Game board`) and optional `board-description` for
overall context. The artwork itself is presentation-only; the generated focus
buttons and compact list are the single accessible interaction model, avoiding
duplicate announcements from SVG-editor metadata.

If the board is a PNG, JPEG, WebP, AVIF, or GIF, keep it as raster artwork and
describe its spaces in normalized coordinates. You do not need to trace the
image into SVG or write resize math:

```ts
import { html, rasterBoardArtwork } from '../../src/client.js';

const artwork = rasterBoardArtwork({
  src: 'game-src/mygame/board.webp',
  // Keep the rendered board square even though the source image is wider.
  viewportAspectRatio: 1,
  fit: 'contain',
  spaces: [
    {
      key: 'harbor',
      label: 'Harbor',
      region: {
        shape: 'circle',
        center: { x: 0.18, y: 0.72 },
        radius: 0.08,
      },
      // Optional: put pieces somewhere other than the hit-region center.
      pieceAnchor: { x: 0.2, y: 0.68 },
    },
    {
      key: 'market',
      label: 'Market',
      region: { shape: 'rect', x: 0.55, y: 0.2, width: 0.3, height: 0.25 },
    },
  ],
});

return html`<boardgame-spatial-board
  .artwork=${artwork}
  pan-zoom
  .action=${this.move(MoveNames.MoveToRoom).targets(
    artwork.spaces.map(space => space.key),
    room => ({ TargetLocation: room }),
  )}>
</boardgame-spatial-board>`;
```

Coordinates run from `0` at the image's left/top to `1` at its right/bottom.
Regions may be circles, rectangles, or polygons. Optional `focusAnchor` and
`pieceAnchor` use the same normalized coordinates. The constructor freezes and
validates the complete descriptor up front: empty or duplicate keys, missing
labels, duplicate keyboard order, out-of-range coordinates, degenerate shapes,
unsupported fit modes, and unsafe image sources fail loudly near the authoring
code.

`fit` is `contain` by default; use `cover` when cropping is intentional or
`fill` when distortion is intentional. The framework puts the raster image and
all regions/anchors in one nested SVG transform, so letterboxing, cropping,
responsive widths, focus placement, piece placement, and animation measurement
cannot silently use different coordinate systems. Omit `viewportAspectRatio`
to use the decoded image's own aspect ratio.

Choose exactly one board source: `svgUrl` for authored SVG or `.artwork` for a
raster descriptor. The custom `.geometry` callback is SVG-only because raster
spaces already live in the descriptor. Everything after source loading is
shared: typed `TargetAction`, piece projections, keyboard/list access, legality
reasons, responsive positioning, geometry inspection, loading errors, and retry.

For a large map, add the boolean `pan-zoom` attribute as above. The board gets
bounded zoom/reset controls, keyboard `+`/`-`/arrow/`0` navigation,
Ctrl/Command-wheel zoom, mouse/touch panning, and pinch zoom. Ordinary taps and
clicks still reach spaces; a gesture that crosses the drag threshold cannot
accidentally propose a move. The compact space list and loading/status content
stay outside the transformed scene. Set `max-zoom` only when the default `4`
is inappropriate. Call `revealSpace(key)` to bring a moved piece or selected
location into view, and `resetViewport()` to return to the complete board.

`boardgame-board-viewport` is also available as a standalone building block for
a custom SVG, canvas, or other large visual. Its default controls work without
configuration; `view` plus `setView(...)` support route-scoped persistence, and
`reveal(element)` brings nested light- or shadow-DOM markers into view. It emits
`board-viewport-change` with the same immutable `{ scale, x, y }` value. Panning
is clamped after zoom, content resize, host resize, or a lower `max-scale`, so
blank space cannot become the main view.

A multi-purpose map may contain different target classes at once: for example,
hex tiles, road edges, and settlement vertices. Give each raster descriptor
space a `group` (or add `data-board-group` to SVG regions), then scope the
current typed action explicitly:

```ts
const VertexGroup = 'vertices';
const artwork = rasterBoardArtwork({
  src: 'game-src/mygame/island.webp',
  spaces: [
    {
      key: 'vertex:0:1',
      label: 'North harbor intersection',
      group: VertexGroup,
      region: { shape: 'circle', center: { x: 0.31, y: 0.18 }, radius: 0.025 },
    },
    // Tile and edge regions may use their own groups in the same descriptor.
  ],
});

return html`<boardgame-spatial-board
  .artwork=${artwork}
  action-group=${VertexGroup}
  .action=${placeSettlementTargets}>
</boardgame-spatial-board>`;
```

The action's candidate keys must still match every key in the selected group
exactly—grouping does not weaken typo or omission detection. A missing group,
candidate from another group, or incomplete candidate set fails loudly. Other
regions remain available as piece/animation anchors but do not activate, appear
in the current compact action list, or receive the visual treatment for an
illegal candidate. Omit `action-group` for ordinary one-purpose boards; the
existing exact match against all geometry remains the zero-configuration path.

Routes, supply lines, patrol paths, and other connections between existing
spaces are data too. Give the board typed path descriptors; it connects each
space's piece anchor (or its region center when no piece anchor is declared),
keeps the line aligned through resize and pan/zoom, and supplies an accessible
route description without making the decorative line interactive:

```ts
import { html, type BoardPathOverlay } from '../../src/client.js';
import { BoardSpaceKeys } from './_board_spaces.js';

type BoardSpace = (typeof BoardSpaceKeys)[number];

const routes = [
  {
    id: 'coastal-supply',
    label: 'Coastal supply line from Harbor through Road to Market',
    spaces: ['harbor', 'road', 'market'],
    tone: 'secondary',
    width: 6,
  },
] as const satisfies readonly BoardPathOverlay<BoardSpace>[];

return html`<boardgame-spatial-board
  svgUrl="game-src/mygame/board.svg"
  .pathOverlays=${routes}>
</boardgame-spatial-board>`;
```

The required stable `id` lets Lit update a path without rebuilding unrelated
routes, and the required `label` makes the route meaningful to assistive
technology. `tone` is the closed set `primary`, `secondary`, `danger`, or
`muted`; theme them with `--board-path-primary`, `--board-path-secondary`,
`--board-path-danger`, and `--board-path-muted`, or target the exported `path`
part. Width is in screen pixels and does not swell when the board zooms.

Unknown spaces, adjacent duplicate points, duplicate IDs, invalid tones or
widths, and excessive path data fail loudly. Paths deliberately do not create
new target semantics: author road edges or route choices as ordinary geometry
spaces/groups and bind a typed `TargetAction`; use `pathOverlays` to show the
resulting connection. Omit it for the common case—the spatial board needs no
route configuration.

`spacePrefix`, `disabledSpaces`, `space-tapped`, `stack`/`stacks`,
`boxForSpace()`, and `tokenPosition()` remain migration adapters for older
numeric-ID boards. New renderers should use `data-board-*`, `.action`, and
`.pieces`; that keeps custom artwork ordinary game-owned SVG without giving up
typed moves, server-authoritative legality, accessibility, or stable animation
anchors.

### Phases

At the core of the engine, there's just a big collection of moves, any of which may be `Legal()` at any time. `ProposeFixUpMove` is called after every move is applied, and any move that is returned is applied. `base.GameDelegate`'s default implementation simply cycles through every move in order, and returns the first one whose `IsFixUp()` returns true, and who is Legal with defaults set for the current state. 

This is fine for simple games like memory, but quickly becomes cumbersome for more complicated games. For example, some games have multiple rounds, where each round is basically a mini-game, where scores accumulate across rounds. For each round you might have to do some set-up tasks (like moving all of the cards from discard to the draw stack, shuffling them, and then dealing out two cards per player), then have the normal play, and then finally some clean-up tasks (collecting the cards remaining in players' hands, tallying up scores).

If you had to write all of your Legal() methods by hand, it would be error-prone and finicky. You'd have to think carefully about how each move could look at the state of the game and figure out that it was its time to be applied. In many cases, it wouldn't be possible to tell that cleanly, and you'd have to add lots of extra properties to your State object to keep track of exactly where you were and what needed to be done.

It'd be a mess!

For that reason, a convention of "Phases" is used. A game can have multiple phases. Moves are only legal to apply in certain phases. In some phases, moves are applied in a specific, prescribed order only.

The concept of Phases is barely represented in the core library at all. Delegates have `CurrentPhase() enum.ImmutableVal` and `PhaseEnum() enum.Enum`, but other than that the notion of Phases is implemented entirely in the (technically optional) `moves` package.

At the core, the notion of Phases is implmented by `moves.Default`'s Legal method--which is why it's so important to always call your super's `Legal` method! `moves.Default` will first check to make sure that the current phase of the game is one that is legal for this move, and then check to see if playing this move at this point in the phase is legal. All other methods and machinery for representing Phases are just about giving moves.Default the information it needs to make this determination.

The actual machinery to implement Moves is not important, other than to know that it can be overriden by swapping out the implementations of a few delegate methods, as covered in the package documentation. This part of the tutorial will primarily just discuss how to use it in practice by examining the blackjack example.

If you're going to support the notion of phases, you'll need to store the current phase somewhere in your state. In `examples/blackjack/state.go` we have:

```go
//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.RoundRobin
	behaviors.CurrentPlayerBehavior
	behaviors.PhaseBehavior
	DiscardStack  boardgame.Stack `stack:"cards" sanitize:"len"`
	DrawStack     boardgame.Stack `stack:"cards" sanitize:"len"`
	UnusedCards   boardgame.Stack `stack:"cards"`
}
```

We also need to define the values of the enum. In `examples/blackjack/components.go` we have:

```go
//boardgame:codegen
const (
	phaseSetUp = iota
	phaseNormalPlay
	phaseScoring
)
```

In general it's easiest to use `boardgame-util codegen`'s enum-generation tool, which we do here.

It's convention to name your phase enum as "phase", and `moves.Default` will rely on that in some cases to create meaningful error messages. If you want to name it something different, override `GameDelegate.PhaseEnum`.

Now we have to tell the engine what the current phase is. We do this by overriding a method on our gamedelegate, much like we do for CurrentPlayerIndex:

```go
func (g *gameDelegate) CurrentPhase(state boardgame.ImmutableState) enum.ImmutableVal {
	game, _ := concreteStates(state)
	return game.Phase
}
```

However, since we're using base.GameDelegate and our Phase property is `Phase` on our `gameState`,
we don't even have to do that. base.GameDelegate's CurrentPhase() already looks for that value
there and returns it.

Now the core engine knows about what phase it is. `moves.Default` will consult that information it is Legal method. But how do we tell `moves.Deafult` which phases a move is legal in?

Moves that are based on `moves.Default` have a `LegalPhases() []EnumKey` method that `moves.Default` consults to see if the game's CurrentPhase is one of those. `LegalPhases()` just returns whatever was passed in `moves.AutoConfigurer` with `WithLegalPhases`. However, setting that manually is error-prone; you have to remember to include it for each move in that phase, and it can be hard to keep track of the order of the moves.

That's why the `moves` package defines `Add`, `AddForPhase`, and `AddOrderedForPhase`, which automatically call the right `WithLegalPhases` and `WithLegalMoveProgression` methods for you. In addition, the `moves` package defines `moves.Combine`, a convenience wrapper to use in your `ConfigureMoves` when you have phases.

You can see this in action in `examples/blackjack/main.go` in `ConfigureMoves`

```go
	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		//...
		moves.AddForPhase(PhaseNormalPlay,
			auto.MustConfig(
				new(moveCurrentPlayerHit),
				moves.WithHelpText("The current player hits, drawing a card."),
			),
			auto.MustConfig(
				new(moveCurrentPlayerStand),
				moves.WithHelpText("If the current player no longer wants to draw cards, they can stand."),
			),
			auto.MustConfig(
				new(moveRevealHiddenCard),
				moves.WithHelpText("Reveals the hidden card in the user's hand"),
				moves.WithIsFixUp(true),
			),
			auto.MustConfig(
				new(moves.FinishTurn),
				moves.WithHelpText("When the current player has either busted or decided to stand, we advance to next player."),
			),
		),//...
	)
```

Of course, there are sometimes moves that are legal in *any* mode. For those, it still makes sense to use `moves.Add`, as blackjack does:

```go
	return moves.Combine(
		moves.Add(
			auto.MustConfig(
				new(moveShuffleDiscardToDraw),
				moves.WithHelpText("When the draw deck is empty, shuffles the discard deck into draw deck."),
			),
		),
		//...
	)
```

Note that moves.Add() doesn't really do anything; it purely exists so that
it's more legible when you have AddForPhase in the same block.

#### Ordered Moves

This machinery works great for moves that legal at any point within a phase, like in blackjack's `phaseNormalPlay`.

However in many cases, like setting up a new round of a game, there are a series of moves that should be applied in a precise order, one after the other. Writing bespoke `Legal` methods that did complicated signaling to each other about when it was their turn would be very error prone.

For that reason, the Phase machinery also has a notion of *ordered* moves in a Phase. When a phase is configured to require certain moves in a specific order, `moves.Default`'s `Legal()` will return an error if the move is applied in the wrong order. 

This means that instead of writing an error-prone Legal method, in many cases you don't need to write a custom Legal method at all, and can just rely on the phase ordering machinery.

The actual machinery to do this uses what are called MoveProgressions, a notion encoded in the `moves` package. You pass `WithLegalMoveProgression` when configuring the move, and `moves.Default.Legal()` consults that information.

Like setting the legal phases, though, it's extremely error prone to call these yourself. That's why `moves.AddOrderedForPhase` exists, which automatically calls `WithLegalPhases` and `WithLegalMoveProgression` on the moves with the right information.

You can see it in action in Blackjack:

```go
	//...
		moves.AddOrderedForPhase(phaseInitialDeal,
			auto.MustConfig(
				new(moves.DealCountComponents),
				moves.WithMoveName("Deal Initial Hidden Card"),
				moves.WithHelpText("Deals a hidden card to each player"),
				moves.WithGameStack("DrawStack"),
				moves.WithPlayerStack("HiddenHand"),
			),
			auto.MustConfig(
				new(moves.DealCountComponents),
				moves.WithMoveName("Deal Initial Visible Card"),
				moves.WithHelpText("Deals a visible card to each player"),
				moves.WithGameStack("DrawStack"),
				moves.WithPlayerStack("VisibleHand"),
			),
			auto.MustConfig(
				new(moves.StartPhase),
				moves.WithPhaseToStart(phaseNormalPlay, phaseEnum),
			),
		),
	)
```

In most cases when you define a progression of moves that are legal in a given phase, you want each move to only be able to be applied a single time in a row. There are some moves that you want to be able to apply multiple times in a row, until their subclasses' Legal() no longer returns nil. For example, for blackjack we want to keep calling MoveDealInitialHiddenCard until each player has a hidden card dealt to them.

Moves signal this by implementing the `interfaces.AllowMultipleInProgression`, and returning true(). You almost never do this yourself, but instead embed moves that do this behavior for you. The move "Deal Initial Visible Card" and "Deal Initial Hidden Card" are both instances of of `moves.DealCountComponents` which is a type of RoundRobin move, which we'll get to in a second.

One more wrinkle: when the engine looks to see if a propose move is legal in this phase in this order, it will ignore any moves that are legal in all phases that may have come in between. This means that if you have a move like ShuffleDiscardToDraw that triggers in any phase if the discard pile runs out, it won't mess up your move progression matching.

By default move progressions are simple serial lists of moves that must occur in order. But if you have more complex logic you can also define groups with more rich semantics. See the section on MoveProgressionGroup below.

#### StartPhase move

The last move in that section is of type `moves.StartPhase`. It needs to be configured with a `WithPhaseToStart`. Often you don't need to override its Legal or Apply at all (the Legal it inherits from Base is sufficient), and can just use the naked `moves.StartPhase` struct itself without embedding it in your own struct.

It is common for the last move of an ordered round to have a move that advances to the next phase. 

#### Round Robin

Another more complex type of move is `moves.RoundRobin`. RoundRobin moves are moves where the move should be repeatedly proposed until some condition is met. For example, a typical RoundRobin move is to deal a card out to each player, until one has been dealt to each person.

A RoundRobin move defines some end-condition (by default the move has gone around one complete cycle and applied for each player) and an action to apply when each Move is applied. It stores some bookkeeping information in your gameState, and has its DefaultsForState handle advancing to the next target player each time.

RoundRobins are pretty complex under the hood because they can model a number of interesting exit criterion. To use a round robin your gameState must implement `moves/interfaces.RoundRobinProperties`. Typically you just embed `behaviors.RoundRobin` to automatically cover those.

RoundRobin moves are very powerful and general, and the `moves.RoundRobin` documentation goes into
more depth on how to configure and use them. In practice you almost always use two types of moves
that are simple sub-classes of RoundRobin: `moves.DealCountComponents` to deal components from a
gameState to specific players, and `moves.CollectCountComponents` to collect components from each
player into gameState. The moves package describes how these moves work and how they fit together.

#### moves.AutoConfigurer

Again, you almost never generate MoveConfigs yourself, but rather use `moves.AutoConfigurer`. See the package doc of `moves` to learn more about how to use it.

### Phases and TreeEnums

In many cases your game has a straightforward progression of phases, and a
normal Enum (described above) will do. But in other cases, there's a complex
progression of phases, some of which might be nested within one another. For
example, maybe during Normal play the game can enter a special sub-phase where
every other player needs to play cards to try to counter a move the primary
player made.

These sub-phases can be finicky to do, and in many cases it's easiest to model them as a phase in themselves, and rely on normal ordered move phase machinery.

To accomplish this use case (and others), the enum package introduces the
notion of a TreeEnum. A TreeEnum is like a normal Enum, except that it also
encodes information about how the various values parent into one another to
form a tree. You can learn more about how a TreeEnum works in the package doc for the enum package.

The whole library (including the moves sub-package) will interpret phaseEnums
that also happen to be TreeEnums specially. They'll make sure, for example,
that the delegate.CurrentPhase() is never in a non-leaf node phase. Also,
moves.Default().Legal() will interpret a move that applies in a certain phase to
also be legal any time delegate.CurrentPhase returns a value that is a child
of that phase.

### MoveProgressionGroup

When you install ordered moves for a game, the default is that each MoveConfig must be matched in order for the progression to be valid (with moves that return true from AllowMultipleInProgression to match multiple times in a row).

But sometimes you want more complex groupings. For example, maybe a move can apply two to three times in a row, or move A is allowed, then either move B or move C, then move D.

For this you may use MoveProgressionGroup's, many of which are defined in `moves/groups`. `moves.AddOrderedForPhase` accepts either basic single move configs, or groups, and groups can be nested within one another to create complex progression matching logic. See the `moves/groups` documentation for more on how to use them.

AllowMultipleInProgression means that the move inherently knows how to terminate its own progression; a move that is in a Repeat group without AllowMultipleInProgression doesn't know how to terminate itself when it's no longer valid and needs the help of the group it's a part of to do that calculation.

Note that move progression groups match greedily as much as they can. In some cases when you have two groups that abut, where the same type of AllowMultipleInProgression moves are next to each other within different groups, the first one consumes all of them in a row, meaning the second group will never match. In this case you can use moves.NoOp to form a barrier.

### Advanced Sanitization

By default, you sanitize with struct tags on properties that use a group of 'all', 'self', or 'other' (or omit the group name and leave it implied). But it's also possible to do more advanced things with group names.

GameDelegate defines GroupEnum, GroupMembership, and ComputedPlayerGroupMembership. These are override points that allow more complex groups to be defined. You can learn more about how they work by looking at the documentation. But in practice, here are some things to know.

If you want to have sanitization that applies to any non-default groups, then you need to create an enum that lists all of the various groups a given player may be in. If you do the following, `boardagme-util codegen` will handle it correctly for you:

```go
//boardgame:codegen
//The next line tells codegen to combine it into a new enum with other enums that also reference the same named item after the colon. 'group' is the one that base.GameDelegate is configured to use automatically when deciding the GroupMembership of a playerState.
//combine:group
const (
	roleGuesser = iota
	roleClueGiver
)

//boardgame:codegen
//combine:group
const (
	colorRed = iota
	colorBlue
)
```

You could then add behaviors like behavior.PlayerRole and behavior.PlayerColor to your playerState.

Then, in your playerStates you could use sanitization policies like: `guesser:hidden` to hide properties when sanitizing a state for a player who has the Role of roleGuesser.

You can also do more advanced things. For example, `different-color:len` would make it so if a player who is a different color than the player in question is looking at a stack, they'll just see the len. This would allow players on the same "team" to see that stack property for each other, while other players not being able to see them. `same-color` also works similarly, but opposite.

### Seats and Inactive players

The core game logic has no idea which actual user is playing as any given
player; it has no concept of who may legally propose a move on behalf of a given
player at all. The server package is where specific users are mapped to specific
players.

However, in some cases it's important for the game logic to know that there's
not actually a user actually there yet to play on behalf of a given player. For
that reason, there exists behaviors.Seat and moves.SeatPlayer. If your game
logic has moves.SeatPlayer, then the server will propose that move when a new
physical user wants to join the game. By defining when in your rounds that move
is legal, you can define when it will be fired. When you use moves.SeatPlayer,
you should also embed behaviors.Seat in your playerState.

You often don't want a player to be seated and immediately active--for example,
if a player joins in the middle of a round you want to wait until the start of
the next round to deal them in. For that reason there's also a
behaviors.InactivePlayer. If that's embedded, then when a player is seated
they'll immediately be marked as "Inactive", meaning the rest of the game logic
will pretend they aren't there. You then need to choose when to activate those
players, typically by having moves.ActivateInactivePlayers fire.

Typically at the setup phase before a round, you want to activate any inactive
players, pause to wait until we have at least the necessary number of players,
and then inactivate any currently empty seats so we won't wait for those
non-existent players in the round. That's such a common series of moves that you
can use moves.DefaultRoundSetup().

You can see idiomatic use of these concepts in the blackjack example.

See the package doc of the behaviors package for more.

### Gathering

The gathering system builds on seats and inactive players to provide a "lobby"
experience. There is no special lobby mode — a gathering phase is just a normal
phase where gathering-related moves are legal. The client automatically renders
appropriate UI (waiting status, share link, team/role/color pickers, start
button) based on which moves are currently legal.

**Zero-code default:** Any game that uses `moves.DefaultRoundSetup` gets a
gathering panel for free. The client shows "Waiting for Players" and a share
link when the game has empty seats. If you use
`moves.DefaultRoundSetup(auto, moves.WithManualStart())`, a "Start Game" button
also appears.

**Team/Role/Color selection:** To let players pick their team, role, or color
during the gathering phase:

1. Embed the corresponding behavior in your playerState:
   `behaviors.PlayerTeam`, `behaviors.PlayerRole`, or `behaviors.PlayerColor`.

2. Define the enum: `const (teamUnset = iota; teamRed; teamBlue)` with a
   `//boardgame:codegen` annotation. The first value should be an "unset"
   sentinel if you want to detect players who haven't picked yet.

3. Register the selection moves for your gathering phase:
```
moves.AddForPhase(phaseGathering, moves.GatheringMoves(auto)...)
```
   `GatheringMoves` auto-detects which behaviors you've embedded and returns
   the corresponding move configs (SelectTeam, SelectRole, SelectColor). Use
   `AddForPhase` (not `AddOrderedForPhase`) so players can pick freely in any
   order.

4. Optionally validate configuration via `ReadyToStart`:
```
func (g *gameDelegate) ReadyToStart(state boardgame.ImmutableState) error {
    // Return nil when ready, or a descriptive error
    return errors.New("each team needs at least 2 players")
}
```
   This error surfaces in the client as a status message and disables the
   "Start Game" button until configuration is valid.

**Uniqueness:** `SelectColor` enforces uniqueness by default (no two players
can share a color). Use `moves.WithAllowDuplicates()` to disable. `SelectRole`
allows duplicates by default; use `moves.WithUnique()` to require unique
selections (e.g., Spirit Island spirits). Both options work on all three moves.

**Multi-round games:** To reopen gathering between rounds, transition your
cleanup phase back to the gathering phase. `DefaultRoundSetup` will re-activate
players, and the gathering UI will reappear naturally.

**Client overrides:** Game authors can customize the gathering UI:
- CSS: set `--boardgame-gathering-team-picker-display: none` to hide the
  framework's picker and render your own in the game renderer.
- Game renderer: check `this.gatheringActive` to conditionally render
  gathering-specific UI alongside the game board.

See the blackjack example for an idiomatic gathering phase, and the moves
package doc for the full API reference.

### Variants

Games can often have different variations. For example, a deck-based card game might be playable with an expansion pack of cards mixed in. 

These are represented in the engine by the notion of a `Variant` which is just an alias of `map[string]string`. When your game is created, a bundle of Variant will be passed to `NewGame`, along with how many players are in the game. That variant is simply passed to your `GameDelegate`'s `BeginSetUp` method, and that's it. It's your game's responsibility to take that information to set properties differently so the game can be configured that way. (Although you can later retrieve the variant a game was created with with game.Variant()).

If you want to support variants in your game, your delegate should return a VariantConfig from its Variants() method. This config defines what the legal keys and values are, what the defaults are, how those keys and values should be displayed to end users.

Here's memory's:

```go
const (
	variantKeyNumCards = "numcards"
	variantKeyCardSet  = "cardset"
)

const (
	numCardsSmall  = "small"
	numCardsMedium = "medium"
	numCardsLarge  = "large"
)

const (
	cardSetAll     = "all"
	cardSetFoods   = "foods"
	cardSetAnimals = "animals"
	cardSetGeneral = "general"
)

func (g *gameDelegate) Variants() boardgame.VariantConfig {

	return boardgame.VariantConfig{
		variantKeyCardSet: {
			VariantDisplayInfo: boardgame.VariantDisplayInfo{
				DisplayName: "Card Set",
				Description: "Which theme of cards to use",
			},
			Default: cardSetAll,
			Values: map[string]*boardgame.VariantDisplayInfo{
				cardSetAll: {
					DisplayName: "All Cards",
					Description: "All cards mixed together",
				},
				cardSetFoods: {
					Description: "Food cards",
				},
				cardSetAnimals: {
					Description: "Animal cards",
				},
				cardSetGeneral: {
					Description: "Random cards with no particular theme",
				},
			},
		},
		variantKeyNumCards: {
			VariantDisplayInfo: boardgame.VariantDisplayInfo{
				DisplayName: "Number of Cards",
				Description: "How many cards to use? Larger numbers are more difficult.",
			},
			Default: numCardsMedium,
			Values: map[string]*boardgame.VariantDisplayInfo{
				numCardsMedium: {
					Description: "A default difficulty game",
				},
				numCardsSmall: {
					Description: "An easy game",
				},
				numCardsLarge: {
					Description: "A challenging game",
				},
			},
		},
	}
}
```

If a variant is passed to the game that is a key/value set that is not legal, the game will fail to be created.

As you can see, a number of times DisplayNames can be omitted because they can be set automatically by just title-casing the name. See boardgame.VariantConfig (and other related docs) for more.

### Agents

Not all players of a game are human. You also want bots or AIs to be able to play. In the engine these are called *agents*.

Agents are configured on the manager when it is created by returning agents in
your delegate's ConfigureAgents() method. There can be multiple agents,
representing different AIs--although in practice you'll likely only have one.
Agents are set up when the game is set up, and then have a callback called
after every move is made to have a chance to propose a move.

The interface that agents must implement is simple:

```sh
type Agent interface {
    Name() string

    DisplayName() string

    SetUpForGame(game *Game, player PlayerIndex) (agentState []byte)

    ProposeMove(game *Game, player PlayerIndex, agentState []byte) (move Move, newState []byte)
}
```

Name() and DisplayName() are similar to the same fields for Games(). The first is a unique-within-this-game-package name, and the latter is what will actually be displayed to the user.

Agents are given access to a Game to act on, which allows them to see the current state as well as the historical moves. But sometimes that state isn't enough. For example, in memory the agent has to remember what cards have been revealed in the past. That state doesn't make sense to store in the main `gameState` or `playerState`. For that reasons agents are also able to store their own state.

Agents state is just a `[]byte` that the engine will persist and then hand back to the agent whenever it is called. Typically agents will encode their state as JSON and then read it back--but that's up to the agent to do as it wishes. Returning an agentState is optional--if it's nil, no new state will be saved. If no state has been saved at all, this means that future calls will have nil state. If state has previously been saved, it just means that no new state versions will be saved.

Agents' ProposeMove is called after every *causal chain* of moves is done. That is, after each playerMove has been applied *and all of the FixUp moves that result*. This is also the timing when normal players are allowed to make moves.

### Constants

Your `GameDelegate` can define constants by returning a map of constants to values from `ConfigureConstants()`. Constants may be an int, bool, or string.

Of course, you don't need to actually return anything from that method to define normal constants in your package. There are two primary reasons to define them: 1) if you need them client-side, and 2) if you want to use them in a tag-based struct auto-inflater.

Constants that are exported via `ConfigureConstants()` will automatically be transmitted client-side.
`boardgame-util emit-types` also generates an exact `GameConstants` contract
for them and binds it to your generated renderer bases. In a renderer,
`this.chest?.Constants?.TOTAL_DIM` is typed and autocompleted; misspelled or
undeclared keys fail compilation. The framework transport remains generic, but
game creators should use the generated renderer base rather than casting the
constants object or declaring a hand-written index signature.

Constants can also be used as the int argument in a tag-based struct auto-inflation. For example, see the tictactoe example:

```go
//In examples/tictactoe/main.go

func (g *gameDelegate) ConfigureConstants() boardgame.PropertyCollection {
	return boardgame.PropertyCollection{
		"TOTAL_DIM": TOTAL_DIM,
	}
}
```

```go
//In examples/tictactoe/state.go

//boardgame:codegen
type gameState struct {
	base.SubState
	CurrentPlayer boardgame.PlayerIndex
	Slots         boardgame.SizedStack `sizedstack:"tokens,TOTAL_DIM"`
	//... Other fields elided
}

```

That allows you to tie the size of the stack automatically to the constant in use elsewhere in the package. The reason you have to export the constant is because constants are not available in go programs at run-time.

### Setting config properties

Many server and `boardgame-util` commands read from a config.json file.

In this tutorial so far you've implicitly been using the `config.SAMPLE.json` file. But in practice you'll generally want to create your own. You can find the canonical help about how those files are structured by running `boardgame-util help config`.

You can modify the config files directly yourself, but it's more common to use `boardgame-util config set` to set properties directly. The first time you call that command, if there isn't an operative config in your directory or its ancestors, it will create a reasonably-named config file in your current directory.

The description of what the various config fields do is in `boardgame/boardgame-util/README.md`.

When creating a new repo or game, it's strongly encouraged to add the following line to your .gitignore:

```gitignore
*.SECRET.*
```

That helps ensure that you don't accidentally check in secret things into version control, like production database DSNs.

### Ensuring your game is well tested

It's important to save robust tests to ensure your games continue to behave as expected.

`boardgame-util` has a special `create-golden` mode that makes it easy to record game play, generating golden game runs that can then be compared to current behavior of your game in `go test`.

You run that tool from within a game package. It's similar to running `boardgame-util serve`, except instead of using all of the game packages listed in your config, it only uses the package in the current directory. It wires it up so that it uses a storage layer that creates json files for each game and its states, stores them in `testdata/golden`, and also creates a `golden_test.go` that automatically loads up all of the games in that directory and ensures that the current behavior matches.

So the workflow is that every so often, sit in the game package, and run `boardgame-util create-golden`. Then create a few new games that exercise interesting behavior (using admin mode's Current Player view to make moves as all players), and as long as they behave as expected, check them in. Every so often you can run the command and create new ones; the existing ones won't be removed.

It's important that your game is deterministic for the same inputs, so its behavior doesn't change and can be compared to tests. In particular, only ever use state.Rand() for randomness, as its state is seeded deterministically based on the game id and version. In fact, if your game package imports math/rand, the package won't be valid to run with the engine unless you have a comment asserting that it the game logic is still deterministic despite the import.

### Client animations

The client side library automatically handles generating rich animations of
components moving from stack to stack, and generally the default ones are
totally sufficient.

Every state version is shipped down to the client to be rendered. When we
render a state, we wait for any animations it kicked off to finish, then
render the next one.

What this means is that basically every individual move you make is eligible
for animating, if it modifies any items on state that would change rendering
and cause an animation to occur. This means that if you want a certain action
to be distinctly visible on the client, you should ensure that there's an
individual move in which it happens. All actions that occur within one move
will be animated simultaneously.

As a concrete example, if you move all cards from one stack to another with
stack.MoveAllTo(), all of the cards will animate moving at exactly the same
time, which isn't particularly clear, visually. If instead you want each card
to be collected one at a time, you'd use moves.MoveAllComponents, which has a
similar effect but renders each individual card movement separately.

You can modify a number of properties of animations. The most simple is the
`--animation-length` CSS var, which the built-in components respect for how
long all of their animations will take. Sometimes you want all animations for a certain move to take a certain amount of time, and it's confusing/error prone to set the values in CSS. If your game renderer defines `animationLength(fromMove, toMove)` then it will be consulted before each state bundle is installed. If the value is 0, then no override is set and the default CSS values for animation length take precedence. If it is greater than zero, than a temporary `--animation-length` value will be set above your renderer (interpreting that number as millisecondes), overriding the default value until another one is set. And if the value is negative, the animation will be skipped entirely. `BoardgameBaseGameRenderer` provides a default `animationLength` that just returns 0.

Both arguments are `ClientMove | null`. `ClientMove` intentionally contains
only readonly `Name` and `Version`: enough to select an animation policy, but
never the storage record's serialized arguments, proposer, initiator, phase, or
timestamp. Those fields can contain private choices that state sanitization
correctly hid. Move type names are public catalog metadata, so do not encode a
secret choice into dynamically generated move names; put the choice in typed
move fields and authoritative sanitized state as usual.

Sometimes you want the completed state to remain visible for a beat before the
next state is installed. For example, Memory leaves a matching pair face-up so
players can recognize it before the cards are captured. Put
`post-animation-delay="1000"` on the relevant `boardgame-component-stack` (or
animatable item). The stack forwards the value to its components; their WAAPI
animations hold the completion gate for that many milliseconds after visible
motion. Prefer deriving the value from current rendered state, as Memory's
`_revealHoldMs()` does. There is no renderer `delayAnimation` hook.

The way the game logic is defined on the server specifies the maximally separate chunking of renderering. However, sometimes you don't want all of those chunks and want to combine some. For example, maybe the user has turned on a 'Fast Animations' option in your game renderer, and instead of animating each card one at a time going from one stack to another, you want all of the cards to move simultaneously. You configure this behavior via `animationLength`, described in the paragraphs above. Instead of returning a positive or 0 length however, you return any negative number to signify that that state should be skipped and the next one should be installed instead. (Note that the last bunlde in the queue is always installed).

Sometimes you want animations to overlap rather than playing fully sequentially. For example, when dealing cards to players, you might want the next card to start moving before the previous one has finished. If your game renderer defines `animationOverlap(fromMove, toMove)`, it will be consulted before each state bundle is installed. The return value is a fraction between 0 and 1 representing how much of the current animation should play before the next state is installed. A return value of 0 (the default) means the current animation must complete entirely before the next state is applied. A value of 0.5 means the next state will be installed when the current animation is 50% complete. Values outside the 0-1 range are clamped. This is useful for cascade effects where multiple animations should overlap smoothly instead of playing one after another.

These controls compose cleanly: `animationLength` controls motion duration (or
skips an intermediate bundle with a negative value), `post-animation-delay`
holds a component's completed state, and `animationOverlap` lets a solo cycle
install its next state before the current cycle finishes.

Companion Table/Hand surfaces add one deliberate constraint: animation cycles
that must agree across physical screens use the framework's version timeline.
The current protocol gives each version an 800ms slot—at most 600ms of motion
plus 200ms to render and pre-arm the next state. For these synchronized cycles,
the framework budgets each component's stagger, visible duration, and
`post-animation-delay` together inside the remaining 600ms motion window. An
effect whose stagger would begin after that window is omitted; an oversized
hold shortens visible motion rather than delaying later slots. Ordinary FLIP
and property effects use the same policy as `animateBetween`, and
`animationOverlap` is disabled. Solo games and explicitly local effects retain
the normal behavior above.

Game renderers normally need no timing code. A call such as
`this.animator?.animateBetween(card, source, 300)` automatically uses the
installed version's companion slot. For an effect that exists only on this
screen, say so explicitly:

```ts
this.animator?.animateBetween(card, source, 900, { timing: 'immediate' });
```

Custom animatable components use the identical policy through `play()`:

```ts
this.play(this, keyframes, { duration: 300 }); // current version slot
this.play(this, keyframes, { duration: 300 }, { timing: 'immediate' });
```

Advanced test or orchestration code can instead pass
`{ timing: { localStartAtMs: timestamp } }`. See
`docs/companion-mode-authoring.md` for the complete Table/Hand conventions.

In the future there will be a number of other attributes and method override
points, and they'll be described here.

For a more thorough overview of how the animation system actually works, check
out `server/static/src/ARCHITECTURE.md`.

### Creating a more production-ready server

Check out the "Creating your own game" section above before reading this section.

The default server in the tutorial uses the bolt db backend because it doesn't
require much configuration. But in practice you'll probably want a mysql
backend.

So far we've used `boardgame-util serve` to run a server. What that command does is effectively `boardgame-util build api` and `boardgame-util build static`, to generate a simple server binary and also generate a linked folder of all of the necessary static HTML files to render the client. `boardgame-util serve` does that in a temporary folder that it then discards when the command is quit. But you can run those other commands directly to generate the server. There's nothing special about these commands; you could manually wire up your own server with the game packages on your own if you wanted.

Each server binary has a specific storage backend it uses. `boardgame-util build api` and `boardgame-util serve` by default use the DefaultStorageType configuration property to select that although an argument of `--storage=TYPE` overrides that. 

The `config.SAMPLE.json` in `github.com/jkomoros/boardgame` (the config you've implicitly been using in this tutorial) sets the default type to `bolt`, but in production or in any real development you'll probably want to use mysql. That requires you to set up your own mysql server and make sure your config file knows how to connect to it. The `config.SAMPLE.json` has a reasonable config string for local development, but your DSN for the production environment will likely be more complicated. See `storage/mysql/README.md` for more about the structure of that connetion string.

`boardgame-util db` and its subcommands can help you configure and set up your database correctly. After starting the mysql server (and ensuring that connection strings are set correctly in your config), run `boardgame-util db setup` to set up the initial configuration. In the future, to ensure your database is fully migrated, you can run `boardgame-util db up`.

### Conclusion

This library is a passion project I'm pursuing in my free time. It's under active development. If you see something that seems to be missing or off, please reach out via a GitHub issue. And pull requests are very appreciated!
