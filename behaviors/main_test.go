package behaviors

import (
	"testing"

	"github.com/jkomoros/boardgame/moves/interfaces"

	"github.com/workfit/tester/assert"
)

func TestRoundRobin(t *testing.T) {
	var b interface{}
	b = &RoundRobin{}
	_, ok := b.(interfaces.RoundRobinProperties)
	assert.For(t).ThatActual(ok).IsTrue()
}

func TestCurrentPlayer(t *testing.T) {
	var b interface{}
	b = &CurrentPlayerBehavior{}
	_, ok := b.(interfaces.CurrentPlayerSetter)
	assert.For(t).ThatActual(ok).IsTrue()
}

func TestPhase(t *testing.T) {
	var b interface{}
	b = &PhaseBehavior{}
	_, ok := b.(interfaces.CurrentPhaseSetter)
	assert.For(t).ThatActual(ok).IsTrue()
}

func TestColor(t *testing.T) {
	var b interface{}
	b = &PlayerColor{}
	_, ok := b.(Connectable)
	assert.For(t).ThatActual(ok).IsTrue()

	//Note that more substantive testing of PlayerColor is done in
	//moves/game_test.go, since testing it requires a whole test game and the
	//moves package has one.
}
