// The Go-shaped facade the translator emits calls to: the third column of
// fabric/TRANSPILE-PROPOSAL.md's API table, written once over the engine
// (rules/state.ts, stack.ts). It has Go's shapes, not the engine's: a stack
// operation takes a component and returns an error (`string | null`)
// instead of taking a slot and throwing, so translated code keeps Go's
// control flow, including discarded errors.
//
// Calls the table maps to "the same" engine method (state.numPlayers(),
// state.currentPlayerIndex(), state.next(p), state.rand().intn(n),
// state.moveByName(n), state.newMove(n), state.timer('P').start(ms, m),
// stack length queries through `stackView`, ...) are emitted against the
// engine directly and have no function here.
//
// Types, from the proposal's "Types" table:
//
//   Go *playerState                     number, the player index
//   Go Stack, SizedStack, MergedStack    StackRef
//   Go ComponentInstance (nil-able)      Comp | null
//   Go *dynamicValues struct             DynamicRef
//   Go enum.EnumKey                      number (the key)
//   Go error                             string | null

import type { LegalBindingValue, LegalVerdict } from '../legal/types.ts';
import type { RulesState } from '../rules/state.ts';
import type { MoveInstance } from '../rules/types.ts';
import type { StackView } from '../stack.ts';
import type { StackRef as EngineStackRef, SubStateJSON } from '../state.ts';

/**
 * A stack, as Go code holds one: any stack property (growable, sized or
 * merged) or one space of a board, of the state it was read from. A plain
 * location, so a handle may be held across stack operations.
 */
export interface StackRef {
	readonly state: RulesState;
	/** Where the stack lives: sub-state, property, and board space. */
	readonly at: EngineStackRef;
}

/**
 * A component instance, as Go's ComponentInstance: the stack and slot it was
 * found at, and its identity in its deck. A handle taken before the
 * component moved (a `components(s)` snapshot) still names the old slot;
 * operating on it fails as Go's does. `index` is GENERIC_COMPONENT (-2) for
 * a sanitized component.
 */
export interface Comp {
	readonly stack: StackRef;
	readonly slot: number;
	readonly deck: string;
	readonly index: number;
}

/** A component's dynamic values struct, as Go code holds one. */
export interface DynamicRef {
	readonly deck: string;
	readonly index: number;
}

/**
 * Who holds an enum property: the game sub-state, a player (by index), a
 * component's dynamic values, or a component's static values.
 */
export type Owner = 'game' | number | { readonly dynamicOf: DynamicRef } | { readonly valuesOf: Comp };

/**
 * A grid graph over a range enum (enum/graph: MustNewEnumGridConnectedness),
 * rebuilt from its constructor's constant arguments: the enum's name, its
 * dimensions (EnumDef.Range) and the Go direction names given.
 */
export interface GraphRef {
	readonly enum: string;
	readonly dimensions: readonly number[];
	readonly directions: readonly string[];
}

// ----- sub-states ---------------------------------------------------------

/**
 * Go: `players[i]`, a player sub-state, for reading and writing its scalar
 * properties (`player(state, i).X`). Throws Go's index panic when i is not
 * a player index of this game.
 */
export function player(state: RulesState, i: number): SubStateJSON {
	void state;
	void i;
	throw new Error('go/api: player not implemented');
}

/** Go: `game.X` where X is a stack property of the game state (merged stacks included). */
export function gameStack(state: RulesState, prop: string): StackRef {
	void state;
	void prop;
	throw new Error('go/api: gameStack not implemented');
}

/** Go: `p.X` where X is a stack property of player i's state (merged stacks included). */
export function playerStack(state: RulesState, i: number, prop: string): StackRef {
	void state;
	void i;
	void prop;
	throw new Error('go/api: playerStack not implemented');
}

/**
 * The engine's read-only view of a stack, for the queries the table maps to
 * the same name: `s.Len()`, `NumComponents()`, `SlotsRemaining()`,
 * `MaxSize()` become `stackView(s).len()`, and so on.
 */
