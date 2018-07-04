package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/roundrobinhelpers"
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

var enums = enum.NewSet()

var phaseEnum = enums.MustAddTree("Phase", map[int]string{
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

//+autoreader
type gameState struct {
	roundrobinhelpers.BaseGameState
	Phase         enum.Val `enum:"Phase"`
	CurrentPlayer boardgame.PlayerIndex
	DrawStack     boardgame.Stack `stack:"cards"`
	DiscardStack  boardgame.Stack `stack:"cards"`
	Counter       int
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
	Hand        boardgame.Stack `stack:"cards"`
	OtherHand   boardgame.Stack `stack:"cards"`
	Counter     int
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func (g *gameState) SetCurrentPhase(phase int) {
	g.Phase.SetValue(phase)
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
	boardgame.DefaultGameDelegate
	moveInstaller func(manager *boardgame.GameManager) []boardgame.MoveTypeConfig
}

func (g *gameDelegate) Name() string {
	return "tester"
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

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return &playerState{
		playerIndex: index,
	}
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveTypeConfig {
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

func newGameManager(moveInstaller func(manager *boardgame.GameManager) []boardgame.MoveTypeConfig) (*boardgame.GameManager, error) {

	return boardgame.NewGameManager(&gameDelegate{moveInstaller: moveInstaller}, memory.NewStorageManager())

}
