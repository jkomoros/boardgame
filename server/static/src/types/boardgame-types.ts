/**
 * Shared type definitions for generated game state types.
 * Imported by auto-generated _types.ts files in each game's client/ directory.
 */

/**
 * A single component instance within an expanded stack.
 * All fields are optional because hidden components (index -2) are bare {}
 * objects at runtime, while normal components have Deck, GameName, etc.
 */
export type Component<
  T extends Record<string, unknown> = Record<string, unknown>,
  D extends Record<string, unknown> = Record<string, unknown>
> = Partial<T & { Deck: string; GameName: string; ID: string; DynamicValues: D }>;

/**
 * An expanded stack as seen by the client after selector expansion.
 * Stacks in raw server state contain indices; the expansion step resolves
 * those indices into full component objects.
 */
export interface ExpandedStack<
  T extends Record<string, unknown> = Record<string, unknown>,
  D extends Record<string, unknown> = Record<string, unknown>
> {
  Deck: string;
  Indexes: number[];
  IDs: string[];
  IDsLastSeen: Record<string, number>;
  ShuffleCount: number;
  Size?: number;
  MaxSize?: number;
  GameName: string;
  Components: (Component<T, D> | null)[];
}

/**
 * A raw stack as stored on the server, before selector expansion.
 * Used for stacks nested inside boards (which are not expanded).
 */
export interface RawStack {
  Deck: string;
  Indexes: number[];
  IDs: string[];
  IDsLastSeen: Record<string, number>;
  ShuffleCount: number;
  Size?: number;
  MaxSize?: number;
}

/**
 * A board as seen by the client before expansion. Boards serialize as an
 * array of spaces, each of which is a raw (non-expanded) stack.
 */
export interface Board {
  Spaces: RawStack[];
}

/**
 * An expanded board after selector expansion. Each space is a fully
 * expanded stack with resolved component objects, ready for rendering.
 */
export interface ExpandedBoard<
  T extends Record<string, unknown> = Record<string, unknown>,
  D extends Record<string, unknown> = Record<string, unknown>
> {
  Spaces: ExpandedStack<T, D>[];
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
  Game: GS;
  Players: PS[];
  Computed?: { Global?: Record<string, unknown>; Players?: Record<string, unknown>[] };
  Components?: Record<string, unknown[]>;
}
