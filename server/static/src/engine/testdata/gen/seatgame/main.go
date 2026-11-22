// Package seatgame is a game that exists only to make the real Go engine run
// the seat-management stock moves the TypeScript engine ports in workstream
// 9 (rules/stock.ts), so ../stockvec can record what Go does with them. It
// is never registered with a server.
//
// With players to seat it goes:
//
//   - Gathering: InactivateEmptySeat marks every empty seat inactive;
//     SeatPlayer (any phase) seats a joiner, and ActivateFilledSeat (any
//     phase) activates them; once two are seated a player proposes
//     CloseAllSeats, after which the custom fix-up Begin enters Open.
//   - Open: ActivateEmptySeat re-activates the closed empty seats, the
//     custom fix-up Reopen opens them again once, CloseEmptySeat closes each
//     again, and the custom fix-up Finish enters Done.
//   - Done: the game is over.
package seatgame

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
	PhaseGathering = iota
	PhaseOpen
	PhaseDone
)

var PhaseEnum = Enums.MustAdd("phase", map[enum.EnumKey]string{
	PhaseGathering: "Gathering",
	PhaseOpen:      "Open",
	PhaseDone:      "Done",
})

//boardgame:codegen reader
type token struct {
	base.ComponentValues
}

//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.CurrentPlayerBehavior
	behaviors.PhaseBehavior
	Tokens   boardgame.Stack `stack:"tokens"`
	Reopened bool
}

//boardgame:codegen
type playerState struct {
	base.SubState
	behaviors.Seat
	behaviors.InactivePlayer
}

func everySeatFilledOrClosed(state boardgame.ImmutableState) bool {
	for _, p := range state.ImmutablePlayerStates() {
		ps := p.(*playerState)
		if !ps.SeatFilled && !ps.SeatClosed {
			return false
		}
	}
	return true
}

// moveBegin enters Open once no seat is open.
//
//boardgame:codegen
type moveBegin struct {
	moves.FixUp
}

func (m *moveBegin) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	if !everySeatFilledOrClosed(state) {
		return errors.New("a seat is still open")
	}
	return nil
}

func (m *moveBegin) Apply(state boardgame.State) error {
	state.GameState().(*gameState).SetCurrentPhase(PhaseOpen)
	return nil
}

// moveReopen opens every empty seat again, once, after none is inactive.
//
//boardgame:codegen
type moveReopen struct {
	moves.FixUp
}

func (m *moveReopen) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	if state.ImmutableGameState().(*gameState).Reopened {
		return errors.New("already reopened")
	}
	for _, p := range state.ImmutablePlayerStates() {
		if p.(*playerState).PlayerInactive {
			return errors.New("a seat is still inactive")
		}
	}
	return nil
}

func (m *moveReopen) Apply(state boardgame.State) error {
	game := state.GameState().(*gameState)
	game.Reopened = true
	for _, p := range state.PlayerStates() {
		ps := p.(*playerState)
		if !ps.SeatFilled {
			ps.SeatClosed = false
		}
	}
	return nil
}

// moveFinish enters Done once the seats were reopened and closed again.
//
//boardgame:codegen
type moveFinish struct {
	moves.FixUp
}

func (m *moveFinish) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	if !state.ImmutableGameState().(*gameState).Reopened {
		return errors.New("not reopened yet")
	}
	if !everySeatFilledOrClosed(state) {
		return errors.New("a seat is still open")
	}
	return nil
}

func (m *moveFinish) Apply(state boardgame.State) error {
	state.GameState().(*gameState).SetCurrentPhase(PhaseDone)
	return nil
}

type Delegate struct {
	base.GameDelegate
}

func (g *Delegate) Name() string        { return "seatgame" }
func (g *Delegate) DisplayName() string { return "Seat Game" }
func (g *Delegate) Description() string {
	return "Runs the seat stock moves the TypeScript engine ports"
}
func (g *Delegate) MinNumPlayers() int     { return 2 }
func (g *Delegate) MaxNumPlayers() int     { return 4 }
func (g *Delegate) DefaultNumPlayers() int { return 4 }

func (g *Delegate) ConfigureEnums() *enum.Set { return Enums }

func (g *Delegate) ConfigureDecks() map[string]*boardgame.Deck {
	tokens := boardgame.NewDeck()
	for i := 0; i < 2; i++ {
		tokens.AddComponent(&token{})
	}
	return map[string]*boardgame.Deck{"tokens": tokens}
}

func (g *Delegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *Delegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *Delegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	return state.ImmutableGameState().(*gameState).Tokens, nil
}

func (g *Delegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	return state.ImmutableGameState().(*gameState).Phase.Value() == PhaseDone
}

func (g *Delegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(g)
	return moves.Combine(
		moves.Add(
			auto.MustConfig(new(moves.SeatPlayer)),
			auto.MustConfig(new(moves.ActivateFilledSeat)),
		),
		moves.AddForPhase(PhaseGathering,
			auto.MustConfig(new(moves.InactivateEmptySeat)),
			auto.MustConfig(new(moves.CloseAllSeats), moves.WithTargetCount(2)),
			auto.MustConfig(new(moveBegin), moves.WithMoveName("Begin"), moves.WithHelpText("Enters Open once no seat is open")),
		),
		moves.AddForPhase(PhaseOpen,
			auto.MustConfig(new(moves.ActivateEmptySeat)),
			auto.MustConfig(new(moves.CloseEmptySeat)),
			auto.MustConfig(new(moveReopen), moves.WithMoveName("Reopen"), moves.WithHelpText("Opens the empty seats again")),
			auto.MustConfig(new(moveFinish), moves.WithMoveName("Finish"), moves.WithHelpText("Enters Done")),
		),
	)
}

// NewDelegate returns the game's delegate.
func NewDelegate() boardgame.GameDelegate {
	return &Delegate{}
}
