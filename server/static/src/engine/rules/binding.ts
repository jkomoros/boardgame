// The legal evaluator's state-access interface (../legal/access.ts) over the
// real engine state. This replaces the fixture-only stand-in in
// ../legal/testing/fixtures.ts: every stack answer here comes from ../stack.ts
// and every property from the state's own JSON.
//
// The evaluator asks three things of a state that Go answers through the
// game delegate rather than the state (current player, who may be active,
// player order). They are passed in as DelegateAnswers, so this module knows
// nothing about game definitions.

import type { GameDefinition, PropDef, PropType, SubStateDef } from '../definition.ts';
import { GENERIC_COMPONENT } from '../ids.ts';
import type {
	LegalBoard,
	LegalChest,
	LegalComponent,
	LegalEnumValue,
	LegalMove,
	LegalPropertyReader,
	LegalStack,
	LegalState,
} from '../legal/access.ts';
import type { PropertyType } from '../legal/types.ts';
import { isMergedStackView, isStack } from '../stack.ts';
import type { MergedStackView, Stack } from '../stack.ts';
import type { StackView } from '../stack.ts';
import type { BoardJSON, EngineState, SubStateJSON, SubStateRef } from '../state.ts';
import type { EnumRegistry } from './enums.ts';

/** PropertyType.String() for each engine property type. */
export const PROPERTY_TYPE_NAMES: Readonly<Record<PropType, PropertyType>> = {
	int: 'TypeInt',
	bool: 'TypeBool',
	string: 'TypeString',
	playerIndex: 'TypePlayerIndex',
	enum: 'TypeEnum',
	intSlice: 'TypeIntSlice',
	boolSlice: 'TypeBoolSlice',
	stringSlice: 'TypeStringSlice',
	playerIndexSlice: 'TypePlayerIndexSlice',
	enumSlice: 'TypeEnumSlice',
	timer: 'TypeTimer',
	stack: 'TypeStack',
	sizedStack: 'TypeStack',
	mergedStack: 'TypeStack',
	board: 'TypeBoard',
};

/** The scalar property types a component's static values or a move may have. */
export interface ValuePropDef {
	Type: PropType;
	Enum?: string;
}

/** Deck name to the declared types of its components' static values. */
export type ComponentSchema = Readonly<Record<string, Readonly<Record<string, ValuePropDef>>>>;

/** What Go asks the delegate, not the state. See LegalState. */
export interface DelegateAnswers {
	/** delegate.CurrentPlayerIndex(state). */
	currentPlayerIndex(): number;
	/** delegate.PlayerMayBeActive(players[index]). */
	playerMayBeActive(index: number): boolean;
	/** delegate.CustomPlayerOrder(state). */
	customPlayerOrder(): readonly number[] | null;
	/** behaviors.PlayerIsInactive(players[index]). */
	playerIsInactive(index: number): boolean;
}

/** Everything static the binding needs beside the state. */
export interface BindingContext {
	enums: EnumRegistry;
	/** Declared types of component values; undeclared values are inferred. */
	components?: ComponentSchema;
}

function inferType(value: unknown): PropertyType {
	if (typeof value === 'number') return 'TypeInt';
	if (typeof value === 'boolean') return 'TypeBool';
	if (typeof value === 'string') return 'TypeString';
	if (Array.isArray(value)) {
		if (value.length > 0 && typeof value[0] === 'boolean') return 'TypeBoolSlice';
		if (value.length > 0 && typeof value[0] === 'string') return 'TypeStringSlice';
		return 'TypeIntSlice';
	}
	return 'TypeIllegal';
}

/**
 * A reader over a plain object of scalar values: a component's static
 * values, or a move's fields. Enum values are stored as their strings.
 */
export function valueReader(
	values: Readonly<Record<string, unknown>>,
	schema: Readonly<Record<string, ValuePropDef>> | undefined,
	enums: EnumRegistry,
): LegalPropertyReader {
	const props: Record<string, PropertyType> = {};
	if (schema) {
		for (const name of Object.keys(schema)) props[name] = PROPERTY_TYPE_NAMES[schema[name].Type];
	}
	for (const name of Object.keys(values)) {
		if (!Object.prototype.hasOwnProperty.call(props, name)) props[name] = inferType(values[name]);
	}
	return {
		props: () => props,
		prop(name) {
			const raw = values[name];
			const declared = schema ? schema[name] : undefined;
			if (declared && declared.Type === 'enum') return enumValue(enums, declared.Enum, raw, name);
			return raw;
		},
	};
}

