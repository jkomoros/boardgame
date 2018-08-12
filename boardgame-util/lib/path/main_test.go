package path

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestPrefix(t *testing.T) {

	tests := []struct {
		from      string
		to        string
		expected  string
		expectErr bool
	}{
		{
			"/a/b/c",
			"/a/b/d",
			"../d",
			false,
		},
		{
			"a/b/c",
			"/a/b/c",
			"",
			true,
		},
		{
			"/a/b/c",
			"a/b/",
			"",
			true,
		},
		{
			"/a/b/c",
			"/a/d/e/f",
			"../../d/e/f",
			false,
		},
		{
			"/a/b/c",
			"/d/e/f",
			"../../../d/e/f",
			false,
		},
		{
			"/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame/boardgame-util/static/",
			"/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame/server/static/webapp/bower.json",
			"../../server/static/webapp/bower.json",
			false,
		},
	}

	for i, test := range tests {
		result, err := RelativizePaths(test.from, test.to)

		if test.expectErr {
			assert.For(t, i).ThatActual(err).IsNotNil()
			continue
		} else {
			assert.For(t, i).ThatActual(err).IsNil()
		}

		assert.For(t, i).ThatActual(result).Equals(test.expected)
	}

}
