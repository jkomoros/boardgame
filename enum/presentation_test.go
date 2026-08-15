package enum

import (
	"testing"

	"github.com/workfit/tester/assert"
)

// climate mirrors the shape of a real game enum: iota-declared, so the key
// order is the declared order, and deliberately NOT alphabetical, so a test
// that passes for a sorted-by-string list fails here.
const (
	climateUnknown = iota
	climateIceAge
	climateFreezing
	climateCold
	climateTemperate
	climateWarm
	climateScorching
)

func climateSet() (*Set, Enum) {
	set := NewSet()
	//Deliberately not written in key order, and stored in a map, so the only
	//thing that can produce a stable order downstream is the enum itself.
	e := set.MustAdd("climate", map[EnumKey]string{
		climateWarm:      "Warm",
		climateUnknown:   "Unknown",
		climateScorching: "Scorching",
		climateCold:      "Cold",
		climateIceAge:    "Ice Age",
		climateTemperate: "Temperate",
		climateFreezing:  "Freezing",
	})
	return set, e
}

// TestValuesAreInDeclaredOrder is the guard against the bug this package has
// already shipped once: ranging over the values map hands a different order to
// every caller on every run. Values() is the single ordering promise the
// client's ordered enum lists rest on, so it is checked repeatedly -- a single
// pass could pass by luck under Go's randomized map iteration.
func TestValuesAreInDeclaredOrder(t *testing.T) {

	_, e := climateSet()

	want := []EnumKey{
		climateUnknown, climateIceAge, climateFreezing, climateCold,
		climateTemperate, climateWarm, climateScorching,
	}

	wantStrings := []string{
		"Unknown", "Ice Age", "Freezing", "Cold", "Temperate", "Warm", "Scorching",
	}

	for i := 0; i < 50; i++ {
		values := e.Values()
		assert.For(t, i).ThatActual(values).Equals(want).ThenDiffOnFail()

		var strings []string
		for _, value := range values {
			strings = append(strings, e.String(value))
		}
		assert.For(t, i).ThatActual(strings).Equals(wantStrings).ThenDiffOnFail()
	}
}

// TestValuesIsNotSortedByString proves the ordering under test is declaration
// order and not some other stable order that happens to look right for simple
// enums.
func TestValuesIsNotSortedByString(t *testing.T) {

	_, e := climateSet()

	values := e.Values()

	assert.For(t).ThatActual(e.String(values[0])).Equals("Unknown")
	assert.For(t).ThatActual(e.String(values[len(values)-1])).Equals("Scorching")
	//Sorted by string, "Cold" would come first and "Warm" last.
	assert.For(t).ThatActual(e.String(values[0])).DoesNotEqual("Cold")
}

func TestPresentationDefaultsToZero(t *testing.T) {

	_, e := climateSet()

	assert.For(t).ThatActual(e.HasPresentation()).IsFalse()

	for _, value := range e.Values() {
		presentation := e.Presentation(value)
		assert.For(t, value).ThatActual(presentation.Zero()).IsTrue()
		assert.For(t, value).ThatActual(presentation).Equals(Presentation{})
	}

	//An out-of-range value is also just the zero Presentation, not a panic.
	assert.For(t).ThatActual(e.Presentation(IllegalValue).Zero()).IsTrue()
}

func TestSetPresentation(t *testing.T) {

	set, e := climateSet()

	err := set.SetPresentation("climate", map[EnumKey]Presentation{
		climateIceAge:    {Label: "Ice", Art: "img/ice-age.png"},
		climateScorching: {CSSColor: "#ef936e", Description: "Nothing survives"},
	})
	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(e.HasPresentation()).IsTrue()

	assert.For(t).ThatActual(e.Presentation(climateIceAge)).Equals(Presentation{
		Label: "Ice",
		Art:   "img/ice-age.png",
	})
	assert.For(t).ThatActual(e.Presentation(climateScorching)).Equals(Presentation{
		CSSColor:    "#ef936e",
		Description: "Nothing survives",
	})

	//Values that were not mentioned round-trip as the zero Presentation, which
	//is what lets a game attach art to only the values that have any.
	assert.For(t).ThatActual(e.Presentation(climateCold).Zero()).IsTrue()
	assert.For(t).ThatActual(e.Presentation(climateUnknown).Zero()).IsTrue()

	//Ordering is untouched by attaching presentation.
	assert.For(t).ThatActual(e.Values()[0]).Equals(EnumKey(climateUnknown))
	assert.For(t).ThatActual(e.Values()[1]).Equals(EnumKey(climateIceAge))
}

func TestSetPresentationMerges(t *testing.T) {

	set, e := climateSet()

	assert.For(t).ThatActual(set.SetPresentation("climate", map[EnumKey]Presentation{
		climateCold: {CSSColor: "#d9eef5"},
	})).IsNil()
	assert.For(t).ThatActual(set.SetPresentation("climate", map[EnumKey]Presentation{
		climateWarm: {CSSColor: "#ef936e"},
		climateCold: {CSSColor: "#ffffff", Art: "img/cold.png"},
	})).IsNil()

	assert.For(t).ThatActual(e.Presentation(climateWarm).CSSColor).Equals("#ef936e")
	//A later entry for the same key wins outright, rather than merging fields.
	assert.For(t).ThatActual(e.Presentation(climateCold)).Equals(Presentation{
		CSSColor: "#ffffff",
		Art:      "img/cold.png",
	})
}

func TestSetPresentationErrors(t *testing.T) {

	set, _ := climateSet()

	assert.For(t).ThatActual(set.SetPresentation("nonexistent", map[EnumKey]Presentation{
		climateCold: {CSSColor: "red"},
	})).IsNotNil()

	assert.For(t).ThatActual(set.SetPresentation("climate", map[EnumKey]Presentation{
		1000: {CSSColor: "red"},
	})).IsNotNil()

	//A rejected call must not partially apply.
	assert.For(t).ThatActual(set.Enum("climate").HasPresentation()).IsFalse()

	set.Finish()

	assert.For(t).ThatActual(set.SetPresentation("climate", map[EnumKey]Presentation{
		climateCold: {CSSColor: "red"},
	})).IsNotNil()
}

func TestMustSetPresentationPanics(t *testing.T) {

	set, _ := climateSet()

	defer func() {
		assert.For(t).ThatActual(recover()).IsNotNil()
	}()

	set.MustSetPresentation("nonexistent", map[EnumKey]Presentation{})

	t.Error("MustSetPresentation did not panic for an unknown enum")
}

func TestCombineCarriesPresentation(t *testing.T) {

	set := NewSet()

	first := set.MustAdd("first", map[EnumKey]string{
		0: "Alpha",
		1: "Beta",
	})
	set.MustAdd("second", map[EnumKey]string{
		2: "Gamma",
	})

	assert.For(t).ThatActual(set.SetPresentation("first", map[EnumKey]Presentation{
		1: {CSSColor: "blue"},
	})).IsNil()

	combined, err := set.Combine("combined", first, set.Enum("second"))
	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(combined.HasPresentation()).IsTrue()
	assert.For(t).ThatActual(combined.Presentation(1).CSSColor).Equals("blue")
	assert.For(t).ThatActual(combined.Presentation(2).Zero()).IsTrue()
	assert.For(t).ThatActual(combined.Values()).Equals([]EnumKey{0, 1, 2})
}
