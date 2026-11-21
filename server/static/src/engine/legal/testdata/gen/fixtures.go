package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	blackjackgame "github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	memorygame "github.com/jkomoros/boardgame/examples/memory"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

/*
The fixture builders below are a faithful port of legalFixtureBuilders in
legal/conformance_test.go. They cannot be imported (they live in a _test.go
file of an external test package), so they are restated here. Two things keep
the restatement honest:

  - corpusVectors (vectors.go) runs every case in legal/testdata/conformance
    against these fixtures with the real Go catalog and refuses to emit
    anything if a single case disagrees with the corpus' own expectation -
    the same assertion TestConformanceCorpus makes.
  - Every game is created with RecreateGame from a fixed id and secret salt,
    so shuffles, deals and component ids are identical on every run and the
    checked-in JSON only changes when the Go semantics do.
*/

// fixture is what a vector needs to build a legal.Context, plus the game it
// came from (for the chest and for delegate lookups at export time).
type fixture struct {
	gameKey string
	game    *boardgame.Game
	state   boardgame.State
	move    boardgame.Move
}

func must(err error, context string) {
	if err != nil {
		panic(fmt.Sprintf("gen: %s: %v", context, err))
	}
}

const fixtureGameID = "LEGALFIXTUREGAME"
const fixtureSecretSalt = "legal-fixture-salt"

func newGame(gameKey string, delegate boardgame.GameDelegate) (*boardgame.Game, boardgame.State) {
	manager, err := boardgame.NewGameManager(delegate, storagememory.NewStorageManager())
	must(err, "building "+gameKey+" manager")
	game, err := manager.Internals().RecreateGame(&boardgame.GameStorageRecord{
		Name:       delegate.Name(),
		ID:         fixtureGameID,
		SecretSalt: fixtureSecretSalt,
	})
	must(err, "building "+gameKey+" game")
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		panic("gen: " + gameKey + " CurrentState() was not mutable")
	}
	return game, state
}

func newMemoryGame() (*boardgame.Game, boardgame.State) {
	return newGame("memory", memorygame.NewDelegate())
}

func newCheckersGame() (*boardgame.Game, boardgame.State) {
	return newGame("checkers", checkers.NewDelegate())
}

func newBlackjackGame() (*boardgame.Game, boardgame.State) {
	return newGame("blackjack", blackjackgame.NewDelegate())
}

func exampleMoveWithProp(game *boardgame.Game, prop string) boardgame.Move {
	for _, mv := range game.Manager().ExampleMoves() {
		if _, ok := mv.Reader().Props()[prop]; ok {
			return mv
		}
	}
	panic("gen: no move with a " + prop + " property found")
}

func memoryMoveWithCardIndex(game *boardgame.Game, cardIndex int) boardgame.Move {
	mv := exampleMoveWithProp(game, "CardIndex")
	must(mv.ReadSetter().SetIntProp("CardIndex", cardIndex), "setting CardIndex")
	return mv
}

func checkersMoveWithTokenIndex(game *boardgame.Game, key int) boardgame.Move {
	mv := exampleMoveWithProp(game, "TokenIndexToMove")
	val, err := mv.ReadSetter().EnumProp("TokenIndexToMove")
	must(err, "reading TokenIndexToMove")
	must(val.SetValue(enum.EnumKey(key)), "setting TokenIndexToMove")
	return mv
}

func checkersMoveWithIndexes(game *boardgame.Game, tokenIndex, spaceIndex int) boardgame.Move {
	mv := checkersMoveWithTokenIndex(game, tokenIndex)
	space, err := mv.ReadSetter().EnumProp("SpaceIndex")
	must(err, "reading SpaceIndex")
	must(space.SetValue(enum.EnumKey(spaceIndex)), "setting SpaceIndex")
	return mv
}

