# Idiomatic Generics API Design for Boardgame Framework

**Status**: Design Proposal
**Date**: 2026-02-07
**Author**: Claude (final-designer agent)

## Executive Summary

This document proposes a complete, idiomatic API design for adding generics to the boardgame framework. The design prioritizes **developer ergonomics** by embracing **code generation** as the primary solution to minimize boilerplate, rather than fighting Go's limitations with type aliases.

**Core Philosophy**: Users should work primarily with concrete types and only see generic syntax during initial setup and in generated code. The framework handles the complexity.

## Design Decisions

### 1. Naming Convention: `StateReader` / `State`

Following Go stdlib patterns (io.Reader/Writer, http.Handler), we use:

- **`StateReader[G, P]`** - Immutable state interface (replaces `ImmutableState`)
- **`State[G, P]`** - Mutable state interface (replaces `State`)
- **`SubStateReader[T]`** - Immutable substate (replaces `ImmutableSubState`)
- **`SubState[T]`** - Mutable substate (replaces `SubState`)

**Rationale**:
- Consistent with Go naming (Reader suffix = read-only)
- Natural progression: StateReader → State (not StateReader → StateWriter)
- Short enough to not be burdensome
- Clearly distinguishes immutable from mutable

Type parameters:
- `G` = Game state type
- `P` = Player state type
- `T` = Generic substate type

### 2. Solution to Verbosity: Code Generation, Not Aliases

