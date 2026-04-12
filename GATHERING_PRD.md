# Gathering System PRD: Scenario Catalog & Design Fitness

This document catalogs 26 real-world game setup scenarios and evaluates them against the proposed gathering system design. The purpose is to stress-test the design and verify coverage.

## Proposed Design Summary

**Go side:** 3 new move types (`SelectTeam`, `SelectRole`, `SelectColor`), 1 helper (`GatheringMoves(auto)`), 1 delegate method (`ReadyToStart(state) error`). All built on existing infrastructure (phases, `AddForPhase`, `AnyPlayerIndex`, behaviors).

**Client side:** A `boardgame-gathering-panel` with auto-detecting sub-components. Each sub-component renders when its corresponding move is legal, hides when it isn't. Full override via `boardgame-render-gathering-GAMENAME`.

**Core principle:** Move legality IS the detection mechanism. No new flags, modes, or concepts.

---

## Dimension Taxonomy

Every game setup can be described along these dimensions:

| Dimension | Values |
|-----------|--------|
| **Player count** | Fixed / Range / Flexible-with-teams |
| **Colors** | None / Auto / Player-picked / Admin-assigned / Functional |
| **Teams** | None / Fixed-2 / Flexible-N / Asymmetric (1-vs-many) / Secret |
| **Roles** | None / Symmetric / Public-chosen / Public-assigned / Secret-random / Secret-admin / Draft-pick / Unique-constraint |
| **Rounds** | Single / Fixed-N / Until-condition / Open-ended |
| **Between-round changes** | None / New-players / Team-reshuffle / Config-change / Role-rotation |
| **Start trigger** | Auto (enough players) / Admin-click / All-ready / Timer |
| **Authority** | No-admin / Creator-is-admin / Transferable-admin / Moderator-role |

---

## Scenario Catalog

### 1. Chess / Tic-Tac-Toe
**Dimensions:** Fixed-2, auto-color (functional), no teams, no roles, single round, auto-start.

**Setup:** Create → 2 players join → auto-start.

**Design fit:** `DefaultRoundSetup(auto)` with no `WithManualStart`. Zero gathering UI needed beyond "Waiting for 1 more player." **Fully covered.**

---

### 2. Blackjack
**Dimensions:** Range 2-7, no color, no teams, no roles, fixed-N rounds (variant), manual or auto start.

**Setup:** Create with variant (max rounds) → players join → admin starts or auto-starts when min reached → deal.

**Between rounds:** Reset hands, re-deal. No player changes. Phase cycles: InitialDeal → NormalPlay → Cleanup → InitialDeal.

**Design fit:** Already works today with `DefaultRoundSetup(auto, WithManualStart())`. Gathering panel shows "Waiting for Players" + "Start Game" button. **Fully covered.**

---

### 3. Memory
**Dimensions:** Range 2-6, no color, no teams, no roles, single round, variant affects components (card count + card set).

**Setup:** Variant chosen at creation → `BeginSetUp` sizes stacks → components distributed → play.

**Design fit:** Variant is fixed at creation (no mutable variants). **Fully covered.** If we ever want lobby-editable variants, this is the hard case (stack resizing). Deferred.

---

### 4. Codenames (Teams + Roles Within Teams)
**Dimensions:** Range 4-10, no color, fixed-2 teams, public roles (spymaster/guesser) player-chosen, single round (or multi), admin start.

**Setup:** Players join → pick team (Red/Blue) → pick role (Spymaster/Guesser) → admin validates (1 spymaster per team, ≥2 per team) → start.

**Between rounds:** Teams/roles often reshuffle. Spymaster rotates.

**How it maps:**
- `behaviors.PlayerTeam` + `behaviors.PlayerRole` on playerState
- `moves.SelectTeam` + `moves.SelectRole` in `AddForPhase(phaseGathering, ...)`
- `ReadyToStart` validates: each team has exactly 1 spymaster, each team ≥ 2 players
- `WithManualStart()` for admin start
- Between rounds: phase cycles back to gathering, roles reset, players re-select

**Design fit: Fully covered.** The `ReadyToStart` hook handles the complex constraint. The framework's team/role pickers handle selection. Game-specific validation is in one delegate method.

---

### 5. The Resistance / Avalon (Secret Roles, Random)
**Dimensions:** Range 5-10 (but balance-critical), no color, secret asymmetric teams, secret roles randomly assigned, single game, auto-start when full.

