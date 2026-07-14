package legal

import "testing"

// TestPassVerdict verifies PassVerdict returns a Pass outcome with no
// Message.
func TestPassVerdict(t *testing.T) {
	v := PassVerdict()
	if v.Outcome != Pass {
		t.Fatalf("Outcome = %v, want Pass", v.Outcome)
	}
	if v.Message != nil {
		t.Fatalf("Message = %+v, want nil", v.Message)
	}
}

// TestFailT verifies FailT returns a Fail outcome carrying the given
// template and bindings, and that omitting bindings is allowed.
func TestFailT(t *testing.T) {
	v := FailT("reveal.no_cards_left", map[string]BindingValue{"left": Int(0)})
	if v.Outcome != Fail {
		t.Fatalf("Outcome = %v, want Fail", v.Outcome)
	}
	if v.Message == nil {
		t.Fatalf("Message = nil, want non-nil")
	}
	if v.Message.Template != "reveal.no_cards_left" {
		t.Fatalf("Message.Template = %q, want %q", v.Message.Template, "reveal.no_cards_left")
	}
	if got := v.Message.Bindings["left"]; got.I == nil || *got.I != 0 {
		t.Fatalf("Message.Bindings[\"left\"] = %+v, want I=0", got)
	}

	noBindings := FailT("some.key")
	if noBindings.Message == nil || noBindings.Message.Template != "some.key" {
		t.Fatalf("FailT with no bindings = %+v", noBindings)
	}
	if noBindings.Message.Bindings != nil {
		t.Fatalf("Message.Bindings = %+v, want nil", noBindings.Message.Bindings)
	}
}

// TestUnknownVerdict verifies UnknownVerdict returns an Unknown outcome
// with the given reason.
func TestUnknownVerdict(t *testing.T) {
	v := UnknownVerdict("reads hidden property HiddenCards")
	if v.Outcome != Unknown {
		t.Fatalf("Outcome = %v, want Unknown", v.Outcome)
	}
	if v.Reason != "reads hidden property HiddenCards" {
		t.Fatalf("Reason = %q, want %q", v.Reason, "reads hidden property HiddenCards")
	}
}

// TestBindingValueHelpers verifies String, Int, and Bool each set exactly
// the corresponding field.
func TestBindingValueHelpers(t *testing.T) {
	s := String("foo")
	if s.S == nil || *s.S != "foo" || s.I != nil || s.B != nil {
		t.Fatalf("String(\"foo\") = %+v", s)
	}

	i := Int(42)
	if i.I == nil || *i.I != 42 || i.S != nil || i.B != nil {
		t.Fatalf("Int(42) = %+v", i)
	}

	b := Bool(true)
	if b.B == nil || *b.B != true || b.S != nil || b.I != nil {
		t.Fatalf("Bool(true) = %+v", b)
	}
}
