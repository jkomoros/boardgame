package boardgame

/*
This file implements the design spec §6 "evaluable" formula's per-Read half:
"every Read's Facet survives this viewer's sanitization" — the piece Task 10's
server ledger needs to decide, per PreconditionEntry per viewer, whether it
may ship Message.Bindings to the client (the #693 guard). The Serializable
half of the formula lives on LegalPredicate itself (legal_predicate.go); the
caller (server/api) combines both: evaluable = entry.Serializable && every
entry.Read is LegalReadEvaluable.

This is deliberately core-side, not server-side: it needs facetSurvives
(legal_path.go, unexported) and the sanitizationTransformation machinery
(sanitization.go, unexported, built by the unexported *state concrete type),
neither of which package server/api can reach.
*/

// LegalReadEvaluable reports whether read's declared Facet would survive
// st's sanitization for viewer — per the design spec §6 table
// (facetSurvives), consulted against the SAME per-property policy
// st.SanitizedForPlayer(viewer) would apply, without actually sanitizing
// st (the server ledger needs the real, unsanitized state to evaluate
// predicates against; it only needs to know what a sanitized COPY would
// have hidden).
//
//   - "move.X" reads are always evaluable: a move's field values are
//     supplied by whoever is framing the request (the proposer themselves,
//     or an admin building the ledger), never sanitized state.
//   - AdminPlayerIndex is omniscient, mirroring state.SanitizedForPlayer's
//     own AdminPlayerIndex bypass (see its doc comment): every Read is
//     evaluable to an Admin viewer.
//   - "player.X" reads resolve against st's actual current player — the
//     same player resolveLegalPath itself reads against — so evaluability
//     is judged against THAT player's sanitization policy for viewer.
//   - "players[*].X" reads (quantifier-only, e.g. AllActivePlayers' inner
//     leaf) require the facet to survive for EVERY player: a per-player
//     quantifier's verdict is a function of every iterated player's value,
//     so if even one player's policy would hide it from viewer, the
//     aggregate verdict can leak information about that one player.
//
// A malformed read.Path, a nil st, or an st whose concrete type isn't this
// package's own (impossible in practice — ImmutableState has exactly one
// implementation) all fail closed (false): this is the same "never leak
// bindings we can't prove are safe" posture as the #693 guard itself.
func LegalReadEvaluable(st ImmutableState, viewer PlayerIndex, read LegalRead) bool {
	parsed, err := parseLegalPath(read.Path)
	if err != nil {
		return false
	}

	if parsed.kind == pathMove {
		return true
	}

	if viewer == AdminPlayerIndex {
		return true
	}

	if st == nil {
		return false
	}

	concrete, ok := st.(*state)
	if !ok {
		return false
	}

	transformation, err := concrete.generateSanitizationTransformation(viewer)
	if err != nil {
		return false
	}

	switch parsed.kind {
	case pathGame:
		return facetSurvives(legalPolicyForProp(transformation.Game, parsed.prop), read.Facet)
	case pathPlayer:
		current := st.CurrentPlayerIndex()
		if current < 0 || int(current) >= len(transformation.Players) {
			return false
		}
		return facetSurvives(legalPolicyForProp(transformation.Players[current], parsed.prop), read.Facet)
	case pathPlayersAll:
		if len(transformation.Players) == 0 {
			return false
		}
		for _, playerPolicies := range transformation.Players {
			if !facetSurvives(legalPolicyForProp(playerPolicies, parsed.prop), read.Facet) {
				return false
			}
		}
		return true
	}

	return false
}

// LegalReadsEvaluable reports whether EVERY read in reads is
// LegalReadEvaluable for (st, viewer) — the conjunction the design spec §6
// formula requires ("every Read's Facet survives"), and (per the same
// spec's "Kleene-honest" note) exactly what makes an "any" compositor's
// single unioned Reads set evaluable iff every child predicate's reads are:
// no special-casing of "any" is needed here because LegalPredicate.Reads is
// already the union for a compositor (see resolveLegalAnySpec).
func LegalReadsEvaluable(st ImmutableState, viewer PlayerIndex, reads []LegalRead) bool {
	for _, read := range reads {
		if !LegalReadEvaluable(st, viewer, read) {
			return false
		}
	}
	return true
}

// legalPolicyForProp looks up propName's Policy in policies, defaulting to
// PolicyVisible for a missing entry — subStateSanitizationTransformation's
// own documented convention ("Missing properties will be treated as
// PolicyVisible").
func legalPolicyForProp(policies subStateSanitizationTransformation, propName string) Policy {
	if policy, ok := policies[propName]; ok {
		return policy
	}
	return PolicyVisible
}
