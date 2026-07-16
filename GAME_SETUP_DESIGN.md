# Gathering System: Concrete Design

## Principle

The framework does not have a concept of "lobby" or "setup mode." Instead:

1. **Move legality is the detection mechanism.** The client shows gathering UI when gathering-related moves are legal, and hides it when they aren't.
2. **Phases are just phases.** A gathering phase is a normal phase where gathering moves happen to be legal.
3. **Behaviors trigger auto-detection.** Embed `behaviors.PlayerTeam` in your playerState → the gathering panel auto-shows a team picker.
4. **Everything is overridable.** CSS custom properties hide individual pieces. A custom `boardgame-render-gathering-GAMENAME` element replaces the whole panel.

## Go Changes

### 1. `moves.AnyPlayer` — Base Move for Self-Selection

A new base move type, analogous to `CurrentPlayer` but for phases where any seated player can act. Lives in `moves/any_player.go`.

```go
//boardgame:codegen
type AnyPlayer struct {
    Default
    TargetPlayerIndex boardgame.PlayerIndex
}
```

**`DefaultsForState`:** Sets `TargetPlayerIndex = boardgame.ObserverPlayerIndex` (-1). This is a fail-safe sentinel — if the caller forgets to set it, `Legal` will reject it rather than silently targeting player 0. The client gathering UI components always send `viewingAsPlayer` as `TargetPlayerIndex`. Agents must explicitly set it to their own player index.

**`Legal`:**
```go
func (a *AnyPlayer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := a.Default.Legal(state, proposer); err != nil {
        return err
    }

    target := a.TargetPlayerIndex

    // Target must be a concrete player index (not a sentinel)
    if target < 0 {
        return errors.New("target player must be a seated player, not a special index")
    }

    // Target must be within bounds
    if int(target) >= len(state.ImmutablePlayerStates()) {
        return errors.New("target player index is out of bounds")
    }

    // Proposer must match target (self-selection) OR be admin
    if !target.Equivalent(proposer) {
        return errors.New("you can only make this move for yourself")
    }

    // Target's seat must be filled (if the game uses seating)
    player := state.ImmutablePlayerStates()[target]
    if seater, ok := player.(interfaces.Seater); ok {
        if !seater.SeatIsFilled() {
            return errors.New("your seat is not yet filled")
        }
    }

    return nil
}
```

**Key design decisions:**
- Uses `target.Equivalent(proposer)` instead of `proposer >= 0`. This allows `AdminPlayerIndex` to propose on behalf of any player (since `AdminPlayerIndex` is a wildcard in `Equivalent`), which is essential for the admin debug panel and for `LegalForAnyone` computation. It also naturally rejects `ObserverPlayerIndex`.
- The seat check is conditional on `interfaces.Seater` — games that don't use seating still work.
- `TargetPlayerIndex` defaults to `ObserverPlayerIndex` which safely fails, matching the framework convention that defaults should be "safe to fail" rather than "dangerous to succeed."

**`FallbackName`:** Returns `"Any Player"`.

**Key difference from `CurrentPlayer`:** No check against `state.CurrentPlayerIndex()`. Any seated player can propose, regardless of whose turn it is. This means these moves are naturally compatible with `AnyPlayerIndex` phases and with non-turn-based gathering phases.

**Agent compatibility:** Agents must explicitly set `TargetPlayerIndex = playerIndex` when proposing `AnyPlayer` moves. This is a documented deviation from the `CurrentPlayer` pattern (where `DefaultsForState` handles it automatically). The reason: there is no single "current" player during gathering, so the framework cannot guess the target.

---

### 2. `moves.SelectTeam`

Lives in `moves/select_team.go`. Embeds `AnyPlayer`.

```go
//boardgame:codegen
type SelectTeam struct {
    AnyPlayer
    SelectedTeam enum.Val `enum:"team"`
}
```

