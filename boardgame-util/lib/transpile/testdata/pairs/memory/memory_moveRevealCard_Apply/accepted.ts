// Stored pair: memory.(*moveRevealCard).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, moves["Reveal Card"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import { moveComponentToSlot } from '../../../../../../../server/static/src/engine/rules/stock.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

function reveal() {
	return moveComponentToSlot({ source: 'game.HiddenCards', destination: 'game.VisibleCards', slotField: 'CardIndex', sourceSlotField: 'CardIndex' });
}

export function memory_moveRevealCard_Apply(state: RulesState, move: MoveInstance): void | string | null {
	reveal().apply(state, move);
	const player = state.players[state.ensureValid(state.g.CurrentPlayer as number)];
	player.CardsLeftToReveal = (player.CardsLeftToReveal as number) - 1;
}
