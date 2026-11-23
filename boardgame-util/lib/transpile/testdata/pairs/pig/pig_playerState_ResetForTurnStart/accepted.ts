// Stored pair: pig.(*playerState).ResetForTurnStart (producer hand, extraction mechanical).
// From server/static/src/engine/games/pig.ts, player.traits.turn.resetForTurnStart.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function pig_playerState_ResetForTurnStart(state: RulesState, player: number): void | string | null {
	const p = state.players[player];
	p.Done = false;
	p.Eliminated = false;
	p.RoundScore = 0;
	p.DieCounted = true;
}