**`Legal`:** Calls `AnyPlayer.Legal`, then validates `SelectedTeam` is set, belongs to the correct enum, and is a valid value within it. Any valid enum value is accepted (including zero) — see the sentinel convention in the doc comment on [SelectTeam] for detecting "unset" players.

**`Apply`:** Sets `player.(HasPlayerTeam).GetPlayerTeam().Team.SetValue(s.SelectedTeam.Value())`.

**`ValidConfiguration`:** Checks that the example player state implements `HasPlayerTeam` and that the `"team"` enum exists in the chest.

**`FallbackName`:** `"Select Team"`. **`FallbackHelpText`:** `"Choose which team to join."`

No uniqueness enforcement. Validation of team balance is the delegate's job via `ReadyToStart`.

---

### 3. `moves.SelectRole`

Lives in `moves/select_role.go`. Same pattern as `SelectTeam`.

```go
//boardgame:codegen
type SelectRole struct {
    AnyPlayer
    SelectedRole enum.Val `enum:"role"`
}
```

**`Legal`:** Validates `SelectedRole` is valid in the `"role"` enum. If configured with `WithUnique()`, also checks that no other seated player already has this role value.

**`Apply`:** Sets `player.(HasPlayerRole).GetPlayerRole().Role.SetValue(...)`.

**`ValidConfiguration`:** Checks for `HasPlayerRole` interface and `"role"` enum.

**`WithUnique()` option:** A `CustomConfigurationOption` that makes `Legal` reject values already claimed by another seated player. For the Spirit Island pattern (globally unique spirits). Captain Sonar (unique per team but shared across teams) would NOT use this — it would use `ReadyToStart` for per-team uniqueness validation.

---

### 4. `moves.SelectColor`

Lives in `moves/select_color.go`. Same pattern.

```go
//boardgame:codegen
type SelectColor struct {
    AnyPlayer
    SelectedColor enum.Val `enum:"color"`
}
```

**`Legal`:** Validates value. **Enforces uniqueness by default** (no two seated players may share a color). This is the safe default — in every real game with player colors, colors are unique. Escape hatch: `WithAllowDuplicates()`.

**`Apply`:** Sets `player.(HasPlayerColor).GetPlayerColor().Color.SetValue(...)`.

---

### 5. Behavior Upgrades

**`behaviors.PlayerRole`** needs interface additions to match the `PlayerTeam` pattern:

```go
// In behaviors/role.go:
type HasPlayerRole interface {
    GetPlayerRole() *PlayerRole
}

func (p *PlayerRole) GetPlayerRole() *PlayerRole {
    return p
}
```

**`behaviors.PlayerColor`** needs the same:

```go
// In behaviors/color.go:
type HasPlayerColor interface {
    GetPlayerColor() *PlayerColor
}

func (p *PlayerColor) GetPlayerColor() *PlayerColor {
    return p
}
```

`PlayerTeam` already has `HasPlayerTeam` and `GetPlayerTeam()`.

---

### 6. `GatheringMoves(auto)` Helper

Lives in `moves/gathering.go`. Auto-detects which selection behaviors exist on the player state and returns the corresponding move configs.

```go
func GatheringMoves(auto *AutoConfigurer) []boardgame.MoveConfig {
    exampleState := auto.Delegate().Manager().ExampleState()
    playerState := exampleState.ImmutablePlayerStates()[0]

    var result []boardgame.MoveConfig

    if _, ok := playerState.(HasPlayerTeam); ok {
        result = append(result, auto.MustConfig(new(SelectTeam)))
    }
    if _, ok := playerState.(HasPlayerRole); ok {
        result = append(result, auto.MustConfig(new(SelectRole)))
    }
    if _, ok := playerState.(HasPlayerColor); ok {
        result = append(result, auto.MustConfig(new(SelectColor)))
    }

    return result
}
```

**Usage:** `moves.AddForPhase(phaseGathering, moves.GatheringMoves(auto)...)` — registers the detected moves as legal in any order during the gathering phase.

