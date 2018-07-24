package golden

import (
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestBasic(t *testing.T) {

	rec, err := record.New("test/basic_blackjack.json")

	assert.For(t).ThatActual(err).IsNil()

	err = Compare(blackjack.NewDelegate(), rec)

	assert.For(t).ThatActual(err).IsNil()

}
