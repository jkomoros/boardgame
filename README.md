# boardgame

boardgame is a work-in-progress package that makes it easy to define multi-player boardgames that can be easily hosted in a high-quality web app with minimal configuration.

The core of your game logic is constructed using the core library into a *game manager* for each game. The server package makes it easy to take those game managers and install them into a server instance. Each game manager defines a basic view that knows how to render any given state of one of its game for the user.

A number of example games are defined in the examples sub-package to demonstrate how to use many of the key concepts. Real documentation for the core game engine is in the [godoc package docs](https://godoc.org/github.com/jkomoros/boardgame).

## Tutorial

*This tutorial will walk through some concrete examples of how to configure a server and create games. For more in-depth documentation of the core concepts, check out the core library's package doc*

Each instantitation of a server includes multiple game packages. These game packages are organized in a canonical way to make it easy to link in game packages into your server even if you didn't write them.

An example server can be found in examples/server. This tutorial will walk through how those work.

The server has two components: the static asset hosting and the core game engine API server. 

### API Server

The vast majority of the logic is encapsulated within the API server. However, even these require minimal configuration to set up. Each game type is represented at its core by a GameManager, an object that encapsulates all of the logic specific to that game.

server/api/main.go demonstrates the core of a server:

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
    storage := api.NewDefaultStorageManager()
    defer storage.Close()
    api.NewServer(storage,
        blackjack.MustNewManager(storage),
        tictactoe.MustNewManager(storage),
        memory.MustNewManager(storage),
        debuganimations.MustNewManager(storage),
        pig.MustNewManager(storage),
    ).Start()
}
```

The server simply imports the packages for each game it wants to install, instantiates a GameManager for each, and then passes those to a new server instance.

By convention game manager packages include a `NewManager(storage) error` that returns a new instantiation of a GameManager for that game type. By convention packages also include a `MustNewManager()`, which is equivalent to NewManager but will panic if NewManager errors. Since that panic will happen immediately as the server is started up, it's fine to panic, especially since it allows the main.go of your server to be fewer lines of code.

The API server also expects a config file named `config.SECRET.json` to exist in the same directory as the server. A sample can be found in `server/api/config.SAMPLE.json`. This config changes a few properties of the server, but most critically includes the information about how to connect to your storage backend. Follow the instructions linked to from config.SAMPLE.json for how to set up the default MySQL storage backend.

Start the API server by running

`go build && ./api`

which will start an API server at `localhost:8888`

However, there is no UI; it is simply a REST endpoint. To get UI you'll need the static server.

### Static Assets

In production the static assets will be built and then hosted on a simple static asset service like Firebase hosting. In development there is a simple stub asset server that rewrites a few routes for convenience.

main.go for the example static asset server is incredibly simple:

```

package main

import (
    "github.com/jkomoros/boardgame/server/static"
)

func main() {
    static.NewServer().Start()
}

```

Most of the configuration takes place in which folders are aliased in to be served.

Inside of examples/server/ you will find game-src. This is where the client-side views of each game the server supports are linked. Each game type, by convention, includes a client/GAMETYPE directory within its package. Inside are the views that the web app will use to render that particular game. Two different views are supported: `boardgame-render-game-GAMETYPE` and `boardgame-render-player-info-GAMETYPE`. The former is where the core visual output of the game is rendered. The latter is optional. If provided it will render out information specific to each player in the game, like their current score.

When setting up your server, you generally soft-link from game-src to the client directory for each game. As long as you soft-link to the canonical location in the GOPATH for that directory, it should work even on other machines.

Run the static server with

`go build && ./static`

which will start a server at `localhost:8080` by default.

Now you can visit the web app in your browser by navigating to `localhost:8080`

## Game Managers

Now that you have the server set up, let's dig into how a given game is constructed.

We'll dig into examples/blackjack because it covers many of the core concepts.

At the core of every game is the GameManager. This is an object that encapsulates all of the logic about a game and can be installed into a server.

Each game type, fundamentally, is about representing all of the semantics of a given game state in a **State** and then configuring when and how modifications may be made by defining **Moves**.

### State

The state is the complete encapsulation of all semantically relevant information for your game at any point. Every time a move is succesfully applied, a new state is created, with a version number one greater than the previous current state.

Game states are represented by a handful of structs specific to your game type. All of these structs are composed only of certain types of simple properties, which are enumerated in `boardgame.PropertyType`. The two most common structs for your game are `GameState` and `PlayerState`.

GameState represents all of the state of the game that is not specific to any player. For example, this is where you might capture who the current player is, and the Draw and Discard decks for a game of cards.

PlayerStates represent the state specific to each individual player in the game. For example, this is where each player's current score would be encoded, and also which cards they have in their hand.

Let's dig into concrete examples in blackjac, in `examples/blackjack/state.go`.

The core of the states are represented here:

```
//+autoreader
type gameState struct {
    DiscardStack  *boardgame.GrowableStack `stack:"cards"`
    DrawStack     *boardgame.GrowableStack `stack:"cards"`
    UnusedCards   *boardgame.GrowableStack `stack:"cards"`
    CurrentPlayer boardgame.PlayerIndex
}

//+autoreader
type playerState struct {
    playerIndex    boardgame.PlayerIndex
    GotInitialDeal bool
    HiddenHand     *boardgame.GrowableStack `stack:"cards,1"`
    VisibleHand    *boardgame.GrowableStack `stack:"cards"`
    Busted         bool
    Stood          bool
}
```

There's a lot going on here, so we'll unpack it piece by piece.

At the core you can see that these objects are simple structs with (mostly) public properties. The game engine will marshal your objects to JSON and back often, so it's important that the properties be public.

It's not explicitly listed, but the only properties on these objects are ones that are legal according to `boardgame.PropertyType`. Your GameManager would fail to be created if your state structs included illegal property types.

Most of the properties are straightforward. Each player has whether they have Busted or Stood, for example.

#### Stacks and Components

As you can see, stacks of cards are represented by something called a GrowableStack. There is also a type of stack called a SizedStack, but they aren't used in blackjack.

Stacks contain 0 or more **Components**. Components are anything in a game that can move around: cards, meeples, resource tokens, dice, etc. Each game type defines a complete enumeration of all components included in their game in something called a **ComponentChest**. We'll get back to that later in the tutorial.

Each component is organized into exactly one **Deck**. A deck is a collection of components all of the same type. For example, you might have a deck of playing cards, a deck of meeples, and a deck of dice in a game. Blackjack only has a single deck of playing cards.

Each Stack is associated with exactly one deck, and only components that are in that deck may be inserted into that stack. In blackjack you can see struct tags that associate a given stack with a given deck. We'll get into how that works later in the tutorial.

**Each component must be in precisely one stack in every state**. Later we will see how the methods available on stacks to move around components help enforce that invariant.

When a blackjack game starts, most of the cards will be in GameState.DrawStack. When new rounds start, players will discard their cards into GameState.DiscardStack. And players can also have cards in a stack in their Hand. You'll note that there are actually two Hand stacks for each player: VisibleHand and HiddenHand. We'll get into why that is later.

Note also that the invariant requires that each card must be in precisely one stack in every version. This means that cards that are included in a normal deck but aren't used in blackjack have to go *somewhere*. That's what GameState.UnusedCards is for in blackjack. In practice those cards will never be used in the UI; conceptually they're left behind in the game box and out of sight.

####autoreader
*TODO*


####PlayerIndex
*TODO*

### Property sanitization
*TODO*

### Renderers
*TODO*



*Tutorial to be continued...*



