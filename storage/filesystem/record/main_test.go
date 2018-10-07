package record

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestFullEncoding(t *testing.T) {
	encodingTestHelper(t, StateEncodingFull, StateEncodingYudai)
}

func TestYudaiEncoding(t *testing.T) {
	encodingTestHelper(t, StateEncodingYudai, StateEncodingFull)
}

func encodingTestHelper(t *testing.T, encoding StateEncoding, others ...StateEncoding) {

	filename := "testdata/" + encoding.name() + ".json"

	rec, err := New(filename)

	assert.For(t).ThatActual(err).IsNil()

	canonicalRec, err := New("testdata/full.json")

	assert.For(t).ThatActual(err).IsNil()

	err = rec.compare(canonicalRec)

	assert.For(t).ThatActual(err).IsNil()

	//this should set detectedEncoder correctly
	enc := rec.encoder()
	assert.For(t).ThatActual(enc).IsNotNil()

	assert.For(t).ThatActual(rec.detectedEncoding).Equals(encoding)

	encodings := append([]StateEncoding{encoding}, others...)

	//We explicitly want to compare to ourselves to ensure that it's a no-op.
	for i, other := range encodings {
		otherRec, err := New(filename)

		assert.For(t, i).ThatActual(err).IsNil()

		err = otherRec.ConvertEncoding(other)

		assert.For(t, i).ThatActual(err).IsNil()

		err = rec.compare(otherRec)

		assert.For(t, i).ThatActual(err).IsNil()
	}

}
