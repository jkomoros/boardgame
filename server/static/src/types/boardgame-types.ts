/**
 * Shared type definitions for generated game state types.
 * Imported by auto-generated _types.ts files in each game's client/ directory.
 */

export type DeepReadonly<T> =
  T extends (...args: never[]) => unknown ? T :
  T extends readonly (infer Item)[] ? readonly DeepReadonly<Item>[] :
  T extends object ? { readonly [Key in keyof T]: DeepReadonly<T[Key]> } :
  T;

/** A visible component resolved from the static deck catalogue. */
export interface VisibleComponent<
  V extends object = Readonly<Record<string, unknown>>,
  D extends object = Readonly<Record<string, unknown>>,
> {
  readonly Index: number;
  readonly Values: DeepReadonly<V>;
  readonly Deck: string;
  readonly GameName: string;
  readonly ID: string;
  /** Absent when this deck has no dynamic values or they were sanitized away. */
  readonly DynamicValues?: DeepReadonly<D>;
}

/** Sanitization's deliberately opaque occupied-slot representation. */
export type OpaqueComponent = Readonly<Record<string, never>>;

export type Component<
  V extends object = Readonly<Record<string, unknown>>,
  D extends object = Readonly<Record<string, unknown>>,
> = VisibleComponent<V, D> | OpaqueComponent;

/** Static chest entry before stack expansion adds instance metadata. */
export interface CatalogComponent<
  V extends object = Readonly<Record<string, unknown>>,
> {
  readonly Index: number;
  readonly Values: DeepReadonly<V>;
}

export function isVisibleComponent<V extends object, D extends object>(
  component: Component<V, D> | null | undefined,
): component is VisibleComponent<V, D>;
export function isVisibleComponent(component: unknown): component is VisibleComponent;
export function isVisibleComponent(component: unknown): component is VisibleComponent {
  if (!isPlainRecord(component)) return false;
  const candidate = component as Partial<VisibleComponent>;
  return Number.isSafeInteger(candidate.Index)
    && (candidate.Index as number) >= 0
    && typeof candidate.Deck === 'string'
    && typeof candidate.GameName === 'string'
    && typeof candidate.ID === 'string'
    && isPlainRecord(candidate.Values)
    && (candidate.DynamicValues === undefined || isPlainRecord(candidate.DynamicValues));
}

function isPlainRecord(value: unknown): value is Readonly<Record<string, unknown>> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false;
  const prototype = Object.getPrototypeOf(value);
  return prototype === Object.prototype || prototype === null;
}

/**
 * An expanded stack as seen by the client after selector expansion.
 * Stacks in raw server state contain indices; the expansion step resolves
 * those indices into full component objects.
 */
export interface ExpandedStack<
  V extends object = Readonly<Record<string, unknown>>,
  D extends object = Readonly<Record<string, unknown>>,
> {
  readonly Deck: string;
  readonly Indexes: readonly number[];
  readonly IDs: readonly string[];
  readonly IDsLastSeen: Readonly<Record<string, number>>;
  readonly ShuffleCount: number;
  readonly Size?: number;
  readonly MaxSize?: number;
  readonly GameName: string;
  readonly Components: readonly (Component<V, D> | null)[];
}

/**
 * A raw stack as stored on the server, before selector expansion.
 * Used for stacks nested inside boards (which are not expanded).
 */
export interface RawStack {
  readonly Deck: string;
  readonly Indexes: readonly number[];
  readonly IDs: readonly string[];
  readonly IDsLastSeen: Readonly<Record<string, number>>;
  readonly ShuffleCount: number;
  readonly Size?: number;
  readonly MaxSize?: number;
}

/**
 * A raw board nested in dynamic component values. Top-level game/player
 * boards are expanded by the renderer selector and use ExpandedBoard.
 */
export interface Board {
  readonly Spaces: readonly RawStack[];
}

/**
 * An expanded board after selector expansion. Each space is a fully
 * expanded stack with resolved component objects, ready for rendering.
 */
export interface ExpandedBoard<
  T extends object = Readonly<Record<string, unknown>>,
  D extends object = Readonly<Record<string, unknown>>,
> {
  readonly Spaces: readonly ExpandedStack<T, D>[];
}

/** Stable timer identity in a renderer snapshot; live clock values are selective signals. */
export interface ExpandedTimer {
  readonly ID: string;
  readonly IsTimer: true;
}

/** Static game metadata delivered beside state snapshots. */
export interface GameChest<
  C extends object = object,
  K extends object = object,
  E extends object = Readonly<Record<string, {
    readonly Values?: Readonly<Record<string, string>>;
  }>>,
> {
  readonly Decks?: DeepReadonly<C>;
  /** Exact enum names and value unions are supplied by generated contracts. */
  readonly Enums?: DeepReadonly<E>;
  /** Exact names and values are supplied by each game's generated contract. */
  readonly Constants?: DeepReadonly<K>;
  readonly LegalTemplates?: Readonly<Record<string, string>>;
}

/**
 * Full game state object passed to game renderers via the `state` property.
 * GS is the game-level state interface, PS is the per-player state interface.
 */
export interface FullGameState<
  GS extends object,
  PS extends object,
  GC extends object = Readonly<Record<string, unknown>>,
  PC extends object = Readonly<Record<string, unknown>>,
  DC extends object = Readonly<Record<string, readonly unknown[]>>,
> {
  readonly Game: DeepReadonly<GS>;
  readonly Players: readonly DeepReadonly<PS>[];
  readonly Computed?: {
    readonly Global?: DeepReadonly<GC>;
    readonly Players?: readonly DeepReadonly<PC>[];
  };
  readonly Components?: DeepReadonly<DC>;
}
