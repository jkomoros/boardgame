package moves

import (
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

var phaseEnum = enums.MustAddTree("phase", map[enum.EnumKey]string{
	phase:                       "",
	phaseSetUp:                  "Set Up",
	phaseNormalPlay:             "Normal Play",
	phaseNormalPlayDrawCard:     "Draw Card",
	phaseNormalPlayActivateCard: "Activate Card",
	phaseDrawAgain:              "Draw Again",
}, map[enum.EnumKey]enum.EnumKey{
	phase:                       phase,
	phaseSetUp:                  phase,
	phaseNormalPlay:             phase,
	phaseNormalPlayDrawCard:     phaseNormalPlay,
	phaseNormalPlayActivateCard: phaseNormalPlay,
	phaseDrawAgain:              phase,
})

var colorEnum = enums.MustAdd("color", map[enum.EnumKey]string{
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
	behaviors.LocationBehavior `location:"TokenLocation"`
	Hand          boardgame.Stack     `stack:"cards"`
	OtherHand     boardgame.Stack     `stack:"cards"`
	TokenLocation boardgame.SizedStack `sizedstack:"tokens,4"`
	Counter       int
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
	moveInstaller func(manager *boardgame.GameManager) []boardgame.MoveConfig
}

func (g *gameDelegate) Name() string {
	return "moves"
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, players := concreteStates(state)

	if c.Deck().Name() == "tokens" {
		// Distribute one token per player into slot 0 of their TokenLocation.
		playerIndex := c.DeckIndex()
		if playerIndex < len(players) {
			return players[playerIndex].TokenLocation, nil
		}
	}

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
	tokens := boardgame.NewDeck()
	// One token per player (DefaultNumPlayers == 4).
	for i := 0; i < 4; i++ {
		tokens.AddComponent(nil)
	}

	return map[string]*boardgame.Deck{
		"cards":  playingcards.NewDeck(false),
		"tokens": tokens,
	}
}

func newGameManager(moveInstaller func(manager *boardgame.GameManager) []boardgame.MoveConfig) (*boardgame.GameManager, error) {

	return boardgame.NewGameManager(&gameDelegate{moveInstaller: moveInstaller}, memory.NewStorageManager())

}
