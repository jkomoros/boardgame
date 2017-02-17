package boardgame

import (
	"encoding/json"
	"testing"
)

type testGameDelegate struct {
	DefaultGameDelegate
}

func (t *testGameDelegate) DistributeComponentToStarterStack(state State, c *Component) error {
	p := state.(*testState)
	return p.Game.DrawDeck.InsertFront(c)
}

func (t *testGameDelegate) Name() string {
	return testGameName
}

func (t *testGameDelegate) CheckGameFinished(state State) (bool, []int) {
	p := state.(*testState)

	var winners []int

	for i, user := range p.Users {
		if user.Score >= 5 {
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

func (t *testGameDelegate) StartingState(numPlayers int) State {

	chest := t.Manager().Chest()

	deck := chest.Deck("test")

	return &testState{
		Game: &testGameState{
			CurrentPlayer: 0,
			DrawDeck:      NewGrowableStack(deck, 0),
		},
		Users: []*testUserState{
			&testUserState{
				playerIndex:       0,
				Score:             0,
				MovesLeftThisTurn: 1,
				IsFoo:             false,
			},
			&testUserState{
				playerIndex:       1,
				Score:             0,
				MovesLeftThisTurn: 0,
				IsFoo:             false,
			},
			&testUserState{
				playerIndex:       2,
				Score:             0,
				MovesLeftThisTurn: 0,
				IsFoo:             true,
			},
		},
	}
}

func (t *testGameDelegate) StateFromBlob(blob []byte, schema int) (State, error) {
	result := &testState{}
	if err := json.Unmarshal(blob, result); err != nil {
		return nil, err
	}

	result.Game.DrawDeck.Inflate(t.Manager().Chest())

	for i, user := range result.Users {
		user.playerIndex = i
	}

	return result, nil
}

func TestTestGameDelegate(t *testing.T) {
	manager := newTestGameManger()

	if manager.Delegate().Name() != testGameName {
		t.Error("Manager.Name() was not overridden")
	}
}