**Critical Finding**: Type aliases cannot be used in struct embedding (from task #2 research).

```go
// This DOESN'T WORK:
type State = boardgame.State[gameState, playerState]
type moveRollDice struct {
    moves.CurrentPlayer[State] // ❌ Cannot embed with alias
}

// Must write:
type moveRollDice struct {
    moves.CurrentPlayer[gameState, playerState] // ✅ But very verbose
}
```

**Solution**: Generate type-specific wrappers via `boardgame-util codegen`.

## Complete API Design

### Core Type Hierarchy

```go
// state.go - Core state interfaces

// StateReader is the read-only view of game state
type StateReader[G any, P any] interface {
    // Game state access
    GameState() SubStateReader[G]
    PlayerStates() []SubStateReader[P]
    CurrentPlayer() SubStateReader[P]

    // Navigation
    CurrentPlayerIndex() PlayerIndex
    Version() int

    // Utilities
    Copy(sanitized bool) (StateReader[G, P], error)
    Diagram() string
    Sanitized() bool
    SanitizedForPlayer(player PlayerIndex) (StateReader[G, P], error)
    Game() *Game[G, P]
    Manager() *GameManager[G, P]
}

// State is the mutable view of game state
type State[G any, P any] interface {
    StateReader[G, P]

    // Mutable access
    MutableGameState() SubState[G]
    MutablePlayerStates() []SubState[P]
    MutableCurrentPlayer() SubState[P]

    // Randomness (only on mutable state)
    Rand() *rand.Rand
}

// SubStateReader is read-only substate
type SubStateReader[T any] interface {
    Reader() PropertyReader
    Value() T  // Get concrete value
}

// SubState is mutable substate
type SubState[T any] interface {
    SubStateReader[T]
    ReadSetter() PropertyReadSetter
}
```

### Move Interfaces

```go
// move.go - Move interfaces

// Move is the base move interface
type Move[G any, P any] interface {
    Legal(state StateReader[G, P], proposer PlayerIndex) error
    Apply(state State[G, P]) error

    // Metadata
    Info() *MoveInfo
    HelpText() string
    DefaultsForState(state StateReader[G, P])

    // Infrastructure
    SetInfo(m *MoveInfo)
    TopLevelStruct() Move[G, P]
    SetTopLevelStruct(m Move[G, P])
    ValidConfiguration(exampleState State[G, P]) error
    ReadSetConfigurer
}

// MoveConfig for registering moves
type MoveConfig[G any, P any] interface {
    Name() string
    Constructor() func() Move[G, P]
    CustomConfiguration() PropertyCollection
}
```

### Delegate Interface

```go
// game_delegate.go - Delegate interface

type GameDelegate[G any, P any] interface {
    // Identity
    Name() string
    DisplayName() string
    Description() string

    // Player configuration
    MinNumPlayers() int
    MaxNumPlayers() int
    DefaultNumPlayers() int

    // State constructors
    GameStateConstructor() G
    PlayerStateConstructor(player PlayerIndex) P

    // Configuration
    ConfigureMoves() []MoveConfig[G, P]
    ConfigureDecks() map[string]*Deck
    ConfigureEnums() *enum.Set
    ConfigureConstants() PropertyCollection

    // Lifecycle
    BeginSetUp(state State[G, P], variant Variant) error
    FinishSetUp(state State[G, P]) error
    DistributeComponentToStarterStack(state StateReader[G, P], c Component) (ImmutableStack, error)

    // Game logic
    CurrentPlayerIndex(state StateReader[G, P]) PlayerIndex
    GameEndConditionMet(state StateReader[G, P]) bool
    CheckGameOver(state StateReader[G, P]) (bool, []PlayerIndex)
    PlayerScore(pState SubStateReader[P]) int

    // Visualization
    Diagram(state StateReader[G, P]) string

    // Fix-up moves
    ProposeFixUpMove(state StateReader[G, P]) Move[G, P]

    // Computed properties
    ComputedGlobalProperties(state StateReader[G, P]) PropertyCollection
    ComputedPlayerProperties(player SubStateReader[P]) PropertyCollection
}
```

## Code Generation Strategy

The key insight: **generate concrete wrappers** so users rarely write generic syntax.

### What Gets Generated

For each game package, `boardgame-util codegen` generates:

```go
// auto_generic.go - Generated by boardgame-util codegen

package pig

import "github.com/jkomoros/boardgame"

// ============================================
// Concrete Type Aliases (for documentation)
// ============================================

// StateReader is the read-only state for pig
type StateReader = boardgame.StateReader[gameState, playerState]

// State is the mutable state for pig
type State = boardgame.State[gameState, playerState]

// Move is a move in pig
type Move = boardgame.Move[gameState, playerState]

// ============================================
// Concrete Wrapper for moves Package
// ============================================

// CurrentPlayer wraps moves.CurrentPlayer with pig's types
type CurrentPlayer struct {
    moves.CurrentPlayer[gameState, playerState]
}

// FinishTurn wraps moves.FinishTurn with pig's types
type FinishTurn struct {
    moves.FinishTurn[gameState, playerState]
}

// DealComponents wraps moves.DealComponents with pig's types
type DealComponents struct {
    moves.DealComponents[gameState, playerState]
}

// ... etc for all commonly used moves

// ============================================
// Helper Functions
// ============================================

// ConcreteStates extracts concrete game and player states
func ConcreteStates(state boardgame.StateReader[gameState, playerState]) (*gameState, []*playerState) {
    game := state.GameState().Value()

    playerReaders := state.PlayerStates()
    players := make([]*playerState, len(playerReaders))
    for i, p := range playerReaders {
        players[i] = p.Value()
    }

    return game, players
}
```

### What Users Write

Users write clean, concise code with no generic syntax:

```go
// state.go - User-written code

package pig

import (
    "github.com/jkomoros/boardgame"
    "github.com/jkomoros/boardgame/base"
    "github.com/jkomoros/boardgame/behaviors"
)

//go:generate boardgame-util codegen

// gameState is the main game state
type gameState struct {
    base.SubState
    behaviors.CurrentPlayerBehavior
    Die         boardgame.SizedStack `sizedstack:"dice"`
    TargetScore int
}

// playerState is per-player state
type playerState struct {
    base.SubState
    Busted     bool
    Done       bool
    DieCounted bool
    RoundScore int
    TotalScore int
}

// Custom methods on states
func (p *playerState) TurnDone() error {
    if !p.DieCounted {
        return errors.New("die not counted")
    }
    if !p.Busted && !p.Done {
        return errors.New("not finished")
    }
    return nil
}
```

```go
// moves.go - User-written moves

package pig

import "errors"

// moveRollDice rolls the die
type moveRollDice struct {
    CurrentPlayer  // Uses generated wrapper, no generics!
}

func (m *moveRollDice) Legal(state StateReader, proposer boardgame.PlayerIndex) error {
    if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
        return err
    }

    game, players := ConcreteStates(state)
    p := players[game.CurrentPlayer.EnsureValid(state)]

    if !p.DieCounted {
        return errors.New("die not counted yet")
    }

    return nil
}

func (m *moveRollDice) Apply(state State) error {
    game, players := ConcreteStates(state)
    p := players[game.CurrentPlayer.EnsureValid(state)]

    die := game.Die.ComponentAt(0)
    die.DynamicValues().(*dice.DynamicValue).Roll(state.Rand())
    p.DieCounted = false

    return nil
}

// moveDoneTurn signals end of turn
type moveDoneTurn struct {
    CurrentPlayer  // Clean!
}

func (m *moveDoneTurn) Legal(state StateReader, proposer boardgame.PlayerIndex) error {
    if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
        return err
    }

    game, players := ConcreteStates(state)
    p := players[game.CurrentPlayer.EnsureValid(state)]

    if !p.DieCounted {
        return errors.New("die not counted")
    }

    return nil
}

func (m *moveDoneTurn) Apply(state State) error {
    game, players := ConcreteStates(state)
    p := players[game.CurrentPlayer.EnsureValid(state)]

    p.Done = true
    return nil
}
```

```go
// main.go - Delegate implementation

package pig

import (
    "github.com/jkomoros/boardgame"
    "github.com/jkomoros/boardgame/base"
    "github.com/jkomoros/boardgame/moves"
)

type gameDelegate struct {
    base.GameDelegate[gameState, playerState]
}

func (g *gameDelegate) Name() string {
    return "pig"
}

func (g *gameDelegate) GameStateConstructor() gameState {
    return gameState{
        TargetScore: 100,
    }
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) playerState {
    return playerState{}
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig[gameState, playerState] {
    auto := moves.NewAutoConfigurer[gameState, playerState](g)

    return moves.Add(
        auto.MustConfig(
            new(moveRollDice),
            moves.WithHelpText("Roll the die"),
        ),
        auto.MustConfig(
            new(moveDoneTurn),
            moves.WithHelpText("End turn"),
        ),
        auto.MustConfig(
            new(FinishTurn),  // Generated wrapper
            moves.WithHelpText("Advance to next player"),
        ),
    )
}

func (g *gameDelegate) GameEndConditionMet(state StateReader) bool {
    game, players := ConcreteStates(state)

    for _, player := range players {
        if player.TotalScore >= game.TargetScore {
            return true
        }
    }

    return false
}

func NewDelegate() boardgame.GameDelegate[gameState, playerState] {
    return &gameDelegate{}
}
```

## Key Benefits

### 1. Users Rarely See Generics

- State definitions: **zero generic syntax**
- Move definitions: **zero generic syntax** (uses generated wrappers)
- Move methods: type parameters hidden in generated aliases
- Delegate methods: mostly using generated aliases

Generic syntax appears only:
- In the delegate struct definition (once)
- In `ConfigureMoves()` return type (once)
- In `NewDelegate()` return type (once)

### 2. No More Type Casting

**Before (current API)**:
```go
func (m *moveRollDice) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    game := state.ImmutableGameState().(*gameState)  // Cast required
    players := state.ImmutablePlayerStates()
    p := players[game.CurrentPlayer].(*playerState)  // Cast required

    if !p.DieCounted {
        return errors.New("die not counted")
    }
    return nil
}
```

**After (generic API)**:
```go
func (m *moveRollDice) Legal(state StateReader, proposer boardgame.PlayerIndex) error {
    game, players := ConcreteStates(state)  // No casts!
    p := players[game.CurrentPlayer]

    if !p.DieCounted {
        return errors.New("die not counted")
    }
    return nil
}
```

### 3. Type Safety

- Compiler catches type mismatches
- Cannot accidentally mix states from different games
- Autocomplete works perfectly
- Refactoring is safe

### 4. Incremental Migration

Games can migrate one method at a time:

```go
// Phase 1: Add generic type parameters to delegate
type gameDelegate struct {
    base.GameDelegate[gameState, playerState]
}

// Phase 2: Update method signatures one by one
func (g *gameDelegate) GameEndConditionMet(state StateReader) bool {
    // Use new API
}

// Old methods continue to work during migration
```

## Code Generation Details

### Codegen Directive

```go
//go:generate boardgame-util codegen
```

### What Codegen Analyzes

1. **Finds state types**: Scans for types with `base.SubState` embedded
2. **Identifies game vs player**: Uses heuristics (name, position in delegate)
3. **Generates wrappers**: For all `moves.*` types used
4. **Creates helpers**: Like `ConcreteStates()`

### Generated File Structure

```go
// auto_generic.go
package <game>

// Part 1: Type aliases for documentation
type StateReader = boardgame.StateReader[<game>, <player>]
type State = boardgame.State[<game>, <player>]
type Move = boardgame.Move[<game>, <player>]

// Part 2: Move wrappers (one per moves.* type used)
type CurrentPlayer struct {
    moves.CurrentPlayer[<game>, <player>]
}

// Part 3: Helper functions
func ConcreteStates(state boardgame.StateReader[<game>, <player>]) (*<game>, []*<player>) {
    // ...
}
```

## Migration Path

### Phase 1: Update Framework (Parallel to v1)

1. Add generic interfaces alongside existing ones
2. Keep old interfaces working
3. No breaking changes

```go
// Both exist:
type ImmutableState interface { ... }  // Old
type StateReader[G, P any] interface { ... }  // New
```

### Phase 2: Codegen Support

1. `boardgame-util codegen` learns to generate generic wrappers
2. Games opt-in with `//go:generate boardgame-util codegen`
3. Generated code references generic APIs

### Phase 3: Game Migration

Games migrate incrementally:

```go
// Step 1: Generate wrappers
//go:generate boardgame-util codegen

// Step 2: Update delegate type parameters
type gameDelegate struct {
    base.GameDelegate[gameState, playerState]
}

// Step 3: Update methods one by one
func (g *gameDelegate) GameEndConditionMet(state StateReader) bool {
    // Old casts still work during migration:
    oldStyle := boardgame.ImmutableState(state)

    // Or use new style:
    game, players := ConcreteStates(state)
}
```

### Phase 4: Deprecation (Far Future)

1. Mark old interfaces as deprecated
2. Provide automated migration tool
3. Eventually remove in v2.0

## Advanced: Dynamic Component Values

For games with dynamic component values, codegen generates 3-parameter versions:

```go
// If the game uses dynamic component values:
type StateReader = boardgame.StateReader[gameState, playerState, cardDynamic]
type State = boardgame.State[gameState, playerState, cardDynamic]

// ConcreteStates includes component values
func ConcreteStates(state StateReader) (*gameState, []*playerState, map[string][]*cardDynamic) {
    // ...
}
```

But most games don't use this, so the 2-parameter version is standard.

## Comparison: Before and After

### Full Move Example

**Before (Current API)**:
```go
type moveRollDice struct {
    moves.CurrentPlayer
}

func (m *moveRollDice) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
        return err
    }

    // Multiple casts required
    game := state.ImmutableGameState().(*gameState)
    players := state.ImmutablePlayerStates()
    p := players[game.CurrentPlayer].(*playerState)

    if !p.DieCounted {
        return errors.New("die not counted")
    }

    return nil
}

func (m *moveRollDice) Apply(state boardgame.State) error {
    // Cast again in Apply
    game := state.GameState().(*gameState)
    players := state.PlayerStates()
    p := players[game.CurrentPlayer].(*playerState)

    die := game.Die.ComponentAt(0)
    die.DynamicValues().(*dice.DynamicValue).Roll(state.Rand())
    p.DieCounted = false

    return nil
}
```

**After (Generic API)**:
```go
type moveRollDice struct {
    CurrentPlayer  // Generated wrapper, no generic syntax
}

func (m *moveRollDice) Legal(state StateReader, proposer boardgame.PlayerIndex) error {
    if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
        return err
    }

    // Single call, no casts
    game, players := ConcreteStates(state)
    p := players[game.CurrentPlayer]

    if !p.DieCounted {
        return errors.New("die not counted")
    }

    return nil
}

func (m *moveRollDice) Apply(state State) error {
    // Same clean pattern
    game, players := ConcreteStates(state)
    p := players[game.CurrentPlayer]

    die := game.Die.ComponentAt(0)
    die.DynamicValues().(*dice.DynamicValue).Roll(state.Rand())
    p.DieCounted = false

    return nil
}
```

**Line count**: Nearly identical
**Type casts**: 4 → 0
**Generic syntax seen**: 0
**Type safety**: Much better

## Open Questions

### 1. Behaviors Package

Current behaviors use embedding:

```go
type gameState struct {
    base.SubState
    behaviors.CurrentPlayerBehavior  // How to parameterize?
}
```

**Options**:
- **A**: Behaviors don't use generics (they're simple)
- **B**: Behaviors are generic: `behaviors.CurrentPlayer[G, P]`
- **C**: Codegen generates behavior wrappers too

