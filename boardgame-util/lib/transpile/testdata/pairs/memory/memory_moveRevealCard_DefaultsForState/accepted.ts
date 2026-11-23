// Stored pair: memory.(*moveRevealCard).DefaultsForState (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, moves["Reveal Card"].defaults.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function memory_moveRevealCard_DefaultsForState(state: RulesState, move: MoveInstance): void {
	const hidden = state.gameStack('HiddenCards');
	for (let i = 0; i < hidden.len(); i++) {
		if (hidden.componentAt(i) === null) continue;
		move.fields.CardIndex = i;
		break;
	}
}
