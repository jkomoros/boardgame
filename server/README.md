
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
4. Create mygame/server/api/main.go with the following content:
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
5. Copy boardgame/server/api/app.yaml to be in your mygame/server/api folder.
5. Ensure your .gitignore file contains the following line:

```
*.SECRET.*
```

6. Copy boardgame/server/api/config.SAMPLE.json to be mygame/server/api/config.SECRET.json . Do not commit this file to version control (your gitignore should help you avoid doing that)
7. Create mygame/server/static/webapp directory
8. Create mygame/server/webapp/game-src directory, which is where you will create all of your game-rendering subviews.
8. In mygame/server/static/webapp/game-src, create boardgame-render-game.html and define a polymer element in it. This is the entrypoint for the rendering of your view, and will be passed Game object. The one in boardgame/server/static/webapp/game-src is a reasonable starting point to copy.
9. Copy the following items from boardgame/server/static/webapp to your own webapp. None of them require modification by default.
* bower.json
* manifest.json
* .gitignore.sample -> .gitignore (this will help ensure you don't check in the symlinks)
10. Create symlinks from the following items:
* src 
* config-src
* index.html
* manifest.json
* polymer.json

Example symlink:
```
#sitting in mygame/server/static/webapp
ln -s ~/Code/go/src/github.com/jkomoros/boadgame/server/static/webapp/index.html
```

11. In mygame/server/webapp directory, run:
```
bower update
```
This will create bower_components.


## Adding new views

You can add new views in game-src/ that are imported directly from other views in game-src/. Remember that game-render-view is the polymer element that is the root of your game rendering.

If you need new bower depenencies, just add them as normal from the command line, sitting in mygame/server/webapp. This will modify your bower.json file, which is correct.

If you want to modify config-src, manifest.json, or index.html just remove the symlink and copy in the example folder from boardgame/server/static/webapp.

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

Install the [https://cloud.google.com/sdk/docs/](Google Cloud SDK).

Install aedeploy:
```
go get google.golang.org/appengine/cmd/aedeploy
```


## Building

By default the server will serve from /src, /game-src, and /bower-components. From mygame/server/webapp, you can run

```
polymer build
```

To create results in mygame/server/webapp/build/{bundled, unbundled}. 

## First deploy

Make sure you have created a project for the static and api servers.

Tell the gcloud commands which project you're operating on.

```
gcloud config set project <project-id>
```

### Static

The static app can be hosted anywhere you want. This section describes how to deploy to Google Cloud Storage.

You will be storing as a static domain-backed bucket on Google Cloud Storage. The main instructions to follow are [https://cloud.google.com/storage/docs/hosting-static-website](here), but this guide pulls out the main steps.

Get a domain. If you get it from Google Domains, it will be pre-verified on Google as owned by you.

Set up your domain to have a CNAME that points to c.storage.googleapis.com

Create the storage bucket to serve the files in. It must be based on the domain you will serve from:

```
gsutil mb gs://www.mydomain.com
```

Set the acls to be world-readable (do this before the first push so all files you push get this ACL by default)

```
gsutil defacl set public-read gs://www.mydomain.com
```

Now do a normal deploy, as described in the "Deploying" section below.

Set it so index.html is returned by default for all routes that don't have other objects:

```
gsutil web set -e static/index.html gs://www.mydomain.com
```

This will only work if it's a domain-backed bucket.

### API

Currently (without a SQL backend) there's no special first time set-up to do.

If you want to set up your API server to be at e.g. https://api.mydomain.com, follow the instructions [https://cloud.google.com/appengine/docs/flexible/go/using-custom-domains-and-ssl](here).

## Deploying

### Static

Do a build, as described above.

Cd into boardgame/webapp/build/bundled

Run

```
gsutil -m rsync -r . gs://www.mydomain.com/static
```

If you were to not use a domain backed bucket you can access the files at https://<your-bucket-name>.storage.googleapis.com/static/index.html . (Note that although the files are also available at https://storage.googleapis.com/<your-bucket-name>/static/index.html, the page won't work because index.html needs to use an absolute path to get to sub-resources.) However, you can't set an errHandler except on domain-backed buckets.

### API

Cd into mygame/server/api.

Run:

```
aedeploy gcloud app deploy
```