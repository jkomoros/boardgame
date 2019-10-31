package record

import (
	"testing"

	"github.com/workfit/tester/assert"
)

const fullJSONFilename = "testdata/full.json"
const diffJSONFilename = "testdata/diff.json"

func TestEncoding(t *testing.T) {

	canonicalRec, err := New(fullJSONFilename)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(canonicalRec.FullStateEncoding()).IsTrue()

	exampleCanonicalPatch := canonicalRec.data.StatePatches[0]

	//Ensure that the full and diff matchers are mutually exclusive
	err = fullEncoder.Matches(exampleCanonicalPatch)
	assert.For(t).ThatActual(err).IsNil()
	err = diffEncoder.Matches(exampleCanonicalPatch)
	assert.For(t).ThatActual(err).IsNotNil()

	rec, err := New("testdata/diff.json")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(rec.FullStateEncoding()).IsFalse()

	exampleDiffPatch := rec.data.StatePatches[0]

	//Ensure that the full and diff matchers are mutually exclusive
	err = diffEncoder.Matches(exampleDiffPatch)
	assert.For(t).ThatActual(err).IsNil()
	err = fullEncoder.Matches(exampleDiffPatch)
	assert.For(t).ThatActual(err).IsNotNil()

	err = rec.compare(canonicalRec)

	assert.For(t).ThatActual(err).IsNil()

}

func TestCompress(t *testing.T) {

	canonicalRec, err := New(fullJSONFilename)

	assert.For(t).ThatActual(err).IsNil()

	rec, err := New(fullJSONFilename)
	assert.For(t).ThatActual(err).IsNil()

	err = rec.Compress()
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(rec.FullStateEncoding()).IsFalse()

	err = canonicalRec.compare(rec)
	assert.For(t).ThatActual(err).IsNil()

}

func TestExpand(t *testing.T) {

	canonicalRec, err := New(fullJSONFilename)

	assert.For(t).ThatActual(err).IsNil()

	rec, err := New("testdata/diff.json")
	assert.For(t).ThatActual(err).IsNil()

	err = rec.Expand()
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(rec.FullStateEncoding()).IsTrue()

	err = canonicalRec.compare(rec)
	assert.For(t).ThatActual(err).IsNil()

}

//TODO: actually test the behavior where we flip from compressed to expanded
//mode in AddCurentGameAndSave. Have the encoder fail to confirm in that mode
//to force the error?
