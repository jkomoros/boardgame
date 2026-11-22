// Package stockgame is a game that exists only to make the real Go engine
// run the stock moves and behaviors the TypeScript engine ports in
// workstream 9 (rules/stock.ts), so ../stockvec can record what Go does with
// them. It is never registered with a server.
//
// It plays in three phases:
//
//   - Deal, an ordered run of fix-ups: every deal, collect and
//     move-components stock move, a ShuffleStack, then StartPhase.
//   - Play: players place cards on a board (a custom CurrentPlayer move),
//     two DefaultComponent fix-ups mark cards that reach certain slots (one
//     with a legal type, one without), FinishTurn passes the turn, Done is
//     a no-op any player may make, and a custom move enters Cleanup.
//   - Cleanup, an ordered run: DealAllComponents, CollectAllComponents, NoOp.
//
// Players embed behaviors.PlayerColor and cards behaviors.ComponentColor.
package stockgame

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

var Enums = enum.NewSet()

const (
	PhaseDeal = iota
	PhasePlay
	PhaseCleanup
)

var PhaseEnum = Enums.MustAdd("phase", map[enum.EnumKey]string{
	PhaseDeal:    "Deal",
	PhasePlay:    "Play",
	PhaseCleanup: "Cleanup",
})

var ColorEnum = Enums.MustAdd("color", map[enum.EnumKey]string{
	behaviors.ColorRed:   "Red",
	behaviors.ColorBlue:  "Blue",
	behaviors.ColorGreen: "Green",
})

const (
	LegalTypeNone = iota
	LegalTypeCorner
)

var LegalTypeEnum = Enums.MustAdd("legaltype", map[enum.EnumKey]string{
	LegalTypeNone:   "None",
	LegalTypeCorner: "Corner",
})

// NumCards is the deck size; BoardSize the board's slots.
const (
	NumCards  = 16
	BoardSize = 6
)

//boardgame:codegen reader
type card struct {
	base.ComponentValues
	behaviors.ComponentColor
	Rank int
}

//boardgame:codegen
type cardDynamic struct {
	base.SubState
	Crowned bool
	Corner  bool
}

//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.CurrentPlayerBehavior
	behaviors.PhaseBehavior
	behaviors.RoundRobin
	DrawStack    boardgame.Stack      `stack:"cards"`
	DiscardStack boardgame.Stack      `stack:"cards"`
	Spare        boardgame.Stack      `stack:"cards"`
	Board        boardgame.SizedStack `sizedstack:"cards,6"`
}

//boardgame:codegen
type playerState struct {
	base.SubState
	behaviors.PlayerColor
	Hand         boardgame.Stack `stack:"cards"`
	FinishedTurn bool
}

func (p *playerState) TurnDone() error {
	if !p.FinishedTurn {
		return errors.New("the player has not placed a card")
	}
	return nil
}

func (p *playerState) ResetForTurnStart() error {
	p.FinishedTurn = false
	return nil
}

func (p *playerState) ResetForTurnEnd() error {
	return nil
}

// boardSlot is the slot of the board holding c, or -1.
func boardSlot(state boardgame.ImmutableState, c boardgame.Component) int {
	game := state.ImmutableGameState().(*gameState)
	for i, other := range game.Board.ImmutableComponents() {
		if other != nil && other.DeckIndex() == c.DeckIndex() {
			return i
		}
	}
	return -1
}

// Legal implements interfaces.LegalComponent: a card on an even slot of the
// board that is not crowned; with the Corner legal type, a card in slot 0
// whose Corner is not set.
func (c *card) Legal(state boardgame.ImmutableState, legalType enum.ImmutableVal) error {
	slot := boardSlot(state, c.ContainingComponent())
	if slot < 0 {
		return errors.New("not on the board")
	}
	dyn := c.ContainingComponent().ImmutableInstance(state).ImmutableDynamicValues().(*cardDynamic)
	if legalType != nil && legalType.Value() == LegalTypeCorner {
		if slot != 0 {
			return errors.New("not in the corner")
		}
		if dyn.Corner {
			return errors.New("already marked as the corner")
		}
		return nil
	}
	if slot%2 != 0 {
		return errors.New("on an odd slot")
	}
	if dyn.Crowned {
		return errors.New("already crowned")
	}
	return nil
}