**Setup:** All seats must fill → engine auto-assigns secret roles based on player count → reveal to each player (sanitized).

**How it maps:**
- No `SelectRole` move — roles are assigned by a FixUp move, not player-chosen
- A custom `AssignSecretRoles` FixUp move runs after `WaitForEnoughPlayers`
- Role is on playerState with `sanitize:"self"` so only you see your own role
- No `ReadyToStart` needed — roles are auto-assigned, not player-chosen

**Design fit: Covered, but via game-specific FixUp move.** The framework's `SelectRole` isn't used here. The game writes `AssignSecretRoles` (a ~20 line FixUp). The gathering panel just shows "Waiting for Players" + auto-start. **The framework doesn't need to handle secret role assignment — it's inherently game-specific.**

---

### 6. Werewolf / Mafia (Secret Roles, Admin-Moderated)
**Dimensions:** Range 6-15+, no color, secret teams, secret roles admin-assigned OR random, elimination, moderator role.

**Setup:** Moderator creates → players join → moderator assigns roles (or random) → moderator starts.

**How it maps:**
- Moderator is player 0 with a special `IsModeratorRole` in their player state
- If admin-assigned: a custom `AdminAssignRole` move (admin proposes, targeting a specific player)
- If random: a FixUp move like Resistance
- `ReadyToStart` validates all players have roles assigned
- `WithManualStart()` — moderator controls start

