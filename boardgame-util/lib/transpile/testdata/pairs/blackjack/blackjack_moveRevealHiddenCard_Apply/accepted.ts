// Stored pair: blackjack.(*moveRevealHiddenCard).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, moves["Reveal Hidden Card"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function blackjack_moveRevealHiddenCard_Apply(state: RulesState, move: MoveInstance): void | string | null {
	const player = move.fields.TargetPlayerIndex as number;
	const hidden = state.playerStack(player, 'HiddenHand');
	hidden.moveComponentToFirstSlot(hidden.firstComponentIndex(), state.playerStack(player, 'VisibleHand'));
}
