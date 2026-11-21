// EngineState: one version of one game's state, as plain JSON in the shape
// the Go server persists (state.go customMarshalJSON / refriedState), plus
// the handles the rules use to change it.
//
// The JSON is the state. Nothing is held beside it except what Go also holds
// beside it: the game's identity (id and secret salt), the version, whether
// it has been sanitized, and the memoized generator.

import { stateRand } from './rand.ts';
import type { GoRand } from './rand.ts';
import { componentId, secretMoveCountFor, EMPTY_SLOT, GENERIC_COMPONENT } from './ids.ts';
import type { GameIdentity } from './ids.ts';
import { indexesOf, newStack, newMergedStackView, isMergedStackView } from './stack.ts';
import type { MergedStackView, Stack } from './stack.ts';
import type { StackJSON, StackView } from './stack.ts';
import type { GameDefinition, PropDef, SubStateDef } from './definition.ts';

export interface BoardJSON {
	Spaces: StackJSON[];
}

/** One sub-state: property name to its JSON value. */
export type SubStateJSON = Record<string, unknown>;

/** state_visibility.go StateVisibility. */
export interface VisibilityJSON {
	Game: Record<string, string[]>;
	Players: Record<string, string[]>[];
}

/** The state blob, as Go's customMarshalJSON writes it. */
export interface StateJSON {
	Game: SubStateJSON;
	Players: SubStateJSON[];
	Components: Record<string, SubStateJSON[]>;
	/** Present only on unsanitized states, and only when non-empty. */
	SecretMoveCount?: Record<string, number[]>;
	/** Present only on per-viewer JSON. */
	Visibility?: VisibilityJSON;
	/** Timer storage records. Passed through untouched; timers are not core. */
	Timers?: unknown;
	/** Present only on JSON for a viewer. Derived; dropped on load. */
	Computed?: unknown;
}

export type SubStateGroup = 'game' | 'player' | 'dynamic';

/** Names one sub-state within a state. */
export interface SubStateRef {
	Group: SubStateGroup;
	/** Player index, for Group 'player'. Absent means 0. */
	Player?: number;
	/** Deck name and deck index (absent means 0), for Group 'dynamic'. */
	Deck?: string;
	Index?: number;
}

/** Names one stack property, or one space of a board property. */
export interface StackRef extends SubStateRef {
	Prop: string;
	/** For a board property: which space. */
	Space?: number;
}

export interface StateError extends Error {
	readonly name: 'StateError';
}

/**
 * Creates a `StateError`. An ordinary `Error` named `StateError`, not a subclass:
 * a Common Fabric pattern module may not declare a class (see README.md,
 * "No classes").
 */
export function stateError(message: string): StateError {
	return Object.assign(new Error(message), { name: 'StateError' as const });
}

/** Whether `thrown` is a `StateError`. */
export function isStateError(thrown: unknown): thrown is StateError {
	return thrown instanceof Error && thrown.name === 'StateError';
}

/** A deep copy of JSON data. */
export function cloneJSON<T>(value: T): T {
	return cloneValue(value) as T;
}

function cloneValue(value: unknown): unknown {
	if (Array.isArray(value)) {
		const out = new Array<unknown>(value.length);
		for (let i = 0; i < value.length; i++) out[i] = cloneValue(value[i]);
		return out;
	}
	if (value !== null && typeof value === 'object') {
		const out: Record<string, unknown> = {};
		for (const key of Object.keys(value)) out[key] = cloneValue((value as Record<string, unknown>)[key]);
		return out;
	}
	return value;
}

/** Go's sort.Strings order: by UTF-8 bytes, which is code point order. */
export function compareGoStrings(a: string, b: string): number {
	const length = Math.min(a.length, b.length);
	for (let i = 0; i < length; i++) {
		let ca = a.charCodeAt(i);
		let cb = b.charCodeAt(i);
		if (ca === cb) continue;
		// UTF-16 surrogates sort below U+E000..U+FFFF, but their code points
		// are above. Lift them so the comparison is by code point.
		if (ca >= 0xd800 && ca <= 0xdfff) ca += 0x2000;
		else if (ca >= 0xe000) ca -= 0x800;
		if (cb >= 0xd800 && cb <= 0xdfff) cb += 0x2000;
		else if (cb >= 0xe000) cb -= 0x800;
		return ca - cb;
	}
	return a.length - b.length;
}

