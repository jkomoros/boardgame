// Stack constraints: what a stack's struct tag says about which components it
// accepts (`stack:"Tokens,same(Color)"`). Ports stack_constraint.go, the
// constraint half of struct_inflater.go's unpackStackStructTag, and the
// default constructors in package constraints (max, unique, same,
// maxdistinct).
//
// A constraint is data here, not a closure: parseStackConstraints turns the
// strings a PropDef carries into StackConstraint records once per state, and
// checkConstraints evaluates them with a switch. Go checks a stack's
// constraints in exactly one place, componentInstance.MayMoveTo, which every
// move into another stack goes through (MoveTo, SecretMoveTo, the
// first/last/next-slot moves, MoveAllTo and MoveCountTo, and every May...
// form); stack.ts calls checkConstraints from the same place. Swaps, slides,
// shuffles, sorts and set-up distribution are not checked, as in Go.

import type { GameDefinition, PropDef, PropType } from './definition.ts';
import { GENERIC_COMPONENT } from './ids.ts';
import type { Stack } from './stack.ts';

/** The constructors base.GameDelegate configures (constraints.DefaultConstructors). */
export type StackConstraintKind = 'max' | 'unique' | 'same' | 'maxdistinct';

/** Where a property path is read: Values(), DynamicValues(), or Values() then DynamicValues(). */
export type PropPathPrefix = 'component' | 'dynamic' | '';

/** One parsed constraint. */
export interface StackConstraint {
	readonly kind: StackConstraintKind;
	/** The constraint string as given, e.g. `same(Color)`. */
	readonly text: string;
	/** max and maxdistinct: the limit. */
	readonly max: number;
	/** unique, same and maxdistinct: the path as written (it appears in Go's errors). */
	readonly propPath: string;
	readonly prefix: PropPathPrefix;
	/** The property name after the prefix. */
	readonly propName: string;
}

/**
 * struct_inflater.go splitTagFields: splits a struct tag on commas outside
 * parentheses, so `cards,maxdistinct(color,2)` is two fields.
 */
export function splitTagFields(tag: string): string[] {
	const fields: string[] = [];
	let depth = 0;
	let start = 0;
	for (let i = 0; i < tag.length; i++) {
		const c = tag[i];
		if (c === '(') depth++;
		else if (c === ')') {
			if (depth > 0) depth--;
		} else if (c === ',' && depth === 0) {
			fields.push(tag.slice(start, i));
			start = i + 1;
		}
	}
	fields.push(tag.slice(start));
	return fields;
}

/**
 * The constraint pieces of a whole stack struct tag (`Tokens,same(Color)`),
 * trimmed, in order: what PropDef.Constraints holds. The deck name and size
 * fields are dropped.
 */
export function constraintsFromStackTag(tag: string): string[] {
	const out: string[] = [];
	const fields = splitTagFields(tag);
	for (let i = 1; i < fields.length; i++) {
		const piece = fields[i].trim();
		if (piece !== '' && parseConstraintExpr(piece) !== null) out.push(piece);
	}
	return out;
}

/**
 * struct_inflater.go parseConstraintExpr: `name(args)` to its name and its
 * comma-separated, trimmed arguments; null when the string is not of that
 * form. `name()` has no arguments.
 */
export function parseConstraintExpr(s: string): { name: string; args: string[] } | null {
	const paren = s.indexOf('(');
	if (paren < 0) return null;
	if (!s.endsWith(')')) return null;
	const name = s.slice(0, paren);
	if (name === '') return null;
	const argStr = s.slice(paren + 1, s.length - 1);
	if (argStr === '') return { name, args: [] };
	return { name, args: argStr.split(',').map(arg => arg.trim()) };
}

/** strconv.Quote for the plain strings constraint arguments are. */
function goQuote(text: string): string {
	return JSON.stringify(text);
}

const MAX_INT64_DIGITS = '9223372036854775807';
const MIN_INT64_DIGITS = '9223372036854775808';

/**
 * strconv.Atoi on a 64-bit platform: the integer, or Go's error text. A value
 * beyond 2^53 is in range for Go but is not exact here.
 */
