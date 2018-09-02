package checkers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	assert.For(t).ThatActual(manager).IsNotNil()

	assert.For(t).ThatActual(err).IsNil()

}
