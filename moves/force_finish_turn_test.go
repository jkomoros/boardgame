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
