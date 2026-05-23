# Per-Person Mobile UI (Projected + Companion Mode)

Spec date: 2026-05-23 (revised same day after critic pass)
Ticket: [#759 — Allow remote control / projected mode](https://github.com/jkomoros/boardgame/issues/759)
Branch: `per-person-mobile-ui`

## Revision Notes

Initial draft was reviewed by four critic agents (architecture, idiomatic Go / framework conventions, robustness, security/privacy). This revision addresses every critical and significant finding:

- **Security**: role/team picker results are now sanitized through new `RolePrivacyDelegate` / `TeamPrivacyDelegate` interfaces so the projector never sees hidden roles; "Free Seat" now triggers a game-defined reset move to prevent the previous player's hand from leaking to the new joiner; HTTPS/`Secure`/`HttpOnly`/`SameSite` cookie requirements made explicit; display-name validation rules specified (NFKC, ASCII, per-game uniqueness); `/api/join` rate-limited per IP; room codes get a 24h grace period before recycling; "Lock room" toggle added.
- **Storage**: avatar+name moved off the user's `StorageRecord` into a new `seatPresentation` table per `(gameID, playerIndex)` — solves the cross-game mutation problem AND the post-Free-Seat orphan-identity problem.
- **Architecture**: layout module is now `layoutFor(state, surface)` function returning a typed `LayoutPlan` with anchor variants (stack / offscreen-stack / badge / secret-value) — handles variable player counts, dynamic stacks, and non-card private info; cross-renderer concerns moved into a shared `BoardgameSurfaceRendererBase` mixin; display-surface signal moved from URL query param to a session cookie set at join time; room code lives as a typed field on `extendedgame.StorageRecord` (no in-memory map).
- **Robustness**: presence is now heartbeat-based (30s timeout), folded into the existing channel-based `versionNotifier` goroutine — no parallel mutex; host transfer promoted from V2 to V1; per-game lock on `/api/join/seat` resolves seat-race ambiguities; clock-sync simplified to a minimum-wins one-way estimator piggybacked on the existing WebSocket message stream, with limitations documented.
- **Idiomatic**: `CompanionGameDelegate` interface split into one marker (`companion.GameDelegate { UsesCompanionMode() }`) plus seven single-purpose optional extension interfaces, matching the framework's existing `behaviors/` pattern; route names switched to the existing camelCase convention (e.g., `/game/<id>/hostSkipTurn`); `HasPlayerTeam` description clarified.
- Minor: animation lead time is per-game configurable via `AnimationLeadDelegate`; lint script now also validates `layoutFor` signature and `SurfaceRendererBase` subclassing.

Most open questions from the initial draft are now resolved inline; remaining items are listed in §16.

## 1. Overview

A new gameplay mode in which one shared screen ("the projector") shows the public game state and each player's phone ("companion") shows their private hand and actions. Cards visually fly off the projector toward a player's seat and arrive on that player's phone (and vice versa). Inspired by the Jackbox model, but applied to turn-based hidden-information board games.

Pairing is anonymous-friendly: the room shows a 4-letter code, phones go to a join URL, type the code, optionally sign in with Google (or stay anonymous), pick an avatar+name, and are bound to a seat for the session.

## 2. Goals & Non-Goals

**Goals**
- A game declares in code whether it supports companion mode. The game-creation form gates the choice; once set, the session is locked into one mode.
- The projector reuses the existing `ObserverPlayerIndex` sanitization. The phone reuses the existing per-player sanitization. No new server-side sanitization variant.
- Players can join without making an account or signing in with Google — Firebase anonymous auth is the default path. Google sign-in remains available.
- Cards animate between the projector and phones with visibly synchronized timing.
- A paired player going offline pauses the game on their turn; the host (game creator at the projector) can skip, replace with an Agent, or free the seat.
- The existing solo-device flow is untouched.

**Non-Goals (deferred)**
- Multi-projector setups (one game, two TVs).
- Spectator-on-phone mode (audience members watching, not playing).
- In-room voice chat (covered by ticket #796).
- Upgrading anonymous identity into a persistent Google account post-hoc (we'll mint the anon UID in a way that *allows* future upgrade, but the UI is not in scope here).
- Cross-game persistent avatar (each game session picks fresh).

## 3. User-Facing UX Flow

### 3.1 Creating a Companion-Mode Game

1. Host opens `/list-games` and picks a game whose `GameDelegate.SupportsCompanionMode()` returns true.
2. The "Create Game" form shows a new toggle: **"Shared projector + phones"** (default *off* — solo-device is still the default).
3. With the toggle on, the host sees: "After creating, this device becomes the projector. Players will join from their phones." Confirm.
4. On submit, server creates the game with `eGame.CompanionMode = true` and generates a 4-letter room code (e.g. `JKLB`).
5. Server sets a `surface=projector` cookie scoped to the gameID (§7.1) on the host's browser, then redirects to `/game/<name>/<id>`. The client reads the cookie and loads the projected renderer.

### 3.2 The Projector View Pre-Game

- Fullscreen-friendly layout. Big, legible room code (e.g. "Go to **boardgame.app/join** and enter **JKLB**").
- A QR code linking directly to `/join?code=JKLB`, generated client-side.
- A "Seats" panel showing N empty avatar slots (N comes from `MinNumPlayers` to `MaxNumPlayers`). As phones join, slots fill with their chosen avatar + name (from `seatPresentation`).
- A "Start Game" button enabled once `MinNumPlayers` slots are filled. Disabled before then.
- For asymmetric-role games (see §6), each filled slot shows the player's role *only if* `RoleIsPublic() == true` for that game. For hidden-role games, the slot shows "Seat N" with avatar+name only.
- A "Lock room" toggle (§5.1) for the host.

### 3.3 Phone Joining Flow

1. Phone visits `/join` (or scans the QR), enters the 4-letter code (case-insensitive). Code is validated; failure shows "Room not found".
2. Server looks up the game by code → returns game metadata to the client.
3. Phone shows an identity screen: **"Continue as guest"** (primary button) or **"Sign in with Google"** (secondary). Note: code is entered *before* identity is chosen, per design decision.
4. Guest path: Firebase anonymous sign-in fires under the hood; the client immediately shows the **avatar/name picker** (see §10).
5. Google path: existing Google sign-in flow; skip avatar picker (use display name + photo URL).
6. Once identity is resolved, the client requests a seat. Two cases (see §6):
   - Symmetric game: server auto-assigns the next open seat and returns the assignment.
   - Asymmetric game: server returns the list of available role slots; phone shows a **role picker** ("Spymaster (red team)", "Operative (blue team)", etc.); selection commits the seat.
7. Server issues a `surface=companion` cookie scoped to the gameID. Phone navigates to `/game/<name>/<id>`; the client reads the cookie and loads the companion renderer.

### 3.4 In-Game

- Projector renders the public board and an avatar strip showing presence/status per player (e.g., "Alice (thinking…)" with a subtle pulse during her turn).
- Each phone renders only its player's private hand and the action affordances (e.g., "Play this card", "Pass"). When a card is dealt, it flies off the projector edge toward the seat and arrives on the phone from the corresponding edge. When played, the reverse.
- Disconnected player: projector shows "Waiting for Alice (1:42)" + host controls (see §9).

## 4. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                              SERVER                                 │
│                                                                     │
│  Existing: GameManager, sanitization-per-player, SeatPlayer,        │
│            WebSocket version notifier (channel-based), Firebase auth │
│                                                                     │
│  NEW (additive — does not alter existing state pipeline):           │
│   • CompanionRoomCode + CompanionLocked on extendedgame row         │
│   • GameByRoomCode lookup (storage method)                          │
│   • seatPresentation table (per-(gameID,playerIndex) avatar+name)   │
│   • Heartbeat-based presence tracker (folded into versionNotifier)  │
│   • Sanitization-aware role/team visibility on projector            │
│   • ServerSentAt/ServerPlayAt fields piggybacked on state pushes    │
│   • Host-action endpoints (skip / agent / free-seat / claim-host)   │
│   • companionHostAudit log                                          │
└─────────────────────────────────────────────────────────────────────┘
                  │              │              │
   ObserverPlayerIndex     PlayerIndex(0)   PlayerIndex(1)
   + surface=projector cookie   + surface=companion cookie  ...
                  │              │              │
       ┌──────────┴──┐     ┌─────┴────┐    ┌────┴─────┐
       │  PROJECTOR  │     │ PHONE A  │    │ PHONE B  │
       │             │     │          │    │          │
       │ host =      │     │ player 0 │    │ player 1 │
       │ eGame.Owner │     │          │    │          │
       │ or transfer │     │          │    │          │
       └─────────────┘     └──────────┘    └──────────┘

Per companion-supporting game (server/static/game-src/<game>/):
  • boardgame-render-game-<X>.ts             (existing — solo)
  • boardgame-render-game-<X>-projected.ts   (NEW — extends SurfaceRendererBase)
  • boardgame-render-game-<X>-companion.ts   (NEW — extends SurfaceRendererBase)
  • boardgame-render-game-<X>-layout.ts      (NEW — exports layoutFor function)

Shared (server/static/src/components/):
  • boardgame-surface-renderer-base.ts       (NEW — cross-cutting concerns)
```

**Cardinal rule**: the server is agnostic about display surfaces *for state purposes*. The projector connects as `ObserverPlayerIndex`; each phone as `PlayerIndex(n)`. Existing sanitization machinery handles the privacy story.

What's new is a *coordination layer* (pairing, presence, host actions, sync timestamps) and a per-surface client renderer system — all additive, none altering the existing solo flow or state pipeline.

## 5. Identity & Pairing

### 5.1 Room Codes

- 4 uppercase letters from a confusion-resistant alphabet (omit O/I/L/Z → 22^4 = 234,256 codes).
- Generated at game-create when `CompanionMode = true`. Stored as a typed field `CompanionRoomCode string` on `extendedgame.StorageRecord` (see `server/api/extendedgame/main.go:15`). **No in-memory registry** — the storage row is the single source of truth.
- **Lookup** via a new storage method `GameByRoomCode(code string) (gameID, error)`, paralleling `UserIDsForGame()` at `server/api/storage.go:58-61`.
- **Collision handling**: generation retries up to 10 random draws; if all collide (extremely rare until the namespace is near-full), falls back to a 5-letter code. Beyond that, game-create returns an error and the host is told to retry later.
- **TTL grace period**: codes remain exclusively assigned for 24 hours after `Finished == true` before becoming eligible for recycling. This prevents the "Alice's just-finished phone still holds the code; new host gets the same code; Alice silently joins a stranger's game" race.
- **Rate limiting**: `/api/join` is rate-limited per IP to 10 lookups/minute with exponential backoff on consecutive 404s. This is the V1 mitigation against namespace enumeration.
- **Room lock**: the host can flip a "Lock room" toggle on the projector at any time. While locked, no new phones can claim a seat. Open by default to preserve the Jackbox-style lightweight feel for in-person play; the toggle exists as the escape hatch for streamed/public games. The locked-state lives on `eGame` as `CompanionLocked bool`.

### 5.2 Phone Join Endpoint

All endpoints in this section require HTTPS in production (`Strict-Transport-Security` enforced at the reverse proxy).

`POST /api/join` with body `{ "code": "JKLB" }`:

1. Rate-limit check (per IP, §5.1).
2. Look up game via `GameByRoomCode(code)`. 404 if not found, locked, or `Finished`.
3. Return `{ gameID, gameName, displayName, minPlayers, maxPlayers, currentPlayers, requiresRolePicker, requiresHostAdmit }`. **The response does *not* include role-slot details** — those come from a separate authenticated request (§6.1) so the response of a brute-force scrape cannot reveal asymmetric-role metadata.

The phone client then runs the identity step (Firebase anonymous or Google), shows the avatar/name picker for anon users, optionally shows the role picker (§6), and finally posts `POST /api/join/seat` with body `{ gameID, roleSlot?, avatarSlug, displayName }` carrying the Firebase ID token. The seat endpoint:

1. Validates the token.
2. Validates `displayName` server-side (§5.5).
3. **Writes `displayName` and `avatarSlug` to a new `seatPresentation` row keyed on `(gameID, playerIndex)` — not to the user's `StorageRecord`.** This is the per-seat identity record (§5.4). The user's `StorageRecord` is *not* mutated by joining a game.
4. Runs the existing `SeatPlayer` proposal path.
5. Issues a `companion` session cookie scoped to the gameID with `Secure; HttpOnly; SameSite=Lax`.

### 5.5 Display-Name Validation

Server-side, on `/api/join/seat`:

- NFKC-normalize the input.
- Allow only `[A-Za-z0-9 ]`, length 2–24 inclusive after trimming whitespace.
- Reject zero-width characters, RTL-override, combining diacritics, and any Unicode in the C0/C1 control or surrogate ranges.
- Reject if the normalized name (lowercased) matches another seat's name in the same `gameID`. The phone client retries with a suggested suffix (e.g., "Alice" → "Alice2"). No global uniqueness; per-game only.
- No slur catalog in V1; if abuse surfaces, hook in a list-based filter at this validation step.

### 5.3 Identity

Two paths:

- **Firebase anonymous**: `firebase.auth().signInAnonymously()` on the phone returns a UID. The server, on `/api/join/seat`, validates the ID token using the existing `firebase.VerifyIDToken()` path in `server/api/auth.go:164` (which already works for anon UIDs — `firebase-verify` does not distinguish). A `StorageRecord` for the *user* is created on first contact with empty `Email` and `DisplayName` (the chosen display name goes to `seatPresentation`, not the user record — see §5.4). The user record exists only to anchor existing identity machinery (`UserIDsForGame()`, cookies).
- **Google sign-in**: existing flow. `StorageRecord` is populated with real email/photo/displayName as today. At the avatar/name picker step (§10), the Google user can optionally override their default name/avatar *for this game*; the override lands in `seatPresentation` without mutating the user record.

A Firebase anon UID lasts across refreshes on the same device (token persistence). If the phone closes the tab and reopens within the token TTL, the UID restores and they auto-rejoin their seat (see §9.3 Reconnection).

### 5.4 Seat Binding & seatPresentation

On successful `/api/join/seat`, the server runs the existing `SeatPlayer` proposal path (see `moves/seat_player.go`) — with the anon UID treated as a UID string for `UserIDsForGame()` purposes only. Existing sanitization-per-player and `calcViewingAsPlayerAndEmptySlots()` work unchanged.

The chosen `displayName` and `avatarSlug` live in a new `seatPresentation` table:

```go
// In a new file, e.g. server/api/seatpresentation/main.go
type StorageRecord struct {
    GameID      string
    PlayerIndex boardgame.PlayerIndex
    DisplayName string
    AvatarSlug  string
}
```

Lookup methods on `StorageManager`:

```go
SeatPresentation(gameID string, p PlayerIndex) (*seatpresentation.StorageRecord, error)
SetSeatPresentation(rec *seatpresentation.StorageRecord) error
ClearSeatPresentation(gameID string, p PlayerIndex) error  // called on Free Seat (§9.2)
```

**Per-seat, not per-user**, deliberately:
- A Google user joining game 2 with a different avatar does not mutate game 1's presentation.
- "Free Seat" (§9.2) clears the row so the next joiner doesn't inherit the previous player's name/avatar.
- For Google-signed-in joiners, the row is auto-populated from the user's profile (with the option to override at join time).

The `companion` cookie issued at `/api/join/seat` is scoped to the gameID and carries the anon UID. Reloading `/game/<name>/<id>` resolves the surface from the cookie (§7.1) and the player index from existing auth.

## 6. Seat Assignment: Symmetric vs Asymmetric

The framework decides whether to show a role picker on the phone, based on whether the game's `playerState` has role or team behaviors.

**Detection** (server-side, at join-time): a game is **asymmetric** if its `playerState` satisfies `behaviors.HasPlayerRole` OR `behaviors.HasPlayerTeam`.

- `behaviors.HasPlayerRole` is the existing detection interface in `behaviors/role.go:26`. A `playerState` satisfies it by embedding `behaviors.PlayerRole`.
- `behaviors.HasPlayerTeam` is **new in this spec**: a one-method interface `GetPlayerTeam() *PlayerTeam`. The existing `behaviors.PlayerTeam` struct (defined today at `behaviors/team.go:29`) gains a one-line `GetPlayerTeam() *PlayerTeam { return p }` method so the type assertion works. The `PlayerTeam` behavior itself is otherwise unchanged.

**Override** is via a separate, single-purpose extension interface on the delegate:

```go
type CompanionSeatAssignmentDelegate interface {
    CompanionSeatAssignment() CompanionSeatAssignment
}

type CompanionSeatAssignment int

const (
    CompanionSeatAssignmentDefault CompanionSeatAssignment = iota
    CompanionSeatAssignmentAuto
    CompanionSeatAssignmentRolePicker
)
```

When unimplemented, `CompanionSeatAssignmentDefault` applies: detection by behaviors as above.

### 6.1 Role Picker — Privacy-Preserving Payload

For asymmetric games, the role-picker payload is **not** returned from `/api/join`. It comes from a separate authenticated endpoint `GET /api/join/role-options?gameID=<>` that requires a valid Firebase ID token, ensuring an enumeration attack on `/api/join` cannot reveal the asymmetric structure of a game.

The role-options payload is sanitized through a new `RoleVisibilityPolicy` on `GameDelegate`:

```go
type CompanionRolePrivacyDelegate interface {
    RoleIsPublic() bool  // default: false
}
```

- If `RoleIsPublic() == true` (e.g., Codenames — teams are public knowledge): the payload labels each slot with its real role, and the projector renders the role on the avatar strip.
- If `RoleIsPublic() == false` (the default; covers Werewolf, Secret Hitler, Mysterium): the payload labels each slot only with a neutral identifier ("Seat 1", "Seat 2") and a generic team color *only if* `TeamIsPublic() == true`. The phone shows the real role only after the joiner picks a seat — and the role is revealed only to that phone (delivered via the per-player sanitized state, not via the seat-options payload). **The projector never sees the role.**

This means a hidden-role game's projector strictly mirrors what an in-room observer could see: seat positions + chosen avatars/names + presence. No role leak.

### 6.2 Presence Channel Sanitization

The `Absent []PlayerIndex` list (§9) is also sanitized via the same policy. For private-role games where role *and* identity are hidden (rare; usually identity is public), presence is delivered as a per-player sanitized list rather than a global broadcast. By default (identity public, role private), presence is a global field on public state — same as it would be for an in-room observer noting "Alice stepped out."

### 6.3 Slot Label Resolution

Single-purpose extension interface:

```go
type CompanionSlotLabelDelegate interface {
    CompanionSlotLabel(playerIndex boardgame.PlayerIndex, state boardgame.ImmutableState) string
}
```

When unimplemented, label is derived from the game's `role` enum (or `team` enum if no role) display name. When `RoleIsPublic() == false`, the label is forced to "Seat N" regardless of override.

## 7. Renderer Model

Per companion-mode game, files in `server/static/game-src/<game>/`:

| File | Purpose |
|------|---------|
| `boardgame-render-game-<X>.ts` | Existing solo-device renderer. Untouched. |
| `boardgame-render-game-<X>-layout.ts` | **NEW.** Module exporting a `layoutFor(state, surface): LayoutPlan` *function* (see §7.2). Imported by both surface-specific renderers below. |
| `boardgame-render-game-<X>-projected.ts` | **NEW.** Projected view, extending `BoardgameSurfaceRendererBase` (§7.3). |
| `boardgame-render-game-<X>-companion.ts` | **NEW.** Companion view, extending the same base. |

### 7.1 Surface Selection (Session-Scoped, Not URL-Scoped)

Surface is **decided at join time and stored in the session cookie**, not in a URL query param. The reasons: cookie-bound surface survives refresh, sharing/bookmarking a phone's URL doesn't accidentally put a friend into projector mode, and the server already needs to know surface for presence tracking (§9).

- Creating a companion-mode game issues a `surface=projector` cookie scoped to the gameID to the creating browser.
- `/api/join/seat` issues a `surface=companion` cookie scoped to the gameID to the joining phone.
- URL stays clean: `/game/<name>/<id>`. The loader (`boardgame-render-game.ts:365`) reads the cookie to pick the renderer:

```ts
const surface = readSurfaceCookie(gameID);   // 'projector' | 'companion' | null
const suffix = surface === 'projector' ? '-projected'
             : surface === 'companion' ? '-companion'
             : '';
await import(`../../game-src/${gameName}/boardgame-render-game-${gameName}${suffix}.ts`);
```

For dev/debug, a `?display=projected|companion` query param overrides the cookie (no production effect, gated on `OfflineDevMode`).

If the suffixed file fails to load (deployment error), the loader falls back to the solo renderer with a console warning. The lint check in §14.1.5 catches missing files at build time.

### 7.2 Layout as a Function over State

The layout module exports a single function, not a constant:

```ts
// boardgame-render-game-<X>-layout.ts
export type Anchor =
  | { kind: 'stack';            stackPath: string;  surface: SurfaceID; offset?: {x:number,y:number} }
  | { kind: 'offscreen-stack';  stackPath: string;  surface: SurfaceID; edge: 'top'|'bottom'|'left'|'right'|'top-left'|... }
  | { kind: 'badge';            forPlayer: PlayerIndex; surface: SurfaceID; ... }
  | { kind: 'secret-value';     forPlayer: PlayerIndex; surface: SurfaceID; valuePath: string };

export type LayoutPlan = { anchors: Anchor[] };

export function layoutFor(state: FullGameState, surface: SurfaceID, viewingAsPlayer: PlayerIndex): LayoutPlan { ... }
```

`layoutFor` is called every render cycle and resolves anchors against the current state. It handles:

- **Variable player counts** — the function reads `state.Players.length` and emits per-player anchors dynamically.
- **Dynamic stacks** — stacks that exist only during certain phases (e.g., a "current trick" pile) are emitted only when present in state.
- **Non-card private info** — `kind: 'secret-value'` anchors handle hidden numbers, secret meeples, etc., declaring where their reveal goes and (for projector) what placeholder.

Each renderer takes the `LayoutPlan` and renders only anchors whose `surface` matches its own — *plus* `offscreen-stack` anchors on the *opposite* surface (positioned off-screen at the declared edge, used as the FLIP animation source/destination). Cross-surface card flight then "just works" via the existing FLIP animator: on the projector, the card animates from on-screen `deck` to off-screen `player-hand-0` parked at the south-west edge; on the player-0 phone, it animates from off-screen `deck` parked at the top edge to on-screen `player-hand-0`.

### 7.3 Shared Surface Renderer Base

Cross-cutting concerns — avatar strip, score ribbon, current-player indicator, presence/absent badges, animation timing, FLIP wiring — live in a new `BoardgameSurfaceRendererBase` mixin (a Lit `ReactiveController` or base class):

```ts
// server/static/src/components/boardgame-surface-renderer-base.ts
export class BoardgameSurfaceRendererBase<GS, PS> extends LitElement {
    @property() state: FullGameState<GS, PS>;
    @property() surface: 'projected' | 'companion';
    @property() viewingAsPlayer: PlayerIndex;
    @property() seatPresentations: SeatPresentation[];   // §5.4
    @property() absentPlayers: PlayerIndex[];
    @property() roomLocked: boolean;
    @property() isHost: boolean;

    protected get layoutPlan() { return this.gameLayout(this.state, this.surface, this.viewingAsPlayer); }
    protected gameLayout(...): LayoutPlan { /* abstract — implemented by each game's layoutFor */ }

    protected renderAvatarStrip(): TemplateResult { /* shared */ }
    protected renderAbsentBadges(): TemplateResult { /* shared, gated by §6.2 sanitization */ }
    protected renderHostControls(): TemplateResult { /* shared, gated by isHost */ }
}
```

Each per-game per-surface renderer subclasses this and overrides `gameLayout` (wiring to the game's `layoutFor`) plus the surface-specific composition. The base never gets per-game logic; the subclasses never reimplement cross-cutting pieces.

### 7.4 Existing per-Player Info Components

`boardgame-render-player-info-<X>.ts` files (seen in `server/static/game-src/murdermrmonroe/`) continue to render avatars in solo and projector views. The companion view typically does not need them.

## 8. Animation Synchronization Primitive

A single piggybacked primitive: every WebSocket frame includes a server timestamp. The client maintains a running estimate of one-way latency from those timestamps and schedules animations against a server-anchored play-at instant. No separate ping endpoint, no clock-sync warmup phase, no `/api/server-time` HTTP endpoint.

### 8.1 Server Stamps

The existing `socketMessage` envelope (see `server/api/websockets.go:19-22`) carries server timing in `Data` for state-push messages:

```go
type stateUpdateData struct {
    // ... existing state fields ...
    ServerSentAt int64 `json:"serverSentAt"`  // ms since epoch, set immediately before write
    ServerPlayAt int64 `json:"serverPlayAt"`  // ms since epoch; serverSentAt + ANIMATION_LEAD_MS
}
```

`ANIMATION_LEAD_MS` defaults to 250ms. Games may override per the `CompanionAnimationLeadDelegate` interface (§11).

This piggybacks on the existing message stream — no protocol bump, no new fields on the outer `socketMessage` envelope.

### 8.2 Client Estimation

On every state push received, the client records:

```js
const localRx   = performance.now();              // monotonic
const serverTx  = msg.serverSentAt;
const oneWayMs  = (localRxAsEpochMs) - serverTx;  // minimum-wins estimator
```

The estimator keeps the **minimum** of `oneWayMs` over the last 30 frames (NTP-style: the lowest-latency sample is closest to true one-way delivery, because variance only adds; it doesn't subtract). The server-to-local offset is then `minOneWayMs`, and `localEquivalent(serverPlayAt) = serverPlayAt - minOneWayMs + (perfNowEpochOffset)`.

Animations are scheduled with `setTimeout(playFn, localEquivalent(serverPlayAt) - now)`.

### 8.3 Known Limitations (Documented, Not Solved)

- **Asymmetric routes** (phone on cell, projector on Wi-Fi): the one-way estimator is biased by the asymmetry but the bias is consistent across frames, so animations on each surface are *self-consistent* (each side animates on its own offset) and the visible cross-surface drift is bounded by the asymmetry, not by the variance.
- **JS GC pauses & background-tab throttling**: a `setTimeout` set to fire in 200ms can fire 200ms+ late under Chrome's hidden-tab throttling or on a low-end Android during GC. We accept this — V1 explicitly does NOT promise frame-perfect sync.
- **First state push beats the estimator's window**: on the very first push, the estimator has 1 sample. Animations play immediately on receive (not via `setTimeout`). The first deal animation may be visibly uncoordinated; we accept this for V1.

### 8.4 Fallback

`localEquivalent(serverPlayAt) - now < 0` means the play-at instant is already past — animation plays immediately on receive. Same fallback when fewer than 3 valid samples have been collected.

### 8.5 Acceptable V1 Outcome

Cross-surface animations look "right" on the median LAN connection (~95% of plays). Pathological networks (mobile in low-signal, asymmetric Wi-Fi) show visible drift up to ~200ms. This is documented in the user-facing FAQ for V1; future work may add an explicit clock-sync round-trip if playtests find the median-LAN baseline insufficient.

## 9. Presence & Host Override

### 9.1 Presence Tracking (Heartbeat-Based)

A counter-based "live socket count" is fragile: TCP RSTs on mobile networks don't fire close frames, zombie tabs keep count > 0, and reconnect-faster-than-unregister leaves a permanently-stuck count. We use **heartbeat-based liveness** instead.

Each companion WebSocket sends a heartbeat ping every 10 seconds. The server tracks `lastHeartbeat map[gameID]map[PlayerIndex]time.Time` and considers a player **absent** when `time.Since(lastHeartbeat) > 30s`. Specifically:

- Server enforces WebSocket read deadlines (30s). Idle sockets get a forced close.
- Heartbeat-tracking lives inside the existing `versionNotifier`'s goroutine, alongside `register`/`unregister`/`notifyVersion`. A new `chan heartbeat { gameID, playerIndex, ts }` is processed in the same select loop — **no mutex**, matching the framework's channel-based concurrency idiom (per `server/api/websockets.go:46-54`).
- A periodic ticker (every 5s) inside the notifier scans `lastHeartbeat` for stale entries and emits `presenceChange` events, which flow to a per-game absent-set computation. Stale-entry detection plus heartbeats jointly close the zombie-tab gap.

The state pushed to clients includes an `Absent []PlayerIndex` field, gated by sanitization per §6.2.

The `lastHeartbeat` map evicts game entries when a game reaches `Finished` to prevent unbounded growth.

### 9.2 Host Actions

The host is identified as `eGame.Owner == currentUser` viewing on the projector surface (per the `surface=projector` cookie of §7.1). Their projector view exposes three actions on any absent player:

| Action | Endpoint | Effect |
|--------|----------|--------|
| Skip turn | `POST /api/game/<id>/hostSkipTurn?player=N` | Server proposes the game-defined absent-player fallback move (§11). |
| Replace with Agent | `POST /api/game/<id>/hostReplaceWithAgent?player=N` | Server engages the existing `Agents` flow for player N. |
| Free seat for new joiner | `POST /api/game/<id>/hostFreeSeat?player=N` | See §9.4 — clears the seat AND its private state. |

Routes use the framework's existing camelCase handler naming (per `server/api/main.go:944,1072,1301`). All three are gated on `IsHost(currentUser, gameID)` and recorded to a per-game audit log (new table `companionHostAudit`).

Rate-limit host actions to 1 per second per `(gameID, hostUserID)` to mitigate trolling-host scenarios.

### 9.3 Host Transfer (V1)

A V1 must-have, given that the original host's projector going down would otherwise permanently lock the game.

Mechanism: if `eGame.Owner` has no live projector socket (no heartbeat within 30s), any seated companion player may claim host via `POST /api/game/<id>/claimHost`. The first claim within a 5-second contention window wins; ties broken by lowest player index. On claim:

- Server sets `eGame.CompanionHostOverride = <claimingUserID>`.
- Subsequent `IsHost(user, gameID)` returns true for either the original Owner OR the override.
- If the original owner returns, both have host powers until the override is dropped (audited).

This is intentionally permissive: a malicious host transfer is in scope only for the "trusted friends in person" threat model. Public/streamed games should keep the room locked from the start (§5.1).

### 9.4 Free-Seat Semantics — Avoiding State Leak

The naive "clear `UserIDsForGame()[N]`" leaves the game's `playerState[N]` intact with the prior player's private cards/info. The next joiner who takes seat N would inherit Alice's hand on first state push.

Fix: every companion-supporting game must implement a one-shot game-defined "reset seat" move:

```go
type CompanionAbsentSeatResetDelegate interface {
    CompanionResetSeatMove() boardgame.MoveType  // returns a configured move that wipes per-seat private state
}
```

When `hostFreeSeat` is invoked, the server:

1. Proposes the reset move via the existing move pipeline (sanitization, validation all apply).
2. Clears `UserIDsForGame()[N]`.
3. Calls `ClearSeatPresentation(gameID, N)` (§5.4).

The reset move is game-specific because "what counts as a clean slate" varies: a hidden-role game needs to redraw a role; a card game needs to return the hand to the deck (or discard). Games that don't supply the move cannot use Free-Seat; the host's only options for those games are Skip or Agent.

### 9.5 Reconnection

A phone reconnects in three flavors:

- **Same browser, refresh**: Cookie still valid → server resolves anon UID → existing `calcViewingAsPlayerAndEmptySlots()` returns same seat. Heartbeat resumes; absent flag clears.
- **Same browser, new tab**: Cookie shared → same seat → second tab joins as the same player. Both send heartbeats; presence remains live as long as either does.
- **New browser, same device**: Anon Firebase UID restored from IndexedDB (Firebase's default persistence). Same as refresh.
- **Truly new browser (cleared storage)**: Anon UID is lost. Phone must re-enter the room code. Server's resolution path first checks `if anonUID in UserIDsForGame() → restore that seat` before falling back to next-open-seat assignment, so a same-UID re-join lands on the same seat as long as the seat hasn't been freed.

## 10. Avatar / Name Picker

Modeled on word-bloom's publisher avatar picker (see `/Users/jkomoros/Code/word-bloom/src/components/publisher-avatar-picker.ts`).

### 10.1 Flow

The picker has a 4-step shape: `random → primary → style → review`.

- **random** (front door): On entry, the phone shows one fully-randomized (avatar + name) pair. Big "Looks good — join!" button. Small "Try another" reroll button. Small "Customize" link to go to step 2.
- **primary**: Grid of 12 character/icon primaries to pick from.
- **style**: Decoration + corner + tint pickers (matches word-bloom's composite avatar model).
- **review**: Confirm, then commit.

Most users will tap "Looks good" on step 1 and skip the rest.

### 10.2 Avatar Composition

Composite id format (4 dash-joined slugs, lifted from word-bloom for consistency):

```
${primaryId}-${decorationId|none}-${cornerId|none}-${tintId|none}
```

Catalog lives in `server/static/src/components/companion-avatar-primaries.ts` etc. We do NOT share the word-bloom catalog literally — these are different products — but we mirror the format and a starter catalog (12 primaries, 12 decorations, 4 corners, 8 tints).

### 10.3 Name Generator

Adjective + animal (e.g., "BrightFox", "ShyOtter") drawn from a small curated list (200 + 200 = 40,000 combinations). Generated client-side. The user can edit the name in step 4 ("review"). Maximum length 24 chars; alphanumerics + spaces only.

### 10.4 Persistence

Avatar + name are stored in the **`seatPresentation` table** keyed on `(gameID, playerIndex)` (§5.4), NOT on the user's `StorageRecord`. This means:

- Joining game B with a different avatar does not mutate game A's presentation.
- Freeing a seat (§9.4) clears the row, so the next joiner doesn't inherit the previous player's identity surface.
- A Google-signed-in user can override their default display name and avatar on a per-game basis at the review step.

The user's persistent `StorageRecord` remains the source of truth for *user identity* (Email, Google photo); `seatPresentation` is the source of truth for *what this seat looks like in this game*.

## 11. Game Capability Declaration

The framework idiom (per `base/game_delegate.go:507`, `behaviors/role.go:26`) is **single-purpose extension interfaces detected via type assertion**, not a wide multi-method extension interface. We follow that idiom strictly: opting into companion mode is one marker interface; each optional behavior is its own.

### 11.1 Capability marker

```go
// boardgame/companion/delegate.go (new package, parallel to behaviors/)
package companion

type GameDelegate interface {
    UsesCompanionMode()  // no-arg marker; implementing this IS the opt-in
}
```

A game opts in with:

```go
func (d *MyGameDelegate) UsesCompanionMode() {}
```

Server check: `if _, ok := delegate.(companion.GameDelegate); ok { /* this game supports companion mode */ }`.

### 11.2 Optional extension interfaces

Each is one method; each is detected separately via type assertion; none are required.

```go
// Seat assignment policy override (§6).
type SeatAssignmentDelegate interface {
    CompanionSeatAssignment() SeatAssignment
}
type SeatAssignment int
const (
    SeatAssignmentDefault SeatAssignment = iota
    SeatAssignmentAuto
    SeatAssignmentRolePicker
)

// Whether roles are publicly visible on the projector (§6.1).
type RolePrivacyDelegate interface { RoleIsPublic() bool }

// Whether team affiliation is publicly visible on the projector (§6.1).
type TeamPrivacyDelegate interface { TeamIsPublic() bool }

// Custom slot label (§6.3).
type SlotLabelDelegate interface {
    CompanionSlotLabel(playerIndex boardgame.PlayerIndex, state boardgame.ImmutableState) string
}

// Animation lead-time override (§8).
type AnimationLeadDelegate interface {
    CompanionAnimationLeadMS() int  // milliseconds
}

// Required-if-supporting-FreeSeat: the move that resets a single seat's private state (§9.4).
type AbsentSeatResetDelegate interface {
    CompanionResetSeatMove() boardgame.MoveType
}

// Required-if-supporting-SkipTurn: the move that no-ops or auto-passes for an absent player (§9.2).
type AbsentSkipTurnDelegate interface {
    CompanionSkipTurnMove() boardgame.MoveType
}
```

### 11.3 Defaults

When a delegate doesn't implement a given interface, the framework applies a default:

| Interface | Default behavior |
|-----------|------------------|
| `SeatAssignmentDelegate` | `SeatAssignmentDefault` — detection via behaviors. |
| `RolePrivacyDelegate` | `false` — assume roles are private. The safer default for hidden-role games; games like Codenames must explicitly opt in. |
| `TeamPrivacyDelegate` | `true` — assume teams are public (most common). |
| `SlotLabelDelegate` | Label derived from `role`/`team` enum display names. |
| `AnimationLeadDelegate` | 250ms. |
| `AbsentSeatResetDelegate` | No default. If unimplemented, the host's Free-Seat action is unavailable for absent players in this game. |
| `AbsentSkipTurnDelegate` | No default. If unimplemented, the host's Skip-Turn action is unavailable. |

## 12. UI / Visual Details

### 12.1 Projector

- Dark theme preferred; legible from 8+ feet.
- Room code typography: ≥120pt. Surrounded by URL: small text "Go to boardgame.app/join" above the code.
- QR code: 25% of viewport width, in a corner.
- Seats panel: avatar + name + presence indicator (pulse if it's their turn; faded if absent).
- Animation: cards fly off the projector edge using FLIP transforms toward the seat's screen-position. Each seat has a known anchor coordinate (passed into the layout module).

### 12.2 Companion

- Mobile-first portrait.
- Top: tiny game-state ribbon (current player, phase, score).
- Middle: this player's hand, fanned. Tap-to-select; selected card has a "Play" affordance.
- Bottom: action buttons (e.g., "Pass", "Draw").
- Animation: incoming cards enter from the top edge (matching the "from the shared screen" direction); outgoing cards exit the top edge.

### 12.3 No Cross-Surface Audio (V1)

Per the Jackbox research, audio cues bridge two screens well, but we defer audio to V2. V1 is silent.

## 13. Edge Cases & Failure Modes

| Case | Behavior |
|------|----------|
| Room code typo'd | 404 / "Room not found". `/api/join` is rate-limited per IP (10/min) with exponential backoff on 404. |
| Two phones race for the last open seat | The `/api/join/seat` handler acquires a per-game lock spanning the SeatPlayer proposal. First request commits, others get 409 Conflict with the latest seat-availability snapshot and retry against it. |
| Two phones race for the same role slot | Same handler/lock; loser gets 409 and is re-shown the role picker with the now-filled slot marked. |
| Host disconnects (closes projector) | Game continues. If `eGame.Owner` heartbeat is stale (>30s) any seated player may claim host (§9.3). |
| Original host returns after transfer | Both have host powers; audit log records the overlap. |
| Anon UID token expires (1h TTL) | Silent re-auth via Firebase SDK. If the device is offline at the refresh moment, on next online attempt the same UID restores from IndexedDB and resumes. |
| Phone went to sleep for 2h, then woke up | Same UID restores; reconnect path of §9.5 applies; same seat as long as not freed. |
| Game gets to `Finished` state | Room code enters 24h grace period (§5.1) before recycling. Phones get a "Game over" screen. |
| Same Google user opens two phones for the same game | First takes the seat; second is observer (existing behavior). |
| `serverPlayAt` reports a time before client-now (slow state-fetch) | Animation plays immediately on the receiving client (§8.4). |
| Network partition phone↔server | Heartbeat goes stale; player flagged absent after 30s. Phone sees "Reconnecting…" overlay; rejoins when network restores. |
| Host frees Alice's seat, Alice reconnects 5s later | Alice's anon UID no longer maps to seat N. She re-enters the room code and is auto-assigned to the next open seat. Server emits a friendly toast: "Your seat was reassigned by the host; you've been moved to seat M." |
| Game reset move fails on Free-Seat | The seat is not freed; host sees an error. The game-defined reset move is responsible for guaranteed-legal cleanup. |
| Display name collides with existing seat | `/api/join/seat` returns 409 with a suggested suffix; phone retries silently. |
| Display name has zero-width/RTL chars | `/api/join/seat` returns 400 with a generic "name not allowed"; phone surfaces the reroll button. |

## 14. Testing Strategy

### 14.1 Unit Tests

- `room_code_test.go`: code generation alphabet, uniqueness, collision handling.
- `seat_assignment_test.go`: behavior detection (with PlayerRole, PlayerTeam, both, neither); `CompanionSeatAssignment` overrides.
- `presence_test.go`: live count increment/decrement, absent flag computation.
- `host_actions_test.go`: skip / agent / free for each host action; non-host caller is rejected.

### 14.1.5 Lint / Build-Time Check

- A small Node script (run in CI) that:
  - Walks `server/static/game-src/*/`, identifies games whose Go delegate implements `companion.GameDelegate`.
  - Asserts each one ships `-projected.ts`, `-companion.ts`, and `-layout.ts` siblings.
  - Imports the `-layout.ts` and verifies it exports a `layoutFor` function (not a constant) with the right signature.
  - Asserts each surface renderer subclasses `BoardgameSurfaceRendererBase`.
  - Catches §7.1 deployment errors at build time.

### 14.2 Integration Tests

- Multi-client websocket scenario: one projector + 2 phones connect; one phone disconnects; presence updates push; reconnect restores.
- Anonymous join → seat → state push → move → state push. Compare projector and phone state JSONs to confirm sanitization differences.

### 14.3 Browser / E2E Tests (Playwright)

- Headless run of `murdermrmonroe` (or chosen MVP game) with projector tab + 3 phone tabs.
- Verify room code visible on projector, code entry works from phone, avatars appear on projector after join.
- Verify a deal animation: capture before/after screenshots on both projector and phone, assert card flies on each.
- Verify drop/pause/host-skip/host-agent flow.

### 14.4 Manual / Playtest

- One real session with 4 humans (per the user's "live, in-person" guidance) before V1 ships. The Jackbox-style group-laughter test cannot be automated.

## 15. Scope / MVP Phasing

### Phase 1 — Foundations (PR 1)

- `companion.GameDelegate` marker + capability gating on the create-game form.
- `CompanionRoomCode` + `CompanionLocked` on `extendedgame.StorageRecord` + `GameByRoomCode` storage method.
- `/api/join` + `/api/join/seat` endpoints with HTTPS-only, rate limiting, display-name validation.
- Firebase anonymous auth flow on the phone.
- `seatPresentation` storage table + accessors.
- Surface cookie routing in `boardgame-render-game.ts`.
- `BoardgameSurfaceRendererBase` mixin.
- Three-renderer + `layoutFor` convention documented in `TUTORIAL.md`.

### Phase 2 — Seat assignment + identity (PR 2)

- `behaviors.HasPlayerTeam` interface + `GetPlayerTeam()` method on `PlayerTeam`.
- Symmetric vs asymmetric detection (behaviors-based) with privacy-respecting role-options endpoint.
- Role picker UI on phone.
- Avatar/name picker (random front door + customize) — depends on the avatar primaries catalog being checked in.
- Optional extension delegates: `SeatAssignmentDelegate`, `RolePrivacyDelegate`, `TeamPrivacyDelegate`, `SlotLabelDelegate`.

### Phase 3 — Presence + host override (PR 3)

- Heartbeat-based presence tracker folded into the existing notifier goroutine.
- Sanitization-aware `Absent []PlayerIndex` channel.
- Host actions: Skip / Agent / Free-Seat (each gated on the relevant `AbsentX` delegate being implemented).
- Host transfer flow with audit log.
- "Waiting for Alice…" projector affordance.

### Phase 4 — Animations (PR 4)

- `layoutFor` function convention for the MVP game (likely `murdermrmonroe`).
- `*-projected.ts` + `*-companion.ts` for MVP game.
- `ServerSentAt`/`ServerPlayAt` piggybacked on state pushes.
- One end-to-end deal animation working across surfaces.

### Phase 5 — Polish + V1 ship (PR 5)

- QR code on projector.
- Visual polish, dark theme.
- Room-lock toggle UI.
- Playtest with 4+ humans.

### Explicitly out of V1

- Audio cues.
- Multi-projector.
- Spectator-on-phone (audience tier).
- Anonymous → Google account upgrade UI.
- Cross-session persistent avatar.
- Improved clock sync beyond the minimum-wins one-way estimator (only if playtests show the V1 sync is insufficient).

## 16. Open Questions

Most prior open questions have been resolved inline by this revision. Remaining items:

1. **MVP game**: assumed `murdermrmonroe` based on the untracked `server/static/game-src/murdermrmonroe/` directory. Confirm before Phase 4. Alternatives include a fresh implementation of Love Letter (`research/love-letter/`) or a new lightweight hidden-role game.
2. **Slur/abuse word filter for display names**: V1 ships with NFKC + ASCII validation only (§5.5). Hook point exists; deciding whether to add a blocklist before V1 or after first incident.
3. **Avatar primaries catalog**: this spec proposes mirroring word-bloom's `(primary, decoration, corner, tint)` 4-tuple model. The actual catalog content (which 12 primaries? which 12 decorations?) is a separate art-direction task tracked outside this spec. Phase 2 cannot land until the starter catalog is committed.
4. **Confusion-resistant alphabet**: proposed omitting O/I/L/Z (letters only). Confirm this is the right set, or revise (some teams prefer numeric-only codes for accessibility).

### Decisions encoded into this spec (closed)

- **Room code persistence**: typed field on `extendedgame.StorageRecord`. (Was Open Q2.)
- **Room code recycling**: 24h grace period after `Finished` before reuse. (Was Open Q3.)
- **Per-game animation timing override**: `CompanionAnimationLeadDelegate` interface on the delegate. (Was Open Q4.)
- **Host transfer**: in V1, not deferred. (Was Open Q6.)
- **Skip-turn fallback move**: `CompanionAbsentSkipTurnDelegate` interface on the delegate; Skip action unavailable if not implemented. (Was Open Q7.)

## 17. Glossary

- **Projector**: the shared screen (laptop, TV, tablet) connected as `ObserverPlayerIndex` with the projected renderer.
- **Companion**: a phone connected as `PlayerIndex(n)` with the companion renderer, bound to a seat.
- **Host**: by default `eGame.Owner` viewing on the projector surface; transferable via §9.3 when stale.
- **Solo mode**: existing single-device-per-player flow. Untouched by this work.
- **Room code**: 4-letter code stored on `extendedgame.StorageRecord` for companion-mode games. Resolves to a gameID via `GameByRoomCode`.
- **Off-screen anchor**: a layout anchor (typically a stack) positioned outside the visible viewport on a given surface, used as the animation source/destination for cards crossing surfaces.
- **`seatPresentation`**: per-`(gameID, playerIndex)` storage of `DisplayName` + `AvatarSlug`. Source of truth for what a seat looks like in this game; cleared on Free-Seat.
- **`serverPlayAt`**: server-clock timestamp stamped on a state push, indicating the wall-clock instant at which all clients should begin the resulting animation.
- **`layoutFor`**: per-game function `(state, surface, viewingAsPlayer) → LayoutPlan` that resolves anchors against current state. Replaces the static-data layout module from earlier drafts.
- **`BoardgameSurfaceRendererBase`**: shared Lit base class owning cross-cutting renderer concerns (avatar strip, presence badges, animation wiring, host controls).