export function sortedKeys(obj: Readonly<Record<string, unknown>>): string[] {
	return Object.keys(obj).sort(compareGoStrings);
}

interface SubStateEntry {
	/** Go's owner-path prefix: "Game", "Players[1]", "DynamicComponentValues[cards][3]". */
	path: string;
	ref: SubStateRef;
	def: SubStateDef;
	node: SubStateJSON;
}

export interface LoadOptions {
	/**
	 * Treat the JSON as a sanitized (per-viewer) state: stacks may hold
	 * generic components, ids are kept as given, nothing may be modified and
	 * conservation is not checked.
	 */
	sanitized?: boolean;
	/** Skip the conservation check Go runs when loading a stored state. */
	skipConservation?: boolean;
}

export interface EngineState {
	readonly def: GameDefinition;
	readonly identity: GameIdentity;
	/** The live JSON. Physical stacks in it are always current. */
	readonly data: StateJSON;
	version: number;
	readonly sanitized: boolean;
	/** On a sanitized state: which facets of each property survived. */
	visibility: VisibilityJSON | null;
	memoizedRand: GoRand | null;
	readonly stacksByPath: Record<string, Stack>;
	readonly stackList: Stack[];
	readonly subStates: SubStateEntry[];
	componentIndex: Record<string, { stack: Stack; slotIndex: number }> | null;
	/**
	 * state.copy(false): an independent state with the same contents. Pass a
	 * version to do what applyMove does to the copy before applying a move.
	 * The copy gets its own generator, seeded from its own version.
	 */
	copy(version?: number): EngineState;
	/** A copy marked sanitized: state.copy(true). Used by sanitize.ts. */
	copySanitized(): EngineState;
	/**
	 * state.Rand(): the generator for this version of this game, seeded from
	 * game id + secret salt + version. Memoized, so every draw made while
	 * applying one move continues a single sequence.
	 */
	rand(): GoRand;
	// ----- sub-states -----------------------------------------------------
	index(): void;
	register(node: StackJSON, kind: 'growable' | 'sized', path: string, inBoard: boolean): void;
	subState(ref: SubStateRef): SubStateEntry;
	/** The JSON object of one sub-state. Scalars are read and written here. */
	subStateNode(ref: SubStateRef): SubStateJSON;
	readonly game: SubStateJSON;
	readonly players: SubStateJSON[];
	numPlayers(): number;
	/** The dynamic values of one component, or undefined if its deck has none. */
	dynamicValues(deckName: string, deckIndex: number): SubStateJSON | undefined;
	// ----- stacks ---------------------------------------------------------
	/** A physical stack. Throws for a merged stack; use view() for those. */
	stack(ref: StackRef): Stack;
	gameStack(prop: string): Stack;
	playerStack(player: number, prop: string): Stack;
	dynamicStack(deckName: string, deckIndex: number, prop: string): Stack;
	/** One space of a board property on the game state. */
	gameBoardSpace(prop: string, space: number): Stack;
	/** Any stack property, physical or merged, as a read-only view. */
	view(ref: StackRef): StackView;
	mergedView(sub: SubStateEntry, propName: string, visiting: string[]): MergedStackView;
	/** Every physical stack, ordered by owner path as Go orders them. */
	physicalStacks(): readonly Stack[];
	/** validateStackAttachment: null when stack is a live stack of this state. */
	stackAttachmentError(stack: Stack): string | null;
	/**
	 * state.containingStack: where a component is, or null if Go would return
	 * an error. See locateComponent for the error text.
	 */
	containingStack(deckName: string, deckIndex: number): { stack: Stack; slotIndex: number } | null;
	/**
	 * state.containingStack, returning Go's error text on failure.
	 *
	 * This is a literal port, including the lazily built component index and
	 * exactly the points at which Go updates it. That matters in one case:
	 * sizedStack.ContractSize removes empty slots without refreshing the
	 * index, so a component that shifted left keeps a stale entry until
	 * something re-adds it, and until then Go refuses to move it ("that
	 * component's location is not public information"). The engine refuses
	 * identically.
	 */
	locateComponent(deckName: string, deckIndex: number): { stack: Stack; slotIndex: number } | string;
	buildComponentIndex(): void;
	componentAddedImpl(deckName: string, deckIndex: number | null, stack: Stack, slotIndex: number): void;
	/** state.componentAdded. Called by stacks; not for game code. */
	componentAdded(deckName: string, deckIndex: number | null, stack: Stack, slotIndex: number): void;
	/** state.updateIndexForAllComponents. Called by stacks; not for game code. */
	updateIndexForAllComponents(stack: Stack): void;
	// ----- ids ------------------------------------------------------------
	/** componentInstance.secretMoveCount. */
	secretMoveCount(deckName: string, deckIndex: number): number;
	/** componentInstance.ID() in this state. */
	componentId(deckName: string, deckIndex: number): string;
	/** componentInstance.movedSecretly: bump one component's count. */
	movedSecretly(deckName: string, deckIndex: number): void;
	// ----- serialization --------------------------------------------------
	/**
	 * Rewrites every merged-stack property in `data` from its parts. Merged
	 * stacks are views, so their JSON goes stale as their parts change; this
	 * is called by toJSON and may be called directly.
	 */
	refreshViews(): void;
	/**
	 * The state blob as Go writes it for storage (customMarshalJSON without
	 * Computed or Visibility): a fresh copy, merged views current,
	 * SecretMoveCount present only when unsanitized and non-empty.
	 */
	toJSON(): StateJSON;
	// ----- conservation ---------------------------------------------------
	/**
	 * validateComponentConservation (stack_ownership.go): every component of
	 * every deck is in exactly one slot of one physical stack. Returns Go's
	 * error text, or null when the invariant holds. Sanitized states are
	 * exempt, as in Go.
	 */
	conservationError(): string | null;
}

