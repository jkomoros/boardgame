// Sanitization: what one viewer may see of a state. Ports sanitization.go,
// state_visibility.go and the base delegate's group logic.
//
// Five policies, least restrictive first: visible, order, len, nonempty,
// hidden. Each property names a policy per group ("all", "self", "other",
// a value of the game's group enum, or a computed group such as
// "same-color"). The viewer's groups select the applicable policies and the
// LEAST restrictive one wins.

import { EMPTY_SLOT, GENERIC_COMPONENT } from './ids.ts';
import { EngineState, StateError, compareGoStrings, sortedKeys } from './state.ts';
import type { BoardJSON, StateJSON, SubStateJSON, SubStateRef, VisibilityJSON } from './state.ts';
import { indexesOf } from './stack.ts';
import type { StackJSON } from './stack.ts';
import { GROUP_ALL, GROUP_OTHER, GROUP_SELF, POLICY_ORDER, policyMapFor } from './definition.ts';
import type { EnumDef, GameDefinition, PolicyName, PropDef, SubStateDef } from './definition.ts';

/** state.go ObserverPlayerIndex: sees only what everyone may see. */
export const OBSERVER_PLAYER_INDEX = -1;
/** state.go AdminPlayerIndex: sees the unsanitized state. */
export const ADMIN_PLAYER_INDEX = -2;
/** state.go AnyPlayerIndex: sanitized as the observer. */
export const ANY_PLAYER_INDEX = -3;

export class SanitizeError extends Error {
	constructor(message: string) {
		super(message);
		this.name = 'SanitizeError';
	}
}

/** The effective policy of every property for one viewer. */
export interface SanitizationTransformation {
	Game: Record<string, PolicyName>;
	Players: Record<string, PolicyName>[];
	DynamicComponentValues: Record<string, Record<string, PolicyName>>;
}

export interface SanitizeOptions {
	/**
	 * How a VISIBLE merged stack over non-visible stacks treats the dynamic
	 * component values of the components it spans.
	 *
	 * Go has no single answer: sanitizeStateObj ranges over a map, so a
	 * visible merged stack is examined either before its parts are sanitized
	 * (it then reveals the dynamic values of every component it spans) or
	 * after (it then reveals nothing, or panics on the generic component).
	 *
	 * false (the default) never reveals through a merged stack: the result
	 * Go gives when the merged stack happens to be examined last, without
	 * the panic. true reveals: the result when it is examined first.
	 */
	mergedStacksRevealDynamicValues?: boolean;
}

/**
 * ResolveSanitizationPolicy: the least restrictive policy among the groups
 * the viewer is in, or fallback when none of them names one.
 */
export function resolveSanitizationPolicy(
	policyMap: Readonly<Record<string, PolicyName>>,
	groupMembership: Readonly<Record<string, boolean>>,
	fallback: PolicyName = 'visible',
): PolicyName {
	let best = -1;
	for (const groupName of Object.keys(groupMembership)) {
		if (!groupMembership[groupName]) continue;
		if (!Object.prototype.hasOwnProperty.call(policyMap, groupName)) continue;
		const rank = POLICY_ORDER.indexOf(policyMap[groupName]);
		if (best === -1 || rank < best) best = rank;
	}
	return best === -1 ? fallback : POLICY_ORDER[best];
}

// ----- groups -------------------------------------------------------------

function groupEnumOf(def: GameDefinition): EnumDef | null {
	if (!def.GroupEnum) return null;
	return def.Enums[def.GroupEnum] ?? null;
}

/** enum.SubsetOf: every key of a is a key of b with the same string. */
function enumSubsetOf(a: EnumDef, b: EnumDef): boolean {
	for (const key of Object.keys(a.Values)) {
		if (!Object.prototype.hasOwnProperty.call(b.Values, key)) return false;
		if (b.Values[key] !== a.Values[key]) return false;
	}
	return true;
}

function enumKeyOf(e: EnumDef, text: unknown): string | null {
	for (const key of Object.keys(e.Values)) if (e.Values[key] === text) return key;
	return null;
}