func firstOccupiedIndex(stack boardgame.ImmutableStack) int {
	for i := 0; i < stack.Len(); i++ {
		if stack.ImmutableComponentAt(i) != nil {
			return i
		}
	}
	return -1
}

func firstEmptyIndex(stack boardgame.ImmutableStack) int {
	for i := 0; i < stack.Len(); i++ {
		if stack.ImmutableComponentAt(i) == nil {
			return i
		}
	}
	return -1
}

func gameStack(state boardgame.State, name string) boardgame.Stack {
	stack, err := state.GameState().ReadSetter().StackProp(name)
	must(err, "reading "+name)
	return stack
}

func firstSpaceWithColor(spaces boardgame.ImmutableStack, currentPlayerColor enum.ImmutableVal, wantMatch bool) int {
	for i := 0; i < spaces.Len(); i++ {
		c := spaces.ImmutableComponentAt(i)
		if c == nil {
			continue
		}
		tokenColor, err := c.Values().Reader().ImmutableEnumProp("Color")
		must(err, "reading token Color")
		if tokenColor.Equals(currentPlayerColor) == wantMatch {
			return i
		}
	}
	return -1
}

func memoryVisiblePair(wantMatch bool) fixture {
	game, state := newMemoryGame()
	hidden := gameStack(state, "HiddenCards")
	visible := gameStack(state, "VisibleCards")
	leftIndex, rightIndex := -1, -1
	for i := 0; i < hidden.Len() && leftIndex < 0; i++ {
		left := hidden.ComponentAt(i)
		if left == nil {
			continue
		}
		leftType, err := left.Values().Reader().StringProp("Type")
		must(err, "reading left memory Type")
		for j := i + 1; j < hidden.Len(); j++ {
			right := hidden.ComponentAt(j)
			if right == nil {
				continue
			}
			rightType, err := right.Values().Reader().StringProp("Type")
			must(err, "reading right memory Type")
			if (leftType == rightType) == wantMatch {
				leftIndex, rightIndex = i, j
				break
			}
		}
	}
	if leftIndex < 0 {
		panic("gen: no memory pair found")
	}
	left, right := hidden.ComponentAt(leftIndex), hidden.ComponentAt(rightIndex)
	must(left.MoveTo(visible, 3), "moving first memory card")
	must(right.MoveTo(visible, 7), "moving second memory card")
	return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 7)}
}

func checkersTokenFixture(wantOwn bool) fixture {
	game, state := newCheckersGame()
	spaces := gameStack(state, "Spaces")
	playerColor, err := state.ImmutableCurrentPlayer().Reader().ImmutableEnumProp("Color")
	must(err, "reading current player's Color")
	idx := firstSpaceWithColor(spaces, playerColor, wantOwn)
	if idx < 0 {
		panic("gen: checkers fixture has no matching space")
	}
	return fixture{gameKey: "checkers", game: game, state: state, move: checkersMoveWithTokenIndex(game, idx)}
}

func blackjackFixture(unfinishedPlayerZero, inactivePlayerZero bool) fixture {
	game, state := newBlackjackGame()
	for i, p := range state.PlayerStates() {
		rs := p.ReadSetter()
		if i == 0 && unfinishedPlayerZero {
			must(rs.SetBoolProp("Stood", false), "setting Stood")
			if inactivePlayerZero {
				must(rs.SetBoolProp("PlayerInactive", true), "setting PlayerInactive")
			}
			continue
		}
		must(rs.SetBoolProp("Stood", true), "setting Stood")
	}
	return fixture{gameKey: "blackjack", game: game, state: state}
}

