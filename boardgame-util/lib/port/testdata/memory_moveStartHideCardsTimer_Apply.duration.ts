import { nonNil } from '../../go/panic.ts';
import type { RulesState } from '../../rules/state.ts';
import type { MoveInstance } from '../../rules/types.ts';

export function memory_moveStartHideCardsTimer_Apply(state: RulesState, m: MoveInstance): string | null {
	const move = state.moveByName('Hide Cards');
	state.timer('HideCardsTimer').start(5000, nonNil(move));
	return null;
}