/**
 * base.GameDelegate.GroupMembership: a player belongs to the group named by
 * each of their enum properties whose enum is a subset of the group enum.
 * Returns group-enum keys (decimal strings), or null for no membership.
 */
export function playerGroupKeys(def: GameDefinition, player: SubStateJSON | null): string[] | null {
	const groupEnum = groupEnumOf(def);
	if (player === null || groupEnum === null) return null;
	const members: string[] = [];
	let any = false;
	for (const propName of Object.keys(def.Player.Props)) {
		const prop = def.Player.Props[propName];
		if (prop.Type !== 'enum' || !prop.Enum) continue;
		const e = def.Enums[prop.Enum];
		if (!e || !enumSubsetOf(e, groupEnum)) continue;
		any = true;
		const key = enumKeyOf(e, player[propName]);
		if (key !== null && members.indexOf(key) < 0) members.push(key);
	}
	return any ? members : null;
}

/** groupMembershipForPlayerState: the named groups, plus "all". */
export function stringMembership(def: GameDefinition, keys: string[] | null): Record<string, boolean> {
	const result: Record<string, boolean> = {};
	const groupEnum = groupEnumOf(def);
	if (keys !== null && groupEnum !== null) {
		for (const key of keys) result[groupEnum.Values[key]] = true;
	}
	result[GROUP_ALL] = true;
	return result;
}

/**
 * StructInflater.sanitizationPolicyGroupNames for the player state: every
 * group a player property names, except "all" and group-enum values. These
 * are the groups whose membership must be computed per player.
 */
function computedPlayerGroupNames(def: GameDefinition): string[] {
	const groupEnum = groupEnumOf(def);
	const names: string[] = [];
	for (const propName of Object.keys(def.Player.Props)) {
		const policyMap = policyMapFor(def.Player.Props[propName], GROUP_OTHER);
		for (const key of Object.keys(policyMap)) {
			if (groupEnum !== null && enumKeyOf(groupEnum, key) !== null) continue;
			if (key === GROUP_ALL) continue;
			if (names.indexOf(key) < 0) names.push(key);
		}
	}
	return names;
}

/** GameManager.computedPlayerGroupMembership + the base delegate's. */
export function computedGroupMembership(
	def: GameDefinition,
	groupName: string,
	player: number,
	viewer: number,
	playerKeys: string[] | null,
	viewerKeys: string[] | null,
): boolean {
	if (groupName === GROUP_SELF) return player === viewer;
	if (groupName === GROUP_OTHER) return player !== viewer;

	const fail = (reason: string): never => {
		throw new SanitizeError("Couldn't get group membership for groupName: " + groupName + ': ' + reason);
	};
	const parts = groupName.split('-');
	if (parts.length !== 2) return fail('Unsupported group name: ' + groupName);
	const fun = parts[0].toLowerCase();
	if (fun !== 'same' && fun !== 'different') {
		return fail(parts[0] + " was used as a computed group name function but it's not a known one");
	}
	const e = def.Enums[parts[1]];
	if (!e) return fail(parts[1] + " was used as a computed group name enum but it's not a legal enum");
	const groupEnum = groupEnumOf(def);
	if (groupEnum === null) return fail('A computed group name used an enum, but there is no Group enum');
	if (!enumSubsetOf(e, groupEnum)) return fail(parts[1] + ' enum is not a subset of GroupEnum');

	for (const key of Object.keys(e.Values)) {
		const p = playerKeys !== null && playerKeys.indexOf(key) >= 0;
		const v = viewerKeys !== null && viewerKeys.indexOf(key) >= 0;
		if (p !== v) return fun === 'different';
	}
	return fun === 'same';
}

// ----- transformation -----------------------------------------------------

function subStateTransformation(
	sub: SubStateDef,
	defaultGroup: string,
	membership: Readonly<Record<string, boolean>>,
): Record<string, PolicyName> {
	const result: Record<string, PolicyName> = {};
	for (const propName of Object.keys(sub.Props)) {
		result[propName] = resolveSanitizationPolicy(policyMapFor(sub.Props[propName], defaultGroup), membership, 'visible');
	}
	return result;
}