export function stackView(s: StackRef): StackView {
	void s;
	throw new Error('go/api: stackView not implemented');
}

// ----- reading stacks -----------------------------------------------------

/** Go: `s.ComponentAt(i)`, `s.ImmutableComponentAt(i)`. Null where Go returns nil. */
export function componentAt(s: StackRef, i: number): Comp | null {
	void s;
	void i;
	throw new Error('go/api: componentAt not implemented');
}

/** Go: `s.First()`, `s.ImmutableFirst()`. Null for an empty stack. */
export function first(s: StackRef): Comp | null {
	void s;
	throw new Error('go/api: first not implemented');
}

/** Go: `s.Last()`, `s.ImmutableLast()`. Null for an empty stack. */
export function last(s: StackRef): Comp | null {
	void s;
	throw new Error('go/api: last not implemented');
}

/**
 * Go: `s.Components()`, `s.ImmutableComponents()`: a snapshot taken now, one
 * entry per slot, null for an empty slot of a sized stack.
 */
export function components(s: StackRef): (Comp | null)[] {
	void s;
	throw new Error('go/api: components not implemented');
}

/** Go: `s.ComponentAtKey(k)`, `s.ImmutableComponentAtKey(k)` on a stack indexed by an enum. */
export function componentAtKey(s: StackRef, key: number): Comp | null {
	void s;
	void key;
	throw new Error('go/api: componentAtKey not implemented');
}

/**
 * Go: `c.Values().(*T)`: the component's static values (the deck's generic
 * values for a generic component). Throws NIL_DEREF_TEXT when c is null.
 */
export function values(state: RulesState, c: Comp | null): Readonly<Record<string, unknown>> {
	void state;
	void c;
	throw new Error('go/api: values not implemented');
}

/**
 * Go: `c.DynamicValues().(*T)`, or a dynamic values struct Go code already
 * holds: its properties, writable. Throws NIL_DEREF_TEXT when c is null.
 */
export function dynamic(state: RulesState, c: Comp | DynamicRef | null): SubStateJSON {
	void state;
	void c;
	throw new Error('go/api: dynamic not implemented');
}

/**
 * Go: `c.ContainingImmutableStack()`: where the component is now, as Go's
 * three results (stack, slot, error). The stack is null when err is not.
 */
export function containingStack(state: RulesState, c: Comp | null): [StackRef | null, number, string | null] {
	void state;
	void c;
	throw new Error('go/api: containingStack not implemented');
}

// ----- moving components --------------------------------------------------

/** Go: `c.MoveTo(dest, slot)`. Null on success, else Go's error text; nothing changes on failure. */
export function moveTo(c: Comp | null, dest: StackRef, slot: number): string | null {
	void c;
	void dest;
	void slot;
	throw new Error('go/api: moveTo not implemented');
}

/** Go: `c.SecretMoveTo(dest, slot)`. */
export function secretMoveTo(c: Comp | null, dest: StackRef, slot: number): string | null {
	void c;
	void dest;
	void slot;
	throw new Error('go/api: secretMoveTo not implemented');
}

/** Go: `c.MoveToFirstSlot(dest)`. */
export function moveToFirstSlot(c: Comp | null, dest: StackRef): string | null {
	void c;
	void dest;
	throw new Error('go/api: moveToFirstSlot not implemented');
}

/** Go: `c.MoveToLastSlot(dest)`. */
export function moveToLastSlot(c: Comp | null, dest: StackRef): string | null {
	void c;
	void dest;
	throw new Error('go/api: moveToLastSlot not implemented');
}

/** Go: `c.MoveToNextSlot(dest)`. */
export function moveToNextSlot(c: Comp | null, dest: StackRef): string | null {
	void c;
	void dest;
	throw new Error('go/api: moveToNextSlot not implemented');
}

