// Stored pair: tictactoe.(*movePlaceToken).DefaultsForState (producer hand, extraction mechanical).
// From server/static/src/engine/games/tictactoe.ts, moves["Place Token"].defaults.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function tictactoe_movePlaceToken_DefaultsForState(state: RulesState, move: MoveInstance): void {
	const slots = state.gameStack('Slots');
	for (let i = 0; i < slots.len(); i++) {
		if (slots.componentAt(i) !== null) continue;
		move.fields.Slot = i;
		break;
	}
}