function enumValue(enums: EnumRegistry, enumName: string | undefined, raw: unknown, propName: string): LegalEnumValue | null {
	const e = enums.get(enumName);
	if (e === null) throw new Error('No such Enum prop: ' + propName);
	const key = typeof raw === 'number' ? raw : e.keyOf(raw);
	if (key === null) return null;
	return e.value(key);
}

interface BoundComponent extends LegalComponent {
	readonly state: EngineState;
	readonly ctx: BindingContext;
	readonly deckName: string;
	readonly deckIndex: number;
	values(): LegalPropertyReader | null;
	/** componentInstance.MayMoveTo. */
	mayMoveTo(destination: LegalStack): string | null;
	/** componentInstance.MayMoveToSlot. */
	mayMoveToSlot(destination: LegalStack, slotIndex: number): string | null;
}

/** Creates a `BoundComponent`. */
function newBoundComponent(state: EngineState, ctx: BindingContext, deckName: string, deckIndex: number): BoundComponent {
	const self: BoundComponent = {
		state,
		ctx,
		deckName,
		deckIndex,
		values(): LegalPropertyReader | null {
			const deck = this.state.def.Decks[this.deckName];
			const all = deck ? deck.Components : undefined;
			if (!all || this.deckIndex < 0) return null;
			const values = all[this.deckIndex];
			if (values === null || typeof values !== 'object') return null;
			const schema = this.ctx.components ? this.ctx.components[this.deckName] : undefined;
			return valueReader(values as Record<string, unknown>, schema, this.ctx.enums);
		},
		/** componentInstance.MayMoveTo. */
		mayMoveTo(destination: LegalStack): string | null {
			const dest = (destination as BoundStack).physical();
			if (dest === null) return 'destination stack is not attached: stack is not an attached owner in this state';
			const found = this.state.locateComponent(this.deckName, this.deckIndex);
			if (typeof found === 'string') return 'component is not in any stack: ' + found;
			return found.stack.mayMoveComponentTo(found.slotIndex, dest);
		},
		/** componentInstance.MayMoveToSlot. */
		mayMoveToSlot(destination: LegalStack, slotIndex: number): string | null {
			const err = this.mayMoveTo(destination);
			if (err !== null) return err;
			const dest = (destination as BoundStack).physical() as Stack;
			if (slotIndex < 0) return 'slot index must be non-negative';
			if (dest.isSized()) {
				if (slotIndex >= dest.len()) return 'slot index ' + slotIndex + ' is out of range (stack length ' + dest.len() + ')';
				if (dest.componentAt(slotIndex) !== null) return 'slot ' + slotIndex + ' is already occupied';
			} else if (slotIndex > dest.len()) {
				return 'slot index ' + slotIndex + ' is out of range (stack length ' + dest.len() + ')';
			}
			return null;
		},
	};
	return self;
}

/** Any stack property of a state, physical or merged, as the evaluator sees it. */
export interface BoundStack extends LegalStack {
	readonly state: EngineState;
	readonly ctx: BindingContext;
	readonly view: StackView;
	/** The physical stack, or null for a merged view. */
	physical(): Stack | null;
	len(): number;
	numComponents(): number;
	componentAt(index: number): LegalComponent | null;
	first(): LegalComponent | null;
	transfer(method: string, destination: LegalStack): { source: Stack; dest: Stack } | string;
	mayMoveAllTo(destination: LegalStack): string | null;
	mayMoveCountTo(destination: LegalStack, count: number): string | null;
	maySwapComponents(i: number, j: number): string | null;
}

/** Creates a `BoundStack`. */
export function newBoundStack(state: EngineState, ctx: BindingContext, view: StackView): BoundStack {
	const self: BoundStack = {
		state,
		ctx,
		view,
		/** The physical stack, or null for a merged view. */
		physical(): Stack | null {
			return isStack(this.view) ? this.view : null;
		},
		len(): number {
			return this.view.len();
		},
		numComponents(): number {
			return this.view.numComponents();
		},
		componentAt(index: number): LegalComponent | null {
			if (typeof index !== 'number' || index !== Math.floor(index)) return null;
			const deckIndex = this.view.componentAt(index);
			if (deckIndex === null) return null;
			return newBoundComponent(this.state, this.ctx, this.view.deckName, deckIndex);
		},
		first(): LegalComponent | null {
			const length = this.view.len();
			for (let i = 0; i < length; i++) {
				const deckIndex = this.view.componentAt(i);
				if (deckIndex !== null) return newBoundComponent(this.state, this.ctx, this.view.deckName, deckIndex);
			}
			return null;
		},
		transfer(method: string, destination: LegalStack): { source: Stack; dest: Stack } | string {
			if (isMergedStackView(this.view)) return method + ' is not supported on MergedStacks';
			const dest = (destination as BoundStack).physical();
			if (dest === null) return 'destination is a view, not a physical stack';
			return { source: this.view as Stack, dest };
		},
		mayMoveAllTo(destination: LegalStack): string | null {
			const ends = this.transfer('MayMoveAllTo', destination);
			if (typeof ends === 'string') return ends;
			return ends.source.mayMoveAllTo(ends.dest);
		},
		mayMoveCountTo(destination: LegalStack, count: number): string | null {
			const ends = this.transfer('MayMoveCountTo', destination);
			if (typeof ends === 'string') return ends;
			return ends.source.mayMoveCountTo(ends.dest, count);
		},
		maySwapComponents(i: number, j: number): string | null {
			if (isMergedStackView(this.view)) return 'MaySwapComponents is not supported on MergedStacks';
			return (this.view as Stack).maySwapComponents(i, j);
		},
	};
	return self;
}

