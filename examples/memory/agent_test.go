package memory

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestCardSeen(t *testing.T) {
	agent := &agentState{
		MemoryLength: 4,
	}

	cards := []string{
		"boop",
		"boop",
		"bam",
		"bam",
		"slam",
		"slam",
		"foo",
		"foo",
	}

	ok := agent.CardSeen(cards[0], 0)

	assert.For(t).ThatActual(ok).IsTrue()

	assert.For(t).ThatActual(len(agent.LastCards)).Equals(1)

	assert.For(t).ThatActual(agent.LastCards[0].Index).Equals(0)
	assert.For(t).ThatActual(agent.LastCards[0].Value).Equals(cards[0])

	ok = agent.CardSeen(cards[1], 1)

	assert.For(t).ThatActual(ok).IsTrue()

	if !assert.For(t).ThatActual(len(agent.LastCards)).Equals(2).Passed() {
		t.FailNow()
	}

	//Make sure that new values are added to the front.
	assert.For(t).ThatActual(agent.LastCards[0].Index).Equals(1)
	assert.For(t).ThatActual(agent.LastCards[0].Value).Equals(cards[1])
	assert.For(t).ThatActual(agent.LastCards[1].Index).Equals(0)
	assert.For(t).ThatActual(agent.LastCards[1].Value).Equals(cards[0])

	//Make sure that a card that is already seen is not added again
	ok = agent.CardSeen(cards[0], 0)

	assert.For(t).ThatActual(ok).IsFalse()

	assert.For(t).ThatActual(len(agent.LastCards)).Equals(2)

	//Fill up memory
	agent.CardSeen(cards[2], 2)
	agent.CardSeen(cards[3], 3)

	//Make sure memory is filled up
	assert.For(t).ThatActual(len(agent.LastCards)).Equals(4)

	//Make sure that adding an extra card past memory expires old memory.
	agent.CardSeen(cards[4], 4)

	assert.For(t).ThatActual(len(agent.LastCards)).Equals(4)

}

func TestCullInvalidCards(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	game := manager.NewGame()

	err = game.SetUp(2, nil, nil)

	assert.For(t).ThatActual(err).IsNil()

	gameState, _ := concreteStates(game.CurrentState())

	agent := &agentState{
		MemoryLength: 4,
		LastCards: []agentCardInfo{
			{
				Index: 0,
			},
			{
				Index: 1,
			},
			{
				Index: 2,
			},
			{
				Index: 3,
			},
		},
	}

	agent.CullInvalidCards(gameState)

	assert.For(t).ThatActual(len(agent.LastCards)).Equals(4)

	gameState.HiddenCards.First().MoveTo(gameState.VisibleCards, 0)

	agent.CullInvalidCards(gameState)

	assert.For(t).ThatActual(agent.LastCards).Equals([]agentCardInfo{
		{
			Index: 1,
		},
		{
			Index: 2,
		},
		{
			Index: 3,
		},
	})

	err = gameState.HiddenCards.ComponentAt(2).MoveTo(gameState.VisibleCards, 2)

	assert.For(t).ThatActual(err).IsNil()

	agent.CullInvalidCards(gameState)

	assert.For(t).ThatActual(agent.LastCards).Equals([]agentCardInfo{
		{
			Index: 1,
		},
		{
			Index: 3,
		},
	})

	err = gameState.HiddenCards.ComponentAt(3).MoveTo(gameState.VisibleCards, 3)

	assert.For(t).ThatActual(err).IsNil()

	agent.CullInvalidCards(gameState)

	assert.For(t).ThatActual(agent.LastCards).Equals([]agentCardInfo{
		{
			Index: 1,
		},
	})
}

func cardsLess(i, j boardgame.ImmutableComponentInstance) bool {
	if i == nil {
		return true
	}
	if j == nil {
		return false
	}
	iType := i.Values().(*cardValue).Type
	jType := j.Values().(*cardValue).Type
	//Break ties with the component that has a lower deckIndex.
	if iType == jType {
		return i.DeckIndex() < j.DeckIndex()
	}
	return iType < jType
}

func TestCardsToFlip(t *testing.T) {

	agent := &agentState{
		MemoryLength: 4,
	}

	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	game := manager.NewGame()

	err = game.SetUp(2, nil, nil)

	assert.For(t).ThatActual(err).IsNil()

	gameState, _ := concreteStates(game.CurrentState())

	err = gameState.HiddenCards.SortComponents(cardsLess)

	assert.For(t).ThatActual(err).IsNil()

	one := agent.FirstCardToFlip(gameState)

	assert.For(t).ThatActual(one).DoesNotEqual(-1)

	gameState.HiddenCards.ComponentAt(one).MoveTo(gameState.VisibleCards, one)

	two := agent.SecondCardToFlip(gameState)

	assert.For(t).ThatActual(two).DoesNotEqual(-1)
	assert.For(t).ThatActual(two).DoesNotEqual(one)

	gameState.VisibleCards.ComponentAt(one).MoveTo(gameState.HiddenCards, one)

	agent.CardSeen(gameState.HiddenCards.ComponentAt(0).Values().(*cardValue).Type, 0)
	agent.CardSeen(gameState.HiddenCards.ComponentAt(2).Values().(*cardValue).Type, 2)
	agent.CardSeen(gameState.HiddenCards.ComponentAt(3).Values().(*cardValue).Type, 3)

	one = agent.FirstCardToFlip(gameState)

	assert.For(t).ThatActual(one).Equals(3)

	gameState.HiddenCards.ComponentAt(one).MoveTo(gameState.VisibleCards, one)

	//This test has been flaky in the past
	for i := 0; i < 20; i++ {

		two = agent.SecondCardToFlip(gameState)

		assert.For(t).ThatActual(two).Equals(2)
	}

	//Verify that cards that are not in hidden are never suggested by CardsToFlip.
	gameState.HiddenCards.First().MoveTo(gameState.VisibleCards, 0)
	gameState.HiddenCards.ComponentAt(1).MoveTo(gameState.VisibleCards, 1)

	for i := 0; i < 50; i++ {
		one = agent.FirstCardToFlip(gameState)

		assert.For(t).ThatActual(one).DoesNotEqual(0)
		assert.For(t).ThatActual(one).DoesNotEqual(1)
		assert.For(t).ThatActual(one).DoesNotEqual(2)
	}

}
