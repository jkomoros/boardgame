// Stored pair: werewolf.voteForPhase (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, voteForPhase.
// voteForPhase(state, player, phase).
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const PHASE_NIGHT = 2;

export function werewolf_voteForPhase(state: RulesState, player: number, phase: number): number {
	return (phase === PHASE_NIGHT ? state.players[player].NightVote : state.players[player].DayVote) as number;
}
