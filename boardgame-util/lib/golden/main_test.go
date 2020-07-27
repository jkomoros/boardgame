package golden

import (
	"flag"
	"testing"

	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/workfit/tester/assert"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files if they're different instead of erroring")

func TestBasic(t *testing.T) {

	//If we also used updateGolden here, then the two tests would collide.
	err := Compare(blackjack.NewDelegate(), "test/basic_blackjack.json", false)

	assert.For(t).ThatActual(err).IsNil()

}

func TestFolder(t *testing.T) {
	err := CompareFolder(blackjack.NewDelegate(), "test", *updateGolden)
	assert.For(t).ThatActual(err).IsNil()
}
