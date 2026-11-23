// Stored pair: blackjack.(*moveCollectCards).Legal (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, moves["Collect Cards"].legal.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const hasCards = (state: RulesState, player: number): boolean =>
	state.playerStack(player, 'HiddenHand').numComponents() > 0 || state.playerStack(player, 'VisibleHand').numComponents() > 0;

export function blackjack_moveCollectCards_Legal(state: RulesState, _move: MoveInstance, _proposer: number, base: () => string | null): string | null {
	const err = base();
	if (err !== null) return err;
	for (let i = 0; i < state.numPlayers(); i++) if (hasCards(state, i)) return null;
	return 'no player has cards to collect';
}
