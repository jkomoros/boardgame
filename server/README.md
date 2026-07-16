
## Running the server

Sitting in a folder that has a valid config file in it or one of its ancestors, run:

`boardgame-util serve`. That will build the api and static servers and run them, so you can visit `localhost:8080`.

`boardgame/boardgame-util/lib/build` is the package that does canonical building of servers for both api and static hosting. You can theoretically build them yourself by hand, but in practice it's best to use those methods (or implicitly use them via `boardgame-util build` and `boardgame-util serve`).

## Writing your client-side views

`boardgame-render-game-GAMENAME` is the generated, typed web component that is
instantiated with the current expanded game state. Extend its generated
`GameRenderer` base and import Lit plus public elements from `src/client.js`.
Bind controls to `this.move(MoveNames.SomeMove)`; use `.with({...})` for one
typed input or `.targets(...)` for a collection. Do not construct
`propose-move` events or side-effect import individual public components.

### Optional: player info

Define `boardgame-render-player-info-GAMENAME` by extending the generated
`PlayerInfoRenderer`. Its typed `state` and `playerIndex` inputs are reactive;
its typed `playerState` getter is always derived from them. Whatever it renders
appears in that player's roster item, which is a natural place for score and
public status.

Override `chip` to customize the small roster badge declaratively:

```typescript
override get chip() {
  return { text: this.playerState?.TokenValue ?? '', color: 'rebeccapurple' };
}
```

