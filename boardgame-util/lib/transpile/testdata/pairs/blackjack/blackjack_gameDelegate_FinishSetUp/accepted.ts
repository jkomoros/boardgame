// Stored pair: blackjack.(*gameDelegate).FinishSetUp (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, setUp.finish.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function blackjack_gameDelegate_FinishSetUp(state: RulesState): void | string | null {
	return state.gameStack('DrawStack').shuffle();
}