export function goAtoi(s: string): number | string {
	if (!/^[+-]?[0-9]+$/.test(s)) return 'strconv.Atoi: parsing ' + goQuote(s) + ': invalid syntax';
	const negative = s[0] === '-';
	let digits = s[0] === '+' || s[0] === '-' ? s.slice(1) : s;
	digits = digits.replace(/^0+(?=.)/, '');
	const limit = negative ? MIN_INT64_DIGITS : MAX_INT64_DIGITS;
	if (digits.length > limit.length || (digits.length === limit.length && digits > limit)) {
		return 'strconv.Atoi: parsing ' + goQuote(s) + ': value out of range';
	}
	return Number(s);
}

/**
 * constraints/helpers.go intEffectiveValue: a named int constant from the
 * chest, else a decimal integer.
 */
function intEffectiveValue(str: string, constants: Readonly<Record<string, unknown>> | undefined): number | string {
	if (constants && Object.prototype.hasOwnProperty.call(constants, str)) {
		const value = constants[str];
		if (value !== null && value !== undefined) {
			if (typeof value !== 'number' || !Number.isInteger(value)) return str + ' is a constant, but not of type int';
			return value;
		}
	}
	const parsed = goAtoi(str);
	if (typeof parsed === 'string') return str + ' is not convertible to int: ' + parsed;
	return parsed;
}

/** constraints/prop_path.go splitPropPath. */
export function splitPropPath(propPath: string): { prefix: PropPathPrefix; name: string } {
	const dot = propPath.indexOf('.');
	if (dot >= 0) {
		const lower = propPath.slice(0, dot).toLowerCase();
		if (lower === 'component' || lower === 'dynamic') return { prefix: lower, name: propPath.slice(dot + 1) };
	}
	return { prefix: '', name: propPath };
}

/** The declared type of a deck's component value property, or null if its components lack it. */
function componentPropType(def: GameDefinition, deckName: string, name: string): PropType | 'unknown' | null {
	const deck = def.Decks[deckName];
	if (!deck) return null;
	if (deck.ComponentProps) {
		const prop = Object.prototype.hasOwnProperty.call(deck.ComponentProps, name) ? deck.ComponentProps[name] : undefined;
		return prop ? prop.Type : null;
	}
	const first = deck.Components && deck.Components.length > 0 ? deck.Components[0] : null;
	if (first === null || typeof first !== 'object') return null;
	return Object.prototype.hasOwnProperty.call(first, name) ? 'unknown' : null;
}

function dynamicPropDef(def: GameDefinition, deckName: string, name: string): PropDef | null {
	const sub = Object.prototype.hasOwnProperty.call(def.DynamicComponentValues, deckName) ? def.DynamicComponentValues[deckName] : undefined;
	if (!sub || !Object.prototype.hasOwnProperty.call(sub.Props, name)) return null;
	return sub.Props[name];
}

/** Property types Go's %v prints from values the state holds, rather than from a pointer. */
function isFormattable(type: PropType | 'unknown'): boolean {
	switch (type) {
		case 'timer':
		case 'stack':
		case 'sizedStack':
		case 'mergedStack':
		case 'board':
			return false;
	}
	return true;
}

/**
 * constraints/prop_path.go validatePropPath: the property must exist on the
 * first component of some deck, or in some deck's dynamic values. Returns
 * Go's error text or null. Also refuses, with the engine's own text, a path
 * that names a stack, board or timer: Go formats those with %v as a pointer,
 * which the engine cannot reproduce.
 */
function validatePropPath(propPath: string, def: GameDefinition): string | null {
	const { prefix, name } = splitPropPath(propPath);
	const checkComponent = prefix === 'component' || prefix === '';
	const checkDynamic = prefix === 'dynamic' || prefix === '';
	let found = false;
	for (const deckName of Object.keys(def.Decks)) {
		if (checkComponent) {
			const type = componentPropType(def, deckName, name);
			if (type !== null) {
				if (!isFormattable(type)) return 'property ' + propPath + ' is a stack, board or timer, which the engine cannot compare';
				found = true;
			}
		}
		if (checkDynamic) {
			const prop = dynamicPropDef(def, deckName, name);
			if (prop !== null) {
				if (!isFormattable(prop.Type)) return 'property ' + propPath + ' is a stack, board or timer, which the engine cannot compare';
				found = true;
			}
		}
	}
	if (found) return null;
	return 'property ' + propPath + " does not exist on any deck's components; check for typos in the property name";
}

