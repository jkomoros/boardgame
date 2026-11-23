// Stored pair: debuganimations.(*gameDelegate).CheckGameFinished (producer hand, extraction none).
// From server/static/src/engine/games/debuganimations.ts, no slot.
// the hand port has no `end` ("Never ends"), which the engine reads as never
// finished; Go returns false, nil. The body is that default.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function debuganimations_gameDelegate_CheckGameFinished(_state: RulesState): [finished: boolean, winners: number[] | null] {
	return [false, null];
}
