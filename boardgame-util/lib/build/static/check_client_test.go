package static

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestClientCheckResultIsDeterministic(t *testing.T) {
	result := NewClientCheckResult([]ClientDiagnostic{
		{Source: "typescript", Code: "TS2", Severity: "error", Package: "z", File: "b.ts", Line: 2, Message: "second"},
		{Source: "boardgame", Code: "BGCLIENT001", Severity: "error", Package: "a", File: "a.ts", Line: 1, Message: "first", Remediation: "repair it"},
	})
	if result.OK || result.Version != 1 || result.Diagnostics[0].Package != "a" {
		t.Fatalf("unexpected normalized result: %#v", result)
	}
	var encoded bytes.Buffer
	if err := result.WriteJSON(&encoded); err != nil {
		t.Fatal(err)
	}
	var decoded ClientCheckResult
	if err := json.Unmarshal(encoded.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Diagnostics[0].Code != "BGCLIENT001" {
		t.Fatalf("unexpected JSON order: %s", encoded.String())
	}

	var human bytes.Buffer
	if err := result.WriteHuman(&human); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"a.ts:1: error BGCLIENT001: first", "Fix: repair it", "b.ts:2: error TS2: second"} {
		if !strings.Contains(human.String(), want) {
			t.Fatalf("human output missing %q:\n%s", want, human.String())
		}
	}
}

func TestEmptyClientCheckResultIsGreen(t *testing.T) {
	result := NewClientCheckResult(nil)
	if !result.OK || result.Diagnostics == nil {
		t.Fatalf("unexpected empty result: %#v", result)
	}
}