/** One constructor of constraints.DefaultConstructors: the constraint, or Go's error text. */
function construct(
	name: string,
	args: string[],
	text: string,
	def: GameDefinition,
	constants: Readonly<Record<string, unknown>> | undefined,
): StackConstraint | string {
	switch (name) {
		case 'max': {
			if (args.length !== 1) return 'max constraint requires exactly 1 argument';
			const max = intEffectiveValue(args[0], constants);
			if (typeof max === 'string') return 'max constraint argument is not a valid int: ' + max;
			return { kind: 'max', text, max, propPath: '', prefix: '', propName: '' };
		}
		case 'unique':
		case 'same': {
			if (args.length !== 1) return name + ' constraint requires exactly 1 argument';
			const invalid = validatePropPath(args[0], def);
			if (invalid !== null) return name + ' constraint: ' + invalid;
			const path = splitPropPath(args[0]);
			return { kind: name === 'same' ? 'same' : 'unique', text, max: 0, propPath: args[0], prefix: path.prefix, propName: path.name };
		}
		case 'maxdistinct': {
			if (args.length !== 2) return 'maxdistinct constraint requires exactly 2 arguments: propPath and max';
			const invalid = validatePropPath(args[0], def);
			if (invalid !== null) return 'maxdistinct constraint: ' + invalid;
			const max = intEffectiveValue(args[1], constants);
			if (typeof max === 'string') return 'maxdistinct second argument is not a valid int: ' + max;
			const path = splitPropPath(args[0]);
			return { kind: 'maxdistinct', text, max, propPath: args[0], prefix: path.prefix, propName: path.name };
		}
	}
	return '';
}

/**
 * The chest's constants, if the definition carries them. GameDefinition has
 * no Constants field yet; an argument that names a constant (`max(LIMIT)`)
 * resolves once one is added, and until then fails as an unknown name does.
 */
function constantsOf(def: GameDefinition): Readonly<Record<string, unknown>> | undefined {
	const withConstants: { readonly Name: string; readonly Constants?: unknown } = def;
	const constants = withConstants.Constants;
	if (constants === null || typeof constants !== 'object' || Array.isArray(constants)) return undefined;
	return constants as Readonly<Record<string, unknown>>;
}

/**
 * Parses a stack property's constraint strings as unpackStackStructTag does,
 * in order. Returns the constraints, or the error Go's NewStructInflater
 * reports for that property (`Prop was a nil SizedStack and its struct tag
 * was not valid: constraint same error: ...`), which Go raises when the
 * manager boots.
 */
export function parseStackConstraints(propName: string, prop: PropDef, def: GameDefinition): readonly StackConstraint[] | string {
	const texts = prop.Constraints;
	if (!texts || texts.length === 0) return [];
	const constants = constantsOf(def);
	const prefix = propName + ' was a nil SizedStack and its struct tag was not valid: ';
	const out: StackConstraint[] = [];
	for (const raw of texts) {
		const piece = raw.trim();
		if (piece === '') continue;
		const expr = parseConstraintExpr(piece);
		if (expr === null) return prefix + 'unexpected field in struct tag: ' + piece;
		const built = construct(expr.name, expr.args, piece, def, constants);
		if (built === '') return prefix + 'unknown constraint ' + expr.name + ' in struct tag';
		if (typeof built === 'string') return prefix + 'constraint ' + expr.name + ' error: ' + built;
		out.push(built);
	}
	return out;
}

/** fmt's %v for a value as the state's JSON holds it. */
function formatValue(value: unknown, type: PropType | 'unknown'): string {
	if (value === null || value === undefined) {
		switch (type) {
			case 'intSlice':
			case 'boolSlice':
			case 'stringSlice':
			case 'playerIndexSlice':
			case 'enumSlice':
				return '[]';
		}
		return '<nil>';
	}
	if (Array.isArray(value)) return '[' + value.map(item => formatValue(item, 'unknown')).join(' ') + ']';
	if (typeof value === 'boolean') return value ? 'true' : 'false';
	if (typeof value === 'number') return String(value);
	if (typeof value === 'string') return value;
	return JSON.stringify(value);
}

