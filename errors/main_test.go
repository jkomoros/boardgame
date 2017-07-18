package errors

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestFriendly(t *testing.T) {

	testMessage := "My Message"

	f := New(testMessage)

	assert.For(t).ThatActual(f).IsNotNil()

	assert.For(t).ThatActual(f.Error()).Equals(testMessage)

	assert.For(t).ThatActual(f.FriendlyError()).Equals(DefaultFriendlyError)

	assert.For(t).ThatActual(f.SecureError()).Equals(testMessage)

	assert.For(t).ThatActual(f.Fields()).Equals(Fields{})

}
