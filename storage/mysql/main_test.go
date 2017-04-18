package mysql

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/test"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestStorageManager(t *testing.T) {

	test.Test(func() test.StorageManager {
		return NewStorageManager(true)
	}, "mysql", "root:root@tcp(localhost:3306)/boardgame_test", t)

}

func TestWinnersConversion(t *testing.T) {
	tests := []struct {
		input       string
		result      []boardgame.PlayerIndex
		expectError bool
	}{
		{
			"",
			nil,
			false,
		},
		{
			"1,2",
			[]boardgame.PlayerIndex{1, 2},
			false,
		},
		{
			"-1",
			[]boardgame.PlayerIndex{-1},
			false,
		},
		{
			"1,2,",
			nil,
			true,
		},
	}

	for i, test := range tests {
		winners, err := stringToWinners(test.input)

		if test.expectError {
			assert.For(t, i).ThatActual(err).IsNotNil()
			continue
		} else {
			assert.For(t, i).ThatActual(err).IsNil()
		}

		assert.For(t, i).ThatActual(winners).Equals(test.result).ThenDiffOnFail()

		reInput := winnersToString(test.result)

		assert.For(t, i).ThatActual(reInput).Equals(test.input)
	}
}
