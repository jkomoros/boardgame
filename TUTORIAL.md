# boardgame

boardgame is a work-in-progress package that aspires to make it easy to define multi-player boardgames that can be easily hosted in a high-quality web app with minimal configuration. It is under active development as a hobby project and different components of it vary in their completeness and polish.

The core of your game logic is constructed using the core library into a *game manager* for each game. The server package makes it easy to take those game managers and install them into a server instance. Each game manager defines a basic view that knows how to render any given state of one of its game for the user.

A number of example games are defined in the examples sub-package to demonstrate how to use many of the key concepts. Real documentation for the core game engine is in the [godoc package docs](https://godoc.org/github.com/jkomoros/boardgame).

## Tutorial

*This tutorial will walk through some concrete examples of how to configure a server and create games, in a way that narratively makes sense but leaves a number of topics unexplored or lightly developed. For more in-depth documentation of the core concepts, check out the core library's package doc, and for more about the server, see `server/README.md`*

Each instantitation of a server includes multiple game packages, each of which defines a Game Manager that describes the logic necessary to run that type of game. These game packages are organized in a canonical way to make it easy to link in game packages into your server even if you didn't write them.

An example server can be found in `examples/server`. This tutorial will walk through how those work.

## Quickstart servers

The server has two components: the static asset hosting and the core game engine API server. The former is mainly used in the development enviornment (in a production environment all of the static assets can be served by Firebase hosting). The API server is easy to set-up; normally it requires only 20 lines of set up, most of which is configuration for MySQL and other aspects.

For simplicity, the server in examples/server is configured to be able to get up and running with no changes in configuration. (It uses a simpler storage backend that doesn't require MySQL).

In `boardgame/examples/server/static`, run

```
go build && ./static
```

In `boardgame/examples/server/api`, run

```
go build && ./api
```

Now you can visit the web app in your browser by navigating to `localhost:8080`

## Game Managers

Now that you have the server set up, let's dig into how a given game is constructed.

We'll dig into `examples/memory` because it covers many of the core concepts. The memory game is the classic childhood game where there's a deck of cards of symbols, with exactly two cards for each symbol. The cards are arrayed face down on the table and players take turn flipping over two cards. If they get a match, they get to keep the cards.

At the core of every game is the `GameManager`. This is an object that encapsulates all of the logic about a game and can be installed into a server. The `GameManager` is a struct provided by the core package, but each game type will configure its behavior to encapsulate its logic.

Each game type, fundamentally, is about representing all of the semantics of a given game state in a versioned **State** and then configuring when and how modifications may be made by defining **Moves**.

### State

The state is the complete encapsulation of all semantically relevant information for your game at any point. Every time a move is succesfully applied, a new state is created, with a version number one greater than the previous current state. States may only be modified by applying moves.

Game states are represented by a handful of structs specific to your game type. All of these structs are composed only of certain types of simple properties, which are enumerated in `boardgame.PropertyType`. The two most common structs for your game are `GameState` and `PlayerState`.

`GameState` represents all of the state of the game that is not specific to any player. For example, this is where you might capture who the current player is, and the Draw and Discard decks for a game of cards.

`PlayerState`s represent the state specific to each individual player in the game. For example, this is where each player's current score would be encoded, and also which cards they have in their hand.

Let's dig into concrete examples in memory, in `examples/memory/state.go`.

The core of the states are represented here:

```
//+autoreader
type gameState struct {
	boardgame.BaseSubState
	CardSet        string
	NumCards       int
	CurrentPlayer  boardgame.PlayerIndex
	HiddenCards    boardgame.MutableStack `sizedstack:"cards,40" sanitize:"order"`
	VisibleCards   boardgame.MutableStack `sizedstack:"cards,40"`
	Cards          boardgame.Stack        `overlap:"VisibleCards,HiddenCards"`
	HideCardsTimer boardgame.MutableTimer
	//Where cards not in use reside most of the time
	UnusedCards boardgame.MutableStack `stack:"cards"`
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex       boardgame.PlayerIndex
	CardsLeftToReveal int
	WonCards          boardgame.MutableStack `stack:"cards"`
}
```

There's a lot going on here, so we'll unpack it piece by piece.

At the core you can see that these objects are simple structs with (mostly) public properties. The game engine will marshal your objects to JSON and back often, so it's important that the properties be public.

It's not explicitly listed, but the only (public) properties on these objects are ones that are
legal according to `boardgame.PropertyType`. Your GameManager would fail to be created if your state structs included illegal property types.

Note the first anonymous field of `boardgame.BaseSubState`. This is a simple struct designed to be anonymously embedded in the substates you define that implements the SetState method that SubStates must define. It's technically optional, but you'll normally just want to anonymously embed it in your gameState and playerStates.

Most of the properties are straightforward. Each player has how many cards they are still allowed to draw this turn, for example.

#### Stacks and Components

As you can see, stacks of cards are represented by type `MutableStack`.

Stacks contain 0 or more **Components**. Components are anything in a game that can move around: cards, meeples, resource tokens, dice, etc. Each game type defines a complete enumeration of all components included in their game in something called a **ComponentChest**. We'll get back to that later in the tutorial.

By default Stacks can grow to accomodate new components and have no empty spaces in the middle. Adding a new component to a slot in the middle of a stack would simply push components from there onward down a slot, and grow the stack by one.

A SizedStack is a special kind of Stack that has a fixed number of slots, each of which may be empty or contain a single component. The default growable Stacks are useful in most instances, including representing a player's Hand or a Draw or Discard deck. SizedStacks are useful when there's a specific fixed size or where there might be gaps between components.

Each component is organized into exactly one **Deck**. A deck is a collection of components all of the same type. For example, you might have a deck of playing cards, a deck of meeples, and a deck of dice in a game. (The terminology makes most sense for cards, but applies to any group of components in a game.) The ComponentChest is simply an enumeration of all of the Decks for this game type. Memory has only has a single deck of cards, but other games will have significantly more decks.

Each Stack is associated with exactly one deck, and only components that are members of that deck may be inserted into that stack. The deck is the complete enumeration of all components in a given set within the game. In memory you can see struct tags that associate a given stack with a given deck. We'll get into how that works later in the tutorial.

**Each component must be in precisely one stack in every state**. This reflects the notion that components are phsyical objects that are in only one location at any given time, and must exist *somewhere*. Later we will see how the methods available on stacks to move around components help enforce that invariant.

When a memory game starts, most of the cards will be in GameState.HiddenCards. Players can also have cards in a stack in their hand when they win them, in WonCards. You'll note that there are actually three stacks for cards in GameState: HiddenCards, VisibleCards, and Cards. We'll get into why that is later.

#### autoreader

Both of the State objects also have a cryptic comment above them: `//+autoreader`. These are actually a critical concept to understand about the core engine.

In a number of cases (including your GameState and PlayerState), your specific game package provides the structs to operate on. The core engine doesn't know their shape. In a number of cases, however, it is necessary to interact with specific fields of that struct, or enumerate how many of a certain type of property there are. It's possible to do that via reflection, but that would be slow. In addition, the engine requires that your structs be simple and only have known types of properties, but if general reflection were used it would be harder to detect that.

The core package has a notion of a `PropertyReader` (as well as `PropertyReadSetter` and `PropertyReadSetConfigurer`), which makes it possible to enumerate, read, and set properties on these types of objects. The signature looks something like this:

```
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
	MutableEnumProp(name string) (enum.MutableVal, error)
	MutableStackProp(name string) (MutableStack, error)
	MutableTimerProp(name string) (MutableTimer, error)

	PropMutable(name string) bool

	SetProp(name string, value interface{}) error
}

type PropertyReadSetConfigurer interface {
	PropertyReadSetter

	ConfigureMutableEnumProp(name string, value enum.MutableVal) error
	ConfigureMutableStackProp(name string, value MutableStack) error
	ConfigureMutableTimerProp(name string, value MutableTimer) error

    ConfigureEnumProp(name string, value enum.Val) error
    ConfigureStackProp(name string, value Stack) error
    ConfigureTimerProp(name string, value Timer) error

	ConfigureProp(name string, value interface{}) error
}
```

This known signature is used a lot within the package for the engine to interact with objects specific to a given game type.

For simple types (like bools, ints, and strings) the signature is
straightforward: a getter and a setter. However, there are three types of
supported properties that are special: `Stack`, `Enum`, and `Timer`. These three types are called "Interface types" because they are a container with some configuration, as well as the specific values within that container. The base interface has read-only methods, and the `MutableTYPE` interface also includes mutators.

A generic Setter for those properties doesn't make sense in a
`PropertyReadSetter` because the configuration of the property doesn't change,
only its value within the container. For that reason the Setters are missing and instead have a MutableTYPE getter, which allows mutation, and also have a ConfigureMutableTYPEProp setters, which are used only after the object is freshly-minted in order to configure the container.

Implementing all of those getters and setters for each custom object type you have is a complete pain. That's why there's a command, suitable for use with `go generate`, that automatically creates PropertyReaders for your structs.

First, install the command by running `go install` from within `$GOPATH/github.com/jkomoros/boardgame/cmd/autoreader`. You only need to do this once.

Somewhere in the package, include:

```
//go:generate autoreader
```

(In the memory package you'll find it near the top of `examples/memory/main.go`)

And then immediately before every struct you want to have a PropertyReader for, include the magic comment:

```
//+autoreader
type MyStruct struct {
	//....
}
```

Then, every time you change the shape of one of your objects, run `go generate` on the command line. That will create `autoreader.go`, with generated getters and setters for all of your objects.

One other thing to note: the actual concrete structs that you define, like `gameState` and `playerState`, should almost always include the Mutable variant of an interface type (`MutableStack`, `MutableEnum`, and `MutableTimer`); the PropertyReader methods will return just the read-only subset of those objects. In general the whole point of having a state object is to represent the state that *changes* which is why you generally want the Mutable variant. However, there are couple of cases where you might want the immutable variant: when you have read-only properties on a component, or when you're using Merged Stacks, which are inherently read-only (more on that later). But for the most part just always use the Mutable variants in your state objects.

The game engine generally reasons about States as one concrete object made up of one GameState, and **n** PlayerStates (one for each player). (There are other components of State that we'll get into later.) The `State` object is defined in the core package, and the getters for Game and Player states return things that generically implement the interface, although under the covers they are the concrete type specific to your game type. Many of the methods you implement will accept a State object. Of course, it would be a total pain if you had to interact with all of your objects within your own package that way--to say nothing of losing a lot of type safety.

That's why it's convention for each game package to define the following private method in their package:

```
func concreteStates(state boardgame.State) (*gameState, []*playerState) {
	game := state.GameState().(*gameState)

	players := make([]*playerState, len(state.PlayerStates()))

	for i, player := range state.PlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}
```

Whenever the game engine hands you a state object, this one-liner will hand you back the concrete states specific to your game type:

```
func (g *gameDelegate) Diagram(state boardgame.State) string {
	game, players := concreteStates(state)
	//do something with game and players, since they are now the concrete types defined in this package
}
```

Many of the methods you'll implement will be passed `State` objects. In the vast majority of cases you are not allowed to modify the State object. To help make the intention clear, you will be passed either a `State` or `MutableState` object (the latter embedding the `State` interface and adding mutation methods) to make the expectation clear.

... Of course, when you pass the State or MutableState object through your concreteStates method you'll just get the naked, modifiable, concrete structs back, and there's nothing to prevent you from changing the properties. Don't do that--at best it won't actually make a change that will be persisted, but at worse it could lead to odd inconsitencies later, if the engine for example re-used the same state object.

#### PlayerIndex

gameState has a property named `CurrentPlayer` of type `boardgame.PlayerIndex`. This property, as you might expect, encodes whose turn it currently is.

It would be reasonable to encode that bit of state as a simple `int` (and indeed, that's basically what a `PlayerIndex` property is). However, it's so common to have to encode a `PlayerIndex` (for example, if there's a move to attack another player), and there are enough convenience methods that apply, that the core engine defines the type as a fundamental type.

`PlayerIndex`es make it easy to increment the `PlayerIndex` to the next player (wrapping around at the end). The engine also won't let you save a State with a `PlayerIndex` that is set to an invalid value.

`PlayerIndex`es have two special values: the `AdminPlayerIndex` and the `ObserverPlayerIndex`. The AdminPlayerIndex encodes the special omnsicient, all-powerful player who can do everything. Special moves like FixUp Moves (more on those below) are applied by the AdminPlayerIndex. In dev mode it's also possible to turn on Admin mode in the UI, which allows you to make moves on behalf of any player. The ObserverPlayerIndex encodes a run-of-the-mill observer: someone who can only see public state (more on public and private state later) and is not allowed to make any moves.

#### Timer

The last type of property in the states for Memory is the HideCardsTimer, which is of type `*boardgame.Timer`. Timers aren't used in most types of games. After a certain amount of time has passed they automatically propose a move. For Memory the timer is used to ensure that the cards that are revealed are re-hidden within 3 seconds by the player who flipped them--and if not, flip them back over automatically.

Timers are rare because they represent parts of the game logic where the time is semantic to the rules of the game. In memory, for example, if players could leave revealed cards showing indefinitely the game would drag on as players competed to exhaustively commit the location of each card to their memory. Contrast that with animations, where the time that passes is merely presentational, to allow the state changes to be visibly demonstrated to players.

### GameDelegate

OK, so we've defined our state objects. How do we tell the engine to actually use them?

The answer to that, and many other questions, is the `GameDelegate`. The `GameManager` is a concrete type of object in the main engine, with many methods and fields. But there are lots of instances where your game type needs to customize the precise behavior. The answer is to define the logic in your `GameDelegate` object. The GameManager will consult your GameDelegate at key points to see if there is behavior it should do specially.

The most basic methods are about the name of your gametype:

```
type GameDelegate interface {
	Name() string
	DisplayName() string
	Description() string
	//... many more methods follow
}
```

Those methods are how you configure the name of the type of the game (e.g. 'memory' or 'blackjack', or 'pig'), what the game type should be called when presented to users (e.g. "Memory", "Blackjack", or "Pig"), and a short description of the game (e.g. "A card game where players draw cards trying to get as close to 21 as possible without going over")

The GameDelegate interface is long and complex. In many cases you only need to override a handful out of the tens of methods. That's why the core engine provides a `DefaultGameDelegate` struct that has default stubs for each of the methods a `GameDelegate` must implement. That way you can embed a `DefaultGameDelegate` in your concrete GameDelegate and only implement the methods where you need special behavior.

Most of the methods on GameDelegate are straightforward, like `LegalNumPlayers(num int) bool` which is consulted when a game is created to ensure that it includes a legal number of players.

GameDelegates are also where you have "Constructors" for your core concrete types:

```
type GameDelegate interface {
	//...
	GameStateConstructor() ConfigurableSubState
	PlayerStateConstructor(player PlayerIndex) ConfigurablePlayerState
	//...
}
```

ConfigurableSubState and ConfigurablePlayerState are simple interfaces that primarily define how to get a `PropertyReader`, `PropertyReadSetter`, and `PropertyReadSetConfigurer` from the object. Many other sub-state values that we'll encounter later have the same shape, which is why the name is generic.

GameStateConstructor and PlayerStateConstructor should return zero-value objects of your concrete types. The only thing that differentiates GameStates (of type ConfigurableSubState) and PlayerStates (of type ConfigurablePlayerState) is that PlayerStates should come back with a hidden property encoding which PlayerIndex they are.

In many cases they can just be a single line or two, as you can see for the PlayerStateConstructor in main.go:

```
func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {

	return &playerState{
		playerIndex: playerIndex,
	}
}
```
If you look at the GameState constructor, it is even simpler:

```
func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}
```

This is actually very interesting. As mentioned above, Interface properties (like Stacks, Timers, and Enums) need to have their container initalized to a reasonable starting state. For stacks this includes what deck they should be affiliated with, whether they should be a fixed size, and their starting size. For these interface types the zero-value is effectively missing type information.

One way to do that is to initalize them to a reasonable value in the GameStateConstructor:

```
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

```
//+autoreader
type gameState struct {
	//...
	HiddenCards    boardgame.MutableStack `sizedstack:"cards,40" sanitize:"order"`
	VisibleCards  boardgame.MutableStack `sizedstack:"cards,40"`
	UnusedCards    boardgame.MutableStack `stack:"cards"`
	//...
}

//+autoreader
type playerState struct {
	//...
	WonCards          boardgame.MutableStack `stack:"cards"`
}
```

For stacks, you can provide a struct tag that has the name of the deck it's affiliated with. Then you can return a nil value from your constructor for that property, and the system will automatically instantiate a zero-value stack of that shape. (Even cooler, this uses reflection only a single time, at engine start up, so it's fast in normal usage) It's also possible to include the starting size (for default stacks, the max size, and for sized stacks the number of slots).

The vast majority of real-world usecases you'll encounter can just use struct tags.

#### Other GameDelegate methods

The GameDelegate has a number of other important methods to override.

One of them is `CheckGameFinished`, which is run after every Move is applied. In it you should check whether the state of the game denotes a game that is finished, and if it is finished, which players (if any) are winners. This allows you to express situations like draws and ties.

Memory's `CheckGameFinished` could look like this:

```
func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []boardgame.PlayerIndex) {

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

However, this pattern--check if the game is finished, and if it is return as a winner any player who has the highest score--is so common that the engine makes it easy to implement with a default behavior built into `DefaultGameDelegate`. Memory uses it, as you can see in `examples/memory/main.go`:

```
func (g *gameDelegate) GameEndConditionMet(state boardgame.State) bool {
	game, _ := concreteStates(state)

	if game.Cards.NumComponents() > 0 {
		return false
	}

	return true
}

func (g *gameDelegate) PlayerScore(pState boardgame.PlayerState) int {
	player := pState.(*playerState)

	return player.WonCards.NumComponents()
}

```

Implementing these two methods is sufficient for DefaultGameDelegate's default CheckGameFinished to do the right thing.

After `CheckGameFinished` returns true, the game is over and no more moves may be applied.

Another method is `CurrentPlayerIndex`. This method should inspect the provided state and return the `PlayerIndex` corresponding to the current player. If any player may make a move, you should return `AdminPlayerIndex`, and if no player may make a move, you should return `ObserverPlayerIndex`. This method is consulted for various convenience methods elsewhere. The reason it can't be done fully automatically is because different games might store this value in a differently-named field, have non obvious rules for when it changes (for example, return the value in this field in the first phase of the game, but a value in another field in the second phase of the game), or not have a notion of current player at all.

The convention is simply to store this value in a property on your gameState called `CurrentPlayer`. If you do that, DefaultGameDelegate's `CurrentPlayerIndex` will just return that.

GameDelegate has a number of other methods that are consulted at various key points and drive certain behaviors. Each is documented to describe what they do. In a number of cases the default implementations in `DefaultGameDelegate` do complex behaviors that are almost always the correct thing, but can theoretically be overriden if necessary. `SanitizationPolicy` is a great example. We'll get to what it does in just a little bit, but although the method is quite generic, `DefaultGameDelegate`'s implementation encodes the formal convention of using struct-based tags to configure its behavior.

#### SetUp

Once you have a GameManager, you can create individual games from it by calling `NewGame`.

Before a Game may be used it must be `SetUp` by passing the number of players in the game. This is where the game's state is initalized and made ready for the first moves to be applied. `SetUp` may fail for any number of reasons. For example, if the provided number of players is not legal according to the `GameDelegate`'s `LegalNumPlayers` method, `SetUp` will fail.

The initalization of the state object is handled in three phases that can be customized by the `GameDelegate`: `BeginSetup`, `DistributeComponentToStarterStack` and `FinishSetup`.

`BeginSetup` is called first. It provides the State, which will be everything's zero-value (as returned from the Constructors, with minimal fixup and sanitization applied by the engine). This is the chance to do any modifications to the state before components are distributed.

`DistributeComponentToStarterStack` is called repeatedly, once per Compoonent in the game. This is the opportunity to distribute each component to the stack that it will reside in. After this phase is completed, components can only be moved around by calling `SwapComponents`, `MoveComponent`, or `Shuffle` (or their variants). This is how the invariant that each component must reside in precisely one stack at every state version is maintained. Each time that `DistributeComponentToStarterStack` is called, your method should return a reference to the `Stack` that they should be inserted into. If no stack is returned, or if there isn't room in that stack, then the Game's `SetUp` will return an error. Components in this phase are always put into the next space in the stack from front to back. If you desire a different ordering you will fix it up in `FinishSetup`.

`FinishSetup` is the last configurable phase of setting up a game. This is the phase after all components are distributed to their starter stacks. This is where stacks will traditionally be `Shuffle`d or otherwise have their components put into the correct order.

After a game is succesfully `SetUp` it is ready to have Moves applied.

### Moves

Up until this point games have existed as a static snapshot of a given state. Outside of the `SetUp` routines, the only modifications to state must be made by `Move`s. 

The bulk of the logic for your game type will be defined as Move structs and then configured onto your GameManager.

The two most important parts of Moves are the methods `Legal` and `Apply`. When a move is proposed on a game, first its `Legal` method will be called. If it returns an error, the move will be rejected. If it returns `nil`, then `Apply` will be called, which should modify the state according to the semantics and configuration of the move. If `Apply` does not return an error, and if the modified state is legal (for example, if all `PlayerIndex` properties are within legal bounds), then the state will be persisted to the database, the `Version` of the game will be incremented, and the game will be ready for the next move.

Moves are proposed on a game by calling `ProposeMove` and providing the Move, along with which player it is being proposed on behalf of. (The server package keeps track of which user corresponds to which player; more on that later.) The moves are appended to a queue. One at a time the engine will remove the first move in the queue, see if it is Legal for the current state, and if it is will Apply it, as described above.

#### Moves, MoveTypes, and MoveTypeConfigs

There are three types of objects related to Moves: `MoveType`, `MoveTypeConfig`, and `Move`s.

A `Move` is a specific instantiation of a particular type of Move. It is a concrete struct that you define and that adheres to the `Move` interface:

```
type Move interface {
    Legal(state State, proposer PlayerIndex) error
    Apply(state MutableState) error
    //... Other minor methods follow
}
```

Your moves also must implement the `PropertyReader` interface. Some moves contain no extra fields, but many will encode things like which player the move operates on, and also things like which slot from a stack the player drew the card from. Moves also implement a method called `DefaultsForState` which is provided a state and sets the properties on the Move to reasonable values. For example, a common pattern is for a move to have a property that encodes which player the move should operate on; this is generally set to the `CurrentPlayerIndex` for the given state via `DefaultsForState`.

A `MoveType` is a conceptual type of Move that can be made in a game and is a generic struct in the main package. It vends new concrete Moves of this type via `MoveConstructor` and also has metadata specific to all moves of this type, like what the name of the move is. All of a MoveType's fields and methods return constants except for `MoveConstructor`. Given a Move `m`, the Type it is from can be recovered by inspecting `m.Info().Type()`

A `MoveTypeConfig` is a configuration object used to create a `MoveType` when you are setting up your `GameManager` to receive a fully formed and ready-to-use `MoveType`. It is mostly a collection of static properties, plus a `MoveConstructor` that returns the specific concrete struct that implements `Move` for that `MoveType`. More on that later. In practice, you almost never create your own `MoveTypeConfig`, but rather use `auto.Config()` to generate them automatically for you. More on that later, too.

#### Player and FixUp Moves

Conceptually there are two types of Moves: Player Moves, and FixUp moves. Player moves are any moves that are legal for normal players to propose at some point in the game. FixUp moves are special moves that are never legal for players to propose, and are instead useful for fixing up a state to ensure it is valid. For example, a common type of FixUp move examines if the DrawStack is empty, and if so moves all cards from the DiscardStack to the DrawStack and then shuffles it. In practice the only thing that distinguishes FixUp moves is that their `MoveType.IsFixUp()` returns true.

After each move is succesfully applied via ProposeMove, and before the next move in the queue of moves is considered, the engine checks if any FixUp moves should be applied. It does this by consulting the `ProposeFixUpMove` method on the GameDelegate. If that method returns a move, it will be immediately applied, so long as it is legal. This will continue until `ProposeFixUpMove` returns nil, at which point the next player move in the proposed move queue will be considered.

Technically it is possible to override the behavior of exactly when to apply certain FixUp moves. Realistically, however, the behavior of `ProposeFixUpMove` on `DefaultGameDelegate` is almost always sufficient. It simply runs through each move configured on the gametype in order, skipping any for whom `Info().Type().IsFixUp()` returns false, setting its values by calling DefaultsForState, and then checking if it is `Legal`. It returns the first fix up move it finds that is legal. This means that it is **important to make sure that your FixUp moves always have well-constructed `Legal` methods**. If a given FixUp move always returns Legal for some state, then the engine will get in an infinite loop. (Technically the engine will detect that it is in an unreasonable state and will panic.)

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

##### moves.Base

Base is the simplest possible base move. It implements stubs for every required method, with the exception of `Apply` which you must implement yourself. This allows you to minimize the boilerplate you have to implement for simple moves. Almost every move you make will embed this move type either directly or indirectly.

Base doesn't do much except implement the required stubs. The one exception is its `Legal()` method, which is where much of the notion of Phases is implemented. More on that in a later section. For now it's important to know that if you embed a move anonymously in your own move struct, it's very important to always call your "super"'s Legal method as well, because non-trivial logic is encoded in it in Base.

Another simple type of move is `FixUp`. It's a simple embedding of `Base`, but
if your move is a FixUp move it's best to embed it so that `auto.Config` will
treat it as a FixUp move automatically.

##### moves.CurrentPlayer

Many Player moves can only be made by the CurrentPlayer. This move encodes which player the move applies to (set automatically in `DefaultsForState`) and also includes the logic to verify that the `proposer` of the move is allowed to make the move, and is modifiying their own state. (This logic is slightly tricky because it needs to accomodate `AdminPlayerIndex` making moves on behalf of any player).

In typical use you embed this struct, and then check its Legal method at the top of your own Legal method, as in this example from memory:
```
type MoveRevealCard struct {
    moves.CurrentPlayer
    CardIndex int
}

func (m *MoveRevealCard) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

    if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
        return err
    }

    // Logic specific to this move type goes here.
}
```

Similarly, note that if you have your own logic in `DefaultsForState`, you should not forget to also call the embedded `DefaultsForState`.

##### moves.FinishTurn

Another common pattern is to have a FixUp move that inspects the state to see if the current player's turn is done, and if it is, advances to the next player and resets their properties for turn start.

`moves.FinishTurn` defines two interafaces that your sub-state objects must implement:

```
type CurrentPlayerSetter interface {
    SetCurrentPlayer(currentPlayer boardgame.PlayerIndex)
}
```

must be implemented by your gameState. Generally this is as simple as setting the CurrentPlayer index to that value, as you can see in the example from memory:

```
func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
    g.CurrentPlayer = currentPlayer
}
```

The next interface must be implemented by your playerStates:

```
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

```
func (p *playerState) TurnDone(state boardgame.State) error {
    if p.CardsLeftToReveal > 0 {
        return errors.New("they still have cards left to reveal")
    }

    game, _ := concreteStates(state)

    if game.VisibleCards.NumComponents() > 0 {
        return errors.New("there are still some cards revealed, which they must hide")
    }

    return nil
}

func (p *playerState) ResetForTurnStart(state boardgame.State) error {
    p.CardsLeftToReveal = 2
    return nil
}

func (p *playerState) ResetForTurnEnd(state boardgame.State) error {
    return nil
}
```

As you can see from the way the errors are constructed in `TurnDone`, the error message will be included in a larger error message. In practice it will return messages like "The current player is not done with their turn because they still have cards left to reveal". 

Because most of the logic for moves that embed `moves.FinishTurn` lives in methods on gameState and playerState, it's common to not need to override the `Legal` or `Apply` methods on `moves.FinishTurn` at all. You can see this in practice on memory's `MoveFinishTurn` which simply embeds `moves.FinishTurn`.

##### Other move types

moves.Base, moves.CurrentPlayer, and moves.FinsihTurn are only three types of moves defined in the moves package. There are a number of others that are useful in other contexts. More detail about how to use some of them is covered below in the Phases section.

#### auto.Config()

The next section will walk through a fully manually example where you define your own MoveTypeConfig and configure that on your game, before showing how to instead do it with `auto.Config()`. In practice `auto.Config()` (in the `moves/auto` package) is almost always used to automatically generate a MoveTypeConfig based on a move, minimizing boilerplate you have to write. You can learn more about how to use it, and good idioms to follow for defining and installing moves, in the `moves/auto` package doc.

#### Worked Move Example

Let's look at a fully-worked example of defining a specific move from memory:

```
//+autoreader readsetter
type MoveHideCards struct {
    moves.CurrentPlayer
}
```

MoveHideCards is a simple concrete struct that embeds a `moves.CurrentPlayer`. This means it is a move that may only be made by the player who turn it is.

MoveHideCards is decorated by the magic autoreader comment, which means its ReadSetter will be automatically generated. The `readsetter` at the end of the comment tells `autoreader` to only bother creating the `PropertyReadSetter` method and not worry about the `PropertyReader` method. It would work fine (just with a tiny bit more code generated) with that argument omitted.

```
var moveHideCardsConfig = boardgame.MoveTypeConfig{
    Name:     "Hide Cards",
    HelpText: "After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player.",
    MoveConstructor: func() boardgame.Move {
        return new(MoveHideCards)
    },
}
```

This is the moveTypeConfig object. This is what we will actually use to install the move type in the GameManager (more on that later).

A `MoveTypeConfig` is basically a bag of straightforward properties. The reason you don't define a MoveType yourself is because it's important that these properties not change once they are configured onto a GameManager. You can think of a MoveTypeConfig as basically just the starter values for properties that will be read-only on the actual MoveType. Again, in practice you almost never generate them yourself, but instead use `auto.Config()`

The `Name` property is a unique-within-this-game-package, human-readable name for the move. It is the string that will be used to retrieve this move type from within the game manager. (You'll rarely do this yourself, but the server package will do this for example to deserialize `POST`s that propose a move).

`HelpText` is a short descriptive stirng that describes generically what this move type accomplishes. It is similar to the `Description` method on `Move`, except that `Description` should include information about the specific properties of this particular instantiation of the move, while `HelpText` is generic. In fact, `moves.Base`'s `Description` method defaults to just returning the `HelpText` for the movetype.

If this move type were a fixUp move, your MoveTypeConfig would also set `IsFixUp` to true.

The most important aspect of `MoveType` is the `MoveConstructor`. Similar to other Constructor methods, this is where your concrete type that implements the interface from the core library will be returned. In almost every case this is a single line method that just `new`'s your concrete Move struct. If you use properties whose zero-value isn't legal (like Enums, which we haven't encountered yet in the tutorial), then as long as you use struct tags, the engine will automatically instantiate them for you, similar to how `GameStateConstructor` works.

```
func (m *MoveHideCards) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

    if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
        return err
    }

    game, players := concreteStates(state)

    p := players[game.CurrentPlayer]

    if p.CardsLeftToReveal > 0 {
        return errors.New("You still have to reveal more cards before your turn is over")
    }

    if game.VisibleCards.NumComponents() < 1 {
        return errors.New("No cards left to hide!")
    }

    return nil
}
```

This is our Legal method. We embed `moves.CurrentPlayer`, but add on our own logic. That's why we call `m.CurrentPlayer.Legal` first, since we want to extend our "superclass". In general you should always call the Legal method of your super class, as even moves.Base includes important logic in its Legal implementation.

```
func (m *MoveHideCards) Apply(state boardgame.MutableState) error {
    game, _ := concreteStates(state)

    //Cancel a timer in case it was still going.
    game.HideCardsTimer.Cancel()

    for i, c := range game.VisibleCards.Components() {
        if c != nil {
            if err := game.VisibleCards.MoveComponent(i, game.HiddenCards, i); err != nil {
                return errors.New("Couldn't move component: " + err.Error())
            }
        }
    }

    return nil
}
```

This is our Apply method. There's not much interesting going on--except to note that calling MoveComponent can fail (for example, if the stack we're moving to is already max size), so we check for that and return an error. If your Move's `Apply` method returns an error than the move will not be applied. In general it is best practice in `Legal` to check for any condition that could cause your `Apply` to fail, so that failures in `Apply` are truly unexpected. But as this example shows, sometimes that's more of a pain than it's worth, as long as you catch those errors in `Apply`. If you didn't catch them in either `Legal` or `Apply` then you could start persisting illegal states to the storage layer, which would get really confusing really fast.

### NewManager

We've now explored enough concepts to build a game. The last remaining piece is to combine everything into a ready-to-use `GameManager` that we can then pass to a server or use in other contexts.

By convention, each game package has a `NewManager` method with the following signature:

```
func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) 
```

The storage argument is a reference to the specific storage backend the manager should use to persist objects. This will generally be provided by the server that your game manager factory is plugged into. A number of different storage implementations can be found in the `storage` directory. In most cases you'll use the MySQL storage layer, but that can be a pain to configure so in some cases, like the simple no-nonsense example server listed in the quickstart above a no-config storage layer like `bolt` is used instead. But in any case your `NewManager` method will generally just pass that directly to the core engine's `NewGameManager` method without inspecting it.

In general your `NewManager` method will define all of the immutable configuration specific to your game. This includes defining all of the components for the game in a ComponentChest, plugging in your `GameDelegate`, and installing all of your MoveTypes.

#### Component structs

Remember that each component is immutable, and lives in precisely one deck in the `ComponentChest` for a game type. Specific instantiations of a Game of this GameType will ensure that each component in the chest lives in exactly one position in one stack at every version. Since the component is immutable, each game's version's stacks have pointers to the same shared components across all games that come from that gametype.

The `Component` struct is a concrete struct defined in the core package. It is immutable, and includes a reference to the deck this component is in, what its index is within that stack, and the `Values` of this Component--the specific properties of this particular component within this game's semantics.

For example, a component that is a card from a traditional American deck of playing cards would have two properties in its Values object; `Rank` and `Suit`. (In fact, American playing cards are so common that for convenience a ready-to-use version of them are defined in `components/playingcards`). The `Values` object will be a concrete struct that you define in your package that adheres to the `SubState` interface. This mean--you guessed it--that the `autoreader` package will be useful.

The components for memory are quite simple:

```
var generalCards []string = []string{
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

//+autoreader reader
type cardValue struct {
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

And then the concrete struct we will use for `Values` is a trivial struct with a single string property, and the `autoreader` magic comment.

In more complicated games, your components and their related constants might be much, much more verbose and effectively be a transcription of the values of a large deck of cards.

#### ConfigureMoves

Your GameDelegate implements a method called `ConfigureMoves()
*boardgame.MoveTypeConfigBundle`. This method will be called during the
creation process for a GameManager. Your job is to create a new move bundle,
add moves to it for your game, and return it. This is the main way you add
moves to your game type.

An example that could be for memory is here:

```
//Not what memory actually does
func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {
	return boardgame.NewMoveTypeConfigBundle().AddMoves(
		//moveRevealCardConfig and others would be defined in the same file as the move structs they are associated with.
		&moveRevealCardConfig,
		&moveHideCardsConfig,
		&moveFinishTurnConfig,
		&moveCaptureCardsConfig,
		&moveStartHideCardsTimerConfig,
	)
}
```

In practice, however, memory uses `auto.Config()`--just as almost every game will--to automatically generate MoveTypeConfigs:

```
func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {

	//...some lines elided...

	return boardgame.NewMoveTypeConfigBundle().AddMoves(
		//...one moveType elided...
		auto.MustConfig(
			new(MoveHideCards),
			moves.WithHelpText("After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player."),
		),
		auto.MustConfig(
			new(moves.FinishTurn),
		),
		auto.MustConfig(
			new(MoveCaptureCards),
			moves.WithHelpText("If two cards are showing and they are the same type, capture them to the current player's hand."),
		),
		auto.MustConfig(
			new(MoveStartHideCardsTimer),
			moves.WithHelpText("If two cards are showing and they are not the same type and the timer is not active, start a timer to automatically hide them."),
		),
	)
}
```

`auto.Config()` is a very powerful tool. It automatically generates move constructors, helptext, and even move names (based on the name of the struct). In this case, you can see that we didn't even need to create a `MoveFinishTurn` in our package--we could simply use `moves.FinishTurn` directly.

You can learn much more about how to use `auto.Config()` in the package doc for `moves/auto`.

More complicated games would use more advanced methods on the bundle (see the
section in the More Concepts section about Phases). Note that the various
AddMove methods all return a reference to the bundle itself, which allows you
to chain the calls and keep the method short.


#### Worked NewManager example

Let's look at memory's NewManager implementation:

```
func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
	chest := boardgame.NewComponentChest(nil)

	if err := chest.AddDeck(cardsDeckName, newDeck()); err != nil {
		return nil, errors.New("Couldn't add deck: " + err.Error())
	}

	return boardgame.NewGameManager(&gameDelegate{}, chest, storage)
}
```

First, we create a new empty `ComponentChest`. Then we start defining the single deck of cards. We create an empty deck, then for each constant in our cardNames we insert two components into the deck with those values. It is important that we always insert the exact same components in the exact same order. Stacks encode which components they contain with a deck/index pair--which means that if the order of the deck is different from when it was saved, the stack will have nonsensical values.

Then we add the deck to the chest. The logic to actually create the deck is included in newDeck().

Now we have the three things we need to get a manager object: the delegate, the chest we just created, and the storage manager that we were passed in. We simply return the result of calling `boardgame.NewGameManager` with those arguments.

`NewGameManager` will fail with an error if anything is not set up correctly:
if the moves are configured incorrectly for any reason, for example, or that
the GameStateConstructor and PlayerStateConstructor return reasonable values.

By following this convention, it will be very easy for instantiations of a server to easily include this game type with minimal overhead.

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

```
//+autoreader
type gameState struct {
	boardgame.BaseSubState
	CardSet        string
	NumCards       int
	CurrentPlayer  boardgame.PlayerIndex
	HiddenCards    boardgame.MutableStack `sizedstack:"cards,40" sanitize:"order"`
	VisibleCards   boardgame.MutableStack `sizedstack:"cards,40"`
	Cards          boardgame.Stack        `overlap:"VisibleCards,HiddenCards"`
	HideCardsTimer boardgame.MutableTimer
	//Where cards not in use reside most of the time
	UnusedCards boardgame.MutableStack `stack:"cards"`
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex       boardgame.PlayerIndex
	CardsLeftToReveal int
	WonCards          boardgame.MutableStack `stack:"cards"`
}
```

HiddenCards is the only stack that is sanitized; everything else is fully visible.

Now that we know about sanitization, we can finally understand why there are three stacks in game: `HiddenCards`, `VisibleCards`, and `Cards`. 

##### Aside: Merged Stacks

Each stack must be sanitized the same way--if the components are hidden, then **all** of the components are hidden. But in memory, there are cards that are hidden and cards that are revealed in the same area. 

The way we do it is by **merging** two stacks together, so they can be used logically as one read-only stack, both server and client-side. There are two types of merged stacks, and they're both created in a similar way. ``NewOveralappedStack`` returns an overlapped stack, and `NewConcatenatedStack` returns a concatenated stack. An overlapped stack takes the first stack provided and returns those components--unless that slot is empty, in which case whatever is in that location of the second slot is returned. For overlapped stacks, both stacks must be fixed size, and they both must be the same size. Concatenated stacks simply have all of the slots of the first stack followed by all of the slots of the second stack.

We can use tag-based auto-inflation for merged stacks, too. We use either `concatenate` or `overlap` and then pass the property names of the input stacks. Note that because Merged Stacks are fundamentally read only, they must be stored in an immutable stack property in your state object. (One of the rare cases where you want a `Stack` property but not a `MutableStack`.) Note that to use tag-based auto inflation the properties must be in the same object. If you want to combine two stacks in different SubStates, you can return them as a Computed Property instead (see the section below on computed properties).

When you use merged stacks, the convention is to name the hidden stack `HiddenFoo`, the visible stack `VisibleFoo`, and the merged stack that combines them just `Foo`.

That's not a *particularly* interesting example. Here's the states for blackjack:

```
//+autoreader
type gameState struct {
	moveinterfaces.RoundRobinBaseGameState
	Phase         enum.MutableVal        `enum:"Phase"`
	DiscardStack  boardgame.MutableStack `stack:"cards" sanitize:"len"`
	DrawStack     boardgame.MutableStack `stack:"cards" sanitize:"len"`
	UnusedCards   boardgame.MutableStack `stack:"cards"`
	CurrentPlayer boardgame.PlayerIndex
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
	HiddenHand  boardgame.MutableStack `stack:"cards,1" sanitize:"len"`
	VisibleHand boardgame.MutableStack `stack:"cards"`
	Hand        boardgame.Stack        `concatenate:"HiddenHand,VisibleHand"`
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

The renderer is a web component with a known name and defined in a known
location that will be instantiated and passed the state object. This is the
primary client-side object that you should define. Your renderer will be passed
four attributes:

* **State**, which is the state for the current version, with many properties expanded for
convenience. This state object will contain all computed properties, for each
Stack will have the DynamicValues for the component added as a direct property
of the component, and will have the computed TimeRemaining provided on the
timer, continuously updated as time passes.
* **Diagram**, which is the result of your GameDelegate's Diagram() method for this state. It's
provided primarily as a useful fallback.
* **viewingAsPlayer**, which is the index of the player who is viewing the game. This might be -1 if
the viewer is a generic observer who isn't themselves playing the game, or -2 if
the player is the all-powerful Admin.
* **currentPlayerIndex**, the index of the player whose turn it is, according to your GameDelegate's
CurrentPlayerIndex method.

The job of your renderer is to take those attributes, render a meaningful visual
representation, and emit events of type `propose-move` when a player has
proposed a specific move that should be passed to the server and proposed. In
practice many renderers look quite similar and basically just define where to
stamp out components.

#### location of renderers

The renderer must be in a specific, known location so it can be imported.

Your renderer web component should be named `boardgame-render-game-GAMENAME`,
where `GAMENAME` is the name of your game (what your GameDelegate returns from
the Name() method).

The import will be looked for in `../game-src/GAMENAME/boardgame-render-game-
GAMENAME.html`.

Your game type might be imported into many different servers, so it's best
practice to keep the renderer definition near the package defining your server
code.

The idiotmatic way to do this is, within the package that defines your game
type's go code, have a sub-folder structure, as you can see by looking at
memory:

```
memory/
	|	client/
	|	|	memory/
	|	|	| boardgame-render-game-memory.html
	|	|	| boardgame-render-player-info-memory.html
	|	agent.go
	|	agent_test.go
	|	auto_reader.go
	|	components.go
	|	main.go
	|	main_test.go
	|	moves.go
	|	state.go
```

(We'll get to what `boardgame-render-player-info-memory.html` in just a bit).

Then, when you're configuring your server, you create a `game-src/` folder in
the right location and then take advantage of the game repo and your server repo
both having a canonical location in your `$GOPATH` to create a relative link for
that directory into `game-src/`, something like `ln -s ../../../../memory/client/memory`.
You can then check that symlink into git, and be reasonably certain that it will work in other places.

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
stack`, when used in conjunction idiotmatically, almost always do exactly what
you want using idiomatic CSS layout with things like flexbox and grid to lay them out and then, with minimal configuration, have high-quality, performant animations created.
Their implementation is non-trivial and handles many edge cases and conditions that are not immediately obvious. They use the `Id` machinery briefly described in the Sanitization section above to keep track of which cards--even cards that are hidden--are which in between states and then animate the cards moving from stack to stack appropriately. They even handle cases like cards flipping from visible to hidden--if done naively, the content of the card would disappear immediately before the flip animation plays! In general, it is strongly recommended to use these components.

boardgame-cards are the basic cards. You can instantiate yourself and set their various properties,
but in practice it is best to bind their `item` attribute to each component item in the state.

boardgame-card's size can be affected by two css properties: --component-scale (a float, with 1.0 being default size) and --card-aspect-ratio (a float, defaulting to 0.6666). Cards are always 100px width by default, with scale affecting the amount of space they take up physically in the layout, as well as applying a transform to their contents to get them to be the right size. --card-aspect-ratio changes how long the minor-axis is compared to the first. If the scale and aspect-ratio are set based on the position in the layout, the size will animate smoothly.

It can be finicky to set all of the cards correctly for the animation to work as
you want; the easiest way is to set boardgame-card-stack's stack property to the
stack in the state, and then ensure you have a template for that deck defined in a `<boardgame-deck-defaults>` element.

In many cases you only have a small number of types of cards in a game, and you want to define their layout only once if possible for consitency. The way to do this is to use the `boardgame-deck-defaults` element in your renderer's template and include a template for your deck.

```
<!-- define a simple front if no processing required -->
<boardgame-deck-defaults>
  <template deck="cards">
    <boardgame-card>
      <div>
        {{item.Values.Type}}
      </div>
    </boardgame-card>
  </template>
</boardgame-deck-defaults>
<!-- boardgame-component-stacks that print from the deck `cards` will automatically stamp that item -->
```

Inside of the template for the deck, include the most general thing to stamp. In general, this is just a `boardgame-card` or `boardgame-token`, perhaps with some inner content. Within that inner content you can bind `item` or `index`. 

Then stamping those components is as simple as using a `boardgame-component-stack` and databinding in the stack property:
```
<boardgame-component-stack layout="stack" stack="{{state.Players.0.WonCards}}" messy component-disabled>
</boardgame-component-stack>
```

The `boardgame-component-stack` will automatically instantiate and bind components as defined in the defaults for that deck name.

Any properties on the `boardgame-stack` of form `component-my-prop` will have `my-prop` stamped on each component that's created. That allows different stacks to, for example, have their components rotated or not. If you want a given attribute to be bound to each component's index in the array, add it in the special attribute `component-index-attributes`, like so:

```
<boardgame-component-stack layout="grid" messy stack="{{state.Game.Cards}}" component-propose-move="Reveal Card" component-index-attributes="data-arg-card-index">
</boardgame-component-stack>
```

If you wanted to do more complex processing, you can create your own custom element and bind that in the same pattern:

```
<link rel='import' href='my-complex-card.html'>
<boardgame-deck-defaults>
  <template deck="cards">
    <boardgame-card>
    	<my-complex-card item="{{item}}"></my-complex-card>
    </boardgame-card>
  </template>
</boardgame-deck-defaults>
```

##### boardgame-fading-text

In many cases you want to draw attention to values that change as the result of moves. For example, when it's the current player's turn you might want to make that fact obvious. A common way to do that is to have that text expand from that location and fade as it does so, drawing attention to the changed value. `boardgame-fading-text` will do this for you.

The boardgame-fading-text element will render text that animates when changed. The font size can be changed with `--message-font-size`. The text will be centered in the nearest ancestor positoned block. When the animation is over the text will be invisible. This is great for animating messages like "Your Turn" that play centered in the middle of your view when it's the user's turn. There are different policies you can apply to control how this text triggers and what text it shows, see the component documenation for more.

In many cases there are parts of your UI that show a value in them, and when that value changes you want to draw attention to it. For example, if you have some text that shows the number of cards in a given stack, you might want users to notice when that changes.

You can use `boardgame-status-text` to render text that will automatically show the fading effect if the value changes. It uses the 'diff-up' strategy by default for fading text, which can be overriden.

```
<!-- you can bind to message attribute -->
<boardgame-status-text message="{{state.Game.Cards.Components.length}}"></boardgame-status-text>

<!-- you can also just include content which automatically sets message -->
<boardgame-status-text>{{state.Game.Cards.Components.length}}</boardgame-status-text>

```

##### boardgame-base-game-renderer

`boardgame-base-game-renderer` is a superclass that it generally makes sense for your renderer to subclass. The primary thing it adds (currently) is machinery to propose moves based on mark-up.

In particular, if an interface element is tapped that has a `propose-move="MOVENAME"`, then it will automatically dispatch a move to the engine to propose that move. You can also define keys/values to be packaged up with the move as attributes in the format `data-arg-my-arg$="val"`, which will result in the ProposeMove event having an arguments bundle of `{MyArg:"val"}`.

#### Worked Example

In general your renderer is mostly concerned with telling the data-binding system where and how to stamp out stacks and buttons. This is one reasons Computed Properties (see the "Other Important Concepts" section below) are useful, because they allow you to define your semantic logic almost entirely on the server and allow the client to be almost entirely about data-binding.

Here's the data-binding for Memory:
```
    <boardgame-deck-defaults>
      <template deck="cards">
        <boardgame-card>
          <div>
            {{item.Values.Type}}
          </div>
        </boardgame-card>
      </template>
    </boardgame-deck-defaults>
    <h2>Memory</h2>
    <div>
      <boardgame-component-stack layout="grid" messy stack="{{state.Game.Cards}}" component-propose-move="Reveal Card" component-index-attributes="data-arg-card-index">
      </boardgame-component-stack>
       <boardgame-fading-text message="Match" trigger="{{state.Game.Cards.NumComponents}}"></boardgame-fading-text>
    </div>
    <div class="layout horizontal around-justified discards">
      <boardgame-component-stack layout="stack" stack="{{state.Players.0.WonCards}}" messy component-disabled>
      </boardgame-component-stack>
      <!-- have a boardgame-card spacer just to keep that row height sane even with no cards -->
      <boardgame-card spacer></boardgame-card>
      <boardgame-component-stack layout="stack" messy stack="{{state.Players.1.WonCards}}" component-disabled>
      </boardgame-component-stack>
    </div>
    <paper-button id="hide" propose-move="Hide Cards" raised disabled="{{state.Computed.Global.CurrentPlayerHasCardsToReveal}}">Hide Cards</paper-button>
    <paper-progress id="timeleft" value="{{state.Game.HideCardsTimer.TimeLeft}}" max="{{maxTimeLeft}}"></paper-progress>
    <boardgame-fading-text trigger="{{isCurrentPlayer}}" message="Your Turn" suppress="falsey"></boardgame-fading-text>
```

It looks like a lot, but it's mostly just abouts stamping out stacks.

#### Player-info

The web app also has a bar along the top of the game that lists each player, their picture, their name, and their player index. It also by default shows whether it's their turn (according to your delegate's `CurrentPlayerIndex`).

You can override this behavior, and also add more information to be rendered for each player (like their current score), by implementing a `boardgame-render-player-info-GAMETYPE` element. If that component exists, it will be passed the full state, as well as the playerState for the specific player. Any text you render out will be shown in the info section beneath each player.

Your player-info renderer can also expose a chipColor and chipText property to override the text of the badge on each player (by default their player index) and what color it is.

memory's player-info just prints out the current score:
```
  <template>
    Won Cards <boardgame-status-text>{{playerState.WonCards.Indexes.length}}</boardgame-status-text>
  </template>
```

The tictactoe example shows how to override the badge/chip color and text.

## Other important concepts

The sections above cover the information you almost always need to know to build a game from start to finish. However, there are other, slightly more complex features and concepts that are optional but sometimes useful for specific types of games. They're described in separate sections below.

### Dynamic Component Values

By default Components are entirely fixed--their values are exactly the same in every game. That works well for things like cards, but isn't sufficiently general. As a simple example, it's not possible to model a Die, because a die has a fixed set of sides that are the same for all games, but also a specific face that is currently face-up. As a much more complex example, the game Evolution allows players to have any number of Species cards in front of them, each with a population size, a body size, consumed food, and up to 4 trait cards.

These use cases are represented by the concept of *Dynamic Component Values*. For decks that have dynamic component values, the values will be stored as an extra section in your State, just like `gameState` and your `playerState`s. On the server, given a state and a component c, you can access the dynamic component values like so:

```
values := c.DynamicValues(state)
```

On the client, these dynamic component values will be merged in directly on the component objects in the state passed to your renderer.

If you look at the JSON output of a state, you'll see that dynamic component values are stored in a section called "Components", with a key for each deck name that has DynamicComponentValues, and then a slot for values associated with each component in that deck. component.DynamicValues is then just a convenience method that fetches the right component values associated with this component.

The way you configure that a given deck has dynamic component values is by the output to `GameDelegate.DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState`. For decks that don't have dynamic values, just return nil. For decks that do have dynamic component values, just return a new concrete struct, just as you would for `GameStateConstructor` and `PlayerStateConstructor`.

When sanitizing dynamic component values, each deck has its own policy. Importantly, though, that policy is only effective if the stack that the component is currently in has a policy of Visible. In most cases it should just behave as you'd naively expect. For more about specific behaviors, see the package doc on Sanitization.

### Computed properties

It's common to define methods on your `gameState` and `playerState` objects to modify the states and also to provide getters for values that can be computed entirely based on the values of specific properties. This works great on the server, but sometimes you want to have those same computed values available on the client in order to do view data-binding more easily.

When a JSON representation of your gameState is being prepared for a player, your delegate's `ComputedGlobalProperties(state State)` and `ComputedPlayerProperties(player PlayerState)` are called, allowing you to return a map of strings to `interface{}` to include in the JSON. 

Typically this is a simple enumeration of the names of the values and the method calls, like you can see in memory:

```
func (g *gameDelegate) ComputedGlobalProperties(state boardgame.State) boardgame.PropertyCollection {
	game, _ := concreteStates(state)
	return boardgame.PropertyCollection{
		"CurrentPlayerHasCardsToReveal": game.CurrentPlayerHasCardsToReveal(),
	}
}
```

Note that when this method is called, your state will likely aready have been sanitized, which means that **your computed property methods should return reasonable values for sanitized states**. In most cases you don't have to think much about this, because all sanitization transformations keep the objects of the same "shape". But it is something to keep an eye out for.

Note that although Merged Stacks might *feel* like computed properties, in most cases (as long as the stacks are on the same SubState object), you can simply use tag-based auto-inflation and have the merged stacks live directly on your state objects.

### Enums

There are a number of cases where a given property can be one of a small set of options--what you'd call in other languages an Enum.

Representing those values as an int is OK, but it doesn't allow you to enumerate which values are legal. In addition, you sometimes want to know the string value of the enum value in question.

Boardgame formalizes this notion as an `enum`, which is a valid property type and is defined in `boardgame/enum`. 

You define your named Enums at set up time as part of an `EnumSet`, and list the values that are legal (and their string equivalents). You can retrieve the EnumSet in use from `manager.Chest().Enums()`.

Given an enum, you can create an `enum.Val`, which is a container for a value from that enum. These `enum.Val` and `enum.MutableVal` are legal properties to add to your states and moves, and like stacks can be configured via struct tags, as you can see in blackjack's `state.go`:

```
//+autoreader
type gameState struct {
	moveinterfaces.RoundRobinBaseGameState
	Phase         enum.MutableVal        `enum:"Phase"`
	DiscardStack  boardgame.MutableStack `stack:"cards" sanitize:"len"`
	DrawStack     boardgame.MutableStack `stack:"cards" sanitize:"len"`
	UnusedCards   boardgame.MutableStack `stack:"cards"`
	CurrentPlayer boardgame.PlayerIndex
}
```

Creating an enum is slightly cumbersome and repetitive. You typically create a const block, enumerate all of the values, and then later install each of those values, while passing their string equivalent.

The autoreader command can also help automate this, as you can see in the blackjack example in `state.go`:

```
//+autoreader
const (
	PhaseInitialDeal = iota
	PhaseNormalPlay
)
```

This will automatically create a global `Enums` EnumSet, and a global `PhaseEnum` that contains the two values, configured with the string values of "Initial Deal" and "Normal Play". You can find much more details on the conventions and how to configure autoreader in the enums package doc.

### RangedEnum and Enum Graphs

Sometimes when you're creating a boardgame--especially one with a board and multiple connected spaces--you need to keep track of which spaces are connected to one another.

The enum package also allows you to create a ranged enum. It's just a normal enum, but created with all of the values in the given dimensions:

```
//returns an enum with 9 items
e := set.MustAddRanged("Spaces", 3, 3)

//returns true
e.IsRange()
```

Under the covers it's just a simple enum with values from 0 to 8, where the string value for 0 is "0,0". But because it was created with AddRange is also has a few additional convience getters to and from the raw index to the multi-dimensional index it represents.

```
//Returns []int{0,1}
e.ValueToRange(3)

//returns 3
e.RangeToValue(0, 1)
```

Typically to model a board with spaces, you create a RangedEnum of the correct dimensions. Then on your gameState you'd have a SizedStack that is the same size as the RangedEnum. You'd use the Ranged getters to convert a multi-dimensional index into a single-dimensional index into the stack. This set-up works if each space on the board can have only one token; if a given space can host more than one, create a Spaces SizedStack for each player.

```
chessBoard := set.MustAddRange("Spaces", 8, 8)

type gameState struct {
	boargame.BaseSubState
	//Note: 64 is the size of the chessBoard.
	Spaces boargame.MutableStack `sizedstack:"Tokens, 64"`
}

//retrive the token at space 3,3 in the chessboard
gState.Spaces.ComponentAt(chessBoard.RangeToValue(3,3))
```

`enum/graph` is a package that allows you to create graphs where each value in an enum is a node, and you add edges between nodes. These graphs are useful to test whether indexes in a stack that represents spaces in a game board are adjacent or not.

You can add your own edges between items, but for grid-based boards, NewGridConnectedness() often does what you want. Check out the package doc for more, but here's a quick example:

```
set := enum.NewSet()
chessBoard := set.MustAddRange("Spaces", 8, 8)

//blackLegalMoves will have moves that are only valid upwards and diagonal.
blackLegalMoves := graph.NewGridConnectedness(chessBoard, DirectionDiagonal, DirectionUp)
redLegalMoves := graph.NewGridConnectedness(chessBoard, DirectionDiagonal, DirectionDown)
```

### Phases

At the core of the engine, there's just a big collection of moves, any of which may be `Legal()` at any time. `ProposeFixUpMove` often just cycles through all FixUp moves in order and returns the first one that is legal.

This is fine for simple games like memory, but quickly becomes cumbersome for more complicated games. For example, some games have multiple rounds, where each round is basically a mini-game, where scores accumulate across rounds. For each round you might have to do some set-up tasks (like moving all of the cards from discard to the draw stack, shuffling them, and then dealing out two cards per player), then have the normal play, and then finally some clean-up tasks (collecting the cards remaining in players' hands, tallying up scores).

If you had to write all of your Legal() methods by hand, it would be error-prone and finicky. You'd have to think carefully about how each move could look at the state of the game and figure out that it was its time to be applied. In many cases, it wouldn't be possible to tell that cleanly, and you'd have to add lots of extra properties to your State object to keep track of exactly where you were and what needed to be done.

It'd be a mess!

For that reason, a convention of "Phases" is used. A game can have multiple phases. Moves are only legal to apply in certain phases. In some phases, moves are applied in a specific, prescribed order only.

The concept of Phases is only lightly represented in the core game engine, and is mostly implemented "in user land" by careful convention and default methods.

At the core, the notion of Phases is implmented by `moves.Base`'s Legal method--which is why it's so important to always call your super's `Legal` method! `moves.Base` will first check to make sure that the current phase of the game is one that is legal for this move, and then check to see if playing this move at this point in the phase is legal. All other methods and machinery for representing Phases are just about giving moves.Base the information it needs to make this determination.

The actual machinery to implement Moves is not important, other than to know that it can be overriden by swapping out the implementations of a few delegate methods, as covered in the package documentation. This part of the tutorial will primarily just discuss how to use it in practice by examining the blackjack example.

If you're going to support the notion of phases, you'll need to store the current phase somewhere in your state. In `examples/blackjack/state.go` we have:

```
//+autoreader
type gameState struct {
	moveinterfaces.RoundRobinBaseGameState
	Phase         enum.MutableVal        `enum:"Phase"`
	DiscardStack  boardgame.MutableStack `stack:"cards" sanitize:"len"`
	DrawStack     boardgame.MutableStack `stack:"cards" sanitize:"len"`
	UnusedCards   boardgame.MutableStack `stack:"cards"`
	CurrentPlayer boardgame.PlayerIndex
}
```

We also need to define the values of the enum. In `examples/blackjack/components.go` we have:

```
//+autoreader
const (
	PhaseSetUp = iota
	PhaseNormalPlay
	PhaseScoring
)
```

In general it's easiest to use autoreader's enum-generation tool, which we do here.

It's convention to name your phase enum as "Phase", and `moves.Base` will rely on that in some cases to create meaningful error messages. If you want to name it something different, override `GameDelegate.PhaseEnum`.

Now we have to tell the engine what the current phase is. We do this by overriding a method on our gamedelegate, much like we do for CurrentPlayerIndex:
```
func (g *gameDelegate) CurrentPhase(state boardgame.State) int {
	game, _ := concreteStates(state)
	return game.Phase.Value()
}
```

However, since we're using DefaultGameDelegate and our Phase property is `Phase` on our `gameState`,
we don't even have to do that. DefaultGameDelegate's CurrentPhase() already looks for that value
there and returns it.

Now the core engine knows about what phase it is. `moves.Base` will consult that information it is Legal method. But how do we tell `moves.Base` which phases a move is legal in?

`MoveType` and `MoveTypeConfig` have a `LegalPhases []int` field. If this field is zero-length (nil or a zero-length slice) that tells `moves.Base` that the move is legal in any phase. If it's non-zero-length, then `moves.Base` will ensure that the current phase is explicitly enumerated in LegalPhases, otherwise it will return an error.

You can set these `LegalPhases` on your own in `MoveTypeConfig`s, but that can be error prone, and it's confusing to have the information about phases coordinated across each move, instead of in one central place.

`GameManager` has an `AddMoves` method that takes a list of move configs to add to the manager during setup. It also has an `AddMovesForPhase` method that takes a phase and a list of moves. For each move config provided, if the LegalPhases doesn't already list the phase in question, it will explicitly add it before installing it. This means that in most cases you can just leave your `LegalPhases` field on your config as nil, and configure the phase for all moves at once when you install them. 

You can see this in action in `examples/blackjack/main.go` in `ConfigureMoves`

```
	//...
	.AddMovesForPhase(PhaseNormalPlay,
		auto.MustConfig(
			new(MoveCurrentPlayerHit),
			moves.WithHelpText("The current player hits, drawing a card."),
		),
		auto.MustConfig(
			new(MoveCurrentPlayerStand),
			moves.WithHelpText("If the current player no longer wants to draw cards, they can stand."),
		),
		auto.MustConfig(
			new(MoveRevealHiddenCard),
			moves.WithHelpText("Reveals the hidden card in the user's hand"),
			moves.WithIsFixUp(true),
		),
		auto.MustConfig(
			new(moves.FinishTurn),
			moves.WithHelpText("When the current player has either busted or decided to stand, we advance to next player."),
		),
	)//...
```

If you were to inspect the MoveTypeConfig that `auto.Config()` returned for MoveCurrentPlayerHit, you'll see it doesn't mention LegalPhases at all, allowing manager.AddMovesForPhase to configure them appropriately. If you did want to specifically define which phases it was legal in, you'd pass `moves.WithLegalPhases(phases)` into `auto.Config()`.

Of course, there are sometimes moves that are legal in *any* mode. For those, it still makes sense to use `AddMoves`, as blackjack does:

```
	return boardgame.NewMoveTypeConfigBundle().AddMoves(
		auto.MustConfig(
			new(MoveShuffleDiscardToDraw),
			moves.WithHelpText("When the draw deck is empty, shuffles the discard deck into draw deck."),
		),
	)...
```

#### Ordered Moves

This machinery works great for moves that legal at any point within a phase, like in blackjack's `PhaseNormalPlay`.

However in many cases, like setting up a new round of a game, there are a series of moves that should be applied in a precise order, one after the other. Writing bespoke `Legal` methods that did complicated signaling to each other about when it was their turn would be very error prone.

For that reason, the Phase machinery also has a notion of *ordered* moves in a Phase. When a phase is configured to require certain moves in a specific order, `moves.Base`'s `Legal()` will return an error if the move is applied in the wrong order. 

This means that instead of writing an error-prone Legal method, in many cases you don't need to write a custom Legal method at all, and can just rely on the phase ordering machinery.

The precise machinery that accomplishes this is covered in `moves.Base` and GameDelegate's `PhaseMoveProgression() []string`. But in practice you rarely need to modify those, and can just use `GameManager.AddOrderedMovesForPhase` to accomplish what you want, as you can see in blackjack's `NewManager`:

```
	//...
	).AddOrderedMovesForPhase(PhaseInitialDeal,
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
			moves.WithPhaseToStart(PhaseNormalPlay, PhaseEnum),
		),
	)
```

In most cases when you define a progression of moves that are legal in a given phase, you want each move to only be able to be applied a single time in a row. There are some moves that you want to be able to apply multiple times in a row, until their subclasses' Legal() no longer returns nil. For example, for blackjack we want to keep calling MoveDealInitialHiddenCard until each player has a hidden card dealt to them.

Moves signal this by implementing the `interfaces.AllowMultipleInProgression`, and returning true(). You almost never do this yourself, but instead embed moves that do this behavior for you. The move "Deal Initial Visible Card" and "Deal Initial Hidden Card" are both instances of of `moves.DealCountComponents` which is a type of RoundRobin move, which we'll get to in a second.

One more wrinkle: when the engine looks to see if a propose move is legal in this phase in this order, it will ignore any moves that are legal in all phases that may have come in between. This means that if you have a move like ShuffleDiscardToDraw that triggers in any phase if the discard pile runs out, it won't mess up your move progression matching.

#### StartPhase move

The last move in that section is of type `moves.StartPhase`. It needs to be configured with a `moves.WithPhaseToStart`. Often you don't need to override its Legal or Apply at all (the Legal it inherits from Base is sufficient), and can just use the naked `moves.StartPhase` struct itself without embedding it in your own struct.

It is common for the last move of an ordered round to have a move that advances to the next phase. 

#### Round Robin

Another more complex type of move is `moves.RoundRobin`. RoundRobin moves are moves where the move should be repeatedly proposed until some condition is met. For example, a typical RoundRobin move is to deal a card out to each player, until one has been dealt to each person.

A RoundRobin move defines some end-condition (by default the move has gone around one complete cycle and applied for each player) and an action to apply when each Move is applied. It stores some bookkeeping information in your gameState, and has its DefaultsForState handle advancing to the next target player each time.

RoundRobins are pretty complex under the hood because they can model a number of interesting exit criterion. To use a round robin your gameState must implement `moveinterfaces.RoundRobinProperties`. Alternatively you can anonymously embed `moveinterfaces.RoundRobinBaseGameState` instead of `boardgame.BaseSubState` to implement it for free. 

RoundRobin moves are very powerful and general, and the `moves.RoundRobin` documentation goes into
more depth on how to configure and use them. In practice you almost always use two types of moves
that are simple sub-classes of RoundRobin: `moves.DealCountComponents` to deal components from a
gameState to specific players, and `moves.CollectCountComponents` to collect components from each
player into gameState. The moves package describes how these moves work and how they fit together.

#### auto.Config

Again, you almost never generate MoveTypeConfigs yourself, but rather use `auto.Config()` from the `moves/auto` package. See the package doc of `moves/auto` to learn more about how to use it.

### Configs

Games can often take different configurations. For example, a deck-based card game might be playable with an expansion pack of cards mixed in. 

These are represented in the engine by the notion of a `GameConfig` which is just a `map[string]string`. When your game is created, a bundle of Config will be passed to `SetUp`, along with how many players are in the game. That config is simply passed to your `GameDelegate`'s `BeginSetUp` method, and that's it. It's your game's responsibility to take that information to set properties differently so the game can be configured that way.

There are a few other extension points for `Config`. One is `GameDelegate.Configs() map[string][]string`. This is a purely optional method that just enumerates different keys your game understands and valid values for them, like in `memory`:

```
const (
	configKeyNumCards = "numcards"
	configKeyCardSet  = "cardset"
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

func (g *gameDelegate) Configs() map[string][]string {
	return map[string][]string{
		configKeyCardSet:  {cardSetAll, cardSetFoods, cardSetAnimals, cardSetGeneral},
		configKeyNumCards: {numCardsMedium, numCardsSmall, numCardsLarge},
	}
}
```

These values are used primarily just so the webapp can create reasonable fields in the UI. 

Another extension point is `GameDelegate.LegalConfig()`. When a Game is being `SetUp`, just after the number of players is checked for legality, the config object is passed to that method. If it returns an error then SetUp will fail. `DefaultGameDelegate` just verifies that all of the keys and values are legal according to the return value of `Configs()`, which is almost always what you want.

There are two other methods on `GameDelegate`,  `ConfigKeyDisplay` and `ConfigValueDisplay`, which are used to get strings to show to the user in the web app UI.

### Agents

Not all players of a game are human. You also want bots or AIs to be able to play. In the engine these are called *agents*.

Agents are configured on the manager when it is created by returning agents in
your delegate's ConfigureAgents() method. There can be multiple agents,
representing different AIs--although in practice you'll likely only have one.
Agents are set up when the game is set up, and then have a callback called
after every move is made to have a chance to propose a move.

The interface that agents must implement is simple:

```
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

### Hooking into your own server

So far this tutorial has told you how to start up an existing server. But configuring your own server, with your own custom collection of games, is incredibly simple too.

The server example in `examples/server/api` is very simple:

```
package main

import (
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/storage/bolt"
)

func main() {
	storage := api.NewServerStorageManager(bolt.NewStorageManager(".database"))
	defer storage.Close()
	api.NewServer(storage,
		api.MustNewManager(blackjack.NewManager(storage)),
		api.MustNewManager(tictactoe.NewManager(storage)),
		api.MustNewManager(memory.NewManager(storage)),
		api.MustNewManager(debuganimations.NewManager(storage)),
		api.MustNewManager(pig.NewManager(storage)),
	).Start()
}
```

It uses the `bolt` storage backend for simplicity because it doesn't require configuration. But if you wanted to use the mysql backend instead, it would require just a couple of lines changed:

```
package main

import (
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api"
)

func main() {
	//The default storage engine is mysql
	storage := api.NewDefaultStorageManager()
	defer storage.Close()
	api.NewServer(storage,
		api.MustNewManager(blackjack.NewManager(storage)),
		api.MustNewManager(tictactoe.NewManager(storage)),
		api.MustNewManager(memory.NewManager(storage)),
		api.MustNewManager(debuganimations.NewManager(storage)),
		api.MustNewManager(pig.NewManager(storage)),
	).Start()
}
```

You'd also need to have a mysql connection string configured in your config.SECRET.json, something like:
```
{
	//...
	"dev" : {
		//...
		"storageconfig": {
			"mysql" : "root:root@tcp(localhost:3306)/boardgame"
		}
	}
}
```

Of course, you'd also need to stand up a mysql server to connect to and configure the tables correctly. The `boardgame-mysql-admin` tool, in `storage/mysql/cmd/boardgame-mysql-admin`, makes this incredibly easy to setup a server and roll its schema forward. 

Check out `storage/mysql/README.md` for more information on configuring the server connection string and using `boardgame-mysql-admin`.

### Conclusion

This library is a passion project I'm pursuing in my free time. It's under active development. If you see something that seems to be missing or off, please reach out via a GitHub issue. And pull requests are very appreciated!