// movePlaceCard moves the first card of the current player's hand to Slot.
//
//boardgame:codegen
type movePlaceCard struct {
	moves.CurrentPlayer
	Slot int
}

func (m *movePlaceCard) DefaultsForState(state boardgame.ImmutableState) {
	m.CurrentPlayer.DefaultsForState(state)
	game := state.ImmutableGameState().(*gameState)
	for i, c := range game.Board.ImmutableComponents() {
		if c == nil {
			m.Slot = i
			return
		}
	}
}

func (m *movePlaceCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}
	p := state.ImmutablePlayerStates()[m.TargetPlayerIndex].(*playerState)
	if p.FinishedTurn {
		return errors.New("you already placed a card this turn")
	}
	first := p.Hand.ImmutableFirst()
	if first == nil {
		return errors.New("your hand is empty")
	}
	return first.MayMoveToSlot(state.ImmutableGameState().(*gameState).Board, m.Slot)
}

func (m *movePlaceCard) Apply(state boardgame.State) error {
	game := state.GameState().(*gameState)
	p := state.PlayerStates()[m.TargetPlayerIndex].(*playerState)
	if err := p.Hand.First().MoveTo(game.Board, m.Slot); err != nil {
		return err
	}
	p.FinishedTurn = true
	return nil
}

// moveCrownCard is a DefaultComponent over the board that crowns the card.
//
//boardgame:codegen
type moveCrownCard struct {
	moves.DefaultComponent
}

func (m *moveCrownCard) Apply(state boardgame.State) error {
	c := state.GameState().(*gameState).Board.ComponentAt(m.ComponentIndex)
	if c == nil {
		return errors.New("no card there")
	}
	c.DynamicValues().(*cardDynamic).Crowned = true
	return nil
}

// moveMarkCorner is a DefaultComponent with the Corner legal type.
//
//boardgame:codegen
type moveMarkCorner struct {
	moves.DefaultComponent
}

func (m *moveMarkCorner) Apply(state boardgame.State) error {
	c := state.GameState().(*gameState).Board.ComponentAt(m.ComponentIndex)
	if c == nil {
		return errors.New("no card there")
	}
	c.DynamicValues().(*cardDynamic).Corner = true
	return nil
}

// moveCallCleanup is a current-player move that enters Cleanup once the
// board is full.
//
//boardgame:codegen
type moveCallCleanup struct {
	moves.CurrentPlayer
}

func (m *moveCallCleanup) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}
	if state.ImmutableGameState().(*gameState).Board.SlotsRemaining() > 0 {
		return errors.New("the board is not full")
	}
	return nil
}

func (m *moveCallCleanup) Apply(state boardgame.State) error {
	state.GameState().(*gameState).SetCurrentPhase(PhaseCleanup)
	return nil
}

type Delegate struct {
	base.GameDelegate
}

func (g *Delegate) Name() string        { return "stockgame" }
func (g *Delegate) DisplayName() string { return "Stock Game" }
func (g *Delegate) Description() string {
	return "Runs the stock moves the TypeScript engine ports"
}
func (g *Delegate) MinNumPlayers() int     { return 2 }
func (g *Delegate) MaxNumPlayers() int     { return 3 }
func (g *Delegate) DefaultNumPlayers() int { return 2 }

func (g *Delegate) ConfigureEnums() *enum.Set { return Enums }

func (g *Delegate) ConfigureDecks() map[string]*boardgame.Deck {
	cards := boardgame.NewDeck()
	for i := 0; i < NumCards; i++ {
		c := &card{Rank: i + 1}
		c.Color = ColorEnum.MustNewVal(enum.EnumKey(i % 3))
		cards.AddComponent(c)
	}
	return map[string]*boardgame.Deck{"cards": cards}
}

func (g *Delegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *Delegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *Delegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	if deck.Name() == "cards" {
		return new(cardDynamic)
	}
	return nil
}

func (g *Delegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	return state.ImmutableGameState().(*gameState).DrawStack, nil
}

func (g *Delegate) FinishSetUp(state boardgame.State) error {
	for i, p := range state.PlayerStates() {
		if err := p.(*playerState).Color.SetValue(enum.EnumKey((i + 1) % 3)); err != nil {
			return err
		}
	}
	return state.GameState().(*gameState).DrawStack.Shuffle()
}

