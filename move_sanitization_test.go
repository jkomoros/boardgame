package boardgame

import (
	"reflect"
	"strings"
	"testing"
)

func TestConfiguredMoveNameSanitization(t *testing.T) {
	config := make(PropertyCollection)
	SetMoveNameSanitization(config, "self:visible,same-team:visible", "Hidden Action")

	policies, hiddenKey, err := configuredMoveNameSanitization(config)
	if err != nil {
		t.Fatal(err)
	}
	if hiddenKey != "Hidden Action" {
		t.Fatalf("hidden key = %q, want Hidden Action", hiddenKey)
	}
	if policies[sanitizationGroupSelf] != PolicyVisible || policies["same-team"] != PolicyVisible {
		t.Fatalf("policies = %#v, want self and same-team visible", policies)
	}
	if got := ResolveSanitizationPolicy(policies, map[string]bool{SanitizationDefaultGroup: true}, PolicyHidden); got != PolicyHidden {
		t.Fatalf("unmatched policy = %v, want hidden", got)
	}
}

func TestConfiguredMoveNameSanitizationRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		set  func(PropertyCollection)
	}{
		{
			name: "structural policy",
			set: func(config PropertyCollection) {
				SetMoveNameSanitization(config, "self:len")
			},
		},
		{
			name: "multiple hidden keys",
			set: func(config PropertyCollection) {
				SetMoveNameSanitization(config, "self:visible", "Hidden One", "Hidden Two")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := make(PropertyCollection)
			test.set(config)
			if _, _, err := configuredMoveNameSanitization(config); err == nil {
				t.Fatal("expected invalid move name sanitization to fail")
			}
		})
	}
}

func TestMoveAnimationKeys(t *testing.T) {
	for _, hidden := range []string{"", "Play Card", "Hidden Action"} {
		info := &MoveInfo{name: "Play Card", runtime: moveRuntime{moveType: &moveType{hiddenAnimationKey: hidden}}}
		want := []string{"Play Card"}
		if hidden == "Hidden Action" {
			want = append(want, hidden)
		}
		got := info.AnimationKeys()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		got[0] = "mutated"
		if info.AnimationKeys()[0] != "Play Card" {
			t.Fatal("keys expose mutable metadata")
		}
	}
	if (&MoveInfo{}).AnimationKeys() != nil || (*MoveInfo)(nil).AnimationKeys() != nil {
		t.Fatal("uninitialized info has keys")
	}
}

func TestAnimationKeyWireBoundary(t *testing.T) {
	for _, key := range []string{" \t\n", "\ufeff", strings.Repeat("a", 257), strings.Repeat("🎲", 129), string([]byte{0xff})} {
		if err := validateAnimationKey(key); err == nil {
			t.Fatalf("accepted invalid key %q", key)
		}
		config := make(PropertyCollection)
		SetMoveNameSanitization(config, "hidden", key)
		if _, _, err := configuredMoveNameSanitization(config); err == nil {
			t.Fatalf("accepted invalid alias %q", key)
		}
		if _, err := newMoveType(defaultMoveConfig{name: key}, nil); err == nil || !strings.Contains(err.Error(), "animation key") {
			t.Fatalf("move name wasn't rejected at construction: %v", err)
		}
	}
	for _, key := range []string{strings.Repeat("a", 256), strings.Repeat("中", 256), strings.Repeat("🎲", 128)} {
		if err := validateAnimationKey(key); err != nil {
			t.Fatalf("rejected valid key: %v", err)
		}
	}
}
