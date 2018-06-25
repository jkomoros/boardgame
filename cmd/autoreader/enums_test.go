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
		{
			map[string]string{
				"Color":         "",
				"ColorBlue_One": "Blue > One",
			},
			map[string]string{
				"Color":                "",
				"-9223372036854775808": "Blue",
				"ColorBlue_One":        "One",
			},
			map[string]string{
				"Color":                "Color",
				"-9223372036854775808": "Color",
				"ColorBlue_One":        "-9223372036854775808",
			},
		},
		{
			map[string]string{
				"Color":            "",
				"ColorGreen_One_A": "Green > One > A",
			},
			map[string]string{
				"Color":                "",
				"-9223372036854775808": "Green",
				"-9223372036854775807": "One",
				"ColorGreen_One_A":     "A",
			},
			map[string]string{
				"Color":                "Color",
				"-9223372036854775808": "Color",
				"-9223372036854775807": "-9223372036854775808",
				"ColorGreen_One_A":     "-9223372036854775807",
			},
		},
	}

	for i, test := range tests {

		e := newEnum("test", transformNone)

		for key, val := range test.strValues {
			e.AddTransformKey(key, true, val, transformNone)
		}

		err := e.Process()
		assert.For(t, i).ThatActual(err).IsNil()

		actualValues := e.ValueMap()
		actualParents := e.Parents()
		if test.expectedValues == nil {
			//Expect no change from test.strValues
			assert.For(t, i).ThatActual(actualValues).Equals(test.strValues).ThenDiffOnFail()
		} else {
			assert.For(t, i).ThatActual(actualValues).Equals(test.expectedValues).ThenDiffOnFail()
		}
		assert.For(t, i).ThatActual(actualParents).Equals(test.expectedParents).ThenDiffOnFail()
	}
}
