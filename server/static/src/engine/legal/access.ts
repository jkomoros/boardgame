// The read-only view of a game the legal evaluator needs. This file is the
// seam between the evaluator and whatever holds the state: the evaluator only
// ever sees these interfaces, never a concrete state, stack or enum type.
//
// Each member names the Go method it stands for. An implementation must give
// the same answers Go's would for the same state; the evaluator does no
// stack, deck or enum arithmetic of its own. See README.md, "Binding the
// state-access interface".

import type { PropertyType } from './types.ts';

/**
 * Go: enum.Enum, as much of it as legality reads. Two values belong to the
 * same enum exactly when their enums have the same name (enum names are
 * unique within a chest, which is what Go's pointer comparison relies on).
 */
export interface LegalEnum {
  /** Go: Enum.Name(). */
  readonly name: string;
  /**
   * Go: Enum.ValueFromString(s). The key whose string value matches `s`
   * (Go normalizes with TrimSpace+ToLower, then falls back to `s` being the
   * decimal key itself), or null where Go returns enum.IllegalValue.
   */
  valueFromString(s: string): number | null;
  /**
   * Go: Enum.TreeEnum().Ancestors(key), the self-inclusive ancestor chain,
   * or null when the enum is not a TreeEnum. Only phase enums need it.
   */
  ancestors?(key: number): readonly number[] | null;
}

/** Go: enum.ImmutableVal. */
export interface LegalEnumValue {
  /** Go: Val.Enum(). */
  readonly enum: LegalEnum;
  /** Go: Val.Value(), the integer key. */
  readonly value: number;
  /** Go: Val.String(), the value's display name. */
  readonly string: string;
}

/**
 * Go: PropertyReader, reduced to the two methods everything else is built
 * from.
 *
 * `prop` returns, by declared type:
 *   TypeInt, TypePlayerIndex  -> number (an integer)
 *   TypeBool                  -> boolean
 *   TypeString                -> string
 *   TypeEnum                  -> LegalEnumValue, or null for an unset enum
 *   TypeStack                 -> LegalStack, or null
 *   TypeBoard                 -> LegalBoard, or null
 *   anything else             -> whatever the holder has; legality never reads it
 *
 * `prop` may throw to report a read failure (Go's `(nil, err)`); the message
 * becomes part of an Unknown verdict's reason.
 */
export interface LegalPropertyReader {
  /** Go: PropertyReader.Props(). Property name -> declared type. */
  props(): Readonly<Record<string, PropertyType>>;
  /** Go: PropertyReader.Prop(name). Only called for names present in props(). */
  prop(name: string): unknown;
}

/** Go: ImmutableComponentInstance. */
export interface LegalComponent {
  /** Go: Values().Reader(), the component's static chest-defined values, or null if it has none. */
  values(): LegalPropertyReader | null;
  /**
   * Go: MayMoveTo(dest). null when the move is allowed, otherwise Go's
   * err.Error() text, which becomes the `detail` binding.
   */
  mayMoveTo(destination: LegalStack): string | null;
  /** Go: MayMoveToSlot(dest, slotIndex). Same contract as mayMoveTo. */
  mayMoveToSlot(destination: LegalStack, slotIndex: number): string | null;
}

/** Go: ImmutableStack (growable, sized or merged). */
export interface LegalStack {
  /** Go: Len(), the number of slots. */
  len(): number;
  /** Go: NumComponents(), the number of occupied slots. */
  numComponents(): number;
  /**
   * Go: ImmutableComponentAt(index). null for an empty slot AND for any index
   * outside [0, len()); it must not throw. ImmutableComponentAtKey(key) is
   * componentAt(key) in every Go stack, so there is no separate method.
   */
  componentAt(index: number): LegalComponent | null;
  /** Go: ImmutableFirst(), the first occupied slot's component, or null. */
  first(): LegalComponent | null;
  /** Go: MayMoveAllTo(dest). null or Go's error text. */
  mayMoveAllTo(destination: LegalStack): string | null;
  /** Go: MayMoveCountTo(dest, count). null or Go's error text. */
  mayMoveCountTo(destination: LegalStack, count: number): string | null;
  /** Go: MaySwapComponents(i, j). null or Go's error text. */
  maySwapComponents(i: number, j: number): string | null;
}

