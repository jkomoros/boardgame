package golden

import (
	"bytes"
	"errors"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
	"text/template"
)

//Path relative to package root where goldens should be stored
const GameRecordsFolder = "testdata/golden"

//The name of the test file to create
const GoldenTestFile = "golden_test.go"

var goldenTestTemplate *template.Template

func init() {
	goldenTestTemplate = template.Must(template.New("golden").Parse(goldenTestTemplateText))
}

//MakeGoldenTest ensures that GameRecordFolder exists and creates
//golden_test.go in the root of the package if it doesn't yet exist.
func MakeGoldenTest(pkg *gamepkg.Pkg) error {

	if pkg == nil {
		return errors.New("No package provided")
	}

	if err := pkg.EnsureDir(GameRecordsFolder); err != nil {
		return errors.New("Couldn't ensure game records dir: " + err.Error())
	}

	buf := new(bytes.Buffer)

	err := goldenTestTemplate.Execute(buf, map[string]string{
		"gametype": pkg.Name(),
		"folder":   GameRecordsFolder,
	})

	if err != nil {
		return errors.New("Couldn't generate blob from template: " + err.Error())
	}

	return pkg.WriteFile(GoldenTestFile, buf.Bytes(), true)

}

var goldenTestTemplateText = `package {{.gametype}}

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
	if err := golden.CompareFolder(NewDelegate(), "{{.folder}}"); err != nil {
		t.Error(err.Error())
	}
}
`