// fixtureBuilders mirrors legalFixtureBuilders (legal/conformance_test.go)
// name for name. See that file for what each arrangement is for.
var fixtureBuilders = map[string]func() fixture{
	"memoryDefault": func() fixture {
		game, state := newMemoryGame()
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 0)}
	},
	"memoryZeroCardsLeft": func() fixture {
		game, state := newMemoryGame()
		must(state.CurrentPlayer().ReadSetter().SetIntProp("CardsLeftToReveal", 0), "setting CardsLeftToReveal")
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 0)}
	},
	"memorySeatFilled": func() fixture {
		game, state := newMemoryGame()
		must(state.CurrentPlayer().ReadSetter().SetBoolProp("SeatFilled", true), "setting SeatFilled")
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 0)}
	},
	"memoryVisibleOccupied": func() fixture {
		game, state := newMemoryGame()
		hidden := gameStack(state, "HiddenCards")
		visible := gameStack(state, "VisibleCards")
		must(hidden.ComponentAt(hidden.Len()-1).MoveTo(visible, 0), "moving a hidden card into VisibleCards[0]")
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 0)}
	},
	"memoryVisibleMatchingSparse":  func() fixture { return memoryVisiblePair(true) },
	"memoryVisibleDifferentSparse": func() fixture { return memoryVisiblePair(false) },
	"memoryNoMove": func() fixture {
		game, state := newMemoryGame()
		return fixture{gameKey: "memory", game: game, state: state}
	},
	"checkersDefault": func() fixture {
		game, state := newCheckersGame()
		idx := firstOccupiedIndex(gameStack(state, "Spaces"))
		if idx < 0 {
			panic("gen: checkers fixture has no occupied space")
		}
		return fixture{gameKey: "checkers", game: game, state: state, move: checkersMoveWithTokenIndex(game, idx)}
	},
	"checkersEmptySpace": func() fixture {
		game, state := newCheckersGame()
		idx := firstEmptyIndex(gameStack(state, "Spaces"))
		if idx < 0 {
			panic("gen: checkers fixture has no empty space")
		}
		return fixture{gameKey: "checkers", game: game, state: state, move: checkersMoveWithTokenIndex(game, idx)}
	},
	"checkersNoMove": func() fixture {
		game, state := newCheckersGame()
		return fixture{gameKey: "checkers", game: game, state: state}
	},
	"checkersDistinctSpaces": func() fixture {
		game, state := newCheckersGame()
		spaces := gameStack(state, "Spaces")
		first, second := firstOccupiedIndex(spaces), firstEmptyIndex(spaces)
		if first < 0 || second < 0 || first == second {
			panic("gen: checkers fixture does not have two distinct valid spaces")
		}
		return fixture{gameKey: "checkers", game: game, state: state, move: checkersMoveWithIndexes(game, first, second)}
	},
	"checkersOwnToken":      func() fixture { return checkersTokenFixture(true) },
	"checkersOpponentToken": func() fixture { return checkersTokenFixture(false) },
	"checkersBlackFirst": func() fixture {
		game, state := newCheckersGame()
		spaces := gameStack(state, "Spaces")
		index := -1
		first := firstOccupiedIndex(spaces)
		for i := 0; i < spaces.Len(); i++ {
			component := spaces.ComponentAt(i)
			if component == nil {
				continue
			}
			color, err := component.Values().Reader().ImmutableEnumProp("Color")
			must(err, "reading token Color")
			if color.String() == "Black" {
				index = i
				break
			}
		}
		if index < 0 {
			panic("gen: no black checker token found")
		}
		if index != first {
			must(spaces.SwapComponents(index, first), "moving black checker to first occupied slot")
		}
		return fixture{gameKey: "checkers", game: game, state: state, move: checkersMoveWithTokenIndex(game, first)}
	},
	"memoryCardAlreadyRevealed": func() fixture {
		game, state := newMemoryGame()
		must(gameStack(state, "HiddenCards").ComponentAt(0).MoveTo(gameStack(state, "VisibleCards"), 0), "moving HiddenCards[0] to VisibleCards[0]")
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 0)}
	},
	"memoryCardNeverThere": func() fixture {
		game, state := newMemoryGame()
		must(gameStack(state, "HiddenCards").ComponentAt(0).MoveTo(gameStack(state, "VisibleCards"), 5), "moving HiddenCards[0] to VisibleCards[5]")
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 0)}
	},
	"memoryTargetPlayerOne": func() fixture {
		game, state := newMemoryGame()
		move := memoryMoveWithCardIndex(game, 0)
		must(move.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", boardgame.PlayerIndex(1)), "setting TargetPlayerIndex")
		return fixture{gameKey: "memory", game: game, state: state, move: move}
	},
	"memoryTargetObserver": func() fixture {
		game, state := newMemoryGame()
		move := memoryMoveWithCardIndex(game, 0)
		must(move.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", boardgame.ObserverPlayerIndex), "setting TargetPlayerIndex")
		return fixture{gameKey: "memory", game: game, state: state, move: move}
	},
	"blackjackAllFinished":     func() fixture { return blackjackFixture(false, false) },
	"blackjackOneUnfinished":   func() fixture { return blackjackFixture(true, false) },
	"blackjackInactiveSkipped": func() fixture { return blackjackFixture(true, true) },
}

// extraFixtureBuilders are arrangements the corpus does not have, added for
// the generated vectors in vectors.go. Names are prefixed "x" so they can
// never collide with a corpus fixture name.
var extraFixtureBuilders = map[string]func() fixture{
	// CardIndex far outside both stacks, and a negative one: every
	// index-taking predicate has an out-of-range branch.
	"xMemoryCardIndexHuge": func() fixture {
		game, state := newMemoryGame()
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 4000)}
	},
	"xMemoryCardIndexNegative": func() fixture {
		game, state := newMemoryGame()
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, -1)}
	},
	// The move targets a seat that does not exist: EnsureValid walks forward
	// from an out-of-bounds index.
	"xMemoryTargetOutOfBounds": func() fixture {
		game, state := newMemoryGame()
		move := memoryMoveWithCardIndex(game, 0)
		must(move.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", boardgame.PlayerIndex(7)), "setting TargetPlayerIndex")
		return fixture{gameKey: "memory", game: game, state: state, move: move}
	},
	"xMemoryTargetAdmin": func() fixture {
		game, state := newMemoryGame()
		move := memoryMoveWithCardIndex(game, 0)
		must(move.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", boardgame.AdminPlayerIndex), "setting TargetPlayerIndex")
		return fixture{gameKey: "memory", game: game, state: state, move: move}
	},
	"xMemoryTargetAny": func() fixture {
		game, state := newMemoryGame()
		move := memoryMoveWithCardIndex(game, 0)
		must(move.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", boardgame.AnyPlayerIndex), "setting TargetPlayerIndex")
		return fixture{gameKey: "memory", game: game, state: state, move: move}
	},
	// Player 1 is inactive and is the move's target: EnsureValid skips them.
	"xMemoryPlayerOneInactiveTargeted": func() fixture {
		game, state := newMemoryGame()
		must(state.PlayerStates()[1].ReadSetter().SetBoolProp("PlayerInactive", true), "setting PlayerInactive")
		move := memoryMoveWithCardIndex(game, 0)
		must(move.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", boardgame.PlayerIndex(1)), "setting TargetPlayerIndex")
		return fixture{gameKey: "memory", game: game, state: state, move: move}
	},
	// VisibleCards completely full, HiddenCards empty: no slots remaining.
	"xMemoryAllRevealed": func() fixture {
		game, state := newMemoryGame()
		hidden := gameStack(state, "HiddenCards")
		visible := gameStack(state, "VisibleCards")
		for i := 0; i < hidden.Len(); i++ {
			if c := hidden.ComponentAt(i); c != nil {
				must(c.MoveTo(visible, i), "revealing a card")
			}
		}
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 0)}
	},
	// Player 0 has won two cards, so a growable player stack is non-empty and
	// the players differ.
	"xMemoryPlayerZeroWonCards": func() fixture {
		game, state := newMemoryGame()
		hidden := gameStack(state, "HiddenCards")
		won, err := state.PlayerStates()[0].ReadSetter().StackProp("WonCards")
		must(err, "reading WonCards")
		must(hidden.ComponentAt(0).MoveToNextSlot(won), "winning a card")
		must(hidden.ComponentAt(1).MoveToNextSlot(won), "winning a card")
		must(state.PlayerStates()[1].ReadSetter().SetIntProp("CardsLeftToReveal", 1), "setting CardsLeftToReveal")
		return fixture{gameKey: "memory", game: game, state: state, move: memoryMoveWithCardIndex(game, 2)}
	},
	// Checkers with player 1 holding one captured token (a constrained
	// growable stack) and the move pointing at an occupied and an empty space.
	"xCheckersCaptured": func() fixture {
		game, state := newCheckersGame()
		spaces := gameStack(state, "Spaces")
		first, second := firstOccupiedIndex(spaces), firstEmptyIndex(spaces)
		captured, err := state.PlayerStates()[1].ReadSetter().StackProp("CapturedTokens")
		must(err, "reading CapturedTokens")
		must(spaces.ComponentAt(first).MoveToNextSlot(captured), "capturing a token")
		return fixture{gameKey: "checkers", game: game, state: state, move: checkersMoveWithIndexes(game, firstOccupiedIndex(spaces), second)}
	},
	"xBlackjackDefault": func() fixture {
		game, state := newBlackjackGame()
		return fixture{gameKey: "blackjack", game: game, state: state}
	},
	"xBlackjackPlayerOneEliminated": func() fixture {
		game, state := newBlackjackGame()
		must(state.PlayerStates()[1].ReadSetter().SetBoolProp("Eliminated", true), "setting Eliminated")
		return fixture{gameKey: "blackjack", game: game, state: state}
	},
	"xBlackjackEveryoneInactive": func() fixture {
		game, state := newBlackjackGame()
		for _, p := range state.PlayerStates() {
			must(p.ReadSetter().SetBoolProp("PlayerInactive", true), "setting PlayerInactive")
		}
		return fixture{gameKey: "blackjack", game: game, state: state}
	},
}

/*
Export. A fixture is exported in the shapes the server already ships -
state.StorageRecord() and json.Marshal(chest) - plus the things a client is
told out of band and a state record cannot carry: property types, which enum
an enum property belongs to, what kind of stack a stack is, and the
delegate's answers (current player, who may be active, the phase).
*/

type exportedProp struct {
	Type string `json:"type"`
	// Enum is the owning enum's name, for TypeEnum properties.
	Enum string `json:"enum,omitempty"`
	// Value is only exported for move properties (state values are in State).
	Value interface{} `json:"value,omitempty"`
}

type exportedStack struct {
	// Kind is "growable", "sized" or "merged".
	Kind string `json:"kind"`
	// Constrained marks a stack declared with stack constraints. The Go API
	// does not expose that, so it is restated from the example's struct tags
	// in constrainedStacks below.
	Constrained bool `json:"constrained,omitempty"`
}

type exportedSubState struct {
	Props  map[string]exportedProp  `json:"props"`
	Stacks map[string]exportedStack `json:"stacks,omitempty"`
}

type exportedPhase struct {
	Key  int    `json:"key"`
	Name string `json:"name"`
	// Ancestors is the self-inclusive TreeEnum ancestor list, or null when
	// the phase enum is not a tree.
	Ancestors []int `json:"ancestors"`
}

type exportedMove struct {
	Name  string                  `json:"name"`
	Props map[string]exportedProp `json:"props"`
}

type exportedFixture struct {
	Game               string             `json:"game"`
	CurrentPlayerIndex int                `json:"currentPlayerIndex"`
	PlayersMayBeActive []bool             `json:"playersMayBeActive"`
	PlayersInactive    []bool             `json:"playersInactive"`
	CustomPlayerOrder  []int              `json:"customPlayerOrder"`
	Phase              *exportedPhase     `json:"phase"`
	GameState          exportedSubState   `json:"gameState"`
	PlayerStates       []exportedSubState `json:"playerStates"`
	Move               *exportedMove      `json:"move"`
	State              json.RawMessage    `json:"state"`
}

type exportedGame struct {
	// Chest is json.Marshal(manager.Chest()), exactly as /info ships it.
	Chest json.RawMessage `json:"chest"`
	// OtherEnums holds the enums a property of this game uses that the chest
	// does NOT ship (blackjack's card Rank and Suit live in package
	// playingcards and are never added to the chest's enum set): enum name to
	// key to display string. A fixture cannot be rebuilt without them.
	OtherEnums map[string]map[string]string `json:"otherEnums"`
	// ComponentProps is each deck's component Values() property table.
	ComponentProps map[string]map[string]exportedProp `json:"componentProps"`
}

// constrainedStacks restates which example stacks carry constraints
// (`stack:"Tokens,same(Color)"` in examples/checkers/state.go).
var constrainedStacks = map[string]bool{
	"checkers/player/CapturedTokens": true,
}

// enumsSeen records every enum reached through an exported property, per
// game, so exportGame can ship the ones the chest leaves out.
var enumsSeen = map[string]map[string]enum.Enum{}

// exportingGame is the game key exportProps files enums under.
var exportingGame string

func exportProps(reader boardgame.PropertyReader, withValues bool) map[string]exportedProp {
	out := map[string]exportedProp{}
	for name, propType := range reader.Props() {
		prop := exportedProp{Type: propType.String()}
		if propType == boardgame.TypeEnum {
			val, err := reader.ImmutableEnumProp(name)
			if err == nil && val != nil && val.Enum() != nil {
				prop.Enum = val.Enum().Name()
				if enumsSeen[exportingGame] == nil {
					enumsSeen[exportingGame] = map[string]enum.Enum{}
				}
				enumsSeen[exportingGame][prop.Enum] = val.Enum()
			}
		}
		if withValues {
			switch propType {
			case boardgame.TypeInt:
				v, _ := reader.IntProp(name)
				prop.Value = v
			case boardgame.TypeBool:
				v, _ := reader.BoolProp(name)
				prop.Value = v
			case boardgame.TypeString:
				v, _ := reader.StringProp(name)
				prop.Value = v
			case boardgame.TypePlayerIndex:
				v, _ := reader.PlayerIndexProp(name)
				prop.Value = int(v)
			case boardgame.TypeEnum:
				v, err := reader.ImmutableEnumProp(name)
				if err == nil && v != nil {
					prop.Value = v.Value().Int()
				}
			}
		}
		out[name] = prop
	}
	return out
}

func exportSubState(gameKey, group string, reader boardgame.PropertyReader) exportedSubState {
	out := exportedSubState{Props: exportProps(reader, false)}
	for name, propType := range reader.Props() {
		if propType != boardgame.TypeStack {
			continue
		}
		stack, err := reader.ImmutableStackProp(name)
		if err != nil || stack == nil {
			continue
		}
		kind := "growable"
		if _, merged := stack.(boardgame.MergedStack); merged {
			kind = "merged"
		} else if stack.ImmutableSizedStack() != nil {
			kind = "sized"
		}
		if out.Stacks == nil {
			out.Stacks = map[string]exportedStack{}
		}
		out.Stacks[name] = exportedStack{Kind: kind, Constrained: constrainedStacks[gameKey+"/"+group+"/"+name]}
	}
	return out
}

func exportPhase(state boardgame.ImmutableState) *exportedPhase {
	val := state.Manager().Delegate().CurrentPhase(state)
	if val == nil {
		return nil
	}
	out := &exportedPhase{Key: val.Value().Int(), Name: val.String()}
	if e := val.Enum(); e != nil {
		if tree := e.TreeEnum(); tree != nil {
			out.Ancestors = []int{}
			for _, key := range tree.Ancestors(val.Value()) {
				out.Ancestors = append(out.Ancestors, key.Int())
			}
		}
	}
	return out
}

func exportMove(move boardgame.Move) *exportedMove {
	if move == nil {
		return nil
	}
	return &exportedMove{Name: move.Info().Name(), Props: exportProps(move.Reader(), true)}
}

func exportFixture(f fixture) exportedFixture {
	return exportFixtureWith(f.gameKey, f.state, f.move)
}

func exportFixtureWith(gameKey string, state boardgame.ImmutableState, move boardgame.Move) exportedFixture {
	exportingGame = gameKey
	delegate := state.Manager().Delegate()
	out := exportedFixture{
		Game:               gameKey,
		CurrentPlayerIndex: int(state.CurrentPlayerIndex()),
		Phase:              exportPhase(state),
		GameState:          exportSubState(gameKey, "game", state.ImmutableGameState().Reader()),
		Move:               exportMove(move),
		State:              stripStackIDs(state.StorageRecord()),
	}
	for _, p := range state.ImmutablePlayerStates() {
		out.PlayersMayBeActive = append(out.PlayersMayBeActive, delegate.PlayerMayBeActive(p))
		out.PlayersInactive = append(out.PlayersInactive, behaviors.PlayerIsInactive(p))
		out.PlayerStates = append(out.PlayerStates, exportSubState(gameKey, "player", p.Reader()))
	}
	if order := delegate.CustomPlayerOrder(state); order != nil {
		out.CustomPlayerOrder = []int{}
		for _, p := range order {
			out.CustomPlayerOrder = append(out.CustomPlayerOrder, int(p))
		}
	}
	return out
}

// exportGame must run after every fixture of gameKey has been exported, so
// that enumsSeen is complete.
func exportGame(gameKey string, manager *boardgame.GameManager) exportedGame {
	exportingGame = gameKey
	chest, err := json.Marshal(manager.Chest())
	must(err, "marshaling chest")
	out := exportedGame{Chest: chest, ComponentProps: map[string]map[string]exportedProp{}, OtherEnums: map[string]map[string]string{}}
	defer func() {
		for name, e := range enumsSeen[gameKey] {
			if manager.Chest().Enums() != nil && manager.Chest().Enums().Enum(name) != nil {
				continue
			}
			values := map[string]string{}
			for _, key := range e.Values() {
				values[fmt.Sprint(key.Int())] = e.String(key)
			}
			out.OtherEnums[name] = values
		}
	}()
	for _, name := range manager.Chest().DeckNames() {
		deck := manager.Chest().Deck(name)
		components := deck.Components()
		if len(components) == 0 || components[0].Values() == nil || components[0].Values().Reader() == nil {
			out.ComponentProps[name] = map[string]exportedProp{}
			continue
		}
		out.ComponentProps[name] = exportProps(components[0].Values().Reader(), false)
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// stripStackIDs removes IDs and IDsLastSeen from every stack in a state
// record. Legality never reads component ids, and they are nine tenths of a
// record's bytes; everything else is exactly what state.StorageRecord() gave.
func stripStackIDs(record []byte) json.RawMessage {
	var tree interface{}
	must(json.Unmarshal(record, &tree), "parsing state record")
	var walk func(node interface{})
	walk = func(node interface{}) {
		switch n := node.(type) {
		case map[string]interface{}:
			if _, isStack := n["Indexes"]; isStack {
				if _, hasDeck := n["Deck"]; hasDeck {
					delete(n, "IDs")
					delete(n, "IDsLastSeen")
				}
			}
			for _, child := range n {
				walk(child)
			}
		case []interface{}:
			for _, child := range n {
				walk(child)
			}
		}
	}
	walk(tree)
	return json.RawMessage(marshal(tree))
}
