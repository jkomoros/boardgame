package codegen

import (
	"testing"

	"github.com/workfit/tester/assert"
)

func TestOldDelimiterFails(t *testing.T) {

	values := map[string]string{
		"OldEnum_Key":    "Key",
		"OldEnum010KeyA": "KeyA",
	}

	e := newEnum("test", transformNone)

	for key, val := range values {
		e.AddTransformKey(key, true, val, transformNone)
	}

	err := e.Process()

	//Make sure that enumsw ith old string delimiters will error
	assert.For(t).ThatActual(err).IsNotNil()
}

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
				"Color":            "",
				"ColorBlue":        "Blue",
				"ColorBlue010One":  "Blue > One",
				"ColorBlue010Two":  "Blue > Two",
				"ColorGreen":       "Green",
				"ColorGreen010One": "Green > One",
			},
			map[string]string{
				"Color":            "",
				"ColorBlue":        "Blue",
				"ColorBlue010One":  "One",
				"ColorBlue010Two":  "Two",
				"ColorGreen":       "Green",
				"ColorGreen010One": "One",
			},
			map[string]string{
				"Color":            "Color",
				"ColorBlue":        "Color",
				"ColorBlue010One":  "ColorBlue",
				"ColorBlue010Two":  "ColorBlue",
				"ColorGreen":       "Color",
				"ColorGreen010One": "ColorGreen",
			},
			nil,
		},
		{
			"Three layers",
			map[string]string{
				"Color":                 "",
				"ColorBlue":             "Blue",
				"ColorBlue010One":       "Blue > One",
				"ColorBlue010Two":       "Blue > Two",
				"ColorBlue010One010One": "Blue > One > One",
				"ColorBlue010One010Two": "Blue > One > Two",
			},
			map[string]string{
				"Color":                 "",
				"ColorBlue":             "Blue",
				"ColorBlue010One":       "One",
				"ColorBlue010Two":       "Two",
				"ColorBlue010One010One": "One",
				"ColorBlue010One010Two": "Two",
			},
			map[string]string{
				"Color":                 "Color",
				"ColorBlue":             "Color",
				"ColorBlue010One":       "ColorBlue",
				"ColorBlue010Two":       "ColorBlue",
				"ColorBlue010One010One": "ColorBlue010One",
				"ColorBlue010One010Two": "ColorBlue010One",
			},
			nil,
		},
		{
			"Single implied layer",
			map[string]string{
				"Color":           "",
				"ColorBlue010One": "Blue > One",
			},
			map[string]string{
				"Color":           "",
				"ColorBlue":       "Blue",
				"ColorBlue010One": "One",
			},
			map[string]string{
				"Color":           "Color",
				"ColorBlue":       "Color",
				"ColorBlue010One": "ColorBlue",
			},
			[]string{
				"ColorBlue",
			},
		},
		{
			"Two implied layers",
			map[string]string{
				"Color":                "",
				"ColorGreen010One010A": "Green > One > A",
			},
			map[string]string{
				"Color":                "",
				"ColorGreen":           "Green",
				"ColorGreen010One":     "One",
				"ColorGreen010One010A": "A",
			},
			map[string]string{
				"Color":                "Color",
				"ColorGreen":           "Color",
				"ColorGreen010One":     "ColorGreen",
				"ColorGreen010One010A": "ColorGreen010One",
			},
			[]string{
				"ColorGreen",
				"ColorGreen010One",
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
				"Color":                 "",
				"ColorBlueGreen":        "Blue Green",
				"ColorBlueGreenOne":     "Blue Green One",
				"ColorBlueGreenOne010A": "Blue Green One > A",
				"ColorBlueGreenOne010B": "Blue Green One > B",
				"ColorBlueGreenTwo":     "Blue Green Two",
			},
			map[string]string{
				"Color":                 "",
				"ColorBlueGreen":        "Blue Green",
				"ColorBlueGreenOne":     "One",
				"ColorBlueGreenOne010A": "A",
				"ColorBlueGreenOne010B": "B",
				"ColorBlueGreenTwo":     "Two",
			},
			map[string]string{
				"Color":                 "Color",
				"ColorBlueGreen":        "Color",
				"ColorBlueGreenOne":     "ColorBlueGreen",
				"ColorBlueGreenOne010A": "ColorBlueGreenOne",
				"ColorBlueGreenOne010B": "ColorBlueGreenOne",
				"ColorBlueGreenTwo":     "ColorBlueGreen",
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
			"PhaseBlueGreen010One",
			[]string{
				"Phase",
				"PhaseBlueGreen",
				"PhaseBlueGreenOneA",
			},
			"PhaseBlueGreenOne",
		},
		{
			"PhaseBlueGreen010One",
			[]string{
				"Phase",
				"PhaseBlueGreen",
				"PhaseBlueGreen010OneA",
			},
			"PhaseBlueGreen010One",
		},
		{
			"PhaseBlueGreen010One010A",
			[]string{
				"Phase",
				"PhaseBlueGreen",
				"PhaseBlueGreenOneAOne",
			},
			"PhaseBlueGreenOneA",
		},
		{
			"PhaseBlueGreen010One010A",
			[]string{
				"Phase",
				"PhaseBlueGreen",
				"PhaseBlueGreenOne010AOne",
			},
			"PhaseBlueGreenOne010A",
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
