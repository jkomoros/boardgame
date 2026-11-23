// Stored pair: werewolf.startVoteTimer (producer hand, extraction wrapped).
// From server/static/src/engine/games/werewolf.ts, startVoteTimer.
// startVoteTimer(state) throws "timed vote resolver was not configured" where
// Go returns it as an error.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const PHASE_NIGHT = 2;

const VOTE_DURATION_MS = 45000;

const DAY_TIMEOUT = 'Resolve Day Votes on Timeout';

const NIGHT_TIMEOUT = 'Resolve Night Votes on Timeout';

const timedVoting = (state: RulesState): boolean => state.variant().timedvoting === 'on';

function startVoteTimer(state: RulesState): void {
	if (!timedVoting(state)) return;
	const name = state.currentPhaseKey() === PHASE_NIGHT ? NIGHT_TIMEOUT : DAY_TIMEOUT;
	// Bound to the state being built, as Go's MoveByNameForState is here.
	const completion = state.newMove(name);
	if (completion === null) throw new Error('timed vote resolver was not configured');
	state.timer('VoteTimer').start(VOTE_DURATION_MS, completion);
}

export function werewolf_startVoteTimer(state: RulesState): string | null {
	startVoteTimer(state);
	return null;
}
