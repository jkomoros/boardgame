// Stored pair: memory.(*gameState).CurrentPlayerHasCardsToReveal (producer hand, extraction wrapped).
// From server/static/src/engine/games/memory.ts, computed["CurrentPlayerHasCardsToReveal"].
// the hand port inlines it into the computed property
// CurrentPlayerHasCardsToReveal, which is copied here (the same text as
// ConfigureComputedProperties.func1).
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function memory_gameState_CurrentPlayerHasCardsToReveal(state: RulesState): boolean {
	return (state.players[state.g.CurrentPlayer as number].CardsLeftToReveal as number) > 0;
}