/**
 * state.generateSanitizationTransformation: the policy each property gets
 * when the state is shown to `viewer` (a player index, or the observer).
 */
export function sanitizationTransformation(state: EngineState, viewer: number): SanitizationTransformation {
	const def = state.def;
	const players = state.data.Players;
	const viewerKeys = viewer >= 0 ? playerGroupKeys(def, players[viewer]) : null;
	const viewerMembership = stringMembership(def, viewerKeys);

	const result: SanitizationTransformation = {
		Game: subStateTransformation(def.Game, GROUP_ALL, viewerMembership),
		Players: [],
		DynamicComponentValues: {},
	};

	const computedNames = computedPlayerGroupNames(def);
	for (let i = 0; i < players.length; i++) {
		// A player state is judged by the groups of the player it belongs
		// to, plus the groups computed between that player and the viewer.
		const playerKeys = playerGroupKeys(def, players[i]);
		const membership = stringMembership(def, playerKeys);
		for (const groupName of computedNames) {
			membership[groupName] = computedGroupMembership(def, groupName, i, viewer, playerKeys, viewerKeys);
		}
		result.Players.push(subStateTransformation(def.Player, GROUP_OTHER, membership));
	}

	for (const deckName of Object.keys(state.data.Components)) {
		if (state.data.Components[deckName].length === 0) throw new SanitizeError('No deck values');
		const sub = def.DynamicComponentValues[deckName];
		if (!sub) throw new SanitizeError('no dynamic component values are defined for deck ' + deckName);
		result.DynamicComponentValues[deckName] = subStateTransformation(sub, GROUP_ALL, viewerMembership);
	}
	return result;
}

// ----- facets -------------------------------------------------------------

type Facet = 'values' | 'count' | 'occupancy' | 'order' | 'nonempty';
const FACETS: readonly Facet[] = ['values', 'count', 'occupancy', 'order', 'nonempty'];

/** legal_path.go facetSurvives. */
export function facetSurvives(policy: PolicyName, facet: Facet): boolean {
	switch (facet) {
		case 'values':
			return policy === 'visible';
		case 'count':
			return policy === 'visible' || policy === 'order' || policy === 'len';
		case 'occupancy':
		case 'order':
			return policy === 'visible' || policy === 'order';
		case 'nonempty':
			return policy === 'visible' || policy === 'order' || policy === 'len' || policy === 'nonempty';
	}
}

function propertyVisibility(sub: SubStateDef, policies: Readonly<Record<string, PolicyName>> | null): Record<string, string[]> {
	const result: Record<string, string[]> = {};
	for (const name of Object.keys(sub.Props)) {
		const policy = policies !== null && Object.prototype.hasOwnProperty.call(policies, name) ? policies[name] : 'visible';
		result[name] = FACETS.filter(facet => facetSurvives(policy, facet));
	}
	return result;
}

/**
 * state.visibleFacets for an unsanitized state: which facets of each game
 * and player property survive. A null transformation means nothing is
 * sanitized (the admin view). Dynamic component values have no entry, as in
 * Go.
 */
export function visibleFacets(state: EngineState, transformation: SanitizationTransformation | null): VisibilityJSON {
	const players: Record<string, string[]>[] = [];
	for (let i = 0; i < state.data.Players.length; i++) {
		players.push(propertyVisibility(state.def.Player, transformation === null ? null : transformation.Players[i]));
	}
	return {
		Game: propertyVisibility(state.def.Game, transformation === null ? null : transformation.Game),
		Players: players,
	};
}

// ----- applying a policy --------------------------------------------------

function sliceLength(value: unknown): number {
	return Array.isArray(value) ? value.length : 0;
}

function sanitizedSlice(policy: PolicyName, value: unknown, zero: unknown): unknown[] {
	let length = 0;
	if (policy === 'len' || policy === 'order') length = sliceLength(value);
	else if (policy === 'nonempty') length = sliceLength(value) > 0 ? 1 : 0;
	const out = new Array<unknown>(length);
	for (let i = 0; i < length; i++) out[i] = zero;
	return out;
}

