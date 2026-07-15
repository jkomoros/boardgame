# Adding Companion (Table+Hand) Mode to Your Game

Companion mode turns one shared screen into "the Table" (public board,
room code, avatars) and each player's phone into "the Hand" (their private
state and actions). Players join by typing a 4-letter code or scanning the
QR on the shared screen — no accounts needed.

**The entire opt-in is shipping two files.** Everything else — room codes,
QR, the join/identity/avatar/seat flow, presence, host controls,
reconnection, deal animations — is provided by the framework.

## The two files

In your game's client directory (`examples/<name>/client/` or your repo's
equivalent that `server/static/game-src/<name>` links to):

```
boardgame-render-game-<name>-table.ts   # the shared screen
boardgame-render-game-<name>-hand.ts    # each player's phone
```

Your existing solo renderer (`boardgame-render-game-<name>.ts`) is
untouched and keeps working for single-device play.

Restart `boardgame-util serve` after adding the files — the build walk
detects the pair at startup and the create-game form grows a
"Use shared projector + phones" toggle for your game.

## Minimal table view

```ts
import { html } from 'lit';
import { customElement } from 'lit/decorators.js';
import { BoardgameTableViewBase } from '../../src/components/boardgame-table-view-base.js';
import type { GameState, PlayerState } from './_types.js';

@customElement('boardgame-render-game-mygame-table')
export class MyGameTableView extends BoardgameTableViewBase<GameState, PlayerState> {
  override renderBoard() {
    return html`<!-- your public board here: this.state.Game.… -->`;
  }
}
```

That's a complete table: the base's default `render()` wraps your
`renderBoard()` with the room-code banner (giant in the lobby, corner
badge once the room fills), the game-over celebration, the avatar strip,
host controls, and the fake-deck row that deal animations fly to. Want a
different arrangement? Override `render()` instead and call the helpers
(`renderRoomCodeBanner()`, `renderGameOverBanner()`, `renderAvatarStrip()`,
`renderHostControls()`, `renderFakeDeckRow()`) à la carte — that's what
blackjack and werewolf do.

## Minimal hand view

```ts
import { html } from 'lit';
import { customElement } from 'lit/decorators.js';
import { BoardgameHandViewBase } from '../../src/components/boardgame-hand-view-base.js';
import { MoveNames, type MoveName } from './_move_names.js';
import type { MoveArgs } from './_move_args.js';
import type { GameState, PlayerState } from './_types.js';

@customElement('boardgame-render-game-mygame-hand')
export class MyGameHandView extends BoardgameHandViewBase<GameState, PlayerState, MoveName, MoveArgs> {
  override renderHand() {
    const me = this.playerState;   // YOUR player's private state
    return html`<!-- cards, role, actions -->`;
  }
}
```

