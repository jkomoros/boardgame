package movenames

import (
	"strings"
	"testing"
)

func TestAnimationKeysRemainSeparateFromProposals(t *testing.T) {
	got := GenerateTypeScript([]string{"Play Card"}, []string{"Play Card", "Retire Sequence", "Hidden Action", "Hidden Action"})
	parts := strings.Split(got, "export const AnimationKeys")
	if len(parts) != 2 {
		t.Fatal(got)
	}
	if strings.Contains(parts[0], "Retire Sequence") || strings.Contains(parts[0], "Hidden Action") {
		t.Fatal("transition-only keys became proposals")
	}
	for _, key := range []string{"PlayCard", "RetireSequence", "HiddenAction"} {
		if !strings.Contains(parts[1], key+":") {
			t.Fatalf("missing %s: %s", key, got)
		}
	}
	if strings.Count(parts[1], "HiddenAction:") != 1 {
		t.Fatal("shared aliases must be deduplicated")
	}
}

func TestEmptyAnimationVocabulary(t *testing.T) {
	got := GenerateTypeScript(nil, nil)
	if !strings.Contains(got, "export const MoveNames = {} as const;") || !strings.Contains(got, "export const AnimationKeys = {} as const;") {
		t.Fatal(got)
	}
}

func TestAnimationAliasesPreserveCollidingAndNumericNames(t *testing.T) {
	got := GenerateTypeScript(nil, []string{"Hidden-Action", "HiddenAction", "123", "!!!"})
	for _, key := range []string{`"Hidden-Action": "Hidden-Action"`, `"HiddenAction": "HiddenAction"`, `"123": "123"`, `"": "!!!"`} {
		if !strings.Contains(got, key) {
			t.Fatalf("missing %s in %s", key, got)
		}
	}
}
