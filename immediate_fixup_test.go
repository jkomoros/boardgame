package boardgame

import (
	"errors"
	"github.com/workfit/tester/assert"
	"testing"
)

type moveTestImmediatePlayerMove struct {
	DefaultMove
}

type moveImmediateFixUpOne struct {
	DefaultMove
}

type moveImmediateFixUpTWo struct {
	DefaultMove
}

func moveTestImmediatePlayerMoveFactory(state State) Move {
	return &moveTestImmediatePlayerMove{
		DefaultMove{
			"Test",
			"This is a test",
		},
	}
}

func (m *moveTestImmediatePlayerMove) Legal(state State, proposer PlayerIndex) error {
	return nil
}

func (m *moveTestImmediatePlayerMove) Apply(state MutableState) error {
	game, _ := concreteStates(state)

	game.CurrentPlayer = game.CurrentPlayer.Next(state)

	return nil
}

func (m *moveTestImmediatePlayerMove) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(m)
}

func (m *moveTestImmediatePlayerMove) ImmediateFixUp(state State) Move {
	return moveImmediateFixUpOneFactory(state)
}

func moveImmediateFixUpOneFactory(state State) Move {
	return &moveImmediateFixUpOne{
		DefaultMove{
			"Immediate FixUp 1",
			"",
		},
	}
}

func (m *moveImmediateFixUpOne) Legal(state State, proposer PlayerIndex) error {
	game, players := concreteStates(state)

	if game.CurrentPlayer == 0 {
		return errors.New("The current player may not be 0")
	}

	p := players[game.CurrentPlayer]

	if p.IsFoo {
		return errors.New("The current player cannot be IsFoo=true")
	}
	return nil
}

func (m *moveImmediateFixUpOne) Apply(state MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	p.IsFoo = true

	return nil
}

func (m *moveImmediateFixUpOne) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(m)
}

func (m *moveImmediateFixUpOne) ImmediateFixUp(state State) Move {
	return moveImmediateFixUpTwoFactory(state)
}

func moveImmediateFixUpTwoFactory(state State) Move {
	return &moveImmediateFixUpTWo{
		DefaultMove{
			"Immediate FixUp 2",
			"",
		},
	}
}

func (m *moveImmediateFixUpTWo) Legal(state State, proposer PlayerIndex) error {
	return errors.New("This move is never legal and that's OK")
}

func (m *moveImmediateFixUpTWo) Apply(state MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	p.Score++

	return nil
}

func (m *moveImmediateFixUpTWo) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(m)
}

func TestImmediateFixUp(t *testing.T) {

	manager := NewGameManager(&testGameDelegate{}, newTestGameChest(), newTestStorageManager())

	manager.AddPlayerMoveFactory(moveTestImmediatePlayerMoveFactory)

	manager.AddFixUpMoveFactory(moveImmediateFixUpOneFactory)

	//TODO: add the FixUp with a fixup chain

	manager.SetUp()

	game := NewGame(manager)

	game.SetUp(0, nil)

	assert.For(t).ThatActual(game).IsNotNil()

	move := game.PlayerMoveByName("Test")

	assert.For(t).ThatActual(move).IsNotNil()

	//Gut check that the move we're proposing actually is a
	//MoveTestImmediatePlayerMove.

	convertedMove, ok := move.(*moveTestImmediatePlayerMove)

	assert.For(t).ThatActual(ok).IsTrue()

	assert.For(t).ThatActual(convertedMove).IsNotNil()

	err := <-game.ProposeMove(move, AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNil()

	//Verify that the move was made and so was the fix up

	assert.For(t).ThatActual(game.Version()).Equals(2)

	gameState, playerStates := concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.CurrentPlayer).Equals(PlayerIndex(1))

	currentPlayerState := playerStates[gameState.CurrentPlayer]

	assert.For(t).ThatActual(currentPlayerState.IsFoo).IsTrue()
	//Make sure that MoveImmediateFixUpTwo DIDN't get applied
	assert.For(t).ThatActual(currentPlayerState.Score).Equals(0)

}
