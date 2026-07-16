
## Running the server

Sitting in a folder that has a valid config file in it or one of its ancestors, run:

`boardgame-util serve`. That will build the api and static servers and run them, so you can visit `localhost:8080`.

`boardgame/boardgame-util/lib/build` is the package that does canonical building of servers for both api and static hosting. You can theoretically build them yourself by hand, but in practice it's best to use those methods (or implicitly use them via `boardgame-util build` and `boardgame-util serve`).

## Writing your client-side views

boardgame-render-game-GAMENAME is the web component that will be instantiated and passed state (where state.Game.Stack.Components is an expanded view of your components for convenience). Your view should render that to the screen in whatever way is reasonable.

That view can fire events of the type "propose-move", with a detail containing "name" for the precise name of the Move to make, and "arguments", which is an object containing the non-default arguments for the move. When that move is emitted, it will effectively fill in the corresponding form fileds for that move (ignoring, and thus leaving at their default, any fields that were not explicitly listed in the arguments object), and then submit the move.

### Optional: player info

If you define an element in your GAMENAME/ folder called boardgame-render-player-info-GAMENAME, then it will be insantiated and passed state, expandedState, and playerIndex. Whatever it renders will be shown in the player roster for that player. This is a natural place to put things like score per player and other important status.

if your player-info renderer contains a property called chipText that is a string and has notify:true, that text will be used in the chip on the player picture (a single character is best). If "" is returned, that chip's text will default to just being the index of the player. Similarly, you can define chipColor in a similar way, which should return a valid CSS color. If provided, it will be used instead of the normal colors.

If you create a new renderer component, make sure it's properly imported so the build system can find it.

### Optional: boardgame-card

One useful element in src/ is boardgame-card, which implements a card that can have an overridable front and back, and can do animations and such.

When you use `boardgame-component-stack` and `boardgame-card` together, the
framework tracks stable component IDs and automatically animates movement,
flips, and content changes between state snapshots.

boardgame-card's size can be affected by two css properties: --component-scale (a float, with 1.0 being default size) and --card-aspect-ratio (a float, defaulting to 0.6666). Cards are always 100px width by default, with scale affecting the amount of space they take up physically in the layout, as well as applying a transform to their contents to get them to be the right size. --card-aspect-ratio changes how long the minor-axis is compared to the first. If the scale and aspect-ratio are set based on the position in the layout, the size will animate via boardgame-component-animator as expected.

Define each deck's appearance once with a renderer-scoped, typed component view.
The generated `GameState` supplies the exact stack type:

```typescript
private readonly cards = cardView<GameState['DrawStack']>({
  render: ({ kind, component }) => kind === 'visible'
    ? html`<div>${component.Values.Type}</div>`
    : null,
});

render() {
  return html`<boardgame-component-stack
    layout="stack"
    .stack=${this.state?.Players[0]?.WonCards ?? null}
    .componentView=${this.cards}
    .componentAttrs=${{ disabled: true }}>
  </boardgame-component-stack>`;
}
```

The view context explicitly distinguishes visible, hidden, and empty slots. The
stack retains stable component hosts across state snapshots so card identity,
focus, pooling, and movement animation continue to work.

Prefer the component view's typed `properties` callback for card or token
presentation. `.componentAttrs` is available for lower-level shared properties,
such as rotating every card in one stack. Bind interactions with typed actions:

```typescript
const reveals = this.move(MoveNames.RevealCard).targets(
  cards.Components.map((_card, CardIndex) => CardIndex),
  CardIndex => ({ CardIndex }),
);

return html`<boardgame-component-stack
  layout="grid"
  .stack=${cards}
  .componentView=${this.cards}
  .componentActions=${reveals.candidates.map(candidate => candidate.action)}>
</boardgame-component-stack>`;
```

For more complex processing, render ordinary Lit content from the view or use
`componentView()` with a fresh registered custom element extending
`BoardgameComponent`. Invalid factories fail loudly.