func (g *Delegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	game := state.ImmutableGameState().(*gameState)
	if game.Phase.Value() != PhaseCleanup {
		return false
	}
	if game.DrawStack.NumComponents() > 0 {
		return false
	}
	for _, p := range state.ImmutablePlayerStates() {
		if p.(*playerState).Hand.NumComponents() > 0 {
			return false
		}
	}
	return true
}

func (g *Delegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(g)
	return moves.Combine(
		moves.AddOrderedForPhase(PhaseDeal,
			auto.MustConfig(new(moves.DealCountComponents),
				moves.WithGameProperty("DrawStack"), moves.WithPlayerProperty("Hand"), moves.WithTargetCount(2)),
			auto.MustConfig(new(moves.DealComponentsUntilPlayerCountReached),
				moves.WithGameProperty("DrawStack"), moves.WithPlayerProperty("Hand"), moves.WithTargetCount(3)),
			auto.MustConfig(new(moves.CollectCountComponents),
				moves.WithGameProperty("DiscardStack"), moves.WithPlayerProperty("Hand")),
			auto.MustConfig(new(moves.CollectComponentsUntilPlayerCountLeft),
				moves.WithGameProperty("DiscardStack"), moves.WithPlayerProperty("Hand"), moves.WithTargetCount(1)),
			auto.MustConfig(new(moves.MoveCountComponents),
				moves.WithSourceProperty("DrawStack"), moves.WithDestinationProperty("Spare"), moves.WithTargetCount(2)),
			auto.MustConfig(new(moves.MoveComponentsUntilCountReached),
				moves.WithSourceProperty("DrawStack"), moves.WithDestinationProperty("Spare"), moves.WithTargetCount(4)),
			auto.MustConfig(new(moves.MoveComponentsUntilCountLeft),
				moves.WithSourceProperty("Spare"), moves.WithDestinationProperty("DiscardStack"), moves.WithTargetCount(1)),
			auto.MustConfig(new(moves.ShuffleStack),
				moves.WithSourceProperty("DiscardStack")),
			auto.MustConfig(new(moves.MoveAllComponents),
				moves.WithSourceProperty("DiscardStack"), moves.WithDestinationProperty("DrawStack")),
			auto.MustConfig(new(moves.DealComponentsUntilGameCountLeft),
				moves.WithGameProperty("DrawStack"), moves.WithPlayerProperty("Hand"), moves.WithTargetCount(7)),
			auto.MustConfig(new(moves.CollectComponentsUntilGameCountReached),
				moves.WithGameProperty("DrawStack"), moves.WithPlayerProperty("Hand"), moves.WithTargetCount(9)),
			auto.MustConfig(new(moves.StartPhase), moves.WithPhaseToStart(PhasePlay, PhaseEnum)),
		),
		moves.AddForPhase(PhasePlay,
			auto.MustConfig(new(moveCrownCard),
				moves.WithSourceProperty("Board"), moves.WithMoveName("Crown Card"), moves.WithHelpText("Crowns a card on an even slot")),
			auto.MustConfig(new(moveMarkCorner),
				moves.WithSourceProperty("Board"), moves.WithLegalType(LegalTypeCorner, LegalTypeEnum),
				moves.WithMoveName("Mark Corner"), moves.WithHelpText("Marks the card in the corner")),
			auto.MustConfig(new(moves.FinishTurn)),
			auto.MustConfig(new(movePlaceCard), moves.WithHelpText("Places the first card of your hand on the board")),
			auto.MustConfig(new(moveCallCleanup), moves.WithHelpText("Ends play once the board is full")),
			auto.MustConfig(new(moves.Done), moves.WithMoveName("Pass"), moves.WithHelpText("Does nothing")),
		),
		moves.AddOrderedForPhase(PhaseCleanup,
			auto.MustConfig(new(moves.DealAllComponents),
				moves.WithGameProperty("DrawStack"), moves.WithPlayerProperty("Hand")),
			auto.MustConfig(new(moves.CollectAllComponents),
				moves.WithGameProperty("DiscardStack"), moves.WithPlayerProperty("Hand")),
			auto.MustConfig(new(moves.NoOp)),
		),
	)
}

// NewDelegate returns the game's delegate.
func NewDelegate() boardgame.GameDelegate {
	return &Delegate{}
}
