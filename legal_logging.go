package boardgame

import "github.com/sirupsen/logrus"

/*
This file implements #65 (design spec §6, "Server": "The fixup loop logs
rejections at debug level ... no exceptions"): fixup-move rejections used to
be silently discarded (game.go's applyMove just turned the Legal() error
into the ProposeMove chain's returned error and moved on). Now the rejection
is also logged at debug, naming the move, the first-failing predicate (for a
move that opted into declarative legality — legal_plan.go), and the
rendered message; an opaque (non-opted-in) move has no plan to introspect,
so only its plain Legal() error string is logged.
*/

// logFixupRejection logs a rejected fix-up move at debug level (#65). Gated
// behind a log-level check up front so the (rare) cost of re-deriving the
// first-failing predicate name — a full-ledger plan re-evaluation, see
// below — is never paid outside of debug logging.
func (g *Game) logFixupRejection(move Move, state ImmutableState, proposer PlayerIndex, err error) {
	logger := g.manager.Logger()
	if logger == nil || logger.Level < logrus.DebugLevel {
		return
	}

	name := move.Info().Name()
	predicateName := legalFixupRejectionPredicateName(g.manager, name, state, move, proposer)

	logger.WithFields(logrus.Fields{
		"move":      name,
		"predicate": predicateName,
	}).Debugln("fixup rejected: " + err.Error())
}

// legalFixupRejectionPredicateName returns the name of the first predicate
// that failed (or was Unknown) in moveName's assembled plan, evaluated
// fresh in full-ledger mode so every predicate's individual verdict is
// available — the hot-path evaluate() used by move.Legal() itself only ever
// returns a single overall LegalVerdict, with no record of which predicate
// produced it. Returns "" for an opaque (non-opted-in) move, or if the
// re-evaluation is (surprisingly) all-Pass despite move.Legal() having just
// reported an error — a discrepancy the caller's "plain error" logging
// already covers regardless.
func legalFixupRejectionPredicateName(manager *GameManager, moveName string, state ImmutableState, move Move, proposer PlayerIndex) string {
	plan := manager.legalPlans[moveName]
	if plan == nil {
		return ""
	}

	_, entries := plan.evaluate(LegalContext{State: state, Move: move, Proposer: proposer, Chest: manager.chest}, true)
	for _, entry := range entries {
		if entry.Verdict.Outcome != LegalPass {
			return entry.Name
		}
	}
	return ""
}
