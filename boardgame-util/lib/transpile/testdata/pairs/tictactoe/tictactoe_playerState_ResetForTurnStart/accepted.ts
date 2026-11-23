// Stored pair: tictactoe.(*playerState).ResetForTurnStart (producer hand, extraction mechanical).
// From server/static/src/engine/games/tictactoe.ts, player.traits.turn.resetForTurnStart.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function tictactoe_playerState_ResetForTurnStart(state: RulesState, player: number): void | string | null {
	state.players[player].TokensToPlaceThisTurn = 1;
}
