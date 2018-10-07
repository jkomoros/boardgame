package record

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestFullEncoding(t *testing.T) {

	rec, err := New("testdata/full.json")

	assert.For(t).ThatActual(err).IsNil()

	//this should set detectedEncoder correctly
	enc := rec.encoder()
	assert.For(t).ThatActual(enc).IsNotNil()

	assert.For(t).ThatActual(rec.detectedEncoding).Equals(StateEncodingFull)

}
