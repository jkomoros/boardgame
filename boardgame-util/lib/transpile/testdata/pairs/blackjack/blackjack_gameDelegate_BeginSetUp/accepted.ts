// Stored pair: blackjack.(*gameDelegate).BeginSetUp (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, setUp.begin.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const DEFAULT_MAX_ROUNDS = 5;

export function blackjack_gameDelegate_BeginSetUp(state: RulesState, variant: Readonly<Record<string, string>>): void | string | null {
	const text = variant.maxrounds ?? '';
	state.g.MaxRounds = /^[+-]?[0-9]+$/.test(text) ? Number(text) : DEFAULT_MAX_ROUNDS;
}
