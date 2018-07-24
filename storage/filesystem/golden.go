package filesystem

import (
	"text/template"
)

var goldenTestTemplate *template.Template

func init() {
	goldenTestTemplate = template.Must(template.New("golden").Parse(goldenTestTemplateText))
}

var goldenTestTemplateText = `package {{.gametype}}

/*

This file was created automatically by the filesystem storage layer with a
golden folder.

It will be overwritten the next time a filesystem is booted that uses this
game package.

*/

import (
	"github.com/jkomoros/boardgame/util/golden"
	"testing"
)

func TestGolden(t *testing.T) {
	if err := golden.CompareFolder(NewDelegate(), "{{.folder}}"); err != nil {
		t.Error(err.Error())
	}
}
`
