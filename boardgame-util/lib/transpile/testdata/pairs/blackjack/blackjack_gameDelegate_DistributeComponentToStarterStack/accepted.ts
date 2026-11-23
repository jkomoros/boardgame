// Stored pair: blackjack.(*gameDelegate).DistributeComponentToStarterStack (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, setUp.distribute.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { Stack } from '../../../../../../../server/static/src/engine/stack.ts';
import type { StackRef } from '../../../../../../../server/static/src/engine/state.ts';

export function blackjack_gameDelegate_DistributeComponentToStarterStack(state: RulesState, _deck: string, index: number): Stack | { state: RulesState; at: StackRef } {
	return state.gameStack(state.component('cards', index).Rank === 'Joker' ? 'UnusedCards' : 'DrawStack');
}
