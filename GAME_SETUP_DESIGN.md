# Game Setup Architecture: Lobby-as-Phase Design

## Problem Statement

The framework has a rigid game creation flow:

1. Server receives a request with `numPlayers`, `variant`, and `agents` fully specified
2. `GameManager.NewGame()` calls `Game.setUp()` which validates players/variants, creates state, distributes components, and starts the game
3. The game is immediately live with all its phase progressions running
4. Players join afterwards and are seated via `moves.SeatPlayer`

This means:
- **Player count is immutable** after creation. If you create a 4-player game but only 3 friends show up, you're stuck.
- **Variants must be decided before anyone sees the game.** You can't create a game, invite friends, and let everyone vote on settings.
- **Team/role assignment has no lobby.** There's no in-game flow for "players pick teams, then admin starts."
- **Multi-round re-setup is awkward.** After a round, you might want to re-open seats or let new players join, but there's no phase-based pattern for this.

### What Already Works

The seating infrastructure is ~95% done (#755):
- `behaviors.Seat` (SeatFilled, SeatClosed) and `behaviors.InactivePlayer`
- `moves.SeatPlayer`, `WaitForEnoughPlayers`, `CloseAllSeats`, `InactivateEmptySeat`, `ActivateInactivePlayer`
- `DefaultRoundSetup(auto, WithManualStart())` for player-initiated game start
- `base.GameDelegate.NumSeatedActivePlayers()` and `PlayerMayBeActive()`
- Client roster already shows "Waiting to be seated" / "Sitting out" with proper dimming
- Server auto-seats first player, has join flow, auto-closes full games

### Related Issues

| Issue | Title | Role in This Design |
|-------|-------|-------------------|
| #754 | SetUp phases in server layer | Core of this design |
| #753 | Variants in server layer, not core | Subsumed into this design |
| #752 | Complex team selection | Enabled by this design |
| #755 | Rationalize seats/active players | ~95% complete; foundation |
| #768 | Manual start | Merged; `WithManualStart()` |
| #771 | GameAdministrator behavior | Orthogonal; designed here |
| #774 | Disable SeatPlayer in debug | Small; designed here |
| #767 | Client SeatPlayer rendering | Partially done; completed here |

---

## Design Overview

**Core principle: the lobby IS a game phase.**

A game that wants lobby functionality defines a `SetUp` phase as its first phase. During this phase, special lobby moves (variant selection, team assignment, "Start Game") are legal. The server detects the SetUp phase and shows lobby UI instead of the game renderer. When setup is finalized, the game transitions to its first "real" phase (e.g., InitialDeal).

This reuses all existing machinery: phases, moves, behaviors, state, sanitization. The lobby is just another phase of the game, with its own moves and its own client rendering.

### High-Level Flow

```
 CREATE GAME               SETUP PHASE                    GAMEPLAY
 -----------               -----------                    --------
 NewGame(maxPlayers,   -->  Phase: SetUp              --> Phase: InitialDeal
   defaultVariant,          - Players join (SeatPlayer)    - Components dealt
   noAgents)                - Variants tweaked             - Normal game moves
                            - Teams/roles assigned
                            - Admin clicks "Start"
                            - FinalizeSetUp move fires
                            - Empty seats closed/inactivated
                            - Transition to first real phase
```

For multi-round games, the round cleanup phase can transition back to SetUp, re-opening seats and allowing new players.

---

## Detailed Design

### 1. Game Creation Changes

#### Current flow (`game_manager.go:636-652`)
```go
func (g *GameManager) NewGame(numPlayers int, variant map[string]string, agents []string) (*Game, error) {
    return g.createGame("", "", numPlayers, variant, agents)
}
```

#### Proposed: No changes to NewGame

`NewGame` stays identical. For lobby games, the **server** calls it differently:

```go
// Server creates a lobby game:
game, err := manager.NewGame(
    manager.Delegate().MaxNumPlayers(),  // Create max seats; unfilled ones get inactivated
    nil,                                  // Default variant values
    nil,                                  // No agents yet (assigned during setup)
)
```

The game delegate's `BeginSetUp` detects lobby mode and sets the phase to SetUp:

```go
func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
    game := state.GameState().(*gameState)

    // Transcribe variant defaults into game state (as today)
    game.MaxRounds = variantToMaxRounds(variant)

    // Start in SetUp phase (lobby mode)
    game.Phase.SetValue(phaseSetUp)

    return nil
}
```

**Why no new API?** The existing `NewGame` already handles everything. The differences for lobby games are:
- Pass `MaxNumPlayers` instead of a specific count
- Pass `nil` variant (use defaults, can be changed during SetUp)
- The delegate's `BeginSetUp` sets phase to SetUp

This avoids any core engine changes for game creation.

### 2. The SetUp Phase

#### Phase Definition

Games define a `phaseSetUp` in their phase enum:

```go
const (
    phaseSetUp       = iota  // NEW: Lobby/setup phase
    phaseInitialDeal         // Existing first "real" phase
    phaseNormalPlay
    phaseRoundCleanup
)
```

#### SetUp Phase Behavior

A new behavior to embed in gameState:

```go
// behaviors/setup_config.go

// SetUpConfig is designed to be embedded in your gameState when using
// lobby-style setup. It tracks setup finalization state and provides
// a signal to the server layer that the game is in setup mode.
type SetUpConfig struct {
    SetUpFinalized bool
}

func (s *SetUpConfig) IsSetUpFinalized() bool {
    return s.SetUpFinalized
}

func (s *SetUpConfig) MarkSetUpFinalized() {
    s.SetUpFinalized = true
}
```

This is intentionally minimal. The phase system already tracks what phase we're in. This behavior exists to:
1. Prevent double-finalization
2. Give the server a fast signal ("is this game in setup?") without parsing the phase enum
3. Give moves a clean interface to check

#### Corresponding interface in `moves/interfaces/`

```go
type SetUpConfigurer interface {
    IsSetUpFinalized() bool
    MarkSetUpFinalized()
}
```

### 3. New Moves

#### 3a. `moves.FinalizeSetUp`

The "Start Game" button. When applied, it closes empty seats, inactivates them, marks setup as finalized, and triggers the phase transition.

```go
// moves/finalize_setup.go

type FinalizeSetUp struct {
    FixUp
}

func (f *FinalizeSetUp) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := f.FixUp.Legal(state, proposer); err != nil {
        return err
    }

    // Check not already finalized
    if configurer, ok := state.ImmutableGameState().(interfaces.SetUpConfigurer); ok {
        if configurer.IsSetUpFinalized() {
            return errors.New("Setup is already finalized")
        }
    }

    // Check enough players
    del := state.Manager().Delegate()
    numSeated := base.NumSeatedActivePlayers(state)
    if numSeated < del.MinNumPlayers() {
        return errors.New("Need at least " + strconv.Itoa(del.MinNumPlayers()) +
            " players, but only " + strconv.Itoa(numSeated) + " are seated")
    }

    // Call delegate validation if it implements it
    if validator, ok := del.(SetUpValidator); ok {
        if err := validator.ValidateSetUp(state); err != nil {
            return err
        }
    }

    return nil
}

func (f *FinalizeSetUp) Apply(state boardgame.State) error {
    // Mark finalized
    if configurer, ok := state.GameState().(interfaces.SetUpConfigurer); ok {
        configurer.MarkSetUpFinalized()
    }

    // Close and inactivate all empty seats
    for _, p := range state.PlayerStates() {
        seater, hasSeat := p.(interfaces.Seater)
        if !hasSeat {
            continue
        }
        if !seater.SeatIsFilled() {
            seater.(mutableSeater).SetSeatClosed()
            if inactiver, ok := p.(interfaces.PlayerInactiver); ok {
                inactiver.SetPlayerInactive()
            }
        }
    }

    return nil
}
```

**Note:** This move replaces the `CloseAllSeats` + `InactivateEmptySeat` + phase transition that `DefaultRoundSetup` currently handles separately. It bundles them into one move for the lobby pattern.

**Open question: Should FinalizeSetUp be a FixUp or a player move?**

- As a **player move**: The admin/creator clicks "Start Game." This is more explicit.
- As a **FixUp**: It auto-fires when all conditions are met (enough players, all variants set, etc.). This is more like the current `WaitForEnoughPlayers` pattern.

**Recommendation: Player move.** The whole point of the lobby is to give the admin control over when to start. Auto-starting defeats the purpose. `WaitForEnoughPlayers` already handles the auto-start case.

#### 3b. `moves.SetVariantValue`

A player move for changing variant selections during SetUp.

```go
// moves/set_variant.go

type SetVariantValue struct {
    Default
    VariantKey   string
    VariantValue string
}

func (s *SetVariantValue) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := s.Default.Legal(state, proposer); err != nil {
        return err
    }

    // Validate key exists in VariantConfig
    config := state.Manager().Variants()
    if config == nil {
        return errors.New("Game has no variant configuration")
    }
    variantKey := config[s.VariantKey]
    if variantKey == nil {
        return errors.New("Unknown variant key: " + s.VariantKey)
    }

    // Validate value is legal for this key
    if _, ok := variantKey.Values[s.VariantValue]; !ok {
        return errors.New("Illegal value '" + s.VariantValue + "' for variant " + s.VariantKey)
    }

    return nil
}

func (s *SetVariantValue) Apply(state boardgame.State) error {
    // Delegate to game-specific logic
    changer, ok := state.Manager().Delegate().(interfaces.VariantApplier)
    if !ok {
        return errors.New("Delegate doesn't implement VariantApplier")
    }
    return changer.ApplyVariantValue(state, s.VariantKey, s.VariantValue)
}
```

#### Corresponding interface

```go
// moves/interfaces/

// VariantApplier is for game delegates that allow variant values to be
// changed during the SetUp phase. The delegate should update the relevant
// game state properties.
type VariantApplier interface {
    ApplyVariantValue(state boardgame.State, key, value string) error
}
```

#### Example implementation in memory

```go
func (g *gameDelegate) ApplyVariantValue(state boardgame.State, key, value string) error {
    game, _ := concreteStates(state)

    switch key {
    case variantKeyCardSet:
        game.CardSet = value
    case variantKeyNumCards:
        switch value {
        case numCardsSmall:
            game.NumCards = 10
        case numCardsMedium:
            game.NumCards = 20
        case numCardsLarge:
            game.NumCards = 40
        }
        // Resize stacks to match
        if err := game.HiddenCards.SetSize(game.NumCards); err != nil {
            return err
        }
        if err := game.VisibleCards.SetSize(game.NumCards); err != nil {
            return err
        }
    }
    return nil
}
```

### 4. The Component Distribution Problem

This is the hardest design challenge. Currently:

1. `BeginSetUp` sizes stacks based on variant values (memory: `HiddenCards.SetSize(game.NumCards)`)
2. The engine distributes ALL components into stacks via `DistributeComponentToStarterStack`
3. `FinishSetUp` does final shuffling/arrangement

If variants change during SetUp, component distribution may be wrong.

#### Proposed Solution: Deferred Distribution via Phases

For lobby games, **component distribution happens in a phase, not during setUp()**.

**Pattern:**
1. `BeginSetUp`: Sizes stacks at their MAXIMUM capacity. All variant properties set to defaults.
2. `DistributeComponentToStarterStack`: Returns a staging stack (e.g., `UnusedCards`) for all components
3. `FinishSetUp`: Minimal prep (no component arrangement)
4. **SetUp phase**: Variants can change, resizing stacks as needed
5. **After FinalizeSetUp**: A "Deal" or "InitialDeal" phase uses moves to distribute components from the staging stack into the right places

This is already the pattern blackjack uses:
- `FinishSetUp` shuffles the DrawStack
- `phaseInitialDeal` has `DealCountComponents` moves to deal from DrawStack to players

For memory, this would mean:
- `FinishSetUp` puts all cards in `UnusedCards` and shuffles
- After FinalizeSetUp, a new phase deals the right cards from `UnusedCards` to `HiddenCards`

**Key insight: games that want lobby mode need to move their component distribution from `FinishSetUp` into a phase progression.** This is a migration cost but results in a cleaner architecture anyway (setup logic is visible as moves, not hidden in a delegate hook).

#### What About Games That Don't Want Lobbies?

Games that return `SetUpModeNone` (the default) work exactly as they do today. `BeginSetUp`, `DistributeComponentToStarterStack`, `FinishSetUp` are called normally. No changes.

#### Alternative Considered: Engine-Level Deferred Distribution

We considered adding a `Game.DistributeComponents()` method callable from a move's Apply(). This would let the FinalizeSetUp move trigger component distribution after variants are finalized.

**Rejected because:**
- Requires core engine changes (distribution is currently internal)
- Component distribution reads from `DistributeComponentToStarterStack` delegate method, which returns one stack per component -- calling it twice would try to double-distribute
- The phase-based approach is more explicit and testable

### 5. Server Layer Changes

#### Detecting Lobby Mode

The server needs to know if a game is in its SetUp phase to show lobby UI.

**Option A: Check phase name**
```go
func (s *Server) isGameInSetUp(game *boardgame.Game) bool {
    phase := game.CurrentState().ImmutableGameState()
    if configurer, ok := phase.(interfaces.SetUpConfigurer); ok {
        return !configurer.IsSetUpFinalized()
    }
    return false
}
```

**Option B: Check a delegate method**
```go
// On GameDelegate interface:
SetUpMode() SetUpMode  // returns SetUpModeNone or SetUpModeLobby
```

**Recommendation: Use both.** `SetUpMode()` on the delegate tells the server "this game type supports lobbies." `IsSetUpFinalized()` on the game state tells the server "this specific game is currently in setup." The delegate method is needed so the server can offer "Create Lobby Game" vs "Create Game" in the UI.

#### New Server Endpoints

```
POST /api/game/new-lobby
  Body: { manager: "blackjack" }
  Creates game with MaxNumPlayers and default variants.
  Returns: { GameID, GameName }

GET /api/game/{name}/{id}/lobby-info
  Returns: { InSetUp: true, VariantConfig: {...}, CurrentVariant: {...},
             Players: [...], MinPlayers: N, MaxPlayers: N }
  Only works when game is in SetUp phase.
```

The existing `/api/game/new` continues to work for non-lobby game creation.

#### Server Game View Changes

When the game is in SetUp phase, the server includes extra data in the game info response:

```go
func (s *Server) gameInfoHandler(c *gin.Context) {
    // ... existing logic ...

    // Add lobby info if game is in SetUp
    if s.isGameInSetUp(game) {
        result["InSetUp"] = true
        result["VariantConfig"] = manager.Variants()
        // Current variant values are in game state, accessible via
        // ComputedGlobalProperties
    }
}
```

### 6. Client Layer Changes

#### Lobby UI Component

A new component `boardgame-lobby.ts` renders when the game is in SetUp:

```
+----------------------------------------------+
|  Blackjack Lobby                        [X]  |
+----------------------------------------------+
|                                               |
|  Players (2 of 4-7 needed)                   |
|  [*] Alice (Host)    [*] Bob                 |
|  [ ] Empty seat      [ ] Empty seat          |
|                                               |
|  Share link: [http://...game/bj/abc ][Copy]  |
|                                               |
|  Settings                                     |
|  Max Rounds: [  5  v]                        |
|                                               |
|  [Start Game]  (need 2 more players)         |
|                                               |
+----------------------------------------------+
```

**Key client behaviors:**
- Shows current players and empty seats (from existing roster data)
- Shows variant selectors (driven by `VariantConfig` from server + current values from game state)
- "Start Game" button proposes `FinalizeSetUp` move (only enabled when legal)
- Variant changes propose `SetVariantValue` moves
- Falls back to normal game rendering after finalization

**Implementation approach:** The existing `boardgame-game-view.ts` checks whether the game is in SetUp:

```typescript
// In boardgame-game-view.ts render():
if (this._inSetUp) {
    return html`<boardgame-lobby ...></boardgame-lobby>`;
} else {
    return html`<boardgame-render-game ...></boardgame-render-game>`;
}
```

The `_inSetUp` property comes from computed game properties or a flag in the game info response.

### 7. Permission System (#771)

#### Design

A new behavior and a move-configuration option:

```go
// behaviors/permission.go

type Permission struct {
    PermissionLevel int  // 0 = player, 1 = admin
}

func (p *Permission) GetPermissionLevel() int {
    return p.PermissionLevel
}

func (p *Permission) SetPermissionLevel(level int) {
    p.PermissionLevel = level
}
```

Default initialization: player 0 gets `PermissionLevel = 1` (admin). All others get 0.

This is done in the SeatPlayer move: when seating the first player (index 0), set their permission to admin.

#### Move Configuration

```go
// moves/with.go

func WithRequiredPermission(level int) CustomConfigurationOption {
    return func(config boardgame.PropertyCollection) {
        config[configPropRequiredPermission] = level
    }
}
```

The `Default.Legal()` method checks permission:

```go
func (d *Default) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    // ... existing checks ...

    // Permission check
    if requiredLevel, ok := d.CustomConfiguration()[configPropRequiredPermission]; ok {
        if proposer == boardgame.AdminPlayerIndex {
            return nil  // Engine admin always passes
        }
        if proposer < 0 {
            return errors.New("Permission denied: no valid proposer")
        }
        player := state.ImmutablePlayerStates()[proposer]
        if checker, ok := player.(interfaces.PermissionChecker); ok {
            if checker.GetPermissionLevel() < requiredLevel.(int) {
                return errors.New("Insufficient permission level")
            }
        }
    }

    return nil
}
```

#### Usage

```go
auto.MustConfig(
    new(moves.FinalizeSetUp),
    moves.WithLegalPhases(phaseSetUp, phaseEnum),
    moves.WithRequiredPermission(1),  // Admin only
)

auto.MustConfig(
    new(moves.SetVariantValue),
    moves.WithLegalPhases(phaseSetUp, phaseEnum),
    moves.WithRequiredPermission(1),  // Admin only (or 0 for any player)
)
```

### 8. Debug Auto-Seating (#774)

When `DisableAdminChecking` is true (dev mode), the server auto-fills all player slots at game creation.

```go
// In server/api/main.go, after doNewGame creates the game:

func (s *Server) doNewGame(...) {
    game, err := manager.NewGame(numPlayers, variant, agents)
    // ... existing logic ...

    // Debug auto-seating: fill remaining slots with synthetic users
    if s.config.DisableAdminChecking {
        for i := 1; i < numPlayers; i++ {
            if agents[i] != "" {
                continue  // Skip agent slots
            }
            debugUser := s.getOrCreateDebugUser(i)
            if err := s.doSeatPlayer(game, boardgame.PlayerIndex(i), debugUser); err != nil {
                s.logger.Warnln("Debug auto-seat failed for slot", i, ":", err)
            }
        }
    }
}

func (s *Server) getOrCreateDebugUser(index int) *users.StorageRecord {
    id := fmt.Sprintf("debug-player-%d", index)
    user, err := s.storage.GetUser(id)
    if err != nil || user == nil {
        user = &users.StorageRecord{
            ID:          id,
            DisplayName: fmt.Sprintf("Debug Player %d", index),
        }
        s.storage.UpdateUser(user)
    }
    return user
}
```

### 9. Client Roster Improvements (#767)

The roster item already shows "Waiting to be seated" and "Sitting out." The remaining gaps:

#### "Waiting for Players" Banner

In `boardgame-player-roster.ts`, update `_bannerText`:

```typescript
private _bannerText(finished: boolean, winners: number[], hasEmptySlots: boolean, inSetUp: boolean): string {
    if (finished) return "Game Over";
    if (inSetUp) return "Setting Up";
    if (hasEmptySlots) return "Waiting for Players";
    return "Playing";
}
```

#### Share/Invite Link

Add a "Copy invite link" button when the game has empty slots:

```typescript
${when(this.hasEmptySlots && !this.isObserver, () => html`
    <md-outlined-button @click="${this._copyInviteLink}">
        Copy invite link
    </md-outlined-button>
`)}
```

---

## How a Game Opts In: Full Example (Blackjack)

### Before (current)

```go
const (
    phaseInitialDeal = iota
    phaseNormalPlay
    phaseRoundCleanup
)

func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
    game, _ := concreteStates(state)
    game.MaxRounds = variantToMaxRounds(variant)
    return nil
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    return moves.Combine(
        moves.Add(
            auto.MustConfig(new(moves.SeatPlayer)),
        ),
        moves.AddOrderedForPhase(phaseInitialDeal,
            moves.DefaultRoundSetup(auto),
            // ... deal moves ...
            auto.MustConfig(new(moves.StartPhase),
                moves.WithPhaseToStart(phaseNormalPlay, phaseEnum)),
        ),
        // ...
    )
}
```

### After (with lobby)

```go
const (
    phaseSetUp       = iota   // NEW
    phaseInitialDeal
    phaseNormalPlay
    phaseRoundCleanup
)

type gameState struct {
    base.SubState
    behaviors.RoundRobin
    behaviors.CurrentPlayerBehavior
    behaviors.PhaseBehavior
    behaviors.SetUpConfig              // NEW
    // ... existing fields ...
}

func (g *gameDelegate) SetUpMode() boardgame.SetUpMode {
    return boardgame.SetUpModeLobby    // NEW
}

func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
    game, _ := concreteStates(state)
    game.MaxRounds = variantToMaxRounds(variant)
    // Phase starts at phaseSetUp (0, the default iota value)
    return nil
}

func (g *gameDelegate) ApplyVariantValue(state boardgame.State, key, value string) error {
    game, _ := concreteStates(state)
    if key == variantKeyMaxRounds {
        maxRounds, err := strconv.Atoi(value)
        if err != nil {
            return err
        }
        game.MaxRounds = maxRounds
    }
    return nil
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    return moves.Combine(
        // Moves available in all phases
        moves.Add(
            auto.MustConfig(new(moves.SeatPlayer)),
        ),
        // SetUp phase: lobby moves
        moves.AddForPhase(phaseSetUp,
            auto.MustConfig(new(moves.SetVariantValue)),
            auto.MustConfig(new(moves.FinalizeSetUp),
                moves.WithMoveName("Start Game"),
                moves.WithPhaseToStart(phaseInitialDeal, phaseEnum),
                moves.WithRequiredPermission(1),
            ),
        ),
        // InitialDeal phase: deal cards (unchanged)
        moves.AddOrderedForPhase(phaseInitialDeal,
            moves.DefaultRoundSetup(auto),
            // ... deal moves ...
            auto.MustConfig(new(moves.StartPhase),
                moves.WithPhaseToStart(phaseNormalPlay, phaseEnum)),
        ),
        // ... remaining phases unchanged ...
    )
}
```

### What Changed

1. Added `phaseSetUp` as the first phase constant
2. Embedded `behaviors.SetUpConfig` in gameState
3. Added `SetUpMode()` returning `SetUpModeLobby`
4. Added `ApplyVariantValue()` to handle variant changes
5. Added `moves.SetVariantValue` and `moves.FinalizeSetUp` for phaseSetUp
6. Everything else is identical

---

## DefaultLobbySetup Helper

Analogous to `DefaultRoundSetup`, a convenience function:

```go
func DefaultLobbySetup(auto *AutoConfigurer, options ...CustomConfigurationOption) []boardgame.MoveConfig {
    config := boardgame.PropertyCollection{}
    for _, opt := range options {
        opt(config)
    }

    result := []boardgame.MoveConfig{
        auto.MustConfig(new(SetVariantValue)),
    }

    // FinalizeSetUp is always included
    finalizeOpts := []CustomConfigurationOption{}
    if phase, ok := config[configPropPhaseToStart]; ok {
        finalizeOpts = append(finalizeOpts, WithPhaseToStart(phase, ...))
    }
    if _, ok := config[configPropRequiredPermission]; ok {
        finalizeOpts = append(finalizeOpts, WithRequiredPermission(...))
    }

    result = append(result, auto.MustConfig(
        new(FinalizeSetUp),
        WithMoveName("Start Game"),
        finalizeOpts...,
    ))

    return result
}
```

Usage:
```go
moves.AddForPhase(phaseSetUp,
    moves.DefaultLobbySetup(auto,
        moves.WithPhaseToStart(phaseInitialDeal, phaseEnum),
        moves.WithRequiredPermission(1),
    )...,
),
```

---

## Backward Compatibility

**No breaking changes.** Every change is opt-in:

| Component | Non-lobby games | Lobby games |
|-----------|----------------|-------------|
| `NewGame()` API | Unchanged | Called with MaxNumPlayers + nil variant |
| `GameDelegate` interface | New methods with defaults | Override `SetUpMode()`, `ApplyVariantValue()` |
| `BeginSetUp/FinishSetUp` | Unchanged | Same, but phase starts at phaseSetUp |
| `VariantConfig` | Unchanged; stays in core | Also used by `SetVariantValue` move |
| Server endpoints | Unchanged | New `/api/game/new-lobby` endpoint |
| Client | Unchanged | New `boardgame-lobby` component |
| Existing example games | No changes needed | Can be migrated incrementally |

### New defaults in `base.GameDelegate`

```go
func (g *GameDelegate) SetUpMode() boardgame.SetUpMode {
    return boardgame.SetUpModeNone  // Default: no lobby
}

func (g *GameDelegate) ApplyVariantValue(state boardgame.State, key, value string) error {
    return errors.New("Delegate does not support variant changes during setup")
}

func (g *GameDelegate) ValidateSetUp(state boardgame.ImmutableState) error {
    return nil  // Default: no extra validation needed
}
```

---

## Migration Path for Existing Games

### Phase 0: No migration needed
All existing games continue working identically. No code changes.

### Phase 1: Add lobby to one example game
Convert blackjack to use lobby mode as a reference implementation. Blackjack is ideal because:
- Already has multi-round support
- Already uses `behaviors.Seat` and `behaviors.InactivePlayer`
- Its variant (MaxRounds) doesn't affect component distribution
- Its InitialDeal phase already handles component dealing

### Phase 2: Convert games where variants affect distribution
Memory is the harder case (NumCards variant changes stack sizes). For memory, the card-selection logic currently in `FinishSetUp` would move to a new DealCards phase between SetUp and normal play.

### Phase 3: Add team/role selection
For games that need it, add `moves.SelectTeam` / `moves.SelectRole` moves legal during phaseSetUp. These use existing `behaviors.PlayerTeam` and `behaviors.PlayerRole`.

---

## Risks and Mitigations

### Risk 1: Component distribution timing
**Problem:** Games like memory size stacks based on variants in `BeginSetUp`. If variants change during SetUp, stacks may need resizing.

**Mitigation:** `ApplyVariantValue` handles stack resizing. For complex cases, the game can store components in a staging stack during SetUp, then distribute them in a post-SetUp phase. Blackjack already does this (DrawStack -> player hands during InitialDeal).

### Risk 2: `Game.NumPlayers()` vs actual player count
**Problem:** Lobby games create `MaxNumPlayers` slots. `Game.NumPlayers()` returns MaxNumPlayers, but only some seats are filled.

**Mitigation:** This is already handled. Games using `behaviors.Seat` + `behaviors.InactivePlayer` already use `NumSeatedActivePlayers()` for real player count. `PlayerIndex.Next()` already skips inactive players. This is the existing pattern; no new risk.

### Risk 3: Server-side state for pending players
**Problem:** `playersToSeat` is in-memory. If server restarts while a player is pending, the seating is lost.

**Mitigation:** After server restart, reconcile storage's `UserIDsForGame` against game state's `SeatFilled` flags. Re-inject any users assigned to unfilled seats. (This is an existing bug, not new to this design.)

### Risk 4: Client complexity
**Problem:** The lobby UI needs to read VariantConfig, propose moves for variant changes, and handle the transition to game view.

**Mitigation:** The lobby component is relatively simple -- it reads game state and proposes moves, like any game renderer. The `VariantConfig` is already available via the manager info endpoint. The transition from lobby to game view is just a re-render when `InSetUp` goes from true to false.

### Risk 5: Blast radius on GameDelegate interface
**Problem:** Adding `SetUpMode()`, `ApplyVariantValue()`, `ValidateSetUp()` to `GameDelegate` is an interface change.

**Mitigation:** All have default implementations in `base.GameDelegate`. Any delegate embedding `base.GameDelegate` (which is all of them) gets the defaults automatically. No external code breaks.

---

## Implementation Order

```
Phase 1: Foundation                         Phase 2: Server + Client
  behaviors/setup_config.go                   server/api/main.go (lobby endpoint)
  moves/interfaces/ (new interfaces)          server/api/ (lobby detection)
  moves/finalize_setup.go                     boardgame-lobby.ts
  moves/set_variant.go                        boardgame-game-view.ts (lobby switch)
  game_delegate.go (new methods)              boardgame-player-roster.ts (banner)
  base/game_delegate.go (defaults)
  moves/with.go (WithRequiredPermission)    Phase 3: Polish + Examples
  behaviors/permission.go                     Convert blackjack to lobby mode
                                              Debug auto-seating (#774)
                                              Tests for all new moves
```

Phases 1 and 2 can be developed in parallel. Phase 3 depends on both.

---

## Open Questions

1. **Should `FinalizeSetUp` embed `StartPhase` or be separate?** Currently proposed as a separate move that handles seat closing + finalization, with a separate `StartPhase` move for the phase transition. Alternative: embed `StartPhase` behavior so one move does everything.

2. **Should `SetVariantValue` use enum values or strings?** Currently uses strings (matching `VariantConfig`'s string-based values). Alternative: use enum values for type safety, but this would require each variant key to have its own enum.

3. **How should multi-round re-setup work?** When a round ends and the game transitions back to SetUp, should `SetUpFinalized` be reset? Should new players be able to join? The current design supports this but doesn't prescribe the flow.

4. **Should the server offer BOTH "Create Game" and "Create Lobby Game"?** Or should lobby games be the default for game types that support it? If a game supports lobbies, should the old-style instant creation still be available (useful for single-player testing)?

5. **What lobby information should be sanitized?** In the SetUp phase, should observers see the current variant selections? Probably yes (they need it to decide whether to join). Should they see team assignments? Probably yes. This is different from in-game sanitization.

6. **How do agents interact with lobbies?** Can you assign agents during the SetUp phase? Or only humans? If agents are supported, there should be a `moves.AssignAgent` move during SetUp.