interface BoundBoard extends LegalBoard {
	readonly spaces: readonly LegalStack[];
	enum(): null;
	spaceAt(index: number): LegalStack | null;
	spaceAtKey(key: number): LegalStack | null;
}

/** Creates a `BoundBoard`. */
function newBoundBoard(spaces: readonly LegalStack[]): BoundBoard {
	const self: BoundBoard = {
		spaces,
		enum(): null {
			return null;
		},
		spaceAt(index: number): LegalStack | null {
			return index === Math.floor(index) && index >= 0 && index < this.spaces.length ? this.spaces[index] : null;
		},
		spaceAtKey(key: number): LegalStack | null {
			return this.spaceAt(key);
		},
	};
	return self;
}

/** The reader of one sub-state of a state. */
export function subStateReader(state: EngineState, ctx: BindingContext, ref: SubStateRef, def: SubStateDef): LegalPropertyReader {
	const props: Record<string, PropertyType> = {};
	for (const name of Object.keys(def.Props)) props[name] = PROPERTY_TYPE_NAMES[def.Props[name].Type];
	return {
		props: () => props,
		prop(name) {
			const prop: PropDef | undefined = def.Props[name];
			if (!prop) throw new Error('No such property with that name: ' + name);
			const node: SubStateJSON = state.subStateNode(ref);
			switch (prop.Type) {
				case 'enum':
					return enumValue(ctx.enums, prop.Enum, node[name], name);
				case 'stack':
				case 'sizedStack':
				case 'mergedStack':
					return newBoundStack(state, ctx, state.view({ ...ref, Prop: name }));
				case 'board': {
					const board = node[name] as BoardJSON;
					const spaces: LegalStack[] = [];
					for (let i = 0; i < board.Spaces.length; i++) {
						spaces.push(newBoundStack(state, ctx, state.stack({ ...ref, Prop: name, Space: i })));
					}
					return newBoundBoard(spaces);
				}
				default:
					return node[name];
			}
		},
	};
}

/** The whole state, as the evaluator reads it. */
export function bindLegalState(state: EngineState, ctx: BindingContext, delegate: DelegateAnswers): LegalState {
	const def: GameDefinition = state.def;
	let game: LegalPropertyReader | null = null;
	let players: LegalPropertyReader[] | null = null;
	return {
		game() {
			if (game === null) game = subStateReader(state, ctx, { Group: 'game' }, def.Game);
			return game;
		},
		players() {
			if (players === null) {
				players = [];
				for (let i = 0; i < state.numPlayers(); i++) {
					players.push(subStateReader(state, ctx, { Group: 'player', Player: i }, def.Player));
				}
			}
			return players;
		},
		currentPlayerIndex: () => delegate.currentPlayerIndex(),
		playerMayBeActive: index => delegate.playerMayBeActive(index),
		customPlayerOrder: () => delegate.customPlayerOrder(),
		playerIsInactive: index => delegate.playerIsInactive(index),
	};
}

/** A move's fields, as the evaluator reads them. */
export function bindLegalMove(
	fields: Readonly<Record<string, unknown>>,
	schema: Readonly<Record<string, ValuePropDef>>,
	enums: EnumRegistry,
): LegalMove {
	const reader = valueReader(fields, schema, enums);
	return { reader: () => reader };
}

export function bindLegalChest(enums: EnumRegistry): LegalChest {
	return { enums: () => enums.chestEnums() };
}

/** True for the generic component a sanitized stack shows. */
export function isGenericComponent(deckIndex: number | null): boolean {
	return deckIndex === GENERIC_COMPONENT;
}