### Optional: boardgame-fading-text

The boardgame-fading-text element will render text that animates when changed. The font size can be changed with `--message-font-size`. The text will be centered in the nearest ancestor positoned block.

You can use boardgame-status-text to render text that will also show the fading effect if the value changes. It uses the 'diff-up' strategy by default for fading text, which can be overriden.

```typescript
html`<boardgame-status-text
  .message=${String(this.state?.Game.Cards.Components.length ?? 0)}>
</boardgame-status-text>`
```

### Optional: BoardgameBaseGameRenderer

Generated game renderers inherit from `BoardgameBaseGameRenderer`. It exposes
typed `move(...)` actions, state and proposal perspective, current-player
helpers, and the animation lifecycle. Bind a zero-input action directly with
`.action=${this.move(MoveNames.RollDice)}`; use `.with({...})` or `.targets(...)`
for typed inputs. Controls fail closed while legality is unresolved, a proposal
is pending, or animation gates interaction.

## Adding new views

You can add new views in game-src/ that are imported directly from other views in game-src/. Remember that game-render-view is the web component that is the root of your game rendering.

If you need new npm dependencies, just add them as normal from the command line using `npm install`.

If you want to modify config-src, manifest.json, or index.html just remove the symlink and copy in the example folder from boardgame/server/static/webapp.

## Checking out an existing webapp server

If you're doing a fresh check out of a webapp server, you'll have to recreate the bits that aren't checked into git.

In particular:

1) Create your own server/api/config.SECRET.json
2) Run step 15 in the "starting a new game from scrathc section above"

## configuring the server

See `boardgame-util/README.md` for how to configure the server via
config.json.

## Installing dependencies

1. Install Node.js (which includes npm)
2. Run `npm install` in the `boardgame/server/static` directory to install all JavaScript dependencies

Install the [https://cloud.google.com/sdk/docs/](Google Cloud SDK).

Install firebase tool:

```sh
npm install -g firebase-tools

firebase login
```


## Building

Use `boardgame-util build` and friends.

## First deploy

Make sure you have created a project for the static and api servers.


### First deploy for static hosting on firebase

Go to console.firebase.google.com, enable hosting for your project, and then
follow the steps in connect domain.

### First Deploy for static hosting on Google Cloud STorage

Tell the gcloud commands which project you're operating on.

```sh
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

```sh
gsutil mb gs://www.mydomain.com
```

Set the acls to be world-readable (do this before the first push so all files you push get this ACL by default)

```sh
gsutil defacl set public-read gs://www.mydomain.com
```

Now do a normal deploy, as described in the "Deploying" section below.

Set it so index.html is returned by default for all routes that don't have other objects:

```sh
gsutil web set -e static/index.html gs://www.mydomain.com
```

This will only work if it's a domain-backed bucket.

### API

Use the `boardgame-util db` tool. Sitting in the directory with your config.SECRET.json, run:

`boardgame-util db --prod setup`

If you want to set up your API server to be at e.g. https://api.mydomain.com, follow the instructions [https://cloud.google.com/appengine/docs/flexible/go/using-custom-domains-and-ssl](here).

## Deploying

### Static

Do a build, as described above.

Cd into boardgame/webapp/build/bundled

Run

```sh
gsutil -m rsync -r . gs://www.mydomain.com/static
```

If you were to not use a domain backed bucket you can access the files at https://<your-bucket-name>.storage.googleapis.com/static/index.html . (Note that although the files are also available at https://storage.googleapis.com/<your-bucket-name>/static/index.html, the page won't work because index.html needs to use an absolute path to get to sub-resources.) However, you can't set an errHandler except on domain-backed buckets.

### API

Cd into mygame/server/api.

Make sure your database is up to date. 

Run:

```sh
boardgame-util db --prod up
```

Run:

```sh
gcloud app deploy
```

If it's been awhile since you installed gcloud, you can run:

```sh
gcloud components update
```

To make sure everything is up to date.