/** Go: `c.SlideToFirstSlot()`. */
export function slideToFirstSlot(c: Comp | null): string | null {
	void c;
	throw new Error('go/api: slideToFirstSlot not implemented');
}

/** Go: `c.SlideToLastSlot()`. */
export function slideToLastSlot(c: Comp | null): string | null {
	void c;
	throw new Error('go/api: slideToLastSlot not implemented');
}

/** Go: `c.MayMoveTo(dest)`: null when the move would succeed, else why not. */
export function mayMoveTo(c: Comp | null, dest: StackRef): string | null {
	void c;
	void dest;
	throw new Error('go/api: mayMoveTo not implemented');
}

/** Go: `c.MayMoveToSlot(dest, slot)`. */
export function mayMoveToSlot(c: Comp | null, dest: StackRef, slot: number): string | null {
	void c;
	void dest;
	void slot;
	throw new Error('go/api: mayMoveToSlot not implemented');
}

/** Go: `s.MoveAllTo(dest)`. */
export function moveAllTo(s: StackRef, dest: StackRef): string | null {
	void s;
	void dest;
	throw new Error('go/api: moveAllTo not implemented');
}

/** Go: `s.MoveCountTo(dest, n)`. */
export function moveCountTo(s: StackRef, dest: StackRef, n: number): string | null {
	void s;
	void dest;
	void n;
	throw new Error('go/api: moveCountTo not implemented');
}

/** Go: `s.MayMoveAllTo(dest)`. */
export function mayMoveAllTo(s: StackRef, dest: StackRef): string | null {
	void s;
	void dest;
	throw new Error('go/api: mayMoveAllTo not implemented');
}

/** Go: `s.MayMoveCountTo(dest, n)`. */
export function mayMoveCountTo(s: StackRef, dest: StackRef, n: number): string | null {
	void s;
	void dest;
	void n;
	throw new Error('go/api: mayMoveCountTo not implemented');
}

// ----- rearranging and resizing stacks ------------------------------------

/** Go: `s.Shuffle()`. Draws from state.rand() as Go does. */
export function shuffle(s: StackRef): string | null {
	void s;
	throw new Error('go/api: shuffle not implemented');
}

/** Go: `s.PublicShuffle()`. */
export function publicShuffle(s: StackRef): string | null {
	void s;
	throw new Error('go/api: publicShuffle not implemented');
}

/** Go: `s.SwapComponents(i, j)`. */
export function swapComponents(s: StackRef, i: number, j: number): string | null {
	void s;
	void i;
	void j;
	throw new Error('go/api: swapComponents not implemented');
}

/** Go: `s.SwapComponentsByKey(i, j)` on a stack indexed by an enum. */
export function swapComponentsByKey(s: StackRef, i: number, j: number): string | null {
	void s;
	void i;
	void j;
	throw new Error('go/api: swapComponentsByKey not implemented');
}

/**
 * Go: `s.SortComponents(less)`, with Go's sort.Sort order among equal
 * elements (gosort.ts). `less` gets components, null for an empty slot.
 */
export function sortComponents(s: StackRef, less: (a: Comp | null, b: Comp | null) => boolean): string | null {
	void s;
	void less;
	throw new Error('go/api: sortComponents not implemented');
}

/** Go: `s.SetSize(n)` on a sized stack. */
export function setSize(s: StackRef, n: number): string | null {
	void s;
	void n;
	throw new Error('go/api: setSize not implemented');
}

/** Go: `s.ExpandSize(n)` on a sized stack. */
export function expandSize(s: StackRef, n: number): string | null {
	void s;
	void n;
	throw new Error('go/api: expandSize not implemented');
}

/** Go: `s.ContractSize(n)` on a sized stack. */
export function contractSize(s: StackRef, n: number): string | null {
	void s;
	void n;
	throw new Error('go/api: contractSize not implemented');
}

