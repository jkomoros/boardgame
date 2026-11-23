// Stored pair: debuganimations.(*moveStartMoveAllComponentsToVisible).Apply (producer hand, extraction wrapped).
// From server/static/src/engine/games/debuganimations.ts, moves["Start Move All Components To Visible"].apply.
// The hand port builds this move with moveAll(options), and its apply reads
// options from the factory; the wrapper calls the factory with the
// configuration the game passes and calls its apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import { stackEmpty, stackNotEmpty } from '../../../../../../../server/static/src/engine/legal/catalogStack.ts';
import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance, MoveSpec } from '../../../../../../../server/static/src/engine/rules/types.ts';

function moveAll(options: { name: string; helpText: string; from: string; to: string; emptyMessage: string; occupiedMessage: string }): MoveSpec {
	return {
		name: options.name,
		helpText: options.helpText,
		fixUp: false,
		preconditions: [
			{ ...stackNotEmpty('game.' + options.from), message: options.emptyMessage },
			{ ...stackEmpty('game.' + options.to), message: options.occupiedMessage },
		],
		apply(state) {
			const from = state.gameStack(options.from);
			const to = state.gameStack(options.to);
			while (from.numComponents() > 0) from.moveComponentToNextSlot(from.firstComponentIndex(), to);
		},
	};
}

export function debuganimations_moveStartMoveAllComponentsToVisible_Apply(state: RulesState, m: MoveInstance): void | string | null {
	return moveAll({
		name: 'Start Move All Components To Visible', helpText: 'Moves all components from hidden to visible',
		from: 'AllHiddenStack', to: 'AllVisibleStack',
		emptyMessage: 'debuganimations.all_hidden_stack_empty', occupiedMessage: 'debuganimations.all_visible_stack_occupied',
	}).apply!(state, m);
}
