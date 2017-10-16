package boardgame

import (
	"errors"
	"github.com/workfit/tester/assert"
	"testing"
)

type moveTestImmediatePlayerMove struct {
	baseMove
}

type moveImmediateFixUpOne struct {
	baseMove
}

type moveImmediateFixUpTWo struct {
	baseMove
}

var moveTestImmediatePlayerMoveConfig = MoveTypeConfig{
	Name:     "Test",
	HelpText: "This is a test",
	MoveConstructor: func() Move {
		return new(moveTestImmediatePlayerMove)
	},
	ImmediateFixUp: func(state State) Move {
		moveType, _ := newMoveType(&moveTestImmediateFixUpOneConfig, state.Game().Manager())
		return moveType.NewMove(state)
	},
}

func (m *moveTestImmediatePlayerMove) Legal(state State, proposer PlayerIndex) error {
	return nil
}

func (m *moveTestImmediatePlayerMove) Apply(state MutableState) error {
	game, _ := concreteStates(state)

	game.CurrentPlayer = game.CurrentPlayer.Next(state)

	return nil
}

func (m *moveTestImmediatePlayerMove) Reader() PropertyReader {
	return getDefaultReader(m)
}

func (m *moveTestImmediatePlayerMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(m)
}

func (m *moveTestImmediatePlayerMove) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(m)
}

var moveTestImmediateFixUpOneConfig = MoveTypeConfig{
	Name: "Immediate FixUp 1",
	MoveConstructor: func() Move {
		return new(moveImmediateFixUpOne)
	},
	ImmediateFixUp: func(state State) Move {
		moveType, _ := newMoveType(&moveTestImmediateFixUpTwoConfig, state.Game().Manager())
		return moveType.NewMove(state)
	},
	IsFixUp: true,
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

func (m *moveImmediateFixUpOne) Reader() PropertyReader {
	return getDefaultReader(m)
}

func (m *moveImmediateFixUpOne) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(m)
}

func (m *moveImmediateFixUpOne) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(m)
}

var moveTestImmediateFixUpTwoConfig = MoveTypeConfig{
	Name: "Immediate FixUp 2",
	MoveConstructor: func() Move {
		return new(moveImmediateFixUpTWo)
	},
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

func (m *moveImmediateFixUpTWo) Reader() PropertyReader {
	return getDefaultReader(m)
}

func (m *moveImmediateFixUpTWo) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(m)
}

func (m *moveImmediateFixUpTWo) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(m)
}

func TestImmediateFixUp(t *testing.T) {

	manager := NewGameManager(&testGameDelegate{}, newTestGameChest(), newTestStorageManager())

	manager.BulkAddMoveTypes([]*MoveTypeConfig{
		&moveTestImmediatePlayerMoveConfig,
		&moveTestImmediateFixUpOneConfig,
	})

	//TODO: add the FixUp with a fixup chain

	manager.SetUp()

	game := manager.NewGame()

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
