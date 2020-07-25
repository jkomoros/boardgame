package blackjack

/*

This file was created automatically by the filesystem storage layer with a
golden folder.

It will be overwritten the next time a filesystem is booted that uses this
game package.

*/

import (
	"flag"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/golden"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files if they're different instead of erroring")

func TestGolden(t *testing.T) {
	if err := golden.CompareFolder(NewDelegate(), "testdata/golden", *updateGolden); err != nil {
		t.Error(err.Error())
	}
}