function readStatic(dest: Stack, deckIndex: number, name: string): string | undefined {
	const def = dest.state.def;
	const deck = def.Decks[dest.node.Deck];
	if (!deck) return undefined;
	const values = deckIndex === GENERIC_COMPONENT ? deck.GenericValues : deck.Components ? deck.Components[deckIndex] : undefined;
	if (values === null || values === undefined || typeof values !== 'object') return undefined;
	const record = values as Readonly<Record<string, unknown>>;
	let type: PropType | 'unknown' = 'unknown';
	if (deck.ComponentProps) {
		const prop = Object.prototype.hasOwnProperty.call(deck.ComponentProps, name) ? deck.ComponentProps[name] : undefined;
		if (!prop) return undefined;
		type = prop.Type;
	} else if (!Object.prototype.hasOwnProperty.call(record, name)) {
		return undefined;
	}
	return formatValue(record[name], type);
}

function readDynamic(dest: Stack, deckIndex: number, name: string): string | undefined {
	if (deckIndex < 0) return undefined;
	const prop = dynamicPropDef(dest.state.def, dest.node.Deck, name);
	if (prop === null) return undefined;
	const node = dest.state.dynamicValues(dest.node.Deck, deckIndex);
	if (!node) return undefined;
	return formatValue(node[name], prop.Type);
}

/** constraints/prop_path.go resolvePropValue: %v of the property, or undefined when the component lacks it. */
function resolvePropValue(dest: Stack, deckIndex: number, constraint: StackConstraint): string | undefined {
	switch (constraint.prefix) {
		case 'component':
			return readStatic(dest, deckIndex, constraint.propName);
		case 'dynamic':
			return readDynamic(dest, deckIndex, constraint.propName);
	}
	const value = readStatic(dest, deckIndex, constraint.propName);
	if (value !== undefined) return value;
	return readDynamic(dest, deckIndex, constraint.propName);
}

/** The components already in dest, in slot order (ImmutableComponents without the nils). */
function existing(dest: Stack): number[] {
	const out: number[] = [];
	const length = dest.len();
	for (let i = 0; i < length; i++) {
		const deckIndex = dest.componentAt(i);
		if (deckIndex !== null) out.push(deckIndex);
	}
	return out;
}

function evaluate(constraint: StackConstraint, dest: Stack, proposed: readonly number[]): string | null {
	switch (constraint.kind) {
		case 'max': {
			const after = dest.numComponents() + proposed.length;
			if (after > constraint.max) {
				return 'stack would have ' + after + ' components, which exceeds max of ' + constraint.max;
			}
			return null;
		}
		case 'same': {
			let first: string | undefined;
			for (const deckIndex of existing(dest).concat(proposed)) {
				const value = resolvePropValue(dest, deckIndex, constraint);
				if (value === undefined) continue;
				if (first === undefined) {
					first = value;
					continue;
				}
				if (value !== first) return 'property ' + constraint.propPath + ' has mixed values: ' + first + ' and ' + value;
			}
			return null;
		}
		case 'unique': {
			const seen = new Set<string>();
			for (const deckIndex of existing(dest)) {
				const value = resolvePropValue(dest, deckIndex, constraint);
				if (value !== undefined) seen.add(value);
			}
			for (const deckIndex of proposed) {
				const value = resolvePropValue(dest, deckIndex, constraint);
				if (value === undefined) continue;
				if (seen.has(value)) return 'duplicate value ' + value + ' for property ' + constraint.propPath;
				seen.add(value);
			}
			return null;
		}
		case 'maxdistinct': {
			const seen = new Set<string>();
			for (const deckIndex of existing(dest).concat(proposed)) {
				const value = resolvePropValue(dest, deckIndex, constraint);
				if (value !== undefined) seen.add(value);
			}
			if (seen.size > constraint.max) {
				return 'property ' + constraint.propPath + ' would have ' + seen.size + ' distinct values, which exceeds max of ' + constraint.max;
			}
			return null;
		}
	}
	return null;
}

/**
 * Stack.CheckConstraints: whether dest, as it is now, would accept the
 * proposed components (deck indexes of dest's deck). Each constraint in
 * order; the first error wins. Null when all pass or there are none.
 */
export function checkConstraints(dest: Stack, proposed: readonly number[]): string | null {
	for (const constraint of dest.constraints) {
		const err = evaluate(constraint, dest, proposed);
		if (err !== null) return err;
	}
	return null;
}
