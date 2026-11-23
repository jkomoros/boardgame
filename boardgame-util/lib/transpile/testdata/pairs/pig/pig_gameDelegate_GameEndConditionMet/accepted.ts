// Stored pair: pig.(*gameDelegate).GameEndConditionMet (producer hand, extraction mechanical).
// From server/static/src/engine/games/pig.ts, end.conditionMet.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function pig_gameDelegate_GameEndConditionMet(state: RulesState): boolean {
	for (const player of state.players) if ((player.Score as number) >= (state.g.TargetScore as number)) return true;
	return false;
}
