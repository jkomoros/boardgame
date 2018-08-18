package config

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestFieldFromString(t *testing.T) {
	tests := []struct {
		in  string
		out ConfigModeField
	}{
		{
			"Firebase",
			FieldFirebase,
		},
		{
			" FireBASE ",
			FieldFirebase,
		},
		{
			" F1REBAS3",
			FieldInvalid,
		},
	}

	for i, test := range tests {
		result := FieldFromString(test.in)
		assert.For(t, i).ThatActual(result).Equals(test.out)

	}
}
