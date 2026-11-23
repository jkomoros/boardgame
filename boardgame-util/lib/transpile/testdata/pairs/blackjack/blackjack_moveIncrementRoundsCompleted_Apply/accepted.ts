// Stored pair: blackjack.(*moveIncrementRoundsCompleted).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, moves["Increment Rounds Completed"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function blackjack_moveIncrementRoundsCompleted_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	state.g.RoundsCompleted = (state.g.RoundsCompleted as number) + 1;
}
