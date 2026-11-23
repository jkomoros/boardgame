// Stored pair: werewolf.(*moveResolveVotes).Legal (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, moves["Resolve Votes"].legal.
// One Go method serves moves["Resolve Votes"].legal and moves["Resolve Night
// Votes"].legal; the hand port builds each from the same function expression,
// copied once.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const PHASE_DAY = 1;

const PHASE_NIGHT = 2;

const WEREWOLF = 'Werewolf';

const isWerewolf = (state: RulesState, player: number): boolean => state.players[player].Role === WEREWOLF;

const eliminated = (state: RulesState, player: number): boolean => state.players[player].Eliminated === true;

function voteForPhase(state: RulesState, player: number, phase: number): number {
	return (phase === PHASE_NIGHT ? state.players[player].NightVote : state.players[player].DayVote) as number;
}

export function werewolf_moveResolveVotes_Legal(state: RulesState, _move: MoveInstance, _proposer: number, base: () => string | null): string | null {
	const err = base();
	if (err !== null) return err;
	const phase = state.currentPhaseKey();
	if (phase !== PHASE_DAY && phase !== PHASE_NIGHT) return 'not in a voting phase';
	for (let i = 0; i < state.numPlayers(); i++) {
		if (state.playerIsInactive(i) || eliminated(state, i)) continue;
		if (phase === PHASE_NIGHT && !isWerewolf(state, i)) continue;
		if (voteForPhase(state, i, phase) < 0) return 'not all eligible players have voted';
	}
	return null;
}
