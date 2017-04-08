package boardgame

import (
	"github.com/workfit/tester/assert"
	"testing"
	"time"
)

func TestTimerManager(t *testing.T) {

	game := testGame()

	game.SetUp(2)

	currentVersion := game.Version()

	move := &testMoveDrawCard{
		TargetPlayerIndex: 0,
	}

	timer := newTimerManager()

	assert.For(t).ThatActual(timer.nextTimerFired()).Equals(false)

	assert.For(t).ThatActual(timer.nextId).Equals(0)

	registeredDuration := time.Duration(50) * time.Millisecond

	id := timer.RegisterTimer(registeredDuration, game, move)

	assert.For(t).ThatActual(id).Equals(0)

	assert.For(t).ThatActual(timer.nextId).Equals(1)

	remaining := timer.GetTimerRemaining(id)

	assert.For(t).ThatActual(registeredDuration-remaining < time.Millisecond*10).IsTrue()

	assert.For(t).ThatActual(timer.nextTimerFired()).Equals(false)

	timer.Tick()

	//Ticking before any time has really passed shouldn't trigger the next timer.
	assert.For(t).ThatActual(timer.nextTimerFired()).Equals(false)

	<-time.After(60 * time.Millisecond)

	assert.For(t).ThatActual(timer.nextTimerFired()).IsTrue()

	timer.Tick()

	//Make sure the game was actually made after the time had elapsed.
	assert.For(t).ThatActual(game.Version()).Equals(currentVersion + 1)

	//Now that the tick has happened, there should be not more records.
	assert.For(t).ThatActual(len(timer.records)).Equals(0)
	assert.For(t).ThatActual(len(timer.recordsById)).Equals(0)

	assert.For(t).ThatActual(timer.GetTimerRemaining(id)).Equals(time.Duration(0))

}
