package boardgame

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestLegalOutcomeZeroValueFailsClosed verifies that a zero-value
// LegalVerdict has an Outcome that is none of the valid outcomes, so a
// forgotten/zero-initialized verdict never silently reads as legal.
func TestLegalOutcomeZeroValueFailsClosed(t *testing.T) {
	var v LegalVerdict
	if v.Outcome == LegalPass || v.Outcome == LegalFail || v.Outcome == LegalUnknown {
		t.Fatalf("zero-value LegalVerdict.Outcome must not equal any valid outcome, got %v", v.Outcome)
	}
	if v.Outcome != legalOutcomeInvalid {
		t.Fatalf("zero-value LegalVerdict.Outcome = %v, want legalOutcomeInvalid", v.Outcome)
	}
}

// TestLegalSpecJSONRoundTripLeaf verifies that a leaf LegalSpec marshals to
// the expected compact JSON (empty fields omitted) and round-trips exactly.
func TestLegalSpecJSONRoundTripLeaf(t *testing.T) {
	spec := LegalSpec{
		Name:    "playerPropAtLeast",
		Args:    []string{"player.CardsLeftToReveal", "1"},
		Message: "reveal.no_cards_left",
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	want := `{"name":"playerPropAtLeast","args":["player.CardsLeftToReveal","1"],"message":"reveal.no_cards_left"}`
	if string(data) != want {
		t.Fatalf("Marshal = %s, want %s", data, want)
	}

	var got LegalSpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if !reflect.DeepEqual(got, spec) {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, spec)
	}
}

// TestLegalSpecJSONRoundTripNested verifies that a compositor LegalSpec
// ("any" with two subs) marshals to the expected JSON and round-trips
// exactly, with empty fields (Args/Message on the parent, and on each sub)
// omitted.
func TestLegalSpecJSONRoundTripNested(t *testing.T) {
	spec := LegalSpec{
		Name: "any",
		Sub: []LegalSpec{
			{Name: "playerBool", Args: []string{"Eliminated"}},
			{Name: "playerBool", Args: []string{"Stood"}},
		},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	want := `{"name":"any","sub":[{"name":"playerBool","args":["Eliminated"]},{"name":"playerBool","args":["Stood"]}]}`
	if string(data) != want {
		t.Fatalf("Marshal = %s, want %s", data, want)
	}

	var got LegalSpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if !reflect.DeepEqual(got, spec) {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, spec)
	}
}

// TestLegalSpecWithMessageReturnsCopy verifies that WithMessage does not
// mutate the receiver and returns a modified copy.
func TestLegalSpecWithMessageReturnsCopy(t *testing.T) {
	original := LegalSpec{Name: "foo", Args: []string{"bar"}}
	originalCopy := original

	modified := original.WithMessage("some.key")

	if !reflect.DeepEqual(original, originalCopy) {
		t.Fatalf("original was mutated: got %+v, want %+v", original, originalCopy)
	}
	if original.Message != "" {
		t.Fatalf("original.Message = %q, want empty", original.Message)
	}
	if modified.Message != "some.key" {
		t.Fatalf("modified.Message = %q, want %q", modified.Message, "some.key")
	}
	if modified.Name != original.Name {
		t.Fatalf("modified.Name = %q, want %q", modified.Name, original.Name)
	}
}

func TestLegalSpecWithTemplateKeyReturnsCopy(t *testing.T) {
	original := LegalSpec{Name: "foo"}
	modified := original.WithTemplateKey("some.key")
	if original.Message != "" || modified.Message != "some.key" {
		t.Fatalf("WithTemplateKey mutated original or failed to set key: original=%+v modified=%+v", original, modified)
	}
}

// TestLegalBindingValueMarshalJSON verifies that LegalBindingValue marshals
// as the bare JSON value of whichever single field is set.
func TestLegalBindingValueMarshalJSON(t *testing.T) {
	s := "foo"
	i := 42
	b := true

	tests := []struct {
		name string
		val  LegalBindingValue
		want string
	}{
		{"string", LegalBindingValue{S: &s}, `"foo"`},
		{"int", LegalBindingValue{I: &i}, `42`},
		{"bool", LegalBindingValue{B: &b}, `true`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.val)
			if err != nil {
				t.Fatalf("Marshal returned error: %v", err)
			}
			if string(data) != tc.want {
				t.Fatalf("Marshal = %s, want %s", data, tc.want)
			}
		})
	}
}

