package boardgame

import (
	"testing"
	"time"

	"github.com/workfit/tester/assert"
)

func TestTimerManager(t *testing.T) {

	game := testDefaultGame(t, true)

	currentVersion := game.Version()

	move := game.MoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	assert.For(t).ThatActual(move.(*testMoveDrawCard).TargetPlayerIndex).Equals(PlayerIndex(0))

	timer := newTimerManager(game.manager)

	assert.For(t).ThatActual(timer.nextTimerFired()).Equals(false)

	registeredDuration := time.Duration(50) * time.Millisecond

	id := timer.PrepareTimer(registeredDuration, game.CurrentState().(*state), move)

	assert.For(t).ThatActual(id).Equals("D732D7BBF5331D57")

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
	assert.For(t).ThatActual(len(timer.recordsByID)).Equals(0)

	assert.For(t).ThatActual(timer.GetTimerRemaining(id)).Equals(time.Duration(0))

}

func TestTimerManagerMultiple(t *testing.T) {

	game := testDefaultGame(t, false)

	move := game.MoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	currentVersion := game.Version()

	timer := newTimerManager(game.manager)

	state := game.CurrentState().(*state)

	firstID := timer.PrepareTimer(time.Duration(50)*time.Millisecond, state, move)
	timer.StartTimer(firstID)
	secondID := timer.PrepareTimer(time.Duration(10)*time.Millisecond, state, move)
	timer.StartTimer(secondID)
	thirdID := timer.PrepareTimer(time.Duration(70)*time.Millisecond, state, move)
	timer.StartTimer(thirdID)

	//Make sure that even though the second timer was added second, it is first.
	assert.For(t).ThatActual(timer.records[0].id).Equals(secondID)
	assert.For(t).ThatActual(timer.records[1].id).Equals(firstID)
	assert.For(t).ThatActual(timer.records[2].id).Equals(thirdID)

	timer.CancelTimer(secondID)

	assert.For(t).ThatActual(len(timer.records)).Equals(2)

	assert.For(t).ThatActual(timer.records[0].id).Equals(firstID)
	assert.For(t).ThatActual(timer.records[1].id).Equals(thirdID)

	assert.For(t).ThatActual(timer.nextTimerFired()).IsFalse()

	<-time.After(70 * time.Millisecond)

	timer.Tick()

	assert.For(t).ThatActual(len(timer.records)).Equals(0)

	assert.For(t).ThatActual(game.Version()).Equals(currentVersion + 2)

}

func TestTimerProp(t *testing.T) {
	game := testDefaultGame(t, false)

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

	assert.For(t).ThatActual(gameState.Timer.id()).Equals("")

}
