// Stored pair: debuganimations.(*gameDelegate).DistributeComponentToStarterStack (producer hand, extraction mechanical).
// From server/static/src/engine/games/debuganimations.ts, setUp.distribute.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { Stack } from '../../../../../../../server/static/src/engine/stack.ts';
import type { StackRef } from '../../../../../../../server/static/src/engine/state.ts';

const count = (state: RulesState, prop: string): number => state.gameStack(prop).numComponents();

export function debuganimations_gameDelegate_DistributeComponentToStarterStack(state: RulesState, deck: string, _index: number): Stack | { state: RulesState; at: StackRef } {
	const fill = (prop: string, target: number): Stack | null => (count(state, prop) < target ? state.gameStack(prop) : null);
	if (deck === 'tokens') {
		return fill('TokensTo', 9) ?? fill('SanitizedTokensFrom', 10) ?? fill('SanitizedTokensTo', 9) ?? state.gameStack('TokensFrom');
	}
	return (
		fill('FirstShortStack', 1) ?? fill('SecondShortStack', 1) ?? fill('DiscardStack', 2) ?? fill('HiddenCard', 1) ?? fill('FanStack', 6) ??
		fill('FanDiscard', 3) ?? fill('VisibleStack', 5) ?? fill('HiddenStack', 4) ?? fill('AllVisibleStack', 4) ?? state.gameStack('DrawStack')
	);
}
