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

	for i, player := range p.Players {
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

func (t *testGameDelegate) StartingState(numPlayers int) State {

	chest := t.Manager().Chest()

	deck := chest.Deck("test")

	return &testState{
		Game: &testGameState{
			CurrentPlayer: 0,
			DrawDeck:      NewGrowableStack(deck, 0),
		},
		Players: []*testPlayerState{
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

func (t *testGameDelegate) StateFromBlob(blob []byte) (State, error) {
	result := &testState{}
	if err := json.Unmarshal(blob, result); err != nil {
		return nil, err
	}

	result.Game.DrawDeck.Inflate(t.Manager().Chest())

	for i, player := range result.Players {
		player.playerIndex = i
		player.Hand.Inflate(t.Manager().Chest())
	}

	return result, nil
}

func TestTestGameDelegate(t *testing.T) {
	manager := newTestGameManger()

	if manager.Delegate().Name() != testGameName {
		t.Error("Manager.Name() was not overridden")
	}
}
