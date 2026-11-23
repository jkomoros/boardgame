// Stored pair: memory.(*moveHideCards).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, moves["Hide Cards"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function memory_moveHideCards_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	// The player hid the cards themself; the timer is no longer needed.
	state.timer('HideCardsTimer').cancel();
	const visible = state.gameStack('VisibleCards');
	const hidden = state.gameStack('HiddenCards');
	for (let i = 0; i < visible.len(); i++) {
		if (visible.componentAt(i) !== null) visible.moveComponent(i, hidden, i);
	}
}
