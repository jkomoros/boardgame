package boardgame

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/enum"
)

func TestComputedPropertiesAreDeclaredAndEvaluatedTogether(t *testing.T) {
	delegate := defaultTestGameDelegate(0)
	manager, err := NewGameManager(delegate, newTestStorageManager())
	if err != nil {
		t.Fatal(err)
	}

	descriptors := manager.ComputedPropertyDescriptors()
	if len(descriptors) != 2 {
		t.Fatalf("descriptors = %+v, want two", descriptors)
	}
	if descriptors[0].Name != "SumAllScores" || descriptors[0].Scope != ComputedPropertyScopeGlobal || descriptors[0].Type != TypeInt {
		t.Fatalf("first descriptor = %+v", descriptors[0])
	}
	if descriptors[1].Name != "EffectiveMovesLeftThisTurn" || descriptors[1].Scope != ComputedPropertyScopePlayer || descriptors[1].Type != TypeInt {
		t.Fatalf("second descriptor = %+v", descriptors[1])
	}

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	computed := game.CurrentState().(*state).computed()
	if _, ok := computed.Global["SumAllScores"].(int); !ok {
		t.Fatalf("global computed value = %#v, want int", computed.Global["SumAllScores"])
	}
	if _, ok := computed.Players[0]["EffectiveMovesLeftThisTurn"].(int); !ok {
		t.Fatalf("player computed value = %#v, want int", computed.Players[0]["EffectiveMovesLeftThisTurn"])
	}
}

func TestComputedPropertyConfigurationFailsLoudly(t *testing.T) {
	otherEnums := enum.NewSet()
	otherColor := otherEnums.MustAdd("color", map[enum.EnumKey]string{0: "Other"})
	tests := map[string][]ComputedProperty{
		"empty name":          {GlobalComputedBool("", func(ImmutableState) bool { return true })},
		"nil callback":        {GlobalComputedBool("Broken", nil)},
		"duplicate":           {GlobalComputedBool("Same", func(ImmutableState) bool { return true }), GlobalComputedInt("Same", func(ImmutableState) int { return 1 })},
		"framework collision": {PlayerComputedInt("Color", func(ImmutableSubState) int { return 1 })},
		"foreign enum": {GlobalComputedEnum("Mood", otherColor, func(ImmutableState) enum.ImmutableVal {
			return otherColor.NewDefaultVal()
		})},
		"foreign enum slice": {GlobalComputedEnumSlice("Moods", otherColor, func(ImmutableState) enum.ImmutableEnumSlice {
			return otherColor.NewEnumSlice()
		})},
	}
	for name, properties := range tests {
		t.Run(name, func(t *testing.T) {
			delegate := defaultTestGameDelegate(0)
			delegate.computedProperties = properties
			_, err := NewGameManager(delegate, newTestStorageManager())
			if err == nil || !strings.Contains(err.Error(), "computed") {
				t.Fatalf("NewGameManager() error = %v, want computed-property error", err)
			}
		})
	}
}

func TestComputedSlicesAreCopiedAndNeverNil(t *testing.T) {
	delegate := defaultTestGameDelegate(0)
	var source []int
	delegate.computedProperties = []ComputedProperty{
		GlobalComputedIntSlice("Numbers", func(ImmutableState) []int { return source }),
	}
	manager, err := NewGameManager(delegate, newTestStorageManager())
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	value := game.CurrentState().(*state).computed().Global["Numbers"].([]int)
	if value == nil || len(value) != 0 {
		t.Fatalf("computed slice = %#v, want non-nil empty slice", value)
	}
}