/** Creates an `EngineState`. */
export function newEngineState(def: GameDefinition, identity: GameIdentity, version: number, data: StateJSON, sanitized: boolean): EngineState {
	const self: EngineState = {
		def,
		identity,
		data,
		version,
		sanitized,
		visibility: null,
		memoizedRand: null,
		stacksByPath: {},
		stackList: [],
		subStates: [],
		componentIndex: null,
		/**
		 * state.copy(false): an independent state with the same contents. Pass a
		 * version to do what applyMove does to the copy before applying a move.
		 * The copy gets its own generator, seeded from its own version.
		 */
		copy(version?: number): EngineState {
			const state = newEngineState(this.def, this.identity, version ?? this.version, cloneJSON(this.data), this.sanitized);
			state.visibility = this.visibility;
			state.index();
			return state;
		},
		/** A copy marked sanitized: state.copy(true). Used by sanitize.ts. */
		copySanitized(): EngineState {
			const data = cloneJSON(this.data);
			const state = newEngineState(this.def, this.identity, this.version, data, true);
			state.index();
			return state;
		},
		/**
		 * state.Rand(): the generator for this version of this game, seeded from
		 * game id + secret salt + version. Memoized, so every draw made while
		 * applying one move continues a single sequence.
		 */
		rand(): GoRand {
			if (this.memoizedRand === null) {
				this.memoizedRand = stateRand(this.identity.gameId, this.identity.secretSalt, this.version);
			}
			return this.memoizedRand;
		},
		// ----- sub-states -----------------------------------------------------
		index(): void {
			const data = this.data;
			this.subStates.push({ path: 'Game', ref: { Group: 'game' }, def: this.def.Game, node: data.Game });
			for (let i = 0; i < data.Players.length; i++) {
				this.subStates.push({
					path: 'Players[' + i + ']',
					ref: { Group: 'player', Player: i },
					def: this.def.Player,
					node: data.Players[i],
				});
			}
			for (const deckName of sortedKeys(data.Components)) {
				const def = this.def.DynamicComponentValues[deckName];
				if (!def) throw stateError("The empty dynamic component state didn't have deck name: " + deckName);
				const values = data.Components[deckName];
				for (let i = 0; i < values.length; i++) {
					this.subStates.push({
						path: 'DynamicComponentValues[' + deckName + '][' + i + ']',
						ref: { Group: 'dynamic', Deck: deckName, Index: i },
						def,
						node: values[i],
					});
				}
			}
			for (const sub of this.subStates) {
				for (const propName of sortedKeys(sub.def.Props)) {
					const prop = sub.def.Props[propName];
					const path = sub.path + '.' + propName;
					if (prop.Type === 'stack' || prop.Type === 'sizedStack') {
						this.register(normalizeStack(sub.node[propName], path, prop), prop.Type === 'sizedStack' ? 'sized' : 'growable', path, false);
					} else if (prop.Type === 'board') {
						const board = sub.node[propName] as BoardJSON | undefined;
						if (!board || !Array.isArray(board.Spaces)) throw stateError(path + ' is nil');
						for (let i = 0; i < board.Spaces.length; i++) {
							const spacePath = path + '[' + i + ']';
							this.register(normalizeStack(board.Spaces[i], spacePath, prop), 'growable', spacePath, true);
						}
					}
				}
			}
			// buildComponentIndex and conservation both walk owners by path.
			this.stackList.sort((a, b) => compareGoStrings(a.path, b.path));
		},
		register(node: StackJSON, kind: 'growable' | 'sized', path: string, inBoard: boolean): void {
			if (kind === 'sized' && indexesOf(node).length !== (node.Size ?? 0)) {
				throw stateError("Couldn't unmarshal sized stack: lenght of indexes didn't agree with size");
			}
			const stack = newStack(this, node, kind, path, inBoard);
			this.stacksByPath[path] = stack;
			this.stackList.push(stack);
		},
		subState(ref: SubStateRef): SubStateEntry {
			for (const sub of this.subStates) {
				if (sub.ref.Group !== ref.Group) continue;
				if (ref.Group === 'game') return sub;
				if (ref.Group === 'player' && sub.ref.Player === (ref.Player ?? 0)) return sub;
				if (ref.Group === 'dynamic' && sub.ref.Deck === ref.Deck && sub.ref.Index === (ref.Index ?? 0)) return sub;
			}
			throw stateError('no such sub-state: ' + JSON.stringify(ref));
		},
		/** The JSON object of one sub-state. Scalars are read and written here. */
		subStateNode(ref: SubStateRef): SubStateJSON {
			return this.subState(ref).node;
		},
		get game(): SubStateJSON {
			return this.data.Game;
		},
		get players(): SubStateJSON[] {
			return this.data.Players;
		},
		numPlayers(): number {
			return this.data.Players.length;
		},
		/** The dynamic values of one component, or undefined if its deck has none. */
		dynamicValues(deckName: string, deckIndex: number): SubStateJSON | undefined {
			const values = this.data.Components[deckName];
			return values ? values[deckIndex] : undefined;
		},
		// ----- stacks ---------------------------------------------------------
		/** A physical stack. Throws for a merged stack; use view() for those. */
		stack(ref: StackRef): Stack {
			const sub = this.subState(ref);
			const prop = sub.def.Props[ref.Prop];
			if (!prop) throw stateError(sub.path + '.' + ref.Prop + ' is not a property');
			let path = sub.path + '.' + ref.Prop;
			if (prop.Type === 'board') {
				if (ref.Space === undefined) throw stateError(path + ' is a board; name a space');
				path += '[' + ref.Space + ']';
			} else if (prop.Type === 'mergedStack') {
				throw stateError(path + ' is a view, not a physical stack');
			}
			const stack = this.stacksByPath[path];
			if (!stack) throw stateError(path + ' is not a physical stack');
			return stack;
		},
		gameStack(prop: string): Stack {
			return this.stack({ Group: 'game', Prop: prop });
		},
		playerStack(player: number, prop: string): Stack {
			return this.stack({ Group: 'player', Player: player, Prop: prop });
		},
		dynamicStack(deckName: string, deckIndex: number, prop: string): Stack {
			return this.stack({ Group: 'dynamic', Deck: deckName, Index: deckIndex, Prop: prop });
		},
		/** One space of a board property on the game state. */
		gameBoardSpace(prop: string, space: number): Stack {
			return this.stack({ Group: 'game', Prop: prop, Space: space });
		},
		/** Any stack property, physical or merged, as a read-only view. */
		view(ref: StackRef): StackView {
			const sub = this.subState(ref);
			const prop = sub.def.Props[ref.Prop];
			if (prop && prop.Type === 'mergedStack') return this.mergedView(sub, ref.Prop, []);
			return this.stack(ref);
		},
		mergedView(sub: SubStateEntry, propName: string, visiting: string[]): MergedStackView {
			const prop = sub.def.Props[propName];
			const path = sub.path + '.' + propName;
			if (visiting.indexOf(propName) >= 0) throw stateError(path + ' contains a cycle of merged stacks');
			const parts: StackView[] = [];
			for (const partName of prop.Stacks ?? []) {
				const partProp = sub.def.Props[partName];
				if (!partProp) throw stateError(path + ' merges ' + partName + ', which is not a property');
				if (partProp.Type === 'mergedStack') {
					parts.push(this.mergedView(sub, partName, visiting.concat(propName)));
				} else {
					const stack = this.stacksByPath[sub.path + '.' + partName];
					if (!stack) throw stateError(path + ' merges ' + partName + ', which is not a stack');
					parts.push(stack);
				}
			}
			return newMergedStackView(parts, prop.Overlap === true);
		},
		/** Every physical stack, ordered by owner path as Go orders them. */
		physicalStacks(): readonly Stack[] {
			return this.stackList;
		},
		/** validateStackAttachment: null when stack is a live stack of this state. */
		stackAttachmentError(stack: Stack): string | null {
			if (stack.state !== this) return 'stack belongs to a different state';
			if (this.stacksByPath[stack.path] !== stack) return 'stack is not an attached owner in this state';
			return null;
		},
		/**
		 * state.containingStack: where a component is, or null if Go would return
		 * an error. See locateComponent for the error text.
		 */
		containingStack(deckName: string, deckIndex: number): { stack: Stack; slotIndex: number } | null {
			const found = this.locateComponent(deckName, deckIndex);
			return typeof found === 'string' ? null : found;
		},
		/**
		 * state.containingStack, returning Go's error text on failure.
		 *
		 * This is a literal port, including the lazily built component index and
		 * exactly the points at which Go updates it. That matters in one case:
		 * sizedStack.ContractSize removes empty slots without refreshing the
		 * index, so a component that shifted left keeps a stale entry until
		 * something re-adds it, and until then Go refuses to move it ("that
		 * component's location is not public information"). The engine refuses
		 * identically.
		 */
		locateComponent(deckName: string, deckIndex: number): { stack: Stack; slotIndex: number } | string {
			if (this.componentIndex === null) this.buildComponentIndex();
			if (deckIndex === GENERIC_COMPONENT) return "The generic component for that deck isn't in any stack";
			const item = this.componentIndex![deckName + '\u0000' + deckIndex];
			if (!item) {
				if (this.sanitized) return "that component's location is not public information";
				return 'component was not found in any attached stack; a captured stack may be stale or replaced';
			}
			const other = item.stack.componentAt(item.slotIndex);
			if (other === null || other === GENERIC_COMPONENT) return "that component's location is not public information";
			if (other !== deckIndex || item.stack.node.Deck !== deckName) {
				return 'Unexpectedly that component was not found in the index';
			}
			return item;
		},
		buildComponentIndex(): void {
			this.componentIndex = {};
			for (const stack of this.stackList) {
				for (let i = 0; i < stack.len(); i++) {
					this.componentAddedImpl(stack.node.Deck, stack.componentAt(i), stack, i);
				}
			}
		},
		componentAddedImpl(deckName: string, deckIndex: number | null, stack: Stack, slotIndex: number): void {
			if (deckIndex === null || deckIndex === GENERIC_COMPONENT) return;
			this.componentIndex![deckName + '\u0000' + deckIndex] = { stack, slotIndex };
		},
		/** state.componentAdded. Called by stacks; not for game code. */
		componentAdded(deckName: string, deckIndex: number | null, stack: Stack, slotIndex: number): void {
			if (this.componentIndex === null) this.buildComponentIndex();
			this.componentAddedImpl(deckName, deckIndex, stack, slotIndex);
		},
		/** state.updateIndexForAllComponents. Called by stacks; not for game code. */
		updateIndexForAllComponents(stack: Stack): void {
			for (let i = 0; i < stack.len(); i++) {
				this.componentAdded(stack.node.Deck, stack.componentAt(i), stack, i);
			}
		},
		// ----- ids ------------------------------------------------------------
		/** componentInstance.secretMoveCount. */
		secretMoveCount(deckName: string, deckIndex: number): number {
			return secretMoveCountFor(this.data.SecretMoveCount, deckName, deckIndex);
		},
		/** componentInstance.ID() in this state. */
		componentId(deckName: string, deckIndex: number): string {
			if (deckIndex === GENERIC_COMPONENT || deckIndex === EMPTY_SLOT) return '';
			return componentId(this.identity, deckName, deckIndex, this.secretMoveCount(deckName, deckIndex));
		},
		/** componentInstance.movedSecretly: bump one component's count. */
		movedSecretly(deckName: string, deckIndex: number): void {
			if (deckIndex < 0) return;
			if (!this.data.SecretMoveCount) this.data.SecretMoveCount = {};
			const counts = this.data.SecretMoveCount;
			let deckCounts = counts[deckName];
			if (!deckCounts) {
				const deck = this.def.Decks[deckName];
				deckCounts = new Array<number>(deck ? deck.Size : 0);
				for (let i = 0; i < deckCounts.length; i++) deckCounts[i] = 0;
				counts[deckName] = deckCounts;
			}
			if (deckIndex >= deckCounts.length) return;
			deckCounts[deckIndex]++;
		},
		// ----- serialization --------------------------------------------------
		/**
		 * Rewrites every merged-stack property in `data` from its parts. Merged
		 * stacks are views, so their JSON goes stale as their parts change; this
		 * is called by toJSON and may be called directly.
		 */
		refreshViews(): void {
			for (const sub of this.subStates) {
				for (const propName of Object.keys(sub.def.Props)) {
					if (sub.def.Props[propName].Type !== 'mergedStack') continue;
					sub.node[propName] = this.mergedView(sub, propName, []).toJSON();
				}
			}
		},
		/**
		 * The state blob as Go writes it for storage (customMarshalJSON without
		 * Computed or Visibility): a fresh copy, merged views current,
		 * SecretMoveCount present only when unsanitized and non-empty.
		 */
		toJSON(): StateJSON {
			this.refreshViews();
			const out = cloneJSON(this.data);
			if (this.sanitized || !out.SecretMoveCount || Object.keys(out.SecretMoveCount).length === 0) {
				delete out.SecretMoveCount;
			}
			return out;
		},
		// ----- conservation ---------------------------------------------------
		/**
		 * validateComponentConservation (stack_ownership.go): every component of
		 * every deck is in exactly one slot of one physical stack. Returns Go's
		 * error text, or null when the invariant holds. Sanitized states are
		 * exempt, as in Go.
		 */
		conservationError(): string | null {
			if (this.sanitized) return null;

			// validateCurrentMergedViews
			for (const sub of this.subStates) {
				for (const propName of sortedKeys(sub.def.Props)) {
					if (sub.def.Props[propName].Type !== 'mergedStack') continue;
					const path = sub.path + '.' + propName;
					let view: MergedStackView;
					try {
						view = this.mergedView(sub, propName, []);
					} catch (e) {
						return (e as Error).message;
					}
					const invalid = view.validityError();
					if (invalid !== null) return path + ' is an invalid merged stack: ' + invalid;
					const repeated = repeatedLeaf(view, path);
					if (repeated !== null) return repeated;
				}
			}

			const found: Record<string, string[]> = {};
			const deckNames = sortedKeys(this.def.Decks);
			for (const deckName of deckNames) {
				const locations = new Array<string>(this.def.Decks[deckName].Size);
				for (let i = 0; i < locations.length; i++) locations[i] = '';
				found[deckName] = locations;
			}

			for (const stack of this.stackList) {
				const deckName = stack.node.Deck;
				const locations = found[deckName];
				if (!locations) return stack.path + ' names unknown deck ' + goQuote(deckName);
				const indexes = indexesOf(stack.node);
				for (let slot = 0; slot < indexes.length; slot++) {
					const index = indexes[slot];
					const location = stack.path + '[' + slot + ']';
					if (index === EMPTY_SLOT) {
						if (stack.kind === 'sized') continue;
						return location + ' contains the empty-slot sentinel in a growable stack';
					}
					if (index === GENERIC_COMPONENT) {
						return location + ' contains a generic component in an unsanitized state';
					}
					if (index < 0 || index >= locations.length) {
						return location + ' contains invalid component index ' + index + ' for deck ' + goQuote(deckName);
					}
					if (locations[index] !== '') {
						return 'deck ' + goQuote(deckName) + ' component ' + index + ' appears at both ' + locations[index] + ' and ' + location;
					}
					locations[index] = location;
				}
			}

			for (const deckName of deckNames) {
				const locations = found[deckName];
				for (let index = 0; index < locations.length; index++) {
					if (locations[index] === '') {
						return 'deck ' + goQuote(deckName) + ' component ' + index + ' is missing from every attached stack';
					}
				}
			}
			return null;
		},
	};
	return self;
}

