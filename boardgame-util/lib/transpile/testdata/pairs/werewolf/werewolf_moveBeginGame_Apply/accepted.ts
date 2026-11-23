// Stored pair: werewolf.(*moveBeginGame).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, moves["Begin Game"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const PHASE_DAY = 1;

const PHASE_NIGHT = 2;

const VILLAGER = 'Villager';

const WEREWOLF = 'Werewolf';

const VOTE_DURATION_MS = 45000;

const DAY_TIMEOUT = 'Resolve Day Votes on Timeout';

const NIGHT_TIMEOUT = 'Resolve Night Votes on Timeout';

const timedVoting = (state: RulesState): boolean => state.variant().timedvoting === 'on';

const isWerewolf = (state: RulesState, player: number): boolean => state.players[player].Role === WEREWOLF;

function startVoteTimer(state: RulesState): void {
	if (!timedVoting(state)) return;
	const name = state.currentPhaseKey() === PHASE_NIGHT ? NIGHT_TIMEOUT : DAY_TIMEOUT;
	// Bound to the state being built, as Go's MoveByNameForState is here.
	const completion = state.newMove(name);
	if (completion === null) throw new Error('timed vote resolver was not configured');
	state.timer('VoteTimer').start(VOTE_DURATION_MS, completion);
}

function populateFellowWolves(state: RulesState): void {
	const wolves: number[] = [];
	for (let i = 0; i < state.numPlayers(); i++) {
		state.players[i].FellowWolves = null;
		if (state.playerIsInactive(i) || !isWerewolf(state, i)) continue;
		wolves.push(i);
	}
	for (const wolf of wolves) {
		const fellows = wolves.filter(other => other !== wolf);
		if (fellows.length > 0) state.players[wolf].FellowWolves = fellows;
	}
}

export function werewolf_moveBeginGame_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	state.setPhase(PHASE_DAY);
	const active: number[] = [];
	for (let i = 0; i < state.numPlayers(); i++) if (!state.playerIsInactive(i)) active.push(i);
	const numWerewolves = active.length >= 6 ? 2 : 1;
	// One draw: a permutation whose first entries name the werewolves.
	const indices = state.rand().perm(active.length);
	active.forEach((player, i) => {
		state.players[player].Role = indices.slice(0, numWerewolves).indexOf(i) >= 0 ? WEREWOLF : VILLAGER;
	});
	populateFellowWolves(state);
	startVoteTimer(state);
}
