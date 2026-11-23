// Stored pair: tictactoe.(*playerState).TurnDone (producer hand, extraction mechanical).
// From server/static/src/engine/games/tictactoe.ts, player.traits.turn.turnDone.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function tictactoe_playerState_TurnDone(state: RulesState, player: number): string | null {
	return (state.players[player].TokensToPlaceThisTurn as number) > 0 ? 'they still have tokens left to place this turn' : null;
}
