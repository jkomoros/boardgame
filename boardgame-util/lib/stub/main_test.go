package stub

import (
	"github.com/workfit/tester/assert"
	"testing"
)

//If true, will save out the files generated. Useful for generating new golden
//output when output is changed.
const generateNewGolden = false

func TestGenerate(t *testing.T) {

	tmpls, err := DefaultTemplateSet()

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(tmpls)).Equals(1)

	opt := &Options{
		Name: "checkers",
	}

	contents, err := tmpls.Generate(opt)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(contents)).Equals(1)

	assert.For(t).ThatActual(contents["checkers/main.go"]).IsNotNil()

	//Now that we unit tested underlying stuff, use Generate() top level,
	//which also formats.

	contents, err = Generate(opt)

	assert.For(t).ThatActual(err).IsNil()

	if generateNewGolden {

		//Save out contents as new golden files to compare against
		contents.Save("test")

		return
	}

	mainGo := contents["checkers/main.go"]

	assert.For(t).ThatActual(mainGo).IsNotNil()

}
