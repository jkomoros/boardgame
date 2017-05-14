package memory

import (
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
