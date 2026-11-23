// Stored pair: memory.(*gameDelegate).BeginSetUp (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, setUp.begin.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function memory_gameDelegate_BeginSetUp(state: RulesState, variant: Readonly<Record<string, string>>): void | string | null {
	state.g.CardSet = variant.cardset;
	const numCards = variant.numcards === 'small' ? 10 : variant.numcards === 'large' ? 40 : 20;
	state.g.NumCards = numCards;
	state.gameStack('HiddenCards').setSize(numCards);
	state.gameStack('VisibleCards').setSize(numCards);
}
