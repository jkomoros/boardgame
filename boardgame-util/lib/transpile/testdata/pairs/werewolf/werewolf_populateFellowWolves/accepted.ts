// Stored pair: werewolf.populateFellowWolves (producer hand, extraction wrapped).
// From server/static/src/engine/games/werewolf.ts, populateFellowWolves.
// populateFellowWolves(state) walks every player; Go walks the slice it is
// passed, which its one caller builds from every player.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const WEREWOLF = 'Werewolf';

const isWerewolf = (state: RulesState, player: number): boolean => state.players[player].Role === WEREWOLF;

function populateFellowWolves(state: RulesState): void {
	const wolves: number[] = [];
	for (let i = 0; i < state.numPlayers(); i++) {
		state.players[i].FellowWolves = null;
		if (state.playerIsInactive(i) || !isWerewolf(state, i)) continue;
		wolves.push(i);
	}
	for (const wolf of wolves) {
		const fellows = wolves.filter(other => other !== wolf);
		if (fellows.length > 0) state.players[wolf].FellowWolves = fellows;
	}
}

export function werewolf_populateFellowWolves(state: RulesState, _players: number[] | null): void {
	populateFellowWolves(state);
}
