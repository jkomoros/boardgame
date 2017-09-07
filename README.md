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

*Tutorial to be continued...*



