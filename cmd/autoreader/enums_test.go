package main

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestEnumParent(t *testing.T) {
	tests := []struct {
		strValues map[string]string
		//If nil, expect no change from strValues
		expectedValues  map[string]string
		expectedParents map[string]string
	}{
		{
			map[string]string{
				"Phase":         "",
				"PhaseAnother":  "Another",
				"PhaseOverride": "Heyo",
			},
			nil,
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
			nil,
		},
		{
			map[string]string{
				"Color":          "",
				"ColorBlue":      "Blue",
				"ColorBlue_One":  "Blue > One",
				"ColorBlue_Two":  "Blue > Two",
				"ColorGreen":     "Green",
				"ColorGreen_One": "Green > One",
			},
			map[string]string{
				"Color":          "",
				"ColorBlue":      "Blue",
				"ColorBlue_One":  "One",
				"ColorBlue_Two":  "Two",
				"ColorGreen":     "Green",
				"ColorGreen_One": "One",
			},
			map[string]string{
				"Color":          "Color",
				"ColorBlue":      "Color",
				"ColorBlue_One":  "ColorBlue",
				"ColorBlue_Two":  "ColorBlue",
				"ColorGreen":     "Color",
				"ColorGreen_One": "ColorGreen",
			},
		},
		{
			map[string]string{
				"Color":             "",
				"ColorBlue":         "Blue",
				"ColorBlue_One":     "Blue > One",
				"ColorBlue_Two":     "Blue > Two",
				"ColorBlue_One_One": "Blue > One > One",
				"ColorBlue_One_Two": "Blue > One > Two",
			},
			map[string]string{
				"Color":             "",
				"ColorBlue":         "Blue",
				"ColorBlue_One":     "One",
				"ColorBlue_Two":     "Two",
				"ColorBlue_One_One": "One",
				"ColorBlue_One_Two": "Two",
			},
			map[string]string{
				"Color":             "Color",
				"ColorBlue":         "Color",
				"ColorBlue_One":     "ColorBlue",
				"ColorBlue_Two":     "ColorBlue",
				"ColorBlue_One_One": "ColorBlue_One",
				"ColorBlue_One_Two": "ColorBlue_One",
			},
		},
	}

	for i, test := range tests {
		actualValues, actualParents := createParents(test.strValues)
		if test.expectedValues == nil {
			//Expect no change from test.strValues
			assert.For(t, i).ThatActual(actualValues).Equals(test.strValues).ThenDiffOnFail()
		} else {
			assert.For(t, i).ThatActual(actualValues).Equals(test.expectedValues).ThenDiffOnFail()
		}
		assert.For(t, i).ThatActual(actualParents).Equals(test.expectedParents).ThenDiffOnFail()
	}
}
