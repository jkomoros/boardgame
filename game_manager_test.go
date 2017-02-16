package boardgame

import (
	"encoding/json"
	"testing"
)

type testGameManager struct {
	DefaultGameManager
}

func (t *testGameManager) DistributeComponentToStarterStack(state State, c *Component) error {
	p := state.(*testState)
	return p.Game.DrawDeck.InsertFront(c)
}

func (t *testGameManager) CheckGameFinished(state State) (bool, []int) {
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

func (t *testGameManager) DefaultNumPlayers() int {
	return 3
}

func (t *testGameManager) StartingState(numPlayers int) State {

	chest := t.Game.Chest()

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

func (t *testGameManager) StateFromBlob(blob []byte, schema int) (State, error) {
	result := &testState{}
	if err := json.Unmarshal(blob, result); err != nil {
		return nil, err
	}

	result.Game.DrawDeck.Inflate(t.Game.Chest())

	for i, user := range result.Users {
		user.playerIndex = i
	}

	return result, nil
}

func newTestGameManger() GameManager {
	manager := &testGameManager{}

	manager.AddPlayerMove(&testMove{})
	manager.AddFixUpMove(&testMoveAdvanceCurentPlayer{})

	return manager
}

func TestGameManagerSetUp(t *testing.T) {

	manager := newTestGameManger().(*testGameManager)

	if manager.PlayerMoves() != nil {
		t.Error("Got moves back before SetUp was called")
	}

	if manager.PlayerMoveByName("Test") != nil {
		t.Error("Move by name returned a move before SetUp was called")
	}

	manager.SetUp()

	moves := manager.PlayerMoves()

	if moves == nil {
		t.Error("Got nil player moves even after setting up")
	}

	for i := 0; i < len(moves); i++ {
		if moves[i] == manager.playerMoves[i] {
			t.Error("PlayerMoves didn't return a copy; got same item at", i)
		}
	}

	if manager.PlayerMoveByName("Test") == nil {
		t.Error("MoveByName didn't return a valid move when provided the proper name after calling setup")
	}

	if manager.PlayerMoveByName("test") == nil {
		t.Error("MoveByName didn't return a valid move when provided with a lowercase name after calling SetUp.")
	}

}
