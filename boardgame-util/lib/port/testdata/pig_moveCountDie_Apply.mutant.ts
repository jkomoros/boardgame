import { componentAt, dynamic, gameStack, player, playerIndexes } from '../../go/api.ts';
import { at } from '../../go/slice.ts';
import type { RulesState } from '../../rules/state.ts';
import type { MoveInstance } from '../../rules/types.ts';

export function pig_moveCountDie_Apply(state: RulesState, m: MoveInstance): string | null {
	const players = playerIndexes(state);
	const p = at(players, state.ensureValid(state.g.CurrentPlayer as number));
	const value = dynamic(state, componentAt(gameStack(state, 'Die'), 0)).Value as number;
	if (value === 2) {
		player(state, p).Eliminated = true;
	} else {
		player(state, p).RoundScore = (player(state, p).RoundScore as number) + value;
	}
	player(state, p).DieCounted = true;
	return null;
}
