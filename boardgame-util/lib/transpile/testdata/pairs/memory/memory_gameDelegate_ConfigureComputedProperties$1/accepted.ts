// Stored pair: memory.(*gameDelegate).ConfigureComputedProperties.func1 (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, computed["CurrentPlayerHasCardsToReveal"].
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function memory_gameDelegate_ConfigureComputedProperties$1(state: RulesState): unknown {
	return (state.players[state.g.CurrentPlayer as number].CardsLeftToReveal as number) > 0;
}
