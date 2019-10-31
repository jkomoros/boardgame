package config

import (
	"testing"

	"github.com/workfit/tester/assert"
)

func TestFieldFromString(t *testing.T) {
	tests := []struct {
		in  string
		out ModeField
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
