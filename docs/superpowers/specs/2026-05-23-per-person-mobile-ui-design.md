# Table + Hand Mode (Per-Person Mobile UI)

Spec date: 2026-05-23 (fifth revision — concrete mechanisms for remaining gaps)
Ticket: [#759 — Allow remote control / projected mode](https://github.com/jkomoros/boardgame/issues/759)
Branch: `per-person-mobile-ui`

## Revision Notes

### Revision 5 (concrete mechanisms)

A second code-tracing critic pass on Revision 4 found that several "fixes" were themselves underspecified or still had real gaps. This revision applies the concrete corrections:

- **Embedding-site sanitize override mechanism**: Go's `reflect.Type.FieldByNameFunc` returns the *inner* field's tag for promoted properties, NOT the outer embedding struct's tag. So `behaviors.PlayerRole \`sanitize:"all:visible"\`` at the embedding site has no effect on the existing inflater. **§6.3.2 (new) specifies the inflater extension required to walk outer embedding tags first, falling back to inner defaults.** The implementation plan will include this work item; we explicitly DO add the struct-tag machinery because the override is load-bearing for V1 public-role games.
- **`switchToSolo` plumbing concretized** (§9.6) into four pieces: `eGame.CompanionMode` flag flip, new `mode-changed` WebSocket message type, client reload handler, response that clears surface cookies via `Set-Cookie: Max-Age=0`. Also added warning: switching mid-game in a hidden-info game destroys privacy and is irreversible in V1.
- **`ForceFinishTurn` code example corrected**: now uses `state.PlayerStates()[state.CurrentPlayerIndex()]` and `interfaces.PlayerTurnFinisher` (the actual interface at `moves/interfaces/main.go:78`). Removed the redundant `ResetForTurnEnd` call (already invoked by `FinishTurn.Apply` at `moves/finish_turn.go:90`).
- **Skip-on-non-current-player**: §9.3 now specifies that the Skip button only appears on the badge of the **current** player when they're absent. Non-current absent players display the "Waiting…" indicator without a Skip button — the host has nothing to skip until that player would be active.
- **`ForceAdvancePhase` dropped from V1**: phase advancement is intrinsically game-specific (which phase next?) and cannot be a generic framework move. Phase-driven games where the absent player blocks phase advance fall back to "host has no recourse for V1" — same bucket as simultaneous-action games (§2 Non-Goals).
- **`gameVersionChanged` backward-compatible**: instead of changing the existing `"version"` message's bare-int `Data` shape (which would break old clients at `boardgame-game-state-manager.ts:411-412`), a **new sibling message type `"version-timing"` carries the timestamps**. Old clients ignore the new type; new clients use it. Old `"version"` payload unchanged. §8.4 rewritten.
- **Config distribution to the api binary**: §5.3 now specifies that `boardgame-util` writes a `companion_capable_games` array into `config.json` (the api binary's existing config source via `config.Get`, `server/api/main.go:1738`). No new "read client_config.js at boot" path needed.
- **Audit deliverable corrected**: a grep confirms zero existing games embed `behaviors.PlayerRole` or `behaviors.PlayerTeam` in this repo today. Phase 2 "audit" is empty work; spec text amended to reflect.

### Revision 4 (implementability fix-up)

A code-tracing critic pass against the third revision found several places where the spec assumed mechanisms that don't exist in the codebase as described. This revision fixes the concrete gaps:

- **SkipTurn**: corrected `moves.AdvanceCurrentPlayer` → `moves.FinishTurn` (the actual move), and surfaced that `FinishTurn.Legal()` requires `TurnDone() == nil` — meaning a mid-turn dropout breaks vanilla SkipTurn. Spec now adds a new `moves.ForceFinishTurn` variant that bypasses `TurnDone()` and runs the same `ResetForTurnEnd` defensively. Phase-based games (Werewolf-style with `moves.StartPhase`) require a separate `ForceAdvancePhase` companion — documented in §9.3 with limitations.
- **Filesystem detection**: relocated from the api binary (which has no access to `server/static/game-src/` in either dev or prod) to `boardgame-util/cmd_serve.go` and `cmd_build_static.go`. Capability map emitted into the existing `client_config.js` (already symlinked into the dev temp dir and served in prod). One code path, not the dev/prod split the previous draft described.
- **Sanitization for Visibility**: respecified to use the framework's existing `sanitize:"…"` struct-tag mechanism (rather than a new `visibility:"…"` namespace), with the default sanitize on `PlayerRole.Role` flipped to `sanitize:"other:hidden"` and a public opt-out wired through `ConfigureFromTags`. Concrete mechanism, not hand-wave.
- **Fake-deck row animation**: corrected the same-`component.id` collision in `boardgame-component-animator.ts` by using synthetic stub IDs + a new `animateBetween(realId, stubId)` animator API. The animator records the real card's "first" position from the stub's location and animates to/from there.
- **State push timestamps**: corrected the claim about extending `stateUpdateData` (no such struct; state JSON arrives over a separate HTTP GET, not the WebSocket). Timestamps now ride on the `gameVersionChanged` socket payload (server-controlled), and clients schedule animation start *before* fetching the new state JSON.
- **`HasPlayerTeam` already exists** at `behaviors/team.go:96-105`. Removed the "framework adds this" claim.
- **`SupportsTableHandMode` moved to `managerInfo`** in `server/api/main.go:84` (not the core `boardgame.GameManager` type, which is deliberately HTTP-agnostic). Surfaced to the client via the existing `doListManager` response.
- **Heartbeat plumbing**: surfaces the real edit (threading `playerIndex` into `socket`) and clarifies that the existing 60s `pongWait` is transport-level; the application heartbeat is a new in-band JSON message.
- **Mode lock**: added a "Switch to solo" host action (§9.6) so a failed projector setup doesn't force a game restart.
- **Host on phone**: §9.4 clarifies that host privileges require Table surface; a returning Owner on a phone is just a player.
- **Codenames descoped**: V1 Goals no longer claim Codenames-style games work. Acknowledged as V2 because the spymaster's clue is `(word, number)`, not a card — the fake-deck-row metaphor has no animation target.
- **Inter-player gift privacy**: added as an explicit tested invariant in §13.1.
- **Storage extension**: §6.5 now calls out the four-place edit (mysql / bolt / memory + test harness) and rate-limiting being a new piece of middleware.

### Revision 3 (opinionated reshape)

Initial drafts (commits `076f6cd2`, `14472ac1`) were technically thorough but pushed too much per-game configuration onto authors. Both game-author DX critics independently called the spec "configuration dressed as composition" — seven optional extension interfaces, a typed `Anchor` DSL, three required renderer files per game, two silently-failing required moves.

This revision is opinionated and minimal: **the renderer files ARE the opt-in.** The framework handles everything else by default. A game author writes two normal Lit renderers (and nothing else) and gets cross-screen pairing, animations, presence, and host controls for free.

What dropped:
- The `companion.GameDelegate` marker interface — replaced by filesystem detection at server boot.
- All seven optional extension interfaces (`SeatAssignment`, `RolePrivacy`, `TeamPrivacy`, `SlotLabel`, `AnimationLead`, `AbsentSeatReset`, `AbsentSkipTurn`).
- The `layoutFor` function and `Anchor` enum (`stack` / `offscreen-stack` / `badge` / `secret-value`).
- The `BoardgameSurfaceRendererBase` bundled inheritance.
- The Free-Seat and Replace-with-AI host actions (deferred — only SkipTurn ships).
- The lint / build-time check script (filesystem walk replaces it).

What survives unchanged: cookie-driven surface selection, reusing `ObserverPlayerIndex` for the projector, the `seatPresentation` per-game-per-seat storage table, the piggybacked `ServerSentAt`/`ServerPlayAt` animation sync. The wins were preserved.

What's new: role/team visibility is now a field on the existing `behaviors.PlayerRole` / `behaviors.PlayerTeam` (default `Private`); the framework auto-renders a "fake deck" row along the Table view's bottom edge for cross-screen animations.

## 1. Overview

A new gameplay mode for hidden-information games where one shared screen ("the Table view") shows the public board and each player's phone ("the Hand view") shows their private cards and actions. Cards visibly fly between the two surfaces. Inspired by Jackbox; applied to turn-based board games.

Pairing is anonymous-friendly: the Table view shows a 4-letter room code, phones go to a join URL and type the code, optionally sign in with Google (or stay anonymous), pick an avatar+name, and are bound to a seat for the session.

## 2. Goals & Non-Goals

**Goals**
- A game opts into Table+Hand mode by shipping two TypeScript renderer files. No Go-side opt-in.
- The existing solo-device flow is unaffected.
- Players can join without an account — Firebase anonymous auth is the default path. Google sign-in remains available.
- Cards animate between the Table view and each player's Hand view with visibly synchronized timing.
- A paired player going offline pauses the game; the host (game creator at the Table) can advance with SkipTurn.
- Role/team-asymmetric games get a seat picker on the phone automatically — based on existing `behaviors.PlayerRole` / `behaviors.PlayerTeam` detection.
- Role visibility on the Table view is opt-in per-behavior: hidden by default, public if the author explicitly declares it.

**Non-Goals (V1)**
- Free-Seat and Replace-with-AI host actions.
- Multi-projector setups (one game, two Table views).
- Audience/spectator mode beyond existing `ObserverPlayerIndex`.
- In-room voice chat (ticket #796).
- Upgrading anonymous identity into a persistent Google account post-hoc.
- Cross-game persistent avatar.
- Non-card private info animation (hidden numbers, secret meeples — V1 assumes card-shaped private state, which is the dominant case).
- **Games whose private state isn't a `Stack`** — e.g., Codenames spymasters give a `(word, number)` clue; Spyfall has no private cards at all. The fake-deck-row convention (§8) has no animation target for these games. Their per-player state would still be sanitized correctly and a Hand view could be authored to display it, but the marquee cross-screen animation feature is inert. Acknowledged as a V2 candidate — a "free-form companion" mode where the Hand view doesn't bind to a specific stack.
- **Simultaneous-action games** (all players act at once via `behaviors.PlayerSubmission`) — SkipTurn doesn't apply when there's no current player. V1 still lets these games adopt Table+Hand mode for the per-player-state-on-phone UX, but the host has no recourse for a dropped player; the game pauses until they reconnect (or someone else uses the V2 Free-Seat).

## 3. Author Surface — "Do One Thing, Get Cool Behavior"

This is the entire author surface for adding Table+Hand mode to a game. Every game-author concept in this spec is in this section; the rest is framework internals.

### 3.1 Minimum Opt-In

Ship two new files in your game's static directory:

```
server/static/game-src/<your-game>/
    boardgame-render-game-<your-game>.ts          # existing solo renderer (untouched)
    boardgame-render-game-<your-game>-table.ts    # NEW — public board view
    boardgame-render-game-<your-game>-hand.ts     # NEW — per-player private view
```

The build (via `boardgame-util serve` in dev, `boardgame-util build static` in prod) walks `server/static/game-src/` and emits the capability map into the already-generated `client_config.js`. The game-creation form reads the map at runtime and surfaces a "Use shared projector + phones" toggle for supporting games. That's the entire opt-in. (Mechanism details in §5.)

### 3.2 What Each Renderer Receives

```ts
// boardgame-render-game-<X>-table.ts
import { BoardgameTableViewBase } from '../../src/components/boardgame-table-view-base.ts';

@customElement('boardgame-render-game-<X>-table')
export class TableView extends BoardgameTableViewBase<MyGameState, MyPlayerState> {
    // this.state              — public sanitized state (as ObserverPlayerIndex)
    // this.seatPresentations  — array of { playerIndex, displayName, avatarSlug }
    // this.absentPlayers      — PlayerIndex[]
    // this.isHost             — boolean

    render() {
        return html`
            ${this.renderAvatarStrip()}                 /* base provides */
            ${this.renderHostControls()}                /* base provides; only visible if isHost */
            <div class="board">
                <!-- your game's public board markup -->
            </div>
            ${this.renderFakeDeckRow()}                 /* base provides — animation anchor */
        `;
    }
}
```

```ts
// boardgame-render-game-<X>-hand.ts
import { BoardgameHandViewBase } from '../../src/components/boardgame-hand-view-base.ts';

@customElement('boardgame-render-game-<X>-hand')
export class HandView extends BoardgameHandViewBase<MyGameState, MyPlayerState> {
    // this.state           — this player's sanitized state (as PlayerIndex(n))
    // this.viewingAs       — PlayerIndex
    // this.playerState     — this.state.Players[this.viewingAs] (convenience)

    render() {
        return html`
            <div class="hand">
                ${this.playerState.Hand.Components.map(card => html`<my-card .card=${card}></my-card>`)}
            </div>
            <div class="actions">
                <!-- "Play", "Pass", etc. -->
            </div>
        `;
    }
}
```

That's it. The two renderers render normal Lit. The bases each provide a small number of helper methods (`renderAvatarStrip`, `renderHostControls`, `renderFakeDeckRow`) that handle cross-cutting UI; everything else is ordinary rendering work over the same sanitized state the existing solo renderer uses.

### 3.3 Optional: Public Role/Team Visibility

By default, when a game has `behaviors.PlayerRole` or `behaviors.PlayerTeam`, the Table view does NOT show roles — players appear as numbered seats with avatars only. The role/team is revealed only on the Hand view of the holding player. This works because the framework now ships `behaviors.PlayerRole.Role` with a default `sanitize:"other:hidden"` tag (see §6.3 for the concrete plumbing); the projector connects as `ObserverPlayerIndex` and naturally sees the property as hidden.

A game where roles are *meant to be public* (Codenames: spymasters are publicly known) opts in by overriding the sanitize tag on the embedding site:

```go
type playerState struct {
    base.SubState
    behaviors.PlayerRole `sanitize:"all:visible"`   // override the default "other:hidden"
}
```

The `sanitize:` tag is the framework's existing struct-tag namespace (see `sanitization.go` and existing uses across `examples/*/state.go`); we're just extending which behaviors carry a default. No new tag namespace.

Identical mechanism for `behaviors.PlayerTeam`.

For games that need programmatic control (e.g., visibility changes mid-game), the per-property sanitization policy can be overridden via `base.GameDelegate.SanitizationPolicy()` — this is the existing API at `game_delegate.go:307-321`; no new hook.

### 3.4 That's the Whole Author Surface

No marker interface. No delegate methods. No layout DSL. No required moves. No lint script.

A game author who doesn't care about Table+Hand mode pays zero cost — their existing code is untouched. A game author who opts in writes two normal renderers and (if asymmetric AND public) one struct tag.

---

The rest of this spec is **framework internals**. Authors don't need to read past here unless they're contributing to the framework or debugging.

## 4. Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                              SERVER                                 │
│                                                                     │
│  Existing: GameManager, sanitization-per-player, SeatPlayer,        │
│            channel-based versionNotifier, Firebase auth             │
│                                                                     │
│  NEW (additive, no existing pipeline changes):                      │
│   • Boot-time filesystem walk → capability map per game             │
│   • CompanionRoomCode + CompanionLocked on extendedgame row         │
│   • GameByRoomCode storage lookup                                   │
│   • seatPresentation table (per-(gameID,playerIndex) avatar+name)   │
│   • Default sanitize:"other:hidden" on PlayerRole.Role/PlayerTeam.Team │
│   • Heartbeat-based presence in the existing notifier goroutine     │
│   • ServerSentAt/ServerPlayAt piggybacked on state pushes           │
│   • Host SkipTurn endpoint                                          │
│   • Host transfer (claim if owner heartbeat is stale)               │
└─────────────────────────────────────────────────────────────────────┘
              │                  │              │
   ObserverPlayerIndex     PlayerIndex(0)   PlayerIndex(1)
   + surface_&lt;gameID&gt;=table cookie  + surface_&lt;gameID&gt;=hand cookie    ...
              │                  │              │
       ┌──────┴───────┐    ┌─────┴────┐    ┌────┴─────┐
       │  TABLE VIEW  │    │ HAND A   │    │ HAND B   │
       │  (laptop/TV) │    │          │    │          │
       │              │    │ player 0 │    │ player 1 │
       │  host=Owner  │    │          │    │          │
       └──────────────┘    └──────────┘    └──────────┘

Per game with Table+Hand mode (server/static/game-src/<game>/):
  • boardgame-render-game-<X>.ts        (existing — solo)
  • boardgame-render-game-<X>-table.ts  (NEW)
  • boardgame-render-game-<X>-hand.ts   (NEW)

Shared (server/static/src/components/):
  • boardgame-table-view-base.ts        (NEW — Table view base class)
  • boardgame-hand-view-base.ts         (NEW — Hand view base class)
```

**Cardinal rule**: the server is agnostic about display surfaces *for state purposes*. The Table view connects as `ObserverPlayerIndex`; each phone as `PlayerIndex(n)`. Existing sanitization machinery handles the privacy story end-to-end — no new sanitization variant.

## 5. Mode Detection (Build-Time Filesystem Walk)

**The detection runs in `boardgame-util`, NOT in the api binary.** The api binary in dev runs from a randomly-named `temp_serve_XXXX/` (created by `boardgame-util/cmd_serve.go:135`) with no reference to the project's `server/static/game-src/`; in prod it runs on App Engine where that directory doesn't exist at all. So filesystem walks at api-server boot are impossible. The walk has to happen at build time, in the build tool that already knows where the static tree lives.

### 5.1 Where the walk runs

Two entry points, both already part of the existing build flow:

- **`boardgame-util/cmd_serve.go`** — for dev. Before `api.Build(...)` (line 82) and `static.Build(...)` (line 91), insert a step that walks `server/static/game-src/<game>/` for each registered `delegate.Name()` and notes whether both `boardgame-render-game-<name>-table.ts` AND `boardgame-render-game-<name>-hand.ts` exist.
- **`boardgame-util/cmd_build_static.go`** — for prod static builds. Same walk; same output target.

### 5.2 Where the capability map lands

The existing `client_config.js` is already generated by `boardgame-util` for both dev and prod, already gets symlinked into the temp serve dir, and is already served to clients (it's how the client knows the API URL, Firebase config, `OfflineDevMode`, etc.). We add a single new field:

```js
// client_config.js (already-generated; new field)
window.CLIENT_CONFIG = {
    // ... existing fields ...
    table_hand_supported_games: ["werewolf", "pass", "valentine"],   // NEW
};
```

The game-creation form reads `CLIENT_CONFIG.table_hand_supported_games` to show/hide the toggle. No api-binary code involved.

### 5.3 Server-side capability exposure

The api binary also needs to know per-game whether Table+Hand mode applies (so it can refuse `/api/join`-flow for non-supporting games, and so it can populate per-game metadata in `doListManager`). The api binary already reads its own `config.json` at boot via `config.Get` (`server/api/main.go:1738`); `boardgame-util` writes a new `companion_capable_games []string` field into config.json as part of the same build step that emits the client-side map. At boot, the api populates `managerInfo.supportsTableHandMode bool` (`server/api/main.go:84`) from this field; surfaced via the existing `doListManager` response (which already carries per-manager flags like `playerHasSeat`).

Concretely:
- **Build step** (`boardgame-util/cmd_serve.go`, `cmd_build_static.go`): walks `server/static/game-src/<delegate.Name()>/`, builds a string slice of games with both `*-table.ts` and `*-hand.ts`. Sets a new typed field on `ClientConfig` (`boardgame-util/lib/config/client.go:6-12` — the same struct that today carries `OfflineDevMode`, Firebase config, etc.). The existing `client_config.js` generator at `boardgame-util/lib/build/static/build.go:167-193` serializes the struct to JS for the browser AND `boardgame-util` writes the same field into `config.json` (under `companion_capable_games`) for the api binary to consume.
- **Server side**: api binary reads `config.json` at boot via existing `config.Get` (`server/api/main.go:1738`), populates per-manager flag in `managerInfo` (`server/api/main.go:84`).
- **Client side**: game-creation form reads `window.CLIENT_CONFIG.tableHandSupportedGames` (the typed-field name will follow existing camelCase conventions in `ClientConfig` JSON output) to decide whether to show the toggle.

One typed source (`ClientConfig.TableHandSupportedGames`), two consumers (browser via the existing JS serialization path; api via the existing config.json read path). No "api parses JS file" trickery; no raw-string emission.

**Not on `boardgame.GameManager` directly** — that type is in the core library, deliberately HTTP-agnostic and filesystem-agnostic. Per-deployment capability belongs in the server's `managerInfo`, not the core library.

### 5.4 Dev hot-add

**Status: not implemented in V1 — restart required.** The capability walk
runs once at `boardgame-util serve` startup, and the resulting list is
baked into both `client_config.js` and the generated api binary. Adding
`-table.ts`/`-hand.ts` files mid-serve therefore requires restarting
`serve`. (The originally-specced design — a watcher that re-runs the walk
and lets Vite HMR refresh the form — remains a V2 candidate; it also needs
a story for the api binary's baked-in `supportsTableHandMode` list.) The
authoring guide (docs/companion-mode-authoring.md) documents the restart.

## 6. Identity & Pairing

### 6.1 Room Codes

- 4 uppercase letters from a confusion-resistant alphabet (omit O/I/L/Z → 22^4 = 234,256 codes).
- Generated at game-create when Table+Hand mode is chosen. Stored as a typed field `CompanionRoomCode string` on `extendedgame.StorageRecord` (see `server/api/extendedgame/main.go:15`). **No in-memory registry** — the storage row is the single source of truth.
- **Lookup** via a new storage method `GameByRoomCode(code string) (gameID, error)`, paralleling `UserIDsForGame()` at `server/api/storage.go:58-61`.
- **Collision handling**: generation retries up to 10 random draws; if all collide, falls back to a 5-letter code.
- **TTL grace period**: codes remain exclusively assigned for 24h after `Finished == true` before becoming eligible for recycling.
- **Rate limiting**: `/api/join` is rate-limited per IP to 10 lookups/minute with exponential backoff on consecutive 404s.
- **Room lock**: the host can flip a "Lock room" toggle on the Table view at any time (`CompanionLocked bool` on `eGame`). Open by default to preserve the Jackbox-style lightweight feel.

### 6.2 Phone Join Endpoint

All endpoints in this section require HTTPS in production. Cookies set with `Secure; HttpOnly; SameSite=Lax`.

`POST /api/join` with body `{ "code": "JKLB" }`:

1. Rate-limit check (per IP).
2. Look up game via `GameByRoomCode(code)`. 404 if not found, locked, or `Finished`.
3. Return `{ gameID, gameName, displayName, minPlayers, maxPlayers, currentPlayers, requiresSeatPicker }`. **No role/team metadata** — those come only after authentication via `/api/join/seat-options` (§6.3) so brute-forcing `/api/join` does not reveal asymmetric structure.

The phone client then runs the identity step (Firebase anonymous or Google), shows the avatar/name picker for anon users, runs the seat picker if `requiresSeatPicker == true` (§6.3), and finally posts `POST /api/join/seat`:

```json
{ "gameID": "<id>", "seatPick": <playerIndex>?, "avatarSlug": "...", "displayName": "..." }
```

(`seatPick` is required iff `requiresSeatPicker == true`.)

The seat endpoint:

1. Validates the Firebase ID token (existing `auth.go:164` path).
2. Validates `displayName` (§6.6).
3. Acquires a per-game lock spanning the SeatPlayer proposal.
4. Writes `displayName` + `avatarSlug` to a new `seatPresentation` row keyed on `(gameID, playerIndex)`.
5. Runs the existing `SeatPlayer` proposal path.
6. Issues a `surface_<gameID>=hand` cookie (one cookie per game, keyed by gameID).

Race resolution: if two phones race for the last seat, the loser receives 409 Conflict + the latest seat-availability snapshot and retries against it.

### 6.3 Seat Picker for Asymmetric Games

The framework auto-detects asymmetry by type-asserting the game's `playerState` against `behaviors.HasPlayerRole` or `behaviors.HasPlayerTeam`. Both interfaces and their detection methods (`GetPlayerRole`, `GetPlayerTeam`) **already exist in the codebase** — see `behaviors/role.go:26` and `behaviors/team.go:96-105`. This spec does not add them; it just consumes them.

For asymmetric games, `requiresSeatPicker = true`. The phone calls `GET /api/join/seat-options?gameID=<>` with its Firebase ID token and receives:

```json
{
  "slots": [
    {"playerIndex": 0, "label": "Seat 1", "filled": true,  "avatar": {...}},
    {"playerIndex": 1, "label": "Seat 2", "filled": false},
    {"playerIndex": 2, "label": "Seat 3", "filled": false},
    {"playerIndex": 3, "label": "Seat 4", "filled": true,  "avatar": {...}}
  ]
}
```

For each slot, the label is computed server-side based on whether the role/team property is sanitized as public for the projector. The projector receives sanitization via `ObserverPlayerIndex` (existing machinery); the same sanitization governs the label.

| `sanitize` tag on PlayerRole.Role / PlayerTeam.Team | What `ObserverPlayerIndex` sees | Slot label on phone |
|-----------------------------------------------------|---------------------------------|---------------------|
| `sanitize:"other:hidden"` (new default, see §6.3.1) | property hidden                 | "Seat N"            |
| `sanitize:"all:visible"`                            | property visible                | "<Role display name>" (e.g., "Spymaster (Red)") |

The choice of `seatPick` becomes the player's `playerIndex` on the seat endpoint.

### 6.3.1 Concrete Sanitization Plumbing

Sanitization in this framework is property-name keyed and resolved at inflater-construction time — see `struct_inflater.go:438` (`PropertySanitizationPolicy`) and the `sanitize:` struct-tag mini-language in `sanitization.go`. A runtime-mutable `Visibility` field on a behavior cannot directly influence per-property policies (the policy map is precomputed). This spec uses the existing tag mechanism, plus a small extension to make embedding-site overrides actually work:

1. **Default**: this spec changes `behaviors.PlayerRole.Role` to ship with a default `sanitize:"other:hidden"` tag (added to its existing `enum:"role"` tag at `behaviors/role.go:16`). Same change applied to `behaviors.PlayerTeam.Team` at `behaviors/team.go:32`. Effect: `ObserverPlayerIndex` (i.e., the projector) sees these properties as hidden by default; the player's own Hand view still sees them (since "other" excludes self).
2. **Opt-out at the embedding site**: a game that wants public roles overrides the tag in its own `playerState` struct (as shown in §3.3). See §6.3.2 — this requires a small extension to the inflater because Go's default reflection of promoted properties doesn't carry the outer embedding-site tag.
3. **Programmatic override (rare)**: `base.GameDelegate.SanitizationPolicy()` at `game_delegate.go:307-321` can return a per-property `Policy` based on runtime state. This is the existing API; no new hook needed.

**Existing-game compatibility**: a grep of `examples/` and `server/static/game-src/` shows **no games today embed `behaviors.PlayerRole` or `behaviors.PlayerTeam`**. So the default flip is safe — there's nothing to migrate. The Phase 2 "audit" deliverable is effectively confirmation that the grep stays empty as of the Phase 2 PR; new games introduced before then would be checked.

**Why no new tag namespace**: the third revision proposed a custom `visibility:"public"` tag; the code-traced critic pointed out this isn't idiomatic (existing tags are `enum:"…"`, `sanitize:"…"`). Reusing `sanitize:` keeps the convention tight.

### 6.3.2 Inflater Extension for Embedding-Site Tag Overrides

**The problem.** Go's `reflect.Type.FieldByNameFunc(propName)` (used at `struct_inflater.go:962-990`) returns the *promoted inner* field's `Tag` when looking up a struct field by its promoted name. It does NOT return the tag on the outer anonymous-embed site. So today, given:

```go
type playerState struct {
    base.SubState
    behaviors.PlayerRole `sanitize:"all:visible"`   // outer tag
}
```

…and `behaviors.PlayerRole.Role` carrying `sanitize:"other:hidden"` as its inner default, the existing inflater sees only `sanitize:"other:hidden"`. The outer override is dropped on the floor.

**The fix.** This spec adds a small extension to `StructInflater` (the one place that resolves per-property sanitization, at `struct_inflater.go:125` and the per-property lookup at `:438`). The new resolution rule:

1. For each anonymous-embedded field on the struct being inflated, capture its outer tag (via `reflect.StructField.Tag` on the embedding field — not the inner promoted one).
2. When resolving a promoted property name to a sanitization policy: first check whether the outer tag has a `sanitize:` value (the outer override); if it does, use that. Otherwise fall back to the inner field's `sanitize:` tag (current behavior).
3. Precedence: outer-embedding-site tag > inner default. This is the standard "child overrides parent" precedence, which matches developer expectations from struct-tag overrides in other Go libraries (e.g., gorm).

This extension is **load-bearing for V1 public-role games** (Codenames-style — if we ever ship one — and any user-defined game wanting public roles). Without it, the §3.3 author experience ("override one tag at the embedding site") is broken.

**Scope**: ~30-50 lines of new code in `struct_inflater.go`, plus tests that:
- Confirm the default (inner `sanitize:"other:hidden"`) hides from observer.
- Confirm the outer override (`sanitize:"all:visible"`) shows to observer.
- Confirm a game NOT embedding either behavior is unaffected (no regression).

This work item is included in **Phase 2** as the foundation for the public-role opt-in — see §14.

### 6.4 Identity Paths

- **Firebase anonymous**: `firebase.auth().signInAnonymously()` on the phone returns a UID. Server validates the ID token via existing `firebase.VerifyIDToken()` at `server/api/auth.go:164` (works for anon UIDs unchanged). A user `StorageRecord` is created on first contact with empty Email/DisplayName — the chosen DisplayName goes to `seatPresentation`, not the user record.
- **Google sign-in**: existing flow. `StorageRecord` populated with real email/photo/displayName as today. At the avatar/name step, Google users can optionally override their default name/avatar *for this game*; the override lands in `seatPresentation` without mutating the user record.

### 6.5 seatPresentation Table

Per-(gameID, playerIndex) storage of display name + avatar. New table:

```go
// server/api/seatpresentation/main.go
type StorageRecord struct {
    GameID      string
    PlayerIndex boardgame.PlayerIndex
    DisplayName string
    AvatarSlug  string
}
```

Storage manager extensions:

```go
SeatPresentation(gameID string, p PlayerIndex) (*seatpresentation.StorageRecord, error)
SetSeatPresentation(rec *seatpresentation.StorageRecord) error
ClearSeatPresentation(gameID string, p PlayerIndex) error
```

Per-seat, not per-user, deliberately: a Google user joining game 2 with a different avatar does not mutate game 1's presentation; the user's persistent `StorageRecord` is never written to from the join flow.

**Implementation note — four-place edit**: adding methods to the storage manager interface (`server/api/storage.go:21-77`) forces parallel implementations in `storage/internal/helpers/memory.go`, `storage/bolt/main.go`, `storage/mysql/main.go`, AND the storage test harness referenced from the comment at `server/api/storage.go:76`. MySQL also needs a schema migration. Phase 1 includes all four edits and the migration.

**Rate-limiting middleware**: `/api/join` (and `/api/join/seat`) need per-IP rate limiting. The framework does **not** ship a rate-limiting middleware today; this spec adds a small middleware in `server/api/middleware_ratelimit.go` (token-bucket per IP, configurable limit). Phase 1 includes it.

### 6.6 Display-Name Validation

Server-side, on `/api/join/seat`:

- NFKC-normalize the input.
- Allow only `[A-Za-z0-9 ]`, length 2–24 inclusive after trim.
- Reject zero-width, RTL-override, combining diacritics, C0/C1 control, surrogate ranges.
- Reject if the lowercased normalized name matches another seat's name in the same gameID. The phone retries with a suggested suffix (e.g., "Alice" → "Alice2").
- No slur catalog in V1; hook point exists for one.

## 7. Renderer Model

### 7.1 Surface Selection (Cookie-Based)

Surface is decided at join/create time and stored in a `surface_<gameID>=table` or `surface_<gameID>=hand` cookie (one per game, keyed by gameID so multiple simultaneous games don't collide). URL stays clean (`/game/<name>/<id>`). The loader (`boardgame-render-game.ts:365`) reads the cookie:

```ts
const surface = readSurfaceCookie(gameID);    // 'table' | 'hand' | null
const suffix  = surface === 'table' ? '-table'
              : surface === 'hand'  ? '-hand'
              : '';
await import(`../../game-src/${gameName}/boardgame-render-game-${gameName}${suffix}.ts`);
```

If `surface` is null (solo mode) or the suffixed file isn't present, falls back to the solo renderer with a console warning. The filesystem-detection step (§5) is the authoritative check.

For dev/debug, a `?display=table|hand` query param overrides the cookie. Production has no query-param override.

### 7.2 Table View Base

`BoardgameTableViewBase` is a thin Lit base class (NOT a god-object — see §7.4 on what it deliberately doesn't bundle).

```ts
export class BoardgameTableViewBase<GS, PS> extends LitElement {
    @property() state: PublicGameState<GS, PS>;     // ObserverPlayerIndex-sanitized
    @property() seatPresentations: SeatPresentation[];
    @property() absentPlayers: PlayerIndex[];
    @property() isHost: boolean;
    @property() roomLocked: boolean;
    @property() roomCode: string;

    // Helper renders — call these from your render() where you want them
    protected renderAvatarStrip(): TemplateResult { ... }
    protected renderHostControls(): TemplateResult { ... }   // SkipTurn button when applicable
    protected renderFakeDeckRow(): TemplateResult { ... }    // §8

    // Shared animator wired automatically; its version timeline is framework-owned.
}
```

Helper renders are *opt-in to call*, not invoked automatically. The author composes their own layout and calls the helpers where they want them. No bundled inheritance — if the author wants a custom avatar strip, they don't call `renderAvatarStrip()`.

### 7.3 Hand View Base

```ts
export class BoardgameHandViewBase<GS, PS> extends LitElement {
    @property() state: FullGameState<GS, PS>;       // PlayerIndex(n)-sanitized
    @property() viewingAs: PlayerIndex;
    @property() seatPresentations: SeatPresentation[];

    // Convenience: shortcut to this player's own state
    protected get playerState(): PS { return this.state.Players[this.viewingAs]; }

    // FLIP animator wired automatically.
}
```

The Hand view base is even thinner than Table view base — there's no avatar strip or host controls to expose.

### 7.4 What the Bases Do NOT Bundle

Cross-cutting concerns the *framework* handles (the author doesn't think about):

- Cookie reading + surface selection (done by the loader).
- FLIP wiring (auto, via existing animator).
- Version-bound `serverPlayAt` scheduling inside the shared animator (auto;
  `animateBetween(..., { timing: 'immediate' })` opts a local effect out).
- Presence indicators (`absentPlayers` is just a prop; author renders it however they want, or uses the avatar-strip helper which renders the standard treatment).

Things deliberately *not* in the bases — kept as Lit `ReactiveController`s or out-of-class entirely so they're opt-in composable, not inherited bundles:

- Score ribbon — not all games have a global score; let authors render their own ribbons.
- Phase indicators, turn timers — game-specific; bases don't impose a layout.
- Custom animations beyond FLIP — author hooks into the existing animation system directly.

This is the answer to the layering critic's "BoardgameSurfaceRendererBase bundled inheritance" complaint: each base exposes a *small* set of typed properties + a *small* set of opt-in helper renders, and nothing else. Authors compose, not subclass-and-pray.

## 8. Animation — Fake Deck Row Convention

The framework provides a single convention for cross-screen card animations that game authors don't have to think about.

### 8.1 The Mechanism

On the Table view, the framework auto-renders a "fake deck row" along the bottom edge. The author opts in by calling `this.renderFakeDeckRow()` from their `render()`; if they don't include it, cross-screen animations to/from player hands are disabled with a one-time console warning (the framework does NOT auto-inject — light-DOM injection into author-controlled markup is too magical and breaks Lit's reactive contract).

The row has one **stub stack** per seated player, in seat order, evenly spaced left-to-right. The stub is rendered hidden (CSS-clipped) but its DOM position is real, so the FLIP animator measures it accurately.

**Critical implementation detail — synthetic component IDs.** A naive "render the same `component.id` in both the player's real hand stack AND the table's fake-deck stub" causes a silent collision in `boardgame-component-animator.ts:113` (the animator's flat `result[component.id] = record` map overwrites one record with the other; only one node gets animated correctly). To avoid this, **stubs render placeholder elements with synthetic IDs derived from the real ID**:

```
real card id:  "c17"
stub element id (in fake-deck-3): "stub:p3:c17"
```

The animator gains a new public method `animateBetween(realId: string, stubId: string)` that records the "first" position from the stub element and the "last" position from the real element (or vice versa), enabling cross-position animation without the flat-map collision.

**Result**:
- Deal from public deck → player N's hand: server state moves card `c17` into player N's `Hand` stack. Hand-view client animates `c17` from off-screen-top-edge to the hand position (normal FLIP). Table-view client renders the placeholder `stub:pN:c17` in the fake-deck row for one frame, calls `animateBetween("c17", "stub:pN:c17")` to animate `c17`'s rendered representation from the on-screen deck to the stub position. Visually: card flies down toward player N.
- Player M plays a card from hand → public discard: reverse of the above. Table renders `stub:pM:c17` for a frame and animates from stub → discard. Visually: card flies up.
- Card transfers from player 3 → player 2: Table renders both `stub:p3:c17` and `stub:p2:c17` for one frame, calls `animateBetween("stub:p3:c17", "stub:p2:c17")`. Visually: card slides between players along the bottom row.

The synthetic-ID approach is also useful generally — it's a real animator capability that other future features can build on (e.g., "ghost preview" animations).

### 8.2 On the Hand View Side

On player N's Hand view, the framework auto-positions an off-screen anchor at the top edge representing "the rest of the game" (the public board on the Table view).

- Card dealt to this player: animates from the top edge (entering "from the Table") to the hand stack.
- Card played from this player: animates from the hand stack to the top edge (exiting "toward the Table").
- Card transferred to this player from another: also enters from the top edge — the Hand view does not currently distinguish "from player M vs from the deck" (V2 if the metaphor matters more).

### 8.3 No DSL, No Edge Enums

The author writes nothing to enable this. The framework knows:

- Seat order — from the existing `state.Players` array.
- Per-player private stacks — they live on `playerState`. The framework iterates `playerState`'s `Stack`-typed fields when building the fake-deck row.
- Seat positions on the Table view — left-to-right evenly spaced; computed from `state.Players.length`.

V1 commits to the left-to-right-evenly-spaced layout convention. Games that want a different seat arrangement (e.g. circular for Settlers of Catan-style) are V2.

### 8.4 Sync Primitive

**There is no "state push" — there's a version-change notify over WebSocket, followed by a separate HTTP `GET /api/game/:name/:id/version/:version` round-trip for the state JSON** (see `server/api/main.go:944` and `websockets.go:19-22`). The WebSocket `socketMessage.Data` for the existing `"version"` message type is a bare integer (the version number). So adding timestamps to that message would break existing clients which parse the integer directly (`boardgame-game-state-manager.ts:411-412` calls `setTargetVersion(msg.data)` expecting an int).

To stay backward-compatible, this spec introduces a **new sibling message type `"version-timing"`** that is broadcast immediately after the existing `"version"` message for state changes. Old clients ignore the new type; new clients use it. The existing `"version"` payload shape is unchanged.

```go
// websockets.go — new message type
type versionTiming struct {
    Version      int   `json:"version"`
    ServerSentAt int64 `json:"serverSentAt"`  // ms since epoch, set at broadcast
    ServerPlayAt int64 `json:"serverPlayAt"`  // reserved slot on the game's animation lane
    SlotDurationMS int `json:"slotDurationMs"`
    MaxAnimationDurationMS int `json:"maxAnimationDurationMs"`
}
```

Outbound socket frames for the client (two frames, in this order):

```json
{ "type": "version",         "data": 42 }
{ "type": "version-timing",  "data": { "version": 42, "serverSentAt": 1779712345678, "serverPlayAt": 1779712346178, "slotDurationMs": 800, "maxAnimationDurationMs": 600 } }
```

The first slot after an idle period is `serverSentAt + 500ms`. Rapid consecutive versions reserve monotonically increasing 800ms slots: at most 600ms of synchronized motion plus 200ms to render and pre-arm the next queued state. The timing frame declares both values so this is an explicit protocol policy, not a transport-layer guess about current CSS. Registration sends the newer of its handshake snapshot and the notifier's retained lane version, reusing that version's reserved slot; this also covers a move committed while the socket was registering.

New clients that handle `"version-timing"` index it by `(gameID, version)`, rather than retaining a global latest value. If a complete, valid `"version-timing"` frame doesn't arrive within 200ms after `"version"`, the client falls back to "play immediately on state-fetch". Raw legacy version notifications still fetch state, but do not provide synchronized timing.

Client logic on receipt:

1. On socket open, send three `clock-sync` request/reply rounds. Compute each clock offset from the request/response midpoint and retain the offset from the lowest-round-trip sample.
2. Record `localRxMs = Date.now()` (epoch ms) immediately for each complete timing-policy frame. When clock-sync replies are unavailable, these frames may still warm the minimum one-way fallback; older timing frames that omit the slot policy are invalid rather than partially interpreted.
3. Compute the local equivalent of that version's `serverPlayAt` from the midpoint offset. Carry it with that version's state bundle; never substitute a later version's timing.
4. Fire the HTTP state-fetch in parallel. The state manager installs each queued bundle in its preparation window; delayed WAAPI uses backwards fill to hold the source pose and launch from the compositor timeline. The common animation primitive applies that slot to ordinary FLIP/property effects as well as `animateBetween`, including game-authored bespoke flights; `{ timing: 'immediate' }` is the explicit local-only escape hatch for the latter.

If the target has no visible-motion budget left, is implausibly far in the future, or fewer than three samples have been collected, the timing context is discarded and animation plays immediately. A merely late target has its remaining duration shortened so it still ends inside its slot.

**Documented limitations**:
- Strongly asymmetric request/response routes can still bias midpoint estimation by half the asymmetry.
- A main thread blocked past the preparation window cannot pre-arm WAAPI; that surface degrades to an immediate late animation rather than holding state indefinitely.

The server enqueues version + timing as one socket batch. A full queue closes the socket so it reconnects and resynchronizes; it never silently receives a version without its timing sibling.

## 9. Presence & Host Controls

### 9.1 Presence (Heartbeat-Based)

Each Hand-view WebSocket sends an **application-level heartbeat** (a small JSON message `{"type":"heartbeat"}`) every 10 seconds. This is distinct from the transport-level ping/pong already wired in `websockets.go:128,168` (`pongWait = 60s`); the existing ping/pong keeps the TCP socket alive but does not carry the playerIndex we need to track per-player liveness. We add the application heartbeat as an in-band message type.

**Socket struct change**: `websockets.go:56-61` defines `socket` with `gameID` but no `playerIndex`. We add `playerIndex boardgame.PlayerIndex` to the struct, populated at socket-handler entry (`server/api/main.go:1816`) from the existing `effectivePlayerIndex` lookup in `context.go`. This is a real edit, not a no-op; phase 3 includes it.

Server-side tracking: `lastHeartbeat map[gameID]map[PlayerIndex]time.Time`. A player is **absent** when `time.Since(lastHeartbeat) > 30s`. Tracking lives inside the existing `versionNotifier`'s goroutine (`server/api/websockets.go:46-54`), alongside `register`/`unregister`/`notifyVersion`. A new `chan heartbeat` is processed in the same select loop — no mutex, matching the framework's channel-based concurrency idiom.

A periodic ticker (every 5s, also in the same goroutine) scans `lastHeartbeat` for stale entries and emits presence-change events. `lastHeartbeat` evicts game entries when a game reaches `Finished`.

Clients receive presence as `Absent []PlayerIndex` on the state JSON returned by `gameInfoHandler` (`server/api/main.go:1273`). The Table view renders "Waiting for Alice…" badges over absent seats.

### 9.2 Host = Game Creator on Table Surface

The host is identified by `eGame.Owner == currentUser` AND `surface_<gameID>=table` cookie. Host privileges:

- See the Skip-Turn button on the absent-player badge.
- See the "Lock room" toggle.
- Audit log: V1 logs all host actions to the server logger with structured fields (action, gameID, userID, result). A dedicated `companionHostAudit` storage table is deferred to a future polish pass if log-based diagnostics prove insufficient.

### 9.3 SkipTurn

The only V1 host action.

`POST /api/game/<id>/hostSkipTurn` (no `?player=` query parameter — the action only ever targets the **current** player; see "Skip-on-non-current behavior" below):

- Gated on `IsHost(currentUser, gameID)`.
- Server proposes the new `moves.ForceFinishTurn` move with `proposer = AdminPlayerIndex` (existing pattern: server-proposed FixUp moves already use `AdminPlayerIndex` — see `game.go:568, 651` for the existing `applyMove(..., AdminPlayerIndex, ...)` and `ProposeMove(move, AdminPlayerIndex)` callers).
- Audited.
- Rate-limited to 1/sec per `(gameID, hostUserID)`.

#### Why a new move, not vanilla `moves.FinishTurn`

The existing `moves.FinishTurn` (`moves/finish_turn.go:25`) is the framework's turn-advance move, BUT `FinishTurn.Legal()` (line 46-76) only permits the move when `TurnDone()` returns `nil` — i.e., the current player has voluntarily satisfied their turn-end condition (played the required card, made the required decision, etc.). A player who drops mid-turn typically has `TurnDone() != nil`, so a host SkipTurn that just proposes `FinishTurn` is rejected as illegal. The game would be stuck — host's button would do nothing.

To fix this without forcing every game author to write game-specific moves, this spec adds:

```go
// moves/force_finish_turn.go
type ForceFinishTurn struct {
    FinishTurn  // embed for behavior reuse
}

func (f *ForceFinishTurn) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    // Bypasses TurnDone() — only AdminPlayerIndex (server-proposed) may invoke
    if proposer != boardgame.AdminPlayerIndex {
        return errors.New("ForceFinishTurn can only be proposed by AdminPlayerIndex")
    }
    return nil
}

// Apply is inherited from FinishTurn, which already calls ResetForTurnEnd on
// the current player's PlayerTurnFinisher (moves/finish_turn.go:90). No
// override needed — the embedded behavior does the right thing once Legal()
// allows the move through.
```

Note: `FinishTurn.Apply` already calls `ResetForTurnEnd` via the `interfaces.PlayerTurnFinisher` interface (`moves/interfaces/main.go:78`); we explicitly do NOT override `Apply` in `ForceFinishTurn` so that we don't double-invoke. The embed cleanly inherits the behavior.

Server-side `hostSkipTurn` proposes `ForceFinishTurn` with `proposer = AdminPlayerIndex`. Existing `moves.FinishTurn` is unchanged.

#### Skip-on-non-current-player behavior

A common case: player Bob is absent, but it's Carol's turn (not Bob's). Host taps Bob's badge — what should happen?

**Answer**: the Skip button only appears on the badge of the **current** player when they're absent. Non-current absent players display a "Waiting…" indicator (just informational) with no Skip button. Rationale: there's nothing meaningful to "skip" for a player who isn't currently making decisions. Once it becomes Bob's turn and he's still absent, the Skip button appears on his badge.

This matches the existing `ForceFinishTurn` semantics (which target the current player) and avoids the ambiguity of "skip until this player would have been current."

#### Phase-based games (Werewolf etc.) — out of V1

Earlier drafts proposed a sibling `ForceAdvancePhase` move for games using `moves.StartPhase`. We dropped it from V1 because phase advancement is intrinsically game-specific — `StartPhase` requires `WithPhaseToStart` config (see `moves/start_phase.go:86-94`) and a generic framework move cannot know which phase comes next without delegate-specific knowledge.

For V1, phase-driven games where an absent player blocks phase advance fall into the same "no recourse" bucket as simultaneous-action games (§2 Non-Goals): the game pauses indefinitely on the absent player; host can wait, switch-to-solo, or end the game. V2 can add a delegate hook for game-specific phase-skip moves once we have actual deployed phase-driven games to drive the design.

### 9.4 Host Transfer

**Host privileges belong to one fenced Table-device lease**, not to the game
owner, a user override, or the JavaScript-readable `surface_<gameID>` renderer
preference. The server gives the initial Table a random HttpOnly credential and
stores only its digest, generation, holder audit ID, and expiry. Its socket
renews that lease on immediate and ten-second heartbeats.

Every host action validates the unexpired persisted lease and its credential.
After the reconnect grace period expires, the owner or any seated player may
use the framework-owned Hand recovery control, which calls
`POST /api/game/<name>/<id>/tableLease/acquire`. Storage compares and swaps the
lease generation, so simultaneous claimants have exactly one winner across
processes and stale credentials are fenced out. The winner becomes the Table;
the displaced screen receives a terminal paused state. `CompanionHostOverride`
and `/claimHost` no longer exist.

Renderer selection remains presentation-only. It can choose a Table module for
developer testing, but cannot grant host controls or mutate host state.

### 9.5 Reconnection

Three flavors, all framework-handled:

- **Same browser, refresh**: cookie still valid → existing `calcViewingAsPlayerAndEmptySlots()` returns same seat. Heartbeat resumes; absent flag clears.
- **Same browser, new tab**: cookie shared → same seat. Both tabs send heartbeats; presence is live as long as either does.
- **Truly new browser**: anon UID is gone unless Firebase IndexedDB persistence restores it. If restored, same seat (resolution path checks `if anonUID in UserIDsForGame() → restore` before falling back to next-open-seat). If not restored, phone re-enters the room code and is reassigned to the next open seat.
- **Same browser reused for a different room**: a phone that played in game A and then goes to `/join` to join game B will carry a stale `surface_<gameA_ID>=hand` cookie. Because cookies are per-gameID (the cookie name includes the gameID), this cookie is invisible to game B (which uses `surface_<gameB_ID>=hand`). The new cookie for game B is issued on `/api/join/seat` for game B without conflict. The old cookie expires naturally when game A reaches `Finished` + the 24h grace period elapses, OR is cleared explicitly if the user invokes `switchToSolo` on game A.

### 9.6 Switch to Solo (Mode Downgrade)

If the projector setup fails mid-game (HDMI handshake drops, AirPlay disconnect, etc.) and the group wants to fall back to solo-device-per-player without losing state: the host can invoke `POST /api/game/<id>/switchToSolo`.

**⚠️ User-facing warning at the confirm step**: "Switching to solo mode will end the shared-screen mode for this game. Each player's phone will start showing the full game view, which **may reveal hidden information** (cards, roles) that was previously private. This change cannot be undone for the current game." The confirm requires a deliberate two-tap (initial button → confirm-the-warning).

This is the price of switching mid-game in a hidden-info game: the solo renderer typically renders all public state + the viewing player's private slice, but a phone shoved next to a neighbor's eyes makes "your private slice" much less private than the Hand view did. For pre-game switching, the warning is unnecessary (no info to leak yet); the spec still surfaces it for consistency.

#### Concrete plumbing — four pieces

The earlier draft said cookies are "invalidated via the WebSocket." That's actually two pieces of plumbing the framework doesn't have today. Here's what's actually needed:

1. **Server-side state change**. Endpoint clears `eGame.CompanionRoomCode = ""` (which is the V1 "not in companion mode" sentinel) and persists it. (One existing pattern — same kind of write as setting `Finished`.)

2. **New WebSocket message type `mode-changed`**. Currently the framework defines only `"version"` and `"chat"` socket message types (`server/api/websockets.go:19-22`). This spec adds a third: `{ "type": "mode-changed", "data": { "newMode": "solo" } }`. Broadcast to every socket in the game's bucket via the existing `versionNotifier` channel-based mechanism.

3. **Client-side handler**. `boardgame-game-state-manager.ts` (or equivalent socket router on the client) gets a new branch for `"mode-changed"` that calls `window.location.reload()`. No state surgery — just a clean reload.

4. **Client-side cookie clearing on reload**. When `boardgame-game-state-manager.ts` receives the `mode-changed` message, it calls `_clearAllSurfaceCookies()` (which iterates `document.cookie`, finds all `surface_*` keys, and sets them to `Max-Age=0`) before calling `location.reload()`. After reload the loader reads no `surface_<gameID>=*` cookie and loads the solo renderer. The server also clears the host's own `surface_<gameID>` cookie in the switchToSolo HTTP response. Phones that are offline when the switch happens will carry a stale `surface_<gameID>=hand` cookie; on their next `doGameInfo` fetch the server detects `CompanionRoomCode == ""` and the game falls through to the solo renderer (stale cookie no-ops).

`seatPresentation` rows are preserved (the avatars still apply in the solo view's "who's playing" UI). Phones still bound to seats stay bound — they just now see the solo renderer.

V1 does NOT support the reverse direction (upgrading a solo game into Table+Hand mid-session) because solo-mode players aren't seat-bound the same way.

## 10. Avatar / Name Picker

Modeled on word-bloom's publisher avatar picker (see `/Users/jkomoros/Code/word-bloom/src/components/publisher-avatar-picker.ts`). 4-step shape:

- **random** (front door): one fully-randomized avatar + name. Big "Looks good — join!" button. Small "Try another" reroll. Small "Customize" link.
- **primary**: grid of 12 primaries.
- **style**: decoration + corner + tint pickers.
- **review**: confirm + edit name.

Most users tap "Looks good" on step 1.

Composite id format (mirrored from word-bloom for consistency):

```
${primaryId}-${decorationId|none}-${cornerId|none}-${tintId|none}
```

Catalog lives in `server/static/src/components/companion-avatar-primaries.ts` etc. Not shared with word-bloom literally — separate products — but the format and starter catalog (12 primaries, 12 decorations, 4 corners, 8 tints) are mirrored.

Name generator: adjective + animal from a small curated list (200 + 200 = 40,000 combos). Generated client-side. Editable at the review step. Length 2–24, ASCII per §6.6.

Persistence: lives in `seatPresentation`, NOT the user `StorageRecord`. Clearing a seat (V2 Free-Seat) clears the row.

## 11. Edge Cases

| Case | Behavior |
|------|----------|
| Room code typo | 404 / "Room not found". Rate-limited per IP. |
| Two phones race for last seat | Per-game lock + 409 to loser; loser retries against fresh snapshot. |
| Host's projector closes | Heartbeat goes stale (30s) → any seated player may claim host. |
| Anon UID token expires (1h) | Silent re-auth via Firebase SDK. If offline, restored from IndexedDB on next online attempt; same seat as long as still bound. |
| Phone went to sleep for 2h | Heartbeat went stale → player flagged absent → on wake, heartbeat resumes → absent flag clears. |
| Game reaches `Finished` | Room code enters 24h grace before recycling. Phones see "Game over". |
| Same Google user opens two phones | First takes the seat; second is observer (existing behavior). |
| State-fetch slower than `serverPlayAt` lead | Animation plays immediately on the receiving client. Cross-surface drift visible; non-blocking. |
| Network partition phone↔server | Heartbeat stales; absent flag raised; reconnect restores. |
| Display-name collision in same game | 409 + suggested suffix; phone retries silently. |
| Display-name has zero-width/RTL chars | 400 + generic rejection; phone shows reroll. |

## 12. UI Details

### Table View
- Dark theme preferred; legible from 8+ feet.
- Room code typography: ≥120pt while pre-game.
- QR code (25% of viewport) in a corner.
- Avatar strip across the top: avatar + name + presence indicator (pulse if current player; faded if absent).
- "Waiting for Alice (1:42)" badge over absent avatar; SkipTurn button visible only to host.
- Fake-deck row along bottom edge (auto-rendered): one stub per seated player, evenly spaced.

### Hand View
- Mobile-first portrait.
- Top: tiny game-state ribbon (current player, phase, score).
- Middle: this player's hand, fanned/stacked per game's renderer.
- Bottom: action buttons.
- Cards animate from the top edge on incoming, to the top edge on outgoing.

### No Cross-Surface Audio in V1
Per Jackbox research, audio cues bridge two screens well but we defer to V2. V1 is silent.

## 13. Testing Strategy

### 13.1 Unit Tests

- `room_code_test.go`: alphabet, collision retry, recycling.
- `presence_test.go`: heartbeat staleness, absent flag, reconnect.
- **`inflater_outer_tag_override_test.go`**: confirms the new inflater extension (§6.3.2) — given a struct that embeds a behavior with an inner `sanitize:"…"` tag AND an outer `sanitize:"…"` tag at the embedding site, the outer tag wins. Tests against synthetic fixtures, not against `PlayerRole` specifically, so the inflater is verified independently of the behavior.
- `role_sanitize_default_test.go`: confirms the new `sanitize:"other:hidden"` default on `PlayerRole.Role` and `PlayerTeam.Team` produces the expected projector vs hand-view sanitization difference. Includes a fixture game with PlayerRole, asserts `ObserverPlayerIndex` JSON omits the role property while `PlayerIndex(n)` JSON includes it.
- `role_sanitize_public_override_test.go`: same fixture but with `sanitize:"all:visible"` override at the embedding site; confirms projector now sees the role. Depends on the inflater extension passing first.
- `force_finish_turn_test.go`: confirms `moves.ForceFinishTurn.Legal` accepts `AdminPlayerIndex` proposers and rejects others; confirms `Apply` delegates correctly to `FinishTurn.Apply` (which invokes `ResetForTurnEnd` via `interfaces.PlayerTurnFinisher`); confirms no double-invocation of reset.
- `host_actions_test.go`: SkipTurn gated on `IsHost` (table-surface check included); rate-limited; audited; non-host rejected; phone-only Owner cannot skip. **Skip button is only available when the absent player IS the current player** (per §9.3).
- `host_transfer_test.go`: claim succeeds after 30s of stale Owner heartbeat; rejects if Owner is fresh; ties broken by lowest player index; promoted player retains host on Hand surface.
- `switch_to_solo_test.go`: confirms `switchToSolo` clears `CompanionRoomCode`, broadcasts the new `"mode-changed"` WebSocket message, and client-side handler calls `_clearAllSurfaceCookies()` before reload.
- **`inter_player_gift_privacy_test.go`** (security regression): set up a 3-player game where each player has a private `Hand` stack with `sanitize:"order:none"` (component IDs hidden from observers). Move a card from player 3's hand to player 2's hand. Confirm the `ObserverPlayerIndex` view of both before and after states sees only obscured component IDs in the fake-deck stub elements — there is no way the Table-view FLIP animation can reveal which specific card was transferred. This regression covers the case where a sanitization mistake on a future game could leak via the cross-screen animation.

### 13.2 Integration Tests

- Multi-client websocket scenario: Table + 2 Hands connect; one Hand disconnects; presence updates push; SkipTurn advances (via `moves.ForceFinishTurn` rather than vanilla `FinishTurn`); Hand reconnects.
- Anonymous join → seat → state-version-change notify → state-fetch → move → notify → fetch. Compare Table and Hand state JSONs to confirm sanitization differences.
- Public-visibility role game: confirm role appears on Table view's seat picker. Private-visibility role game: confirm role does NOT appear on Table view, only on Hand view.
- New `"version-timing"` WebSocket message arrives after `"version"` for state changes; carries valid `serverSentAt`/`serverPlayAt`. Old `"version"` payload shape unchanged (bare int). Backward-compat: a stubbed "old client" that only handles `"version"` doesn't break.
- `animateBetween(realId, stubId)` correctly animates a single card via synthetic-ID stub without colliding with the real `component.id` in the animator's `_infoById` map.
- `switchToSolo` end-to-end: Table host invokes endpoint → all sockets receive `"mode-changed"` → reload → cookies cleared → solo renderer loads.

### 13.3 Browser / E2E (Playwright)

- Headless run of `werewolf` (or chosen MVP game) with Table tab + 3 Hand tabs.
- Verify room code visible, code entry works, avatars appear on Table after join.
- Verify a deal: capture before/after screenshots on both Table and Hand; assert card flies on each.
- Verify drop → absent → host SkipTurn → game advances.

### 13.4 Manual Playtest

One real session with 4 humans before V1 ships.

## 14. Phasing

### Phase 1 — Foundations (PR 1)
- `CompanionRoomCode` + `CompanionLocked` fields on `extendedgame.StorageRecord` + `GameByRoomCode` storage method (with four-place storage implementation: memory/bolt/mysql + test harness; MySQL schema migration).
- `/api/join` + `/api/join/seat` endpoints with HTTPS-only, display-name validation.
- New token-bucket rate-limiting middleware (`middleware_ratelimit.go`); apply to `/api/join` and `/api/join/seat`.
- Firebase anonymous auth flow on the phone.
- `seatPresentation` storage table + accessors (same four-place implementation).
- Build-time filesystem capability walk in `boardgame-util/cmd_serve.go` and `cmd_build_static.go`; emit `companion_capable_games` into config.json (for api binary) AND `table_hand_supported_games` into client_config.js (for browser). Server-side `managerInfo.supportsTableHandMode` populated from config.json at boot; surfaced via `doListManager`.
- Surface cookie routing in `boardgame-render-game.ts`.
- `BoardgameTableViewBase` + `BoardgameHandViewBase` (minimal — just typed props + helper renders).

### Phase 2 — Identity + Seat Picker (PR 2)
- **Inflater extension (§6.3.2)**: extend `StructInflater` in `struct_inflater.go` to walk outer embedding-site tags first, falling back to inner defaults. ~30-50 LOC + tests. **Load-bearing** for the public-role opt-in mechanism.
- Add default `sanitize:"other:hidden"` to `behaviors.PlayerRole.Role` (at `behaviors/role.go:16`) and `behaviors.PlayerTeam.Team` (at `behaviors/team.go:32`).
- Audit confirmation: grep shows zero games currently embed `PlayerRole`/`PlayerTeam`. If any new games landed before this PR, audit them and add explicit `sanitize:"all:visible"` overrides where current behavior was intentional.
- Seat picker UI on phone (`/api/join/seat-options` endpoint).
- Avatar/name picker (random front door + customize) — depends on the avatar catalog being committed.

### Phase 3 — Presence + Host (PR 3)
- Thread `playerIndex` into the `socket` struct (`websockets.go:56-61`); populate from `effectivePlayerIndex` at handler entry.
- In-band JSON heartbeat message + server-side `lastHeartbeat` map folded into the existing notifier goroutine via a new `chan heartbeat`.
- "Waiting for Alice…" projector affordance.
- `moves.ForceFinishTurn` (new file `moves/force_finish_turn.go`).
- Host `hostSkipTurn` endpoint — Skip button appears ONLY on the current player's badge when absent (§9.3).
- Fenced Table-device lease + atomic Hand recovery + audit logging from host endpoints.
- Host = table-surface requirement encoded in `IsHost`.

### Phase 4 — Animations (PR 4)
- Fake-deck row helper render on Table view base (synthetic-ID stub elements).
- New animator API: `animateBetween(realId, stubId)` in `boardgame-component-animator.ts`. Scope: ~50-80 LOC; a new code path that takes element references directly and runs mini-FLIP outside the main collection iteration (the existing `prepare()` only walks `_sharedStackList` so synthetic stubs outside the stack tree need their own measure path).
- Top-edge off-screen anchor on Hand view base.
- New WebSocket message type `"version-timing"` carrying `serverSentAt`/`serverPlayAt`; broadcast sibling to existing `"version"` message (backward compatible — old clients ignore the new type). Client-side minimum-wins one-way estimator.
- MVP game's `*-table.ts` and `*-hand.ts` shipped (likely `werewolf`).
- End-to-end deal animation working across surfaces.

### Phase 5 — Polish + V1 Ship (PR 5)
- QR code on Table view.
- Visual polish, dark theme.
- Room-lock toggle UI.
- `switchToSolo` host action (four-piece plumbing per §9.6): server-side flag, new `"mode-changed"` WebSocket message type, client reload handler, cookie clearing via Set-Cookie Max-Age=0 on reload.
- 4+ human playtest.

### Out of V1 (deferred)
- Free-Seat host action.
- Replace-with-AI host action (existing Agent system covers some of this already at game-create time).
- Phase-skip host action (`ForceAdvancePhase` or per-delegate hook) for games using `moves.StartPhase`.
- Skip host action for simultaneous-action games (those using `behaviors.PlayerSubmission`).
- Audio cues.
- Multi-projector.
- Spectator/audience tier.
- Anonymous → Google account upgrade UI.
- Cross-session persistent avatar.
- Non-card private info animation (hidden numbers, secret meeples).
- Non-left-to-right seat layouts (circular, around-a-board).
- Improved clock sync beyond the minimum-wins estimator.
- Mid-game upgrade from solo → Table+Hand (Switch-to-Solo is one-way only in V1).

## 15. Open Questions

Most have been resolved by the reshape and the implementability fix-up. Remaining:

1. **MVP game**: assumed `werewolf` based on the untracked `server/static/game-src/werewolf/` directory. Confirm before Phase 4.
2. **Avatar primaries catalog**: this spec proposes mirroring word-bloom's `(primary, decoration, corner, tint)` 4-tuple format. The actual catalog content (which 12 primaries? which 12 decorations?) is a separate art-direction task. Phase 2 cannot land until the starter catalog is committed.
3. **Slur/abuse word filter for display names**: V1 ships with NFKC + ASCII validation only. Hook point exists; deciding whether to add a blocklist before V1 or after first incident.
4. **Existing-game sanitization migration**: Phase 2 audits all games embedding `PlayerRole`/`PlayerTeam` and adds explicit `sanitize:"all:visible"` where current behavior should be preserved. Confirm the audit list before Phase 2: at minimum any examples in `examples/` plus any games under `server/static/game-src/` that use either behavior. Spot-check via grep.

## 16. Glossary

- **Table view**: the shared screen (laptop, TV, tablet) connected as `ObserverPlayerIndex` with the `-table.ts` renderer.
- **Hand view**: a phone connected as `PlayerIndex(n)` with the `-hand.ts` renderer, bound to a seat.
- **Solo mode**: existing single-device-per-player flow. Untouched by this work; can be switched into mid-game via Switch-to-Solo (§9.6).
- **Host**: `eGame.Owner` (or transferred override) viewing on the Table surface. Phone-only Owner is not host (§9.4).
- **Fake deck row**: framework-rendered row of off-screen-anchored stub stacks along the bottom of the Table view, one per seated player in left-to-right seat order. Stubs use synthetic IDs (`stub:pN:<realId>`) to avoid `component.id` collision in the FLIP animator.
- **`animateBetween(realId, stubId)`**: new public method on `BoardgameComponentAnimator` that drives cross-position animation using a synthetic-ID stub element as the source or destination anchor.
- **Room code**: 4-letter code on `extendedgame.StorageRecord`. Resolved via `GameByRoomCode`.
- **`seatPresentation`**: per-`(gameID, playerIndex)` storage of display name + avatar. Source of truth for what a seat looks like in this game.
- **`sanitize:"other:hidden"`** (default on `PlayerRole.Role` / `PlayerTeam.Team` as of this spec): hides the property from `ObserverPlayerIndex` (Table view) while leaving it visible to the player's own Hand view. Override at the embedding site with `sanitize:"all:visible"` for public-role games.
- **`moves.ForceFinishTurn`**: new framework move that bypasses `TurnDone()` to advance the current player when a host invokes SkipTurn for an absent player. Only accepts `AdminPlayerIndex` proposers (i.e., server-initiated).
- **`"version-timing"` (WebSocket message type)**: new message type carrying `serverSentAt` and `serverPlayAt` timestamps. Broadcast as a sibling to the existing `"version"` message; old clients ignore it. Backward-compatible.
- **`"mode-changed"` (WebSocket message type)**: new message type signaling that the game's mode flipped (e.g., Table+Hand → solo via `switchToSolo`). Client handler triggers a reload.

## 17. What a Game Author Actually Writes — Worked Example

For `werewolf` (or any hidden-info game) to opt into Table+Hand mode:

1. Create `server/static/game-src/werewolf/boardgame-render-game-werewolf-table.ts`:

```ts
import { html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { BoardgameTableViewBase } from '../../src/components/boardgame-table-view-base.ts';
import type { GameState, PlayerState } from './_types.ts';

@customElement('boardgame-render-game-werewolf-table')
export class TableView extends BoardgameTableViewBase<GameState, PlayerState> {
    static styles = css`/* big-screen styles */`;

    render() {
        return html`
            ${this.renderAvatarStrip()}
            ${this.renderHostControls()}
            <div class="board">
                <boardgame-component-stack .stack=${this.state.PublicDeck}></boardgame-component-stack>
                <!-- ...rest of public board... -->
            </div>
            ${this.renderFakeDeckRow()}
        `;
    }
}
```

2. Create `boardgame-render-game-werewolf-hand.ts`:

```ts
import { html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { BoardgameHandViewBase } from '../../src/components/boardgame-hand-view-base.ts';

@customElement('boardgame-render-game-werewolf-hand')
export class HandView extends BoardgameHandViewBase<GameState, PlayerState> {
    static styles = css`/* mobile-first styles */`;

    render() {
        return html`
            <div class="hand">
                <boardgame-component-stack .stack=${this.playerState.Hand} layout="fan">
                </boardgame-component-stack>
            </div>
            <div class="actions">
                <button @click=${this.onPlay}>Play</button>
                <button @click=${this.onPass}>Pass</button>
            </div>
        `;
    }
}
```

3. (If roles should appear on the Table view, which werewolf doesn't): override the sanitize tag on the embedded behavior — `behaviors.PlayerRole \`sanitize:"all:visible"\`` — at the embedding site in your `playerState` struct. Otherwise: do nothing else.

That's the whole opt-in. Two files, roughly the size of the existing solo renderer combined. The framework handles pairing, identity, sanitization, animations, presence, host controls, sync.
