// Stored pair: blackjack.(*gameDelegate).GameEndConditionMet (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, end.conditionMet.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function blackjack_gameDelegate_GameEndConditionMet(state: RulesState): boolean {
	return (state.g.RoundsCompleted as number) >= (state.g.MaxRounds as number);
}
