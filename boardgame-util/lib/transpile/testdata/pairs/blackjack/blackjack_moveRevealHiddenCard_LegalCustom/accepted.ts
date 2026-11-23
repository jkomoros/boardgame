// Stored pair: blackjack.(*moveRevealHiddenCard).LegalCustom (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, moves["Reveal Hidden Card"].legalCustom.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { LegalVerdict } from '../../../../../../../server/static/src/engine/legal/types.ts';
import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function blackjack_moveRevealHiddenCard_LegalCustom(state: RulesState, move: MoveInstance, _proposer: number): string | null | LegalVerdict {
	const player = move.fields.TargetPlayerIndex as number;
	const hidden = state.playerStack(player, 'HiddenHand');
	// Go dereferences the first component without checking for one; the
	// declared precondition above is what guarantees it.
	if (hidden.numComponents() === 0) throw new Error('runtime error: invalid memory address or nil pointer dereference');
	return hidden.mayMoveComponentTo(hidden.firstComponentIndex(), state.playerStack(player, 'VisibleHand'));
}