function enumDefault(def: GameDefinition, prop: PropDef, path: string): string {
	const e = prop.Enum ? def.Enums[prop.Enum] : undefined;
	if (!e) throw new SanitizeError(path + ' names an enum the definition does not have');
	return e.Values[String(e.DefaultValue)];
}

/** growableStack / sizedStack applySanitizationPolicy, on the JSON in place. */
function sanitizeStack(node: StackJSON, sized: boolean, policy: PolicyName): void {
	if (policy === 'visible') return;
	const indexes = indexesOf(node);

	if (policy === 'len' || policy === 'order') {
		// Ids are kept (the stack's already hold each component's id, ""
		// for an empty slot). Under len they are sorted, which in a sized
		// stack moves every "" to the front while Indexes keeps the real
		// occupancy: the two no longer line up, exactly as in Go.
		const ids = (node.IDs ?? []).slice();
		if (policy === 'len') ids.sort(compareGoStrings);
		node.IDs = ids;
		const next = new Array<number>(indexes.length);
		for (let i = 0; i < next.length; i++) {
			next[i] = sized && indexes[i] === EMPTY_SLOT ? EMPTY_SLOT : GENERIC_COMPONENT;
		}
		node.Indexes = next;
		return;
	}

	node.ShuffleCount = 0;
	node.IDsLastSeen = {};

	if (!sized) {
		if (policy === 'nonempty' && indexes.length > 0) {
			node.Indexes = [GENERIC_COMPONENT];
			node.IDs = [''];
		} else {
			node.Indexes = [];
			node.IDs = [];
		}
		return;
	}

	let hasComponents = false;
	for (const index of indexes) if (index !== EMPTY_SLOT) hasComponents = true;
	const next = new Array<number>(indexes.length);
	const ids = new Array<string>(indexes.length);
	for (let i = 0; i < next.length; i++) {
		next[i] = EMPTY_SLOT;
		ids[i] = '';
	}
	if (policy === 'nonempty' && hasComponents) next[0] = GENERIC_COMPONENT;
	node.Indexes = next;
	node.IDs = ids;
}

/** applyPolicy for one property of one sub-state, on the JSON in place. */
function applyPolicy(def: GameDefinition, node: SubStateJSON, propName: string, prop: PropDef, policy: PolicyName, path: string): void {
	if (policy === 'invalid') throw new SanitizeError('Effective policy computed to PolicyInvalid');
	if (policy === 'visible') return;
	switch (prop.Type) {
		case 'bool':
			node[propName] = false;
			return;
		case 'int':
		case 'playerIndex':
			node[propName] = 0;
			return;
		case 'string':
			node[propName] = '';
			return;
		case 'timer':
			node[propName] = { ID: '', IsTimer: true };
			return;
		case 'enum':
			node[propName] = enumDefault(def, prop, path);
			return;
		case 'enumSlice':
			node[propName] = sanitizedSlice(policy, node[propName], enumDefault(def, prop, path));
			return;
		case 'intSlice':
		case 'playerIndexSlice':
			node[propName] = sanitizedSlice(policy, node[propName], 0);
			return;
		case 'boolSlice':
			node[propName] = sanitizedSlice(policy, node[propName], false);
			return;
		case 'stringSlice':
			node[propName] = sanitizedSlice(policy, node[propName], '');
			return;
		case 'board':
			for (const space of (node[propName] as BoardJSON).Spaces) sanitizeStack(space, false, policy);
			return;
		case 'stack':
			sanitizeStack(node[propName] as StackJSON, false, policy);
			return;
		case 'sizedStack':
			sanitizeStack(node[propName] as StackJSON, true, policy);
			return;
		case 'mergedStack':
			// applyPolicy asserts the property to a mutable Stack, which a
			// merged stack is not: Go panics. A merged stack is a view; hide
			// the stacks it is made of instead.
			throw new SanitizeError(
				path + ' is a merged stack and cannot carry policy "' + policy + '" (Go panics here); sanitize its parts',
			);
	}
}

