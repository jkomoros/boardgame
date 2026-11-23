// Stored pair: werewolf.(*moveResolveVotes).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, moves["Resolve Votes"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import { checkGameFinished } from '../../../../../../../server/static/src/engine/rules/engine.ts';
import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const PHASE_DAY = 1;

const PHASE_NIGHT = 2;

const WEREWOLF = 'Werewolf';

const VOTE_DURATION_MS = 45000;

const DAY_TIMEOUT = 'Resolve Day Votes on Timeout';

const NIGHT_TIMEOUT = 'Resolve Night Votes on Timeout';

const timedVoting = (state: RulesState): boolean => state.variant().timedvoting === 'on';

const isWerewolf = (state: RulesState, player: number): boolean => state.players[player].Role === WEREWOLF;

const eliminated = (state: RulesState, player: number): boolean => state.players[player].Eliminated === true;

function voteForPhase(state: RulesState, player: number, phase: number): number {
	return (phase === PHASE_NIGHT ? state.players[player].NightVote : state.players[player].DayVote) as number;
}

function startVoteTimer(state: RulesState): void {
	if (!timedVoting(state)) return;
	const name = state.currentPhaseKey() === PHASE_NIGHT ? NIGHT_TIMEOUT : DAY_TIMEOUT;
	// Bound to the state being built, as Go's MoveByNameForState is here.
	const completion = state.newMove(name);
	if (completion === null) throw new Error('timed vote resolver was not configured');
	state.timer('VoteTimer').start(VOTE_DURATION_MS, completion);
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

export function werewolf_moveResolveVotes_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	state.timer('VoteTimer').cancel();
	resolveVotes(state);
	if (checkGameFinished(state).finished) return;
	startVoteTimer(state);
}
