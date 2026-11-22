// The Go-shaped facade the translator emits calls to: the third column of
// fabric/TRANSPILE-PROPOSAL.md's API table, written once over the engine
// (rules/state.ts, stack.ts). It has Go's shapes, not the engine's: a stack
// operation takes a component and returns an error (`string | null`)
// instead of taking a slot and throwing, so translated code keeps Go's
// control flow, including discarded errors. A failed operation changes
// nothing (the engine validates before it mutates, as Go does).
//
// Calls the table maps to "the same" engine method (state.numPlayers(),
// state.currentPlayerIndex(), state.next(p), state.rand().intn(n),
// state.moveByName(n), state.newMove(n), state.timer('P').start(ms, m),
// t.active(), t.cancel(), stack length queries through `stackView`, ...) are
// emitted against the engine directly and have no function here.
//
// Types, from the proposal's "Types" table:
//
//   Go *playerState                     number, the player index
//   Go Stack, SizedStack, MergedStack    StackRef
//   Go ComponentInstance (nil-able)      Comp | null
//   Go *dynamicValues struct             DynamicRef
//   Go enum.EnumKey                      number (the key)
//   Go error                             string | null
//
// A Go ComponentInstance is the component's identity (deck and index) and
// the state it was read from, not a slot: every operation on it finds where
// the component is now (state.containingStack). So is a Comp. A handle from
// a `components(s)` snapshot whose component has since moved still names
// that component, and operating on it acts on it where it now is, exactly
// as in Go. What fails is what fails in Go: a component the state's index
// has lost track of (after sizedStack.ContractSize, "that component's
// location is not public information"), or a generic one.
//
// Every answer is held to Go's by testdata/gen/facadevec (go/stacks.test.ts,
// go/lang.test.ts).

