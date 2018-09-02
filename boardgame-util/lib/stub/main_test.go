package stub

import (
	"github.com/workfit/tester/assert"
	"testing"
)

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

	mainGo := contents["checkers/main.go"]

	assert.For(t).ThatActual(mainGo).IsNotNil()

}
