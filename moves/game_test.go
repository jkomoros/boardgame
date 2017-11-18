package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/storage/memory"
)

const (
	phaseSetUp = iota
	phaseNormalPlay
)

var enums = enum.NewSet()

var phaseEnum = enums.MustAdd("Phase", map[int]string{
	phaseSetUp:      "Set Up",
	phaseNormalPlay: "Normal Play",
})

//+autoreader
type gameState struct {
	boardgame.BaseSubState
	Phase         enum.MutableVal `enum:"Phase"`
	CurrentPlayer boardgame.PlayerIndex
	DrawStack     boardgame.MutableStack `stack:"cards"`
	DiscardStack  boardgame.MutableStack `stack:"cards"`
	Counter       int
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
	Hand        boardgame.MutableStack `stack:"cards"`
	Counter     int
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func concreteStates(state boardgame.State) (*gameState, []*playerState) {
	game := state.GameState().(*gameState)

	players := make([]*playerState, len(state.PlayerStates()))

	for i, player := range state.PlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "tester"
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {
	game, _ := concreteStates(state)

	return game.DrawStack, nil
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	return state.GameState().(*gameState).CurrentPlayer
}

func (g *gameDelegate) CurrentPhase(state boardgame.State) int {
	return state.GameState().(*gameState).Phase.Value()
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return &playerState{
		playerIndex: index,
	}
}

func newGameManager(moveInstaller func(manager *boardgame.GameManager) error) (*boardgame.GameManager, error) {
	chest := boardgame.NewComponentChest(enums)

	if err := chest.AddDeck("cards", playingcards.NewDeck(false)); err != nil {
		return nil, errors.New("couldn't add deck: " + err.Error())
	}

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, memory.NewStorageManager())

	if manager == nil {
		return nil, errors.New("No manager returned")
	}

	if err := moveInstaller(manager); err != nil {
		return nil, errors.New("Couldn't add moves: " + err.Error())
	}

	if err := manager.SetUp(); err != nil {
		return nil, errors.New("Couldn't set up manager: " + err.Error())
	}

	return manager, nil
}
