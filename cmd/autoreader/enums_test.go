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
		expectedNewKeys []string
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
			nil,
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
			nil,
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
			nil,
		},
		{
			"Single implied layer",
			map[string]string{
				"Color":         "",
				"ColorBlue_One": "Blue > One",
			},
			map[string]string{
				"Color":         "",
				"ColorBlue":     "Blue",
				"ColorBlue_One": "One",
			},
			map[string]string{
				"Color":         "Color",
				"ColorBlue":     "Color",
				"ColorBlue_One": "ColorBlue",
			},
			[]string{
				"ColorBlue",
			},
		},
		{
			"Two implied layers",
			map[string]string{
				"Color":            "",
				"ColorGreen_One_A": "Green > One > A",
			},
			map[string]string{
				"Color":            "",
				"ColorGreen":       "Green",
				"ColorGreen_One":   "One",
				"ColorGreen_One_A": "A",
			},
			map[string]string{
				"Color":            "Color",
				"ColorGreen":       "Color",
				"ColorGreen_One":   "ColorGreen",
				"ColorGreen_One_A": "ColorGreen_One",
			},
			[]string{
				"ColorGreen",
				"ColorGreen_One",
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
			nil,
		},
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
			nil,
		},
		{
			"Implied node with implied nesting",
			map[string]string{
				"Color":             "",
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
			[]string{
				"ColorBlueGreen",
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
				"ColorBlueGreenOne":  "One",
				"ColorBlueGreenOneA": "A",
				"ColorBlueGreenOneB": "B",
				"ColorBlueGreenTwo":  "Two",
			},
			map[string]string{
				"Color":              "Color",
				"ColorBlueGreen":     "Color",
				"ColorBlueGreenOne":  "ColorBlueGreen",
				"ColorBlueGreenOneA": "ColorBlueGreenOne",
				"ColorBlueGreenOneB": "ColorBlueGreenOne",
				"ColorBlueGreenTwo":  "ColorBlueGreen",
			},
			nil,
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
				"Color":              "",
				"ColorBlueGreen":     "Blue Green",
				"ColorBlueGreenOne":  "One",
				"ColorBlueGreenOneA": "A",
				"ColorBlueGreenOneB": "B",
				"ColorBlueGreenTwo":  "Two",
			},
			map[string]string{
				"Color":              "Color",
				"ColorBlueGreen":     "Color",
				"ColorBlueGreenOne":  "ColorBlueGreen",
				"ColorBlueGreenOneA": "ColorBlueGreenOne",
				"ColorBlueGreenOneB": "ColorBlueGreenOne",
				"ColorBlueGreenTwo":  "ColorBlueGreen",
			},
			[]string{
				"ColorBlueGreenOne",
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
				"Color":               "",
				"ColorBlueGreen":      "Blue Green",
				"ColorBlueGreenOne":   "One",
				"ColorBlueGreenOne_A": "A",
				"ColorBlueGreenOne_B": "B",
				"ColorBlueGreenTwo":   "Two",
			},
			map[string]string{
				"Color":               "Color",
				"ColorBlueGreen":      "Color",
				"ColorBlueGreenOne":   "ColorBlueGreen",
				"ColorBlueGreenOne_A": "ColorBlueGreenOne",
				"ColorBlueGreenOne_B": "ColorBlueGreenOne",
				"ColorBlueGreenTwo":   "ColorBlueGreen",
			},
			nil,
		},
		{
			"Multiple implied layers in a row",
			map[string]string{
				"Color":              "",
				"ColorBlueGreenOneA": "Blue Green One A",
				"ColorBlueGreenOneB": "Blue Green One B",
				"ColorBlueGreenTwo":  "Blue Green Two",
			},
			map[string]string{
				"Color":              "",
				"ColorBlueGreen":     "Blue Green",
				"ColorBlueGreenOne":  "One",
				"ColorBlueGreenOneA": "A",
				"ColorBlueGreenOneB": "B",
				"ColorBlueGreenTwo":  "Two",
			},
			map[string]string{
				"Color":              "Color",
				"ColorBlueGreen":     "Color",
				"ColorBlueGreenOne":  "ColorBlueGreen",
				"ColorBlueGreenOneA": "ColorBlueGreenOne",
				"ColorBlueGreenOneB": "ColorBlueGreenOne",
				"ColorBlueGreenTwo":  "ColorBlueGreen",
			},
			[]string{
				"ColorBlueGreen",
				"ColorBlueGreenOne",
			},
		},
		{
			"Elided parent beneath non-elided multi-word node",
			map[string]string{
				"Color":              "",
				"ColorBlueGreenOne":  "Blue Green One",
				"ColorBlueGreenOneA": "Blue Green One A",
				"ColorBlueGreenOneB": "Blue Green One B",
				"ColorBlueGreenTwoA": "Blue Green Two A",
			},
			map[string]string{
				"Color":              "",
				"ColorBlueGreen":     "Blue Green",
				"ColorBlueGreenOne":  "One",
				"ColorBlueGreenOneA": "A",
				"ColorBlueGreenOneB": "B",
				"ColorBlueGreenTwoA": "Two A",
			},
			map[string]string{
				"Color":              "Color",
				"ColorBlueGreen":     "Color",
				"ColorBlueGreenOne":  "ColorBlueGreen",
				"ColorBlueGreenOneA": "ColorBlueGreenOne",
				"ColorBlueGreenOneB": "ColorBlueGreenOne",
				"ColorBlueGreenTwoA": "ColorBlueGreen",
			},
			[]string{
				"ColorBlueGreen",
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
		assert.For(t, i).ThatActual(e.NewKeys()).Equals(test.expectedNewKeys).ThenDiffOnFail()
	}
}

func TestReduceProposedKey(t *testing.T) {
	tests := []struct {
		in        string
		otherKeys []string
		expected  string
	}{
		{
			"PhaseBlueGreen",
			nil,
			"PhaseBlueGreen",
		},
		{
			"PhaseBlueGreen_One",
			[]string{
				"Phase",
				"PhaseBlueGreen",
				"PhaseBlueGreenOneA",
			},
			"PhaseBlueGreenOne",
		},
		{
			"PhaseBlueGreen_One",
			[]string{
				"Phase",
				"PhaseBlueGreen",
				"PhaseBlueGreen_OneA",
			},
			"PhaseBlueGreen_One",
		},
		{
			"PhaseBlueGreen_One_A",
			[]string{
				"Phase",
				"PhaseBlueGreen",
				"PhaseBlueGreenOneAOne",
			},
			"PhaseBlueGreenOneA",
		},
		{
			"PhaseBlueGreen_One_A",
			[]string{
				"Phase",
				"PhaseBlueGreen",
				"PhaseBlueGreenOne_AOne",
			},
			"PhaseBlueGreenOne_A",
		},
	}

	for i, test := range tests {
		e := newEnum("test", transformNone)
		for _, key := range test.otherKeys {
			e.AddTransformKey(key, false, "", transformNone)
		}
		actual := e.reduceProposedKey(test.in)
		assert.For(t, i).ThatActual(actual).Equals(test.expected)
	}
}
