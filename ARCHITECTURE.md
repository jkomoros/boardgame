# Boardgame Framework - Architecture Documentation

> **Last Updated:** 2026-02-03
> **Framework Version:** Go 1.13+ (compiles with Go 1.25.6)
> **Frontend:** Polymer 3 + lit-element 0.7.1

## Table of Contents

1. [Introduction & Overview](#introduction--overview)
2. [Core Framework Architecture](#core-framework-architecture)
3. [Game Development Patterns](#game-development-patterns)
4. [Storage Architecture](#storage-architecture)
5. [Server Architecture](#server-architecture)
6. [Web Frontend Architecture](#web-frontend-architecture)
7. [Development Tools](#development-tools)
8. [Key Design Patterns & Strengths](#key-design-patterns--strengths)
9. [Technical Debt & Known Issues](#technical-debt--known-issues)
10. [Modernization Strategies](#modernization-strategies)
11. [Execution Recommendations](#execution-recommendations)
12. [Verification Plan](#verification-plan)
13. [References & Further Reading](#references--further-reading)

---

## Introduction & Overview

### What is the Boardgame Framework?

The boardgame framework is a comprehensive Go-based game engine for creating multi-player board and card games with sophisticated web user interfaces. It enables game developers to define game rules and logic formally in Go, while providing automatic state management, persistence, networking, and rich web-based UIs with minimal configuration.

**Key Capabilities:**
- Define game rules once in Go, get multiplayer web app automatically
- Sophisticated animation system with automatic FLIP animations
- Multiple storage backends (Memory, Filesystem, Bolt, MySQL)
- Real-time updates via WebSockets
- Built-in sanitization for hidden information (cards, hidden state)
- Comprehensive move library with 30+ reusable move types
- Progressive Web App (PWA) support
- AI agent framework

**Primary Use Case:** Creating browser-based multiplayer board and card games without writing networking code, state synchronization logic, or animation systems from scratch.

### Documentation Ecosystem

This document provides a comprehensive architectural overview. For other aspects of the framework:

- **[README.md](README.md)** - Project overview, getting started, design goals, current status
- **[TUTORIAL.md](TUTORIAL.md)** - Comprehensive 127KB tutorial walking through building a complete game
- **[server/static/src/ARCHITECTURE.md](server/static/src/ARCHITECTURE.md)** - Detailed frontend animation system architecture
- **[boardgame-util/README.md](boardgame-util/README.md)** - CLI tool configuration and usage
- **[moves/doc.go](moves/doc.go)** - Extensive documentation of the moves package

**Audience for this document:**
- New contributors wanting to understand the framework's design
- Game developers needing to understand how their games interact with the framework
- Maintainers considering modernization strategies
- Developers evaluating technical debt and improvement opportunities

### Design Philosophy

From the project README, the framework was built with these core principles:

**Don't Repeat Yourself**
- Write your normative game logic in Go a single time
- The framework handles persistence, networking, state management, and UI automatically

**Minimize Code**
- Writing a game should feel like transcribing rules into a formal model
- Not like a challenging coding exercise

**Batteries Included**
- Common operations (e.g., "deal cards from draw stack to each player's hand until each has 3 cards") require minimal logic
- 30+ reusable move types cover most common game patterns
- 9 composable behaviors for common game mechanics

**Minimize Boilerplate**
- Structs with powerful defaults to anonymously embed
- Code generation tools eliminate repetitive code
- AutoConfigurer pattern reduces move configuration to a few lines

**Clean Layering**
- Override just the parts you need to customize
- Embed base types for instant functionality

**Flexible**
- Powerful enough to model any real-world board or card game
- Extensible through interfaces

**Make Cheaters' Lives Hard**
- Don't rely on security by obscurity
- Sanitize properties before transmitting to the client
- Built-in sanitization policies

**Identify Errors ASAP**
- Configuration errors caught at type-check time or at boot (NewGameManager)
- Not when someone's playing a game
- Minimal use of `interface{}` for type safety

**Fast**
- Minimal reliance on reflection at runtime
- Code generation provides reflection-free property access

**Minimize Javascript**
- Most client views are tens of lines of templates and databinding
- Sometimes without any JavaScript at all

**Rich Animations and UI**
- Automatic FLIP animations for component movement
- Smooth, fast animations computed automatically

**Robust Tooling**
- Swiss-army-knife `boardgame-util` utility
- Generate boilerplate, create starter projects, run dev servers, manage databases

### High-Level Architecture

The framework is organized in clear layers from low-level engine to high-level tools:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Development Tools Layer                      │
│  boardgame-util CLI: codegen, serve, build, db management       │
└─────────────────────────────────────────────────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                      Web Frontend Layer                          │
│  Polymer 3 / lit-element Components + Redux State Management    │
│  Animations: FLIP animation system with boardgame-component-     │
│             animator orchestrating smooth transitions            │
└─────────────────────────────────────────────────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                    Server & API Layer                            │
│  REST API (/api/game/{name}/{id}/{action})                      │
│  WebSocket real-time updates                                     │
│  Authentication (Firebase integration)                           │
└─────────────────────────────────────────────────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                      Storage Layer                               │
│  StorageManager interface with multiple backends:               │
│  Memory | Filesystem | Bolt | MySQL                             │
└─────────────────────────────────────────────────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                  High-Level Game Logic Layer                     │
│  moves/ package: 30+ reusable move types                        │
│  behaviors/ package: 9 composable game patterns                 │
│  base/ package: Embeddable implementations                      │
└─────────────────────────────────────────────────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                    Core Engine Layer                             │
│  boardgame/ package: GameManager, State, Move, Component        │
│  Interfaces: GameDelegate, Move, State, PropertyReader          │
│  Systems: Property system, Component chest, Sanitization        │
└─────────────────────────────────────────────────────────────────┘
```

**Data Flow:**

1. **Game Definition (Go):** Developer implements GameDelegate interface, defines state structs, creates move types
2. **Code Generation:** `//boardgame:codegen` directive triggers auto_reader.go generation for reflection-free property access
3. **Initialization:** NewGameManager validates configuration, sets up moves, connects to storage
4. **Game Creation:** Server API creates game, stores in database, assigns players
5. **Move Proposal:** Player clicks UI → propose-move event → REST API → GameManager.ProposeMove()
6. **Move Application:** Legal() check → Apply() modifies state → Storage saves → WebSocket notifies clients
7. **State Fetch:** Clients fetch new state via REST → State expansion → Redux store update
8. **Animation:** FLIP animation system animates components from old to new positions
9. **Rendering:** Game-specific renderer databinds state to Polymer/Lit components

### Technology Stack Summary

**Backend (Go):**
- **Language:** Go 1.13+ (declared in go.mod, compiles with Go 1.25.6)
- **Web Framework:** gin-gonic/gin
- **WebSocket:** gorilla/websocket
- **Logging:** sirupsen/logrus
- **Storage:** database/sql (MySQL), boltdb/bolt, filesystem, in-memory
- **Note:** Pre-generics Go (before 1.18) - relies heavily on code generation instead

**Frontend (JavaScript):**
- **Web Components:** Polymer 3.3.0 (20 components) + lit-element 2.3.1 (6 components)
- **State Management:** Redux 4.0.5 + Redux Thunk 2.3.0
- **Animations:** Web Animations API + CSS Transitions
- **Backend Integration:** Firebase 5.11.1 (authentication)
- **Build:** NO BUILD STEP - native ES modules with polyfills
- **PWA:** Service Worker + Web App Manifest

**Development Tools:**
- **Code Generation:** Custom codegen in boardgame-util (generates auto_reader.go, auto_enum.go)
- **CLI:** boardgame-util (serve, build, codegen, db commands)
- **Linting:** ESLint with Google config

**Architecture Timeline Context:**

This framework was created before Go 1.18 (which introduced generics in March 2022). As a result:
- Heavy use of `interface{}` where generics would now be used
- Extensive code generation to eliminate reflection at runtime
- PropertyReader pattern for type-safe property access without generics

The frontend was built on Polymer 3 and an early version of lit-element (0.7.1):
- Polymer is now a legacy framework (last major release 2018)
- Modern lit (v3.x) is significantly different from lit-element 0.7.1
- No TypeScript, no bundler, direct ES module loading

**Despite its age, the framework compiles and runs successfully on modern Go (1.25.6) and modern browsers (Chrome, Safari with limitations).**

### Package Structure Overview

```
boardgame/
├── boardgame/           # Core engine (~17,900 lines)
│   ├── game_manager.go  # Central orchestrator
│   ├── state.go         # State interfaces and implementations
│   ├── move.go          # Move system
│   ├── stack.go         # Stack and component collections
│   ├── component.go     # Component system
│   ├── deck.go          # Deck management
│   ├── property_reader.go # Property system
│   └── sanitization.go  # State sanitization for hidden info
│
├── base/                # Base implementations (~1,274 lines)
│   ├── game_delegate.go # Base GameDelegate to embed
│   ├── main.go          # SubState base struct
│   └── move.go          # Base Move implementation
│
├── moves/               # Reusable move library (~27,450 lines)
│   ├── doc.go           # Comprehensive package documentation
│   ├── default.go       # Base move with AutoConfigurer support
│   ├── auto_config.go   # AutoConfigurer for minimal boilerplate
│   ├── deal_components.go
│   ├── collect_components.go
│   ├── move_components.go
│   ├── round_robin.go
│   ├── start_phase.go
│   ├── finish_turn.go
│   └── ... (30+ move types)
│
├── behaviors/           # Composable patterns (~575 lines)
│   ├── current_player.go  # Track current player
│   ├── phase.go           # Game phase management
│   ├── round_robin.go     # Turn order cycling
│   ├── seat.go            # Player seating
│   ├── inactive_player.go # Player join/leave
│   ├── color.go           # Player colors
│   └── role.go            # Player roles
│
├── enum/                # Type-safe enumeration system (~2,112 lines)
│   ├── main.go
│   ├── tree.go
│   └── range.go
│
├── storage/             # Storage abstraction layer
│   ├── memory/          # In-memory storage (testing)
│   ├── filesystem/      # File-based persistence
│   ├── bolt/            # BoltDB backend
│   └── mysql/           # MySQL backend (production)
│
├── server/              # Web server (~2,448 lines API)
│   ├── api/             # REST API and WebSocket handlers
│   └── static/          # Frontend application
│       ├── index.html
│       ├── src/
│       │   ├── components/  # 26 Web Components
│       │   ├── actions/     # Redux actions
│       │   ├── reducers/    # Redux reducers
│       │   └── ARCHITECTURE.md  # Frontend animation docs
│       └── package.json
│
├── boardgame-util/      # CLI tool
│   ├── cmd_*.go         # 30+ commands
│   └── lib/
│       └── codegen/     # Code generation library
│
├── examples/            # 6 complete example games
│   ├── pig/             # Simple dice game
│   ├── memory/          # Card matching
│   ├── blackjack/       # Card game with phases
│   ├── checkers/        # Board game
│   ├── tictactoe/       # Classic game
│   └── debuganimations/ # Animation testing
│
└── components/          # Reusable components
    ├── playingcards/    # Standard 52-card deck
    └── dice/            # Dice components
```

**Package Size Summary:**
- Core framework: ~50,000 lines
- Example games: ~25,000 lines
- Web frontend: ~10,000 lines (HTML/JS/CSS)
- Total: ~85,000+ lines of code

---

## Core Framework Architecture

The core `boardgame` package (~17,900 lines) provides the fundamental abstractions and engine that power all games. This section describes the key systems and how they work together.

### GameManager: The Central Orchestrator

The GameManager is the heart of the framework. It coordinates all aspects of a game type:

```go
// From game_manager.go
type GameManager struct {
    delegate                  GameDelegate              // Game-specific logic
    gameValidator             *StructInflater           // Validates game state structs
    playerValidator           *StructInflater           // Validates player state structs
    dynamicComponentValidator map[string]*StructInflater // Validates component values
    chest                     *ComponentChest            // All components for this game type
    storage                   StorageManager             // Persistence layer
    agents                    []Agent                    // AI agents
    moves                     []*moveType                // All configured moves
    movesByName               map[string]*moveType       // Fast move lookup
    agentsByName              map[string]Agent           // Fast agent lookup
    modifiableGamesLock       sync.RWMutex               // Thread safety for game instances
    modifiableGames           map[string]*Game           // Active game instances
    timers                    *timerManager              // Timer coordination
    initialized               bool                       // Whether setup is complete
    logger                    *logrus.Logger             // Logging
    variantConfig             VariantConfig              // Game variants configuration
}
```

**Key Responsibilities:**
- **Validation:** At `NewGameManager()` time, validates all configuration (state structs, moves, components)
- **Move Management:** Registers and validates all moves, checks legality, applies moves
- **Game Lifecycle:** Creates games, loads from storage, manages active instances
- **Storage Coordination:** Reads/writes games to configured storage backend
- **Agent Management:** Coordinates AI agents
- **Timer Management:** Handles game timers for time-based mechanics

**Fail-Fast Philosophy:**
Most configuration errors are caught at `NewGameManager()` creation time, not during gameplay. This includes:
- Invalid state struct tags
- Missing required methods on moves
- Invalid component deck configurations
- Malformed sanitization policies

### GameDelegate: Your Game-Specific Brain

The GameDelegate interface is how you define your game's specific logic. The GameManager calls methods on your delegate at key points:

**Required Methods:**
```go
type GameDelegate interface {
    // Identification
    Name() string
    DisplayName() string
    Description() string

    // Player configuration
    MinNumPlayers() int
    MaxNumPlayers() int
    DefaultNumPlayers() int

    // State constructors
    GameStateConstructor() ConfigurableSubState
    PlayerStateConstructor(playerIndex boardgame.PlayerIndex) ConfigurableSubState
    DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState

    // Configuration
    ConfigureMoves() []MoveConfig
    ConfigureDecks() map[string]*Deck
    ConfigureConstants() PropertyCollection
    ConfigureAgents() []Agent

    // Computed properties (optional)
    ComputedGlobalProperties(state ImmutableState) PropertyCollection
    ComputedPlayerProperties(player ImmutablePlayerState) PropertyCollection

    // Lifecycle hooks (optional)
    BeginSetUp(state State, variant Variant) error
    DistributeComponentToStarterStack(state ImmutableState, c Component) (ImmutableStack, error)
    FinishSetUp(state State) error
    CheckGameFinished(state ImmutableState) (finished bool, winners []PlayerIndex)
    GameEndConditionMet(state ImmutableState) bool
    PlayerScore(player ImmutableSubState) int

    // Utilities (optional)
    Diagram(state ImmutableState) string
    Variants() VariantConfig
    LegalNumPlayers(numPlayers int) bool

    // ... additional methods for sanitization, graphics, etc.
}
```

**Common Pattern - Embedding base.GameDelegate:**

```go
// From examples/pig/main.go
type gameDelegate struct {
    base.GameDelegate  // Embed for default implementations
}

func (g *gameDelegate) Name() string {
    // Use reflection to match package name
    if memoizedDelegateName == "" {
        pkgPath := reflect.ValueOf(g).Elem().Type().PkgPath()
        pathPieces := strings.Split(pkgPath, "/")
        memoizedDelegateName = pathPieces[len(pathPieces)-1]
    }
    return memoizedDelegateName
}

func (g *gameDelegate) Description() string {
    return "Players roll the dice, collecting points, but bust if they roll a one."
}

func (g *gameDelegate) MinNumPlayers() int { return 2 }
func (g *gameDelegate) MaxNumPlayers() int { return 6 }
func (g *gameDelegate) DefaultNumPlayers() int { return 2 }

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
    return &gameState{}
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurableSubState {
    return &playerState{}
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    // Configure moves (see Moves section)
    auto := moves.NewAutoConfigurer(g)
    return moves.Add(
        auto.MustConfig(new(moveRollDice)),
        auto.MustConfig(new(moveCountDie), moves.WithIsFixUp(true)),
        auto.MustConfig(new(moveFinishTurn)),
    )
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
    return map[string]*boardgame.Deck{
        diceDeckName: dice.NewDeck(false),  // Standard 6-sided die
    }
}
```

### State System: Immutability and Versioning

The state system is built around immutability and versioning for determinism and animation support.

**State Interfaces:**

```go
// ImmutableState - Read-only view of game state
type ImmutableState interface {
    ImmutableGameState() ImmutableSubState
    ImmutablePlayerStates() []ImmutablePlayerState
    ImmutableCurrentPlayer() ImmutablePlayerState
    Version() int
    Game() *Game
    // ... many more read methods
}

// State - Mutable view (only during Move.Apply())
type State interface {
    ImmutableState  // Inherits all read methods
    GameState() SubState
    PlayerStates() []PlayerState
    CurrentPlayer() PlayerState
    // ... mutable accessors
}

// SubState - Base for gameState and playerState structs
type SubState interface {
    Reader
    ReadSetter
    ReadSetConfigurer
    ContainingState() State
}
```

**Key Design Principles:**

1. **Copy-on-Write:** Each move application creates a new state version
2. **Immutable Outside Apply():** Only `Move.Apply()` gets mutable `State` interface
3. **Versioned:** Each state has a version number for animation diffing
4. **Typed Access:** Game code uses concrete types (`*gameState`, `*playerState`), framework uses interfaces

**Pattern - concreteStates() Helper:**

```go
// From examples/pig/state.go
func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
    game := state.ImmutableGameState().(*gameState)

    players := make([]*playerState, len(state.ImmutablePlayerStates()))
    for i, player := range state.ImmutablePlayerStates() {
        players[i] = player.(*playerState)
    }

    return game, players
}

// Usage in move:
func (m *moveRollDice) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    game, players := concreteStates(state)
    currentPlayer := players[game.CurrentPlayer]

    if currentPlayer.Done {
        return errors.New("current player is already done")
    }
    return nil
}
```

### Code Generation System: Eliminating Boilerplate

Since the framework predates Go generics (1.18), it uses code generation extensively to provide type-safe property access without reflection at runtime.

**The `//boardgame:codegen` Directive:**

```go
// From examples/pig/state.go
//boardgame:codegen
type gameState struct {
    base.SubState
    behaviors.CurrentPlayerBehavior
    Die         boardgame.SizedStack `sizedstack:"dice"`
    TargetScore int
}

//boardgame:codegen
type playerState struct {
    base.SubState
    Busted     bool
    Done       bool
    DieCounted bool
    RoundScore int
    TotalScore int
}
```

**Generated Code (`auto_reader.go`):**

Running `boardgame-util codegen` generates PropertyReader implementations:

```go
// Generated code provides reflection-free property access
func (g *gameState) Reader() PropertyReader {
    return &gameStateReader{g}
}

type gameStateReader struct {
    data *gameState
}

func (g *gameStateReader) Props() map[string]PropertyType {
    return map[string]PropertyType{
        "Die":         TypeSizedStack,
        "TargetScore": TypeInt,
        "CurrentPlayer": TypePlayerIndex,
    }
}

func (g *gameStateReader) Prop(name string) (interface{}, error) {
    switch name {
    case "Die":
        return g.data.Die, nil
    case "TargetScore":
        return g.data.TargetScore, nil
    case "CurrentPlayer":
        return g.data.CurrentPlayer, nil
    }
    return nil, errors.New("no such property")
}

// Similar SetProp, IntProp, BoolProp, etc. methods...
```

**Why Code Generation?**

1. **No Runtime Reflection:** Property access is direct field access after codegen
2. **Type Safety:** Generated code is type-checked by the compiler
3. **Performance:** No reflection overhead during gameplay
4. **Validation:** Struct tags are validated at generation time

**What Gets Generated:**
- `auto_reader.go` - PropertyReader/ReadSetter/ReadSetConfigurer implementations
- `auto_enum.go` - Enum constants and tree structures (if enums are defined)

###Move System: Defining Game Actions

Moves are the only way to modify game state. The framework enforces a clear separation between validation (`Legal()`) and application (`Apply()`).

**Move Interface:**

```go
type Move interface {
    Legal(state ImmutableState, proposer PlayerIndex) error
    Apply(state State) error
    ValidConfiguration(exampleState State) error  // Called at setup time
}
```

**Move Types:**

1. **Player Moves:** Proposed by players (e.g., "Roll Dice", "Play Card")
2. **Fix-Up Moves:** Auto-applied by the framework when legal (e.g., "Deal Cards", "Shuffle Discard to Draw")

**Example from Blackjack:**

```go
//boardgame:codegen
type moveShuffleDiscardToDraw struct {
    moves.FixUp  // Base type for auto-applied moves
}

func (m *moveShuffleDiscardToDraw) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := m.FixUp.Legal(state, proposer); err != nil {
        return err
    }

    game, _ := concreteStates(state)
    if game.DrawStack.Len() > 0 {
        return errors.New("The draw stack is not yet empty")
    }

    return nil
}

func (m *moveShuffleDiscardToDraw) Apply(state boardgame.State) error {
    game, _ := concreteStates(state)

    game.DiscardStack.MoveAllTo(game.DrawStack)
    game.DrawStack.Shuffle()

    return nil
}
```

**Move Configuration - The Old Way (Verbose):**

```go
var moveRollDiceConfig = boardgame.MoveConfig {
    Name: "Roll Dice",
    Constructor: func() boardgame.Move {
        return new(moveRollDice)
    },
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    return []boardgame.MoveConfig{
        &moveRollDiceConfig,
    }
}
```

**Move Configuration - The AutoConfigurer Way (Minimal):**

```go
// From moves/doc.go example
func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    auto := moves.NewAutoConfigurer(g)

    return moves.Add(
        auto.MustConfig(
            new(moves.DealComponentsUntilPlayerCountReached),
            moves.WithGameProperty("DrawStack"),
            moves.WithPlayerProperty("Hand"),
            moves.WithTargetCount(2),
        )
    )
}
```

The AutoConfigurer:
- Derives move names from struct names ("MoveDealCards" → "Deal Cards")
- Accepts configuration options (WithGameProperty, WithPlayerProperty, etc.)
- Validates configuration at setup time
- Dramatically reduces boilerplate

### Component System: Cards, Dice, Tokens

The component system enforces a critical invariant: **"One component, one location"**. A component can only exist in one stack/deck at a time.

**Key Types:**

```go
type Component interface {
    Deck() *Deck
    DeckIndex() int
    Values() ComponentValues
    DynamicValues() SubState
}

type Deck struct {
    // Collection of all components of a type
}

type Stack interface {
    // Ordered collection of components
    NumComponents() int
    ComponentAt(index int) Component
    First() Component
    MoveComponent(index int, destination Stack) error
    // ... many more methods
}
```

**Component Values vs Dynamic Values:**

- **Values:** Immutable properties set at deck creation (e.g., suit, rank of a card)
- **DynamicValues:** Mutable properties that change during gameplay (e.g., whether a piece is "kinged" in checkers)

**Stack Types:**

1. **Stack:** Variable-size ordered collection (e.g., a hand of cards)
2. **SizedStack:** Fixed-size slots (e.g., a board with 64 squares)
3. **MergedStack:** Read-only view merging multiple stacks (e.g., visible + hidden cards)

**Struct Tags for Stacks:**

```go
type gameState struct {
    base.SubState
    DrawStack    boardgame.Stack      `stack:"cards"`               // Variable-size
    Board        boardgame.SizedStack `sizedstack:"pieces,64"`      // 64 fixed slots
    AllCards     boardgame.MergedStack `overlap:"DrawStack,Discard"` // Merged view
}
```

### Property System: Structured Access

The property system provides uniform access to state fields across the framework, enabling features like sanitization, serialization, and client-side databinding.

**14 Legal Property Types:**

1. `TypeInt`
2. `TypeBool`
3. `TypeString`
4. `TypePlayerIndex`
5. `TypeEnum`
6. `TypeIntSlice`
7. `TypeBoolSlice`
8. `TypeStringSlice`
9. `TypePlayerIndexSlice`
10. `TypeStack`
11. `TypeBoard` (sized stack)
12. `TypeTimer`
13. `TypeEnum` (with specific enum type)
14. `TypeIllegal` (for computed/disallowed properties)

**PropertyReader Interface:**

```go
type PropertyReader interface {
    Props() map[string]PropertyType
    Prop(name string) (interface{}, error)
    IntProp(name string) (int, error)
    BoolProp(name string) (bool, error)
    StringProp(name string) (string, error)
    PlayerIndexProp(name string) (PlayerIndex, error)
    // ... type-specific getters
}
```

Generated code provides these implementations, allowing the framework to access properties without reflection.

### Base Package: Embeddable Defaults

The `base` package (~1,274 lines) provides default implementations that you embed in your game types to minimize boilerplate.

**base.SubState:**

```go
// Embed in gameState and playerState
type SubState struct {
    state State
}

func (s *SubState) ContainingState() State {
    return s.state
}

func (s *SubState) State() ImmutableState {
    return s.state
}
```

Provides the connection between your state struct and the containing State object.

**base.GameDelegate:**

```go
type GameDelegate struct{}

// Provides default implementations for optional GameDelegate methods
func (g *GameDelegate) LegalNumPlayers(numPlayers int) bool {
    return numPlayers >= g.MinNumPlayers() && numPlayers <= g.MaxNumPlayers()
}

func (g *GameDelegate) ConfigureConstants() PropertyCollection {
    return nil
}

func (g *GameDelegate) ConfigureAgents() []Agent {
    return nil
}

// ... many more default implementations
```

By embedding `base.GameDelegate`, you only need to implement the methods specific to your game.

### Moves Package: 30+ Reusable Move Types

The `moves` package (~27,450 lines) is a comprehensive library of common game operations. It's organized hierarchically:

**Base Types (embed these):**

- `moves.Base` - Minimal move implementation
- `moves.Default` - Base with AutoConfigurer support
- `moves.CurrentPlayer` - Only legal for current player
- `moves.FixUp` - Auto-applied when legal
- `moves.FinishTurn` - Standard turn-ending move

**Component Manipulation:**

- `DealComponents` - Deal from game stack to player stacks
- `DealComponentsUntilPlayerCountReached` - Deal until each player has N cards
- `CollectComponentsUntilCountReached` - Collect components into a stack
- `MoveComponent` - Move single component between stacks
- `ShuffleStack` - Shuffle a stack
- `SwapComponents` - Swap two components

**Turn Management:**

- `RoundRobin` - Cycle through players
- `StartPhase` - Change game phase
- `FinishTurn` - End current player's turn

**Move Progression:**

- `Serial` - Execute moves in sequence
- `Parallel` - All players take action before continuing
- `Repeat` - Repeat a set of moves N times

**Example - Blackjack's ConfigureMoves:**

```go
func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    auto := moves.NewAutoConfigurer(g)

    return moves.Combine(
        moves.AddOrderedForPhase(phaseInitialDeal,
            moves.DefaultRoundSetup(auto),
            auto.MustConfig(new(moves.DealCountComponents),
                moves.WithMoveName("Deal Initial Hidden Card"),
                moves.WithGameProperty("DrawStack"),
                moves.WithPlayerProperty("HiddenHand"),
                moves.WithCount(1)),
            auto.MustConfig(new(moves.DealCountComponents),
                moves.WithMoveName("Deal Initial Visible Card"),
                moves.WithGameProperty("DrawStack"),
                moves.WithPlayerProperty("VisibleHand"),
                moves.WithCount(1)),
            auto.MustConfig(new(moveStartNormalPlay)),
        ),
        moves.AddForPhase(phaseNormalPlay,
            auto.MustConfig(new(moveRevealHiddenCard)),
            auto.MustConfig(new(moveCurrentPlayerHit)),
            auto.MustConfig(new(moveCurrentPlayerStand)),
            auto.MustConfig(new(moveFinishTurn)),
        ),
    )
}
```

The `moves` package dramatically reduces the amount of code needed to implement common patterns.

### Behaviors Package: Composable Game Mechanics

The `behaviors` package (~575 lines) provides embeddable interfaces for common game patterns. Embed them in your state structs to instantly get functionality.

**Available Behaviors:**

1. **CurrentPlayerBehavior** - Track whose turn it is
   ```go
   type CurrentPlayerBehavior interface {
       CurrentPlayer() PlayerIndex
       SetCurrentPlayer(PlayerIndex)
   }
   ```

2. **PhaseBehavior** - Game phase tracking
   ```go
   type PhaseBehavior interface {
       Phase() int
       SetPhase(int)
   }
   ```

3. **RoundRobinBehavior** - Cycle through players
   ```go
   type RoundRobinBehavior interface {
       CurrentPlayerBehavior
       StartRoundRobin()
       RoundRobinNextPlayer() PlayerIndex
   }
   ```

4. **SeatBehavior** - Player seating (for playerState)
   ```go
   type SeatBehavior interface {
       PlayerIndex() PlayerIndex
   }
   ```

5. **InactivePlayerBehavior** - Dynamic join/leave
   ```go
   type InactivePlayerBehavior interface {
       IsActive() bool
       SetIsActive(bool)
   }
   ```

6. **ColorBehavior** - Player colors
   ```go
   type ColorBehavior interface {
       Color() string
   }
   ```

7. **RoleBehavior** - Player roles (e.g., "werewolf", "villager")

**Usage Example:**

```go
// From examples/memory/state.go
//boardgame:codegen
type gameState struct {
    base.SubState
    behaviors.CurrentPlayerBehavior  // Embed behavior
    CardSet        string
    NumCards       int
    HiddenCards    boardgame.SizedStack `sizedstack:"cards,40"`
    VisibleCards   boardgame.SizedStack `sizedstack:"cards,40"`
}
```

Now your `gameState` automatically has `CurrentPlayer()` and `SetCurrentPlayer()` methods, and the framework knows how to use them for turn management.

**Behavior + Moves Integration:**

Behaviors work seamlessly with moves from the `moves` package:

```go
// moves.CurrentPlayer checks that the proposer is the current player
// It works with any state that embeds behaviors.CurrentPlayerBehavior

//boardgame:codegen
type moveRollDice struct {
    moves.CurrentPlayer  // Only legal for current player
}

// The moves.CurrentPlayer.Legal() implementation:
func (c *CurrentPlayer) Legal(state ImmutableState, proposer PlayerIndex) error {
    currentPlayer, ok := behaviors.CurrentPlayerBehavior(state)
    if !ok {
        return errors.New("GameState doesn't have CurrentPlayerBehavior")
    }
    if currentPlayer.CurrentPlayer() != proposer {
        return errors.New("Proposer is not current player")
    }
    return nil
}
```

This is the power of the behavior + moves combination: common patterns implemented once, reused everywhere.

---

## Game Development Patterns

This section describes the patterns and conventions used when building games with the framework, based on the 6 example games (pig, memory, blackjack, checkers, tictactoe, debuganimations).

### Standard Game File Structure

All games follow a consistent structure:

```
game-name/
├── main.go           # GameDelegate implementation, ConfigureMoves, ConfigureDecks
├── state.go          # gameState and playerState struct definitions
├── moves.go          # Custom move type implementations
├── components.go     # Custom component definitions (if needed)
├── agent.go          # AI agent implementation (optional)
├── auto_reader.go    # AUTO-GENERATED by //boardgame:codegen
├── auto_enum.go      # AUTO-GENERATED if enums are defined
├── *_test.go         # Unit tests
└── client/           # Web components for rendering
    ├── boardgame-render-game-{name}.js        # Main game renderer
    └── boardgame-render-player-info-{name}.js # Player info renderer
```

**Key Files:**

- **main.go** - The game's entry point and configuration hub
- **state.go** - Defines what information the game tracks
- **moves.go** - Defines how players interact with the game
- **client/** - Defines how the game looks in the browser

### Anatomy of main.go

The `main.go` file is where your `gameDelegate` lives and is configured.

**Complete Example from Pig:**

```go
package pig

import (
    "github.com/jkomoros/boardgame"
    "github.com/jkomoros/boardgame/base"
    "github.com/jkomoros/boardgame/components/dice"
    "github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

type gameDelegate struct {
    base.GameDelegate
}

func (g *gameDelegate) Name() string {
    // Reflection-based package name matching
    // (boilerplate pattern used in all examples)
    return "pig"
}

func (g *gameDelegate) DisplayName() string {
    return "Pig"
}

func (g *gameDelegate) Description() string {
    return "Players roll the dice, collecting points, but bust if they roll a one."
}

func (g *gameDelegate) MinNumPlayers() int { return 2 }
func (g *gameDelegate) MaxNumPlayers() int { return 6 }
func (g *gameDelegate) DefaultNumPlayers() int { return 2 }

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
    return &gameState{}
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
    return &playerState{}
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
    return map[string]*boardgame.Deck{
        "dice": dice.NewDeck(false),  // One standard 6-sided die
    }
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    auto := moves.NewAutoConfigurer(g)
    return moves.Add(
        auto.MustConfig(new(moveRollDice)),
        auto.MustConfig(new(moveCountDie), moves.WithIsFixUp(true)),
        auto.MustConfig(new(moveFinishTurn)),
    )
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
    game, _ := concreteStates(state)
    return game.Die, nil  // All components (the die) go to the Die stack
}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
    game, _ := concreteStates(state)

    // Pick a random starting player
    startingPlayer := boardgame.PlayerIndex(state.Rand().Intn(len(state.PlayerStates())))
    game.SetCurrentPlayer(startingPlayer)

    game.TargetScore = 100

    return nil
}

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
    game, players := concreteStates(state)

    for _, player := range players {
        if player.TotalScore >= game.TargetScore {
            return true
        }
    }

    return false
}

func (g *gameDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
    return pState.(*playerState).TotalScore
}
```

**Common Patterns:**
- Embed `base.GameDelegate` for default implementations
- Use reflection to derive Name() from package name
- Use `ConfigureMoves()` with AutoConfigurer
- Implement lifecycle hooks (DistributeComponentToStarterStack, FinishSetUp, GameEndConditionMet)

### Anatomy of state.go

The `state.go` file defines what information your game tracks.

**Example from Pig (Simple):**

```go
package pig

import (
    "github.com/jkomoros/boardgame"
    "github.com/jkomoros/boardgame/base"
    "github.com/jkomoros/boardgame/behaviors"
)

//boardgame:codegen
type gameState struct {
    base.SubState
    behaviors.CurrentPlayerBehavior
    Die         boardgame.SizedStack `sizedstack:"dice"`
    TargetScore int
}

//boardgame:codegen
type playerState struct {
    base.SubState
    Busted     bool
    Done       bool
    DieCounted bool
    RoundScore int
    TotalScore int
}

// Helper to convert interface types to concrete types
func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
    game := state.ImmutableGameState().(*gameState)

    players := make([]*playerState, len(state.ImmutablePlayerStates()))
    for i, player := range state.ImmutablePlayerStates() {
        players[i] = player.(*playerState)
    }

    return game, players
}

// TurnDone - called by moves.FinishTurn to check if turn can end
func (p *playerState) TurnDone() error {
    if !p.DieCounted {
        return errors.New("the most recent die roll has not been counted")
    }

    if !p.Busted && !p.Done {
        return errors.New("they have not either busted or signaled that they are done")
    }

    return nil
}

// ResetForTurnStart - called by moves.FinishTurn at turn start
func (p *playerState) ResetForTurnStart() error {
    p.Done = false
    p.Busted = false
    p.RoundScore = 0
    p.DieCounted = true
    return nil
}

// ResetForTurnEnd - called by moves.FinishTurn at turn end
func (p *playerState) ResetForTurnEnd() error {
    if p.Done {
        p.TotalScore += p.RoundScore
    }
    p.ResetForTurn()
    return nil
}
```

**Example from Memory (Medium Complexity):**

```go
//boardgame:codegen
type gameState struct {
    base.SubState
    behaviors.CurrentPlayerBehavior
    CardSet        string
    NumCards       int
    HiddenCards    boardgame.SizedStack  `sizedstack:"cards,40" sanitize:"order"`
    VisibleCards   boardgame.SizedStack  `sizedstack:"cards,40"`
    Cards          boardgame.MergedStack `overlap:"VisibleCards,HiddenCards"`
    HideCardsTimer boardgame.Timer
    UnusedCards    boardgame.Stack `stack:"cards"`
}

//boardgame:codegen
type playerState struct {
    base.SubState
    CardsLeftToReveal int
    WonCards          boardgame.Stack `stack:"cards"`
}
```

**Key Patterns:**

1. **Always Embed base.SubState**
   - Provides connection to containing State

2. **Embed Behaviors for Common Patterns**
   - `behaviors.CurrentPlayerBehavior` - Track current player
   - `behaviors.PhaseBehavior` - Track game phases
   - `behaviors.RoundRobinBehavior` - Cycling turn order

3. **Struct Tags Configure Stacks**
   - `stack:"deckname"` - Variable-size stack
   - `sizedstack:"deckname,size"` - Fixed-size stack
   - `overlap:"stack1,stack2"` - Merged view
   - `sanitize:"policy"` - Hide information from non-owners

4. **Add //boardgame:codegen Directive**
   - Triggers auto_reader.go generation

5. **Implement Interfaces for Move Integration**
   - `TurnDone()` error - For moves.FinishTurn
   - `ResetForTurnStart()` error - For moves.FinishTurn
   - `ResetForTurnEnd()` error - For moves.FinishTurn

6. **Create concreteStates() Helper**
   - Converts interfaces to concrete types
   - Used throughout moves and delegate

### Anatomy of moves.go

The `moves.go` file defines how players interact with the game.

**Pattern 1: Simple Player Move**

```go
//boardgame:codegen
type moveRollDice struct {
    moves.CurrentPlayer  // Only legal for current player
}

func (m *moveRollDice) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    // Call parent Legal() first
    if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
        return err
    }

    game, players := concreteStates(state)
    currentPlayer := players[game.CurrentPlayer]

    if currentPlayer.Done {
        return errors.New("current player is already done")
    }

    if !currentPlayer.DieCounted {
        return errors.New("the current die hasn't been counted yet")
    }

    return nil
}

func (m *moveRollDice) Apply(state boardgame.State) error {
    game, players := concreteStates(state)
    currentPlayer := players[game.CurrentPlayer]

    // Roll the die
    game.Die.First().MoveToNextSlot()
    currentPlayer.DieCounted = false

    return nil
}
```

**Pattern 2: Fix-Up Move (Auto-Applied)**

```go
//boardgame:codegen
type moveCountDie struct {
    moves.FixUp  // Automatically applied when legal
}

func (m *moveCountDie) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := m.FixUp.Legal(state, proposer); err != nil {
        return err
    }

    game, players := concreteStates(state)
    currentPlayer := players[game.CurrentPlayer]

    if currentPlayer.DieCounted {
        return errors.New("die is already counted")
    }

    return nil
}

func (m *moveCountDie) Apply(state boardgame.State) error {
    game, players := concreteStates(state)
    currentPlayer := players[game.CurrentPlayer]

    dieValue := game.Die.ComponentAt(0).DynamicValues().(*dice.DynamicValue).Value

    if dieValue == 1 {
        currentPlayer.Busted = true
    } else {
        currentPlayer.RoundScore += dieValue
    }

    currentPlayer.DieCounted = true

    return nil
}
```

**Pattern 3: Using AutoConfigurer for Complex Moves**

```go
// From blackjack - no custom move implementation needed!
func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
    auto := moves.NewAutoConfigurer(g)

    return moves.Add(
        auto.MustConfig(
            new(moves.DealCountComponents),
            moves.WithMoveName("Deal Initial Hidden Card"),
            moves.WithGameProperty("DrawStack"),
            moves.WithPlayerProperty("HiddenHand"),
            moves.WithCount(1),
        ),
        auto.MustConfig(
            new(moves.DealCountComponents),
            moves.WithMoveName("Deal Initial Visible Card"),
            moves.WithGameProperty("DrawStack"),
            moves.WithPlayerProperty("VisibleHand"),
            moves.WithCount(1),
        ),
    )
}
```

**Move Implementation Guidelines:**

1. **Always call parent Legal() first**
   - Ensures base validation is performed

2. **Use concreteStates() for type safety**
   - Easier than type assertions everywhere

3. **Legal() uses ImmutableState**
   - Cannot modify state
   - Pure validation

4. **Apply() uses mutable State**
   - Modify game state here
   - Keep it simple and deterministic

5. **Return descriptive errors from Legal()**
   - Error messages shown to players

6. **Use //boardgame:codegen**
   - Generates PropertyReader for move if needed

### Lifecycle Hooks

The GameDelegate has several hooks called at specific times during game setup and play.

**1. BeginSetUp(state State, variant Variant) error**

Called before component distribution. Use for:
- Initializing variant-specific configuration
- Setting up sized stacks with correct counts

```go
// From memory game
func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
    game, _ := concreteStates(state)

    game.CardSet = variant["cardset"]
    game.NumCards = getNumCards(variant["numcards"])

    return nil
}
```

**2. DistributeComponentToStarterStack(state ImmutableState, c Component) (ImmutableStack, error)**

Called for each component during setup. Returns the stack where the component should go.

```go
// From blackjack - route components to appropriate stacks
func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
    game, _ := concreteStates(state)

    if c.Deck().Name() == "cards" {
        // Playing cards go to draw stack
        return game.DrawStack, nil
    }

    return nil, errors.New("Unknown deck: " + c.Deck().Name())
}
```

**3. FinishSetUp(state State) error**

Called after all components are distributed. Use for:
- Shuffling decks
- Selecting starting player
- Initializing game state values

```go
// From pig
func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
    game, _ := concreteStates(state)

    // Pick random starting player
    startingPlayer := boardgame.PlayerIndex(state.Rand().Intn(len(state.PlayerStates())))
    game.SetCurrentPlayer(startingPlayer)

    game.TargetScore = defaultTargetScore

    return nil
}
```

**4. GameEndConditionMet(state ImmutableState) bool**

Called after every move to check if the game is over.

```go
func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
    game, players := concreteStates(state)

    for _, player := range players {
        if player.TotalScore >= game.TargetScore {
            return true
        }
    }

    return false
}
```

**5. PlayerScore(pState ImmutableSubState) int**

Returns the score for a player (used for ranking when game ends).

```go
func (g *gameDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
    return pState.(*playerState).TotalScore
}
```

**6. CheckGameFinished(state ImmutableState) (bool, []PlayerIndex)**

Alternative to GameEndConditionMet() - returns both whether game is finished AND who won.

```go
func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (bool, []boardgame.PlayerIndex) {
    // Check if game is finished and determine winners
    // Return (true, []PlayerIndex{0, 2}) for a tie between players 0 and 2
}
```

### Example Complexity Levels

The framework supports games of varying complexity. Here's how the examples compare:

**Simple: Pig (Dice Game)**
- **Lines of Code:** ~500 total
- **State:** 7 properties (gameState: 2, playerState: 5)
- **Moves:** 3 moves (RollDice, CountDie, FinishTurn)
- **Components:** 1 die
- **Behaviors:** CurrentPlayerBehavior
- **Phases:** None (single phase)
- **Complexity:** Perfect starting point for learning

**Medium: Memory (Card Matching)**
- **Lines of Code:** ~1,000 total
- **State:** 12 properties, including Timer
- **Moves:** 4 moves + AutoConfigured moves
- **Components:** 40 cards with custom values
- **Behaviors:** CurrentPlayerBehavior
- **Features:** Timers, sanitization (hidden cards), variants, AI agent
- **Complexity:** Good next step, introduces important concepts

**Medium: Blackjack**
- **Lines of Code:** ~1,200 total
- **State:** 8 properties across game/player state
- **Moves:** 5 custom + many AutoConfigured
- **Components:** Standard 52-card deck
- **Behaviors:** RoundRobinBehavior, CurrentPlayerBehavior, PhaseBehavior, InactivePlayerBehavior, SeatBehavior
- **Phases:** 2 (InitialDeal, NormalPlay)
- **Features:** Multiple phases, dynamic player joining, computed properties (hand values)
- **Complexity:** Shows phase-based game structure

**Complex: Checkers**
- **Lines of Code:** ~2,000 total
- **State:** Custom board representation
- **Moves:** Complex move validation (legal checker moves)
- **Components:** Dynamic component values (crowned pieces)
- **Behaviors:** ColorBehavior for player colors
- **Features:** Board game mechanics, complex move validation
- **Complexity:** Advanced patterns, but still compact

**Testing: DebugAnimations**
- **Lines of Code:** ~7,000 (largest)
- **Purpose:** Test and demonstrate animation features
- **Not a real game:** Designed to exercise animation system
- **Features:** All animation types, stress testing

### Client-Side Web Component Structure

Each game provides custom renderers in the `client/` directory.

**Naming Convention:**
```
client/
├── boardgame-render-game-{gamename}.js
└── boardgame-render-player-info-{gamename}.js
```

**Example from Pig:**

```javascript
import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';

class BoardgameRenderGamePig extends BoardgameBaseGameRenderer {
    static get is() {
        return "boardgame-render-game-pig"
    }

    static get template() {
        return html`
            <style>
                /* Custom styles */
            </style>

            <!-- Data binding to game state -->
            <div>Target Score: [[state.Game.TargetScore]]</div>
            <div>Current Player: [[state.Game.CurrentPlayer]]</div>

            <!-- Render die -->
            <boardgame-component-stack layout="pile" messy="{{}}">
                <boardgame-component id="die-0" item="[[state.Game.Die.Components.0]]">
                </boardgame-component>
            </boardgame-component-stack>

            <!-- Player actions -->
            <button propose-move="Roll Dice">Roll Dice</button>
            <button propose-move="Finish Turn">Finish Turn</button>
        `;
    }
}

customElements.define(BoardgameRenderGamePig.is, BoardgameRenderGamePig);
```

**Key Patterns:**

1. **Extend BoardgameBaseGameRenderer**
   - Provides base functionality (move proposing, event handling)

2. **Access State via `[[state.Game.*]]` and `[[state.Players.*]]`**
   - Polymer databinding syntax
   - Automatically updates when state changes

3. **Use `propose-move` Attribute for Buttons**
   - `<button propose-move="MoveName">` triggers move

4. **Use boardgame-component for Game Pieces**
   - `<boardgame-component item="[[state.Game.Die.Components.0]]">`
   - Automatically animates when component moves

5. **Optional: Override Animation Timing**
   ```javascript
   delayAnimation(fromMove, toMove) {
       if (fromMove && fromMove.Name == "Reveal Card") {
           return 1000;  // 1 second delay
       }
       return 0;
   }
   ```

---

## Storage Architecture

The storage layer provides persistence for games, allowing them to be saved, loaded, and resumed. The framework defines a `StorageManager` interface and provides multiple backend implementations.

### StorageManager Interface

```go
type StorageManager interface {
    // Game operations
    State(gameID string, version int) (StateStorageRecord, error)
    Moves(gameID string, fromVersion, toVersion int) ([]*MoveStorageRecord, error)
    Game(gameID string) (*GameStorageRecord, error)
    SaveGameAndCurrentState(game *GameStorageRecord, state StateStorageRecord, move *MoveStorageRecord) error

    // Player operations
    SetPlayerForGame(gameID string, playerIndex PlayerIndex, userID string) error
    UpdatePlayer(gameID string, playerIndex PlayerIndex, userID string) error

    // Listing and querying
    ListGames(max int, since int64, excludeIDs map[string]bool) []*GameStorageRecord
    UserIdsForGame(gameID string) []string

    // Agents
    AgentState(gameID string, player PlayerIndex) ([]byte, error)
    SaveAgentState(gameID string, player PlayerIndex, state []byte) error

    // Extended info (optional)
    ExtendedGame(gameID string) (*ExtendedGameStorageRecord, error)

    // Lifecycle
    Connect(config string) error
    Close()
    Name() string
}
```

**Key Concepts:**

1. **State Versioning:** Each move application creates a new state version
2. **Atomic Saves:** `SaveGameAndCurrentState()` saves game + state + move atomically
3. **Player Association:** Track which user controls which player seat
4. **Agent Persistence:** AI agent state persisted as opaque JSON blobs

### Available Storage Backends

#### 1. Memory Storage

**Location:** `storage/memory/`

**Use Cases:**
- Testing
- Development
- Demos
- Games that don't need persistence

**Characteristics:**
- In-memory only (lost on restart)
- Fast
- No configuration needed
- No dependencies

**Usage:**
```go
import "github.com/jkomoros/boardgame/storage/memory"

storage := memory.NewStorageManager()
```

#### 2. Filesystem Storage

**Location:** `storage/filesystem/`

**Use Cases:**
- Single-user games
- Development
- Simple deployments

**Characteristics:**
- Stores each game as directory with JSON files
- Human-readable (useful for debugging)
- No database dependencies
- Not suitable for high concurrency

**Structure:**
```
games/
└── {gameID}/
    ├── game.json
    ├── state_000001.json
    ├── state_000002.json
    ├── move_000001.json
    └── ...
```

**Usage:**
```go
import "github.com/jkomoros/boardgame/storage/filesystem"

storage := filesystem.NewStorageManager("./games")
```

#### 3. Bolt Storage

**Location:** `storage/bolt/`

**Use Cases:**
- Single-server deployments
- Embedded applications
- SQLite-like use cases

**Characteristics:**
- BoltDB embedded key-value store
- Single file database
- ACID transactions
- Good performance for single server
- Cannot scale horizontally

**Usage:**
```go
import "github.com/jkomoros/boardgame/storage/bolt"

storage, err := bolt.NewStorageManager("./boardgame.db")
```

#### 4. MySQL Storage

**Location:** `storage/mysql/`

**Use Cases:**
- Production deployments
- Multi-server setups
- High concurrency
- Need for SQL queries

**Characteristics:**
- Full relational database
- Horizontal scaling
- ACID transactions
- Supports replication
- Requires MySQL server

**Schema:**
```sql
-- Games table
CREATE TABLE games (
    id VARCHAR(16) PRIMARY KEY,
    name VARCHAR(64),
    version INT,
    winners TEXT,
    finished BOOLEAN,
    num_players INT,
    agents TEXT,
    created TIMESTAMP,
    modified TIMESTAMP
);

-- States table (one per version)
CREATE TABLE states (
    game_id VARCHAR(16),
    version INT,
    blob MEDIUMBLOB,
    PRIMARY KEY (game_id, version)
);

-- Moves table
CREATE TABLE moves (
    game_id VARCHAR(16),
    version INT,
    initiator INT,
    timestamp BIGINT,
    blob MEDIUMBLOB,
    PRIMARY KEY (game_id, version)
);
```

**Usage:**
```go
import "github.com/jkomoros/boardgame/storage/mysql"

config := &mysql.Config{
    Host:     "localhost",
    Port:     3306,
    User:     "boardgame",
    Password: "password",
    Database: "boardgame",
}

storage, err := mysql.NewStorageManager(config.ConnectString())
```

**See also:** `storage/mysql/README.md` for schema setup instructions

### Storage Configuration via config.json

Games are configured via a `config.json` file in the game directory:

```json
{
    "DefaultStorageType": "mysql",
    "StorageConfig": {
        "mysql": "user:password@tcp(localhost:3306)/boardgame"
    },
    "ServerConfig": {
        "ApiHost": "http://localhost:8080",
        "DefaultNumPlayers": 2,
        "AllowedOrigins": ["http://localhost:8080"]
    },
    "Firebase": {
        "ApiKey": "...",
        "AuthDomain": "...",
        "DatabaseURL": "...",
        "ProjectId": "...",
        "StorageBucket": "...",
        "MessagingSenderId": "..."
    }
}
```

---

## Server Architecture

The server package (`server/`) provides a web server with REST API, WebSocket support, authentication, and static file serving for the web frontend.

### Server Components

```
server/
├── api/
│   ├── main.go          # Server setup and routing
│   ├── listing.go       # Game listing endpoints
│   ├── moves.go         # Move proposal endpoint
│   ├── state.go         # State retrieval endpoints
│   ├── manager.go       # GameManager management
│   ├── auth.go          # Firebase authentication
│   ├── users.go         # User management
│   ├── websockets.go    # WebSocket connections
│   └── cors.go          # CORS handling
└── static/
    ├── index.html       # App entry point
    ├── manifest.json    # PWA manifest
    ├── service-worker.js # Service worker for offline
    └── src/             # Frontend source code
```

### REST API Structure

**Base Path:** `/api/`

**Game Management:**
- `POST /api/new` - Create new game
  - Body: `{name: "blackjack", numPlayers: 2, agents: [], variant: {}}`
  - Returns: `{GameID: "abc123"}`

- `GET /api/list/{list-type}/{user-id}` - List games
  - list-type: `participating`, `visible`, `finished`
  - Returns: Array of game info

**Game State:**
- `GET /api/game/{gameName}/{gameId}` - Get game info
  - Returns: Game metadata, player names, etc.

- `GET /api/game/{gameName}/{gameId}/state/{version}` - Get specific state version
  - Returns: State JSON

- `GET /api/game/{gameName}/{gameId}/state/{fromVersion}/to/{toVersion}` - Get state range
  - Returns: Array of states for animation

**Move Proposal:**
- `POST /api/game/{gameName}/{gameId}/move` - Propose move
  - Body: `{Name: "Roll Dice", PlayerIndex: 0}`
  - Returns: `{Success: true}` or error

**WebSocket:**
- `GET /api/game/{gameName}/{gameId}/socket` - WebSocket upgrade
  - Server pushes version numbers when state changes
  - Client fetches new states via REST

### Authentication Flow

The server integrates with Firebase Authentication:

```
1. User signs in via Firebase (client-side)
2. Client receives ID token
3. Client includes token in requests: Authorization: Bearer {token}
4. Server validates token with Firebase
5. Server associates user ID with player seat
```

**auth.go:**
```go
func getEffectiveUser(r *http.Request, users UserManager) string {
    // Extract Authorization header
    authHeader := r.Header.Get("Authorization")

    // Parse "Bearer {token}"
    token := strings.TrimPrefix(authHeader, "Bearer ")

    // Validate with Firebase
    decodedToken, err := firebaseApp.Auth().VerifyIDToken(token)

    // Return user ID
    return decodedToken.UID
}
```

### WebSocket Implementation

WebSockets provide real-time updates when game state changes.

**Flow:**

1. **Client connects:** `ws://host/api/game/{name}/{id}/socket`
2. **Server tracks connection:** Maintains map of gameID → []connections
3. **Move applied:** State version increments
4. **Server broadcasts:** Sends version number to all connected clients
5. **Clients fetch:** Use REST API to fetch new states

**websockets.go Pattern:**
```go
type socketManager struct {
    connections map[string][]*websocket.Conn
    lock        sync.RWMutex
}

func (s *socketManager) AddConnection(gameID string, conn *websocket.Conn) {
    s.lock.Lock()
    defer s.lock.Unlock()

    s.connections[gameID] = append(s.connections[gameID], conn)
}

func (s *socketManager) NotifyGame(gameID string, version int) {
    s.lock.RLock()
    conns := s.connections[gameID]
    s.lock.RUnlock()

    message := fmt.Sprintf("%d", version)
    for _, conn := range conns {
        conn.WriteMessage(websocket.TextMessage, []byte(message))
    }
}
```

**Why Not Push Full State?**

- States can be large (especially with many components)
- Clients might already have that state cached
- Separates notification from data transfer
- Allows clients to batch fetch multiple states

### CORS Configuration

For development, the server needs CORS configuration to allow frontend on different port:

```go
// From cors.go
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")

        // Check if origin is allowed
        if isAllowedOrigin(origin, allowedOrigins) {
            c.Header("Access-Control-Allow-Origin", origin)
            c.Header("Access-Control-Allow-Credentials", "true")
            c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
        }

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

**Configured via config.json:**
```json
{
    "ServerConfig": {
        "AllowedOrigins": ["http://localhost:8080", "https://yourdomain.com"]
    }
}
```

### Server Startup

**Using boardgame-util:**
```bash
# Development server (with hot reload)
boardgame-util serve

# Production build
boardgame-util build
./server
```

**Programmatic:**
```go
import (
    "github.com/jkomoros/boardgame/server/api"
)

func main() {
    config := api.LoadConfig("config.json")

    // Create GameManager for each game type
    managers := make(map[string]*boardgame.GameManager)
    managers["pig"] = pig.NewManager(config.Storage)
    managers["memory"] = memory.NewManager(config.Storage)

    // Start server
    server := api.NewServer(config, managers)
    server.Start(":8080")
}
```

---

## Web Frontend Architecture

The web frontend is a Progressive Web App (PWA) built with Polymer 3 and lit-element, using Redux for state management and a sophisticated FLIP animation system.

### Technology Stack

**Framework & UI:**
- **Polymer 3.3.0** - 20 components (game rendering, animations, UI)
- **lit-element 2.3.1** - 6 components (app shell, main views)
  - Note: Uses old `@polymer/lit-element 0.7.1`, not modern `lit`
- **Web Components** - Standard custom elements
- **@webcomponents/webcomponentsjs** - Polyfills for older browsers

**State Management:**
- **Redux 4.0.5** - Central state store
- **Redux Thunk 2.3.0** - Async actions
- **reselect-tools 0.0.7** - Memoized selectors

**Backend Integration:**
- **Firebase 5.11.1** - Authentication
- **iron-ajax** - HTTP requests (Polymer element)
- **WebSocket** - Real-time updates

**Build & PWA:**
- **NO BUNDLER** - Native ES modules with imports
- **Service Worker** - Offline support
- **Web App Manifest** - Installable PWA

**Animation:**
- **Web Animations API** - Smooth animations
- **CSS Transitions** - Transform-based animations
- **web-animations-js** - Polyfill

### Frontend Architecture Overview

```
index.html (entry point)
    ↓
boardgame-app (app shell, lit-element)
    ├── Redux Store (state management)
    ├── Routing (URL → page)
    └── Views
        ├── boardgame-list-games-view (lit-element)
        └── boardgame-game-view (lit-element)
            ├── boardgame-game-state-manager (Polymer)
            │   ├── HTTP: Fetch game info & states
            │   ├── WebSocket: Listen for updates
            │   └── Queue state bundles for animation
            ├── boardgame-render-game (Polymer)
            │   ├── boardgame-component-animator (Polymer)
            │   │   └── FLIP Animation System
            │   └── Game-specific renderer (Polymer)
            │       └── boardgame-component (Polymer)
            │           ├── boardgame-card (Polymer)
            │           └── boardgame-token (Polymer)
            └── boardgame-player-roster (Polymer)
```

### Component Breakdown

**Lit Element Components (6)** - Modern, lightweight:
1. **boardgame-app** - App shell, routing, Redux connection
2. **boardgame-game-view** - Main game page
3. **boardgame-list-games-view** - Game list page
4. **boardgame-create-game** - Game creation dialog
5. **boardgame-user** - User info component
6. **shared-styles-lit** - Shared styles

**Polymer Components (20)** - Legacy but functional:

*Animation System (Critical):*
1. **boardgame-component-animator** - FLIP animation coordinator
2. **boardgame-animatable-item** - Base class for animated elements
3. **boardgame-component** - Base game component (cards, tokens)
4. **boardgame-card** - Card with flip/rotation animations
5. **boardgame-token** - Simple game token

*Game Infrastructure:*
6. **boardgame-render-game** - Dynamic game renderer loader
7. **boardgame-game-state-manager** - State fetching & queuing
8. **boardgame-base-game-renderer** - Base class for game renderers
9. **boardgame-component-stack** - Container for components (grid, stack, fan, pile layouts)

*UI Components:*
10. **boardgame-admin-controls** - Admin panel
11. **boardgame-player-roster** - Player list
12. **boardgame-player-chip** - Individual player display
13. **boardgame-render-player-info** - Player info renderer
14. **boardgame-status-text** - Game status messages
15. **boardgame-fading-text** - Animated text effects
16. **boardgame-deck-defaults** - Default component templates

*Utilities:*
17. **boardgame-ajax** - HTTP request wrapper (extends iron-ajax)
18. **game-path-mixin** - URL routing helper

### Redux State Structure

**Store Configuration** (`src/store.js`):
```javascript
import { createStore, compose, applyMiddleware, combineReducers } from 'redux';
import thunk from 'redux-thunk';
import { lazyReducerEnhancer } from 'pwa-helpers/lazy-reducer-enhancer.js';

const store = createStore(
    (state, action) => state,  // Initial reducer
    compose(
        lazyReducerEnhancer(combineReducers),
        applyMiddleware(thunk),
        window.__REDUX_DEVTOOLS_EXTENSION__ ? window.__REDUX_DEVTOOLS_EXTENSION__() : f => f
    )
);

// Register reducers
store.addReducers({
    app,    // UI state
    user,   // Authentication
});

// Lazy-load game/list reducers when needed
```

**State Shape:**
```javascript
{
    app: {
        page: "game",
        pageExtra: "blackjack/abc123",
        offline: false,
        snackbarOpened: false,
        headerPanelOpen: false
    },
    user: {
        loggedIn: true,
        user: { uid: "...", email: "..." },
        admin: false
    },
    game: {
        id: "abc123",
        name: "blackjack",
        currentState: {  // Expanded state
            Game: { ... },
            Players: [ ... ]
        },
        chest: { ... },  // Component definitions
        playersInfo: [ ... ]
    },
    list: {
        managers: [ ... ],  // Available game types
        participatingGames: [ ... ],
        visibleGames: [ ... ]
    }
}
```

### Animation System Architecture

> **See [server/static/src/ARCHITECTURE.md](server/static/src/ARCHITECTURE.md) for detailed animation system documentation.**

**High-Level Overview:**

The animation system uses the FLIP (First, Last, Invert, Play) technique to create smooth animations when components move between states.

**FLIP Steps:**

1. **First:** Record initial position of all components
2. **Last:** New state is databound (components jump to new positions)
3. **Invert:** Calculate transforms to visually move components back to old positions
4. **Play:** Animate transforms to 0, creating smooth transitions

**Implemented in boardgame-component-animator.js:**

```javascript
// Simplified algorithm
class BoardgameComponentAnimator {
    prepareAnimations(nextState) {
        // FIRST: Record current positions
        this.recordPositions();
    }

    animateChanges(nextState) {
        // LAST: New state databound (components jump)
        this.installState(nextState);

        // INVERT: Calculate inverse transforms
        for (let component of this.components) {
            let oldPos = component.oldPosition;
            let newPos = component.getBoundingClientRect();
            let dx = oldPos.left - newPos.left;
            let dy = oldPos.top - newPos.top;
            component.style.transform = `translate(${dx}px, ${dy}px)`;
        }

        // PLAY: Animate to identity transform
        requestAnimationFrame(() => {
            for (let component of this.components) {
                component.style.transform = 'translate(0, 0)';
                // CSS transition handles the animation
            }
        });
    }
}
```

**Key Features:**

- **Component Tracking:** Components identified by `id` across states
- **Faux Components:** When component appears/disappears, create temporary element for animation
- **Sanitized Stacks:** Animate to/from stack center with opacity fade
- **Card Flips:** Internal animations (face up/down, rotation) separate from movement
- **Animation Timing:** Controlled via CSS `--animation-length` custom property

### Communication: WebSocket + REST

**Pattern: WebSocket for Notification, REST for Data**

**1. Initial Load:**
```javascript
// Fetch game info
fetch(`/api/game/${gameName}/${gameId}`)
    .then(r => r.json())
    .then(info => {
        // Game metadata, player names, etc.
        dispatch(updateGameInfo(info));
    });

// Fetch current state
fetch(`/api/game/${gameName}/${gameId}/state/0`)
    .then(r => r.json())
    .then(state => {
        dispatch(installGameState(state));
    });
```

**2. Open WebSocket:**
```javascript
const socket = new WebSocket(`ws://${host}/api/game/${gameName}/${gameId}/socket`);

socket.onmessage = (event) => {
    const newVersion = parseInt(event.data);
    dispatch(fetchStateRange(currentVersion, newVersion));
};
```

**3. Fetch State Range:**
```javascript
// Fetch all states between versions for smooth animation
fetch(`/api/game/${gameName}/${gameId}/state/${fromVersion}/to/${toVersion}`)
    .then(r => r.json())
    .then(states => {
        // Queue states for sequential animation
        states.forEach(state => queueStateBundle(state));
    });
```

**4. State Expansion:**

Before installing in Redux, states are expanded:
```javascript
function expandState(state, chest) {
    // Expand component indexes to full objects
    for (let stack of state.Game.AllStacks) {
        stack.Components = stack.ComponentIndexes.map(i => chest.components[i]);
    }

    // Expand timers
    for (let timer of state.Timers) {
        timer.timeLeft = calculateTimeLeft(timer);
    }

    return state;
}
```

### Progressive Web App Features

**Service Worker (`service-worker.js`):**
- Caches static assets
- Offline support for app shell
- Background sync for moves

**Web App Manifest (`manifest.json`):**
```json
{
    "name": "Boardgame",
    "short_name": "Boardgame",
    "start_url": "/",
    "display": "standalone",
    "background_color": "#ffffff",
    "theme_color": "#3f51b5",
    "icons": [ ... ]
}
```

**Installation:**
- Add to home screen on mobile
- Standalone window on desktop
- Offline gameplay (with limitations)

---

## Development Tools

The `boardgame-util` CLI is a Swiss-army-knife tool for developing games with the framework.

### Installation

```bash
go install github.com/jkomoros/boardgame/boardgame-util@latest
```

### Primary Commands

**1. codegen - Generate Boilerplate**

```bash
# In your game directory
boardgame-util codegen
```

Generates:
- `auto_reader.go` - PropertyReader implementations for state structs
- `auto_enum.go` - Enum definitions (if enums are declared)

Looks for `//boardgame:codegen` directives in `.go` files.

**Alternative:** Use `//go:generate boardgame-util codegen` in your source file, then run `go generate`.

**2. serve - Development Server**

```bash
# Start dev server with hot reload
boardgame-util serve [game-package-path]

# Example
cd examples/pig
boardgame-util serve
```

Features:
- Automatic reload on Go source changes
- Serves static frontend
- Opens browser automatically
- Uses memory storage by default
- Reads `config.json` for configuration

**3. build - Production Build**

```bash
# Build server binary and prepare static assets
boardgame-util build [output-dir]

# Example
boardgame-util build ./dist
```

Output:
```
dist/
├── server          # Compiled binary
└── static/         # Frontend assets
```

**4. db - Database Management**

```bash
# Setup database schema
boardgame-util db setup

# Check current schema version
boardgame-util db version

# Upgrade schema to latest
boardgame-util db upgrade
```

Supports:
- MySQL schema creation
- Schema migrations
- Version tracking

**5. stub - Generate Game Starter**

```bash
# Create new game from template
boardgame-util stub [game-name]

# Example
boardgame-util stub mygame
```

Generates:
```
mygame/
├── main.go
├── state.go
├── moves.go
└── client/
    └── boardgame-render-game-mygame.js
```

**6. golden - Test Golden Files**

```bash
# Update golden test files
boardgame-util golden -update

# Run golden tests
boardgame-util golden
```

For testing that code generation produces expected output.

**7. config - Configuration Management**

```bash
# Validate config.json
boardgame-util config validate

# Print merged configuration
boardgame-util config print
```

### boardgame-util Library

The CLI is built on `boardgame-util/lib/` which provides:

- **codegen/** - Code generation library
  - Struct parsing
  - Template rendering
  - PropertyReader generation
  - Enum generation

- **stub/** - Project templating
  - Game starter templates
  - File generation

- **config/** - Configuration management
  - Config file parsing
  - Environment variable merging

### Code Generation Workflow

**Typical Development Cycle:**

1. Define your state structs with `//boardgame:codegen`
   ```go
   //boardgame:codegen
   type gameState struct {
       base.SubState
       DrawStack boardgame.Stack `stack:"cards"`
   }
   ```

2. Run codegen
   ```bash
   boardgame-util codegen
   ```

3. Generated `auto_reader.go` provides PropertyReader

4. Compile and test
   ```bash
   go test ./...
   ```

5. Iterate

**What Gets Generated:**

```go
// In auto_reader.go

func (g *gameState) Reader() PropertyReader {
    return &gameStateReader{g}
}

type gameStateReader struct {
    data *gameState
}

func (g *gameStateReader) Props() map[string]PropertyType {
    return map[string]PropertyType{
        "DrawStack": TypeStack,
    }
}

func (g *gameStateReader) Prop(name string) (interface{}, error) {
    switch name {
    case "DrawStack":
        return g.data.DrawStack, nil
    }
    return nil, errors.New("no such property")
}

func (g *gameStateReader) SetProp(name string, value interface{}) error {
    switch name {
    case "DrawStack":
        val, ok := value.(Stack)
        if !ok {
            return errors.New("invalid type")
        }
        g.data.DrawStack = val
        return nil
    }
    return errors.New("no such property")
}

// ... many more type-specific getters/setters
```

---

## Key Design Patterns & Strengths

This section highlights what makes the boardgame framework well-designed and effective.

### Major Strengths

**1. Type Safety**

Despite predating Go generics, the framework achieves strong type safety through:
- Minimal use of `interface{}` (only where necessary)
- Code generation eliminates reflection at runtime
- Compiler catches configuration errors
- PropertyReader provides type-specific accessors (`IntProp`, `BoolProp`, etc.)

**2. Fail-Fast Error Detection**

Configuration errors are caught early:
- **At compile time:** Type mismatches, missing methods
- **At NewGameManager():** Struct tag validation, move configuration, deck setup
- **Not at runtime:** Players never see configuration errors

Example:
```go
// Invalid struct tag caught at NewGameManager()
type gameState struct {
    base.SubState
    Cards boardgame.Stack `stack:"invalid_deck_name"`  // ERROR: deck not defined
}

// Missing required interface method caught by compiler
type moveRoll struct {
    moves.Base
}
// ERROR: must implement Legal() and Apply()
```

**3. Don't Repeat Yourself (DRY)**

Write game logic once, framework handles the rest:
- State management (versioning, copy-on-write)
- Persistence (save/load games)
- Networking (REST API, WebSockets)
- Authentication (Firebase integration)
- UI rendering (automatic databinding)
- Animations (automatic FLIP animations)

**4. Composability**

Build complex functionality from simple pieces:
- **Embed base types:** `base.GameDelegate`, `base.SubState`, `base.Move`
- **Compose behaviors:** `CurrentPlayerBehavior + PhaseBehavior + RoundRobinBehavior`
- **Stack move types:** `moves.CurrentPlayer` embedding for legality checks
- **Move progression:** Serial, Parallel, Repeat for complex sequences

Example - Build a complex move from simple pieces:
```go
type moveDealCard struct {
    moves.FixUp                                  // Auto-applied
    moves.DealComponents                         // Dealing logic
}

// Inherits:
// - FixUp legality checking
// - Deal operation implementation
// - AutoConfigurer support
// Only need to configure via WithGameProperty(), WithPlayerProperty(), etc.
```

**5. Batteries Included**

The framework provides extensive reusable components:
- **30+ move types** covering common patterns
- **9 behaviors** for standard game mechanics
- **Property system** with 14 legal types
- **Sanitization policies** for hidden information
- **Timer system** for time-based mechanics
- **Animation system** for smooth UI
- **Component system** (cards, dice, tokens)

Actual code savings example:
```go
// Without framework: ~100 lines of dealing logic
// With framework: 5 lines
auto.MustConfig(
    new(moves.DealComponentsUntilPlayerCountReached),
    moves.WithGameProperty("DrawStack"),
    moves.WithPlayerProperty("Hand"),
    moves.WithTargetCount(5),
)
```

**6. Rich Animations**

Automatic FLIP animations make games feel polished:
- Components smoothly move between positions
- Cards flip and rotate
- Opacity fades for appearing/disappearing
- No manual animation code needed
- Configurable timing per move

**7. Security by Design**

Sanitization policies prevent cheating:
- Framework enforces sanitization policies
- Server never sends hidden information to wrong players
- Policies defined via struct tags: `sanitize:"order"`, `sanitize:"len"`
- Client never sees unsanitized state

**8. Strong Validation**

Multiple validation layers:
- Type system (compile-time)
- NewGameManager() (setup-time)
- Move.Legal() (proposal-time)
- Move.ValidConfiguration() (setup-time)

**9. Determinism**

Games are fully deterministic:
- Seeded randomness (`state.Rand()`)
- Same moves + same seed = same outcome
- Enables replay, testing, debugging
- Move history is complete record

**10. Developer Experience**

Framework optimizes for game developer happiness:
- AutoConfigurer eliminates boilerplate
- Code generation handles tedious work
- Hot-reload development server
- Helpful error messages
- Comprehensive tutorial (127KB)
- Well-documented examples

### Core Design Patterns

**1. Struct Embedding Over Inheritance**

Go doesn't have traditional inheritance, so the framework uses embedding:

```go
type gameDelegate struct {
    base.GameDelegate  // Embedding, not inheriting
}

// gameDelegate now has all methods from base.GameDelegate
// Can selectively override specific methods
```

Benefits:
- Composition over inheritance
- Explicit method overriding
- No hidden complexity
- Flat hierarchy

**2. Interface-Based Abstraction**

Core concepts are interfaces:
- `GameDelegate` - Game-specific logic
- `Move` - State modifications
- `StorageManager` - Persistence
- `SubState` - State structs
- `PropertyReader` - Property access

Benefits:
- Multiple implementations (storage backends)
- Testing with mocks
- Flexibility and extensibility

**3. Code Generation for Performance**

Generate code at development time, not runtime:

```go
// Developer writes
//boardgame:codegen
type gameState struct { ... }

// Tool generates
type gameStateReader struct { ... }
func (g *gameStateReader) Prop(name string) (interface{}, error) { ... }
```

Benefits:
- Zero reflection at runtime
- Type-safe property access
- Compiler catches errors
- No runtime code generation

**4. Configuration via Struct Tags**

Struct tags provide declarative configuration:

```go
type gameState struct {
    DrawStack    boardgame.Stack      `stack:"cards"`
    Board        boardgame.SizedStack `sizedstack:"pieces,64"`
    HiddenCards  boardgame.Stack      `stack:"cards" sanitize:"order"`
}
```

Benefits:
- Configuration close to data
- Validated at setup time
- No separate config files
- Type-safe

**5. Immutable State Pattern**

State modifications only in Move.Apply():

```go
func (m *moveRollDice) Legal(state boardgame.ImmutableState, proposer PlayerIndex) error {
    // Can only READ state
    // state.GameState() returns ImmutableSubState
}

func (m *moveRollDice) Apply(state boardgame.State) error {
    // Can WRITE state
    // state.GameState() returns mutable SubState
}
```

Benefits:
- Clear separation of concerns
- No accidental mutations
- Easier reasoning about code
- Thread safety

**6. Move Progression Groups**

Complex turn sequences via move groups:

```go
moves.AddOrderedForPhase(phaseSetup,
    // These moves happen in order
    moves.Serial(
        auto.MustConfig(new(moveShuffleDeck)),
        auto.MustConfig(new(moveDealCards)),
    ),
    // Then all players act in parallel
    moves.Parallel(
        auto.MustConfig(new(moveChooseRole)),
    ),
)
```

Benefits:
- Declarative turn structure
- Framework handles progression
- Complex sequences without boilerplate

**7. Component Invariant: "One Location"**

A component can only be in one stack at a time:

```go
// This is enforced by the framework:
card.MoveTo(stack2)  // Automatically removes from stack1
```

Benefits:
- No duplicate components
- Simplified state
- Easy to track components
- Animation system relies on this

### Design Trade-offs

**Trade-off 1: Pre-Generics Approach**

**Decision:** Use code generation instead of generics
**When:** Framework created before Go 1.18 (March 2022)

Benefits:
- No runtime reflection
- Type-safe property access
- Works on older Go versions

Costs:
- Must run codegen after state changes
- Generated files in source control
- More verbose code
- Could be simpler with generics today

**Trade-off 2: Polymer 3 Frontend**

**Decision:** Use Polymer 3 + lit-element 0.7.1
**When:** ~2018 when Polymer was still actively developed

Benefits:
- Web Components standard
- No build step required
- Direct ES module loading
- Rich component library (paper-*, iron-*)

Costs:
- Polymer is now legacy
- Old lit-element version
- No TypeScript
- Limited modern tooling

**Trade-off 3: Component-Based Moves**

**Decision:** Each move is fine-grained (e.g., separate moves for dealing each card)
**Rationale:** Enables smooth animations between moves

Benefits:
- Granular animation control
- Each move is simple
- Easy to reason about state transitions

Costs:
- More moves than necessary
- More verbose configuration
- Higher database storage (one record per move)

**Trade-off 4: WebSocket for Notification, REST for Data**

**Decision:** WebSocket sends version numbers, client fetches states via REST

Benefits:
- Simpler WebSocket handling
- States can be cached
- Bandwidth efficient (don't push full states unnecessarily)

Costs:
- Extra round-trip for state
- Slightly higher latency
- More complex client code

---

## Technical Debt & Known Issues

This section documents known limitations, bugs, and areas for improvement.

### Framework-Level Issues (from README.md)

**Issue #394: Browser Support**

**Problem:** Currently only fully works in Chrome
- Safari: Partial animation support
- Firefox: Limited testing
- Other browsers: Untested

**Impact:** Limits audience

**Workaround:** Recommend Chrome to users

**Solution:** Test and fix browser-specific issues, polyfill missing APIs

**Issue #396: Animation Control**

**Problem:** Sequential moves don't pause for animations
- Moves applied faster than animations complete
- UI shows final state before animations finish

**Impact:** Less polished user experience, hard to follow rapid changes

**Workaround:** Use `delayAnimation()` in game renderers to slow move application

**Solution:** Framework should automatically pause move application until animations complete

**Issue #184: Schema Migration**

**Problem:** No upgrade path when state shape changes
- Adding/removing fields breaks old saved games
- Renaming fields breaks old saved games
- Changing stack sizes breaks old saved games

**Impact:** Cannot evolve game design after deployment

**Workaround:** Manual database migrations (error-prone)

**Solution:** Implement version-aware state deserializer with migration hooks

### Pre-Generics Limitations

**Problem:** Framework predates Go 1.18 generics

**Manifestations:**
1. **interface{} usage:** PropertyCollection, Component.Values(), storage methods
2. **Type assertions everywhere:** `pState.(*playerState)`
3. **Code generation required:** For type-safe property access
4. **Verbose move configuration:** Without generic constraints

**Example - Current vs Potential:**

```go
// Current (pre-generics)
type PropertyCollection map[string]interface{}

func (r *Reader) Prop(name string) (interface{}, error) {
    // Returns interface{}, caller must type assert
}

// With generics (potential)
func Prop[T any](r *Reader, name string) (T, error) {
    // Returns T, no type assertion needed
}

// Current
game.DrawStack.ComponentAt(0).Values().(CardValues).Suit

// With generics (potential)
game.DrawStack.ComponentAt(0).Values().Suit  // Type-safe
```

**Impact:**
- More verbose code
- Runtime type assertion errors possible
- Code generation maintenance burden
- Steeper learning curve

**Estimated Fix Effort:** Large (4-6 months, breaking changes)

### Dependency Age

**Go Dependencies (from go.mod):**
- **Declared Go version:** 1.13 (March 2019)
- **Current Go version:** 1.23+ (August 2024)
- **Missing features:** Generics, better error handling, performance improvements

**Old Dependencies:**
- gin-gonic/gin (web framework) - Older version, security updates available
- sirupsen/logrus (logging) - Still maintained but dated
- Many transitive dependencies from 2017-2018 era

**Impact:**
- Potential security vulnerabilities
- Missing performance improvements
- Incompatibility with modern tools

**JavaScript Dependencies (from package.json):**
- **lit-element:** 0.7.1 (2018) → Current: Lit 3.x (2024)
- **Polymer:** 3.3.0 (2018, last major version) → Legacy framework
- **Redux:** 4.0.5 (2019) → Current: Redux Toolkit recommended
- **Firebase:** 5.11.1 (2018) → Current: 10.x (2024, breaking changes)

**Impact:**
- Security vulnerabilities
- Missing features
- Harder to find documentation/support
- Cannot use modern JavaScript features

### Documentation Gaps

**Documented:**
- ✅ TUTORIAL.md (comprehensive 127KB walkthrough)
- ✅ README.md (overview and getting started)
- ✅ moves/doc.go (moves package documentation)
- ✅ server/static/src/ARCHITECTURE.md (frontend animation system)
- ✅ THIS DOCUMENT (overall architecture)

**Missing:**
- ❌ Performance characteristics (memory usage, throughput, latency)
- ❌ Deployment guide (production setup, monitoring, scaling)
- ❌ Storage backend comparison (when to use which)
- ❌ Security best practices (sanitization patterns, common pitfalls)
- ❌ Migration guide (if moving from pre-1.0 versions)
- ❌ API reference (generated docs from godoc)

### Code Quality Observations

From code exploration:

**TODOs and FIXMEs:**
- ~42 TODO/FIXME comments across the codebase
- Key areas: game_manager.go (5), state.go (3), property_reader.go (3)

**"Hack" Comments:**
- "This is a hack" appears 12+ times
- "Ugly hack" appears 5+ times
- Indicates technical debt, workarounds, or non-obvious code

**Thread Safety:**
- Some concurrent access patterns lack mutex protection
- shadowComponents cache in deck.go may need mutex
- modifiableGames map protected, but could be more granular

**Timer Management:**
- Complex timer coordination
- Some edge cases with timer updates
- Needs improvement (per comments)

### Missing Features

**State Migration:**
- Issue #184 (documented above)
- Critical for production games

**Historical State Access:**
- Can fetch old states, but no convenient API
- No UI for replay/rewind
- Useful for debugging, spectating

**Complete Sanitization Policies:**
- Limited built-in policies (order, len)
- Custom policies require code
- No visual editor/validator

**Board Game Tooling:**
- Examples don't demonstrate board games well
- Sized stacks work but underexplored
- Path-finding, adjacency helpers missing

**Test Coverage:**
- Core framework has tests
- Example games have minimal tests
- Frontend lacks comprehensive tests
- Integration tests limited

### Browser Compatibility Issues (Issue #394)

**Chrome:** ✅ Fully working
- All animations
- WebSocket stable
- PWA features work

**Safari:** ⚠️ Partial support
- Some animations work
- WebSocket issues reported
- PWA support varies by version

**Firefox:** ⚠️ Limited testing
- Basic functionality works
- Animation edge cases unknown
- WebSocket untested at scale

**Mobile Browsers:** ⚠️ Varies
- Chrome Android: Generally works
- Safari iOS: Similar issues to desktop Safari
- Other mobile browsers: Untested

**Root Causes:**
- Reliance on specific Web Animations API features
- Polymer 3 compatibility issues
- CSS custom properties support
- WebSocket implementation differences

---

## Modernization Strategies

This section outlines strategies for modernizing the framework to use current best practices and technologies.

### Strategy 1: Go Generics Migration

**Goal:** Eliminate `interface{}` usage and type assertions using Go 1.18+ generics

**Estimated Impact:** Could eliminate 30-40% of interface{} usage, improve type safety significantly

**Timeline:** 4-6 months (breaking changes, requires coordination)

#### High-Value Generic Candidates

**1. PropertyCollection**

**Current:**
```go
type PropertyCollection map[string]interface{}

func (g *gameDelegate) ComputedGlobalProperties(state ImmutableState) PropertyCollection {
    return PropertyCollection{
        "CurrentPlayerName": state.PlayerStates()[game.CurrentPlayer].Name,
    }
}

// Client code needs type assertion
name := props["CurrentPlayerName"].(string)
```

**With Generics:**
```go
type PropertyCollection[T any] map[string]T

// Or type-safe getter:
func GetProp[T any](props PropertyCollection, name string) (T, error) {
    // Type-safe retrieval
}

// Client code is type-safe:
name, err := GetProp[string](props, "CurrentPlayerName")
```

**Files to modify:**
- game_delegate.go
- property_reader.go
- All code generation templates

**2. Reader/Setter Prop() Methods**

**Current:**
```go
type PropertyReader interface {
    Prop(name string) (interface{}, error)
    IntProp(name string) (int, error)      // Type-specific accessors
    BoolProp(name string) (bool, error)
    // ... 14 type-specific methods
}
```

**With Generics:**
```go
type PropertyReader interface {
    Prop[T any](name string) (T, error)  // Single generic method
}

// Usage:
score, err := state.Prop[int]("TotalScore")
done, err := state.Prop[bool]("Done")
```

**Benefits:**
- Eliminate 13 type-specific accessor methods
- Compile-time type checking
- Cleaner API

**Challenges:**
- Breaking change for all game code
- Code generation templates need updates
- Backward compatibility difficult

**3. Storage Interfaces**

**Current:**
```go
type StateStorageRecord struct {
    GameID  string
    Version int
    Blob    []byte  // Serialized state
}

// Returns interface{} that must be type-asserted
func (s *StorageManager) State(gameID string, version int) (interface{}, error)
```

**With Generics:**
```go
type StateStorageRecord[T SubState] struct {
    GameID  string
    Version int
    Data    T  // Strongly typed
}

func State[T SubState](s *StorageManager, gameID string, version int) (T, error) {
    // Returns strongly-typed state
}
```

**Benefits:**
- Type-safe state retrieval
- No serialization errors at runtime
- Better IDE support

**4. Component System**

**Current:**
```go
type Component interface {
    Values() ComponentValues  // Returns interface{}
    DynamicValues() SubState  // Returns interface{}
}

// Usage requires type assertion:
cardValues := component.Values().(*playingcards.CardValue)
suit := cardValues.Suit
```

**With Generics:**
```go
type Component[V ComponentValues] interface {
    Values() V  // Strongly typed
}

type Card = Component[playingcards.CardValue]

// Usage is type-safe:
suit := card.Values().Suit  // No type assertion
```

**5. Stack Types**

**Current:**
```go
type Stack interface {
    ComponentAt(index int) Component  // Returns generic Component
}

// Must type-assert Values():
card := stack.ComponentAt(0)
suit := card.Values().(*playingcards.CardValue).Suit
```

**With Generics:**
```go
type Stack[C Component[V], V ComponentValues] interface {
    ComponentAt(index int) C
}

type CardStack = Stack[Card, playingcards.CardValue]

// Type-safe:
card := cardStack.ComponentAt(0)
suit := card.Values().Suit  // No assertion needed
```

#### Backward Compatibility Approach

**Phase 1: Add Generic Versions Alongside Existing**

```go
// Keep old interface
type PropertyReader interface {
    Prop(name string) (interface{}, error)
    IntProp(name string) (int, error)
    // ...
}

// Add new generic interface
type GenericPropertyReader interface {
    PropGeneric[T any](name string) (T, error)
}

// Deprecation warning in docs
// @deprecated: Use PropGeneric[T]() instead
func (r *Reader) Prop(name string) (interface{}, error) {
    // Implementation
}
```

**Phase 2: Update Codegen to Generate Both**

Generated code implements both old and new interfaces for 2-3 versions.

**Phase 3: Deprecate Old Interfaces**

Clear deprecation warnings, migration guide, automated migration tool.

**Phase 4: Remove Old Interfaces**

After sufficient adoption period (6-12 months).

#### Migration Tool

Provide automated migration:

```bash
boardgame-util migrate-generics ./path/to/game
```

Transforms:
```go
// Before
score := pState.(*playerState).IntProp("TotalScore")

// After
score, _ := pState.Prop[int]("TotalScore")
```

#### Estimated Timeline

- **Months 1-2:** Design generic APIs, prototype
- **Months 3-4:** Implement generic versions, update codegen, update core framework
- **Month 5:** Update all example games, write migration guide
- **Month 6:** Community testing, bug fixes, migration tool
- **Months 7-18:** Deprecation period
- **Month 19:** Remove old APIs (v2.0)

### Strategy 2: Web Frontend Migration

**Goal:** Migrate from Polymer 3 + lit-element 0.7.1 to modern Lit 3+ with TypeScript

**Estimated Impact:** Modern development experience, better performance, easier maintenance

**Timeline:** 6-9 months (substantial effort, high risk)

#### Target Architecture

**Technology Stack:**
- **Lit 3.x** - Modern web components (currently at Lit 3.2)
- **TypeScript 5.x** - Type safety for frontend
- **Vite or Rollup** - Modern bundler (fast, good DX)
- **Redux Toolkit** - Modern Redux patterns
- **Firebase 10.x** - Latest SDK

**Reference:** Use `/Users/jkomoros/Code/card-web/` as blueprint (already uses Lit + TypeScript + Redux)

#### Migration Phases

**Phase A: Foundation (Months 1-2)**

1. **Add Build System**
   ```
   npm install -D vite typescript @vitejs/plugin-legacy
   npm install lit@3
   ```

2. **TypeScript Configuration**
   ```json
   // tsconfig.json
   {
       "compilerOptions": {
           "target": "ES2020",
           "module": "ES2020",
           "lib": ["ES2020", "DOM", "DOM.Iterable"],
           "moduleResolution": "node",
           "strict": true,
           "esModuleInterop": true,
           "skipLibCheck": true
       }
   }
   ```

3. **Update package.json**
   ```json
   {
       "scripts": {
           "dev": "vite",
           "build": "vite build",
           "preview": "vite preview"
       },
       "dependencies": {
           "lit": "^3.2.0",
           "redux": "^5.0.0",
           "@reduxjs/toolkit": "^2.0.0",
           "firebase": "^10.0.0"
       }
   }
   ```

4. **Create vite.config.ts**
   ```typescript
   import { defineConfig } from 'vite';

   export default defineConfig({
       build: {
           rollupOptions: {
               output: {
                   manualChunks: {
                       'lit': ['lit'],
                       'redux': ['redux', '@reduxjs/toolkit'],
                       'firebase': ['firebase/app', 'firebase/auth']
                   }
               }
           }
       }
   });
   ```

**Phase B: Core Infrastructure (Months 3-4)**

1. **Migrate Redux to TypeScript + Redux Toolkit**

   Before (Redux 4 + plain actions):
   ```javascript
   // actions/game.js
   export const UPDATE_GAME = 'UPDATE_GAME';
   export const updateGame = (game) => ({
       type: UPDATE_GAME,
       game
   });
   ```

   After (Redux Toolkit + TypeScript):
   ```typescript
   // features/gameSlice.ts
   import { createSlice, PayloadAction } from '@reduxjs/toolkit';

   interface GameState {
       id: string;
       name: string;
       currentState: any;  // TODO: Type this
   }

   const gameSlice = createSlice({
       name: 'game',
       initialState: {} as GameState,
       reducers: {
           updateGame(state, action: PayloadAction<GameState>) {
               return action.payload;
           }
       }
   });
   ```

2. **Replace boardgame-ajax with Modern Fetch**

   Before (iron-ajax wrapper):
   ```html
   <boardgame-ajax path="/game/blackjack/abc123"></boardgame-ajax>
   ```

   After (typed fetch wrapper):
   ```typescript
   // api/client.ts
   class BoardgameClient {
       async getGame(name: string, id: string): Promise<GameInfo> {
           const response = await fetch(`/api/game/${name}/${id}`, {
               credentials: 'include'
           });
           return response.json();
       }
   }
   ```

3. **Update WebSocket Handling**

   TypeScript types for WebSocket messages:
   ```typescript
   interface VersionNotification {
       version: number;
   }

   class GameSocket {
       private socket: WebSocket;

       connect(gameName: string, gameId: string) {
           this.socket = new WebSocket(`ws://${host}/api/game/${gameName}/${gameId}/socket`);
           this.socket.onmessage = (event) => {
               const version = parseInt(event.data);
               this.onVersionUpdate(version);
           };
       }
   }
   ```

**Phase C: Animation System Migration (Months 5-6) ⚠️ CRITICAL**

This is the highest-risk phase. The animation system is sophisticated and must be preserved.

**Approach:**

1. **Extract Animation Logic to Pure TypeScript**

   Separate animation logic from Polymer-specific code:
   ```typescript
   // animation/flipAnimator.ts
   export class FLIPAnimator {
       private components: Map<string, ComponentPosition> = new Map();

       recordPositions(elements: Element[]) {
           // FIRST: Record positions
       }

       applyInverseTransforms(elements: Element[]) {
           // INVERT: Apply inverse transforms
       }

       play(elements: Element[]) {
           // PLAY: Animate to identity
       }
   }
   ```

2. **Create Lit Version of boardgame-component**

   ```typescript
   import { LitElement, html, css } from 'lit';
   import { customElement, property } from 'lit/decorators.js';

   @customElement('boardgame-component')
   export class BoardgameComponent extends LitElement {
       @property({ type: Object }) item?: Component;
       @property({ type: String }) id?: string;

       render() {
           return html`
               <div class="component" data-component-id="${this.id}">
                   ${this.renderContent()}
               </div>
           `;
       }
   }
   ```

3. **Migrate boardgame-card with Internal Animations**

   ```typescript
   @customElement('boardgame-card')
   export class BoardgameCard extends BoardgameComponent {
       @property({ type: Boolean }) faceUp = false;
       @property({ type: Boolean }) rotated = false;

       static styles = css`
           :host {
               transition: transform var(--animation-length, 0.3s);
           }

           .card.face-down {
               transform: rotateY(180deg);
           }

           .card.rotated {
               transform: rotate(90deg);
           }
       `;
   }
   ```

4. **Extensive Testing**

   Test matrix:
   - Card flip animations
   - Card rotation animations
   - Component movement between stacks
   - Appearing/disappearing components
   - Sanitized stack animations (opacity fade)
   - All 6 example games
   - All browsers (Chrome, Firefox, Safari)

**Phase D: UI Components (Months 7-8)**

1. **Replace paper-* Components**

   | Polymer Component | Replacement |
   |-------------------|-------------|
   | paper-button | Custom Lit button or Material Web Components |
   | paper-input | Custom Lit input or native \<input\> with styling |
   | paper-dialog | Custom Lit dialog or native \<dialog\> |
   | paper-checkbox | Custom Lit checkbox or native \<input type="checkbox"\> |
   | paper-slider | Custom Lit slider or native \<input type="range"\> |

2. **Replace iron-* Components**

   | Polymer Component | Replacement |
   |-------------------|-------------|
   | iron-ajax | fetch API |
   | iron-pages | Lit router or custom page switching |
   | iron-flex-layout | CSS Flexbox/Grid |
   | iron-icons | SVG icons or icon font |
   | iron-selector | Custom Lit selector |

**Phase E: Game Renderers (Month 9)**

1. **Migration Guide for Game Developers**

   Document how to migrate game renderers:
   ```markdown
   # Migrating Game Renderers to Lit 3

   ## Before (Polymer 3)
   ```javascript
   class BoardgameRenderGamePig extends BoardgameBaseGameRenderer {
       static get template() {
           return html`<div>[[state.Game.TargetScore]]</div>`;
       }
   }
   ```

   ## After (Lit 3)
   ```typescript
   @customElement('boardgame-render-game-pig')
   class BoardgameRenderGamePig extends BoardgameBaseGameRenderer {
       render() {
           return html`<div>${this.state.Game.TargetScore}</div>`;
       }
   }
   ```
   ```

2. **Update All 6 Example Games**

3. **Provide Compatibility Layer (Optional)**

   Shim layer for gradual migration if needed.

#### Risk Mitigation

**High-Risk Areas:**
1. **Animation System** - Most complex, most visible
   - Mitigation: Extensive testing, keep Polymer version running in parallel initially

2. **Breaking Game Renderers** - Affects all downstream games
   - Mitigation: Provide migration tool, detailed guide, compatibility period

3. **Performance Regression** - New stack might be slower
   - Mitigation: Performance testing, benchmarks

#### Alternative: Incremental Migration

Instead of big-bang migration:
1. Keep Polymer components for game rendering
2. Migrate app shell to Lit 3 (already mostly Lit)
3. Add TypeScript gradually
4. Migrate game components one-by-one over 12-18 months

**Benefits:** Lower risk, gradual adoption
**Costs:** Longer timeline, maintaining two systems

### Strategy 3: Dependency Updates

**Goal:** Update outdated dependencies for security, performance, and compatibility

**Timeline:** 2-3 months (medium risk)

#### Go Dependency Updates

**Priority 1: Critical Security Updates**

1. Check for CVEs:
   ```bash
   go list -json -m all | nancy sleuth
   ```

2. Update go.mod:
   ```
   go 1.13  →  go 1.22 (or latest LTS)
   ```

3. Update dependencies with security issues

**Priority 2: Major Dependencies**

| Dependency | Current | Latest | Breaking Changes? |
|------------|---------|--------|-------------------|
| gin-gonic/gin | old | v1.10+ | Review changelog |
| sirupsen/logrus | old | v1.9+ | Minimal |
| gorilla/websocket | old | v1.5+ | Minimal |
| database drivers | old | latest | Check compatibility |

**Update Process:**
```bash
# Update one at a time
go get -u github.com/gin-gonic/gin
go mod tidy
go test ./...

# Check for breaking changes in each
```

**Testing After Updates:**
- Run full test suite
- Test each storage backend
- Test each example game
- Test WebSocket connections
- Load test (ensure no performance regression)

#### JavaScript Dependency Updates

**Priority 1: Security & Compatibility**

1. **Firebase 5.11.1 → 10.x**
   - **Breaking Changes:** Major API changes
   - **Migration:** https://firebase.google.com/docs/web/modular-upgrade
   ```javascript
   // Old
   import firebase from 'firebase/app';
   import 'firebase/auth';
   firebase.initializeApp(config);
   firebase.auth().signInWithEmailAndPassword(email, password);

   // New
   import { initializeApp } from 'firebase/app';
   import { getAuth, signInWithEmailAndPassword } from 'firebase/auth';
   const app = initializeApp(config);
   const auth = getAuth(app);
   signInWithEmailAndPassword(auth, email, password);
   ```

2. **@webcomponents/webcomponentsjs**
   - Check if still needed (most browsers support web components natively now)
   - Update to latest for better browser support

**Priority 2: Framework Updates (if not doing full migration)**

If sticking with Polymer temporarily:
- Keep Polymer 3.3.0 (no updates available, it's EOL)
- Update redux: `4.0.5 → 5.0.1` (minimal breaking changes)
- Update redux-thunk: `2.3.0 → 3.1.0`

**Priority 3: Development Dependencies**

- ESLint updates
- Test framework updates
- Build tool updates

#### Dependency Update Strategy

**Approach: Incremental, Test-Driven**

1. Update one dependency at a time
2. Run full test suite after each
3. Test manually in browser
4. Commit each update separately (easier rollback)

**Don't:** Update all at once (too risky)

### Strategy 4: Bug Fixes & Technical Debt Reduction

**Goal:** Address known issues and improve code quality

**Timeline:** Ongoing, prioritize critical bugs

#### High-Priority Bug Fixes

**1. Browser Compatibility (Issue #394)**

**Chrome:** Already works
**Safari:** Fix animation issues
**Firefox:** Test and fix

Specific issues to address:
- Web Animations API polyfill for Safari
- WebSocket stability across browsers
- CSS custom properties fallbacks
- Test on iOS Safari

**Approach:**
```javascript
// Add feature detection
const supportsWebAnimations = 'animate' in document.createElement('div');
if (!supportsWebAnimations) {
    // Load polyfill
    import('web-animations-js');
}
```

**2. Animation Sequencing (Issue #396)**

**Problem:** Moves don't wait for animations

**Solution:** Add animation completion tracking
```javascript
class GameStateManager {
    async installNextState(state) {
        await this.animator.animate(state);  // Wait for animations
        this.setState(state);
    }
}
```

**3. Timer Edge Cases**

Review and fix timer management code:
- Synchronization issues
- Race conditions
- Memory leaks

**4. Thread Safety**

Add mutexes where needed:
```go
// deck.go - shadowComponents cache
type Deck struct {
    shadowComponents map[string]*Component
    shadowLock       sync.RWMutex  // ADD THIS
}

func (d *Deck) getShadowComponent(id string) *Component {
    d.shadowLock.RLock()
    defer d.shadowLock.RUnlock()
    return d.shadowComponents[id]
}
```

#### Code Quality Improvements

**1. Reduce "Hack" Comments**

Go through each "hack" comment and either:
- Refactor to remove the hack
- Document why it's necessary
- File issue to address later

**2. Complete TODO Items**

Triage ~42 TODOs:
- Fix now
- File issue for later
- Remove if obsolete

**3. Improve Error Messages**

Replace string matching hacks with proper error types:
```go
// Before
if strings.Contains(err.Error(), "not found") {
    // Handle not found
}

// After
var ErrNotFound = errors.New("not found")

if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

### Strategy 5: Testing & Documentation Improvements

**Goal:** Increase test coverage and improve documentation

**Timeline:** Ongoing

#### Testing Improvements

**Current State:**
- Core framework has tests
- Example games have minimal tests
- Frontend lacks comprehensive tests
- No integration tests

**Target State:**
- 80%+ core framework coverage
- Each example game has comprehensive tests
- Frontend component tests
- Integration tests for full game flows

**Add Tests For:**

1. **Move Validation**
   ```go
   func TestMoveRollDice_Legal(t *testing.T) {
       game := setupTestGame()
       move := &moveRollDice{}

       // Test when legal
       err := move.Legal(game.CurrentState(), game.CurrentPlayer())
       assert.NoError(t, err)

       // Test when illegal (already done)
       game.CurrentPlayerState().Done = true
       err = move.Legal(game.CurrentState(), game.CurrentPlayer())
       assert.Error(t, err)
   }
   ```

2. **State Transitions**
   ```go
   func TestGameFlow(t *testing.T) {
       game := NewTestGame()

       // Initial state
       assert.Equal(t, phaseSetup, game.State().Phase())

       // Apply moves, check state
       game.ApplyMove("Deal Cards")
       assert.Equal(t, phasePlay, game.State().Phase())
   }
   ```

3. **Frontend Components**
   ```typescript
   import { fixture, html } from '@open-wc/testing';
   import './boardgame-card';

   describe('BoardgameCard', () => {
       it('flips when faceUp changes', async () => {
           const el = await fixture(html`
               <boardgame-card .faceUp=${false}></boardgame-card>
           `);

           expect(el.faceUp).to.be.false;
           el.faceUp = true;
           await el.updateComplete;
           expect(el.faceUp).to.be.true;
       });
   });
   ```

#### Documentation Improvements

**Add Missing Docs:**

1. **Performance Guide**
   - Memory usage per game
   - Recommended scaling approach
   - Database optimization
   - Caching strategies

2. **Deployment Guide**
   - Production setup
   - Monitoring and logging
   - Backup strategies
   - Scaling horizontally

3. **Storage Backend Comparison**
   | Backend | Use Case | Pros | Cons |
   |---------|----------|------|------|
   | Memory | Testing, demos | Fast, simple | No persistence |
   | Filesystem | Single-user, dev | Human-readable | No concurrency |
   | Bolt | Single-server | Embedded, fast | No horizontal scaling |
   | MySQL | Production | Scales, reliable | Requires setup |

4. **Security Best Practices**
   - Sanitization patterns
   - Authentication setup
   - CORS configuration
   - Input validation

---

## Execution Recommendations

This section provides guidance on how to approach modernization in practice.

### Recommended Priority Order

**Phase 1: Stabilization & Documentation** (1-2 months)

**Goals:** Fix critical bugs, document current state

**Tasks:**
1. ✅ Create ARCHITECTURE.md (this document) - DONE
2. Fix browser compatibility issues (Issue #394)
3. Fix animation sequencing (Issue #396)
4. Update README with current status
5. Document known limitations
6. Add tests for critical paths
7. Security audit of dependencies

**Deliverables:**
- Updated documentation
- Critical bugs fixed
- Test coverage report
- Security audit report

**Risk:** Low - No breaking changes

---

**Phase 2: Dependency Updates** (2-3 months)

**Goals:** Update to modern, secure dependencies

**Tasks:**
1. Update Go to 1.22+
   - Update go.mod
   - Update go.sum
   - Test with new version
   - Update CI/CD

2. Update safe Go dependencies
   - Review each dependency
   - Check for breaking changes
   - Update one at a time
   - Test after each update

3. Update JavaScript dependencies (non-breaking)
   - Redux 4 → 5
   - Firebase 5 → 10 (breaking, but manageable)
   - Development dependencies

4. Test all example games
5. Performance testing

**Deliverables:**
- Updated dependencies
- All tests passing
- Performance benchmarks
- Migration notes

**Risk:** Low-Medium - Potential for compatibility issues

---

**Phase 3: Go Generics (Optional)** (4-6 months)

**Goals:** Modernize Go code with generics

**Tasks:**
1. Design generic APIs
   - PropertyCollection[T]
   - Prop[T]() methods
   - Generic Stack[T]
   - Generic Component[T]

2. Implement alongside existing
   - Both old and new APIs work
   - Deprecation warnings

3. Update code generation
   - Generate both old and new
   - Allow opt-in to generics

4. Update examples
   - Migrate one example
   - Document migration process
   - Migration tool

5. Community testing period

**Deliverables:**
- Generic APIs available
- Migration guide
- Automated migration tool
- Example games updated

**Risk:** High - Breaking changes, requires careful planning

---

**Phase 4: Frontend Modernization** (6-9 months)

**Goals:** Migrate to Lit 3 + TypeScript

**Tasks:**
1. **Months 1-2:** Foundation
   - Add TypeScript
   - Add Vite/Rollup
   - Upgrade Lit
   - Update package.json

2. **Months 3-4:** Infrastructure
   - Migrate Redux to TypeScript
   - Replace iron-ajax
   - Update WebSocket handling

3. **Months 5-6:** Animation System ⚠️
   - Extract animation logic
   - Create Lit versions
   - Extensive testing

4. **Months 7-8:** UI Components
   - Replace paper-* components
   - Replace iron-* components

5. **Month 9:** Game Renderers
   - Migration guide
   - Update examples
   - Community support

**Deliverables:**
- Modern Lit 3 + TypeScript frontend
- All animations working
- All example games migrated
- Migration guide for game developers

**Risk:** High - Complex migration, affects all games

---

**Phase 5: Polish & New Features** (Ongoing)

**Goals:** Improve framework based on feedback

**Tasks:**
- Schema migration system (Issue #184)
- Board game tooling
- Performance optimizations
- Additional example games
- Community feature requests

**Deliverables:**
- Continuous improvements
- New features
- Bug fixes

**Risk:** Low-Medium - Depends on specific features

### Risk Assessment

**Low Risk Changes:**
- ✅ Documentation improvements
- ✅ Adding tests
- ✅ Bug fixes (non-breaking)
- ✅ Security updates (patch versions)
- ✅ Code quality improvements

**Medium Risk Changes:**
- ⚠️ Dependency updates (minor/major versions)
- ⚠️ Go version update
- ⚠️ Frontend framework updates (if careful)
- ⚠️ Performance optimizations

**High Risk Changes:**
- ❌ Go generics migration (breaking changes)
- ❌ Complete frontend rewrite
- ❌ Storage layer changes
- ❌ Core engine refactoring
- ❌ Animation system changes

### Backward Compatibility Strategy

**For Go Generics Migration:**

1. **Deprecation Period:** 12-18 months
   - Old APIs marked `@deprecated`
   - Compiler warnings
   - Clear migration path

2. **Parallel APIs:** Both old and new work
   ```go
   // Old (deprecated)
   func (r *Reader) Prop(name string) (interface{}, error)

   // New
   func Prop[T any](r *Reader, name string) (T, error)
   ```

3. **Migration Tool:** Automated code transformation

4. **Semantic Versioning:**
   - v1.x - Old APIs
   - v2.0 - Remove old APIs (breaking)

**For Frontend Migration:**

1. **Component-by-Component:** Migrate incrementally
2. **Compatibility Layer:** Shim if needed
3. **Documentation:** Clear migration guide
4. **Support Period:** 6 months for questions

### Success Metrics

**How to measure success:**

1. **Compilation:** All example games compile
   ```bash
   for game in examples/*/; do
       (cd "$game" && go build) || exit 1
   done
   ```

2. **Tests:** All tests pass
   ```bash
   go test ./... -v
   ```

3. **Performance:** No regression
   - Measure before and after
   - Latency, throughput, memory

4. **Browser Compatibility:** Works in Chrome, Firefox, Safari

5. **Community Adoption:** Games successfully migrate

---

## Verification Plan

This section describes how to verify the framework works correctly after changes.

### Pre-Change Baseline

Before any modernization:

1. **Record Current State**
   ```bash
   # Compilation
   go build ./...  # Should succeed

   # Tests
   go test ./... -v  # Record pass/fail count

   # Example games
   for game in examples/*/; do
       echo "Testing $game"
       (cd "$game" && go test ./...)
   done
   ```

2. **Manual Testing Checklist**
   - [ ] Create new game via UI
   - [ ] Multiple players can join
   - [ ] Propose and apply moves
   - [ ] Animations work smoothly
   - [ ] WebSocket updates work
   - [ ] Game persists across server restart
   - [ ] Works in Chrome
   - [ ] Works in Firefox
   - [ ] Works in Safari

3. **Performance Baseline**
   ```bash
   # Measure current performance
   go test -bench=. -benchmem ./...
   ```

### Testing Strategy

**Unit Tests:**
```bash
# Run all unit tests
go test ./... -v

# With coverage
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Integration Tests:**
```bash
# Test each storage backend
go test ./storage/memory -v
go test ./storage/filesystem -v
go test ./storage/bolt -v
go test ./storage/mysql -v  # Requires MySQL running
```

**Example Game Tests:**
```bash
cd examples/pig && go test ./...
cd examples/memory && go test ./...
cd examples/blackjack && go test ./...
cd examples/checkers && go test ./...
cd examples/tictactoe && go test ./...
cd examples/debuganimations && go test ./...
```

**Frontend Tests:**
```bash
cd server/static
npm test
```

### Post-Change Verification

After each modernization phase:

**1. Compilation Check**
```bash
#!/bin/bash
set -e

echo "Checking core framework..."
go build ./...

echo "Checking examples..."
for game in examples/*/; do
    echo "Building $game"
    (cd "$game" && go build)
done

echo "✅ All builds successful"
```

**2. Test Suite**
```bash
#!/bin/bash
set -e

echo "Running core tests..."
go test ./... -v

echo "Running example tests..."
for game in examples/*/; do
    echo "Testing $game"
    (cd "$game" && go test ./...)
done

echo "✅ All tests passed"
```

**3. Storage Backend Verification**

For each backend (memory, filesystem, bolt, mysql):

```go
func TestStorageBackend(t *testing.T, storage StorageManager) {
    // Create game
    game := NewTestGame(storage)
    err := storage.SaveGameAndCurrentState(game.StorageRecord(), game.CurrentState(), nil)
    assert.NoError(t, err)

    // Apply 10 moves
    for i := 0; i < 10; i++ {
        move := game.ProposeLegalMove()
        game.ApplyMove(move)
    }

    // Retrieve game
    loaded, err := storage.Game(game.ID())
    assert.NoError(t, err)
    assert.Equal(t, 10, loaded.Version)

    // Retrieve historical state
    state5, err := storage.State(game.ID(), 5)
    assert.NoError(t, err)
    assert.NotNil(t, state5)
}
```

**4. Browser Compatibility Testing**

| Browser | Test | Expected |
|---------|------|----------|
| Chrome Latest | Full manual test | ✅ Everything works |
| Firefox Latest | Full manual test | ✅ Everything works |
| Safari Latest | Full manual test | ✅ Everything works |
| Chrome Android | Basic smoke test | ✅ Core functionality works |
| Safari iOS | Basic smoke test | ✅ Core functionality works |

**5. Animation Verification**

Critical: Test all animation types work correctly

- [ ] Card movement between stacks
- [ ] Card flip (face up/down)
- [ ] Card rotation (90 degrees)
- [ ] Component appearing (fade in)
- [ ] Component disappearing (fade out)
- [ ] Multiple simultaneous animations
- [ ] Sanitized stack animations (to/from center with opacity)
- [ ] Animation timing feels smooth (not too fast/slow)

**Test each game:**
- [ ] Pig - Die rolling animation
- [ ] Memory - Card revealing animation
- [ ] Blackjack - Card dealing animation
- [ ] Checkers - Piece movement
- [ ] TicTacToe - Token placement
- [ ] DebugAnimations - All animation types

**6. Performance Verification**

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Compare to baseline
# Memory usage should not increase significantly
# Throughput should not decrease significantly
```

**7. End-to-End Game Flow**

Test complete game flows:

```
1. Create Game
   - Via API: POST /api/new
   - Via UI: Click "New Game"
   - ✅ Game appears in list

2. Multiple Players Join
   - Open in multiple browsers
   - Each joins as different player
   - ✅ All players see each other

3. Play Complete Game
   - Players take turns
   - Propose moves
   - ✅ Moves apply correctly
   - ✅ Animations play
   - ✅ WebSocket updates all clients
   - ✅ State stays synchronized

4. Game Completion
   - Reach win condition
   - ✅ Game marked as finished
   - ✅ Winner displayed correctly
   - ✅ Scores calculated correctly

5. Persistence
   - Restart server
   - ✅ Game still exists
   - ✅ Can load and continue
```

### Rollback Plan

If verification fails:

1. **Identify Issue**
   - Which test failed?
   - What's the error message?
   - Can it be quickly fixed?

2. **Quick Fix or Rollback**
   - If fixable in < 1 hour: Fix
   - Otherwise: Rollback

3. **Rollback Process**
   ```bash
   git revert HEAD  # Revert last commit
   # Or
   git reset --hard <previous-commit>

   # Rebuild and test
   go build ./...
   go test ./...
   ```

4. **Document Issue**
   - File GitHub issue
   - Document what went wrong
   - Plan fix for next iteration

### Continuous Integration

Recommended CI/CD setup (GitHub Actions example):

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Build
        run: go build ./...

      - name: Test
        run: go test ./... -v

      - name: Test Examples
        run: |
          for game in examples/*/; do
            (cd "$game" && go test ./...)
          done
```

---

## References & Further Reading

### Internal Documentation

**Comprehensive Guides:**
- **[TUTORIAL.md](TUTORIAL.md)** - 127KB comprehensive tutorial walking through building a complete game
  - Covers all major concepts
  - Step-by-step instructions
  - Best starting point for new game developers

- **[README.md](README.md)** - Project overview and getting started
  - Design goals and philosophy
  - Current status and known issues
  - Quick start guide

- **[server/static/src/ARCHITECTURE.md](server/static/src/ARCHITECTURE.md)** - Detailed frontend animation system documentation
  - FLIP animation technique
  - Component hierarchy
  - Animation timing and coordination

- **[boardgame-util/README.md](boardgame-util/README.md)** - CLI tool configuration and commands
  - All command documentation
  - Configuration file format
  - Development workflow

**Package Documentation:**
- **[moves/doc.go](moves/doc.go)** - Extensive moves package documentation
  - Move type hierarchy
  - AutoConfigurer usage
  - Move progression groups

- **[behaviors/main.go](behaviors/main.go)** - Behaviors package overview
  - All 9 behavior interfaces
  - Usage patterns

- **[storage/mysql/README.md](storage/mysql/README.md)** - MySQL storage setup
  - Database schema
  - Setup instructions

### Key Source Files

**Core Framework:**
- **[game_manager.go](game_manager.go)** (38,138 lines) - Central orchestrator
- **[state.go](state.go)** (44,580 lines) - State system and interfaces
- **[move.go](move.go)** (17,044 lines) - Move system
- **[stack.go](stack.go)** (56,933 lines) - Stack and component collections
- **[component.go](component.go)** - Component system
- **[deck.go](deck.go)** - Deck management
- **[property_reader.go](property_reader.go)** (45,103 lines) - Property system
- **[sanitization.go](sanitization.go)** - State sanitization

**High-Level Packages:**
- **[base/game_delegate.go](base/game_delegate.go)** - Base GameDelegate implementation
- **[base/main.go](base/main.go)** - SubState base struct
- **[base/move.go](base/move.go)** - Base Move implementation

**Example Games (in order of complexity):**
- **[examples/pig/](examples/pig/)** - Simple dice game (~500 lines)
- **[examples/tictactoe/](examples/tictactoe/)** - Classic game (~600 lines)
- **[examples/memory/](examples/memory/)** - Card matching (~1,000 lines)
- **[examples/blackjack/](examples/blackjack/)** - Card game with phases (~1,200 lines)
- **[examples/checkers/](examples/checkers/)** - Board game (~2,000 lines)
- **[examples/debuganimations/](examples/debuganimations/)** - Animation testing (~7,000 lines)

### External Resources

**Go Language:**
- [Go Documentation](https://golang.org/doc/) - Official Go docs
- [Effective Go](https://golang.org/doc/effective_go) - Go programming patterns
- [Go Generics Tutorial](https://go.dev/doc/tutorial/generics) - Introduction to generics (Go 1.18+)

**Web Components & Lit:**
- [Lit Documentation](https://lit.dev/) - Modern web components framework
- [Polymer Project](https://www.polymer-project.org/) - Legacy framework (reference only)
- [Web Components](https://developer.mozilla.org/en-US/docs/Web/Web_Components) - MDN documentation

**Redux:**
- [Redux Documentation](https://redux.js.org/) - State management
- [Redux Toolkit](https://redux-toolkit.js.org/) - Modern Redux patterns

**Animation:**
- [FLIP Animation Technique](https://aerotwist.com/blog/flip-your-animations/) - Paul Lewis's article
- [Web Animations API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Animations_API) - MDN documentation

**Firebase:**
- [Firebase Documentation](https://firebase.google.com/docs) - Authentication and backend

### Package Statistics

**Framework Size:**
```
Core Package (boardgame/):          17,907 lines
Moves Package (moves/):             27,450 lines
Behaviors Package (behaviors/):        575 lines
Base Package (base/):                1,274 lines
Enum Package (enum/):                2,112 lines
Server API (server/api/):            2,448 lines
Total Core Framework:              ~51,766 lines
```

**Example Games:**
```
pig:                ~500 lines
tictactoe:          ~600 lines
memory:           ~1,000 lines
blackjack:        ~1,200 lines
checkers:         ~2,000 lines
debuganimations:  ~7,000 lines
Total Examples:  ~12,300 lines
```

**Web Frontend:**
```
26 Web Components
Redux state management
~10,000 lines HTML/JavaScript/CSS
```

**Total Project:**
```
~85,000+ lines of code
279 Go files
46 JavaScript files
6 complete example games
```

### Community & Support

**GitHub Repository:**
- [github.com/jkomoros/boardgame](https://github.com/jkomoros/boardgame)
- File issues, contribute, ask questions

**Issue Tracker:**
- [Known Issues](https://github.com/jkomoros/boardgame/issues)
- Report bugs
- Request features

### Version History

- **Pre-1.0:** Active development, hobby project
- **Go 1.13 Era:** Initial development (2019-2020)
- **Current:** Compiles with Go 1.25.6, still functional

### License

Check repository for license information.

---

## Conclusion

The boardgame framework is a **well-architected, production-quality** game engine that has aged gracefully despite being built before modern Go and JavaScript features were available.

### Current State

**Strengths:**
- ✅ Fully functional and compiles with modern Go (1.25.6)
- ✅ Comprehensive feature set (30+ move types, 9 behaviors, multiple storage backends)
- ✅ Excellent developer experience (AutoConfigurer, code generation, hot reload)
- ✅ Sophisticated animation system
- ✅ Well-documented (127KB tutorial, this architecture doc)
- ✅ 6 complete example games demonstrating patterns

**Limitations:**
- Pre-generics Go (relies on interface{} and code generation)
- Legacy frontend (Polymer 3, old lit-element)
- Old dependencies (2017-2018 era)
- Browser compatibility limited to Chrome primarily
- No schema migration system

### Modernization Opportunities

**Priority 1 (Low Risk, High Value):**
1. Update documentation ✅ (this document)
2. Fix browser compatibility bugs
3. Update Go to 1.22+
4. Update safe dependencies
5. Add test coverage

**Priority 2 (Medium Risk, High Value):**
1. Migrate to Lit 3 + TypeScript frontend
2. Update Firebase and other JavaScript dependencies
3. Fix animation sequencing issue

**Priority 3 (High Risk, High Value):**
1. Go generics migration (breaking changes)
2. Schema migration system
3. Complete Polymer removal

### Recommended Approach

**For Active Projects:** Start with Phase 1 (Stabilization) and Phase 2 (Dependencies)

**For New Projects:** Consider starting with modern stack (Lit 3 + TypeScript) but use existing backend

**For Learning:** Use as-is, it's fully functional and demonstrates excellent patterns

### Final Assessment

This framework demonstrates:
- Strong architectural principles
- Thoughtful API design
- Excellent separation of concerns
- Powerful abstractions (behaviors, moves, components)
- Production-ready reliability

While modernization would improve maintainability and developer experience, **the framework is fully usable in its current state** for building sophisticated board and card games.

The animation system, in particular, is a standout feature that provides smooth, professional-looking games with minimal developer effort.

---

**Document Version:** 1.0
**Last Updated:** 2026-02-03
**Framework Versions:** Go 1.13+ (compiles with 1.25.6), Polymer 3.3.0, lit-element 0.7.1

