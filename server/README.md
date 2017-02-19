
## Running the server

Assuming you have already set up your game, you will need to start both your
static server and your api server.

From mygame/server/static, run:

```go build && ./static```

From mygame/server/api, run:

```go build && ./api```

You can now visit localhost:8080.

## Starting a new game from scratch

1. Create your mygame directory at the right place in your $GOPATH
2. Define your moves, components, etc in the directory.
3. Create mygame/server/static/main.go with the following content:
```
package main

import (
	"<mygame import>"
	"github.com/jkomoros/boardgame/server/static"
)

func main() {
	static.NewServer().Start()
}
```
4. Creat mygame/server/api/main.go with the following content:
```
package main

import (
	"<mygame import>"
	"github.com/jkomoros/boardgame/server/api"
)

func main() {
	storage := api.NewDefaultStorageManager()
	defer storage.Close()
	api.NewServer(mygame.NewManager(storage), storage).Start()
}

```
(The rest of these steps are in theory, not yet in practice)
5. Create mygame/server/static/webapp directory
6. Create mygame/server/webapp/game-src directory, which is where you will create all of your game-rendering subviews.
7. In mygame/server/webapp/game-src, create boardgame-render-game.html and define a polymer element in it. This is the entrypoint for the rendering of your view, and will be passed Game object. The one in boardgame/server/webapp/game-src is a reasonable starting point to copy.
8. Copy the following items from boardgame/server/webapp. None of them require modification by default.
* bower.json
* index.html
* manifest.json
* polymer.json

_TODO: should polymer.json be a symlink?_

8. Create a symlink from mygame/server/webapp/src to boardgame/server/webapp/src
9. In mygame/server/webapp directory, run the bower update command to install bower components

_TODO: what is the bower update command exactly?_ 

## Adding new views

You can add new views in game-src/ that are imported directly from other views in game-src/. Remember that game-render-view is the polymer element that is the root of your game rendering.

If you need new bower depenencies, just add them as normal from the command line, sitting in mygame/server/webapp. This will modify your bower.json file, which is correct.

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

If you have a fresh checkout, cd into boardgame/server/webapp and run:

```
bower install
```

## Building

By default the server will serve from /src, /game-src, and /bower-components. From mygame/server/webapp, you can run

```
polymer build
```

To create results in mygame/server/webapp/build/{bundled, unbundled}. 

Change the RELEASE_MODE environment variable to change where we serve files from.

## First deploy

Make sure you have created a project for the static and api servers.

Tell the gcloud commands which project you're operating on.

```
gcloud config set project <project-id>
```

### Static

The static app can be hosted anywhere you want. This section describes how to deploy to Google Cloud Storage.

Create the storage bucket to serve the files in:

```
gsutil mb gs://<your-bucket-name>
```

Your bucket name can be whatever you want.

Set the acls to be world-readable:

```
gsutil defacl set public-read gs://<your-bucket-name>
```

Now do a normal deploy, below.

Set it so index.html is returned by default:

```
gsutil web set -e static/index.html gs://<your-bucket-name>
```

## Deploying

### Static

Do a build, as described above.

Cd into boardgame/webapp/build/bundled

Run

```
gsutil -m rsync -r . gs://<your-bucket-name>/static
```

Now you can access the files at https://storage.googleapis.com/<your-bucket-name>/static/index.html

_TODO: describe how to deploy_