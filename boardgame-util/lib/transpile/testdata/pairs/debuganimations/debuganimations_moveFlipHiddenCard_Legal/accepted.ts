// Stored pair: debuganimations.(*moveFlipHiddenCard).Legal (producer hand, extraction mechanical).
// From server/static/src/engine/games/debuganimations.ts, moves["Flip Card Between Hidden and Revealed"].legal.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const count = (state: RulesState, prop: string): number => state.gameStack(prop).numComponents();

export function debuganimations_moveFlipHiddenCard_Legal(state: RulesState, _move: MoveInstance, _proposer: number, base: () => string | null): string | null {
	const err = base();
	if (err !== null) return err;
	const hidden = count(state, 'HiddenCard');
	const visible = count(state, 'VisibleCard');
	if (hidden < 1 && visible < 1) return 'Neither the HiddenCard nor RevealedCard is set';
	if (hidden > 0 && visible > 0) return 'both hidden and revealed are full';
	return null;
}
