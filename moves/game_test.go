package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/roundrobinhelpers"
	"github.com/jkomoros/boardgame/storage/memory"
)

const (
	phaseSetUp = iota
	phaseNormalPlay
	phaseDrawAgain
)

var enums = enum.NewSet()

var phaseEnum = enums.MustAdd("Phase", map[int]string{
	phaseSetUp:      "Set Up",
	phaseNormalPlay: "Normal Play",
	phaseDrawAgain:  "Draw Again",
})

//+autoreader
type gameState struct {
	roundrobinhelpers.BaseGameState
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
	OtherHand   boardgame.MutableStack `stack:"cards"`
	Counter     int
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func (g *gameState) SetCurrentPhase(phase int) {
	g.Phase.SetValue(phase)
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
	moveInstaller func(manager *boardgame.GameManager) *boardgame.MoveTypeConfigBundle
}

func (g *gameDelegate) Name() string {
	return "tester"
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {
	game, _ := concreteStates(state)

	return game.DrawStack, nil
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 4
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	return state.GameState().(*gameState).CurrentPlayer
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return &playerState{
		playerIndex: index,
	}
}

func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {
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

func newGameManager(moveInstaller func(manager *boardgame.GameManager) *boardgame.MoveTypeConfigBundle) (*boardgame.GameManager, error) {

	return boardgame.NewGameManager(&gameDelegate{moveInstaller: moveInstaller}, memory.NewStorageManager())

}
