package legal_test

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

func TestPlayerBoolAtSelectors(t *testing.T) {
	proposerSpec := legal.PlayerBoolAt(legal.Proposer(), "SeatFilled", true)
	if proposerSpec.Name != "playerBoolAt" || len(proposerSpec.Args) != 2 || proposerSpec.Args[0] != "proposer.SeatFilled" || proposerSpec.Args[1] != "true" {
		t.Fatalf("proposer spec = %+v", proposerSpec)
	}
	currentSpec := legal.PlayerBoolAt(legal.CurrentPlayer(), "SeatFilled", false)
	if currentSpec.Args[0] != "player.SeatFilled" {
		t.Fatalf("current-player path = %q", currentSpec.Args[0])
	}
	moveSpec := legal.PlayerBoolAt(legal.PlayerFromMove("TargetPlayerIndex"), "SeatFilled", true)
	if moveSpec.Args[0] != "players[move.TargetPlayerIndex].SeatFilled" {
		t.Fatalf("move-player path = %q", moveSpec.Args[0])
	}

	pred := resolvePredicateForTest(t, proposerSpec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "proposer.SeatFilled" {
		t.Fatalf("reads = %+v", pred.Reads)
	}
	if pred.RequiredReadTypes["proposer.SeatFilled"] != boardgame.TypeBool {
		t.Fatalf("required read types = %+v", pred.RequiredReadTypes)
	}
	filled := buildLegalFixture(t, "memorySeatFilled")
	if got := pred.Evaluate(filled.context(0)); got.Outcome != legal.Pass {
		t.Fatalf("filled proposer outcome = %v (%+v)", got.Outcome, got)
	}
	empty := buildLegalFixture(t, "memoryDefault")
	if got := pred.Evaluate(empty.context(0)); got.Outcome != legal.Fail {
		t.Fatalf("empty proposer outcome = %v (%+v)", got.Outcome, got)
	}
}

func TestBehaviorPredicateBuilders(t *testing.T) {
	tests := []struct {
		name     string
		spec     legal.Spec
		path     string
		want     string
		template string
	}{
		{"submitted", legal.PlayerHasSubmitted(legal.Proposer()), "proposer.PlayerSubmitted", "true", legal.TemplatePlayerNotSubmitted},
		{"not submitted", legal.PlayerHasNotSubmitted(legal.Proposer()), "proposer.PlayerSubmitted", "false", legal.TemplatePlayerAlreadySubmitted},
		{"active", legal.PlayerIsActive(legal.Proposer()), "proposer.PlayerInactive", "false", legal.TemplatePlayerInactive},
		{"inactive", legal.PlayerIsInactive(legal.Proposer()), "proposer.PlayerInactive", "true", legal.TemplatePlayerActive},
		{"seat filled", legal.PlayerSeatIsFilled(legal.Proposer()), "proposer.SeatFilled", "true", legal.TemplateSeatNotFilled},
		{"seat closed", legal.PlayerSeatIsClosed(legal.Proposer()), "proposer.SeatClosed", "true", legal.TemplateSeatNotClosed},
		{"admin", legal.PlayerIsAdmin(legal.Proposer()), "proposer.IsAdmin", "true", legal.TemplatePlayerNotAdmin},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.spec.Name != "playerBoolAt" || len(tc.spec.Args) != 2 || tc.spec.Args[0] != tc.path || tc.spec.Args[1] != tc.want || tc.spec.Message != tc.template {
				t.Fatalf("spec = %+v", tc.spec)
			}
			if tc.spec.AdminPolicy != boardgame.LegalAdminBypass {
				t.Fatalf("admin policy = %q, want bypass", tc.spec.AdminPolicy)
			}
		})
	}

	current := legal.PlayerIsActive(legal.CurrentPlayer())
	if current.AdminPolicy != boardgame.LegalAdminEvaluate {
		t.Fatalf("current-player helper unexpectedly bypasses admin: %+v", current)
	}
}
