# Per-Person Mobile UI (Projected + Companion Mode)

Spec date: 2026-05-23
Ticket: [#759 — Allow remote control / projected mode](https://github.com/jkomoros/boardgame/issues/759)
Branch: `per-person-mobile-ui`

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
5. Host's browser navigates to `/game/<name>/<id>?display=projected`.

### 3.2 The Projector View Pre-Game

- Fullscreen-friendly layout. Big, legible room code (e.g. "Go to **boardgame.app/join** and enter **JKLB**").
- A QR code linking directly to `/join?code=JKLB`, generated client-side.
- A "Seats" panel showing N empty avatar slots (N comes from `MinNumPlayers` to `MaxNumPlayers`). As phones join, slots fill with their chosen avatar + name.
- A "Start Game" button enabled once `MinNumPlayers` slots are filled. Disabled before then.
- For asymmetric-role games (see §6), each filled slot shows the player's chosen role.

### 3.3 Phone Joining Flow

1. Phone visits `/join` (or scans the QR), enters the 4-letter code (case-insensitive). Code is validated; failure shows "Room not found".
2. Server looks up the game by code → returns game metadata to the client.
3. Phone shows an identity screen: **"Continue as guest"** (primary button) or **"Sign in with Google"** (secondary). Note: code is entered *before* identity is chosen, per design decision.
4. Guest path: Firebase anonymous sign-in fires under the hood; the client immediately shows the **avatar/name picker** (see §10).
5. Google path: existing Google sign-in flow; skip avatar picker (use display name + photo URL).
6. Once identity is resolved, the client requests a seat. Two cases (see §6):
   - Symmetric game: server auto-assigns the next open seat and returns the assignment.
   - Asymmetric game: server returns the list of available role slots; phone shows a **role picker** ("Spymaster (red team)", "Operative (blue team)", etc.); selection commits the seat.
7. Phone navigates to `/game/<name>/<id>?display=companion` with its session cookie.

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
│            WebSocket version notifier, Firebase auth                │
│                                                                     │
│  NEW:                                                               │
│   • Room-code registry  (code → gameID)                             │
│   • Anon UID → seat binding  (per-game cookie/session)              │
│   • Presence tracker  (player → live companion socket?)             │
│   • Animation envelope  (serverPlayAt timestamp on state pushes)    │
│   • /api/server-time endpoint  (clock-offset estimation)            │
│   • Host-action endpoints  (skip / agent-takeover / free-seat)      │
└─────────────────────────────────────────────────────────────────────┘
                  │              │              │
   ObserverPlayerIndex     PlayerIndex(0)   PlayerIndex(1)
   + ?display=projected   + ?display=companion  ...
                  │              │              │
       ┌──────────┴──┐     ┌─────┴────┐    ┌────┴─────┐
       │  PROJECTOR  │     │ PHONE A  │    │ PHONE B  │
       │  (laptop/   │     │          │    │          │
       │   TV)       │     │ player 0 │    │ player 1 │
       └─────────────┘     └──────────┘    └──────────┘

Per game:
  • boardgame-render-game-<X>.ts              (existing — solo-device)
  • boardgame-render-game-<X>-projected.ts    (NEW)
  • boardgame-render-game-<X>-companion.ts    (NEW)
  • boardgame-render-game-<X>-layout.ts       (NEW — shared pile positions)
```

The cardinal rule: **the server is largely agnostic about display surfaces.** A connection is identified by `PlayerIndex` (Observer, or 0..N) and serves the existing sanitized state. The `?display=` query param is consumed entirely by the client to pick a renderer.

The server-side novelty is in the *coordination layer*: pairing, presence, host actions, animation timestamps — none of which require changes to the existing state pipeline.

## 5. Identity & Pairing

### 5.1 Room Codes

- 4 uppercase letters, drawn from a confusion-resistant alphabet (omit O/I/L/Z to reduce typo risk → 22 letters × 4 positions = 234,256 codes).
- Generated at game-create time when `CompanionMode = true`. Stored on the `eGame` row alongside the game ID.
- Indexed by code in a server-side registry (in-memory `map[string]gameID` plus storage-backed fallback for restart durability — see §13 Open Questions).
- Expire when the game is `Finished == true`. While a game is open, the code is stable.

### 5.2 Phone Join Endpoint

`POST /api/join` with body `{ "code": "JKLB" }`:

1. Look up game by code. 404 if not found.
2. Check game is not `Finished` and `CompanionMode == true`.
3. Return `{ gameID, gameName, displayName, minPlayers, maxPlayers, currentPlayers, requiresRolePicker }`.

The phone client then drives the identity step locally (Firebase anonymous or Google), runs the avatar/name picker (anon path only), and finally posts `POST /api/join/seat` with body `{ gameID, roleSlot?, avatarSlug?, displayName? }` (carrying the Firebase ID token in the auth header). The seat endpoint validates the token, **creates or updates the user's `StorageRecord` with the chosen `DisplayName` and `AvatarSlug`** (for anon users; Google-signed-in users have these populated already), and runs the existing `SeatPlayer` proposal path.

### 5.3 Identity

Two paths:

- **Firebase anonymous**: `firebase.auth().signInAnonymously()` on the phone. Returns a UID. The server, on `/api/join/seat`, validates the ID token using the existing `firebase.VerifyIDToken()` path in `server/api/auth.go:164` (which already works for anon UIDs — `firebase-verify` does not distinguish). A `StorageRecord` for users is created on first contact (Email is empty, DisplayName is the user's chosen avatar name).
- **Google sign-in**: existing flow. Same `StorageRecord` shape, populated with real email/photo/displayName.

A Firebase anon UID lasts across refreshes on the same device (token persistence). If the phone closes the tab and reopens within the token TTL, the UID restores and they auto-rejoin their seat (see §9.3 Reconnection).

### 5.4 Seat Binding

On successful `/api/join/seat`, the server runs the existing `SeatPlayer` proposal path (see `moves/seat_player.go`) — but with the new anon UID treated identically to a Google UID. The user-to-seat mapping in `UserIDsForGame()` (per `server/api/storage.go:58-61`) now contains the anon UID string. From there, all existing flows — sanitization per player, viewing-as-player resolution in `calcViewingAsPlayerAndEmptySlots()` — work unchanged.

A cookie on the phone carries the session, so re-loading `/game/<name>/<id>?display=companion` resolves the player index via existing auth.

## 6. Seat Assignment: Symmetric vs Asymmetric

The framework decides whether to show a role picker on the phone, based on whether the game's `playerState` has role or team behaviors.

**Detection** (server-side, at join-time):
- A game is **asymmetric** if any of:
  - Its `playerState` type-asserts to `behaviors.HasPlayerRole` (exists today — see `behaviors/role.go`).
  - Its `playerState` type-asserts to `behaviors.HasPlayerTeam` (does **not** exist today; this spec adds it mirroring `HasPlayerRole`: a one-method interface `GetPlayerTeam() *PlayerTeam`, added to `behaviors/team.go`).
- Default rule: asymmetric → show role picker; symmetric → auto-assign next open seat.

**Override**: `GameDelegate.CompanionSeatAssignment()` may return one of:
- `SeatAssignmentAuto` — force auto.
- `SeatAssignmentRolePicker` — force the picker.
- `SeatAssignmentDefault` — apply the detection rule above (zero value).

**Role picker payload** (returned from `/api/join` for asymmetric games):

```json
{
  "requiresRolePicker": true,
  "slots": [
    {"playerIndex": 0, "label": "Spymaster", "team": "red", "filled": true,  "avatar": {...}},
    {"playerIndex": 1, "label": "Operative", "team": "red", "filled": false},
    {"playerIndex": 2, "label": "Spymaster", "team": "blue","filled": false},
    {"playerIndex": 3, "label": "Operative", "team": "blue","filled": true,  "avatar": {...}}
  ]
}
```

Slot label resolution: framework reads the game's `role` enum and `team` enum display names. Each game's `GameDelegate` may override the displayed label per slot via a new optional method `CompanionSlotLabel(playerIndex) string`.

## 7. Renderer Model

Per companion-mode game, four files live in `server/static/game-src/<game>/`:

| File | Purpose |
|------|---------|
| `boardgame-render-game-<X>.ts` | Existing solo-device renderer. Untouched. |
| `boardgame-render-game-<X>-layout.ts` | **NEW.** Pure data module exporting shared layout: pile IDs, on-screen vs off-screen positions, per-player edge direction. Imported by both renderers below. |
| `boardgame-render-game-<X>-projected.ts` | **NEW.** Projected view: public board, avatar strip, off-screen piles parked at edges adjacent to each seat. |
| `boardgame-render-game-<X>-companion.ts` | **NEW.** Companion view: this player's hand + actions, off-screen piles for the "shared deck" parked at the inbound edge. |

### 7.1 Surface Selection

`boardgame-render-game.ts` (the loader at `server/static/src/components/boardgame-render-game.ts:365`) is extended to honor the `?display=` query param:

```ts
// pseudocode
const suffix = display === 'projected' ? '-projected'
             : display === 'companion' ? '-companion'
             : '';
await import(`../../game-src/${gameName}/boardgame-render-game-${gameName}${suffix}.ts`);
```

If `?display=` is absent or the suffixed file fails to load, the loader falls back to the solo renderer (`-X.ts`) with a console warning. A companion-mode-supporting game *must* ship the suffixed files; missing them in production is a deployment error caught by the lint script in §14.

### 7.2 Off-Screen Pile Layout

The `*-layout.ts` module is a pure TypeScript export, e.g.:

```ts
// boardgame-render-game-murdermrmonroe-layout.ts
export const PILES = {
  'deck': { onScreen: { surface: 'projected', x: 50, y: 50 } },
  'discard': { onScreen: { surface: 'projected', x: 70, y: 50 } },
  'player-hand-0': {
    onScreen: { surface: 'companion', playerIndex: 0, x: 50, y: 80 },
    offScreen: { surface: 'projected', edge: 'south-west', distance: 200 }
  },
  'player-hand-1': {
    onScreen: { surface: 'companion', playerIndex: 1, x: 50, y: 80 },
    offScreen: { surface: 'projected', edge: 'south', distance: 200 }
  },
  // ...one per player slot
};
```

Each renderer iterates `PILES` and renders only the entries whose `onScreen.surface` matches its surface — *plus* the entries whose `offScreen.surface` matches (positioned off-screen as a hidden stack). The FLIP animator sees all stacks normally; cards move from on-screen "deck" to off-screen "player-hand-0" on the projector (animates south-west toward seat 0), and from off-screen "deck" to on-screen "player-hand-0" on player 0's phone (animates in from the inbound edge).

The choice to keep this as a pure-data module (not a class hierarchy) is deliberate: it stays a single source of truth that both renderers consume without coupling.

### 7.3 Existing per-Player Info Components

The `boardgame-render-player-info-<X>.ts` files seen in `server/static/game-src/murdermrmonroe/` etc. are NOT replaced. They continue to serve the avatar-strip presentation in both projected and solo views.

## 8. Animation Synchronization Primitive

A two-piece system: clock-offset estimation, plus per-state-update `serverPlayAt` stamps.

### 8.1 Clock Offset Estimation

New endpoint `GET /api/server-time` returns `{ "serverNowMs": <int64 ms since epoch> }`.

Client logic (on connect, then once per 30s while idle):

1. Record `t0 = performance.now()` + wall clock.
2. Fetch endpoint.
3. Record `t1 = performance.now()`.
4. `rtt = t1 - t0`. `serverNowAtT1 = response.serverNowMs + rtt / 2` (one-way ≈ RTT/2).
5. `offset = serverNowAtT1 - (wallClockAt(t1))`.
6. Maintain a rolling buffer of the last 5 offsets; use the **median** as the current offset. Discard samples whose RTT exceeds 3× the median RTT (likely a transient stall).

`serverNow()` on the client = `Date.now() + medianOffset`.

### 8.2 serverPlayAt on State Pushes

The existing WebSocket `socketMessage` (see `server/api/websockets.go:19-22`) gets a new field:

```go
type socketMessage struct {
    Type string
    Data interface{}
    ServerPlayAt int64 `json:"serverPlayAt,omitempty"`  // ms since epoch
}
```

When the server pushes a `version` notification for a state change, it sets `ServerPlayAt = serverNowMs + ANIMATION_LEAD_MS` (default 250ms). This is the "play this animation starting at" wall-clock instant.

Client: when it receives a state push, it fetches the new state JSON, then schedules the animation to begin at the local equivalent of `serverPlayAt`. If `serverPlayAt` is already in the past on this client (e.g. extreme latency), animation plays immediately — degraded but not broken.

`ANIMATION_LEAD_MS` is configurable; 250ms is enough for typical LAN/Wi-Fi to converge while not feeling laggy.

### 8.3 Fallback

If the client has fewer than 3 valid offset samples (insufficient data for the median), animations play immediately on state receive. This guarantees we never *block* gameplay on sync.

## 9. Presence & Host Override

### 9.1 Presence Tracking

The existing `versionNotifier.sockets` map (per `server/api/websockets.go:46-54`) is keyed by `gameID`. We add a parallel map:

```go
type playerPresence struct {
    mu sync.RWMutex
    livePerPlayer map[gameID]map[PlayerIndex]int // count of live companion sockets
}
```

A companion socket — identified by `?display=companion` on the connection or an explicit query param at handshake — increments the count on register, decrements on unregister.

**Absent**: `livePerPlayer[gameID][playerIndex] == 0` for a paired (seated) player.

The state pushed to the projector includes an `Absent []PlayerIndex` field (computed at fetch time). The projector renderer uses this to show "Waiting for Alice…" badges.

### 9.2 Host Actions

The host is identified as the game creator (`eGame.Owner == currentUser`). Their projector view exposes three actions on any absent player:

| Action | Endpoint | Effect |
|--------|----------|--------|
| Skip turn | `POST /api/game/<>/host-skip-turn?player=N` | Server proposes the game-specific "no-op for this player" or "auto-pass" move. Game-dependent fallback. |
| Replace with Agent | `POST /api/game/<>/host-replace-with-agent?player=N` | Server runs the existing `Agents` flow for player N — same as games created with agent slots. |
| Free seat for new joiner | `POST /api/game/<>/host-free-seat?player=N` | Clears `UserIDsForGame()[N]`. The next phone joining with the room code can take the slot. |

These three endpoints are gated on `IsHost(currentUser, gameID)`. The actions surface in the projector UI as inline menus on each absent-player badge.

### 9.3 Reconnection

A phone reconnects in three flavors:

- **Same browser, refresh**: Cookie still valid → server resolves anon UID → existing `calcViewingAsPlayerAndEmptySlots()` returns same seat. Phone resumes.
- **Same browser, new tab**: Cookie shared → same seat → second tab joins as the same player. Both receive state pushes (this matches existing multi-tab semantics).
- **New browser**: Anon UID is lost. Player must re-enter the room code. They'll auto-assign to the *next* open seat — which is their old seat *only* if the host has freed it, or if no one else has joined in the meantime. This is documented behavior — anon identity is browser-bound.

When a phone reconnects and `livePerPlayer[gameID][playerIndex]` transitions from 0 → 1, the server clears the absent flag and broadcasts. The projector's "Waiting…" badge dismisses.

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

Avatar + name stored on the user's `StorageRecord` (`DisplayName` field). The avatar composite slug goes into a new `AvatarSlug` field on the user record. This is sent back to the projector to render the avatar strip.

## 11. Game Capability Declaration

New optional method on `GameDelegate`:

```go
type CompanionGameDelegate interface {
    GameDelegate
    SupportsCompanionMode() bool
    CompanionSeatAssignment() CompanionSeatAssignment // optional; defaults to SeatAssignmentDefault
    CompanionSlotLabel(playerIndex PlayerIndex, state ImmutableState) string // optional
}
```

We use an interface-extension pattern rather than adding methods to the base `GameDelegate` to keep existing games compiling unchanged. The server checks via type assertion: `if d, ok := delegate.(CompanionGameDelegate); ok && d.SupportsCompanionMode() { ... }`.

The new `CompanionSeatAssignment` enum:

```go
type CompanionSeatAssignment int

const (
    SeatAssignmentDefault CompanionSeatAssignment = iota  // detect via role/team behaviors
    SeatAssignmentAuto                                    // force auto next-open-seat
    SeatAssignmentRolePicker                              // force role picker
)
```

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
| Room code typo'd | 404 / "Room not found". No retry limiting in V1; party games are low risk for brute-forcing 234k codes in a session. |
| Two phones try to take the same role slot | Race resolved at server. First request wins; loser sees "That seat was just taken" and is re-shown the role picker. |
| Host disconnects (closes projector) | Game continues — no host-only actions are *required* to play. If host returns, they resume host controls. If they never return, the game stalls only if a player goes absent. |
| Anon UID expires | Token TTL is 1 hour by default (Firebase setting). On expiry, a silent re-auth runs. If silent re-auth fails, phone shows "Reconnect" and re-enters the room. |
| Game gets to `Finished` state | Room code retires. Phones get a "Game over" screen. |
| Same Google user opens two phones for the same game | First phone takes the seat (existing `calcViewingAsPlayerAndEmptySlots()` behavior). Second phone is reflected as "this user is already seated" — degraded view; can opt into observer mode (existing). |
| `serverPlayAt` reports a time before client-now | Animation plays immediately on the receiving client. Out-of-sync but never frozen. |
| Network partition between phone and server | Phone enters "Reconnecting…" overlay (existing reconnect logic); presence count drops on the server side after the websocket close. |

## 14. Testing Strategy

### 14.1 Unit Tests

- `room_code_test.go`: code generation alphabet, uniqueness, collision handling.
- `seat_assignment_test.go`: behavior detection (with PlayerRole, PlayerTeam, both, neither); `CompanionSeatAssignment` overrides.
- `presence_test.go`: live count increment/decrement, absent flag computation.
- `host_actions_test.go`: skip / agent / free for each host action; non-host caller is rejected.

### 14.1.5 Lint / Build-Time Check

- A small Node script (run in CI) that walks `server/static/game-src/*/`, identifies games whose Go delegate returns `SupportsCompanionMode() == true`, and asserts that each one ships `-projected.ts`, `-companion.ts`, and `-layout.ts` siblings. Catches the §7.1 deployment-error case.

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

- Room-code registry + `/api/join` + `/api/join/seat` endpoints.
- Anonymous Firebase auth flow on the phone.
- `CompanionGameDelegate` interface + capability gating on the create-game form.
- Display-mode routing on `boardgame-render-game.ts`.
- Three-renderer convention documented in `TUTORIAL.md`.

### Phase 2 — Seat assignment + avatar picker (PR 2)

- Symmetric vs asymmetric detection (behaviors-based).
- Role picker UI on the phone.
- Avatar/name picker (random front door + customize) — depends on the avatar primaries catalog being checked in.

### Phase 3 — Presence + host override (PR 3)

- Presence tracker.
- "Waiting for Alice…" projector affordance.
- Skip / Agent / Free-seat host actions.

### Phase 4 — Animations (PR 4)

- Off-screen pile layout module convention.
- `*-layout.ts` + `*-projected.ts` + `*-companion.ts` for the MVP game (likely `murdermrmonroe`).
- `/api/server-time` + `serverPlayAt` plumbing.
- One end-to-end deal animation working across surfaces.

### Phase 5 — Polish + V1 ship (PR 5)

- QR code on projector.
- Visual polish, dark theme.
- Playtest with 4+ humans.

### Explicitly out of V1

- Audio cues.
- Multi-projector.
- Spectator-on-phone.
- Anonymous → Google account upgrade UI.
- Cross-session persistent avatar.

## 16. Open Questions

1. **MVP game**: assumed `murdermrmonroe` based on the untracked `server/static/game-src/murdermrmonroe/` directory and its hidden-info nature. Confirm.
2. **Room code persistence across server restart**: in-memory map is simplest, but a restart loses the code → gameID mapping. Should we store the code on the `eGame` row? (Recommendation: yes — adds one column, costs nothing.)
3. **Room code reuse after game ends**: codes are short. If we never reuse them, we'll exhaust the space eventually. Recommendation: codes recycle once a game reaches `Finished`.
4. **Per-game animation timing override**: `ANIMATION_LEAD_MS = 250` is a global default. Should games be able to override? (Recommendation: yes, via the existing `animationLength()` renderer hook.)
5. **Confusion-resistant alphabet**: proposed omitting O/I/L/Z. Should we also omit numbers entirely, or include digits? (Recommendation: letters only — easier to read out loud.)
6. **Host transfer**: if the original host loses their device, can the host role transfer to another player? (Recommendation: V2.)
7. **What "Skip turn" means for games where pass isn't legal**: needs a game-defined fallback move per delegate. Could surface a `CompanionAbsentPlayerMove() MoveType` method. (Recommendation: include in Phase 3.)

## 17. Glossary

- **Projector**: the shared screen (laptop, TV, tablet) connected as `ObserverPlayerIndex` with the projected renderer.
- **Companion**: a phone connected as `PlayerIndex(n)` with the companion renderer, bound to a seat.
- **Host**: the user who created the game; identified by `eGame.Owner`. Their projector view has host-action controls.
- **Solo mode**: existing single-device-per-player flow. Untouched by this work.
- **Room code**: 4-letter code generated at game-create for companion-mode games. Resolves to a gameID.
- **Off-screen pile**: a stack positioned outside the visible viewport on a given surface, used as the animation source/destination for cards crossing surfaces.
- **`serverPlayAt`**: server-clock timestamp stamped on a state push, indicating the wall-clock instant at which all clients should begin the resulting animation.