The framework propagates changes automatically. Do not dispatch chip-change
events or maintain separate chip properties. Invalid fields, types, colors, or
player indexes fail loudly.

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
  return html`<boardgame-component-zone
    label="Won cards"
    layout="stack"
    .stack=${this.state?.Players[0]?.WonCards ?? null}
    .componentView=${this.cards}>
  </boardgame-component-zone>`;
}
```

The view context explicitly distinguishes visible, hidden, and empty slots. The
zone adds a semantic heading, count, empty state, responsive surface, CSS parts,
and theme tokens, and makes an actionless stack display-only automatically. Its
internal stack retains stable component hosts across state snapshots so card
identity, focus, pooling, and movement animation continue to work. Use the
lower-level `boardgame-component-stack` directly for board/spatial geometry or
unusual animation plumbing.

Stack layout is a closed TypeScript contract: `stack`, `grid`, `fan`, `pile`,
`spread`, `board`, or `spatial`. Use `isStackLayout()` to narrow values from a
dynamic control. Unknown layouts and invalid geometry fail loudly at runtime.
Importing `src/client.js` registers the full curated element set; game renderers
should not side-effect import individual public component modules. The strict
client checker reports those deep imports as `BGCLIENT0105`.

Prefer the component view's typed `properties` callback for card or token
presentation. Use `.componentView=${this.cards.withProperties({ rotated: true })}`
for stack-specific typed properties. `.unsafeComponentAttrs` is reserved for
custom presentation properties with no typed representation; removed proposal
keys (`proposeMove`, `indexAttributes`, and `data-arg-*`) fail loudly there.
Bind interactions with typed actions:

```typescript
const reveals = this.move(MoveNames.RevealCard).targets(
  cards.Components.map((_card, CardIndex) => CardIndex),
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

For more complex processing, render ordinary Lit content from the view or use
`componentView()` with a fresh registered custom element extending
`BoardgameComponent`. Invalid factories fail loudly.

Wrap one game-owned panel per player in `boardgame-player-grid`. It supplies a
named Players region, heading, empty state, and an auto-fitting responsive grid
without imposing a player-state schema or panel appearance:

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

Use the renderer's `playerPresentation(playerIndex)` for the host-supplied,
sanitized player name and color. It always returns a useful numbered fallback,
so ordinary badges require one binding and no Redux/store knowledge:

```typescript
html`<boardgame-player-badge
  .player=${this.playerPresentation(playerIndex)}>
</boardgame-player-badge>`
```

Add `compact` for an avatar-only badge with the accessible name retained. The
badge rejects missing presentations, invalid indices, blank/oversized labels,
and invalid CSS colors loudly. Renderer fixtures accept an optional contiguous
`playerPresentations` array, making long names, colors, and fallback behavior
deterministic in browser tests.

Use `hide-heading` only when another visible heading already names the
collection. Tune `--boardgame-player-grid-min-width` and
`--boardgame-player-grid-gap`; children remain arbitrary Lit content.

Use `boardgame-player-panel` inside that grid when one player area combines
scores, zones, status, and controls:

```typescript
html`<boardgame-player-panel
    label="Player 1"
    .active=${this.currentPlayerIndex === 0}>
  <boardgame-component-zone label="Hand" ...></boardgame-component-zone>
  <boardgame-action-button slot="actions" .action=${pass}>Pass</boardgame-action-button>
</boardgame-player-panel>`
```

It supplies the semantic heading, responsive panel surface, current-player
badge and `aria-current`, plus optional header/status/actions/footer regions.
Game-specific selection, elimination, and role states remain ordinary classes
and content. Style its stable parts or `--boardgame-player-panel-*` tokens.

### Game surface

Use `boardgame-game-surface` as the semantic, responsive root of an ordinary
solo renderer:

```typescript
html`<boardgame-game-surface heading="Memory">
  <boardgame-game-outcome slot="status" ...></boardgame-game-outcome>
  <!-- Primary board and zones use the default slot. -->
  <boardgame-action-bar slot="actions" label="Memory actions">
    <!-- Typed controls. -->
  </boardgame-action-bar>
  <boardgame-turn-status slot="status" .turn=${this.turnStatus}>
  </boardgame-turn-status>
</boardgame-game-surface>`
```

Its optional `status`, `actions`, and `footer` regions are omitted when their
slots are unassigned; `header` content sits beside the required heading. The
centered max width, responsive padding, CSS
parts, and `--boardgame-game-surface-*` tokens make the zero-CSS result useful
while keeping all game content and styling game-owned. Blank headings and
invalid heading levels fail loudly.

The renderer base's typed `turnStatus` getter gives the turn primitive everything
it needs in one binding. It distinguishes the acting player, other players,
observers, admins, and simultaneous turns, and suppresses stale output during
animation or after completion. Optional `.playerLabels` replaces fallback
“Player 1” names. The client facade exports the named `ObserverPlayerIndex`,
`AdminPlayerIndex`, and `AnyPlayerIndex` constants and player-index guards; do
not repeat their numeric values in renderer code.


### Optional: boardgame-fading-text

`boardgame-fading-text` renders and politely announces a callout when its typed
scalar `.trigger` changes. Fixed `message` text and the `new`, `diff`, and
`diff-up` auto-message policies cover the common cases; falsey/truthy
suppression is explicit. Invalid policies and non-finite numeric triggers throw
actionable errors, and reduced-motion preferences collapse the visual effect.

Use `boardgame-status-text` to display a typed string or number, announce its
changes politely, and show a fading change effect. It uses the `diff-up`
strategy by default.

```typescript
html`<boardgame-status-text
  .value=${this.state?.Game.Cards.Components.length ?? 0}>
</boardgame-status-text>`
```

Bind a generated timer reference to `boardgame-timer` for an accessible
countdown and progress indicator without rerendering the whole game at 60Hz:

```typescript
html`<boardgame-timer
  label="Cards hide in"
  .timer=${this.state?.Game.HideCardsTimer ?? null}>
</boardgame-timer>`
```

Use `format="clock"`, `hide-progress`, or `hide-value` for the common variants.
For fully custom Lit markup, `TimerController` exposes an immutable selective
reading and defaults to one update per displayed second; opt into frame cadence
only for continuous visuals.

Use `boardgame-game-outcome` with the renderer's `gameFinished`, `gameWinners`,
and `animating` properties. It withholds and does not announce the verdict until
the final animation settles, handles wins/losses/draws, and rejects contradictory
winner data. `viewer=null` is the public form; a nonnegative player index uses
personal wording.

### Optional: BoardgameBaseGameRenderer

Generated game renderers inherit from `BoardgameBaseGameRenderer`. It exposes
typed `move(...)` actions, state and proposal perspective, current-player
helpers, and the animation lifecycle. Bind a zero-input action directly with
`.action=${this.move(MoveNames.RollDice)}`; use `.with({...})` or `.targets(...)`
for typed inputs. Controls fail closed while legality is unresolved, a proposal
is pending, or animation gates interaction.

Register the concrete class with the exact decorator from the generated
`_game_renderer.js` module:

```typescript
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';

@registerGameRenderer
export class MyRenderer extends GameRenderer { /* ... */ }
```

The generated `registerTableRenderer`, `registerHandRenderer`, and
`registerPlayerInfoRenderer` variants enforce both the exact host tag and the
matching base class. Do not hand-type `customElements.define(...)` for framework
renderer surfaces; reserve it for genuinely game-local auxiliary elements.

For an ordinary labeled target menu, wrap the target collection with
`targetList(targets, key => label)` and bind it once to
`<boardgame-target-list .choices=${...}>`. The list owns batched preview,
disabled reasons, native controls, empty state, and responsive stack/grid
layout while the callback remains exact-key typed. Render candidates directly
only when a choice row needs richer game-specific content.

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

Table+Hand deployments must also provide the same high-entropy join-ticket
secret to every API instance:

```sh
BOARDGAME_JOIN_TICKET_SECRET="at-least-32-random-characters"
```

Release-mode servers fail closed when it is absent. For zero-downtime key
rotation, deploy the new value as `BOARDGAME_JOIN_TICKET_SECRET` and the old
value as `BOARDGAME_JOIN_TICKET_PREVIOUS_SECRET`; remove the previous value
after the ten-minute ticket lifetime. These values belong in the deployment's
secret manager, never in client configuration or source control.

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
