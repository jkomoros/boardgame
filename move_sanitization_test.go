package boardgame

import "testing"

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
