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

func TestTimerManagerMultiple(t *testing.T) {

	game := testGame()

	game.SetUp(2)

	move := &testMoveDrawCard{
		TargetPlayerIndex: 0,
	}

	currentVersion := game.Version()

	timer := newTimerManager()

	firstId := timer.RegisterTimer(time.Duration(50)*time.Millisecond, game, move)
	secondId := timer.RegisterTimer(time.Duration(10)*time.Millisecond, game, move)
	thirdId := timer.RegisterTimer(time.Duration(70)*time.Millisecond, game, move)

	//Make sure that even though the second timer was added second, it is first.
	assert.For(t).ThatActual(timer.records[0].id).Equals(secondId)
	assert.For(t).ThatActual(timer.records[1].id).Equals(firstId)
	assert.For(t).ThatActual(timer.records[2].id).Equals(thirdId)

	timer.CancelTimer(secondId)

	assert.For(t).ThatActual(len(timer.records)).Equals(2)

	assert.For(t).ThatActual(timer.records[0].id).Equals(firstId)
	assert.For(t).ThatActual(timer.records[1].id).Equals(thirdId)

	assert.For(t).ThatActual(timer.nextTimerFired()).IsFalse()

	<-time.After(70 * time.Millisecond)

	timer.Tick()

	assert.For(t).ThatActual(len(timer.records)).Equals(0)

	assert.For(t).ThatActual(game.Version()).Equals(currentVersion + 2)

}

func TestTimerProp(t *testing.T) {
	game := testGame()

	game.SetUp(2)

	move := &testMoveDrawCard{
		TargetPlayerIndex: 0,
	}

	currentVersion := game.Version()

	gameState, _ := concreteStates(game.CurrentState())

	gameState.Timer.Start(time.Millisecond*5, move)

	<-time.After(time.Millisecond * 300)

	assert.For(t).ThatActual(game.Version()).Equals(currentVersion + 1)

}
