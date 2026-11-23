// Stored pair: memory.(*moveStartHideCardsTimer).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, moves["Start Hide Cards Timer"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const HIDE_CARDS_DURATION_MS = 4000;

const HIDE_CARDS = 'Hide Cards';

export function memory_moveStartHideCardsTimer_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	state.timer('HideCardsTimer').start(HIDE_CARDS_DURATION_MS, state.moveByName(HIDE_CARDS));
}
