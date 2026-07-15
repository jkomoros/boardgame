import { BoardgameBaseGameRenderer } from '../components/boardgame-base-game-renderer.js';
import type { FullGameState } from '../types/boardgame-types.js';

type Inputs = {
  Empty: Record<string, never>;
  DefaultOnly: { readonly Label?: string };
  Required: { readonly Count: number };
  Choose: { readonly Mode: 'Fast' | 'Safe' };
};

type State = FullGameState<object, object>;
declare const renderer: BoardgameBaseGameRenderer<State, object, keyof Inputs, Inputs>;

renderer.proposeMove('Empty');
renderer.proposeMove('DefaultOnly');
renderer.proposeMove('DefaultOnly', { Label: 'custom' });
renderer.proposeMove('Required', { Count: 1 });
renderer.proposeMove('Choose', { Mode: 'Fast' });

// @ts-expect-error A move with required creator input cannot omit its argument.
renderer.proposeMove('Required');
// @ts-expect-error Exact creator inputs reject extra fields.
renderer.proposeMove('DefaultOnly', { Surprise: true });
const aliasedExtra = { Label: 'custom', Surprise: true };
// @ts-expect-error Aliasing cannot bypass exact creator inputs.
renderer.proposeMove('DefaultOnly', aliasedExtra);
const spreadExtra = { ...aliasedExtra };
// @ts-expect-error Spreading cannot bypass exact creator inputs.
renderer.proposeMove('DefaultOnly', spreadExtra);
// @ts-expect-error Context-owned and unknown fields are absent from creator inputs.
renderer.proposeMove('Required', { Count: 1, TargetPlayerIndex: 0 });
// @ts-expect-error Named enums preserve their exact generated value union.
renderer.proposeMove('Choose', { Mode: 'Slow' });
// @ts-expect-error Zero-input moves reject argument fields.
renderer.proposeMove('Empty', { Surprise: true });
