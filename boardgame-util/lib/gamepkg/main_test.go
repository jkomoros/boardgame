package gamepkg

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	tests := []struct {
		description   string
		input         string
		inputIsImport bool
		errExpected   bool
	}{
		{
			"Basic legal",
			"github.com/jkomoros/boardgame/examples/blackjack",
			true,
			false,
		},
		{
			"Basic illegal import",
			"github.com/jkomoros/boardgame/NONEXISTENTFOLDER/blackjack",
			true,
			true,
		},
		{
			"No go files",
			"github.com/jkomoros/boardgame/boardgame-util/lib",
			true,
			true,
		},
		{
			"Package doesn't have NewDelegate",
			"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg",
			true,
			true,
		},
	}

	for i, test := range tests {
		pkg, err := New(test.input)
		if test.errExpected {
			if err == nil {
				assert.For(t, i, test.description).ThatActual(err).IsNotNil()
			}
			continue
		}
		assert.For(t, i, test.description).ThatActual(err).IsNil()
		assert.For(t, i, test.description).ThatActual(pkg).IsNotNil()

		if !test.inputIsImport {
			continue
		}

		calculatedImport, err := pkg.Import()

		assert.For(t, i, test.description).ThatActual(err).IsNil()

		assert.For(t, i, test.description).ThatActual(calculatedImport).Equals(test.input)
	}
}
