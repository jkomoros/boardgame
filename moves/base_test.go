package moves

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestMoveProgression(t *testing.T) {

	tests := []struct {
		progression    []string
		pattern        []string
		expectedResult bool
	}{
		{
			[]string{
				"A",
			},
			[]string{
				"A",
				"B",
				"C",
			},
			true,
		},
		{
			[]string{
				"B",
			},
			[]string{
				"A",
				"B",
				"C",
			},
			false,
		},
		{
			[]string{
				"A",
				"A",
				"A",
			},
			[]string{
				"A",
				"B",
				"C",
			},
			true,
		},
		{
			[]string{
				"A",
				"A",
				"B",
			},
			[]string{
				"A",
				"B",
				"C",
			},
			true,
		},
		{
			[]string{
				"A",
				"A",
				"C",
			},
			[]string{
				"A",
				"B",
				"C",
			},
			false,
		},
		{
			[]string{
				"A",
				"A",
				"B",
				"A",
			},
			[]string{
				"A",
				"B",
				"A",
				"C",
			},
			true,
		},
		{
			[]string{
				"A",
				"A",
				"B",
				"B",
			},
			[]string{
				"A",
				"B",
				"A",
				"C",
			},
			true,
		},
		{
			[]string{
				"A",
				"A",
				"B",
				"C",
			},
			[]string{
				"A",
				"B",
				"A",
				"C",
			},
			false,
		},
		{
			[]string{
				"Multi Word Move",
				"B",
			},
			[]string{
				"Multi Word Move",
				"B",
				"Multi Word Move",
			},
			true,
		},
	}

	for i, test := range tests {
		result := progressionMatches(test.progression, test.pattern)
		assert.For(t, i).ThatActual(result).Equals(test.expectedResult)
	}

}
