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

	friendlyTestMessage := "MY friendly message"

	fields := Fields{"Foo": true}

	friendly := f.WithFriendly(friendlyTestMessage, fields)

	assert.For(t).ThatActual(friendly).IsNotNil()

	assert.For(t).ThatActual(friendly).DoesNotEqual(f)

	assert.For(t).ThatActual(friendly.FriendlyError()).Equals(friendlyTestMessage)

	assert.For(t).ThatActual(friendly.Error()).Equals(testMessage)

	assert.For(t).ThatActual(f.Fields()).Equals(Fields{})

	assert.For(t).ThatActual(friendly.Fields()).Equals(fields)

}
