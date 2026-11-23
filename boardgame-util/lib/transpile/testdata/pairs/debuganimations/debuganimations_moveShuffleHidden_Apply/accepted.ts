// Stored pair: debuganimations.(*moveShuffleHidden).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/debuganimations.ts, moves["Shuffle Hidden"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function debuganimations_moveShuffleHidden_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	state.gameStack('FanDiscard').shuffle();
	state.g.FanShuffleCount = (state.g.FanShuffleCount as number) + 1;
}
