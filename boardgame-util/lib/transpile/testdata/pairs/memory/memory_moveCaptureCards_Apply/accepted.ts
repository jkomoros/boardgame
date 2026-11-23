// Stored pair: memory.(*moveCaptureCards).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, moves["Capture Cards"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function memory_moveCaptureCards_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	const won = state.playerStack(state.g.CurrentPlayer as number, 'WonCards');
	const visible = state.gameStack('VisibleCards');
	for (let i = 0; i < visible.len(); i++) {
		if (visible.componentAt(i) !== null) visible.moveComponentToNextSlot(i, won);
	}
}
