// Stored pair: tictactoe.(*gameDelegate).DistributeComponentToStarterStack (producer hand, extraction mechanical).
// From server/static/src/engine/games/tictactoe.ts, setUp.distribute.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { Stack } from '../../../../../../../server/static/src/engine/stack.ts';
import type { StackRef } from '../../../../../../../server/static/src/engine/state.ts';

const X = 'X';

export function tictactoe_gameDelegate_DistributeComponentToStarterStack(state: RulesState, _deck: string, index: number): Stack | { state: RulesState; at: StackRef } {
	return state.playerStack(state.component('tokens', index).Value === X ? 0 : 1, 'UnusedTokens');
}
