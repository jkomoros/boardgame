// Package staticgame is a fixture for the static half: a small game whose
// shape exercises each rule of Static (roles, erasure, callee-refused
// propagation, buckets, stock ancestors, data-only moves, DefaultRoundSetup
// expansion, traits, literal units, generated files). It is type-checked,
// never run.
package staticgame

import (
	"errors"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

type gameDelegate struct {
	base.GameDelegate
}

// NewDelegate returns the fixture's delegate.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}

var moveName string

//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.CurrentPlayerBehavior
	behaviors.PhaseBehavior
	behaviors.RoundRobin
	behaviors.DrawDiscardPair `draw:"Cards" discard:"Discards"`
	Cards                     boardgame.Stack `stack:"cards"`
	Discards                  boardgame.Stack `stack:"cards"`
	Count                     int
}

//boardgame:codegen
type playerState struct {
	base.SubState
	behaviors.Seat
	behaviors.ScoreBehavior
	Done bool
}

//boardgame:codegen
type moveData struct {
	moves.NoOp
}

//boardgame:codegen
type moveCode struct {
	moves.CurrentPlayer
}

//boardgame:codegen
type moveNested struct {
	moveCode
}

//boardgame:codegen
type moveGap struct {
	moves.DefaultComponent
}

func (g *gameDelegate) Name() string {
	return "staticgame"
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return &playerState{}
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(g)
	return moves.Combine(
		moves.Add(
			auto.MustConfig(new(moveData), moves.WithMoveName("Data Move")),
			auto.MustConfig(new(moveCode),
				moves.WithLegalPreconditions(
					legal.StackNotEmpty("game.Cards"),
					legal.StackEmpty("game.Discards"),
				),
			),
			auto.MustConfig(new(moveNested), moves.WithMoveNameSuffix("Nested")),
			auto.MustConfig(new(moveGap)),
		),
		moves.AddOrderedForPhase(0,
			moves.DefaultRoundSetup(auto),
			auto.MustConfig(new(moves.StartPhase)),
		),
		moves.AddOrderedForPhase(1,
			moves.Repeat(func(currentCount, length int) error {
				return nil
			}, auto.MustConfig(new(moves.NoOp), moves.WithMoveName("Counted"))),
			auto.MustConfig(new(moves.StartPhase), moves.WithMoveName("Second Start")),
		),
	)
}

func (g *gameDelegate) ConfigureComputedProperties() []boardgame.ComputedProperty {
	return []boardgame.ComputedProperty{
		boardgame.GlobalComputedBool("Refused", func(state boardgame.ImmutableState) bool {
			return readsMoveName() != ""
		}),
		boardgame.GlobalComputedInt("Gapped", gappedCount),
	}
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, _ := concreteStates(state)
	return game.Cards, nil
}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game, players := concreteStates(state)
	game.Count = len(players)
	return nil
}

func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {
	return strings.Join(diagramLines(), "\n")
}

func (g *gameDelegate) DefaultNumPlayeres() int {
	return 2
}

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)
	players := make([]*playerState, len(state.ImmutablePlayerStates()))
	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}
	return game, players
}

func diagramLines() []string {
	return []string{"diagram"}
}

func readsMoveName() string {
	return moveName
}

func gappedCount(state boardgame.ImmutableState) int {
	return countKeyed(state)
}

func countKeyed(state boardgame.ImmutableState) int {
	game, _ := concreteStates(state)
	if game.Cards.ImmutableComponentAtKey(0) == nil {
		return 0
	}
	return 1
}

func (m *moveCode) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}
	return legalHelper(state)
}

func legalHelper(state boardgame.ImmutableState) error {
	if countKeyed(state) > 0 {
		return errors.New("keyed")
	}
	return nil
}

func (m *moveCode) Apply(state boardgame.State) error {
	moveName = "applied"
	return nil
}

func (m *moveCode) HelpText() string {
	return "help"
}

func (m *moveGap) Apply(state boardgame.State) error {
	return nil
}

func (p *playerState) TurnDone() error {
	if !p.Done {
		return errors.New("not done")
	}
	return nil
}

func (p *playerState) ResetForTurnStart() error {
	p.Done = false
	return nil
}

func (p *playerState) ResetForTurnEnd() error {
	return nil
}

func unused() int {
	return 1
}
