// Stored pair: memory.(*moveStartHideCardsTimer).LegalCustom (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, moves["Start Hide Cards Timer"].legalCustom.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { LegalVerdict } from '../../../../../../../server/static/src/engine/legal/types.ts';
import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function memory_moveStartHideCardsTimer_LegalCustom(state: RulesState, _m: MoveInstance, _proposer: number): string | null | LegalVerdict {
	return (state.timer('HideCardsTimer').active() ? 'the timer is already active' : null);
}
