package checkers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestNewManager(t *testing.T) {

	//A lot of validation goes on in boardgame.NewGameManager, which means
	//that simply testing that we don't get an error with our delegate is a
	//useful test. However, this is not a very robust test because it doesn't
	//verify that moves are legal when they should be or do the right things,
	//among other things. Typically you should also create golden game
	//examples to verify the behavior of your game matches expectations. See
	//TUTORIAL.md for more on goldens.

	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	assert.For(t).ThatActual(manager).IsNotNil()

	assert.For(t).ThatActual(err).IsNil()

}
