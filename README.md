# boardgame

boardgame is a work-in-progress package that makes it easy to define multi-player boardgames that can be easily hosted in a high-quality web app with minimal configuration.

The core of your game logic is constructed using the core library into a *game manager* for each game. The server package makes it easy to take those game managers and install them into a server instance. Each game manager defines a basic view that knows how to render any given state of one of its game for the user.

A number of example games are defined in the examples sub-package to demonstrate how to use many of the key concepts. Real documentation for the core game engine is in the [godoc package docs](https://godoc.org/github.com/jkomoros/boardgame).

## Tutorial

*This tutorial will walk through some concrete examples of how to configure a server and create games, in a way that narratively makes sense but leaves a number of topics unexplored or lightly developed. For more in-depth documentation of the core concepts, check out the core library's package doc, and for more about the server, see `server/README.md`*

Each instantitation of a server includes multiple game packages, each of which defines a Game Manager that describes the logic necessary to run that type of game. These game packages are organized in a canonical way to make it easy to link in game packages into your server even if you didn't write them.

An example server can be found in examples/server. This tutorial will walk through how those work.

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

Let's dig into concrete examples in blackjack, in `examples/memory/state.go`.

The core of the states are represented here:

```
//+autoreader
type gameState struct {
	CurrentPlayer  boardgame.PlayerIndex
	HiddenCards    *boardgame.SizedStack
	RevealedCards  *boardgame.SizedStack
	HideCardsTimer *boardgame.Timer
}

//+autoreader
type playerState struct {
	playerIndex       boardgame.PlayerIndex
	CardsLeftToReveal int
	WonCards          *boardgame.GrowableStack `stack:"cards"`
}
```

There's a lot going on here, so we'll unpack it piece by piece.

At the core you can see that these objects are simple structs with (mostly) public properties. The game engine will marshal your objects to JSON and back often, so it's important that the properties be public.

It's not explicitly listed, but the only properties on these objects are ones that are legal according to `boardgame.PropertyType`. Your GameManager would fail to be created if your state structs included illegal property types.

Most of the properties are straightforward. Each player has how many cards they are still allowed to draw this turn, for example.

#### Stacks and Components

As you can see, stacks of cards are represented by either `GrowableStack` or `SizedStack`. A SizedStack has a fixed number of slots, each of which may be empty or contain a single component. A GrowableStack is a variable-length stack with no gaps, that can grow and shrink as components are inserted and removed.

Stacks contain 0 or more **Components**. Components are anything in a game that can move around: cards, meeples, resource tokens, dice, etc. Each game type defines a complete enumeration of all components included in their game in something called a **ComponentChest**. We'll get back to that later in the tutorial.

Each component is organized into exactly one **Deck**. A deck is a collection of components all of the same type. For example, you might have a deck of playing cards, a deck of meeples, and a deck of dice in a game. (The terminology makes most sense for cards, but applies to any group of components in a game.) Memory has only has a single deck of cards.

Each Stack is associated with exactly one deck, and only components that are in that deck may be inserted into that stack. The deck is the complete enumeration of all components in a given set within the game. In blackjack you can see struct tags that associate a given stack with a given deck. We'll get into how that works later in the tutorial.

**Each component must be in precisely one stack in every state**. Later we will see how the methods available on stacks to move around components help enforce that invariant.

When a memory game starts, most of the cards will be in GameState.HiddenCards. And players can also have cards in a stack in their hand when they win them, in WonCards. You'll note that there are actually two stacks for cards in GameState: HiddenCards and RevealedCards. We'll get into why that is later.

#### autoreader

Both of the State objects also have a cryptic comment above them: `//+autoreader`. These are actually a critical concept to understand about the core engine.

In a number of cases (including your GameState and PlayerState), the game package provides the structs to operate on. The core engine doesn't know their shape. In a number of cases, however, it is necessary to interact with specific fields of that struct, or enumerate how many of a certain type of property there are. It's possible to do that via reflection, but that would be slow. In addition, the engine requires that your structs be simple and only have known types of properties, but if general reflection were used it would be harder to detect that.

The core package has a notion of a `PropertyReader`, which makes it possible to enumerate, read, and set properties on these types of objects. The signature looks something like this:

```
type PropertyReader interface {
	Props() map[string]PropertyType
	IntProp(name string) (int, error)
	//... Getters for all of the other PropertyTypes
	Prop(name string) (interface{}, error)
}

type PropertyReadSetter interface {
	PropertyReader
	SetIntProp(name string, value int) error
	//... setters for all of the other PropertyTypes
	SetProp(name string, value interface{}) error
}
```

This known signature is used a lot within the package for the engine to interact with objects specific to a given game type.

Implementing all of those getters and setters for each custom object type you have is a complete pain. That's why there's a cmd, suitable for use with `go generate`, that automatically creates PropertyReaders for your structs.

Somewhere in the package, include:

```
//go:generate autoreader
```

(You'll find it near the top of examples/memory/main.go)

And then immediately before every struct you want to have a PropertyReader for, include the magic comment:

```
//+autoreader
type MyStruct struct {
	//....
}
```

Then, every time you change the shape of one of your objects, run `go generate` on the command line. (That assumes that you have already run `go install` from within `$GOPATH/github.com/jkomoros/boardgame/cmd/autoreader` to install the autoreader command.) That will create `autoreader.go`, with generated getters and setters for all of your objects.

The game engine generally reasons about States as one concrete object made up of one GameState, and n PlayerStates (one for each player). (There are other components of State that we'll get into later.) This object is defined in the core package, and the getters for Game and Player states return things that generically implement the interface. Many of the methods you implement will accept a State object. Of course, it would be a total pain if you had to interact with all of your objects within your own package that way--to say nothing of losing a lot of type safety.

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

#### PlayerIndex

gameState has a property named `CurrentPlayer` of type `boardgame.PlayerIndex`. This property, as you might expect, encodes whose turn it currently is.

It would be reasonable to encode that bit of state as a simple int (and indeed, that's basically what a PlayerIndex property is). However, it's so common to have to encode a PlayerIndex (for example, if there's a move to attack another player), and there are enough convenience methods that apply, that the core engine defines the type as a fundamental type.

PlayerIndexes make it easy to increment the PlayerIndex to the next player (wrapping around at the end). The engine also won't let you save a State with a PlayerIndex that is set to an invalid value.

PlayerIndexes have two special values: the `AdminPlayerIndex` and the `ObserverPlayerIndex`. The AdminPlayerIndex encodes the special omnsicient, all-powerful player who can do everything. Special moves like FixUp Moves (more on those below) are applied by the AdminPlayerIndex. In dev mode it's also possible to turn on Admin mode in the UI, which allows you to make moves on behalf of any player. The ObserverPlayerIndex encodes a run-of-the-mill observer: someone who can only see public state (more on public and private state later) and is not allowed to make any moves.

#### Timer

The last type of property in the states for Memory is the HideCardsTimer, which is of type *boardgame.Timer. Timers aren't used in most types of games. After a certain amount of time has passed they automatically propose a move. For Memory the timer is used to ensure that the cards that are revealed are re-hidden within 3 seconds by the player who flipped them--and if not, flip them back over automatically.

Timers are rare because they represent parts of the game logic where the time is semantic to the rules of the game. Contrast that with animations, where the time that passes is merely presentational.

### GameDelegate

OK, so we've defined our state objects. How do we tell the engine to actually use them?

The answer to that, and many other questions, is the `GameDelegate`. The `GameManager` is a concrete type of object in the main engine, with many methods and fields. But there are lots of instances where your game type needs to customize the precise behavior. The answer is to define the logic in your `GameDelegate` object. The GameManager will consult your GameDelegate at key points to see if there is behavior it should do specially.

The most basic methods are about the name of your gametype:

```
type GameDelegate interface {
	Name() string
	DisplayName() string
	//... many more methods follow
}
```

Those methods are how you configure the name of the type of the game (e.g. 'memory' or 'blackjack', or 'pig') and also what the game type should be called when presented to users (e.g. "Memory", "Blackjack", or "Pig").


The GameDelegate interface is long and complex. In many cases you only need to override a handful out of the tens of methods. That's why the core engine provides a `DefaultGameDelegate` struct that has default stubs for each of the methods a `GameDelegate` must implement. That way you can embed a `DefaultGameDelegate` in your concrete GameDelegate and only implement the methods where you need special behavior.

Most of the methods on GameDelegate are straightforward, like `LegalNumPlayers(num int) bool` which is consulted when a game is created to ensure that it includes a legal number of players.

GameDelegates are also where you have "Constructors" for your core concrete types:

```
type GameDelegate interface {
	//...
	GameStateConstructor() MutableSubState
	PlayerStateConstructor(player PlayerIndex) MutablePlayerState
	//...
}
```

GameStateConstructor and PlayerStateConstructor should return zero-value objects of your concrete types. The only special thing is that PlayerStates should come back with a hidden property encoding which PlayerIndex they are.

In many cases they can just be a single line or two, as you can see for the PlayerStateConstructor in main.go:

```
func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.MutablePlayerState {

	return &playerState{
		playerIndex: playerIndex,
	}
}
```
However, if you look at the GameStateConstructor for memory, you'll see that it is a bit more complicated:

```
func (g *gameDelegate) GameStateConstructor() boardgame.MutableSubState {

	cards := g.Manager().Chest().Deck(cardsDeckName)

	if cards == nil {
		return nil
	}

	//We want to size the stack based on the size of the deck, so we'll do it
	//ourselves and not use tag-based auto-inflation.
	return &gameState{
		HiddenCards:   cards.NewSizedStack(len(cards.Components())),
		RevealedCards: cards.NewSizedStack(len(cards.Components())),
	}
}
```

The reason is because some of the pointer-based types (like Stacks) do not have a reasonable zero-value. Stacks are tied specifically to a particular deck; it is illegal to add a component to a stack that doesn't match its deck. But a nil stack doesn't encode which deck it is affiliated with. You need a zero-valued stack that is tied to the given deck.

The GameState constructor does this by getting a reference to the deck in question and then returning an object with those stacks initalized. You'll note that the other properties are omitted because their zero value is reasonable.

However, it's kind of a pain to have to do this imperative instantitation for all of your pointer types.

If you look closely at the playerState, you'll see that it has a stack, too, of WonCards. But that stack isn't initalized in the PlayerStateConstructor. What gives?

The answer is in the struct tag for playerState. For stacks, you can provide a struct tag that has the name of the deck it's affiliated with. Then you can return a nil value from your constructor for that property, and the system will automatically instantiate a zero-value stack of that shape. (Even cooler, this uses reflection only a single time, at engine start up, so it's fast.) Memory doesn't demonstrate it, but it's also possible to include the size of the stack in that struct tag.

The reason GameState can't use it is because the size of the SizedStack is not known statically, because it varies with the size of the deck. So it has to be done the old fashioned way.

#### Other GameDelegate methods

The GameDelegate has a number of other important methods to override.

One of them is `CheckGameFinished`, which is run after every Move is applied. In it you should check whether the state of the game denotes a game that is finished, and if it is finished, which players (if any) are winners. This allows you to express situations like draws and ties.

After `CheckGameFinished` returns true, the game is over and no more moves may be applied.

Another method is `CurrentPlayerIndex`. This method should inspect the provided state and return the `PlayerIndex` corresponding to the current player. If any player may make a move, you should return `AdminPlayerIndex`, and if no player may make a move, you should return `ObserverPlayerIndex`. This method is consulted for various convenience methods elsewhere.

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

Moves are proposed on a game by calling `ProposeMove` and providing the Move, along with which player it is being proposed on behalf of. (The server package keeps track of which user corresponds to which player; more on that later.) The moves is appended to a queue. One a time the engine will remove the first move in the queue, see if it is Legal for the current state, and if it is will Apply it, as described above.

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

Your moves also must implement the `PropertyReader` interface. Some moves contain no extra fields, but many will encode things like which player the move operates on, and also things like which slot from a stack the player drew the card from.

A `MoveType` is a conceptual type of Move that can be made in a game and is a generic struct in the main package. It vends new concrete Moves of this type via `MoveConstructor` and also has metadata specific to all moves of this type, like what the name of the move is. All of a MoveType's fields and methods return constants except for `MoveConstructor`.

A `MoveTypeConfig` is a configuration object used to create a `MoveType` when you are setting up your `GameManager` to receive a fully formed and ready-to-use `MoveType`.

#### Worked Move Example

#### Player Moves

#### FixUp Moves

#### common Move Types

### NewManager

### Property sanitization
*TODO*

#### Computed properties

### Enums

### Renderers / Client
*TODO*

#### Users vs Players (roles and responsibilities of server and core engine)

### Agents



*Tutorial to be continued...*



