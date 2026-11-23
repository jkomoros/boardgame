// Stored pair: pig.(*gameDelegate).FinishSetUp (producer hand, extraction mechanical).
// From server/static/src/engine/games/pig.ts, setUp.finish.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const DEFAULT_TARGET_SCORE = 100;

export function pig_gameDelegate_FinishSetUp(state: RulesState): void | string | null {
	// A random player starts.
	state.g.CurrentPlayer = state.rand().intn(state.numPlayers());
	state.g.TargetScore = DEFAULT_TARGET_SCORE;
}
