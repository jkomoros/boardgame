package moves

import (
	"github.com/workfit/tester/assert"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/storage/memory"
)

const (
	phase = iota
	phaseSetUp
	phaseNormalPlay
	phaseNormalPlayDrawCard
	phaseNormalPlayActivateCard
	phaseDrawAgain
)

const (
	colorRed = iota
	colorGreen
	colorBlue
)

var enums = enum.NewSet()

var phaseEnum = enums.MustAddTree("phase", map[int]string{
	phase:                       "",
	phaseSetUp:                  "Set Up",
	phaseNormalPlay:             "Normal Play",
	phaseNormalPlayDrawCard:     "Draw Card",
	phaseNormalPlayActivateCard: "Activate Card",
	phaseDrawAgain:              "Draw Again",
}, map[int]int{
	phase:                       phase,
	phaseSetUp:                  phase,
	phaseNormalPlay:             phase,
	phaseNormalPlayDrawCard:     phaseNormalPlay,
	phaseNormalPlayActivateCard: phaseNormalPlay,
	phaseDrawAgain:              phase,
})

var colorEnum = enums.MustAdd("color", map[int]string{
	colorRed:   "Red",
	colorGreen: "Green",
	colorBlue:  "Blue",
})

//boardgame:codegen
type gameState struct {
	behaviors.RoundRobin
	base.SubState
	behaviors.CurrentPlayerBehavior
	behaviors.PhaseBehavior
	DrawStack    boardgame.Stack `stack:"cards"`
	DiscardStack boardgame.Stack `stack:"cards"`
	Counter      int
}

//boardgame:codegen
type playerState struct {
	base.SubState
	behaviors.PlayerColor
	Hand      boardgame.Stack `stack:"cards"`
	OtherHand boardgame.Stack `stack:"cards"`
	Counter   int
}

func (p *playerState) FinishStateSetUp() {
	if p.State().Manager().Delegate().(*gameDelegate).skipConnectBehaviors {
		return
	}
	p.PlayerColor.ConnectBehavior(p)
}

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

type gameDelegate struct {
	base.GameDelegate
	moveInstaller        func(manager *boardgame.GameManager) []boardgame.MoveConfig
	skipConnectBehaviors bool
}

func (g *gameDelegate) Name() string {
	return "moves"
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, _ := concreteStates(state)

	return game.DrawStack, nil
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 4
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.ImmutableState) boardgame.PlayerIndex {
	return state.ImmutableGameState().(*gameState).CurrentPlayer
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
	return g.moveInstaller(g.Manager())
}

func (g *gameDelegate) ConfigureEnums() *enum.Set {
	return enums
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		"cards": playingcards.NewDeck(false),
	}
}

func newGameManager(moveInstaller func(manager *boardgame.GameManager) []boardgame.MoveConfig, skipConnectBehaviors bool) (*boardgame.GameManager, error) {

	return boardgame.NewGameManager(&gameDelegate{moveInstaller: moveInstaller, skipConnectBehaviors: skipConnectBehaviors}, memory.NewStorageManager())

}

func TestNoBehaviorConnectErrors(t *testing.T) {

	//Tests that if you have a ConnectableBehavior (which
	//playerState.PlayerColor is) and you don't call ConnectBehavior that the
	//game manager fails to be created.

	_, err := newGameManager(defaultMoveInstaller, true)
	assert.For(t).ThatActual(err).IsNotNil()
}
