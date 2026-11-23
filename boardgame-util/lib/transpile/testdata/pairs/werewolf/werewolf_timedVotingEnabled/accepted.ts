// Stored pair: werewolf.timedVotingEnabled (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, timedVoting.
// timedVoting(state).
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function werewolf_timedVotingEnabled(state: RulesState): boolean {
	return state.variant().timedvoting === 'on';
}
