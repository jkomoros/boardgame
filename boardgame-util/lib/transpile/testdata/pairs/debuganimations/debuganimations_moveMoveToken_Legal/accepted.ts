// Stored pair: debuganimations.(*moveMoveToken).Legal (producer hand, extraction wrapped).
// From server/static/src/engine/games/debuganimations.ts, moves["Move Token"].legal.
// The hand port builds this move with toggleThird(options), and its legal
// reads options from the factory; the wrapper calls the factory with the
// configuration the game passes and calls its legal.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { LegalResult, MoveInstance, MoveSpec } from '../../../../../../../server/static/src/engine/rules/types.ts';

const count = (state: RulesState, prop: string): number => state.gameStack(prop).numComponents();

function toggleThird(options: { name: string; helpText: string; full: string; other: string; fullCount: number; otherCount: number; error: string }): MoveSpec {
	return {
		name: options.name,
		helpText: options.helpText,
		fixUp: false,
		legal(state, _move, _proposer, base): LegalResult {
			const err = base();
			if (err !== null) return err;
			const full = count(state, options.full);
			const other = count(state, options.other);
			if (full === options.fullCount && other === options.otherCount) return null;
			if (full === options.fullCount - 1 && other === options.otherCount + 1) return null;
			return options.error;
		},
		apply(state) {
			const full = state.gameStack(options.full);
			const other = state.gameStack(options.other);
			if (full.numComponents() < options.fullCount) other.moveComponent(other.firstComponentIndex(), full, 2);
			else full.moveComponentToFirstSlot(2, other);
		},
	};
}

export function debuganimations_moveMoveToken_Legal(state: RulesState, m: MoveInstance, proposer: number, base: () => string | null): string | null {
	return toggleThird({
		name: 'Move Token', helpText: 'Moves tokens',
		full: 'TokensFrom', other: 'TokensTo', fullCount: 10, otherCount: 9, error: "tokens aren't in known position",
	}).legal!(state, m, proposer, base);
}
