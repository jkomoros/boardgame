package golden

import (
	"testing"

	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/workfit/tester/assert"
)

func TestBasic(t *testing.T) {

	err := Compare(blackjack.NewDelegate(), "test/basic_blackjack.json", false)

	assert.For(t).ThatActual(err).IsNil()

}

func TestFolder(t *testing.T) {
	err := CompareFolder(blackjack.NewDelegate(), "test", false)

	assert.For(t).ThatActual(err).IsNil()
}
