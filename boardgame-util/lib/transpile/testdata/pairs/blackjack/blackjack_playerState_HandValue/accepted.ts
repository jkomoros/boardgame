// Stored pair: blackjack.(*playerState).HandValue (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, handValue.
// handValue(state, player), which the hand port also installs as the computed
// property HandValue.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import { GENERIC_COMPONENT } from '../../../../../../../server/static/src/engine/ids.ts';
import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const TARGET_SCORE = 21;

const RANKS: Record<string, string> = {
	0: 'Unknown', 1: 'Ace', 2: '2', 3: '3', 4: '4', 5: '5', 6: '6', 7: '7', 8: '8', 9: '9', 10: '10', 11: 'Jack', 12: 'Queen', 13: 'King', 14: 'Joker',
};

const RANK_ACE = 1;

const RANK_JACK = 11;

const RANK_KING = 13;

function rankKey(rank: string): number {
	for (const key of Object.keys(RANKS)) {
		if (RANKS[key] === rank) return Number(key);
	}
	return 0;
}

export function blackjack_playerState_HandValue(state: RulesState, player: number): number {
	const hand = state.playerView(player, 'Hand');
	let aces = 0;
	let value = 0;
	for (let i = 0; i < hand.len(); i++) {
		const index = hand.componentAt(i);
		if (index === null) continue;
		const rank = index === GENERIC_COMPONENT ? 0 : rankKey(state.component('cards', index).Rank as string);
		if (rank === RANK_ACE) {
			aces++;
			value++;
		} else if (rank >= RANK_JACK && rank <= RANK_KING) {
			value += 10;
		} else {
			value += rank;
		}
	}
	while (aces > 0) {
		if (value >= TARGET_SCORE - 10) break;
		aces--;
		value += 10;
	}
	return value;
}
