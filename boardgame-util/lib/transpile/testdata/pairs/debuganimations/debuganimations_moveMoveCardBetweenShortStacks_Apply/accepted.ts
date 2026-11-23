// Stored pair: debuganimations.(*moveMoveCardBetweenShortStacks).Apply (producer hand, extraction wrapped).
// From server/static/src/engine/games/debuganimations.ts, moves["Move Card Between Short Stacks"].apply.
// The hand port builds this move with moveBetween(options), and its apply
// reads ends, options from the factory; the wrapper calls the factory with
// the configuration the game passes and calls its apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance, MoveSpec } from '../../../../../../../server/static/src/engine/rules/types.ts';
import type { Stack } from '../../../../../../../server/static/src/engine/stack.ts';

const count = (state: RulesState, prop: string): number => state.gameStack(prop).numComponents();

function firstToFirstSlot(from: Stack, to: Stack): void {
	from.moveComponentToFirstSlot(from.firstComponentIndex(), to);
}

function moveBetween(options: { name: string; helpText: string; field: string; whenTrue: [string, string]; defaultTrue: (state: RulesState) => boolean }): MoveSpec {
	const ends = (state: RulesState, flag: unknown): [Stack, Stack] => {
		const [from, to] = flag === true ? options.whenTrue : [options.whenTrue[1], options.whenTrue[0]];
		return [state.gameStack(from), state.gameStack(to)];
	};
	return {
		name: options.name,
		helpText: options.helpText,
		fixUp: false,
		fields: { [options.field]: { type: 'bool', sanitize: 'all:visible' } },
		defaults(state, move) {
			move.fields[options.field] = options.defaultTrue(state);
		},
		legal(state, move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			const [from, to] = ends(state, move.fields[options.field]);
			if (from.numComponents() === 0) return 'source stack has no cards to move';
			return from.mayMoveComponentTo(from.firstComponentIndex(), to);
		},
		apply(state, move) {
			const [from, to] = ends(state, move.fields[options.field]);
			firstToFirstSlot(from, to);
		},
	};
}

export function debuganimations_moveMoveCardBetweenShortStacks_Apply(state: RulesState, m: MoveInstance): void | string | null {
	return moveBetween({
		name: 'Move Card Between Short Stacks',
		helpText: 'Moves a card between two short stacks',
		field: 'FromFirst',
		whenTrue: ['FirstShortStack', 'SecondShortStack'],
		defaultTrue: state => count(state, 'FirstShortStack') >= 1,
	}).apply!(state, m);
}
