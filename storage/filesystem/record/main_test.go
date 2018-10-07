package record

import (
	"github.com/workfit/tester/assert"
	"testing"
)

const fullJsonFilename = "testdata/full.json"
const diffJsonFilename = "testdata/diff.json"

func TestEncoding(t *testing.T) {

	canonicalRec, err := New(fullJsonFilename)

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

	canonicalRec, err := New(fullJsonFilename)

	assert.For(t).ThatActual(err).IsNil()

	rec, err := New(fullJsonFilename)
	assert.For(t).ThatActual(err).IsNil()

	err = rec.Compress()
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(rec.FullStateEncoding()).IsFalse()

	err = canonicalRec.compare(rec)
	assert.For(t).ThatActual(err).IsNil()

}

func TestExpand(t *testing.T) {

	canonicalRec, err := New(fullJsonFilename)

	assert.For(t).ThatActual(err).IsNil()

	rec, err := New("testdata/diff.json")
	assert.For(t).ThatActual(err).IsNil()

	err = rec.Expand()
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(rec.FullStateEncoding()).IsTrue()

	err = canonicalRec.compare(rec)
	assert.For(t).ThatActual(err).IsNil()

}