func TestLegalBindingValueMarshalJSONRejectsMultipleFields(t *testing.T) {
	s := "foo"
	i := 42
	if _, err := json.Marshal(LegalBindingValue{S: &s, I: &i}); err == nil {
		t.Fatal("Marshal accepted a LegalBindingValue with multiple fields set")
	}
}

// TestLegalBindingValueMarshalJSONZeroValueErrors verifies that marshaling
// a zero-value LegalBindingValue (no field set) is an error, since that
// violates the exactly-one-field-set invariant.
func TestLegalBindingValueMarshalJSONZeroValueErrors(t *testing.T) {
	var v LegalBindingValue
	if _, err := json.Marshal(v); err == nil {
		t.Fatalf("Marshal of zero-value LegalBindingValue returned nil error, want error")
	}
}

// TestLegalBindingValueUnmarshalJSON verifies that UnmarshalJSON infers the
// correct field from the JSON value's type.
func TestLegalBindingValueUnmarshalJSON(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var v LegalBindingValue
		if err := json.Unmarshal([]byte(`"foo"`), &v); err != nil {
			t.Fatalf("Unmarshal returned error: %v", err)
		}
		if v.S == nil || *v.S != "foo" {
			t.Fatalf("got %+v, want S=foo", v)
		}
		if v.I != nil || v.B != nil {
			t.Fatalf("got %+v, want only S set", v)
		}
	})

	t.Run("int", func(t *testing.T) {
		var v LegalBindingValue
		if err := json.Unmarshal([]byte(`42`), &v); err != nil {
			t.Fatalf("Unmarshal returned error: %v", err)
		}
		if v.I == nil || *v.I != 42 {
			t.Fatalf("got %+v, want I=42", v)
		}
		if v.S != nil || v.B != nil {
			t.Fatalf("got %+v, want only I set", v)
		}
	})

	t.Run("bool", func(t *testing.T) {
		var v LegalBindingValue
		if err := json.Unmarshal([]byte(`true`), &v); err != nil {
			t.Fatalf("Unmarshal returned error: %v", err)
		}
		if v.B == nil || *v.B != true {
			t.Fatalf("got %+v, want B=true", v)
		}
		if v.S != nil || v.I != nil {
			t.Fatalf("got %+v, want only B set", v)
		}
	})

	t.Run("non-integer JSON number errors", func(t *testing.T) {
		var v LegalBindingValue
		if err := json.Unmarshal([]byte(`1.5`), &v); err == nil {
			t.Fatalf("Unmarshal of 1.5 returned nil error, want error")
		}
		if v.I != nil || v.S != nil || v.B != nil {
			t.Fatalf("got %+v, want no fields set after error", v)
		}
	})

	t.Run("integral float succeeds", func(t *testing.T) {
		var v LegalBindingValue
		if err := json.Unmarshal([]byte(`1e2`), &v); err != nil {
			t.Fatalf("Unmarshal of 1e2 returned error: %v", err)
		}
		if v.I == nil || *v.I != 100 {
			t.Fatalf("got %+v, want I=100", v)
		}
		if v.S != nil || v.B != nil {
			t.Fatalf("got %+v, want only I set", v)
		}
	})

	t.Run("unmarshal bool into value with int clears int", func(t *testing.T) {
		i := 42
		v := LegalBindingValue{I: &i}
		if err := json.Unmarshal([]byte(`true`), &v); err != nil {
			t.Fatalf("Unmarshal returned error: %v", err)
		}
		if v.B == nil || *v.B != true {
			t.Fatalf("got %+v, want B=true", v)
		}
		if v.I != nil {
			t.Fatalf("got I=%v, want I=nil", v.I)
		}
		if v.S != nil {
			t.Fatalf("got %+v, want only B set", v)
		}
	})
}

// TestLegalMessageJSONBindings verifies that a LegalMessage's Bindings map
// (whose values are LegalBindingValue) round-trips through JSON correctly,
// exercising LegalBindingValue's custom (un)marshaling as a map value.
func TestLegalMessageJSONBindings(t *testing.T) {
	left := 0
	msg := LegalMessage{
		Template: "cleanup.players_unfinished",
		Bindings: map[string]LegalBindingValue{
			"left": {I: &left},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	want := `{"Template":"cleanup.players_unfinished","Bindings":{"left":0}}`
	if string(data) != want {
		t.Fatalf("Marshal = %s, want %s", data, want)
	}

	var got LegalMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if !reflect.DeepEqual(got, msg) {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, msg)
	}
}
