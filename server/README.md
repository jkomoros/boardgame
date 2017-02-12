
## Running the server

_TODO: describe steps_

## Starting a new game from scratch

1. Create your mygame directory at the right place in your $GOPATH
2. Define your moves, components, etc in the directory.
3. Create mygame/server/main.go with the following content:
```
package main

import (
	"<mygame import>"
	"github.com/jkomoros/boardgame/server"
)

func main() {
	server.NewServer(<mygame-import-name>.NewGame()).Start()
}
```
4. Create mygame/server/webapp directory
5. Create mygame/server/webapp/game-src directory, which is where you will create all of your game-rendering subviews.
6. In mygame/server/webapp/game-src, create game-render-view.html and define a polymer element in it. This is the entrypoint for the rendering of your view, and will be passed Game object.
6. Copy the following items from boardgame/server/webapp. None of them require modification by default.
* bower.json
* index.html
* manifest.json
* polymer.json
7. Create a symlink from mygame/server/webapp/src to boardgame/server/webapp/src
8. In mygame/server/webapp directory, run the bower update command to install bower components

_TODO: what is the bower update command exactly?_ 

## Adding new views

You can add new views f

## Installing dependencies

1. Install Node.js
2. Install bower:

```
npm install -g bower
```

1. Install polymer CLI:

```
npm install -g polymer-cli
```