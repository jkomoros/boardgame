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
import { html } from '../../src/client.js';
import { TableRenderer, registerTableRenderer } from './_game_renderer.js';

@registerTableRenderer
export class MyGameTableView extends TableRenderer {
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
import { html } from '../../src/client.js';
import { HandRenderer, registerHandRenderer } from './_game_renderer.js';

@registerHandRenderer
export class MyGameHandView extends HandRenderer {
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

1. **Import public framework APIs only from `../../src/client.js`.** The
   generated `_game_renderer.js` is the other framework-facing import. Deep
   `src/components/...` paths are implementation details and fail the client
   checker with `BGCLIENT0105`.

2. **Never branch on another player's sanitized property.** Hidden values
   arrive as the property's ZERO VALUE, not as "hidden" — on the table,
   every werewolf's hidden `Role` literally reads `"Villager"`. Verdicts
   that depend on private state (who won, how many wolves remain) must be
   computed server-side and shipped as a public field. The framework's
   `Finished`/`Winners` arrive on every renderer as
   `this.gameFinished`/`this.gameWinners` (indexes) — `renderGameOverBanner()`
   and the hand header consume them for you.

3. **Give controls a typed move action.** Prefer
   `.action=${this.move(MoveNames.X)}` (or
   `.action=${this.move(MoveNames.X).with({ Field: value })}` for move input).
   Framework controls then own disabled, pending, stale-snapshot, and accessible
   error states. Use `isMovePossible()` only when presentation needs to hide an
   action that is structurally irrelevant.

4. **Label people, not seats.** `this.seatPresentations` (both bases)
   carries each seat's avatar slug + display name from the join flow.
   The facade's `glyphForSlug()` renders the avatar. "🐺 WolfBot2", never
   "Player 1".

5. **Keep move creation typed end to end.** The generated move name and input
   maps make unknown names, missing fields, and extra fields compile errors.
   Required-input actions cannot be proposed before `.with(...)` binds their
   exact input. If you find yourself writing `as MoveName`, the name is wrong.

For example:

```typescript
html`<boardgame-action-button
  .action=${this.move(MoveNames.ChooseRole).with({ Role: role })}>
  Choose ${role}
</boardgame-action-button>`
```

The old renderer `propose-move`/`data-arg-*` DOM protocol is removed. Companion
controls use the same snapshot-bound typed actions as solo controls.

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

Custom components that call the framework's `play()` inherit this policy too.
Use `{ timing: 'immediate' }` as the fourth argument for a tap flourish or
other local-only effect. Stack stagger, visible duration, and
`post-animation-delay` share the slot's remaining budget; an effect that can no
longer begin in the slot is omitted rather than snapping late.

## Host actions and ForceFinishTurn

If your game uses turn-based play, register `moves.ForceFinishTurn` in a
phase-agnostic slot so the host's SkipTurn works for absent players. The server
prints a startup warning when a companion-capable game omits it; simultaneous
games may intentionally omit it, while turn-based games should treat the
warning as a broken host recovery path.

Shared-screen ownership and recovery are framework-owned. The Table holds a
short-lived, HttpOnly device lease that its socket renews; local renderer
selection never grants host authority. If the Table disappears, seated Hands
automatically receive a takeover control after the reconnect grace period.
Simultaneous attempts are resolved atomically, and a displaced screen is
paused. Game renderers do not add a recovery button, heartbeat, host identity,
or lease handling—the framework chrome remains present even when an author
completely replaces the Table or Hand renderer layout.

An active Table can also be moved deliberately, without waiting for failure.
The framework's **Move shared Table** control creates a five-minute, one-use QR
link plus a room-code/manual-code fallback. The receiving screen previews the
game and asks for confirmation before it takes control. Redemption rotates the
same durable lease atomically, so the old Table remains fully active until the
new one succeeds and is then fenced and given a clear completion screen. A
receiver does not need an account: the transferred device capability, not a
browser login, authorizes framework-owned host controls. Shared Tables are
always serialized as observers—even if the receiving browser also belongs to
a seated player—so moving a phone to a projector cannot expose that player's
private state. This entire journey is framework-owned; game authors add no
routes, transfer methods, dialogs, or state fields.

## Testing your surfaces on one machine

`?display=table` / `?display=hand` on a game URL override renderer selection
for visual testing only; `?display=table` never grants the fenced Table lease
or host actions. Open the same game in two tabs to inspect both sides. Note the two
tabs share one signed-in identity; for true multi-identity testing use a
second browser profile, or claim seats via `curl` against
`POST /api/join/seat` in `--offline-dev-mode`.

## Type-checking your renderers

`boardgame-util serve` regenerates every client contract in one failure-atomic
transaction and stops immediately if extraction, validation, or installation
fails. Production builds also fail on assembled renderer type errors. Run
`boardgame-util check-client --fix` locally to refresh and check; commit its
generated changes. Run plain `boardgame-util check-client` in CI for the
read-only strict gate: unsafe escapes, stale generation, Lit bindings, and
package-isolated type checking. The
generated `TableRenderer` and `HandRenderer` bases already bind the complete
state, component catalog, move-name, and move-input contract. You get the same
compile-time checking of `state`, `playerState`, and `move(...).with(...)` as
the ordinary generated `GameRenderer`, without repeating generic arguments.
The facade also exports the underlying `BoardgameTableViewBase`,
`BoardgameHandViewBase`, and `SeatPresentation` for advanced framework
adapters.

## Known limitations (V1)

- Adding the two files requires a `serve` restart (no hot-add watcher).
- Google sign-in in the join flow is stubbed; guests are the only path.
- Games whose private state isn't card-shaped get correct sanitization
  and a working hand view, but the marquee deal animation has no target.
