// Stored pair: memory.(*playerState).TurnDone (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, player.traits.turn.turnDone.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function memory_playerState_TurnDone(state: RulesState, player: number): string | null {
	if ((state.players[player].CardsLeftToReveal as number) > 0) return 'they still have cards left to reveal';
	if (state.gameStack('VisibleCards').numComponents() > 0) return 'there are still some cards revealed, which they must hide';
	return null;
}
