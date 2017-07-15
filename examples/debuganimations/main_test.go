package debuganimations

import (
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestManager(t *testing.T) {
	manager, err := NewManager(memory.NewStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(manager).IsNotNil()

}
