// Stored pair: pig.(*playerState).ResetForTurn (producer hand, extraction mechanical).
// From server/static/src/engine/games/pig.ts, resetForTurn.
// resetForTurn, which the hand port also installs as resetForTurnStart.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function pig_playerState_ResetForTurn(state: RulesState, player: number): void {
	const p = state.players[player];
	p.Done = false;
	p.Eliminated = false;
	p.RoundScore = 0;
	p.DieCounted = true;
}
