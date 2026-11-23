// Stored pair: pig.(*playerState).ResetForTurnEnd (producer hand, extraction mechanical).
// From server/static/src/engine/games/pig.ts, player.traits.turn.resetForTurnEnd.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

function resetForTurn(state: RulesState, player: number): void {
	const p = state.players[player];
	p.Done = false;
	p.Eliminated = false;
	p.RoundScore = 0;
	p.DieCounted = true;
}

export function pig_playerState_ResetForTurnEnd(state: RulesState, player: number): void | string | null {
	const p = state.players[player];
	if (p.Done === true) p.Score = (p.Score as number) + (p.RoundScore as number);
	resetForTurn(state, player);
}
