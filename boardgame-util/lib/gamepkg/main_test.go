package gamepkg

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	tests := []struct {
		description  string
		input        string
		basePath     string
		errExpected  bool
		expectedName string
		//Leave empty to signal that input is the expected import.
		expectedImport string
	}{
		{
			"Basic legal",
			"github.com/jkomoros/boardgame/examples/blackjack",
			"SHOULDNTBEUSED",
			false,
			"blackjack",
			"",
		},
		{
			"Basic illegal import",
			"github.com/jkomoros/boardgame/NONEXISTENTFOLDER/blackjack",
			"",
			true,
			"blackjack",
			"",
		},
		{
			"No go files",
			"github.com/jkomoros/boardgame/boardgame-util/lib",
			"",
			true,
			"",
			"",
		},
		{
			"Package doesn't have NewDelegate",
			"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg",
			"",
			true,
			"",
			"",
		},
		{
			"Relative path to blackjack",
			"../../../examples/blackjack/",
			"",
			false,
			"blackjack",
			"github.com/jkomoros/boardgame/examples/blackjack",
		},
		{
			"Relative path to non-existant folder",
			"../../../blackjack",
			"",
			true,
			"",
			"",
		},
		{
			"Relative path to non-game pkg",
			"../../lib/gamepkg/",
			"",
			true,
			"",
			"",
		},
		{
			"Relative path to blackjack from parent",
			"../../examples/blackjack/",
			"../",
			false,
			"blackjack",
			"github.com/jkomoros/boardgame/examples/blackjack",
		},
	}

	for i, test := range tests {
		pkg, err := New(test.input, test.basePath)
		if test.errExpected {
			if err == nil {
				assert.For(t, i, test.description).ThatActual(err).IsNotNil()
			}
			continue
		}
		assert.For(t, i, test.description).ThatActual(err).IsNil()
		assert.For(t, i, test.description).ThatActual(pkg).IsNotNil()

		if test.expectedImport == "" {
			test.expectedImport = test.input
		}

		assert.For(t, i, test.description).ThatActual(pkg.Name()).Equals(test.expectedName)
		assert.For(t, i, test.description).ThatActual(pkg.Import()).Equals(test.expectedImport)

	}
}
