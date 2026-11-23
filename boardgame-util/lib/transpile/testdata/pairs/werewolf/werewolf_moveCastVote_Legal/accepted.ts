// Stored pair: werewolf.(*moveCastVote).Legal (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, moves["Cast Vote"].legal.
// One Go method serves moves["Cast Vote"].legal and moves["Cast Night
// Vote"].legal; the hand port builds each from the same function expression,
// copied once.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import { anyPlayerLegal } from '../../../../../../../server/static/src/engine/rules/stock.ts';
import type { LegalResult, MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const PHASE_DAY = 1;

const PHASE_NIGHT = 2;

const WEREWOLF = 'Werewolf';

const isWerewolf = (state: RulesState, player: number): boolean => state.players[player].Role === WEREWOLF;

const eliminated = (state: RulesState, player: number): boolean => state.players[player].Eliminated === true;

function voteForPhase(state: RulesState, player: number, phase: number): number {
	return (phase === PHASE_NIGHT ? state.players[player].NightVote : state.players[player].DayVote) as number;
}

export function werewolf_moveCastVote_Legal(state: RulesState, move: MoveInstance, proposer: number, base: () => string | null): string | null {
	const err = anyPlayerLegal(state, move, proposer, base);
	if (err !== null) return err;
	const voter = move.fields.TargetPlayerIndex as number;
	if (eliminated(state, voter)) return 'eliminated players cannot vote';
	const phase = state.currentPhaseKey();
	if (voteForPhase(state, voter, phase) >= 0) return 'you have already voted this phase';
	if (phase === PHASE_NIGHT) {
		if (!isWerewolf(state, voter)) return 'only werewolves may act during the night';
	} else if (phase !== PHASE_DAY) {
		return 'voting is not allowed during this phase';
	}
	const target = move.fields.VoteTarget as number;
	if (target < 0 || target >= state.numPlayers()) return 'invalid vote target';
	if (eliminated(state, target)) return 'cannot vote for an eliminated player';
	if (state.playerIsInactive(target)) return 'cannot vote for an inactive player';
	if (target === voter) return 'you cannot vote for yourself';
	return null;
}