The function returns `nil` (empty slice) if no selection behaviors are detected. `AddForPhase` with an empty slice is a no-op. So calling `GatheringMoves(auto)` is always safe.

---

### 7. `ReadyToStart(state) error` Delegate Method

Added to `GameDelegate` interface in `game_delegate.go`:

```go
// ReadyToStart is called to check whether the game's configuration is
// ready to proceed past the gathering phase. Return nil if ready, or a
// descriptive error explaining what's still needed (e.g., "each team
// needs at least 2 players"). The default returns nil.
//
// This method must be cheap (O(n) in player count or better) because it
// is called on every fix-up check cycle.
ReadyToStart(state ImmutableState) error
```

Default in `base/game_delegate.go`:

```go
func (g *GameDelegate) ReadyToStart(state boardgame.ImmutableState) error {
    return nil
}
```

**Integration points — TWO places, not one:**

`ReadyToStart` is checked in both `WaitForEnoughPlayers` and `CloseAllSeats` to prevent a critical deadlock.

**In `WaitForEnoughPlayers.Legal()`** — after the player-count check:

```go
if err := state.Manager().Delegate().ReadyToStart(state); err != nil {
    return errors.New("not ready to start: " + err.Error())
}
```

**In `CloseAllSeats.Legal()`** — when configured via `WithManualStart()`:

```go
if err := state.Manager().Delegate().ReadyToStart(state); err != nil {
    return errors.New("not ready to start: " + err.Error())
}
```

