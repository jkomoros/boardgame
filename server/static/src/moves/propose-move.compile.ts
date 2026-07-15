import { BoardgameBaseGameRenderer } from '../components/boardgame-base-game-renderer.js';

type Inputs = {
  Empty: Record<string, never>;
  DefaultOnly: { readonly Label?: string };
  Required: { readonly Count: number };
};

declare const renderer: BoardgameBaseGameRenderer<object, object, keyof Inputs, Inputs>;

renderer.proposeMove('Empty');
renderer.proposeMove('DefaultOnly');
renderer.proposeMove('DefaultOnly', { Label: 'custom' });
renderer.proposeMove('Required', { Count: 1 });

// @ts-expect-error A move with required creator input cannot omit its argument.
renderer.proposeMove('Required');
// @ts-expect-error Exact creator inputs reject extra fields.
renderer.proposeMove('DefaultOnly', { Surprise: true });
