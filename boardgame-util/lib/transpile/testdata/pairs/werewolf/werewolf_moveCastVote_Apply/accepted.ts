// Stored pair: werewolf.(*moveCastVote).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, moves["Cast Vote"].apply.
// One Go method serves moves["Cast Vote"].apply and moves["Cast Night
// Vote"].apply; the hand port builds each from the same function expression,
// copied once.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const PHASE_NIGHT = 2;

export function werewolf_moveCastVote_Apply(state: RulesState, move: MoveInstance): void | string | null {
	const voter = state.players[move.fields.TargetPlayerIndex as number];
	if (state.currentPhaseKey() === PHASE_NIGHT) voter.NightVote = move.fields.VoteTarget;
	else voter.DayVote = move.fields.VoteTarget;
}
