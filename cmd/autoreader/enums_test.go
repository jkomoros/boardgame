package main

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestEnumParent(t *testing.T) {
	tests := []struct {
		strValues       map[string]string
		expectedParents map[string]string
	}{
		{
			map[string]string{
				"Phase":         "",
				"PhaseAnother":  "Another",
				"PhaseOverride": "Heyo",
			},
			map[string]string{
				"Phase":         "Phase",
				"PhaseAnother":  "Phase",
				"PhaseOverride": "Phase",
			},
		},
		{
			map[string]string{
				"ColorBlue":  "Blue",
				"ColorGreen": "Green",
				"ColorRed":   "Red",
			},
			nil,
		},
	}

	for i, test := range tests {
		result := createParents(test.strValues)
		assert.For(t, i).ThatActual(result).Equals(test.expectedParents)
	}
}
