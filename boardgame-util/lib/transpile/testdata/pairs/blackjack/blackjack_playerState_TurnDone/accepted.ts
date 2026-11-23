// Stored pair: blackjack.(*playerState).TurnDone (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, player.traits.turn.turnDone.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function blackjack_playerState_TurnDone(state: RulesState, player: number): string | null {
	return state.players[player].Eliminated !== true && state.players[player].Stood !== true ? 'they have neither busted nor decided to stand' : null;
}
