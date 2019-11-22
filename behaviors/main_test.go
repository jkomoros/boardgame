package behaviors

import (
	"testing"

	"github.com/workfit/tester/assert"
)

func TestRoundRobin(t *testing.T) {
	var b interface{}
	b = &RoundRobin{}
	_, ok := b.(Interface)
	assert.For(t).ThatActual(ok).IsTrue()
}
