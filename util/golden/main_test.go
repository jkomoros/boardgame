package golden

import (
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestBasic(t *testing.T) {

	err := Compare(blackjack.NewDelegate(), "test/basic_blackjack.json")

	assert.For(t).ThatActual(err).IsNil()

}

func TestFolder(t *testing.T) {
	err := CompareFolder(blackjack.NewDelegate(), "test")

	assert.For(t).ThatActual(err).IsNil()
}