The default `render()` adds the hand header ("🦊 CleverFox · Your turn /
Waiting for <name>…", and the win/lose verdict at game end), the top-edge
animation anchor, and the base auto-flies newly-dealt cards in from the
top edge. The framework also buzzes the phone (`navigator.vibrate`) when
it becomes this player's turn.

## The rules that keep you out of trouble

1. **Import framework code as `../../src/…`** — never
   `../../../server/static/src/…`. The second form resolves in your repo
   but 500s through the dev server's symlinked layout. The build prints a
   warning naming any file that gets this wrong.

2. **Never branch on another player's sanitized property.** Hidden values
   arrive as the property's ZERO VALUE, not as "hidden" — on the table,
   every werewolf's hidden `Role` literally reads `"Villager"`. Verdicts
   that depend on private state (who won, how many wolves remain) must be
   computed server-side and shipped as a public field. The framework's
   `Finished`/`Winners` arrive on every renderer as
   `this.gameFinished`/`this.gameWinners` (indexes) — `renderGameOverBanner()`
   and the hand header consume them for you.

3. **Gate action buttons on `isMoveCurrentlyLegal(MoveNames.X)`.** Moves
   proposed out of turn are rejected client-side with no user feedback;
   a disabled button is honest UI. See blackjack's hand view.

4. **Label people, not seats.** `this.seatPresentations` (both bases)
   carries each seat's avatar slug + display name from the join flow.
   `glyphForSlug()` renders the avatar. "🐺 WolfBot2", never "Player 1".

5. **Use `proposeMove(MoveNames.X, {…})` with the generated constants** —
   the typed API catches wrong move names at compile time. If you find
   yourself writing `as MoveName`, the name is wrong.

## Private state and the seat picker

- Per-player privacy is the existing `sanitize:` struct-tag machinery —
  nothing companion-specific to configure. `sanitize:"other:hidden"` on a
  player-state property means: my phone sees it, other phones and the
  table don't.
- Embedding `behaviors.PlayerRole` or `behaviors.PlayerTeam` makes your
  game "asymmetric": phones automatically get a seat picker before
  claiming a seat, and `Role`/`Team` default to `other:hidden`. To make
  roles public (e.g. Codenames spymasters), override the tag at your
  embedding site with `sanitize:"all:visible"`.

## Cross-screen deal animation

Works out of the box for card games:

- **Phone side** (`autoFlyIncoming`, default on): any card id newly
  appearing in your player's own stacks flies in from the top edge.
- **Table side** (`autoFlyDeals`, default on): mark your draw pile with
  `id="deal-source"`; when a player's hand grows, their name-stub flies
  from the deck toward the bottom edge. No element = no animation — the
  id's presence is the entire opt-in.
- Bespoke needs: set the flags false and call
  `this.animator?.animateBetween(cardIdOrElement, targetIdOrElement, ms)`.
  The first argument visually ARRIVES FROM the second's position. The framework
  automatically schedules this against the current version's cross-screen
  timeline; there is no timing property to pass through your renderer.
- For an intentionally local effect (for example, a tap flourish that has no
  matching event on another screen), opt out explicitly:
  `this.animator?.animateBetween(card, source, 300, { timing: 'immediate' })`.
  Advanced code may instead use
  `{ timing: { localStartAtMs: someTimestamp } }`.

Each game version owns its own animation slot. Rapid automatic/fix-up moves are
spaced on the server's per-game lane, and queued HTTP state bundles retain their
matching slot. The protocol currently reserves 800ms per synchronized version:
up to 600ms of visible motion and 200ms to prepare the next queued state. The
framework applies the slot to its whole animation pipeline: ordinary FLIP
movement, card/die property effects, automatic deals, and `animateBetween`
calls. Visible motion is capped at 600ms and `animationOverlap` is disabled for
those cycles; use immediate timing for a longer effect that has no cross-screen
counterpart. State installs during the 200ms preparation window before the
target, and WAAPI holds each opening frame until launch. A client joining a
cycle late receives only its remaining visible-motion budget, so it cannot
spill into the next slot. If timing is unavailable or no visible budget remains,
the context is discarded completely and state installs immediately.

## Host actions and ForceFinishTurn

If your game uses turn-based play, register `moves.ForceFinishTurn` in a
phase-agnostic slot so the host's SkipTurn works for absent players. The
server refuses to boot a companion-capable game without it — you'll get a
clear error at startup rather than a broken button at game night.

## Testing your surfaces on one machine

`?display=table` / `?display=hand` on a game URL override the surface
cookie — open the same game in two tabs to see both sides. Note the two
tabs share one signed-in identity; for true multi-identity testing use a
second browser profile, or claim seats via `curl` against
`POST /api/join/seat` in `--offline-dev-mode`.

## Type-checking your renderers

`boardgame-util serve` transpiles renderers WITHOUT type-checking (fast dev
loop). A production `boardgame-util build static` type-checks all game
renderers against the framework types and prints any errors as warnings —
so run a prod build (or `tsc` over the assembled dir) before shipping. The
base classes are generic (`BoardgameTableViewBase<GameState, PlayerState>`
etc.); pass your generated `_types` and you'll get compile-time checking of
`playerState` access and `proposeMove` args, with no casts needed.

## Known limitations (V1)

- Adding the two files requires a `serve` restart (no hot-add watcher).
- Google sign-in in the join flow is stubbed; guests are the only path.
- Games whose private state isn't card-shaped get correct sanitization
  and a working hand view, but the marquee deal animation has no target.
