import { BoardgameBaseGameRenderer } from '../components/boardgame-base-game-renderer.js';
import type { FullGameState } from '../types/boardgame-types.js';

type Inputs = {
  Empty: Record<string, never>;
  Required: { readonly Count: number };
  Choose: { readonly Mode: 'Fast' | 'Safe' };
};

type State = FullGameState<object, object>;
declare const renderer: BoardgameBaseGameRenderer<State, object, keyof Inputs, Inputs>;

renderer.move('Empty').propose();
renderer.move('Required').with({ Count: 1 }).propose();
renderer.move('Choose').with({ Mode: 'Fast' }).activate();

// @ts-expect-error the ungated proposal shortcut is intentionally absent
renderer.proposeMove('Empty');
// @ts-expect-error required input must be bound before proposal exists
renderer.move('Required').propose();
// @ts-expect-error exact creator inputs reject extra fields
renderer.move('Required').with({ Count: 1, Surprise: true });
// @ts-expect-error named enums preserve their generated value union
renderer.move('Choose').with({ Mode: 'Slow' });
