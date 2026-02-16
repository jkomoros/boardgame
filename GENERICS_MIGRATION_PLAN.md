# Generics Migration Plan for Boardgame Framework

> **Version:** 1.0
> **Date:** 2026-02-07
> **Status:** Draft
> **Authors:** Generics Planning Team

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Prerequisites](#prerequisites)
3. [Migration Strategy Overview](#migration-strategy-overview)
4. [API Changes by Component](#api-changes-by-component)
5. [Detailed Changes by File/System](#detailed-changes-by-filesystem)
6. [Implementation Order](#implementation-order)
7. [Impact Analysis](#impact-analysis)
8. [Testing Strategy](#testing-strategy)
9. [Documentation Updates](#documentation-updates)
10. [Risk Mitigation](#risk-mitigation)
11. [Appendices](#appendices)

---

## Executive Summary

The boardgame framework (~85,000 lines of Go code) was designed for Go 1.13, before generics were available. This plan outlines a phased migration to add idiomatic generics support, eliminating:

- **125 `concreteStates()` calls** - The primary type casting pattern used in every game
- **220 type assertions** throughout the codebase
- **299 `interface{}` usages** in the framework core
- **87 reflection calls** in the property system

### Key Benefits

1. **Type Safety**: Compile-time type checking instead of runtime type assertions
2. **Developer Experience**: Eliminate boilerplate `concreteStates()` helper in every game
3. **Performance**: Reduce reflection overhead in hot paths
4. **Maintainability**: Clearer code with explicit type parameters
5. **Modern Go**: Align with Go 1.18+ best practices

### Migration Approach

**Backward Compatible Phases**: Each phase maintains compatibility with existing games while introducing new generic APIs. Games can migrate incrementally.

**Estimated Timeline**: 6-8 weeks for core framework changes, plus 2-4 weeks per example game migration.

---

## Prerequisites

### Go Version Requirements

**Minimum Version**: Go 1.18 (generics support)
**Recommended Version**: Go 1.21+ (improved type inference, better error messages)
**Current Version**: Go 1.13 (declared in go.mod)

### Migration Steps

1. **Update go.mod**
   ```
   go 1.21
   ```

2. **Dependency Updates**
   - All dependencies are indirect or compatible with Go 1.18+
   - No breaking dependency changes required
   - Test all storage backends (MySQL, Bolt, filesystem)

3. **Tooling Updates**
   - `boardgame-util codegen`: Update code generation templates for generic types
   - VSCode/GoLand: Ensure IDE supports Go 1.18+ for type parameter hints
   - CI/CD: Update build pipelines to use Go 1.21+

### Compatibility Matrix

| Component | Go 1.13 | Go 1.18+ | Notes |
|-----------|---------|----------|-------|
| Core framework | ✓ | ✓ | Current implementation |
| Generic APIs | ✗ | ✓ | New APIs require 1.18+ |
| Code generation | ✓ | ✓ | Templates updated |
| Storage backends | ✓ | ✓ | No changes needed |
| Example games | ✓ | ✓ | Can use either API |

---

## Migration Strategy Overview

### Phased Approach

**Phase 1: State System (Weeks 1-2)**
- Add generic State/ImmutableState interfaces
- Implement `State[G, P]` type with backward compatibility
- Eliminate `concreteStates()` pattern

**Phase 2: Component System (Week 3)**
- Add generic Component interfaces
- Type-safe component values access
- Update deck and stack types

**Phase 3: PropertyReader Modernization (Weeks 4-5)**
- Generate type-safe property accessors
- Reduce reflection in hot paths
- Maintain backward compatibility with existing auto_reader.go

**Phase 4: Move System (Week 6)**
- Generic move base types
- Type-safe state access in Legal/Apply
- Update moves package

**Phase 5: Code Generation (Week 7)**
- Update boardgame-util templates
- Generate generic-aware auto_reader.go
- Simplify generated code

**Phase 6: Example Migration (Week 8+)**
- Migrate each example game
- Document patterns
- Create migration guide

### Backward Compatibility Strategy

**Dual API Approach**: Maintain both legacy and generic APIs during transition.

```go
// Legacy API (preserved)
type ImmutableState interface {
    ImmutableGameState() ImmutableSubState
    ImmutablePlayerStates() []ImmutableSubState
    // ...
}

// New Generic API (added)
type TypedImmutableState[G GameState, P PlayerState] interface {
    ImmutableState  // Embed legacy interface
    TypedGameState() *G
    TypedPlayerStates() []*P
    // ...
}
```

**Migration Path**:
1. Existing games continue using `concreteStates()` pattern
2. New games use generic API
3. Games migrate incrementally by switching to generic types
4. Legacy API deprecated in v2.0, removed in v3.0

---

## API Changes by Component

### 1. State System

#### Current API (state.go:24-113)

```go
type ImmutableState interface {
    ImmutableGameState() ImmutableSubState
    ImmutablePlayerStates() []ImmutableSubState
    ImmutableCurrentPlayer() ImmutableSubState
    // ... other methods
}

type State interface {
    ImmutableState
    GameState() SubState
    PlayerStates() []PlayerState
    CurrentPlayer() PlayerState
}
```

**Current Usage Pattern** (examples/memory/state.go:32):
```go
func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
    game := state.ImmutableGameState().(*gameState)
    players := make([]*playerState, len(state.ImmutablePlayerStates()))
    for i, player := range state.ImmutablePlayerStates() {
        players[i] = player.(*playerState)
    }
    return game, players
}

// Used everywhere:
func (m *moveRevealCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    game, players := concreteStates(state)
    // Now use game and players with concrete types
}
```

#### New Generic API

```go
// New generic interfaces
type TypedImmutableState[G GameState, P PlayerState] interface {
    ImmutableState  // Embed for compatibility

    // Type-safe accessors
    TypedGameState() *G
    TypedPlayerStates() []*P
    TypedCurrentPlayer() *P
}

type TypedState[G GameState, P PlayerState] interface {
    TypedImmutableState[G, P]
    State  // Embed for compatibility

    // Mutable type-safe accessors
    MutableGameState() *G
    MutablePlayerStates() []*P
    MutableCurrentPlayer() *P
}

// Constraint interfaces
type GameState interface {
    SubState
}

type PlayerState interface {
    SubState
}
```

#### Usage After Migration

```go
// No concreteStates() needed!
func (m *moveRevealCard) Legal(state boardgame.TypedImmutableState[gameState, playerState], proposer boardgame.PlayerIndex) error {
    game := state.TypedGameState()
    players := state.TypedPlayerStates()
    // Direct access, type-safe!
}
```

**Breaking Changes**: None. New methods added to existing interfaces.

**Migration Steps**:
1. Add generic type parameters to GameDelegate
2. Update move method signatures to use `TypedImmutableState[G, P]`
3. Replace `concreteStates()` calls with direct `TypedGameState()` calls
4. Delete `concreteStates()` helper

---

### 2. Component System

#### Current API (component.go:24-65)

```go
type Component interface {
    Values() ComponentValues  // Returns interface{}
    Deck() *Deck
    DeckIndex() int
    // ...
}

type ComponentValues interface {
    Reader
    ReadSetter
    ReadSetConfigurer
}
```

**Current Usage** (components/playingcards/main_test.go:78):
```go
for i := 0; i < 52; i++ {
    card := components[i].Values().(*Card)  // Type assertion required
    // Use card
}
```

#### New Generic API

```go
// Generic component interface
type TypedComponent[V ComponentValues] interface {
    Component  // Embed for compatibility

    // Type-safe value access
    TypedValues() *V
}

// Generic deck interface
type TypedDeck[V ComponentValues] struct {
    *Deck  // Embed
    // Internal type information
}

func (d *TypedDeck[V]) NewComponent(values *V) TypedComponent[V] {
    // ...
}

func (d *TypedDeck[V]) ComponentAt(index int) TypedComponent[V] {
    // ...
}
```

#### Usage After Migration

```go
// In GameDelegate.ConfigureDecks()
func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
    return map[string]*boardgame.Deck{
        "cards": boardgame.NewTypedDeck[Card](),
    }
}

// Type-safe access
deck := manager.Chest().TypedDeck[Card]("cards")
card := deck.ComponentAt(0)
values := card.TypedValues()  // *Card, no casting!
```

**Breaking Changes**: None for existing code using `Values()`.

---

### 3. Stack System

#### Current API (stack.go:18-200)

```go
type ImmutableStack interface {
    Deck() *Deck
    ImmutableComponentAt(index int) ImmutableComponentInstance
    ImmutableComponents() []ImmutableComponentInstance
    // ...
}

type Stack interface {
    ImmutableStack
    ComponentAt(index int) ComponentInstance
    // ...
}
```

#### New Generic API

```go
type TypedStack[V ComponentValues] interface {
    Stack  // Embed for compatibility

    // Type-safe component access
    TypedComponentAt(index int) TypedComponentInstance[V]
    TypedComponents() []TypedComponentInstance[V]

    // Deck accessor returns typed deck
    TypedDeck() *TypedDeck[V]
}

type TypedComponentInstance[V ComponentValues] interface {
    ComponentInstance

    // Type-safe value access
    TypedValues() *V
    TypedDynamicValues() *V  // If dynamic values are same type
}
```

**Breaking Changes**: None.

---

### 4. GameDelegate Interface

#### Current API (game_delegate.go:18+, ARCHITECTURE.md:348-388)

```go
type GameDelegate interface {
    Name() string
    // ...

    GameStateConstructor() ConfigurableSubState
    PlayerStateConstructor(playerIndex PlayerIndex) ConfigurableSubState
    DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState

    ConfigureMoves() []MoveConfig
    ConfigureDecks() map[string]*Deck
    // ...
}
```

#### New Generic API

```go
// New generic delegate interface
type TypedGameDelegate[G GameState, P PlayerState] interface {
    GameDelegate  // Embed for compatibility

    // Type information (used internally by framework)
    GameStateType() G
    PlayerStateType() P
}

// Helper function to create typed delegates
func NewTypedDelegate[G GameState, P PlayerState](base GameDelegate) TypedGameDelegate[G, P] {
    return &typedDelegateWrapper[G, P]{base}
}
```

#### Usage After Migration

```go
type gameDelegate struct {
    base.GameDelegate
}

// In main.go NewManager()
func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
    delegate := &gameDelegate{}

    // Wrap with type information
    typedDelegate := boardgame.NewTypedDelegate[gameState, playerState](delegate)

    return boardgame.NewGameManager(typedDelegate, storage)
}
```

**Breaking Changes**: None for existing delegates.

---

### 5. Move Interface

#### Current API (move.go:189-219)

```go
type Move interface {
    Legal(state ImmutableState, proposer PlayerIndex) error
    Apply(state State) error
    // ...
}
```

#### New Generic API

```go
// Generic move interface
type TypedMove[G GameState, P PlayerState] interface {
    Move  // Embed for compatibility

    // Type-safe methods
    TypedLegal(state TypedImmutableState[G, P], proposer PlayerIndex) error
    TypedApply(state TypedState[G, P]) error
}

// Base move with generics
type BaseMove[G GameState, P PlayerState] struct {
    base.Move
}

func (m *BaseMove[G, P]) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    // Call TypedLegal with typed state
    return m.TypedLegal(state.(boardgame.TypedImmutableState[G, P]), proposer)
}
```

#### Usage After Migration

```go
type moveRollDice struct {
    boardgame.BaseMove[gameState, playerState]
}

func (m *moveRollDice) TypedLegal(state boardgame.TypedImmutableState[gameState, playerState], proposer boardgame.PlayerIndex) error {
    game := state.TypedGameState()  // No casting!
    players := state.TypedPlayerStates()

    // Game logic here
    return nil
}

func (m *moveRollDice) TypedApply(state boardgame.TypedState[gameState, playerState]) error {
    game := state.MutableGameState()
    // Modify state
    return nil
}
```

**Breaking Changes**: None. TypedMove is optional enhancement.

---

### 6. PropertyReader System

#### Current API (property_reader.go:159-175, type_casting_analysis.md:155-187)

```go
type PropertyReader interface {
    Prop(name string) (interface{}, error)
    IntProp(name string) (int, error)
    StringProp(name string) (string, error)
    // ... many typed accessors
}

type PropertyReadSetter interface {
    PropertyReader
    SetProp(name string, value interface{}) error
    SetIntProp(name string, value int) error
    // ...
}
```

**Current Generated Code** (auto_reader.go in each game):
```go
func (g *gameState) Reader() boardgame.PropertyReader {
    return &gameStateReader{g}
}

type gameStateReader struct {
    data *gameState
}

func (g *gameStateReader) Props() map[string]boardgame.PropertyType {
    return map[string]boardgame.PropertyType{
        "CurrentPlayer": boardgame.TypeInt,
        "Phase": boardgame.TypeInt,
        // ... 50+ more lines
    }
}

func (g *gameStateReader) Prop(name string) (interface{}, error) {
    switch name {
    case "CurrentPlayer":
        return g.data.CurrentPlayer, nil
    case "Phase":
        return g.data.Phase, nil
    // ... massive switch statement
    }
}
```

#### New Generic API

**Strategy**: Generate simpler, type-safe accessors while maintaining backward compatibility.

```go
// Keep existing PropertyReader for serialization
// Add type-safe generated accessors

// Generated in auto_reader.go
type GameStateReader struct {
    data *gameState
}

// Type-safe accessor methods
func (r *GameStateReader) CurrentPlayer() int {
    return r.data.CurrentPlayer
}

func (r *GameStateReader) Phase() int {
    return r.data.Phase
}

// Implement PropertyReader for backward compatibility
func (r *GameStateReader) Prop(name string) (interface{}, error) {
    // Same as before
}
```

**Breaking Changes**: None. Adds type-safe accessors alongside existing reflection-based API.

---

## Detailed Changes by File/System

### Core Framework Files

#### state.go (~1500 lines)

**Lines to Change**:
- Add `TypedImmutableState[G, P]` interface after line 113
- Add `TypedState[G, P]` interface after existing State
- Implement generic methods on `*state` struct (line 300+)

**New Code**:
```go
// Add after line 113
type TypedImmutableState[G GameState, P PlayerState] interface {
    ImmutableState
    TypedGameState() *G
    TypedPlayerStates() []*P
    TypedCurrentPlayer() *P
}

type TypedState[G GameState, P PlayerState] interface {
    TypedImmutableState[G, P]
    State
    MutableGameState() *G
    MutablePlayerStates() []*P
    MutableCurrentPlayer() *P
}

// Implement on *state
func (s *state) TypedGameState[G GameState]() *G {
    return s.gameState.(*G)
}

func (s *state) TypedPlayerStates[P PlayerState]() []*P {
    result := make([]*P, len(s.playerStates))
    for i, p := range s.playerStates {
        result[i] = p.(*P)
    }
    return result
}
```

**Testing**: All existing state_test.go tests must pass. Add new tests for generic accessors.

---

#### component.go (~800 lines)

**Lines to Change**:
- Add `TypedComponent[V]` interface after line 65
- Add `TypedDeck[V]` struct wrapping `*Deck`
- Implement type-safe methods (lines 266, 279)

**Current Type Assertions** (type_casting_analysis.md:63-90):
- Line 266: `st.(*state)` - Internal framework use
- Line 279: `st.(*state)` - Internal framework use

**New Code**:
```go
type TypedComponent[V ComponentValues] interface {
    Component
    TypedValues() *V
}

type TypedDeck[V ComponentValues] struct {
    *Deck
}

func NewTypedDeck[V ComponentValues]() *TypedDeck[V] {
    return &TypedDeck[V]{
        Deck: NewDeck(),
    }
}

func (d *TypedDeck[V]) ComponentAt(index int) TypedComponent[V] {
    // ...
}
```

---

#### stack.go (~2000 lines)

**Lines to Change**:
- Add `TypedStack[V]` interface after line 200
- Add `TypedComponentInstance[V]` interface
- Update stack implementations (lines 584, 608, 1642)

**Current Type Assertions** (type_casting_analysis.md:93-121):
- Line 584: `other.(*growableStack)`
- Line 608: `other.(*sizedStack)`
- Line 1642: `stack.(*growableStack)`

**Strategy**: Keep internal assertions, add generic wrapper types.

---

#### game_manager.go (~800 lines)

**Changes**:
- Update `NewGameManager` to accept `TypedGameDelegate[G, P]`
- Store type information for runtime use
- No breaking changes to existing API

**Current Type Assertions** (type_casting_analysis.md:79-89):
- Line 103: `st.(*state)` - Internal use, keep

---

#### move.go (~400 lines)

**Changes**:
- Add `TypedMove[G, P]` interface after line 219
- Update `moves/` package base types

---

#### property_reader.go (~1500 lines)

**Strategy**:
- Keep existing reflection-based implementation for serialization
- Generate additional type-safe accessors in auto_reader.go
- No changes to core property_reader.go needed

**Reflection Usage** (type_casting_analysis.md:155-187):
- 54+ reflection calls - KEEP for backward compatibility and serialization
- Generate parallel type-safe API in code generation phase

---

### Moves Package (~27,450 lines)

#### moves/default.go

**Add Generic Base Types**:
```go
type Base[G boardgame.GameState, P boardgame.PlayerState] struct {
    Base  // Embed existing
}

func (b *Base[G, P]) TypedLegal(state boardgame.TypedImmutableState[G, P], proposer boardgame.PlayerIndex) error {
    // Delegate to Legal for backward compatibility
    return b.Legal(state, proposer)
}
```

**Migration**: Each of 30+ move types gets generic version:
- `DealComponents[G, P]`
- `CollectComponents[G, P]`
- `MoveComponents[G, P]`
- `FinishTurn[G, P]`
- etc.

---

### Code Generation (boardgame-util/lib/codegen/)

#### Current Generated Code Pattern

**auto_reader.go** (1000+ lines per game):
```go
type gameStateReader struct {
    data *gameState
}

func (g *gameStateReader) Prop(name string) (interface{}, error) {
    switch name {
    case "CurrentPlayer":
        return g.data.CurrentPlayer, nil
    // ... 100+ cases
    }
}

func (g *gameStateReader) SetProp(name string, value interface{}) error {
    switch name {
    case "CurrentPlayer":
        val, ok := value.(int)
        if !ok {
            return errors.New("wrong type")
        }
        g.data.CurrentPlayer = val
    // ... 100+ cases
    }
}
```

#### New Generated Code (With Generics)

**Phase 1**: Keep existing generation, add type-safe methods
```go
// Keep all existing PropertyReader implementation

// Add type-safe accessors
func (r *gameStateReader) GetCurrentPlayer() int {
    return r.data.CurrentPlayer
}

func (r *gameStateReader) SetCurrentPlayer(value int) error {
    r.data.CurrentPlayer = value
    return nil
}
```

**Phase 2** (Future optimization): Generate generic struct accessors
```go
// Much simpler generated code
type gameStateAccessor struct {
    data *gameState
}

func (a *gameStateAccessor) CurrentPlayer() *int {
    return &a.data.CurrentPlayer
}
```

**Files to Update**:
- `/Users/jkomoros/Code/boardgame/boardgame-util/lib/codegen/reader.go` - Main template
- `/Users/jkomoros/Code/boardgame/boardgame-util/lib/stub/templates.go` - Starter templates

---

## Implementation Order

### Phase 1: Foundation (Week 1-2)

**Goal**: Add generic State interfaces without breaking existing code.

**PR #1: Add Generic State Interfaces**
- File: `/Users/jkomoros/Code/boardgame/state.go`
- Add `TypedImmutableState[G, P]` interface
- Add `TypedState[G, P]` interface
- Add constraint interfaces `GameState`, `PlayerState`
- Implement on `*state` type
- Tests: `state_test.go`

**PR #2: Add TypedGameDelegate**
- File: `/Users/jkomoros/Code/boardgame/game_delegate.go`
- Add `TypedGameDelegate[G, P]` interface
- Add wrapper implementation
- Tests: `game_delegate_test.go`

**Verification**:
```bash
cd /Users/jkomoros/Code/boardgame
go test ./... -v  # All tests pass
```

---

### Phase 2: Component System (Week 3)

**PR #3: Add Generic Component Types**
- File: `/Users/jkomoros/Code/boardgame/component.go`
- Add `TypedComponent[V]` interface
- Add `TypedDeck[V]` struct
- Update ComponentChest to support typed decks

**PR #4: Add Generic Stack Types**
- File: `/Users/jkomoros/Code/boardgame/stack.go`
- Add `TypedStack[V]` interface
- Add `TypedComponentInstance[V]` interface

**Dependencies**: PR #1, #2 merged

---

### Phase 3: Move System (Week 4)

**PR #5: Add Generic Move Interfaces**
- File: `/Users/jkomoros/Code/boardgame/move.go`
- Add `TypedMove[G, P]` interface
- Add `BaseMove[G, P]` implementation

**PR #6: Update Moves Package**
- Files: `/Users/jkomoros/Code/boardgame/moves/*.go`
- Add generic variants of all base move types
- `Base[G, P]`, `FixUp[G, P]`, etc.

**Dependencies**: PR #1-4 merged

---

### Phase 4: Code Generation (Week 5-6)

**PR #7: Update Code Generation Templates**
- File: `/Users/jkomoros/Code/boardgame/boardgame-util/lib/codegen/reader.go`
- Generate type-safe accessor methods alongside PropertyReader
- Update test fixtures

**PR #8: Update Starter Templates**
- File: `/Users/jkomoros/Code/boardgame/boardgame-util/lib/stub/templates.go`
- New game template uses generic APIs
- Includes migration examples

**Dependencies**: PR #1-6 merged

---

### Phase 5: Example Migration (Week 7-10)

**PR #9: Migrate examples/pig** (Simplest)
- Remove `concreteStates()`
- Update move signatures
- Regenerate auto_reader.go
- Document patterns

**PR #10: Migrate examples/memory**
- Similar to pig
- More complex state

**PR #11: Migrate examples/blackjack**
- Multiple phases
- Component values usage

**PR #12: Migrate examples/tictactoe**
**PR #13: Migrate examples/checkers**
**PR #14: Migrate examples/debuganimations**

**Order Rationale**: Start with simplest (pig), progress to most complex (checkers).

---

### Phase 6: Documentation (Week 11)

**PR #15: Update Documentation**
- `/Users/jkomoros/Code/boardgame/TUTORIAL.md` - Add generics section
- `/Users/jkomoros/Code/boardgame/ARCHITECTURE.md` - Update architecture docs
- `/Users/jkomoros/Code/boardgame/README.md` - Update prerequisites
- Add `/Users/jkomoros/Code/boardgame/MIGRATION_GUIDE.md`

---

## Impact Analysis

### What Breaks for Existing Games

**Immediate Impact**: **NONE**

All existing games continue to work without modification:
- Legacy interfaces preserved
- `concreteStates()` pattern still works
- No API removals
- Backward compatible at binary level

**Optional Migration Impact**:

When a game chooses to migrate:
1. **Update GameDelegate wrapper** (5 minutes)
   ```go
   // Add single line in NewManager:
   typedDelegate := boardgame.NewTypedDelegate[gameState, playerState](delegate)
   ```

2. **Update move signatures** (5 minutes per move)
   ```go
   // Before:
   func (m *moveRollDice) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error

   // After:
   func (m *moveRollDice) TypedLegal(state boardgame.TypedImmutableState[gameState, playerState], proposer boardgame.PlayerIndex) error
   ```

3. **Replace concreteStates calls** (2 minutes per call, ~20 calls per game)
   ```go
   // Before:
   game, players := concreteStates(state)

   // After:
   game := state.TypedGameState()
   players := state.TypedPlayerStates()
   ```

4. **Delete concreteStates helper** (10 seconds)

5. **Regenerate auto_reader.go** (automated)
   ```bash
   boardgame-util codegen
   ```

**Total Migration Time per Game**: 2-4 hours

---

### What Stays the Same

1. **Storage Layer**: No changes to serialization format
2. **Network Protocol**: JSON API unchanged
3. **Web Frontend**: No JavaScript changes needed
4. **Storage Backends**: MySQL, Bolt, filesystem all unchanged
5. **Game Logic**: Core game rules and state management unchanged
6. **Property System**: Reflection-based properties still work
7. **Sanitization**: Policy system unchanged
8. **Animation System**: Frontend animations unchanged

---

### Migration Effort by Game

Based on example game analysis:

| Game | Lines of Code | concreteStates Calls | Moves | Estimated Time |
|------|---------------|---------------------|-------|----------------|
| pig | ~500 | 10 | 3 | 2 hours |
| memory | ~800 | 15 | 5 | 2-3 hours |
| tictactoe | ~600 | 12 | 3 | 2 hours |
| blackjack | ~1200 | 25 | 8 | 3-4 hours |
| checkers | ~2000 | 40 | 12 | 4-6 hours |
| debuganimations | ~400 | 8 | 2 | 1-2 hours |

**Framework Migration**: 6-8 weeks (core team)
**Per-Game Migration**: 2-6 hours (game developers, optional)

---

### Performance Implications

**Expected Performance Improvements**:

1. **Type Assertions Eliminated**: ~125 calls per game → 0
   - **Impact**: Negligible (type assertions are very fast in Go)
   - **Benefit**: Compile-time safety

2. **Reflection Reduction**: Potential 20-30% reduction in reflection calls
   - **Impact**: Measurable in property-heavy operations
   - **Benefit**: Faster state serialization/deserialization

3. **Generated Code Simplification**: Smaller auto_reader.go files
   - **Impact**: Faster compilation times
   - **Benefit**: Better maintainability

4. **Memory Allocation**: Slightly reduced due to fewer interface conversions
   - **Impact**: Minor (Go's interface conversions are efficient)
   - **Benefit**: Cleaner memory profiles

**Benchmarks Needed**:
```bash
# Before and after comparison
go test -bench=. -benchmem ./...

# Focus areas:
# - State.Copy() performance
# - Property reading/writing
# - Move Legal/Apply cycles
# - Serialization/deserialization
```

**Expected Outcome**: 0-5% performance improvement, 100% type safety improvement.

---

## Testing Strategy

### Test Categories

#### 1. Framework Core Tests

**Existing Tests** (~50 test files):
- Must all pass without modification
- Add new generic API tests alongside

**New Test Files**:
```
state_generic_test.go
component_generic_test.go
move_generic_test.go
```

**Test Pattern**:
```go
func TestTypedStateAccessors(t *testing.T) {
    // Test generic state access
    manager := setupTestManager[testGameState, testPlayerState](t)
    game := manager.NewGame()
    state := game.State().(TypedState[testGameState, testPlayerState])

    gameState := state.TypedGameState()
    assert.NotNil(t, gameState)

    playerStates := state.TypedPlayerStates()
    assert.Equal(t, 2, len(playerStates))
}
```

---

#### 2. Integration Tests

**Full Game Simulation**:
```go
func TestGenericGameFlow(t *testing.T) {
    // Pig game using generic API
    manager := pig.NewGenericManager(storage.NewMemoryStorageManager())
    game := manager.NewGame()

    // Propose moves with type safety
    move := &moveRollDice{}
    err := game.ProposeMove(move, 0)
    assert.NoError(t, err)

    // Verify state changes
    state := game.State().(boardgame.TypedImmutableState[pig.GameState, pig.PlayerState])
    gameState := state.TypedGameState()
    assert.NotZero(t, gameState.Round)
}
```

---

#### 3. Backward Compatibility Tests

**Legacy API Still Works**:
```go
func TestLegacyAPICompatibility(t *testing.T) {
    // Old-style game creation
    manager := pig.NewManager(storage.NewMemoryStorageManager())
    game := manager.NewGame()

    // Old-style state access
    state := game.State()
    gameState := state.ImmutableGameState().(*pig.gameState)
    assert.NotNil(t, gameState)

    // concreteStates still works
    game2, players := pig.concreteStates(state)
    assert.NotNil(t, game2)
    assert.Equal(t, 2, len(players))
}
```

---

#### 4. Code Generation Tests

**Template Tests**:
```bash
# Test codegen produces valid code
cd /Users/jkomoros/Code/boardgame/boardgame-util
go test ./lib/codegen/... -v

# Test generated code compiles
cd /Users/jkomoros/Code/boardgame/examples/pig
boardgame-util codegen
go build ./...
```

**Round-Trip Tests**:
```go
func TestGeneratedCodeRoundTrip(t *testing.T) {
    // Serialize and deserialize using generated readers
    state := createTestState()

    // Use generated PropertyReader
    reader := state.GameState().Reader()
    value, err := reader.Prop("CurrentPlayer")
    assert.NoError(t, err)

    // Use generated type-safe accessor
    typedReader := state.GameState().(*gameState).TypedReader()
    typedValue := typedReader.GetCurrentPlayer()
    assert.Equal(t, value, typedValue)
}
```

---

#### 5. Migration Tests

**Test Each Example Game**:
```bash
# Automated test suite
./test-migration.sh examples/pig
./test-migration.sh examples/memory
./test-migration.sh examples/blackjack
# etc.
```

**Script Contents**:
```bash
#!/bin/bash
GAME=$1
cd $GAME

# Run tests before migration
go test ./... -v > before.txt

# Perform migration (manual or automated)
# ...

# Regenerate code
boardgame-util codegen

# Run tests after migration
go test ./... -v > after.txt

# Compare results
diff before.txt after.txt
```

---

### Test Coverage Goals

**Target Coverage**: 85%+ for new generic code

**Measurement**:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Critical Paths**:
1. Generic state accessors - 100%
2. Component type safety - 100%
3. Move system generics - 95%
4. Code generation - 90%
5. Backward compatibility - 100%

---

## Documentation Updates

### Files to Update

#### 1. TUTORIAL.md (~3500 lines)

**Section**: "The ConcreteStates Pattern" (around line 500)

**Add New Section**: "Using Generic Types (Go 1.18+)"

```markdown
## Using Generic Types (Go 1.18+)

As of Go 1.18, the boardgame framework supports generic types that eliminate
the need for the `concreteStates()` pattern.

### Before (Legacy Pattern)

```go
func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
    game := state.ImmutableGameState().(*gameState)
    players := make([]*playerState, len(state.ImmutablePlayerStates()))
    for i, player := range state.ImmutablePlayerStates() {
        players[i] = player.(*playerState)
    }
    return game, players
}
```

### After (Generic Pattern)

```go
// No helper needed! Use TypedImmutableState[G, P]
func (m *moveRollDice) TypedLegal(state boardgame.TypedImmutableState[gameState, playerState], proposer boardgame.PlayerIndex) error {
    game := state.TypedGameState()  // *gameState, type-safe!
    players := state.TypedPlayerStates()  // []*playerState
    // ...
}
```

### Migration Steps

1. Wrap your delegate with type parameters in `NewManager()`:
   ```go
   typedDelegate := boardgame.NewTypedDelegate[gameState, playerState](delegate)
   return boardgame.NewGameManager(typedDelegate, storage)
   ```

2. Update move methods to use `TypedLegal` and `TypedApply`

3. Replace `concreteStates()` calls with direct typed accessors

4. Regenerate auto_reader.go: `boardgame-util codegen`

For complete migration guide, see MIGRATION_GUIDE.md.
```

---

#### 2. ARCHITECTURE.md (~4000 lines)

**Section**: "State System: Immutability and Versioning" (around line 443)

**Add Subsection**: "Generic State Types (Go 1.18+)"

```markdown
### Generic State Types (Go 1.18+)

The framework now provides generic variants of State interfaces that eliminate type assertions:

```go
type TypedImmutableState[G GameState, P PlayerState] interface {
    ImmutableState  // Embeds legacy interface
    TypedGameState() *G
    TypedPlayerStates() []*P
    TypedCurrentPlayer() *P
}
```

**Benefits**:
- Compile-time type safety
- No type assertions needed
- Better IDE autocomplete
- Clearer code intent

**When to use**:
- New games: Always use generic types
- Existing games: Migrate when convenient (backward compatible)
- Libraries: Provide both APIs during transition

See TypedState and TypedMove interfaces for complete API.
```

---

#### 3. README.md

**Section**: "Prerequisites"

**Update Requirements**:
```markdown
## Prerequisites

- **Go 1.18 or later** (required for generic types)
  - Generics support added in Go 1.18
  - Type inference improvements in Go 1.21+
  - Older Go versions: Use legacy API (pre-generics branch)

- **Database** (optional): MySQL 5.7+ or BoltDB
```

---

#### 4. NEW: MIGRATION_GUIDE.md

**Create Comprehensive Migration Guide**:

```markdown
# Migration Guide: Legacy to Generic Types

This guide walks through migrating an existing boardgame from legacy
interface{} types to generic types introduced in Go 1.18.

## Overview

**Time Required**: 2-4 hours per game
**Difficulty**: Easy
**Breaking Changes**: None (opt-in migration)

## Prerequisites

1. Update Go version: `go 1.18` or later in go.mod
2. Update boardgame framework to v2.0+
3. Backup your code (or create a git branch)

## Step-by-Step Migration

### Step 1: Update GameDelegate Wrapper (5 minutes)

**File**: `main.go`

**Before**:
```go
func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
    return boardgame.NewGameManager(&gameDelegate{}, storage)
}
```

**After**:
```go
func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
    delegate := &gameDelegate{}
    typedDelegate := boardgame.NewTypedDelegate[gameState, playerState](delegate)
    return boardgame.NewGameManager(typedDelegate, storage)
}
```

### Step 2: Update Move Interfaces (5 min per move)

**Before**:
```go
func (m *moveRollDice) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    game, players := concreteStates(state)
    // ...
}

func (m *moveRollDice) Apply(state boardgame.State) error {
    game, players := concreteStates(state)
    // ...
}
```

**After**:
```go
func (m *moveRollDice) TypedLegal(state boardgame.TypedImmutableState[gameState, playerState], proposer boardgame.PlayerIndex) error {
    game := state.TypedGameState()
    players := state.TypedPlayerStates()
    // ... rest of code unchanged
}

func (m *moveRollDice) TypedApply(state boardgame.TypedState[gameState, playerState]) error {
    game := state.MutableGameState()
    players := state.MutablePlayerStates()
    // ... rest of code unchanged
}
```

**Find/Replace Hints**:
- Find: `game, players := concreteStates(state)`
- Replace: `game := state.TypedGameState(); players := state.TypedPlayerStates()`

### Step 3: Update State Methods (2 min per method)

Methods on your gameState/playerState that accept state parameters:

**Before**:
```go
func (p *playerState) TurnDone() error {
    game, _ := concreteStates(p.State())
    // ...
}
```

**After**:
```go
func (p *playerState) TurnDone() error {
    state := p.State().(boardgame.TypedImmutableState[gameState, playerState])
    game := state.TypedGameState()
    // ...
}
```

### Step 4: Delete concreteStates Helper

**File**: `state.go`

Delete the entire function:
```go
func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
    // DELETE THIS FUNCTION
}
```

### Step 5: Regenerate auto_reader.go

```bash
boardgame-util codegen
```

### Step 6: Test

```bash
go test ./... -v
```

All tests should pass. If not, check for:
- Missed concreteStates() calls
- State parameter types incorrect
- Missing type assertions

## Common Patterns

### Pattern: Current Player Access

**Before**:
```go
game, players := concreteStates(state)
currentPlayer := players[game.CurrentPlayer]
```

**After**:
```go
currentPlayer := state.TypedCurrentPlayer()
```

### Pattern: Component Values

**Before**:
```go
card := component.Values().(*Card)
```

**After**:
```go
card := component.TypedValues()  // *Card, type-safe
```

### Pattern: Stack Iteration

**Before**:
```go
for i := 0; i < stack.Len(); i++ {
    component := stack.ComponentAt(i)
    card := component.Values().(*Card)
    // ...
}
```

**After**:
```go
stack := gameState.DrawDeck.(boardgame.TypedStack[Card])
for i := 0; i < stack.Len(); i++ {
    component := stack.TypedComponentAt(i)
    card := component.TypedValues()  // Type-safe!
    // ...
}
```

## Troubleshooting

### Error: "cannot use state (type boardgame.ImmutableState) as type boardgame.TypedImmutableState[gameState,playerState]"

**Solution**: Add type assertion:
```go
typedState := state.(boardgame.TypedImmutableState[gameState, playerState])
```

Or update the function signature to accept typed state.

### Error: "undefined: TypedGameState"

**Solution**: Update boardgame framework to v2.0+:
```bash
go get github.com/jkomoros/boardgame@latest
```

### All Tests Fail After Migration

**Solution**: Check that you regenerated auto_reader.go:
```bash
boardgame-util codegen
```

## Complete Example

See `/examples/pig` for a fully migrated example game.

## Questions?

- GitHub Issues: https://github.com/jkomoros/boardgame/issues
- Documentation: /Users/jkomoros/Code/boardgame/ARCHITECTURE.md
```

---

### Documentation Timeline

| Week | Document | Owner | Status |
|------|----------|-------|--------|
| 11 | MIGRATION_GUIDE.md | Team | New |
| 11 | TUTORIAL.md updates | Team | Update |
| 11 | ARCHITECTURE.md updates | Team | Update |
| 11 | README.md updates | Team | Update |
| 11 | Code comments | Team | Update |

---

## Risk Mitigation

### Identified Risks

#### Risk 1: Breaking Backward Compatibility

**Likelihood**: Low
**Impact**: High
**Mitigation**:
- Maintain dual API (legacy + generic) for 2+ major versions
- Extensive backward compatibility testing
- Deprecation warnings before removal
- Clear migration timeline (v2.0 introduces, v3.0 removes legacy)

---

#### Risk 2: Code Generation Complexity

**Likelihood**: Medium
**Impact**: Medium
**Mitigation**:
- Start with additive changes (add methods, don't remove)
- Extensive codegen tests with fixtures
- Manual review of generated code samples
- Rollback plan: Keep old templates

---

#### Risk 3: Type Inference Issues

**Likelihood**: Medium
**Impact**: Low
**Mitigation**:
- Recommend Go 1.21+ for better type inference
- Provide explicit type parameter examples
- Document common inference failures
- IDE plugin recommendations (gopls)

---

#### Risk 4: Performance Regression

**Likelihood**: Low
**Impact**: Medium
**Mitigation**:
- Benchmark before/after each phase
- Profile hot paths (state.Copy(), property access)
- Regression testing in CI
- Performance gates (reject PRs with >5% regression)

**Benchmark Commands**:
```bash
# Baseline (before changes)
go test -bench=. -benchmem ./... > bench-before.txt

# After changes
go test -bench=. -benchmem ./... > bench-after.txt

# Compare
benchstat bench-before.txt bench-after.txt
```

---

#### Risk 5: Adoption Resistance

**Likelihood**: Medium
**Impact**: Low
**Mitigation**:
- Make migration optional (not forced)
- Provide automated migration tools where possible
- Excellent documentation and examples
- Migration office hours / support channel
- Celebrate early adopters

---

### Rollback Plan

**If critical issues arise**:

1. **Phase-level rollback**: Revert specific PR set
2. **Feature flagging**: Add runtime flag to disable generic APIs
3. **Git branches**: Maintain `pre-generics` branch indefinitely
4. **Documentation**: Clear "how to stay on legacy API" guide

**Rollback Trigger Criteria**:
- >10% performance regression
- Unfixable backward compatibility break
- Critical bug affecting all users
- Community feedback overwhelmingly negative

---

## Appendices

### Appendix A: Go Generics Best Practices

**References**:
- [Go Generics Tutorial](https://go.dev/doc/tutorial/generics)
- [Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md)
- [When to Use Generics](https://go.dev/blog/when-generics)

**Key Principles Applied**:

1. **Use generics for type containers**: ✓ State, Stack, Component
2. **Keep interfaces for behavior**: ✓ GameDelegate, Move still interfaces
3. **Type parameters for identical logic**: ✓ All state access patterns identical
4. **Avoid over-generalization**: ✓ Limited to clear use cases

---

### Appendix B: File Paths Reference

**Core Framework**:
- `/Users/jkomoros/Code/boardgame/state.go` - State interfaces
- `/Users/jkomoros/Code/boardgame/component.go` - Component system
- `/Users/jkomoros/Code/boardgame/stack.go` - Stack implementations
- `/Users/jkomoros/Code/boardgame/move.go` - Move interfaces
- `/Users/jkomoros/Code/boardgame/game_delegate.go` - GameDelegate interface
- `/Users/jkomoros/Code/boardgame/property_reader.go` - Property system

**Code Generation**:
- `/Users/jkomoros/Code/boardgame/boardgame-util/lib/codegen/reader.go` - auto_reader.go template
- `/Users/jkomoros/Code/boardgame/boardgame-util/lib/stub/templates.go` - New game templates

**Example Games**:
- `/Users/jkomoros/Code/boardgame/examples/pig/` - Simplest example
- `/Users/jkomoros/Code/boardgame/examples/memory/` - Card game
- `/Users/jkomoros/Code/boardgame/examples/blackjack/` - Complex phases
- `/Users/jkomoros/Code/boardgame/examples/checkers/` - Board game

**Documentation**:
- `/Users/jkomoros/Code/boardgame/TUTORIAL.md` - Comprehensive tutorial
- `/Users/jkomoros/Code/boardgame/ARCHITECTURE.md` - Architecture docs
- `/Users/jkomoros/Code/boardgame/README.md` - Project overview

---

### Appendix C: Type Casting Statistics

**From**: `/Users/jkomoros/Code/boardgame/type_casting_analysis.md`

**Priority Tier 1** (User-Facing):
- 125 `concreteStates()` calls → 0 with generics
- 20+ component value casts → 0 with generics

**Priority Tier 2** (Framework Internal):
- 299 `interface{}` usages → Reduced by ~40%
- 54 reflection calls → Maintained for serialization

**Priority Tier 3** (Medium Impact):
- 50+ state interface casts → Mostly internal, keep
- 30+ move type casts → Mostly tests, low priority

**Projected Elimination**: 60-70% of user-facing type casts

---

### Appendix D: Code Size Comparison

**Before Migration (Typical Game)**:

```
state.go:         150 lines (includes concreteStates)
moves.go:         300 lines (20 concreteStates calls)
auto_reader.go:   1200 lines (generated)
TOTAL:            1650 lines
```

**After Migration**:

```
state.go:         120 lines (no concreteStates)
moves.go:         280 lines (typed accessors)
auto_reader.go:   1000 lines (simpler generation)
TOTAL:            1400 lines (-15%)
```

**Boilerplate Reduction**: ~15% fewer lines, higher type safety.

---

### Appendix E: Timeline Summary

```
Week 1-2:   State System (PR #1-2)
Week 3:     Component System (PR #3-4)
Week 4:     Move System (PR #5-6)
Week 5-6:   Code Generation (PR #7-8)
Week 7:     Pig Example (PR #9)
Week 8:     Memory Example (PR #10)
Week 9:     Blackjack Example (PR #11)
Week 10:    Remaining Examples (PR #12-14)
Week 11:    Documentation (PR #15)

TOTAL: 11 weeks for complete migration
```

**Critical Path**: State System → Component System → Move System → Examples

**Parallelizable**: Documentation can start week 7, code generation concurrent with moves.

---

### Appendix F: Success Metrics

**Technical Metrics**:
- [ ] Zero type assertions in migrated game code
- [ ] All tests pass (100% of 200+ tests)
- [ ] No performance regression (±5% acceptable)
- [ ] Code coverage maintained (85%+)

**Adoption Metrics**:
- [ ] 3+ example games migrated
- [ ] Migration guide published
- [ ] Tutorial updated
- [ ] Community feedback collected

**Quality Metrics**:
- [ ] Zero backward compatibility breaks
- [ ] Documentation complete
- [ ] All PRs reviewed by 2+ people
- [ ] Benchmark suite passing

---

## Conclusion

This migration plan provides a comprehensive, phased approach to adding generics support to the boardgame framework while maintaining 100% backward compatibility. The migration eliminates the primary pain point (`concreteStates()` pattern) while providing a clear path forward for modern Go development.

**Next Steps**:
1. Review this plan with core team
2. Approve Go version upgrade (1.13 → 1.21)
3. Begin Phase 1: State System implementation
4. Iterate based on learnings from early phases

**Questions or Feedback**: Open issues at https://github.com/jkomoros/boardgame/issues

---

*Generated by the Generics Planning Team - 2026-02-07*
