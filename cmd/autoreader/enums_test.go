package main

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestEnumParent(t *testing.T) {
	tests := []struct {
		description string
		strValues   map[string]string
		//If nil, expect no change from strValues
		expectedValues  map[string]string
		expectedParents map[string]string
	}{
		{
			"Single layer",
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
			"No Tree",
			map[string]string{
				"ColorBlue":  "Blue",
				"ColorGreen": "Green",
				"ColorRed":   "Red",
			},
			nil,
			nil,
		},
		{
			"Two layers",
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
			"Three layers",
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
			"Single implied layer",
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
			"Two implied layers",
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
		{
			"Single word implied nesting",
			map[string]string{
				"Color":        "",
				"ColorBlue":    "Blue",
				"ColorBlueOne": "Blue One",
				"ColorBlueTwo": "Blue Two",
			},
			map[string]string{
				"Color":        "",
				"ColorBlue":    "Blue",
				"ColorBlueOne": "One",
				"ColorBlueTwo": "Two",
			},
			map[string]string{
				"Color":        "Color",
				"ColorBlue":    "Color",
				"ColorBlueOne": "ColorBlue",
				"ColorBlueTwo": "ColorBlue",
			},
		},
		/*
			{
				"Multi-Word implied nesting",
				map[string]string{
					"Color":             "",
					"ColorBlueGreen":    "Blue Green",
					"ColorBlueGreenOne": "Blue Green One",
					"ColorBlueGreenTwo": "Blue Green Two",
				},
				map[string]string{
					"Color":             "",
					"ColorBlueGreen":    "Blue Green",
					"ColorBlueGreenOne": "One",
					"ColorBlueGreenTwo": "Two",
				},
				map[string]string{
					"Color":             "Color",
					"ColorBlueGreen":    "Color",
					"ColorBlueGreenOne": "ColorBlueGreen",
					"ColorBlueGreenTwo": "ColorBlueGreen",
				},
			},
			{
				"Implied node with implied nesting",
				map[string]string{
					"Color":             "",
					"ColorBlueGreenOne": "Blue Green One",
					"ColorBlueGreenTwo": "Blue Green Two",
				},
				map[string]string{
					"Color":                "",
					"-9223372036854775808": "Blue Green",
					"ColorBlueGreenOne":    "Blue Green > One",
					"ColorBlueGreenTwo":    "Blue Green > Two",
				},
				map[string]string{
					"Color":                "Color",
					"-9223372036854775808": "Color",
					"ColorBlueGreenOne":    "-9223372036854775808",
					"ColorBlueGreenTwo":    "-9223372036854775808",
				},
			},
			{
				"Multiple implied layers",
				map[string]string{
					"Color":              "",
					"ColorBlueGreen":     "Blue Green",
					"ColorBlueGreenOne":  "Blue Green One",
					"ColorBlueGreenOneA": "Blue Green One A",
					"ColorBlueGreenOneB": "Blue Green One B",
					"ColorBlueGreenTwo":  "Blue Green Two",
				},
				map[string]string{
					"Color":              "",
					"ColorBlueGreen":     "Blue Green",
					"ColorBlueGreenOne":  "Blue Green > One",
					"ColorBlueGreenOneA": "Blue Green > One > A",
					"ColorBlueGreenOneB": "Blue Green > One > B",
					"ColorBlueGreenTwo":  "Blue Green > Two",
				},
				map[string]string{
					"Color":              "Color",
					"ColorBlueGreen":     "Color",
					"ColorBlueGreenOne":  "ColorBlueGreen",
					"ColorBlueGreenOneA": "ColorBlueGreenOne",
					"ColorBlueGreenOneB": "ColorBlueGreenOne",
					"ColorBlueGreenTwo":  "ColorBlueGreen",
				},
			},
			{
				"Multiple implied layers with implied node",
				map[string]string{
					"Color":              "",
					"ColorBlueGreen":     "Blue Green",
					"ColorBlueGreenOneA": "Blue Green One A",
					"ColorBlueGreenOneB": "Blue Green One B",
					"ColorBlueGreenTwo":  "Blue Green Two",
				},
				map[string]string{
					"Color":                "",
					"ColorBlueGreen":       "Blue Green",
					"-9223372036854775808": "Blue Green > One",
					"ColorBlueGreenOneA":   "Blue Green > One > A",
					"ColorBlueGreenOneB":   "Blue Green > One > B",
					"ColorBlueGreenTwo":    "Blue Green > Two",
				},
				map[string]string{
					"Color":                "Color",
					"ColorBlueGreen":       "Color",
					"-9223372036854775808": "ColorBlueGreen",
					"ColorBlueGreenOneA":   "-9223372036854775808",
					"ColorBlueGreenOneB":   "-9223372036854775808",
					"ColorBlueGreenTwo":    "ColorBlueGreen",
				},
			},
			{
				"Mix implicit and explicit layers",
				map[string]string{
					"Color":               "",
					"ColorBlueGreen":      "Blue Green",
					"ColorBlueGreenOne":   "Blue Green One",
					"ColorBlueGreenOne_A": "Blue Green One > A",
					"ColorBlueGreenOne_B": "Blue Green One > B",
					"ColorBlueGreenTwo":   "Blue Green Two",
				},
				map[string]string{
					"Color":              "",
					"ColorBlueGreen":     "Blue Green",
					"ColorBlueGreenOne":  "Blue Green > One",
					"ColorBlueGreenOneA": "Blue Green > One > A",
					"ColorBlueGreenOneB": "Blue Green > One > B",
					"ColorBlueGreenTwo":  "Blue Green > Two",
				},
				map[string]string{
					"Color":              "Color",
					"ColorBlueGreen":     "Color",
					"ColorBlueGreenOne":  "ColorBlueGreen",
					"ColorBlueGreenOneA": "ColorBlueGreenOne",
					"ColorBlueGreenOneB": "ColorBlueGreenOne",
					"ColorBlueGreenTwo":  "ColorBlueGreen",
				},
			},
		*/
	}

	for i, test := range tests {

		if i > 5 {
			continue
		}

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