**Design fit: Covered.** Admin assignment is a game-specific move (not `SelectRole` since it's admin → player, not self-selection). Random assignment is a FixUp. The gathering panel shows "Start Game" for the moderator. **The framework doesn't need a generic "admin assigns role to player" move — it's too game-specific.**

---

### 7. Risk (Functional Colors, Long Game, Drop-Out)
**Dimensions:** Range 2-6, functional color (army identity), no teams, no roles, single long game, color is player-chosen.

**Setup:** Players join → pick colors → territories distributed → initial army placement → play.

**How it maps:**
- `behaviors.PlayerColor` + `moves.SelectColor` with uniqueness enforcement
- Color is functional (ties to component identity) — the `behaviors.PlayerColor.OwnsToken()` method already handles this
- Territory distribution is game-specific (a dealing phase after gathering)

**Drop-out:** Mid-game, a player leaves. Their armies remain on the board but the player is marked inactive. Game continues. This is handled by `behaviors.InactivePlayer` — `PlayerIndex.Next()` skips them.

**Design fit: Fully covered.** `SelectColor` with uniqueness, `PlayerColor.OwnsToken()` for functional color. Drop-out is existing infrastructure.

---

### 8. Poker Cash Game (Open-Ended, Drop-In/Drop-Out)
**Dimensions:** Range 2-10, no color, no teams, no roles, open-ended rounds, drop-in/drop-out between hands.

**Setup:** Table created → players join → auto-start when min players seated.

**Between rounds:** New players can sit in empty seats. Players can leave (seat reopens). Phase cycles: Gathering → Deal → Bet → Showdown → Gathering.

**How it maps:**
- Phase cycles back to gathering between every hand
- `ActivateInactivePlayer` runs at round start (new joiners become active)
- `ActivateEmptySeat` reopens closed-but-unfilled seats
- `WaitForEnoughPlayers` gates each hand
- Server must re-open `eGame.Open` when phase returns to gathering (currently the server auto-closes and never reopens — **gap**)

**Design fit: Mostly covered.** The one gap is server-side: the server auto-closes a game when all seats are filled/closed and never reopens it. When the game returns to gathering phase and seats reopen, the server needs to set `eGame.Open = true` again. **This is a server-layer fix, not a design gap.** The fix: when `ActivateEmptySeat` fires, the server detects the game has open seats again and sets `eGame.Open = true`.

---

### 9. Spirit Island (Draft-Pick Unique Roles, Cooperative)
**Dimensions:** Range 1-4, no color, no teams, unique asymmetric roles (spirits) via draft or free-pick, single scenario.

**Setup:** Players join → select spirits (no duplicates) → expansion modules configured → play.

**How it maps:**
- `behaviors.PlayerRole` with a spirit enum
- `moves.SelectRole` with uniqueness enforcement (same as `SelectColor` uniqueness)
- Expansion modules = variant at creation time
- `ReadyToStart` validates: all seated players have selected a spirit, no duplicates

**Design fit: Covered.** `SelectRole` needs the same uniqueness logic as `SelectColor`. The `ReadyToStart` hook validates. The gathering panel auto-detects `SelectRole` and shows a spirit picker.

**Question:** Should `SelectRole` support a `WithUnique()` configuration option (like `SelectColor` enforces uniqueness by default)? Probably yes — the move should take a `WithUniqueSelection()` option.

---

### 10. Diplomacy (Fixed Asymmetric Positions)
**Dimensions:** Exactly 7, functional position (country = power set), no teams, simultaneous secret orders, single long game.

**Setup:** 7 players join → country assignment (fixed by seat, random, or drafted) → play.

**How it maps:**
- Country assignment = role assignment. `behaviors.PlayerRole` with a country enum.
- If fixed by seat: a FixUp sets roles from player index. No `SelectRole` needed.
- If drafted: `SelectRole` with uniqueness.
- Requires exactly 7 players — `MinNumPlayers() == MaxNumPlayers() == 7`.

**Design fit: Fully covered.**

---

### 11. Cosmic Encounter (Variable Powers, Shifting Alliances)
**Dimensions:** Range 3-5, functional color, no fixed teams (alliances shift), unique alien powers via draft.

**Setup:** Players join → draft alien powers (deal N, pick 1) → colors assigned → colonies established.

**How it maps:**
- Alien power draft = a game-specific drafting phase (not `SelectRole`, since it involves dealing cards and picking, not just selecting from a list)
- Color = auto-assigned or player-chosen via `SelectColor`
- The draft phase is game-specific: `DealCountComponents` to deal power cards, then a custom `PickPower` move

**Design fit: Partially covered.** The draft mechanic is game-specific (the framework provides `DealCountComponents` but not "pick one from your dealt hand and discard the rest"). Colors work. The gathering panel wouldn't show a power picker — the game renderer would handle the draft UI. **This is fine — the draft is game logic, not framework lobby logic.**

---

### 12. Captain Sonar (Teams + Roles Within Teams, Fixed Team Size)
**Dimensions:** Exactly 8 (or 2-4 variant), fixed-2 teams of exactly 4, 4 unique roles per team, admin start.

**Setup:** 8 players join → split into 2 teams of 4 → within each team, assign 4 unique roles (Captain, First Mate, Engineer, Radio Operator).

**How it maps:**
- `behaviors.PlayerTeam` (2 teams) + `behaviors.PlayerRole` (4 roles)
- `SelectTeam` + `SelectRole` in gathering phase
- `ReadyToStart` validates: each team has exactly 4 players, each team has all 4 roles filled, no duplicate roles within a team
- Roles are unique **within** a team but shared **across** teams (both teams have a Captain)

**Design fit: Covered with custom `ReadyToStart`.** The complex constraint (unique per team, not globally unique) is game-specific validation. `SelectRole` doesn't enforce uniqueness in this case — `ReadyToStart` does.

**Question:** Should `SelectRole` / `SelectColor` uniqueness be configurable? Options: `WithUnique()` (globally unique), `WithUniquePerTeam()` (unique within team), no uniqueness. Or just let `ReadyToStart` handle it all. **Recommendation:** Keep the moves simple. Uniqueness enforcement in `ReadyToStart`, not in the move itself. The move just sets the value; validation is the delegate's job.

---

### 13. Settlers of Catan (Color Selection, Expansion Modules)
**Dimensions:** Range 3-6, player-chosen color, no teams, no roles, expansion variant.

**Setup:** Expansion chosen at creation → players join → pick colors → board randomized → initial placement phase.

**How it maps:**
- `SelectColor` with uniqueness. Colors are functional (piece identity).
- Expansion = variant at creation time.
- Board randomization + initial placement = game-specific phases after gathering.

**Design fit: Fully covered.**

---

### 14. One Night Ultimate Werewolf (Secret Roles, Role Swapping, No Moderator)
**Dimensions:** Range 3-10, no color, secret roles randomly assigned, roles mutate during night phase, single round.

**Setup:** Players join → admin selects which roles to include (variant) → roles randomly dealt (N+3: N to players, 3 to center) → night phase.

**How it maps:**
- Which-roles-to-include = variant at creation time (or a custom "role deck builder" move)
- Random assignment = FixUp move after `WaitForEnoughPlayers`
- Role mutation during night = game-specific moves (swap, peek, etc.)
- Sanitization: each player sees only their own initial role. Center cards hidden from all.

**Design fit: Covered.** The "which roles to include" could be a variant. Random assignment is a FixUp. The night phase is game logic. The interesting bit is N+3 roles — the game needs `NumSeatedActivePlayers + 3` role cards. The game handles this in its assignment logic.

---

### 15. Pandemic Legacy (Campaign, Persistent State)
**Dimensions:** Range 2-4, no color, no teams, role selection from evolving pool, campaign (multi-session persistence).

**Setup per session:** Load campaign state → select roles from available pool → play scenario.

**How it maps:**
- Campaign persistence = **out of scope** for the framework's game layer. This is server/storage layer logic (save/load campaign state across Game objects).
- Role selection from a pool = `SelectRole` where the available roles are game-state-dependent (some roles may have been "destroyed" in previous sessions)
- Per-session setup otherwise maps to the standard gathering flow

**Design fit: Session setup is covered. Campaign persistence is explicitly out of scope** (as established in our design discussion). A campaign system would sit above `GameManager`.

---

### 16. Decrypto / Wavelength (Teams, Rotating Clue-Giver)
**Dimensions:** Range 4-8, no color, fixed-2 teams, role rotates each round (clue-giver), multi-round.

**Setup:** Players join → pick teams → game starts. Clue-giver rotates automatically.

**Between rounds:** Clue-giver role advances to next player on each team. This is a FixUp, not a player choice. Teams stay fixed.

**How it maps:**
- `behaviors.PlayerTeam` + `SelectTeam`
- Rotating clue-giver = game-specific logic using `behaviors.RoundRobin` or similar per-team tracking
- `ReadyToStart` validates: each team ≥ 2 players

**Design fit: Fully covered.** The rotating role is game logic, not gathering logic.

---

### 17. D&D / RPG Session (Moderator, Admin-Assigned)
**Dimensions:** Range 1-DM + 3-6, no color, no teams, admin-assigned complex roles, persistent campaign.

**Setup:** DM creates → players join → DM approves characters → DM prepares adventure → start.

**How it maps:**
- DM = player 0 with a moderator flag (game-specific behavior)
- Character creation/approval = entirely game-specific (too complex for framework moves)
- Campaign persistence = out of scope (same as Pandemic Legacy)

**Design fit: Gathering (player joining + DM starts) is covered. Character creation is game-specific.** The framework's gathering panel shows "Waiting for Players" + "Start Game" for the DM. Everything else is the game's custom renderer.

---

### 18. Blood on the Clocktower (Admin-Assigned Roles, Complex Script)
**Dimensions:** Range 5-20, no color, secret teams, admin-assigned roles from a "script" (variant), eliminated players still participate.

**Setup:** Storyteller creates with script selection (variant) → players join → Storyteller assigns roles → start.

**How it maps:**
- Script = variant at creation time
- Role assignment = admin-to-player assignment. NOT `SelectRole` (players don't choose). A game-specific `AdminAssignRole` move where the Storyteller targets a player and assigns a role.
- Sanitization: each player sees only their own role. The Storyteller sees all.
- Elimination with continued participation: the `InactivePlayer` behavior might not be right here (eliminated players still vote). Game-specific.

**Design fit: Gathering (player joining + admin start) is covered. Role assignment is game-specific.** The Storyteller's custom assignment UI lives in the game renderer. This is a Level 2 override scenario — the game provides `boardgame-render-gathering-clocktower`.

---

### 19. MtG Draft (Draft Phase → Deck Building → Tournament)
**Dimensions:** Range 2-8, no color, no teams, draft-pick unique components (not roles), multi-phase structure.

**Setup:** Players join → draft phase (open packs, pick cards, pass) → deck building → games.

**How it maps:**
- The entire draft phase is game-specific (framework has `DealCountComponents` but not "pick one and pass")
- Deck building is game-specific
- The tournament structure (Swiss/round-robin) = out of scope (meta-game orchestration)

**Design fit: Gathering (player joining + start) is covered. The draft is game logic.** This is fine — drafting is a gameplay mechanic, not a lobby/setup concern.

---

### 20. Twilight Imperium (Everything Complex)
**Dimensions:** Range 3-6, functional color, no fixed teams (shifting alliances), unique factions via draft, complex map setup.

**Setup:** Players join → faction draft (ban + snake-pick) → map construction (player tile placement or preset) → speaker token assigned → play.

**How it maps:**
- Faction draft = a game-specific drafting phase (like Cosmic Encounter)
- Map construction = game-specific
- Speaker token = game-specific first-player logic
- Everything after "players join" is game-specific

**Design fit: Gathering (player joining + admin start) is covered. Everything else is game logic.** This is the game where the full override (`boardgame-render-gathering-ti`) makes sense.

---

### 21. Jackbox / Party Games (Quick Join, Spectators, Rounds)
**Dimensions:** Range 3-8 + spectators, no color, no teams (usually), no roles, quick rounds, admin start.

**Setup:** Host creates → room code displayed → players join on phones → host starts.

**Between rounds:** New players can join. Quick turnaround. Mini-game selection might happen.

**How it maps:**
- `WithManualStart()` — host controls start
- Share link = room code (the gathering panel's share link)
- Spectators = observers (already supported via `ObserverPlayerIndex`)
- Between rounds: phase cycles back to gathering, `ActivateEmptySeat` reopens seats
- Mini-game selection between rounds = game-specific

**Design fit: Fully covered** for the core pattern. Mini-game selection is game-specific.

---

### 22. Gloomhaven (Character Selection from Unlocked Pool)
**Dimensions:** Range 1-4, no color, cooperative, role selection from an evolving pool.

**Setup per scenario:** Select scenario → pick characters from unlocked pool → equip items → select card hand → play.

**How it maps:**
- Scenario selection = variant or game-state-dependent phase
- Character selection = `SelectRole` where available roles are filtered by campaign state
- Item equipping + card hand selection = game-specific phases after gathering

**Design fit: Gathering is covered. Character customization is game-specific.**

---

### 23. Mysterium (Asymmetric Cooperative, One vs Many)
**Dimensions:** Range 2-7, no color, no teams, one special role (Ghost) player-chosen or admin-assigned.

**Setup:** Players join → one player becomes the Ghost → remaining are Psychics → play.

**How it maps:**
- The Ghost role could be assigned via `SelectRole` (a player volunteers)
- Or admin-assigned
- `ReadyToStart` validates: exactly 1 Ghost, rest are Psychics
- The Ghost sees different state than Psychics (sanitization handles this via group membership)

**Design fit: Covered.** `SelectRole` for volunteering. `ReadyToStart` for validation. Sanitization for information asymmetry.

---

### 24. Secret Hitler (Secret Teams, Random)
**Dimensions:** Range 5-10, no color, secret teams random, secret roles random (Liberal/Fascist/Hitler), single game.

**Setup:** All players join → roles randomly assigned → play. Identical to The Resistance pattern.

**Design fit: Covered** (same as Scenario 5). FixUp auto-assigns roles.

---

### 25. Mahjong (Exactly 4, Functional Seating Position)
**Dimensions:** Exactly 4, no color, no teams, seating position is functionally significant (East/South/West/North), multi-round.

**Setup:** 4 players join → seat positions assigned (ritual/random) → tiles dealt → play.

**Between rounds:** East position rotates. This is automatic.

**How it maps:**
- Seat position = `behaviors.PlayerRole` with a wind enum (East/South/West/North)
- Position assignment = auto-assigned based on seat index (FixUp) or random
- Position rotation between rounds = FixUp move in tally phase

**Design fit: Fully covered.** Position is just a role that rotates.

---

### 26. Tournament / Swiss Bracket (Meta-Game)
**Dimensions:** N players, multiple games, pairings change between rounds.

**How it maps:** **Out of scope.** This is a meta-game system that creates and manages multiple Game objects. It sits above `GameManager` in the server layer. Each individual game within the tournament uses the normal gathering flow.

**Design fit: Individual games are covered. The tournament orchestration is out of scope.**

---

## Coverage Matrix

| # | Scenario | Gathering Panel | SelectTeam | SelectRole | SelectColor | ReadyToStart | Game-Specific Moves | Override Needed |
|---|----------|----------------|------------|------------|-------------|--------------|-------------------|----------------|
| 1 | Chess | Status only | - | - | - | - | - | No |
| 2 | Blackjack | Status + Start | - | - | - | - | - | No |
| 3 | Memory | Status | - | - | - | - | - | No |
| 4 | Codenames | Status + Start + Team + Role | Yes | Yes | - | Yes (complex) | - | No |
| 5 | Resistance | Status only | - | - | - | - | AssignSecretRoles (FixUp) | No |
| 6 | Werewolf | Status + Start | - | - | - | Yes | AdminAssignRole | Maybe |
| 7 | Risk | Status + Start + Color | - | - | Yes | - | - | No |
| 8 | Poker | Status (recurring) | - | - | - | - | - | No |
| 9 | Spirit Island | Status + Start + Role | - | Yes (unique) | - | Yes | - | No |
| 10 | Diplomacy | Status + Role (or auto) | - | Maybe | - | - | - | No |
| 11 | Cosmic Encounter | Status + Start | - | - | Maybe | - | DraftPower | Yes (draft UI) |
| 12 | Captain Sonar | Status + Start + Team + Role | Yes | Yes | - | Yes (complex) | - | No |
| 13 | Catan | Status + Start + Color | - | - | Yes | - | - | No |
| 14 | ONUW | Status | - | - | - | - | AssignSecretRoles | No |
| 15 | Pandemic Legacy | Status + Role | - | Yes | - | - | - | No |
| 16 | Decrypto | Status + Start + Team | Yes | - | - | Yes | - | No |
| 17 | D&D | Status + Start | - | - | - | - | Character creation | Yes (full) |
| 18 | Clocktower | Status + Start | - | - | - | Yes | AdminAssignRole | Yes (role assignment) |
| 19 | MtG Draft | Status + Start | - | - | - | - | Draft moves | Yes (draft UI) |
| 20 | TI | Status + Start | - | - | - | - | Faction draft, map build | Yes (full) |
| 21 | Jackbox | Status + Start + Share | - | - | - | - | - | No |
| 22 | Gloomhaven | Status + Start + Role | - | Yes | - | - | Card/item selection | Maybe |
| 23 | Mysterium | Status + Start + Role | - | Yes | - | Yes | - | No |
| 24 | Secret Hitler | Status only | - | - | - | - | AssignSecretRoles | No |
| 25 | Mahjong | Status | - | - | - | - | AssignPosition (FixUp) | No |
| 26 | Tournament | N/A | - | - | - | - | N/A | Out of scope |

## Summary

**Fully handled by framework defaults (zero game code):** 8 scenarios (1, 2, 3, 8, 14, 21, 24, 25)

**Handled by framework building blocks (embed behavior + 1-3 lines):** 10 scenarios (4, 7, 9, 10, 12, 13, 15, 16, 22, 23)

**Framework gathering + game-specific moves:** 4 scenarios (5, 6, 11, 19)

**Full or partial override needed:** 3 scenarios (17, 18, 20)

**Out of scope:** 1 scenario (26 — tournament orchestration)

## Design Implications

### Confirmed decisions:
1. **`SelectTeam` and `SelectRole` should NOT enforce uniqueness by default.** Captain Sonar has roles unique per team but shared across teams. Spirit Island has globally unique roles. The constraint varies. **Let `ReadyToStart` handle validation.**

2. **`SelectColor` SHOULD enforce uniqueness by default** (with a `WithAllowDuplicateColors()` escape hatch). In every scenario with player colors (Risk, Catan), colors are unique. This is a safe default.

3. **Secret role assignment is always game-specific.** The framework should not try to generalize it. It's too varied (random, admin-assigned, depends on player count, may involve dealing from a deck).

4. **Admin-assigns-to-player is a game-specific move pattern.** The framework's `SelectTeam`/`SelectRole` handle self-selection. Admin assigning roles to specific players (Werewolf moderator, Clocktower Storyteller) is game logic.

5. **The "recurring gathering" pattern (Poker, Jackbox) works naturally** via phase cycling + `ActivateEmptySeat`. The one gap is server-side: re-opening `eGame.Open` when seats reopen between rounds.

6. **Draft-pick is game logic, not framework logic.** MtG Draft, Cosmic Encounter, TI faction selection — these all involve dealing components and making sequential picks. Too varied for framework moves.

### New considerations from scenarios:

7. **`SelectRole` should support a `WithUnique()` option** for the Spirit Island pattern (globally unique). This is a convenience — `ReadyToStart` could also check it, but `Legal()` rejecting a duplicate pick gives immediate feedback.

8. **The "admin can override player choices" pattern** (admin reassigns someone's color, kicks someone off a team) is a separate move type. Not `SelectColor` (that's self-selection). A custom `AdminOverrideColor` or generic `AdminSetPlayerProperty` move. Probably game-specific for now.

9. **Rotating roles between rounds** (Decrypto clue-giver, Mahjong wind position) is not a gathering concern — it's a FixUp in the tally phase. The framework doesn't need to handle it in the gathering system.

10. **The framework's gathering UI handles 18 of 25 in-scope scenarios with zero or minimal game code.** The 4 scenarios needing game-specific moves are fine — they're genuinely unique. The 3 needing full override are the most complex games (D&D, Clocktower, TI).