**Recommendation**: Option A initially. Behaviors are simple state holders, don't need generic methods.

### 2. Type Parameter Ordering

Should it be `[G, P]` or `[P, G]`?

**Decision**: `[G, P]` (game first, players second)
- Matches mental model (game contains players)
- Matches order in delegate methods
- Matches usual access patterns

### 3. Component Values

3-parameter version for dynamic components feels heavy:

```go
type State[G, P, C any] interface { ... }
```

**Recommendation**: Use map for component values instead:

```go
type State[G, P any] interface {
    DynamicComponentValues() map[string][]SubStateReader[any]
}
```

Users cast when needed (rare operation). Keeps 90% of games simple.

## Implementation Order

1. **Week 1-2**: Core generic interfaces
   - `StateReader[G, P]`, `State[G, P]`
   - `Move[G, P]`, `GameDelegate[G, P]`
   - Parallel to existing APIs

2. **Week 3-4**: Code generation
   - Parse state types
   - Generate wrappers
   - Generate helpers

3. **Week 5**: Base package updates
   - `base.GameDelegate[G, P]`
   - `base.Move[G, P]`
   - Test with pig example

4. **Week 6**: Moves package
   - Update all standard moves
   - Maintain backward compatibility
   - Document migration

5. **Week 7-8**: Documentation and migration guide
   - Tutorial for new games
   - Migration guide for existing games
   - Examples

## Success Metrics

A successful design means:

1. **New games**: No generic syntax in user code (except delegate definition)
2. **Existing games**: Can migrate incrementally
3. **Type safety**: Zero type assertions in move methods
4. **Code size**: Similar line count before/after
5. **Learning curve**: Minimal (codegen handles complexity)

## Conclusion

This design resolves the core tension:

- **Generics add verbosity in declarations** → Solved by code generation
- **But eliminate casting in bodies** → Users get this benefit
- **Users should rarely see generics** → Achieved (only in delegate setup)

The key insight is embracing code generation rather than fighting Go's limitations with type aliases. Generated wrappers make generic types invisible to users while preserving all type safety benefits.

**Users write clean, type-safe code. The framework handles the complexity.**
