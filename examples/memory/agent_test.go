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
	manager := NewManager(memory.NewStorageManager())

	game := boardgame.NewGame(manager)

	err := game.SetUp(2, nil)

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

	gameState.HiddenCards.MoveComponent(0, gameState.RevealedCards, 0)

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

	gameState.HiddenCards.MoveComponent(2, gameState.RevealedCards, 2)

	agent.CullInvalidCards(gameState)

	assert.For(t).ThatActual(agent.LastCards).Equals([]agentCardInfo{
		{
			Index: 1,
		},
		{
			Index: 3,
		},
	})

	gameState.HiddenCards.MoveComponent(3, gameState.RevealedCards, 3)

	agent.CullInvalidCards(gameState)

	assert.For(t).ThatActual(agent.LastCards).Equals([]agentCardInfo{
		{
			Index: 1,
		},
	})
}

func TestCardsToFlip(t *testing.T) {
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

	agent := &agentState{
		MemoryLength: 4,
	}

	manager := NewManager(memory.NewStorageManager())

	game := boardgame.NewGame(manager)

	err := game.SetUp(2, nil)

	assert.For(t).ThatActual(err).IsNil()

	gameState, _ := concreteStates(game.CurrentState())

	//Cheat and stuff in values we know what they are into the cards. We're
	//the only one using the chest.
	for i, card := range cards {
		c := gameState.HiddenCards.ComponentAt(i)
		values := c.Values.(*cardValue)
		values.Type = card
	}

	one := agent.FirstCardToFlip(gameState)

	assert.For(t).ThatActual(one).DoesNotEqual(-1)

	gameState.HiddenCards.MoveComponent(one, gameState.RevealedCards, one)

	two := agent.SecondCardToFlip(gameState)

	assert.For(t).ThatActual(two).DoesNotEqual(-1)
	assert.For(t).ThatActual(two).DoesNotEqual(one)

	gameState.RevealedCards.MoveComponent(one, gameState.HiddenCards, one)

	agent.CardSeen(cards[0], 0)
	agent.CardSeen(cards[2], 2)
	agent.CardSeen(cards[3], 3)

	one = agent.FirstCardToFlip(gameState)

	assert.For(t).ThatActual(one).Equals(3)

	gameState.HiddenCards.MoveComponent(one, gameState.RevealedCards, one)

	two = agent.SecondCardToFlip(gameState)

	assert.For(t).ThatActual(two).Equals(2)

	//Verify that cards that are not in hidden are never suggested by CardsToFlip.
	gameState.HiddenCards.MoveComponent(0, gameState.RevealedCards, 0)
	gameState.HiddenCards.MoveComponent(1, gameState.RevealedCards, 1)

	for i := 0; i < 50; i++ {
		one = agent.FirstCardToFlip(gameState)

		assert.For(t).ThatActual(one).DoesNotEqual(0)
		assert.For(t).ThatActual(one).DoesNotEqual(1)
		assert.For(t).ThatActual(one).DoesNotEqual(2)
	}

}
