# Table + Hand Mode (Per-Person Mobile UI)

Spec date: 2026-05-23 (third revision — opinionated reshape after game-author DX critic pass)
Ticket: [#759 — Allow remote control / projected mode](https://github.com/jkomoros/boardgame/issues/759)
Branch: `per-person-mobile-ui`

## Revision Notes

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

The server walks `server/static/game-src/` at boot, notes which games ship both files, and surfaces a "Use shared projector + phones" toggle on the game-creation form. That's the entire opt-in.

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

By default, when a game has `behaviors.PlayerRole` or `behaviors.PlayerTeam`, the Table view does NOT show roles — players appear as numbered seats with avatars only. The role/team is revealed only on the Hand view of the holding player.

A game where roles are *meant to be public* (Codenames: spymasters are publicly known) opts in by setting a field on the behavior:

```go
type playerState struct {
    base.SubState
    behaviors.PlayerRole `visibility:"public"`  // struct-tag opt-in; default is "private"
}
```

Or programmatically:

```go
type playerState struct {
    base.SubState
    Role behaviors.PlayerRole
}

func (p *playerState) ConnectBehavior(s boardgame.SubState) {
    p.Role.Visibility = behaviors.VisibilityPublic
    p.Role.ConnectBehavior(s)
}
```

Identical mechanism for `behaviors.PlayerTeam`.

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
│   • behaviors.Visibility field on PlayerRole / PlayerTeam           │
│   • Heartbeat-based presence in the existing notifier goroutine     │
│   • ServerSentAt/ServerPlayAt piggybacked on state pushes           │
│   • Host SkipTurn endpoint                                          │
│   • Host transfer (claim if owner heartbeat is stale)               │
└─────────────────────────────────────────────────────────────────────┘
              │                  │              │
   ObserverPlayerIndex     PlayerIndex(0)   PlayerIndex(1)
   + surface=table cookie  + surface=hand cookie    ...
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

## 5. Mode Detection (Filesystem Walk)

At server boot, after `boardgame.NewServer(storage, delegates...)` registers each `GameDelegate`, the server walks `server/static/game-src/<delegate.Name()>/` and records on the `GameManager` whether both `boardgame-render-game-<name>-table.ts` AND `boardgame-render-game-<name>-hand.ts` are present. This capability map is exposed by `Manager.SupportsTableHandMode() bool` (read by the game-creation form to show/hide the toggle).

In dev (`OfflineDevMode`), the walk re-runs on every page load so newly-added games appear without a server restart.

In production (Vite/static-built deployment), the build process emits a manifest file `companion-capability.json` listing the games with both files; the server reads the manifest at boot instead of walking. Same capability map either way.

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
6. Issues a `surface=hand` cookie scoped to the gameID.

Race resolution: if two phones race for the last seat, the loser receives 409 Conflict + the latest seat-availability snapshot and retries against it.

### 6.3 Seat Picker for Asymmetric Games

The framework auto-detects asymmetry by type-asserting the game's `playerState` against `behaviors.HasPlayerRole` or `behaviors.HasPlayerTeam`. (The framework adds `behaviors.HasPlayerTeam` as a one-method interface `GetPlayerTeam() *PlayerTeam`, paralleling the existing `HasPlayerRole`; `PlayerTeam` gains a one-line `GetPlayerTeam() *PlayerTeam { return p }` so the assertion works.)

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

For each slot, the label is computed server-side based on `behaviors.Visibility`:

| Visibility on PlayerRole/PlayerTeam | Slot label                     |
|-------------------------------------|--------------------------------|
| `Private` (default)                 | "Seat N"                       |
| `Public`                            | "<Role display name>" (e.g., "Spymaster (Red)") |

The choice of `seatPick` becomes the player's `playerIndex` on the seat endpoint.

**Why the framework owns label resolution**: the layered alternative (delegate-supplies-label-string) was the source of the projector role-leak issue in earlier drafts. By making visibility a field on the behavior itself, there is no layer-spanning policy to forget.

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

### 6.6 Display-Name Validation

Server-side, on `/api/join/seat`:

- NFKC-normalize the input.
- Allow only `[A-Za-z0-9 ]`, length 2–24 inclusive after trim.
- Reject zero-width, RTL-override, combining diacritics, C0/C1 control, surrogate ranges.
- Reject if the lowercased normalized name matches another seat's name in the same gameID. The phone retries with a suggested suffix (e.g., "Alice" → "Alice2").
- No slur catalog in V1; hook point exists for one.

## 7. Renderer Model

### 7.1 Surface Selection (Cookie-Based)

Surface is decided at join/create time and stored in a `surface=table` or `surface=hand` cookie scoped to the gameID. URL stays clean (`/game/<name>/<id>`). The loader (`boardgame-render-game.ts:365`) reads the cookie:

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
    @property() serverPlayAt: number | null;        // for animation timing

    // Helper renders — call these from your render() where you want them
    protected renderAvatarStrip(): TemplateResult { ... }
    protected renderHostControls(): TemplateResult { ... }   // SkipTurn button when applicable
    protected renderFakeDeckRow(): TemplateResult { ... }    // §8

    // FLIP animator wired automatically; respects serverPlayAt for cross-screen sync.
}
```

Helper renders are *opt-in to call*, not invoked automatically. The author composes their own layout and calls the helpers where they want them. No bundled inheritance — if the author wants a custom avatar strip, they don't call `renderAvatarStrip()`.

### 7.3 Hand View Base

```ts
export class BoardgameHandViewBase<GS, PS> extends LitElement {
    @property() state: FullGameState<GS, PS>;       // PlayerIndex(n)-sanitized
    @property() viewingAs: PlayerIndex;
    @property() seatPresentations: SeatPresentation[];
    @property() serverPlayAt: number | null;

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
- `serverPlayAt` scheduling (auto).
- Presence indicators (`absentPlayers` is just a prop; author renders it however they want, or uses the avatar-strip helper which renders the standard treatment).

Things deliberately *not* in the bases — kept as Lit `ReactiveController`s or out-of-class entirely so they're opt-in composable, not inherited bundles:

- Score ribbon — not all games have a global score; let authors render their own ribbons.
- Phase indicators, turn timers — game-specific; bases don't impose a layout.
- Custom animations beyond FLIP — author hooks into the existing animation system directly.

This is the answer to the layering critic's "BoardgameSurfaceRendererBase bundled inheritance" complaint: each base exposes a *small* set of typed properties + a *small* set of opt-in helper renders, and nothing else. Authors compose, not subclass-and-pray.

## 8. Animation — Fake Deck Row Convention

The framework provides a single convention for cross-screen card animations that game authors don't have to think about.

### 8.1 The Mechanism

On the Table view, the framework auto-renders a "fake deck row" along the bottom edge (via `renderFakeDeckRow()` when the author calls it, or auto-added as a sibling if not). The row has one stub stack per seated player, in seat order, evenly spaced left-to-right.

Each stub is a hidden stack mirror of that player's private state — same `component.id`s as the player's actual Hand view stack, but rendered off-screen / underneath the avatar at the bottom edge. The FLIP animator on the Table side sees the cards present in the stub when the server pushes a state where they belong to player N's hand.

**Result**:
- Deal from public deck → player N's hand: card animates from on-screen deck to the bottom-row stub at seat N's position. Visually: card flies down toward the player.
- Player M plays a card to a public discard: card animates from the bottom-row stub at seat M to the on-screen discard. Visually: card flies up from the player to the table.
- Card moves from player 3's hand to player 2's hand: card animates between fake-deck-3 and fake-deck-2 along the bottom — horizontally, left or right based on relative seat position. Visually: card slides between players.

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

Every WebSocket frame includes a server timestamp piggybacked in `Data` for state-push messages:

```go
type stateUpdateData struct {
    // ... existing state fields ...
    ServerSentAt int64 `json:"serverSentAt"`  // ms since epoch, set immediately before write
    ServerPlayAt int64 `json:"serverPlayAt"`  // ms since epoch; serverSentAt + 250ms default
}
```

Client maintains a minimum-wins one-way latency estimator over the last 30 frames:

```js
const oneWayMs = (localRxAsEpochMs) - serverSentAt;
// minOneWayMs over rolling window is the offset
```

Animations schedule via `setTimeout(playFn, localEquivalent(serverPlayAt) - now)`. Falls back to "play immediately on receive" if the play-at instant is already past or fewer than 3 samples have been collected.

**Documented limitations**:
- Asymmetric routes (cell + Wi-Fi) bias the estimator but the bias is consistent per surface; visible cross-surface drift is bounded by the asymmetry, not by the variance.
- JS GC pauses and background-tab throttling can shift `setTimeout` firing by 50-200ms; we accept this in V1.
- First state push beats the sample window; first deal animation may be visibly uncoordinated.

V1 ships this minimal primitive; future iteration can layer in explicit clock sync rounds if playtests show the median-LAN baseline is insufficient.

## 9. Presence & Host Controls

### 9.1 Presence (Heartbeat-Based)

Each Hand-view WebSocket sends a heartbeat ping every 10 seconds. Server tracks `lastHeartbeat map[gameID]map[PlayerIndex]time.Time`. A player is **absent** when `time.Since(lastHeartbeat) > 30s`.

Tracking lives inside the existing `versionNotifier`'s goroutine (`server/api/websockets.go:46-54`), alongside `register`/`unregister`/`notifyVersion`. A new `chan heartbeat` is processed in the same select loop — no mutex, matching the framework's channel-based concurrency idiom.

A periodic ticker (every 5s) scans for stale entries and emits presence-change events. WebSocket read deadlines (30s) close idle sockets so the OS-keepalive default doesn't leave zombies. `lastHeartbeat` evicts game entries when a game reaches `Finished`.

The state pushed to clients includes `Absent []PlayerIndex` so the Table view can render a "Waiting for Alice…" badge over the absent seat's avatar in the strip.

### 9.2 Host = Game Creator on Table Surface

The host is identified by `eGame.Owner == currentUser` AND `surface=table` cookie. Host privileges:

- See the Skip-Turn button on the absent-player badge.
- See the "Lock room" toggle.
- Audit log: all host actions recorded to a new `companionHostAudit` table for diagnostics.

### 9.3 SkipTurn

The only V1 host action.

`POST /api/game/<id>/hostSkipTurn?player=N` (camelCase to match existing route style at `server/api/main.go:944,1072,1301`):

- Gated on `IsHost(currentUser, gameID)`.
- Server proposes the existing `moves.AdvanceCurrentPlayer` FixUp move (or equivalent — the framework's standard turn-advance). No game-specific opt-in required.
- Audited.
- Rate-limited to 1/sec per `(gameID, hostUserID)`.

Games whose turn structure doesn't fit "advance current player" (e.g., simultaneous-action games) have host-SkipTurn surfaced as a no-op for V1; future iterations can plumb in game-specific fallbacks via a single optional delegate method. Most turn-based games work out of the box.

### 9.4 Host Transfer

If `eGame.Owner` has no heartbeat-fresh Table connection for 30s, any seated player may claim host via `POST /api/game/<id>/claimHost`. First claim wins; ties broken by lowest player index. Subsequent `IsHost` returns true for either the original Owner OR the override. If the original owner returns, both have host powers (audited). The override is durable for the rest of the session.

### 9.5 Reconnection

Three flavors, all framework-handled:

- **Same browser, refresh**: cookie still valid → existing `calcViewingAsPlayerAndEmptySlots()` returns same seat. Heartbeat resumes; absent flag clears.
- **Same browser, new tab**: cookie shared → same seat. Both tabs send heartbeats; presence is live as long as either does.
- **Truly new browser**: anon UID is gone unless Firebase IndexedDB persistence restores it. If restored, same seat (resolution path checks `if anonUID in UserIDsForGame() → restore` before falling back to next-open-seat). If not restored, phone re-enters the room code and is reassigned to the next open seat.

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
- `visibility_test.go`: `behaviors.Visibility` field affects sanitization correctly for projector vs hand views (this is the security regression test).
- `host_actions_test.go`: SkipTurn gated on `IsHost`; rate-limited; audited; non-host rejected.
- `host_transfer_test.go`: claim succeeds after 30s; doesn't succeed if owner is fresh; ties broken by lowest player index.

### 13.2 Integration Tests

- Multi-client websocket scenario: Table + 2 Hands connect; one Hand disconnects; presence updates push; SkipTurn advances; Hand reconnects.
- Anonymous join → seat → state push → move → state push. Compare Table and Hand state JSONs to confirm sanitization differences.
- Public-visibility role game: confirm role appears on Table view's seat picker. Private-visibility role game: confirm role does NOT appear on Table view, only on Hand view.

### 13.3 Browser / E2E (Playwright)

- Headless run of `murdermrmonroe` (or chosen MVP game) with Table tab + 3 Hand tabs.
- Verify room code visible, code entry works, avatars appear on Table after join.
- Verify a deal: capture before/after screenshots on both Table and Hand; assert card flies on each.
- Verify drop → absent → host SkipTurn → game advances.

### 13.4 Manual Playtest

One real session with 4 humans before V1 ships.

## 14. Phasing

### Phase 1 — Foundations (PR 1)
- `CompanionRoomCode` + `CompanionLocked` fields on `extendedgame.StorageRecord` + `GameByRoomCode` storage method.
- `/api/join` + `/api/join/seat` endpoints with HTTPS-only, rate limiting, display-name validation.
- Firebase anonymous auth flow on the phone.
- `seatPresentation` storage table + accessors.
- Filesystem capability walk at server boot.
- Surface cookie routing in `boardgame-render-game.ts`.
- `BoardgameTableViewBase` + `BoardgameHandViewBase` (minimal — just the typed props + helper renders).

### Phase 2 — Identity + Seat Picker (PR 2)
- `behaviors.Visibility` field on `PlayerRole` + `PlayerTeam` (with struct-tag support).
- `behaviors.HasPlayerTeam` interface + `GetPlayerTeam()` on `PlayerTeam`.
- Seat picker UI on phone (`/api/join/seat-options` endpoint).
- Avatar/name picker (random front door + customize) — depends on the avatar catalog being committed.

### Phase 3 — Presence + Host (PR 3)
- Heartbeat-based presence folded into the existing notifier goroutine.
- "Waiting for Alice…" projector affordance.
- Host SkipTurn endpoint + UI button.
- Host transfer + audit log.

### Phase 4 — Animations (PR 4)
- Fake-deck row auto-rendering on Table view base.
- Top-edge anchor on Hand view base.
- `ServerSentAt`/`ServerPlayAt` piggybacked on state pushes; client estimator.
- MVP game's `*-table.ts` and `*-hand.ts` shipped (likely `murdermrmonroe`).
- End-to-end deal animation working across surfaces.

### Phase 5 — Polish + V1 Ship (PR 5)
- QR code on Table view.
- Visual polish, dark theme.
- Room-lock toggle UI.
- 4+ human playtest.

### Out of V1 (deferred)
- Free-Seat host action.
- Replace-with-AI host action (existing Agent system covers some of this already at game-create time).
- Audio cues.
- Multi-projector.
- Spectator/audience tier.
- Anonymous → Google account upgrade UI.
- Cross-session persistent avatar.
- Non-card private info animation (hidden numbers, secret meeples).
- Non-left-to-right seat layouts (circular, around-a-board).
- Improved clock sync beyond the minimum-wins estimator.

## 15. Open Questions

Most have been resolved by the reshape. Remaining:

1. **MVP game**: assumed `murdermrmonroe` based on the untracked `server/static/game-src/murdermrmonroe/` directory. Confirm before Phase 4.
2. **Avatar primaries catalog**: this spec proposes mirroring word-bloom's `(primary, decoration, corner, tint)` 4-tuple format. The actual catalog content (which 12 primaries? which 12 decorations?) is a separate art-direction task. Phase 2 cannot land until the starter catalog is committed.
3. **Build-time manifest vs filesystem walk**: §5 describes both paths (dev = walk, prod = manifest). Confirm whether to ship both or just the walk in V1 (recommendation: walk-only for V1; add manifest in V2 when the deployment story is more formal).
4. **Slur/abuse word filter for display names**: V1 ships with NFKC + ASCII validation only. Hook point exists; deciding whether to add a blocklist before V1 or after first incident.

## 16. Glossary

- **Table view**: the shared screen (laptop, TV, tablet) connected as `ObserverPlayerIndex` with the `-table.ts` renderer.
- **Hand view**: a phone connected as `PlayerIndex(n)` with the `-hand.ts` renderer, bound to a seat.
- **Solo mode**: existing single-device-per-player flow. Untouched by this work.
- **Host**: by default `eGame.Owner` viewing the Table; transferable per §9.4.
- **Fake deck row**: framework-rendered row of off-screen-anchored stub stacks along the bottom of the Table view, one per seated player in left-to-right seat order. The destination/source for cross-screen card animations.
- **Room code**: 4-letter code on `extendedgame.StorageRecord`. Resolved via `GameByRoomCode`.
- **`seatPresentation`**: per-`(gameID, playerIndex)` storage of display name + avatar. Source of truth for what a seat looks like in this game.
- **`behaviors.Visibility`**: enum `Private | Public` on `PlayerRole` / `PlayerTeam` controlling whether role/team appears on the Table view. Default: `Private`.
- **`ServerSentAt` / `ServerPlayAt`**: timestamps piggybacked on state pushes; client uses them for cross-screen animation timing.

## 17. What a Game Author Actually Writes — Worked Example

For `murdermrmonroe` (or any hidden-info game) to opt into Table+Hand mode:

1. Create `server/static/game-src/murdermrmonroe/boardgame-render-game-murdermrmonroe-table.ts`:

```ts
import { html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { BoardgameTableViewBase } from '../../src/components/boardgame-table-view-base.ts';
import type { GameState, PlayerState } from './_types.ts';

@customElement('boardgame-render-game-murdermrmonroe-table')
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

2. Create `boardgame-render-game-murdermrmonroe-hand.ts`:

```ts
import { html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { BoardgameHandViewBase } from '../../src/components/boardgame-hand-view-base.ts';

@customElement('boardgame-render-game-murdermrmonroe-hand')
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

3. (If roles should appear on the Table view, which murdermrmonroe doesn't): set `visibility:"public"` on the role behavior. Otherwise: do nothing else.

That's the whole opt-in. Two files, roughly the size of the existing solo renderer combined. The framework handles pairing, identity, sanitization, animations, presence, host controls, sync.