/** Go: `s.SizeToFit()` on a sized stack. */
export function sizeToFit(s: StackRef): string | null {
	void s;
	throw new Error('go/api: sizeToFit not implemented');
}

// ----- game, players, behaviors -------------------------------------------

/** Go: `state.Game().Variant()[k]`. A missing key gives ''. */
export function variantValue(state: RulesState, key: string): string {
	void state;
	void key;
	throw new Error('go/api: variantValue not implemented');
}

/** Go: `p.SetEliminated()` (behaviors.PlayerElimination). */
export function setEliminated(state: RulesState, i: number): void {
	void state;
	void i;
	throw new Error('go/api: setEliminated not implemented');
}

/** Go: `p.SetPlayerInactive()` (behaviors.InactivePlayer). */
export function setPlayerInactive(state: RulesState, i: number): void {
	void state;
	void i;
	throw new Error('go/api: setPlayerInactive not implemented');
}

/** Go: `p.SetSeatFilled()` (behaviors.Seat). */
export function setSeatFilled(state: RulesState, i: number): void {
	void state;
	void i;
	throw new Error('go/api: setSeatFilled not implemented');
}

// ----- enums --------------------------------------------------------------

/** Go: `v.Value()` on an enum property: its key, converted from the stored string. */
export function enumKey(state: RulesState, owner: Owner, prop: string): number {
	void state;
	void owner;
	void prop;
	throw new Error('go/api: enumKey not implemented');
}

/** Go: `v.SetValue(k)` on an enum property: stores k's string. Null, or Go's error for a key the enum lacks. */
export function setEnumKey(state: RulesState, owner: Owner, prop: string, key: number): string | null {
	void state;
	void owner;
	void prop;
	void key;
	throw new Error('go/api: setEnumKey not implemented');
}

/** Go: `v.Equals(o)` on two enum properties: the same enum and the same key. */
export function enumEquals(state: RulesState, owner: Owner, prop: string, otherOwner: Owner, otherProp: string): boolean {
	void state;
	void owner;
	void prop;
	void otherOwner;
	void otherProp;
	throw new Error('go/api: enumEquals not implemented');
}

/** Go: `v.RangeValue()` on a range-enum property: its indexes, one per dimension. */
export function rangeValue(state: RulesState, owner: Owner, prop: string): number[] {
	void state;
	void owner;
	void prop;
	throw new Error('go/api: rangeValue not implemented');
}

/** Go: `e.RangeToValue(indexes...)`: the key at those indexes, or enum.IllegalValue. */
export function rangeToValue(state: RulesState, enumName: string, indexes: readonly number[]): number {
	void state;
	void enumName;
	void indexes;
	throw new Error('go/api: rangeToValue not implemented');
}

/** Go: `e.ValueToRange(k)`: the indexes of key k. */
export function valueToRange(state: RulesState, enumName: string, key: number): number[] {
	void state;
	void enumName;
	void key;
	throw new Error('go/api: valueToRange not implemented');
}

/** Go: `g.NeighborsByKey(k)` on an enum graph: the neighboring keys, in Go's order. */
export function neighborsByKey(graph: GraphRef, key: number): number[] {
	void graph;
	void key;
	throw new Error('go/api: neighborsByKey not implemented');
}

// ----- moves and legality -------------------------------------------------

/**
 * Go: `m.<Stock>.Apply(state)` inside a move's own Apply: the embedded
 * stock move's apply, which the loader resolves from MoveData.Stock.
 */
export function stockApply(state: RulesState, m: MoveInstance): string | null {
	void state;
	void m;
	throw new Error('go/api: stockApply not implemented');
}

/** Go: `legal.Errorf(key, bindings)`: a failing verdict with that template and bindings. */
export function legalError(key: string, bindings?: Readonly<Record<string, LegalBindingValue>>): LegalVerdict {
	void key;
	void bindings;
	throw new Error('go/api: legalError not implemented');
}
