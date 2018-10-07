package record

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestFullEncoding(t *testing.T) {
	encodingTestHelper(t, StateEncodingFull)
}

func TestYudaiEncoding(t *testing.T) {
	encodingTestHelper(t, StateEncodingYudai)
}

func encodingTestHelper(t *testing.T, encoding StateEncoding) {

	filename := "testdata/" + encoding.name() + ".json"

	rec, err := New(filename)

	assert.For(t).ThatActual(err).IsNil()

	//this should set detectedEncoder correctly
	enc := rec.encoder()
	assert.For(t).ThatActual(enc).IsNotNil()

	assert.For(t).ThatActual(rec.detectedEncoding).Equals(encoding)
}
