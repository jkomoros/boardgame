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
//   - "player.X" reads resolve against the move's player. This function
//     sees only the read, not the move, so it judges the read against st's
//     current player when that is concrete. That is right for moves without
//     a player of their own, and for moves.CurrentPlayer moves, whose turn
//     check ties their player to the current one. When the current player
//     is a wildcard (AnyPlayerIndex during simultaneous play, or
//     AdminPlayerIndex) the read is judged like "proposer.X" below. Ledger
//     consumers should call LegalVerdictEntry.ReadsEvaluable instead, which
//     also judges a moves.AnyPlayer-style move's reads like "proposer.X".
//   - "proposer.X" reads conservatively require the facet to survive for
//     every player. LegalReadEvaluable intentionally has no proposer argument;
//     this avoids assuming that proposer and viewer are the same identity.
//   - "players[*].X" reads (quantifier-only, e.g. AllActivePlayers' inner
//     leaf) require the facet to survive for EVERY player: a per-player
//     quantifier's verdict is a function of every iterated player's value,
//     so if even one player's policy would hide it from viewer, the
//     aggregate verdict can leak information about that one player.
//   - "players[move.<Field>].<Prop>" reads require the facet to survive for
//     EVERY player, exactly like "players[*].X": which player <Field>
//     resolves to is only known once the move is fixed, and the server
//     ledger's job here is to judge evaluability WITHOUT trusting the move
//     to pick a favorable player, so it must be conservative over every
//     player the field could name (same rationale as pathPlayersAll above).
//   - Component paths ("game.Stack[selector].Prop") need the stack hop to
//     be PolicyVisible for every owner the path can name, and, when Prop is
//     a DynamicComponentValues property, read.Facet must also survive the
//     deck's dynamic-value policy (legalComponentReadEvaluable).
//
// A malformed read.Path, a nil st, or an st whose concrete type isn't this
// package's own (impossible in practice — ImmutableState has exactly one
// implementation) all fail closed (false): this is the same "never leak
// bindings we can't prove are safe" posture as the #693 guard itself.
func LegalReadEvaluable(st ImmutableState, viewer PlayerIndex, read LegalRead) bool {
	return legalReadEvaluable(st, viewer, read, false)
}

// ReadsEvaluable reports whether every one of e's Reads is evaluable for
// (st, viewer). It is LegalReadsEvaluable, except that when e's "player.X"
// reads name the proposer rather than the current player
// (e.MovePlayerIsProposer: a move made for its own player that nothing ties
// to the current player, such as a moves.AnyPlayer move), those reads are
// judged like "proposer.X": conservatively, across every player.
func (e LegalVerdictEntry) ReadsEvaluable(st ImmutableState, viewer PlayerIndex) bool {
	for _, read := range e.Reads {
		if !legalReadEvaluable(st, viewer, read, e.MovePlayerIsProposer) {
			return false
		}
	}
	return true
}

func legalReadEvaluable(st ImmutableState, viewer PlayerIndex, read LegalRead, movePlayerIsProposer bool) bool {
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

	if parsed.hasComponent() {
		return legalComponentReadEvaluable(st, transformation, parsed, read, movePlayerIsProposer)
	}

	switch parsed.kind {
	case pathGame:
		return facetSurvives(legalPolicyForProp(transformation.Game, parsed.prop), read.Facet)
	case pathPlayer:
		current := st.CurrentPlayerIndex()
		if movePlayerIsProposer || legalPlayerIndexIsWildcard(current) {
			// Simultaneous play: the move's player is the proposer (or whom
			// an admin acts for), which this viewer-keyed API cannot trust.
			// Judge it like proposer.X: conservatively, across every player.
			return legalFacetSurvivesForEveryPlayer(transformation.Players, parsed.prop, read.Facet)
		}
		if current < 0 || int(current) >= len(transformation.Players) {
			return false
		}
		return facetSurvives(legalPolicyForProp(transformation.Players[current], parsed.prop), read.Facet)
	case pathProposer, pathPlayersAll, pathPlayersMoveField:
		return legalFacetSurvivesForEveryPlayer(transformation.Players, parsed.prop, read.Facet)
	default:
		// Every legalPathKind must be classified explicitly above. A new
		// path kind added to legal_path.go without a corresponding case here
		// would otherwise silently fall through to fail-closed false for
		// every read of that kind — which is exactly the bug this default
		// exists to make loud instead of silent: if you're here because a
		// new kind hit this default, go add its case above, don't just
		// accept the false.
		return false
	}
}

// legalFacetSurvivesForEveryPlayer reports whether facet of prop survives
// every player's sanitization policy for the viewer (and false when there
// are no players at all).
func legalFacetSurvivesForEveryPlayer(players []subStateSanitizationTransformation, prop string, facet LegalFacet) bool {
	if len(players) == 0 {
		return false
	}
	for _, playerPolicies := range players {
		if !facetSurvives(legalPolicyForProp(playerPolicies, prop), facet) {
			return false
		}
	}
	return true
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
