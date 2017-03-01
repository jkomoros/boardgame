package boardgame

import (
	"encoding/json"
	"testing"
)

type testGameDelegate struct {
	DefaultGameDelegate
}

func (t *testGameDelegate) DistributeComponentToStarterStack(state *State, c *Component) error {
	game, _ := concreteStates(state)
	return game.DrawDeck.InsertFront(c)
}

func (t *testGameDelegate) Name() string {
	return testGameName
}

func (t *testGameDelegate) CheckGameFinished(state *State) (bool, []int) {
	_, players := concreteStates(state)

	var winners []int

	for i, player := range players {
		if player.Score >= 5 {
			//This user won!
			winners = append(winners, i)

			//Keep going through to see if anyone else won at the same time
		}
	}

	if len(winners) > 0 {
		return true, winners
	}

	return false, nil
}

func (t *testGameDelegate) DefaultNumPlayers() int {
	return 3
}

func (t *testGameDelegate) StartingStateProps(numPlayers int) *StateProps {

	chest := t.Manager().Chest()

	deck := chest.Deck("test")

	return &StateProps{
		Game: &testGameState{
			CurrentPlayer: 0,
			DrawDeck:      NewGrowableStack(deck, 0),
		},
		Players: []PlayerState{
			&testPlayerState{
				playerIndex:       0,
				Score:             0,
				MovesLeftThisTurn: 1,
				Hand:              NewSizedStack(deck, 2),
				IsFoo:             false,
			},
			&testPlayerState{
				playerIndex:       1,
				Score:             0,
				MovesLeftThisTurn: 0,
				Hand:              NewSizedStack(deck, 2),
				IsFoo:             false,
			},
			&testPlayerState{
				playerIndex:       2,
				Score:             0,
				MovesLeftThisTurn: 0,
				Hand:              NewSizedStack(deck, 2),
				IsFoo:             true,
			},
		},
	}
}

func (t *testGameDelegate) GameStateFromBlob(blob []byte) (GameState, error) {
	var result testGameState

	if err := json.Unmarshal(blob, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (t *testGameDelegate) PlayerStateFromBlob(blob []byte, index int) (PlayerState, error) {
	var result testPlayerState

	if err := json.Unmarshal(blob, &result); err != nil {
		return nil, err
	}

	result.playerIndex = index

	return &result, nil
}

func TestTestGameDelegate(t *testing.T) {
	manager := newTestGameManger()

	if manager.Delegate().Name() != testGameName {
		t.Error("Manager.Name() was not overridden")
	}
}
