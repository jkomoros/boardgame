
## Running the server

Assuming you have already set up your game, you will need to start both your
static server and your api server.

From mygame/server/static, run:

`go build && ./static`

From mygame/server/api, run:

`go build && ./api`

You can now visit localhost:8080.

## Starting a new game from scratch

1. Create your mygame directory at the right place in your $GOPATH
2. Define your moves, components, etc in the directory.
3. Create mygame/server/static/main.go with the following content:
```
package main

import (
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
	api.NewServer(storage, mygame.NewManager(storage)).Start()
}

```
5. Copy boardgame/server/api/app.yaml to be in your mygame/server/api folder. You  may need to modify the cloud_sql_instances property (see the README in the mysql directory on how to set that).
5. Ensure your .gitignore file contains the following line:

```
*.SECRET.*
```

6. Copy boardgame/server/api/config.SAMPLE.json to be mygame/server/api/config.SECRET.json . Do not commit this file to version control (your gitignore should help you avoid doing that)
7. Create mygame/server/static/webapp directory
8. Create mygame/server/webapp/game-src directory, which is where you will link to all of your game-rendering subviews.
9. Create mygame/client/mygame/
10. Create a relative symlink from mygame/server/webapp/game-src to mygame/client/mygame/ (see #14 for an example)
11. In mygame/client/mygame, create boardgame-render-game-GAMENAME.html (where GAMENAME is the short name of the game) and define a polymer element in it. This is the entrypoint for the rendering of your view, and will be passed Game object.
12. Copy the following items from boardgame/server/static/webapp to your own webapp. None of them require modification by default.
* bower.json
* manifest.json
* firebase.json
* .gitignore
13. Copy polymer.json to your own webapp. Modify it to add the game-src/mygame/boardgame-render-game-GAMENAME.html fragment.
14. Copy config.js and modify the values as necessary.
15. Create symlinks from the following items:
* src 
* index.html (note that you may want to copy this to put in your firebase id and analytics code)

Example symlink:
```
#sitting in mygame/server/static/webapp
ln -s ../../../../server/static/webapp/index.html
```

By doing relative paths, they can be checked into and managed by git, as long as everything is in its canonical path in $GOPATH--because it will work for everyone.

15. In mygame/server/webapp directory, run:
```
bower update
```
This will create bower_components.

## Writing your client-side views

boardgame-render-game-GAMENAME is the Polymer element that will be instantiated and passed state (where state.Game.Stack.Components is an expanded view of your components for convenience). Your view should render that to the screen in whatever way is reasonable.

That view can fire events of the type "propose-move", with a detail containing "name" for the precise name of the Move to make, and "arguments", which is an object containing the non-default arguments for the move. When that move is emitted, it will effectively fill in the corresponding form fileds for that move (ignoring, and thus leaving at their default, any fields that were not explicitly listed in the arguments object), and then submit the move.

### Optional: player info

If you define an element in your GAMENAME/ folder called boardgame-render-player-info-GAMENAME, then it will be insantiated and passed state, expandedState, and playerIndex. Whatever it renders will be shown in the player roster for that player. This is a natural place to put things like score per player and other important status.

if your player-info renderer contains a property called chipText that is a string and has notify:true, that text will be used in the chip on the player picture (a single character is best). If "" is returned, that chip's text will default to just being the index of the player. Similarly, you can define chipColor in a similar way, which should return a valid CSS color. If provided, it will be used instead of the normal colors.

If you don't, don't forget to add it to the polymer.json fragments list above.

### Optional: boardgame-card

One useful element in src/ is boardgame-card, which implements a card that can have an overridable front and back, and can do animations and such.

When you use boardgame-component-stack and boardgame-card in conjunction you'll get powerful animations that just do what you want. They use card.Id and stack.LastIdsSeen to calculate which card is which and animate. It will also do advanced things like cloning old content into the card for when the new state has flipped the card hidden. 

It can be finicky to set all of the cards correctly for the animation to work as you want; the easiest way is to set boardgame-component-stack's stack property to the stack in the state, and then have a dom-repeat with boardgame-card that have item="{{item}}" index="{{index}}", and the card's children how to render if there is content. If you do that, everything will work as expected! This will also automatically set type (see below).

boardgame-card's size can be affected by two css properties: --component-scale (a float, with 1.0 being default size) and --card-aspect-ratio (a float, defaulting to 0.6666). Cards are always 100px width by default, with scale affecting the amount of space they take up physically in the layout, as well as applying a transform to their contents to get them to be the right size. --card-aspect-ratio changes how long the minor-axis is compared to the first. If the scale and aspect-ratio are set based on the position in the layout, the size will animate via boardgame-component-animator as expected.

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


### Optional: boardgame-fading-text

The boardgame-fading-text element will render text that animates when changed. The font size can be changed with `--message-font-size`. The text will be centered in the nearest ancestor positoned block.

You can use boardgame-status-text to render text that will also show the fading effect if the value changes. It uses the 'diff-up' strategy by default for fading text, which can be overriden.

```
<!-- you can bind to message attribute -->
<boardgame-status-text message="{{state.Game.Cards.Components.length}}"></boardgame-status-text>

<!-- you can also just include content which automatically sets message -->
<boardgame-status-text>{{state.Game.Cards.Components.length}}</boardgame-status-text>

```

### Optional: BoardgameBaseGameRenderer

If your game renderer inherits from BoardgameBaseGameRenderer, you'll get a few convenience goodies.

Elements that have a propose-move attribute on them anywhere below will, when tapped, fire a propose-move event with that name. It will also include as arguments to that move any attributes named like `data-arg-my-foo`, where the argument would be represnted in the event as `MyFoo`. If you data-bind to that attribute, remember to use `$=` so that Polymer binds them as attributes, not as properties.

BoardgameBaseGameRenderer also defines a few extra properties, like isCurrentPlayer. 

## Adding new views

You can add new views in game-src/ that are imported directly from other views in game-src/. Remember that game-render-view is the polymer element that is the root of your game rendering.

If you need new bower depenencies, just add them as normal from the command line, sitting in mygame/server/webapp. This will modify your bower.json file, which is correct.

If you want to modify config-src, manifest.json, or index.html just remove the symlink and copy in the example folder from boardgame/server/static/webapp.

## Checking out an existing webapp server

If you're doing a fresh check out of a webapp server, you'll have to recreate the bits that aren't checked into git.

In particular:

1) Create your own server/api/config.SECRET.json
2) Run step 15 in the "starting a new game from scrathc section above"