type VisibleSet = Record<string, Record<number, boolean>>;

function markStack(visible: VisibleSet, deckName: string, indexes: readonly number[], deckSize: number): number[] {
	const marked: number[] = [];
	const deckVisible = visible[deckName];
	if (!deckVisible) return marked;
	for (const index of indexes) {
		if (index < 0 || index >= deckSize) continue;
		if (!deckVisible[index]) marked.push(index);
		deckVisible[index] = true;
	}
	return marked;
}

/** Stacks of a sub-state as (node) list: stack props and board spaces. */
function physicalStackNodes(sub: SubStateDef, node: SubStateJSON): StackJSON[] {
	const out: StackJSON[] = [];
	for (const propName of sortedKeys(sub.Props)) {
		const type = sub.Props[propName].Type;
		if (type === 'stack' || type === 'sizedStack') out.push(node[propName] as StackJSON);
		else if (type === 'board') for (const space of (node[propName] as BoardJSON).Spaces) out.push(space);
	}
	return out;
}

function sanitizeSubState(
	def: GameDefinition,
	sub: SubStateDef,
	node: SubStateJSON,
	policies: Readonly<Record<string, PolicyName>>,
	visible: VisibleSet | null,
	path: string,
	/** The slots a merged property showed before anything was sanitized. */
	mergedIndexesBefore: ((propName: string) => { deck: string; indexes: number[] }) | null,
): void {
	const deckSize = (name: string) => (def.Decks[name] ? def.Decks[name].Size : 0);
	for (const propName of sortedKeys(sub.Props)) {
		const prop = sub.Props[propName];
		const policy = Object.prototype.hasOwnProperty.call(policies, propName) ? policies[propName] : 'visible';
		applyPolicy(def, node, propName, prop, policy, path + '.' + propName);
		if (visible === null || policy !== 'visible') continue;

		// A visible stack makes the dynamic values of its components visible.
		if (prop.Type === 'stack' || prop.Type === 'sizedStack') {
			const stack = node[propName] as StackJSON;
			markStack(visible, stack.Deck, indexesOf(stack), deckSize(stack.Deck));
		} else if (prop.Type === 'board') {
			for (const space of (node[propName] as BoardJSON).Spaces) {
				markStack(visible, space.Deck, indexesOf(space), deckSize(space.Deck));
			}
		} else if (prop.Type === 'mergedStack' && mergedIndexesBefore !== null) {
			// See SanitizeOptions: reproduce Go examining the merged stack
			// before any of its parts was sanitized.
			const before = mergedIndexesBefore(propName);
			markStack(visible, before.deck, before.indexes, deckSize(before.deck));
		}
	}
}

/**
 * state.SanitizedForPlayer: the state as `viewer` may see it.
 *
 * `viewer` is a player index, OBSERVER_PLAYER_INDEX, or ANY_PLAYER_INDEX
 * (treated as the observer). ADMIN_PLAYER_INDEX, and any index at or past
 * the number of players, get the state itself, unsanitized, as in Go.
 */