/**
	 * Builds a state from JSON the Go server wrote (a storage record, or a
	 * golden history entry). The input is copied. As GameManager.
	 * stateFromRecord does, recorded ids are ignored and recomputed, and the
	 * state must satisfy component conservation.
	 */
export function loadEngineState(
		def: GameDefinition,
		identity: GameIdentity,
		version: number,
		json: StateJSON,
		options: LoadOptions = {},
	): EngineState {
		const data = cloneJSON(json);
		if (!data.Components) data.Components = {};
		// Derived, never stored: Go's loader ignores both. (A sanitized
		// state keeps its visibility beside the data, as Go does.)
		const visibility = data.Visibility ?? null;
		delete data.Computed;
		delete data.Visibility;
		const sanitized = options.sanitized === true;
		if (sanitized) delete data.SecretMoveCount;
		else if (!data.SecretMoveCount) data.SecretMoveCount = {};
		const state = newEngineState(def, identity, version, data, sanitized);
		if (sanitized) state.visibility = visibility;
		state.index();
		if (!sanitized) {
			for (const stack of state.stackList) stack.syncIds();
			if (!options.skipConservation) {
				const err = state.conservationError();
				if (err !== null) throw stateError('Loaded state violated component conservation: ' + err);
			}
		}
		return state;
	}

