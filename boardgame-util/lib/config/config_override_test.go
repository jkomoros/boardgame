package config

import "testing"

func TestOverrideAllowedOriginsIsTemporaryAndAppliesToBothModes(t *testing.T) {
	c := DefaultStarterConfig("")
	beforeDev := c.Dev.AllowedOrigins
	beforeProd := c.Prod.AllowedOrigins
	const origins = "http://localhost:49152,http://127.0.0.1:49152"

	c.AddOverride(OverrideAllowedOrigins(origins))

	if c.Dev.AllowedOrigins != origins {
		t.Errorf("dev AllowedOrigins = %q, want %q", c.Dev.AllowedOrigins, origins)
	}
	if c.Prod.AllowedOrigins != origins {
		t.Errorf("prod AllowedOrigins = %q, want %q", c.Prod.AllowedOrigins, origins)
	}
	if c.rawPublicConfig == nil {
		t.Fatal("starter config unexpectedly lacks a raw public config")
	}
	c.overriders = nil
	c.derive()
	if c.Dev.AllowedOrigins != beforeDev || c.Prod.AllowedOrigins != beforeProd {
		t.Fatal("allowed-origins override mutated persisted source configuration")
	}
}

func TestOriginAllowedTrimsCommaDelimitedEntries(t *testing.T) {
	mode := &Mode{ModeCommon: ModeCommon{AllowedOrigins: "http://localhost:49152, http://127.0.0.1:49152"}}
	for _, origin := range []string{
		"http://localhost:49152",
		"http://127.0.0.1:49152",
	} {
		if !mode.OriginAllowed(origin) {
			t.Errorf("OriginAllowed(%q) = false, want true", origin)
		}
	}
}