export function sanitizedForPlayer(state: EngineState, viewer: number, options: SanitizeOptions = {}): EngineState {
	if (state.sanitized) throw new SanitizeError('the state is already sanitized');
	if (viewer === ANY_PLAYER_INDEX) viewer = OBSERVER_PLAYER_INDEX;
	if (viewer < -1 || viewer >= state.numPlayers()) return state;

	const transformation = sanitizationTransformation(state, viewer);
	const def = state.def;
	const sanitized = state.copySanitized();
	const data = sanitized.data;

	const visible: VisibleSet = {};
	for (const deckName of Object.keys(data.Components)) visible[deckName] = {};

	const extend = (err: unknown, message: string): never => {
		if (err instanceof SanitizeError) throw new SanitizeError(message + ' : ' + err.message);
		throw err;
	};

	const reveal = options.mergedStacksRevealDynamicValues === true;
	const before = (ref: SubStateRef) =>
		reveal
			? (propName: string) => {
					const view = state.view({ ...ref, Prop: propName });
					return { deck: view.deckName, indexes: view.indexes() };
				}
			: null;

	try {
		sanitizeSubState(def, def.Game, data.Game, transformation.Game, visible, 'Game', before({ Group: 'game' }));
	} catch (err) {
		extend(err, "Couldn't sanitize game state");
	}
	for (let i = 0; i < data.Players.length; i++) {
		try {
			sanitizeSubState(def, def.Player, data.Players[i], transformation.Players[i], visible, 'Players[' + i + ']', before({ Group: 'player', Player: i }));
		} catch (err) {
			extend(err, "Couldn't sanitize player state number " + i);
		}
	}

	// transativelyMarkDynamicComponentsAsVisible: a visible component's own
	// stacks make the components in them visible too, and so on. (Go appends
	// to its work list without checking for repeats, so two components that
	// hold each other would loop forever there. Here each is visited once.)
	const work: { deck: string; index: number }[] = [];
	for (const deckName of sortedKeys(visible)) {
		for (const key of Object.keys(visible[deckName])) work.push({ deck: deckName, index: Number(key) });
	}
	for (let i = 0; i < work.length; i++) {
		const item = work[i];
		const sub = def.DynamicComponentValues[item.deck];
		const values = data.Components[item.deck][item.index];
		if (!sub || !values) continue;
		for (const stack of physicalStackNodes(sub, values)) {
			const size = def.Decks[stack.Deck] ? def.Decks[stack.Deck].Size : 0;
			for (const index of markStack(visible, stack.Deck, indexesOf(stack), size)) work.push({ deck: stack.Deck, index });
		}
	}

	// sanitizeDynamicComponentValues: a visible component's values get their
	// configured policies; everything else is hidden outright.
	try {
		for (const deckName of sortedKeys(data.Components)) {
			const sub = def.DynamicComponentValues[deckName];
			const values = data.Components[deckName];
			const hiddenAll: Record<string, PolicyName> = {};
			for (const propName of Object.keys(sub.Props)) {
				// Merged stacks are views and follow their parts.
				hiddenAll[propName] = sub.Props[propName].Type === 'mergedStack' ? 'visible' : 'hidden';
			}
			for (let i = 0; i < values.length; i++) {
				const path = 'DynamicComponentValues[' + deckName + '][' + i + ']';
				if (visible[deckName][i]) {
					try {
						sanitizeSubState(def, sub, values[i], transformation.DynamicComponentValues[deckName], null, path, null);
					} catch (err) {
						extend(err, "Couldn't sanitize random dynamic component");
					}
				} else {
					sanitizeSubState(def, sub, values[i], hiddenAll, null, path, null);
				}
			}
		}
	} catch (err) {
		extend(err, "Couldn't sanitize dyanmic component values");
	}

	sanitized.visibility = visibleFacets(state, transformation);
	sanitized.refreshViews();
	return sanitized;
}

export interface ViewOptions extends SanitizeOptions {
	/**
	 * Computes the "Computed" block from the (sanitized) state, as the
	 * game's computed properties would. Computed properties are game rules,
	 * outside the engine core; with no callback the block is omitted.
	 */
	computed?: (sanitized: EngineState) => unknown;
}

/**
 * The JSON a viewer receives as the current state: what Go's
 * Game.JSONForPlayer puts in "CurrentState" (state.customMarshalJSON with
 * Visibility). SecretMoveCount is present only on the unsanitized (admin)
 * view.
 */
export function viewFor(state: EngineState, viewer: number, options: ViewOptions = {}): StateJSON {
	const sanitized = sanitizedForPlayer(state, viewer, options);
	const out = sanitized.toJSON();
	delete out.Timers;
	delete out.Computed;
	out.Visibility = sanitized.sanitized && sanitized.visibility ? sanitized.visibility : visibleFacets(sanitized, null);
	if (options.computed) out.Computed = options.computed(sanitized);
	return out;
}

export { StateError };
