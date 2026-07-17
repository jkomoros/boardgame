package path

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

func TestAbsoluteGoPkgPath(t *testing.T) {
	got, err := AbsoluteGoPkgPathWithOptions("github.com/jkomoros/boardgame/examples/pig", Options{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "pig" || !strings.Contains(filepath.ToSlash(got), "/examples/pig") {
		t.Fatalf("resolved path = %q, want examples/pig", got)
	}
}

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