/** fmt's %q for the plain names decks have. */
function goQuote(text: string): string {
	return JSON.stringify(text);
}

function normalizeStack(value: unknown, path: string, prop: PropDef): StackJSON {
	if (value === null || typeof value !== 'object') throw stateError(path + ' is nil');
	const node = value as StackJSON;
	if (typeof node.Deck !== 'string' || node.Deck === '') node.Deck = prop.Deck ?? '';
	if (!Array.isArray(node.Indexes)) node.Indexes = null; // Go's nil slice
	if (!Array.isArray(node.IDs)) node.IDs = [];
	if (node.IDsLastSeen === null || typeof node.IDsLastSeen !== 'object') node.IDsLastSeen = {};
	if (typeof node.ShuffleCount !== 'number') node.ShuffleCount = 0;
	return node;
}

/** validateMergedOwnerLeaves' rule that a view uses each stack once. */
function repeatedLeaf(view: MergedStackView, path: string): string | null {
	const leaves: { stack: Stack; path: string }[] = [];
	const walk = (current: MergedStackView, currentPath: string): string | null => {
		for (let i = 0; i < current.parts.length; i++) {
			const child = current.parts[i];
			const childPath = currentPath + '[' + i + ']';
			if (isMergedStackView(child)) {
				const nested = walk(child, childPath);
				if (nested !== null) return nested;
				continue;
			}
			const physical = child as Stack;
			for (const leaf of leaves) {
				if (leaf.stack === physical) {
					return currentPath + ' repeats physical stack ' + physical.path + ' (first used at ' + leaf.path + ')';
				}
			}
			leaves.push({ stack: physical, path: childPath });
		}
		return null;
	};
	return walk(view, path);
}