/** Go: ImmutableBoard, for "Prop[move.Field]" indexed paths. */
export interface LegalBoard {
  /** Go: Board.Enum(), the enum keying the spaces, or null. */
  enum(): LegalEnum | null;
  /** Go: ImmutableSpaceAt(index), or null when out of range. */
  spaceAt(index: number): LegalStack | null;
  /** Go: ImmutableSpaceAtKey(key), or null when out of range. */
  spaceAtKey(key: number): LegalStack | null;
}

/**
 * Go: ImmutableState plus the three delegate questions legality asks about
 * players. Everything is a method so an implementation can be a thin view.
 */
export interface LegalState {
  /** Go: ImmutableGameState().Reader(). */
  game(): LegalPropertyReader;
  /** Go: ImmutablePlayerStates(), one reader per seat, in seat order. */
  players(): readonly LegalPropertyReader[];
  /** Go: CurrentPlayerIndex() (the delegate's answer; may be a special negative index). */
  currentPlayerIndex(): number;
  /** Go: Manager().Delegate().PlayerMayBeActive(players[index]). Only called for in-bounds indexes. */
  playerMayBeActive(index: number): boolean;
  /** Go: Manager().Delegate().CustomPlayerOrder(state), or null for seat order. */
  customPlayerOrder(): readonly number[] | null;
  /**
   * Go: behaviors.PlayerIsInactive(players[index]): true only when the player
   * state implements PlayerInactiver and reports inactive. Used by
   * allActivePlayers to skip players.
   */
  playerIsInactive(index: number): boolean;
}

/** Go: Move, as much of it as legality reads. */
export interface LegalMove {
  /** Go: Move.Reader(). */
  reader(): LegalPropertyReader;
}

/** Go: ComponentChest, as much of it as predicate construction reads. */
export interface LegalChest {
  /** Go: Enums().Enum(name) for every EnumNames() entry. */
  enums(): readonly LegalEnum[];
}

/**
 * The current phase, as Go's LegalInPhaseCheck sees it. Belongs to the phase
 * machinery, which is a later task; this is the interface it must satisfy.
 */
export interface LegalPhase {
  /** Go: CurrentPhase(state).Value(). */
  readonly key: number;
  /** Go: CurrentPhase(state).String(), used in the "not legal in phase X" text. */
  readonly name: string;
  /**
   * Go: the phase enum's TreeEnum().Ancestors(key), self-inclusive, or null
   * when the phase enum is not a tree (then only `key` itself counts).
   */
  readonly ancestors: readonly number[] | null;
}

/** What a base check supplied by the host reports. */
export type LegalHostCheck =
  /** Go: nil error. */
  | null
  /** Go: a non-nil error; the string is err.Error() and becomes the `detail` binding. */
  | string
  /** The check could not be decided; the string is the Unknown reason. */
  | { unknown: string };

/**
 * The base checks whose logic lives outside this package: phases and move
 * progressions. The plan evaluator takes this as an input (LegalContext.
 * framework); the engine task that owns phases and progressions implements
 * it.
 */
export interface LegalFramework {
  /**
   * Go: Manager().Delegate().CurrentPhase(state). null when the game has no
   * phase configured (Go then treats the phase key as 0 and names it "0").
   */
  currentPhase(state: LegalState): LegalPhase | null;
  /**
   * Go: moves.Default.legalMoveInProgression for the move type named
   * `moveName` - the move-tape match behind the contributed "inProgression"
   * predicate. Optional: without it "inProgression" is Unknown.
   */
  moveInProgression?(moveName: string, state: LegalState, proposer: number): LegalHostCheck;
}

/**
 * Go: LegalContext, the entire vocabulary a predicate may read.
 */
export interface LegalContext {
  /** The state being evaluated against. */
  state: LegalState | null;
  /** The move being evaluated, or null (move.* reads are then Unknown). */
  move: LegalMove | null;
  /** Go: ProposerPlayerIndex. */
  proposer: number;
  /** The phase and progression base checks. See LegalFramework. */
  framework?: LegalFramework | null;
}
