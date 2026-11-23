// Stored pair: werewolf.resolveVotes (producer hand, extraction wrapped).
// From server/static/src/engine/games/werewolf.ts, resolveVotes.
// resolveVotes(state) returns nothing; Go returns an error (from
// SetCurrentPhase), which the hand port's setPhase throws instead.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const PHASE_DAY = 1;

const PHASE_NIGHT = 2;

const WEREWOLF = 'Werewolf';

const isWerewolf = (state: RulesState, player: number): boolean => state.players[player].Role === WEREWOLF;

const eliminated = (state: RulesState, player: number): boolean => state.players[player].Eliminated === true;

function voteForPhase(state: RulesState, player: number, phase: number): number {
	return (phase === PHASE_NIGHT ? state.players[player].NightVote : state.players[player].DayVote) as number;
}

function resolveVotes(state: RulesState): void {
	const phase = state.currentPhaseKey();
	const counts: number[] = state.players.map(() => 0);
	for (let i = 0; i < state.numPlayers(); i++) {
		if (state.playerIsInactive(i) || eliminated(state, i)) continue;
		if (phase === PHASE_NIGHT && !isWerewolf(state, i)) continue;
		const vote = voteForPhase(state, i, phase);
		if (vote >= 0) counts[vote] = (counts[vote] ?? 0) + 1;
	}
	let maxVotes = 0;
	let chosen = -1;
	let tied = false;
	for (let target = 0; target < state.numPlayers(); target++) {
		const count = counts[target];
		if (count === 0) continue;
		if (count > maxVotes) {
			maxVotes = count;
			chosen = target;
			tied = false;
		} else if (count === maxVotes) {
			tied = true;
		}
	}
	// A tied day vote eliminates nobody; a tied night vote keeps the first.
	if (phase === PHASE_DAY && tied) chosen = -1;
	if (chosen >= 0 && chosen < state.numPlayers()) state.players[chosen].Eliminated = true;
	for (const player of state.players) {
		player.DayVote = -1;
		player.NightVote = -1;
	}
	if (phase === PHASE_DAY) {
		state.setPhase(PHASE_NIGHT);
	} else {
		state.g.RoundNumber = (state.g.RoundNumber as number) + 1;
		state.setPhase(PHASE_DAY);
	}
}

export function werewolf_resolveVotes(state: RulesState): string | null {
	resolveVotes(state);
	return null;
}
