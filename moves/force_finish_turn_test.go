package moves

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
)

func TestForceFinishTurnLegalRequiresAdmin(t *testing.T) {
	m := &ForceFinishTurn{}

	// AdminPlayerIndex should be accepted at the Legal check itself.
	// (Note: a full Legal call also reads CurrentPlayerIndex from state,
	// which requires a fixture state; the admin-gate check is what we're
	// covering here. We construct a minimal state below for the second
	// test.)
	err := m.Legal(nil, boardgame.PlayerIndex(0))
	assert.For(t).ThatActual(err).IsNotNil()

	err = m.Legal(nil, boardgame.ObserverPlayerIndex)
	assert.For(t).ThatActual(err).IsNotNil()

	// Hitting AdminPlayerIndex with nil state will panic before the
	// admin-check returns nil — we can't easily run the success path
	// without a fixture state. The non-admin rejection is the meaningful
	// safety check and is what this test pins.
}

func TestForceFinishTurnEmbedsFinishTurn(t *testing.T) {
	// ForceFinishTurn embeds FinishTurn so that Apply (which calls
	// ResetForTurnEnd + CurrentPlayerSetter) is inherited. This test
	// pins that the embedding compiles and that ForceFinishTurn can be
	// treated as a FinishTurn — protecting against accidental future
	// refactors that would un-embed.
	var ff ForceFinishTurn
	var ft *FinishTurn = &ff.FinishTurn
	_ = ft
	// If this compiles, the embedding is intact.
	assert.For(t).ThatActual(true).IsTrue()
}

// TestForceFinishTurnRefusesToBootMisconfigured pins that the two options
// ForceFinishTurn's own doc comment calls "load-bearing" are enforced at boot
// rather than merely described in prose.
//
// Without WithIsFixUp(false) the move inherits FinishTurn's FixUp-ness, so the
// framework's fixup pipeline auto-proposes it under AdminPlayerIndex -- the
// exact identity its Legal() accepts -- and it advances the current player on
// every fixup pass, forever. Without WithMoveName the auto-configurer derives
// "Finish Turn" from the embedded parent, silently shadowing or colliding with
// the real FinishTurn registration. Both used to boot without complaint.
func TestForceFinishTurnRefusesToBootMisconfigured(t *testing.T) {

	tests := []struct {
		description string
		options     []CustomConfigurationOption
		wantErr     bool
	}{
		{
			description: "neither option",
			options:     nil,
			wantErr:     true,
		},
		{
			description: "name only, still a fixup",
			options:     []CustomConfigurationOption{WithMoveName("Force Finish Turn")},
			wantErr:     true,
		},
		{
			description: "not a fixup, but no name override",
			options:     []CustomConfigurationOption{WithIsFixUp(false)},
			wantErr:     true,
		},
		{
			description: "both options, as the doc comment requires",
			options: []CustomConfigurationOption{
				WithMoveName("Force Finish Turn"),
				WithIsFixUp(false),
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		options := test.options
		moveInstaller := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
			auto := NewAutoConfigurer(manager.Delegate())
			return []boardgame.MoveConfig{
				auto.MustConfig(new(ForceFinishTurn), options...),
			}
		}

		_, err := newGameManager(moveInstaller)

		if test.wantErr && err == nil {
			t.Errorf("%s: expected NewGameManager to refuse the move, but it booted", test.description)
		}
		if !test.wantErr && err != nil {
			t.Errorf("%s: expected NewGameManager to succeed, got %v", test.description, err)
		}
	}
}
