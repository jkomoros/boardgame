package tictactoe

/*

This file was created automatically by the filesystem storage layer with a
golden folder.

It will be overwritten the next time a filesystem is booted that uses this
game package.

*/

import (
	"github.com/jkomoros/boardgame/boardgame-util/lib/golden"
	"testing"
)

func TestGolden(t *testing.T) {
	if err := golden.CompareFolder(NewDelegate(), "testdata/golden"); err != nil {
		t.Error(err.Error())
	}
}
