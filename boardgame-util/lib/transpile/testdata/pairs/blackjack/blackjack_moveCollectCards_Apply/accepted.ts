// Stored pair: blackjack.(*moveCollectCards).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, moves["Collect Cards"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function blackjack_moveCollectCards_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	const discard = state.gameStack('DiscardStack');
	for (let i = 0; i < state.numPlayers(); i++) {
		state.playerStack(i, 'HiddenHand').moveAllTo(discard);
		state.playerStack(i, 'VisibleHand').moveAllTo(discard);
	}
}
