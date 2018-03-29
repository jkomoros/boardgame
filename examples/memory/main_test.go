package memory

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
	"testing"
)

func BenchmarkMoves(b *testing.B) {
	manager, _ := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	for j := 0; j < b.N; j++ {

		game := manager.NewGame()

		game.SetUp(2, nil, nil)

		for i := 0; i < 10; i++ {

			move := game.MoveByName("Reveal Card")

			<-game.ProposeMove(move, game.CurrentState().CurrentPlayerIndex())

			move = game.MoveByName("Reveal Card")

			<-game.ProposeMove(move, game.CurrentState().CurrentPlayerIndex())

			move = game.MoveByName("Hide Cards")

			<-game.ProposeMove(move, game.CurrentState().CurrentPlayerIndex())
		}
	}

}

func TestMain(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	game := manager.NewGame()

	assert.For(t).ThatActual(game).IsNotNil()

	err = game.SetUp(2, nil, nil)

	if !assert.For(t).ThatActual(err).IsNil().Passed() {
		t.FailNow()
	}

	move := game.MoveByName("Reveal Card")

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, boardgame.AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNil()

	gameState, _ := concreteStates(game.CurrentState())

	var revealedType string
	var revealedIndex int

	for i, c := range gameState.VisibleCards.Components() {
		if c != nil {
			revealedType = c.Values.(*cardValue).Type
			revealedIndex = i
			break
		}
	}

	assert.For(t).ThatActual(revealedType).DoesNotEqual("")

	//Find a card that does NOT match

	cardToFlip := 0

	for i, c := range gameState.HiddenCards.Components() {
		if c == nil {
			continue
		}
		if c.Values.(*cardValue).Type != revealedType {
			cardToFlip = i
			break
		}
	}

	move = game.MoveByName("Reveal Card")

	move.(*MoveRevealCard).CardIndex = cardToFlip

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, boardgame.AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNil()

	gameState, _ = concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.VisibleCards.NumComponents()).Equals(2)

	move = game.MoveByName("Hide Cards")

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, boardgame.AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNil()

	gameState, _ = concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.VisibleCards.NumComponents()).Equals(0)

	assert.For(t).ThatActual(gameState.CurrentPlayer).Equals(boardgame.PlayerIndex(1))

	move = game.MoveByName("Reveal Card")

	assert.For(t).ThatActual(move).IsNotNil()

	move.(*MoveRevealCard).CardIndex = revealedIndex

	err = <-game.ProposeMove(move, boardgame.AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNil()

	//Look for a card that DOES match.

	cardToFlip = -1

	for i, c := range gameState.HiddenCards.Components() {
		if c == nil {
			continue
		}
		if i == revealedIndex {
			continue
		}
		if c.Values.(*cardValue).Type == revealedType {
			cardToFlip = i
			break
		}
	}

	assert.For(t).ThatActual(cardToFlip).DoesNotEqual(-1)

	move = game.MoveByName("Reveal Card")

	move.(*MoveRevealCard).CardIndex = cardToFlip

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, boardgame.AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNil()

	gameState, playerStates := concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.VisibleCards.NumComponents()).Equals(0)

	assert.For(t).ThatActual(playerStates[1].WonCards.NumComponents()).Equals(2)

	assert.For(t).ThatActual(gameState.CurrentPlayer).Equals(boardgame.PlayerIndex(0))

}
