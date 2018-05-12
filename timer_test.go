package boardgame

import (
	"github.com/workfit/tester/assert"
	"testing"
	"time"
)

func TestTimerManager(t *testing.T) {

	game := testGame(t)

	err := game.SetUp(0, nil, nil)

	assert.For(t).ThatActual(err).IsNil()

	currentVersion := game.Version()

	move := game.MoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	assert.For(t).ThatActual(move.(*testMoveDrawCard).TargetPlayerIndex).Equals(PlayerIndex(0))

	timer := newTimerManager(game.manager)

	assert.For(t).ThatActual(timer.nextTimerFired()).Equals(false)

	assert.For(t).ThatActual(timer.nextId).Equals(1)

	registeredDuration := time.Duration(50) * time.Millisecond

	id := timer.PrepareTimer(registeredDuration, game, move)

	assert.For(t).ThatActual(id).Equals(1)

	assert.For(t).ThatActual(timer.nextId).Equals(2)

	remaining := timer.GetTimerRemaining(id)

	assert.For(t).ThatActual(remaining).Equals(registeredDuration)

	assert.For(t).ThatActual(timer.records[0].fireTime.Sub(time.Now()) > time.Hour)

	timer.StartTimer(id)

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

	game := testGame(t)

	err := game.SetUp(0, nil, nil)

	assert.For(t).ThatActual(err).IsNil()

	move := game.MoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	currentVersion := game.Version()

	timer := newTimerManager(game.manager)

	firstId := timer.PrepareTimer(time.Duration(50)*time.Millisecond, game, move)
	timer.StartTimer(firstId)
	secondId := timer.PrepareTimer(time.Duration(10)*time.Millisecond, game, move)
	timer.StartTimer(secondId)
	thirdId := timer.PrepareTimer(time.Duration(70)*time.Millisecond, game, move)
	timer.StartTimer(thirdId)

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
	game := testGame(t)

	err := game.SetUp(0, nil, nil)

	assert.For(t).ThatActual(err).IsNil()

	move := game.MoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	currentVersion := game.Version()

	gameState, _ := concreteStates(game.CurrentState())

	gameState.Timer.Start(time.Millisecond*5, move)

	assert.For(t).ThatActual(gameState.Timer.Active()).IsFalse()

	assert.For(t).ThatActual(gameState.Timer.TimeLeft()).Equals(time.Millisecond * 5)

	//Trigger the timers to actually be added
	game.CurrentState().(*state).committed()

	assert.For(t).ThatActual(gameState.Timer.Active()).IsTrue()

	assert.For(t).ThatActual(gameState.Timer.TimeLeft() < time.Millisecond*50).IsTrue()

	<-time.After(time.Millisecond * 300)

	assert.For(t).ThatActual(game.Version()).Equals(currentVersion + 1)

	gameState, _ = concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.Timer.Active()).IsFalse()

	assert.For(t).ThatActual(gameState.Timer.TimeLeft()).Equals(time.Duration(0))

	gameState.Timer.Cancel()

	assert.For(t).ThatActual(gameState.Timer.id()).Equals(0)

}
