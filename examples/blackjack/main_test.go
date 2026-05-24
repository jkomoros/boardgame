package blackjack

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestManager(t *testing.T) {

	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	if err != nil {
		t.Fatalf("NewGameManager error: %v", err)
	}
	assert.For(t).ThatActual(manager).IsNotNil()

	game, err := manager.NewDefaultGame()

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(game).IsNotNil()

}
