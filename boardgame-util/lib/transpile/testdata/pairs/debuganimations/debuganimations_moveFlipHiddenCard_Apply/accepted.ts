// Stored pair: debuganimations.(*moveFlipHiddenCard).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/debuganimations.ts, moves["Flip Card Between Hidden and Revealed"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';
import type { Stack } from '../../../../../../../server/static/src/engine/stack.ts';

function firstToFirstSlot(from: Stack, to: Stack): void {
	from.moveComponentToFirstSlot(from.firstComponentIndex(), to);
}

export function debuganimations_moveFlipHiddenCard_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	const hidden = state.gameStack('HiddenCard');
	const visible = state.gameStack('VisibleCard');
	if (hidden.numComponents() > 0) firstToFirstSlot(hidden, visible);
	else firstToFirstSlot(visible, hidden);
}