## configuring the server

This is technically about the api server, but here just to have it in one place.

You configure the server via one or more config.json files. The name of the
file must be config.json, or config.*.json, where * is anything other than
"SECRET". These files must be in the directory you will start the server from.

Some of the information in these files is secret and should not be committed
to source control, while some is not. You may have two files: one named
generically, for any information it's OK to check into source control. If you
have a file named "config.SECRET.json", then that will also be loaded up, and
override anything in the base file (or used directly if nothing is in the base
file). This allows you to add a line to your .gitignore that makes it
impossible for you to accidentally commit the information in
config.SECRET.json. Generally you want to keep all of the non-secret aspects
in config.

Within a config file there are three sub configs: "base", dev" and "prod".
"base" is never used directly, but it sets defaults that "dev" and "prod" will
override and extend.

Both "dev" and "prod" have the same possible fields to set. The server picks
which one to use at start up based on the GIN_MODE environment variable.

### AllowedOrigins

AllowedOrigins*is a comma-delimited list of origins to use in CORS that should
be allowed to access your endpoint.

### DefaultPort

DefaultPort is the port (e.g. "8080") to use when no port is specified in
environment variables.

### FirebaseProjectId

FirebaseProjectId is the ID of your firebase project. It is necessary for user
authentication. Note that you'll also likely need to update it in index.html

### DisableAdminChecking

This option *should only be enabled in dev*. When set to true, it disables all
admin checking. That means that any user can enable admin mode clientside and
then operate as an admin (e.g. make whatever moves they want on a game, view
the state from the perspective of any user, etc).

### AdminUserIds

When adminmode chcecking is enabled (which is the default, see above), only
users whose userId is in this list will be allowed to enable admin mode. You
can find the userIds in the Firebase user console.

### StorageConfig

StorageConfig is how you configure the parameter to be passed to
Storage.Connect(). Different storage backends have different expectations, and
many are fine with just "". When the server is started up, it will fetch the
connection string from this map that matches the Name of the storage engine in
use.

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

Install firebase tool:

```
npm install -g firebase-tools

firebase login
```


## Building

By default the server will serve from /src, /game-src, and /bower-components. From mygame/server/webapp, you can run

```
polymer build
```

To create results in mygame/server/webapp/build/{bundled, unbundled}. 

## First deploy

Make sure you have created a project for the static and api servers.


### First deploy for static hosting on firebase

Go to console.firebase.google.com, enable hosting for your project, and then
follow the steps in connect domain.

### First Deploy for static hosting on Google Cloud STorage

Tell the gcloud commands which project you're operating on.

```
gcloud config set project <project-id>
```

### Static

The static app can be hosted anywhere you want. 

#### Static hosting on Firebase

Make sure that you have a firebase config file in the
mygame/server/static/webapp directory. You can create one with `firebase init`
in that directory, or just link in the one from
boardgame/server/static/webapp. This file configures where the root item is.

Run your build

Sitting in mygame/server/static/webapp, run `firebase deploy`

#### Static Hosting on Google Cloud Storage

This section describes how to deploy to Google Cloud Storage (and may be slightly out of date)

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

Use the `boardgame-mysql-admin` tool. Sitting in the directory with your config.SECRET.json, run:

`boardgame-mysql-admin -prod setup`

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

Make sure your database is up to date. 

Run:
```
boardgame-mysql-admin -prod up
```

Run:

```
gcloud app deploy
```

If it's been awhile since you installed gcloud, you can run:

```
gcloud components update
```

To make sure everything is up to date.