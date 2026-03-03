/**
 * Shared type definitions for generated game state types.
 * Imported by auto-generated _types.ts files in each game's client/ directory.
 */

/**
 * A single component instance within an expanded stack.
 * Partial<T> handles generic/hidden components (index -2 = {} at runtime).
 */
export type Component<T extends Record<string, unknown> = Record<string, unknown>> =
  Partial<T> & { Deck: string; GameName: string; ID?: string; DynamicValues?: Record<string, unknown> };

/**
 * An expanded stack as seen by the client after selector expansion.
 * Stacks in raw server state contain indices; the expansion step resolves
 * those indices into full component objects.
 */
export interface ExpandedStack<T extends Record<string, unknown> = Record<string, unknown>> {
  Deck: string;
  Indexes: number[];
  IDs: string[];
  GameName: string;
  Components: (Component<T> | null)[];
  NumComponents: number;
}

/**
 * An expanded timer as seen by the client after selector expansion.
 */
export interface ExpandedTimer {
  ID: string;
  IsTimer: true;
  TimeLeft: number;
  originalTimeLeft: number;
}

/**
 * Full game state object passed to game renderers via the `state` property.
 * GS is the game-level state interface, PS is the per-player state interface.
 */
export interface FullGameState<
  GS extends Record<string, unknown> = Record<string, unknown>,
  PS extends Record<string, unknown> = Record<string, unknown>
> {
  Version: number;
  Game: GS;
  Players: PS[];
  Computed?: { Global?: Record<string, unknown>; Players?: Record<string, unknown>[] };
  Components?: Record<string, Record<number, unknown>>;
}
