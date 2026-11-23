// Stored pair: tictactoe.(*movePlaceToken).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/tictactoe.ts, moves["Place Token"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import { moveComponentToSlot } from '../../../../../../../server/static/src/engine/rules/stock.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const PHASE_AFTER_FIRST_MOVE = 1;

function place() {
	return moveComponentToSlot({ source: 'players[move.TargetPlayerIndex].UnusedTokens', destination: 'game.Slots', slotField: 'Slot' });
}

export function tictactoe_movePlaceToken_Apply(state: RulesState, move: MoveInstance): void | string | null {
	place().apply(state, move);
	const player = state.players[state.ensureValid(move.fields.TargetPlayerIndex as number)];
	player.TokensToPlaceThisTurn = (player.TokensToPlaceThisTurn as number) - 1;
	state.setPhase(PHASE_AFTER_FIRST_MOVE);
}