import type { LegalBindingValue, LegalVerdict } from '../legal/types.ts';
import { failT } from '../legal/types.ts';
import { EMPTY_SLOT, GENERIC_COMPONENT } from '../ids.ts';
import type { PropDef } from '../definition.ts';
import type { BoundEnum } from '../rules/enums.ts';
import type { RulesState } from '../rules/state.ts';
import type { MoveInstance } from '../rules/types.ts';
import { isStack, isStackError } from '../stack.ts';
import type { Stack, StackView } from '../stack.ts';
import type { EngineState, StackRef as EngineStackRef, SubStateJSON } from '../state.ts';
import { indexOutOfRange, nilDeref } from './panic.ts';
import { atoi, itoa } from './str.ts';

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
 * A component instance, as Go's ComponentInstance: its identity in its deck,
 * and the stack and slot it was found at. Operations use the identity (and
 * the state of `stack`), never `slot`, as Go's do. `slot` is -1 for a
 * component handed to a `sortComponents` less function, which has no slot.
 * `index` is GENERIC_COMPONENT (-2) for a sanitized component.
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
 * component's dynamic values, or a component's static values. The component
 * may be null (Go's nil), which panics on use as Go's does.
 */
export type Owner = 'game' | number | { readonly dynamicOf: Comp | DynamicRef | null } | { readonly valuesOf: Comp | null };

/**
 * A grid graph over a range enum (enum/graph: MustNewEnumGridConnectedness),
 * rebuilt from its constructor's constant arguments: the enum's name, its
 * dimensions (EnumDef.Range) and the Go direction names given
 * (`DirectionUp`, `DirectionDown`, `DirectionLeft`, `DirectionRight`,
 * `DirectionPerpendicular`, `DirectionDiagonal`), in order.
 */
export interface GraphRef {
	readonly enum: string;
	readonly dimensions: readonly number[];
	readonly directions: readonly string[];
}

/**
 * Go: enum.IllegalValue (math.MaxInt64), as a number. It is not a safe
 * integer (it reads as 2^63), but it is the same number the translator's
 * folded constant `enum.IllegalValue` becomes, so comparisons agree. It is
 * never stored: storing it fails in Go too.
 */
export const ILLEGAL_ENUM_VALUE = 9223372036854775807;

// ----- internals ------------------------------------------------------------

/** A Go stack error as the facade returns it: run f, and turn a StackError into its text. */
function stackCall(f: () => void): string | null {
	try {
		f();
		return null;
	} catch (e) {
		if (isStackError(e)) return e.message;
		throw e;
	}
}

function engineOf(s: StackRef): EngineState {
	return s.state.engine;
}

function viewOf(s: StackRef): StackView {
	return engineOf(s).view(s.at);
}

/** The physical stack a ref names, or null for a merged stack. */
function physicalOf(s: StackRef): Stack | null {
	const view = viewOf(s);
	return isStack(view) ? view : null;
}

/**
 * The physical stack for a Go call that takes a `Stack`, which a
 * MergedStack is not: a translation that passes one did not compile in Go.
 */
function requirePhysical(s: StackRef, fn: string): Stack {
	const stack = physicalOf(s);
	if (stack === null) throw new Error('go/api: ' + fn + ': ' + JSON.stringify(s.at) + ' is a merged stack, which Go cannot pass here');
	return stack;
}

/** The facade's handle for one physical stack of a state. */
function refOf(state: RulesState, stack: Stack): StackRef {
	for (const sub of state.engine.subStates) {
		const prefix = sub.path + '.';
		if (!stack.path.startsWith(prefix)) continue;
		const rest = stack.path.slice(prefix.length);
		const bracket = rest.indexOf('[');
		if (bracket < 0) return { state, at: { ...sub.ref, Prop: rest } };
		return { state, at: { ...sub.ref, Prop: rest.slice(0, bracket), Space: Number(rest.slice(bracket + 1, -1)) } };
	}
	throw new Error('go/api: no sub-state owns stack ' + stack.path);
}

function compAt(s: StackRef, view: StackView, slot: number): Comp | null {
	const index = view.componentAt(slot);
	if (index === null) return null;
	return { stack: s, slot, deck: view.deckName, index };
}

/** state.containingStack for a component handle: where it is now, or Go's error text. */
function locate(c: Comp): { stack: Stack; slotIndex: number } | string {
	return engineOf(c.stack).locateComponent(c.deck, c.index);
}

const NOT_IN_MUTABLE_STACK = 'The source component was not in a mutable stack: ';

// ----- sub-states ---------------------------------------------------------

/**
 * Go: `players[i]`, a player sub-state, for reading and writing its scalar
 * properties (`player(state, i).X`). Throws Go's index panic when i is not
 * a player index of this game.
 */
export function player(state: RulesState, i: number): SubStateJSON {
	const count = state.numPlayers();
	if (!Number.isInteger(i) || i < 0 || i >= count) indexOutOfRange(i, count);
	return state.players[i];
}

function requireStackProp(prop: PropDef | undefined, where: string): void {
	if (!prop || (prop.Type !== 'stack' && prop.Type !== 'sizedStack' && prop.Type !== 'mergedStack')) {
		throw new Error('go/api: ' + where + ' is not a stack property');
	}
}

/** Go: `game.X` where X is a stack property of the game state (merged stacks included). */
export function gameStack(state: RulesState, prop: string): StackRef {
	requireStackProp(state.engine.def.Game.Props[prop], 'game.' + prop);
	return { state, at: { Group: 'game', Prop: prop } };
}

/** Go: `p.X` where X is a stack property of player i's state (merged stacks included). */
export function playerStack(state: RulesState, i: number, prop: string): StackRef {
	player(state, i);
	requireStackProp(state.engine.def.Player.Props[prop], 'player.' + prop);
	return { state, at: { Group: 'player', Player: i, Prop: prop } };
}

/**
 * The engine's read-only view of a stack, for the queries the table maps to
 * the same name: `s.Len()`, `NumComponents()`, `SlotsRemaining()`,
 * `MaxSize()` become `stackView(s).len()`, and so on.
 */
export function stackView(s: StackRef): StackView {
	return viewOf(s);
}

// ----- reading stacks -----------------------------------------------------

/** Go: `s.ComponentAt(i)`, `s.ImmutableComponentAt(i)`. Null where Go returns nil. */
export function componentAt(s: StackRef, i: number): Comp | null {
	return compAt(s, viewOf(s), i);
}

/** Go: `s.First()`, `s.ImmutableFirst()`. Null for an empty stack. */
export function first(s: StackRef): Comp | null {
	const view = viewOf(s);
	if (isStack(view)) return compAt(s, view, view.firstComponentIndex());
	for (let i = 0; i < view.len(); i++) {
		const c = compAt(s, view, i);
		if (c !== null) return c;
	}
	return null;
}

/** Go: `s.Last()`, `s.ImmutableLast()`. Null for an empty stack. */
export function last(s: StackRef): Comp | null {
	const view = viewOf(s);
	if (isStack(view)) return compAt(s, view, view.lastComponentIndex());
	for (let i = view.len() - 1; i >= 0; i--) {
		const c = compAt(s, view, i);
		if (c !== null) return c;
	}
	return null;
}

/**
 * Go: `s.Components()`, `s.ImmutableComponents()`: a snapshot taken now, one
 * entry per slot, null for an empty slot of a sized stack.
 */
export function components(s: StackRef): (Comp | null)[] {
	const view = viewOf(s);
	const out: (Comp | null)[] = [];
	for (let i = 0; i < view.len(); i++) out.push(compAt(s, view, i));
	return out;
}

/**
 * Go: `s.ComponentAtKey(k)`, `s.ImmutableComponentAtKey(k)` on a stack
 * indexed by an enum: Go's is `ComponentAt(int(k))`.
 */
export function componentAtKey(s: StackRef, key: number): Comp | null {
	return componentAt(s, key);
}

/**
 * Go: `c.Values().(*T)`: the component's static values (the deck's generic
 * values for a generic component). Throws NIL_DEREF_TEXT when c is null.
 * A generic component of a deck with no generic values has nil values in
 * Go, and the type assertion panics; that throws here too, with its own text.
 */
export function values(state: RulesState, c: Comp | null): Readonly<Record<string, unknown>> {
	if (c === null) nilDeref();
	if (c.index === GENERIC_COMPONENT) {
		const generic = state.game.spec.decks[c.deck]?.generic ?? state.game.def.Decks[c.deck]?.GenericValues;
		if (generic === undefined || generic === null) {
			throw new Error('go/api: values: the generic component of deck ' + c.deck + ' has no values (Go: interface conversion: interface is nil)');
		}
		return generic;
	}
	return state.component(c.deck, c.index);
}

/** What Go's DynamicComponentValuesConstructor returns before inflation: scalars at zero, nothing else. */
function zeroDynamic(state: RulesState, deck: string): SubStateJSON {
	const def = state.engine.def.DynamicComponentValues[deck];
	if (!def) throw new Error('go/api: dynamic: deck ' + deck + ' has no dynamic component values');
	const out: SubStateJSON = {};
	for (const name of Object.keys(def.Props)) {
		switch (def.Props[name].Type) {
			case 'int':
			case 'playerIndex':
				out[name] = 0;
				break;
			case 'bool':
				out[name] = false;
				break;
			case 'string':
				out[name] = '';
				break;
			case 'intSlice':
			case 'boolSlice':
			case 'stringSlice':
			case 'playerIndexSlice':
				out[name] = null;
				break;
		}
	}
	return out;
}

/**
 * Go: `c.DynamicValues().(*T)`, or a dynamic values struct Go code already
 * holds: its properties, writable. Throws NIL_DEREF_TEXT when c is null.
 * For a Comp, the values are those of the state the component was read
 * from, as in Go. A generic component gets a fresh zero struct, as Go's
 * DynamicValues does (writes to it are lost); its enum and stack
 * properties are absent, Go's nil.
 */
export function dynamic(state: RulesState, c: Comp | DynamicRef | null): SubStateJSON {
	if (c === null) nilDeref();
	const owner = 'stack' in c ? c.stack.state : state;
	if (c.index === GENERIC_COMPONENT) return zeroDynamic(owner, c.deck);
	const values = owner.engine.dynamicValues(c.deck, c.index);
	if (values === undefined) throw new Error('go/api: dynamic: deck ' + c.deck + ' has no dynamic component values');
	return values;
}

/**
 * Go: `c.ContainingImmutableStack()`: where the component is now, as Go's
 * three results (stack, slot, error). The stack is null when err is not.
 * The answer is from the state the component was read from.
 */
export function containingStack(state: RulesState, c: Comp | null): [StackRef | null, number, string | null] {
	void state;
	if (c === null) nilDeref();
	const found = locate(c);
	if (typeof found === 'string') return [null, 0, found];
	return [refOf(c.stack.state, found.stack), found.slotIndex, null];
}

// ----- moving components --------------------------------------------------

/** Go: `c.MoveTo(dest, slot)`. Null on success, else Go's error text; nothing changes on failure. */
export function moveTo(c: Comp | null, dest: StackRef, slot: number): string | null {
	if (c === null) nilDeref();
	const to = requirePhysical(dest, 'moveTo');
	if (slot < 0) return 'Invalid slotIndex';
	const found = locate(c);
	if (typeof found === 'string') return NOT_IN_MUTABLE_STACK + found;
	return stackCall(() => found.stack.moveComponent(found.slotIndex, to, slot));
}

/** Go: `c.SecretMoveTo(dest, slot)`. */
export function secretMoveTo(c: Comp | null, dest: StackRef, slot: number): string | null {
	if (c === null) nilDeref();
	const to = requirePhysical(dest, 'secretMoveTo');
	if (slot < 0) return 'Invalid slotIndex';
	const found = locate(c);
	if (typeof found === 'string') return NOT_IN_MUTABLE_STACK + found;
	return stackCall(() => found.stack.secretMoveComponent(found.slotIndex, to, slot));
}

function moveToEnd(c: Comp | null, dest: StackRef, fn: string, kind: 'first' | 'last' | 'next'): string | null {
	if (c === null) nilDeref();
	const to = requirePhysical(dest, fn);
	const found = locate(c);
	if (typeof found === 'string') return NOT_IN_MUTABLE_STACK + found;
	return stackCall(() => {
		if (kind === 'first') found.stack.moveComponentToFirstSlot(found.slotIndex, to);
		else if (kind === 'last') found.stack.moveComponentToLastSlot(found.slotIndex, to);
		else found.stack.moveComponentToNextSlot(found.slotIndex, to);
	});
}

/** Go: `c.MoveToFirstSlot(dest)`. */
export function moveToFirstSlot(c: Comp | null, dest: StackRef): string | null {
	return moveToEnd(c, dest, 'moveToFirstSlot', 'first');
}

/** Go: `c.MoveToLastSlot(dest)`. */
export function moveToLastSlot(c: Comp | null, dest: StackRef): string | null {
	return moveToEnd(c, dest, 'moveToLastSlot', 'last');
}

/** Go: `c.MoveToNextSlot(dest)`. */
export function moveToNextSlot(c: Comp | null, dest: StackRef): string | null {
	return moveToEnd(c, dest, 'moveToNextSlot', 'next');
}

function slide(c: Comp | null, toStart: boolean): string | null {
	if (c === null) nilDeref();
	const found = locate(c);
	if (typeof found === 'string') return NOT_IN_MUTABLE_STACK + found;
	return stackCall(() => (toStart ? found.stack.slideToFirstSlot(found.slotIndex) : found.stack.slideToLastSlot(found.slotIndex)));
}

/** Go: `c.SlideToFirstSlot()`. */
export function slideToFirstSlot(c: Comp | null): string | null {
	return slide(c, true);
}

/** Go: `c.SlideToLastSlot()`. */
export function slideToLastSlot(c: Comp | null): string | null {
	return slide(c, false);
}

/** Go: `c.MayMoveTo(dest)`: null when the move would succeed, else why not. */
export function mayMoveTo(c: Comp | null, dest: StackRef): string | null {
	if (c === null) nilDeref();
	// componentInstance.MayMoveTo's checks, in its order. Those after the
	// source is found are the engine's (and so pick up its constraints).
	if (engineOf(dest) !== engineOf(c.stack)) return 'source and destination belong to different states';
	const to = physicalOf(dest);
	if (to === null) return 'destination stack is not attached: stack is not a physical stack';
	const attachment = engineOf(dest).stackAttachmentError(to);
	if (attachment !== null) return 'destination stack is not attached: ' + attachment;
	const found = locate(c);
	if (typeof found === 'string') return 'component is not in any stack: ' + found;
	return found.stack.mayMoveToError(c.index, to);
}

/** Go: `c.MayMoveToSlot(dest, slot)`. */
export function mayMoveToSlot(c: Comp | null, dest: StackRef, slot: number): string | null {
	const err = mayMoveTo(c, dest);
	if (err !== null) return err;
	if (slot < 0) return 'slot index must be non-negative';
	const view = viewOf(dest);
	const outOfRange = 'slot index ' + itoa(slot) + ' is out of range (stack length ' + itoa(view.len()) + ')';
	if (view.isSized()) {
		if (slot >= view.len()) return outOfRange;
		if (view.componentAt(slot) !== null) return 'slot ' + itoa(slot) + ' is already occupied';
		return null;
	}
	if (slot > view.len()) return outOfRange;
	return null;
}

/** Go: `s.MoveAllTo(dest)`. */
export function moveAllTo(s: StackRef, dest: StackRef): string | null {
	const from = requirePhysical(s, 'moveAllTo');
	const to = requirePhysical(dest, 'moveAllTo');
	return stackCall(() => from.moveAllTo(to));
}

/** Go: `s.MoveCountTo(dest, n)`. */
export function moveCountTo(s: StackRef, dest: StackRef, n: number): string | null {
	const from = requirePhysical(s, 'moveCountTo');
	const to = requirePhysical(dest, 'moveCountTo');
	return stackCall(() => from.moveCountTo(to, n));
}

/** Go: `s.MayMoveAllTo(dest)`. */
export function mayMoveAllTo(s: StackRef, dest: StackRef): string | null {
	const from = physicalOf(s);
	if (from === null) return 'MayMoveAllTo is not supported on MergedStacks';
	const to = physicalOf(dest);
	if (to === null) return 'destination is not a physical stack';
	return from.mayMoveAllTo(to);
}

/** Go: `s.MayMoveCountTo(dest, n)`. */
export function mayMoveCountTo(s: StackRef, dest: StackRef, n: number): string | null {
	const from = physicalOf(s);
	if (from === null) return 'MayMoveCountTo is not supported on MergedStacks';
	const to = physicalOf(dest);
	if (to === null) return 'destination is not a physical stack';
	return from.mayMoveCountTo(to, n);
}

// ----- rearranging and resizing stacks ------------------------------------

/** Go: `s.Shuffle()`. Draws from state.rand() as Go does. */
export function shuffle(s: StackRef): string | null {
	const stack = requirePhysical(s, 'shuffle');
	return stackCall(() => stack.shuffle());
}

/** Go: `s.PublicShuffle()`. */
export function publicShuffle(s: StackRef): string | null {
	const stack = requirePhysical(s, 'publicShuffle');
	return stackCall(() => stack.publicShuffle());
}

/** Go: `s.SwapComponents(i, j)`. */
export function swapComponents(s: StackRef, i: number, j: number): string | null {
	const stack = requirePhysical(s, 'swapComponents');
	return stackCall(() => stack.swapComponents(i, j));
}

/** Go: `s.SwapComponentsByKey(i, j)` on a stack indexed by an enum: Go's is `SwapComponents(int(i), int(j))`. */
export function swapComponentsByKey(s: StackRef, i: number, j: number): string | null {
	return swapComponents(s, i, j);
}

/**
 * Go: `s.SortComponents(less)`, with Go's sort.Sort order among equal
 * elements (gosort.ts). `less` gets components, null for an empty slot;
 * their `slot` is -1. A `less` that throws (a Go panic) leaves the stack
 * as it was, as Go's sort of a copy does.
 */
export function sortComponents(s: StackRef, less: (a: Comp | null, b: Comp | null) => boolean): string | null {
	const stack = requirePhysical(s, 'sortComponents');
	const comp = (index: number): Comp | null => (index === EMPTY_SLOT ? null : { stack: s, slot: -1, deck: stack.deckName, index });
	return stackCall(() => stack.sortComponents((a, b) => less(comp(a), comp(b))));
}

/** Go: `s.SetSize(n)` on a sized stack. */
export function setSize(s: StackRef, n: number): string | null {
	const stack = requirePhysical(s, 'setSize');
	return stackCall(() => stack.setSize(n));
}

/** Go: `s.ExpandSize(n)` on a sized stack. */
export function expandSize(s: StackRef, n: number): string | null {
	const stack = requirePhysical(s, 'expandSize');
	return stackCall(() => stack.expandSize(n));
}

/** Go: `s.ContractSize(n)` on a sized stack. */
export function contractSize(s: StackRef, n: number): string | null {
	const stack = requirePhysical(s, 'contractSize');
	return stackCall(() => stack.contractSize(n));
}

/** Go: `s.SizeToFit()` on a sized stack. */
export function sizeToFit(s: StackRef): string | null {
	const stack = requirePhysical(s, 'sizeToFit');
	return stackCall(() => stack.sizeToFit());
}

// ----- game, players, behaviors -------------------------------------------

/** Go: `state.Game().Variant()[k]`. A missing key gives ''. */
export function variantValue(state: RulesState, key: string): string {
	const variant = state.variant();
	return Object.prototype.hasOwnProperty.call(variant, key) ? variant[key] : '';
}

/** Go: `p.SetEliminated()` (behaviors.PlayerElimination). */
export function setEliminated(state: RulesState, i: number): void {
	player(state, i).Eliminated = true;
}

/** Go: `p.SetPlayerInactive()` (behaviors.InactivePlayer). */
export function setPlayerInactive(state: RulesState, i: number): void {
	player(state, i).PlayerInactive = true;
}

/** Go: `p.SetSeatFilled()` (behaviors.Seat): a filled seat is also closed. */
export function setSeatFilled(state: RulesState, i: number): void {
	const p = player(state, i);
	p.SeatFilled = true;
	p.SeatClosed = true;
}

// ----- enums --------------------------------------------------------------

function isDynamicOwner(owner: Owner): owner is { readonly dynamicOf: Comp | DynamicRef | null } {
	return typeof owner === 'object' && 'dynamicOf' in owner;
}

function isValuesOwner(owner: Owner): owner is { readonly valuesOf: Comp | null } {
	return typeof owner === 'object' && 'valuesOf' in owner;
}

/** The JSON object an owner's properties live in. */
function ownerNode(state: RulesState, owner: Owner): Readonly<Record<string, unknown>> {
	if (owner === 'game') return state.g;
	if (typeof owner === 'number') return player(state, owner);
	if (isDynamicOwner(owner)) return dynamic(state, owner.dynamicOf);
	return values(state, owner.valuesOf);
}

/**
 * The bound enum of an owner's enum property. A nil component panics first,
 * as Go's `c.DynamicValues()` does before anything reads the property.
 */
function enumOf(state: RulesState, owner: Owner, prop: string): BoundEnum {
	ownerNode(state, owner);
	const def = state.game.def;
	let name: string | undefined;
	if (owner === 'game') name = def.Game.Props[prop]?.Enum;
	else if (typeof owner === 'number') name = def.Player.Props[prop]?.Enum;
	else if (isDynamicOwner(owner)) name = def.DynamicComponentValues[owner.dynamicOf!.deck]?.Props[prop]?.Enum;
	else {
		const deck = owner.valuesOf!.deck;
		name = state.game.components[deck]?.[prop]?.Enum ?? def.Decks[deck]?.ComponentProps?.[prop]?.Enum;
	}
	const bound = state.game.enums.get(name);
	if (bound === null) throw new Error('go/api: ' + prop + ' is not an enum property');
	return bound;
}

/** The stored string of an enum property; Go's nil enum.Val panics on use. */
function storedEnum(state: RulesState, owner: Owner, prop: string): string {
	const value = ownerNode(state, owner)[prop];
	if (value === undefined || value === null) nilDeref();
	return value as string;
}

/** Go: `v.Value()` on an enum property: its key, converted from the stored string. */
export function enumKey(state: RulesState, owner: Owner, prop: string): number {
	const e = enumOf(state, owner, prop);
	const stored = storedEnum(state, owner, prop);
	const key = e.keyOf(stored);
	if (key === null) throw new Error('go/api: ' + prop + ' holds ' + JSON.stringify(stored) + ', which is not a value of enum ' + e.name);
	return key;
}

/** Go: `v.SetValue(k)` on an enum property: stores k's string. Null, or Go's error for a key the enum lacks. */
export function setEnumKey(state: RulesState, owner: Owner, prop: string, key: number): string | null {
	if (isValuesOwner(owner)) throw new Error('go/api: setEnumKey: component values are immutable');
	const e = enumOf(state, owner, prop);
	storedEnum(state, owner, prop);
	if (!Object.prototype.hasOwnProperty.call(e.def.Values, String(key))) return 'That value is not valid for this enum';
	(ownerNode(state, owner) as SubStateJSON)[prop] = e.stringOf(key);
	return null;
}

/** Go: `v.Equals(o)` on two enum properties: the same enum and the same key. */
export function enumEquals(state: RulesState, owner: Owner, prop: string, otherOwner: Owner, otherProp: string): boolean {
	const e = enumOf(state, owner, prop);
	const other = enumOf(state, otherOwner, otherProp);
	return e.name === other.name && enumKey(state, owner, prop) === enumKey(state, otherOwner, otherProp);
}

/** enum's indexesForKey: a range value's string split at `_`, or null if a part is not an int. */
function indexesForKey(key: string): number[] | null {
	const out: number[] = [];
	for (const part of key.split('_')) {
		const [value, err] = atoi(part);
		if (err !== null) return null;
		out.push(value);
	}
	return out;
}

/**
 * Go: `v.RangeValue()` on a range-enum property: its indexes, one per
 * dimension, parsed from the stored string. Null where Go returns nil.
 */
export function rangeValue(state: RulesState, owner: Owner, prop: string): number[] | null {
	enumOf(state, owner, prop);
	return indexesForKey(storedEnum(state, owner, prop));
}

function requireEnum(state: RulesState, enumName: string): BoundEnum {
	const e = state.game.enums.get(enumName);
	if (e === null) throw new Error('go/api: no enum named ' + enumName);
	return e;
}

/**
 * Go: `e.RangeToValue(indexes...)`: the key whose string is the indexes
 * joined with `_` (as `e.ValueFromString` finds it), or ILLEGAL_ENUM_VALUE.
 */
export function rangeToValue(state: RulesState, enumName: string, indexes: readonly number[]): number {
	const key = requireEnum(state, enumName).valueFromString(indexes.map(itoa).join('_'));
	return key === null ? ILLEGAL_ENUM_VALUE : key;
}

/** Go: `e.ValueToRange(k)`: the indexes of key k. Null where Go returns nil (a key the enum lacks). */
export function valueToRange(state: RulesState, enumName: string, key: number): number[] | null {
	return indexesForKey(requireEnum(state, enumName).stringOf(key));
}

/** One of graph's Direction* edge filters, by its Go name. */
function directionPasses(direction: string, from: readonly number[], to: readonly number[]): boolean {
	switch (direction) {
		case 'DirectionUp':
			return from[0] > to[0];
		case 'DirectionDown':
			return from[0] < to[0];
		case 'DirectionLeft':
			return from[1] > to[1];
		case 'DirectionRight':
			return from[1] < to[1];
		case 'DirectionPerpendicular':
			return from[0] === to[0] || from[1] === to[1];
		case 'DirectionDiagonal':
			return !(from[0] === to[0] || from[1] === to[1]);
	}
	throw new Error('go/api: neighborsByKey: unknown direction ' + direction);
}

/**
 * Go: `g.NeighborsByKey(k)` on a graph from MustNewEnumGridConnectedness:
 * the keys of the (up to eight) adjacent cells that pass every direction
 * filter, or null where Go returns nil (no neighbors, or a key the enum
 * lacks). Go returns them in a map's iteration order, which differs run to
 * run; they are returned ascending here, and Go code whose result depends on
 * the order has no single answer to match.
 */
export function neighborsByKey(graph: GraphRef, key: number): number[] | null {
	for (const direction of graph.directions) directionPasses(direction, [0, 0], [0, 0]);
	if (graph.dimensions.length !== 2) throw new Error('The enum did not have two dimensions');
	const [rows, cols] = graph.dimensions;
	if (!Number.isInteger(key) || key < 0 || key >= rows * cols) return null;
	const from = [Math.floor(key / cols), key % cols];
	const out: number[] = [];
	for (let dr = -1; dr < 2; dr++) {
		for (let dc = -1; dc < 2; dc++) {
			if (dr === 0 && dc === 0) continue;
			const to = [from[0] + dr, from[1] + dc];
			if (to[0] < 0 || to[0] >= rows || to[1] < 0 || to[1] >= cols) continue;
			if (graph.directions.every(direction => directionPasses(direction, from, to))) out.push(to[0] * cols + to[1]);
		}
	}
	if (out.length === 0) return null;
	return out.sort((a, b) => a - b);
}

// ----- moves and legality -------------------------------------------------

/**
 * Go: `m.<Stock>.Apply(state)` inside a move's own Apply: the embedded
 * stock move's apply, which the loader resolves from MoveData.Stock.
 */
export function stockApply(state: RulesState, m: MoveInstance): string | null {
	void state;
	void m;
	throw new Error(
		'engine-gap: stockApply needs the loader (rules/data.ts gameFromData with rules/stockRegistry.ts stockMove) to resolve ' +
			"MoveData.Stock into the move's StockParts and make its apply reachable from a MoveInstance",
	);
}

/** Go: `legal.Errorf(key, bindings)`: a failing verdict with that template and bindings. */
export function legalError(key: string, bindings?: Readonly<Record<string, LegalBindingValue>>): LegalVerdict {
	return bindings === undefined ? failT(key) : failT(key, { ...bindings });
}