**Why both?** This prevents a deadlock: without the `CloseAllSeats` check, the admin could click "Start Game" → `CloseAllSeats` closes all empty seats → `WaitForEnoughPlayers` calls `ReadyToStart` which fails → game is stuck (seats are closed, no new players can join, and there's no `SetSeatOpen()` to undo it). By gating `CloseAllSeats` on `ReadyToStart`, the "Start Game" button stays disabled with a visible error until configuration is valid. Seats never close prematurely.

Since `CloseAllSeats` is a player move (extends `Default`, not `FixUp`), its `LegalForPlayerError` is included in the `moveForms` sent to the client. The gathering-start component can display this error next to the disabled "Start Game" button. This solves the error visibility problem — FixUp errors (from `WaitForEnoughPlayers`) are invisible to the client, but player move errors (from `CloseAllSeats`) are visible.

**Error delivery to the client — framework computed properties:**

In addition to the move-form error on `CloseAllSeats`, the `ReadyToStart` error is also surfaced via the framework-owned global computed values for display in the gathering status area:

```go
// In base.GameDelegate.FrameworkComputedGlobalProperties():
if err := g.Manager().Delegate().ReadyToStart(state); err != nil {
    result["ReadyToStartError"] = err.Error()
}
```

This ensures the error is visible even in auto-start games (without `WithManualStart`), where there is no `CloseAllSeats` move to carry the error. The gathering-status component displays `ReadyToStartError` as a status message below the player count.

**Why this works:** `WaitForEnoughPlayers` is a FixUp in `DefaultRoundSetup`'s ordered progression. It blocks the progression from advancing. Meanwhile, `SelectTeam`/`SelectRole`/`SelectColor` are registered via `AddForPhase` (unordered, any-time), so they remain legal while the progression is blocked. Players can keep changing their selections until the validation passes.

**Adding `SetSeatOpen()` to `behaviors.Seat`:**

As a safety valve, add a method to reopen a closed seat:

```go
// In behaviors/seat.go:
func (s *Seat) SetSeatOpen() {
    s.SeatClosed = false
}
```

This is not needed in normal flow (the `CloseAllSeats` gate prevents premature closing), but provides an escape hatch for game-specific moves that might need to reopen seats.

---

### 8. Framework Computed-Property Extensions

In `base/game_delegate.go`, extend the defaults to auto-detect gathering behaviors:

**FrameworkComputedPlayerProperties** — per-player data for the client:

```go
func (g *GameDelegate) FrameworkComputedPlayerProperties(player boardgame.ImmutableSubState) boardgame.PropertyCollection {
    result := boardgame.PropertyCollection{
        "Color":       behaviors.CSSColorForPlayer(player),
        "MayBeActive": g.Manager().Delegate().PlayerMayBeActive(player),
    }
    if score, ok := behaviors.PlayerGameScore(player); ok {
        result["GameScore"] = score
    }

    // Gathering: current selections
    if th, ok := player.(behaviors.HasPlayerTeam); ok {
        result["TeamValue"] = th.GetPlayerTeam().Team.String()
    }
    if rh, ok := player.(behaviors.HasPlayerRole); ok {
        result["RoleValue"] = rh.GetPlayerRole().Role.String()
    }
    if ch, ok := player.(behaviors.HasPlayerColor); ok {
        result["ColorValue"] = ch.GetPlayerColor().Color.String()
    }

    return result
}
```

**FrameworkComputedGlobalProperties** — available values for pickers + readiness error:

```go
func (g *GameDelegate) FrameworkComputedGlobalProperties(state boardgame.ImmutableState) boardgame.PropertyCollection {
    result := boardgame.PropertyCollection{}

    // Existing: player order
    // ... (existing code) ...

    // Gathering: available enum values for pickers
    chest := g.Manager().Chest()
    if teamEnum := chest.Enums().Enum("team"); teamEnum != nil {
        result["AvailableTeams"] = enumValues(teamEnum)
    }
    if roleEnum := chest.Enums().Enum("role"); roleEnum != nil {
        result["AvailableRoles"] = enumValues(roleEnum)
    }
    if colorEnum := chest.Enums().Enum("color"); colorEnum != nil {
        result["AvailableColors"] = enumValues(colorEnum)
    }

    // Gathering: readiness error for client display
    if err := g.Manager().Delegate().ReadyToStart(state); err != nil {
        result["ReadyToStartError"] = err.Error()
    }

    return result
}

// enumValues returns a list of {Key, Name} for all values in the enum.
func enumValues(e enum.Enum) []map[string]interface{} {
    var result []map[string]interface{}
    for _, key := range e.Keys() {
        result = append(result, map[string]interface{}{
            "Key":  key,
            "Name": e.String(key),
        })
    }
    return result
}
```

The client uses `AvailableTeams`/`AvailableRoles`/`AvailableColors` to populate picker dropdowns, and the per-player `TeamValue`/`RoleValue`/`ColorValue` to show current selections.

---

### 9. Server: Reopen Games When Seats Reopen

In `server/api/main.go`, when a game's state changes and unfilled/unclosed seats exist again (e.g., between rounds when `ActivateEmptySeat` fires), set `eGame.Open = true`.

The simplest hook: in the version-notification handler (when the server detects a new game version), check if the game has open seats and is currently closed. If so, reopen it.

```go
func (s *Server) maybeReopenGame(game *boardgame.Game) {
    closedSeats := s.closedSeatsForGame(game)
    userIds := s.storage.UserIDsForGame(game.ID())
    agents := game.Agents()

    for i, uid := range userIds {
        if uid == "" && agents[i] == "" && !closedSeats[i] {
            // There's an open, unfilled, non-agent slot.
            eGame, err := s.storage.ExtendedGame(game.ID())
            if err == nil && !eGame.Open {
                eGame.Open = true
                s.storage.UpdateExtendedGame(game.ID(), eGame)
            }
            return
        }
    }
}
```

---

## Client Changes

### 10. `boardgame-gathering-panel`

New component: `server/static/src/components/boardgame-gathering-panel.ts`

**Position:** Rendered by `boardgame-game-view` between `boardgame-player-roster` and `boardgame-render-game`. Always in the DOM, but auto-hides when it has nothing to show.

**Detection logic:** Scans `moveForms` for gathering-related move names. Each sub-component independently decides whether to render based on its corresponding move's legality.

**Properties received from `boardgame-game-view`:**
- `moveForms` — the full move forms array (already computed)
- `state` — current expanded game state (has Computed.Global and Computed.Players)
- `viewingAsPlayer` — current player index
- `hasEmptySlots`, `gameOpen`, `finished` — existing game state flags
- `isOwner`, `loggedIn` — existing auth state
- `gameRoute` — for constructing the share URL

**Sub-components:**

| Component | Visible when | What it shows |
|-----------|-------------|--------------|
| `boardgame-gathering-status` | `hasEmptySlots && gameOpen && !finished`, OR `Computed.Global.ReadyToStartError` is non-empty | "Waiting for 2 more players" / `ReadyToStartError` text / "Ready to start" |
| `boardgame-gathering-share` | `hasEmptySlots && gameOpen && !isObserver` | Copy invite link button |
| `boardgame-gathering-team-picker` | A move with a `SelectedTeam` field (`EnumName: "team"`) exists and `LegalForAnyone` | Dropdown per seated player; interactive for self (if `LegalForPlayer`), read-only for others and observers |
| `boardgame-gathering-color-picker` | A move with a `SelectedColor` field (`EnumName: "color"`) exists and `LegalForAnyone` | Color swatch per seated player; interactive for self, read-only for others |
| `boardgame-gathering-role-picker` | A move with a `SelectedRole` field (`EnumName: "role"`) exists and `LegalForAnyone` | Dropdown per seated player; interactive for self, read-only for others |
| `boardgame-gathering-start` | A move with no non-TargetPlayerIndex fields exists in known start-move names, OR fallback: a `CloseAllSeats`-type move is `LegalForAnyone` | "Start Game" button; enabled when `LegalForPlayer`; shows `LegalForPlayerError` when disabled |

**Move detection — field signatures, not name strings:**

Each sub-component detects its corresponding move by inspecting the `Fields` array on each move form, not by matching the move name. This is robust against game authors subclassing moves or using `WithMoveName()`.

- **Team picker:** Looks for a move form that has a field with `EnumName == "team"` and type `TypeEnum`.
- **Color picker:** Looks for `EnumName == "color"`.
- **Role picker:** Looks for `EnumName == "role"`.
- **Start button:** Looks for a move form where `LegalForAnyone` is true and the move name matches known start names (`"Confirm Players"`, `"Close All Seats"`, `"Start Game"`). As a fallback, any non-FixUp move with zero fields (other than `TargetPlayerIndex`) could be treated as a start candidate.

Field signature detection means: if a game author creates `type MovePickYourSpirit struct { moves.SelectRole }` and it gets named `"Pick Your Spirit"` by `DeriveName`, the gathering panel still detects it as a role picker because it has a `SelectedRole` field with `EnumName: "role"`.

**Observer/spectator behavior:** Sub-components use two signals:
- `LegalForAnyone` controls **visibility** — should this widget render at all?
- `LegalForPlayer` controls **interactivity** — should the widget be interactive (dropdown) or read-only (label)?

Observers have `LegalForPlayer = false` for all moves (the server skips legality computation for `ObserverPlayerIndex`). So observers see read-only displays of current team/color/role assignments but cannot interact.

**Auto-hide:** The panel checks if any sub-component has content to render. If none do, it sets `display: none`. This means:
- Before anyone joins: status + share link visible
- During gathering with team picking: status + share + team picker + start button visible
- After game starts: all moves become illegal → all sub-components hide → panel hides
- Between rounds (if phase cycles back): moves become legal again → panel reappears

---

### 11. How Pickers Work

**Team picker example:**

1. Component reads `Computed.Global.AvailableTeams` → `[{Key: 0, Name: "Red"}, {Key: 1, Name: "Blue"}]`
2. For each seated player, reads `Computed.Players[i].TeamValue` → current selection (e.g., `"Red"` or `""`)
3. For the viewing player, renders a `<md-filled-select>` dropdown with the available teams
4. For other players, renders a read-only label showing their current selection
5. When the user selects a value, dispatches:
   ```js
   this.dispatchEvent(new CustomEvent('propose-move', {
       composed: true, bubbles: true,
       detail: {
           name: 'Select Team',
           arguments: {
               TargetPlayerIndex: String(this.viewingAsPlayer),
               SelectedTeam: selectedTeamName  // e.g., "Red"
           }
       }
   }));
   ```
6. The existing move proposal pipeline handles the rest.

**Color picker** is identical but renders color swatches instead of a dropdown. Already-claimed colors are shown with a player avatar or dimmed (since `SelectColor` enforces uniqueness, proposing a taken color would fail with a legality error).

---

### 12. Override System

**Level 0 (zero work):** Game author does nothing. Gathering panel auto-detects from moves. Every game with `DefaultRoundSetup` gets status + share link + start button.

**Level 1 (CSS hide):** Each sub-component respects a CSS custom property:
```css
boardgame-gathering-team-picker {
    display: var(--boardgame-gathering-team-picker-display, block);
}
```
Game author hides the framework's picker: `--boardgame-gathering-team-picker-display: none;` and renders their own in the game renderer.

**Level 2 (full replacement — future):** Register `boardgame-render-gathering-GAMENAME` as a custom element. Not yet implemented; planned for a future release. For now, use the CSS override to hide framework pickers and render custom UI in your game renderer via `gatheringActive`:
```typescript
// In your game renderer:
if (this.gatheringActive) {
    // Render custom gathering UI
}
```
Same convention as `boardgame-render-game-GAMENAME`.

**Game renderer signal:** `BoardgameBaseGameRenderer` gains a computed `gatheringActive: boolean` property, derived from whether any gathering-related moves are legal. Game renderers can use `if (this.gatheringActive) { ... }` to conditionally render gathering-specific UI (e.g., Codenames' drag-and-drop team picker).

---

### 13. `boardgame-player-roster` Enhancements

**Banner text change:**
```typescript
private _bannerText(): string {
    if (this.finished) return "Game Over";
    if (this.readyToStartError) return "Setting Up";
    if (this.hasEmptySlots && this.gameOpen) return "Waiting for Players";
    return "Playing";
}
```

The `readyToStartError` comes from `Computed.Global.ReadyToStartError`. When present, the banner says "Setting Up" rather than "Playing" — this handles the case where all seats are filled but configuration isn't valid yet (e.g., not everyone has picked a team). The detailed error message is shown in the gathering-status sub-component, not the banner.

**Per-player gathering metadata:** Each `boardgame-player-roster-item` already receives `state` and `playerIndex`. It can read `Computed.Players[i].TeamValue`, `Computed.Players[i].ColorValue`, etc. to show team badges or color indicators next to player names. This is a small enhancement to the existing roster item rendering.

---

## Conventions

### Phase naming
By convention, the first phase in the enum (iota 0) is the gathering phase. Since `PhaseBehavior.Phase` defaults to 0, the game starts in this phase with no code needed.

### Move organization
Gathering selection moves go in `AddForPhase` (legal any order, any number of times). The `DefaultRoundSetup` progression goes in `AddOrderedForPhase` for the same phase. Both coexist: the progression gates phase advancement while the selection moves remain freely available.

### Multi-round games
To reopen gathering between rounds, the round cleanup phase transitions back to the gathering phase. `DefaultRoundSetup` runs again (activating new players, waiting for enough, etc.). The gathering UI naturally reappears because the gathering moves become legal again.

---

## Examples

### Simple card game (zero gathering code)

```go
const (
    phaseGathering = iota
    phasePlaying
)

type playerState struct {
    base.SubState
    behaviors.Seat
    behaviors.InactivePlayer
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    auto := moves.NewAutoConfigurer(g)
    return moves.Combine(
        moves.Add(
            auto.MustConfig(new(moves.SeatPlayer)),
        ),
        moves.AddOrderedForPhase(phaseGathering,
            moves.DefaultRoundSetup(auto, moves.WithManualStart()),
            auto.MustConfig(new(moves.StartPhase),
                moves.WithPhaseToStart(phasePlaying, phaseEnum)),
        ),
        moves.AddForPhase(phasePlaying,
            // ... game-specific moves ...
        ),
    )
}
```

**Result:** Gathering panel shows "Waiting for Players," share link, and "Start Game" button. Zero new code beyond what games already write today.

### Codenames (teams + roles + validation)

```go
const (
    phaseGathering = iota
    phaseClueGiving
    phaseGuessing
)

type playerState struct {
    base.SubState
    behaviors.Seat
    behaviors.InactivePlayer
    behaviors.PlayerTeam   // +1 line
    behaviors.PlayerRole   // +1 line
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    auto := moves.NewAutoConfigurer(g)
    return moves.Combine(
        moves.Add(
            auto.MustConfig(new(moves.SeatPlayer)),
        ),
        moves.AddForPhase(phaseGathering,
            moves.GatheringMoves(auto)...,           // +1 line: auto-detects Team + Role
        ),
        moves.AddOrderedForPhase(phaseGathering,
            moves.DefaultRoundSetup(auto, moves.WithManualStart()),
            auto.MustConfig(new(moves.StartPhase),
                moves.WithPhaseToStart(phaseClueGiving, phaseEnum)),
        ),
        // ... game phases ...
    )
}

func (g *gameDelegate) ReadyToStart(state boardgame.ImmutableState) error {
    // Game-specific validation
    _, players := concreteStates(state)
    redSpymasters, blueSpymasters := 0, 0
    for _, p := range players {
        if !p.SeatIsFilled() { continue }
        if p.Team.Value() == teamUnset {
            return errors.New("all players must select a team")
        }
        if p.Role.Value() == roleUnset {
            return errors.New("all players must select a role")
        }
        if p.Role.Value() == roleSpymaster {
            if p.Team.Value() == teamRed { redSpymasters++ }
            if p.Team.Value() == teamBlue { blueSpymasters++ }
        }
    }
    if redSpymasters != 1 { return errors.New("Red team needs exactly 1 spymaster") }
    if blueSpymasters != 1 { return errors.New("Blue team needs exactly 1 spymaster") }
    return nil
}
```

**Result:** Gathering panel auto-shows team picker + role picker + start button. "Start Game" is disabled with the `ReadyToStart` error message until teams are balanced. Delta from simple game: +2 behavior embeds, +1 `GatheringMoves` line, +1 `ReadyToStart` override.

### Spirit Island (unique roles, cooperative)

```go
type playerState struct {
    base.SubState
    behaviors.Seat
    behaviors.InactivePlayer
    behaviors.PlayerRole   // spirit selection
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    auto := moves.NewAutoConfigurer(g)
    return moves.Combine(
        moves.Add(auto.MustConfig(new(moves.SeatPlayer))),
        moves.AddForPhase(phaseGathering,
            auto.MustConfig(new(moves.SelectRole),
                moves.WithUnique(),  // no two players can pick the same spirit
            ),
        ),
        moves.AddOrderedForPhase(phaseGathering,
            moves.DefaultRoundSetup(auto, moves.WithManualStart()),
            auto.MustConfig(new(moves.StartPhase),
                moves.WithPhaseToStart(phasePlaying, phaseEnum)),
        ),
        // ...
    )
}
```

**Note:** Here the game uses `SelectRole` directly with `WithUnique()` instead of `GatheringMoves(auto)`, because it wants the uniqueness option. `GatheringMoves` is a convenience that returns default-configured moves — games can always configure moves manually for more control.

### Poker cash game (recurring gathering, drop-in/drop-out)

```go
const (
    phaseGathering = iota
    phaseDealing
    phaseBetting
    phaseShowdown
)

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    auto := moves.NewAutoConfigurer(g)
    return moves.Combine(
        moves.Add(auto.MustConfig(new(moves.SeatPlayer))),
        // Gathering phase: wait for players, start
        moves.AddOrderedForPhase(phaseGathering,
            moves.DefaultRoundSetup(auto),  // auto-start, no manual start needed
            auto.MustConfig(new(moves.StartPhase),
                moves.WithPhaseToStart(phaseDealing, phaseEnum)),
        ),
        // ... dealing, betting, showdown phases ...
        // Showdown transitions back to gathering:
        moves.AddOrderedForPhase(phaseShowdown,
            // ... resolve hand ...
            auto.MustConfig(new(moves.StartPhase),
                moves.WithPhaseToStart(phaseGathering, phaseEnum)),  // loop back
        ),
    )
}
```

**Result:** Between every hand, the game returns to `phaseGathering`. `DefaultRoundSetup` runs: `ActivateInactivePlayer` activates new joiners, `WaitForEnoughPlayers` ensures enough are seated. The gathering panel briefly appears ("Waiting for Players" if a seat opened), then disappears when the next hand starts. The server reopens `eGame.Open` when seats reopen.

### Secret role game (Resistance — game-specific assignment)

```go
// No SelectTeam/SelectRole — roles are auto-assigned
func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    auto := moves.NewAutoConfigurer(g)
    return moves.Combine(
        moves.Add(auto.MustConfig(new(moves.SeatPlayer))),
        moves.AddOrderedForPhase(phaseGathering,
            moves.DefaultRoundSetup(auto),  // auto-start when all seats filled
            auto.MustConfig(new(moveAssignSecretRoles)),  // game-specific FixUp
            auto.MustConfig(new(moves.StartPhase),
                moves.WithPhaseToStart(phaseMission, phaseEnum)),
        ),
    )
}
```

**Result:** Gathering panel shows only "Waiting for Players" status (no team/role pickers since no selection moves are registered). When all seats fill, `WaitForEnoughPlayers` passes, `moveAssignSecretRoles` fires as a FixUp (randomly assigns roles from a game-specific algorithm), then the game transitions to play.

---

## What's NOT in This Design

- **No `SetUpMode` / `SetUpConfig` / `FinalizeSetUp`** — no new framework concepts
- **No `ApplyVariantValue` / `SetVariantValue`** — variants stay fixed at creation time
- **No `DefaultLobbySetup`** — `DefaultRoundSetup` IS the gathering
- **No new server endpoints** — existing data is sufficient
- **No core engine changes** — `game.go`, `game_manager.go` unchanged
- **No mutable variants** — deferred; can be added later without breaking this design
- **No tournament/campaign orchestration** — explicitly out of scope (server-layer concern)
- **No admin-assigns-to-other-player** — game-specific moves for that pattern

## Implementation Order

### Phase 1: Client-only lobby UX (zero Go changes)
1. `boardgame-gathering-panel` with status, share link, start button sub-components
2. Wire into `boardgame-game-view`
3. Update `boardgame-player-roster` banner text

### Phase 2: Go-side selection moves
1. Upgrade `PlayerRole` and `PlayerColor` behaviors (add interfaces)
2. Implement `moves.AnyPlayer` base type
3. Implement `moves.SelectTeam`, `moves.SelectRole`, `moves.SelectColor`
4. Implement `GatheringMoves(auto)` helper
5. Add `ReadyToStart` to `GameDelegate` and integrate into `WaitForEnoughPlayers`
6. Extend `ComputedProperties` with team/role/color data
7. Run codegen

### Phase 3: Client-side pickers
1. Team picker, role picker, color picker sub-components
2. Override system (CSS custom properties + custom element detection)
3. `gatheringActive` signal on base game renderer

### Phase 4: Server fix + polish
1. Reopen `eGame.Open` when seats reopen between rounds
2. Debug auto-seating (#774)
3. Convert one example game to demonstrate the pattern

### Phase 5: Test against PRD scenarios
Verify coverage of all 26 scenarios in `GATHERING_PRD.md`.
