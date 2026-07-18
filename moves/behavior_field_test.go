package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/behaviors"
)

type promotedBehaviorHolder struct {
	Market behaviors.FaceUpMarket
}

type promotedBehaviorState struct {
	*promotedBehaviorHolder
}

func TestNamedBehaviorValueRejectsUnsafeOrAmbiguousFields(t *testing.T) {
	tests := []struct {
		name      string
		state     any
		fieldName string
		want      string
	}{
		{"nil game state", (*struct{ Market behaviors.FaceUpMarket })(nil), "Market", "non-nil"},
		{"empty name", &struct{ Market behaviors.FaceUpMarket }{}, " ", "empty"},
		{"missing", &struct{ Market behaviors.FaceUpMarket }{}, "Other", "no field"},
		{"wrong type", &struct{ Market int }{}, "Market", "want direct value"},
		{"pointer", &struct{ Market *behaviors.FaceUpMarket }{}, "Market", "want direct value"},
		{"unexported", &struct{ market behaviors.FaceUpMarket }{}, "market", "not accessible"},
		{"promoted through nil pointer", &promotedBehaviorState{}, "Market", "promoted"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := namedBehaviorValue(test.state, test.fieldName, new(behaviors.FaceUpMarket))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestNamedBehaviorValueReturnsDirectValueAddress(t *testing.T) {
	state := &struct{ Market behaviors.FaceUpMarket }{}
	value, err := namedBehaviorValue(state, "Market", new(behaviors.FaceUpMarket))
	if err != nil {
		t.Fatalf("namedBehaviorValue: %v", err)
	}
	if value != &state.Market {
		t.Fatalf("value = %p, want %p", value, &state.Market)
	}
}
