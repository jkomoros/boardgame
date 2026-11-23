// Stored pair: werewolf.(*moveResolveVotesOnTimeout).Legal (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, moves["Resolve Day Votes on Timeout"].legal.
// One Go method serves moves["Resolve Day Votes on Timeout"].legal and
// moves["Resolve Night Votes on Timeout"].legal; the hand port builds each
// from the same function expression, copied once.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import { ADMIN_PLAYER_INDEX } from '../../../../../../../server/static/src/engine/legal/types.ts';
import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const timedVoting = (state: RulesState): boolean => state.variant().timedvoting === 'on';

export function werewolf_moveResolveVotesOnTimeout_Legal(state: RulesState, _move: MoveInstance, proposer: number, base: () => string | null): string | null {
	const err = base();
	if (err !== null) return err;
	if (proposer !== ADMIN_PLAYER_INDEX) return 'only the timer may resolve incomplete votes';
	if (!timedVoting(state) || !state.timer('VoteTimer').active()) return 'timed vote resolution is not active';
	return null;
}
