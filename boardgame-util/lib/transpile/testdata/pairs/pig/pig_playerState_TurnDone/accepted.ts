// Stored pair: pig.(*playerState).TurnDone (producer hand, extraction mechanical).
// From server/static/src/engine/games/pig.ts, player.traits.turn.turnDone.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function pig_playerState_TurnDone(state: RulesState, player: number): string | null {
	const p = state.players[player];
	if (p.DieCounted !== true) return 'the most recent die roll has not been counted';
	if (p.Eliminated !== true && p.Done !== true) return 'they have not either busted or signaled that they are done';
	return null;
}
