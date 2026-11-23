// Stored pair: memory.(*gameDelegate).FinishSetUp (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, setUp.finish.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

type Card = { Type: string; CardSet: string };

export function memory_gameDelegate_FinishSetUp(state: RulesState): void | string | null {
	const unused = state.gameStack('UnusedCards');
	const hidden = state.gameStack('HiddenCards');
	const numCards = state.g.NumCards as number;
	const cardSet = state.g.CardSet as string;
	const card = (slot: number): Card => state.component('cards', unused.componentAt(slot) as number) as Card;

	unused.shuffle();
	// Take pairs from the front: into the layout if they are of the
	// chosen set, otherwise out of the way at the back.
	while (hidden.numComponents() < numCards) {
		const first = card(0);
		let pairIndex = 0;
		for (let i = 1; i < unused.len(); i++) {
			if (card(i).Type === first.Type) {
				pairIndex = i;
				break;
			}
		}
		if (pairIndex === 0) throw new Error('unexpectedly unable to find the pair when sorting cards to include');
		if (cardSet === 'all' || cardSet === first.CardSet) {
			unused.moveComponentToNextSlot(pairIndex, hidden);
			unused.moveComponentToNextSlot(0, hidden);
		} else {
			unused.slideToLastSlot(pairIndex);
			unused.slideToLastSlot(0);
		}
	}
	hidden.shuffle();
	state.players[0].CardsLeftToReveal = 2;
}
