// Stored pair: debuganimations.(*moveShuffleCards).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/debuganimations.ts, moves["Shuffle"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function debuganimations_moveShuffleCards_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	return state.gameStack('FanStack').shuffle();
}
