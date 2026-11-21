// compileGame: turns a GameSpec into a game ready to run. This is the work
// Go's NewGameManager does at boot: derive the state definition, resolve
// every move's fields, visibility and legal plan, and build the schemas a
// client is told about. A spec that Go would refuse to boot throws here.

import { GROUP_ALL, parseSanitizeTag } from '../definition.ts';
import type { DeckDef, EnumDef, GameDefinition, PolicyName, PropDef, SubStateDef } from '../definition.ts';
import { defaultLegalRegistry } from '../legal/catalog.ts';
import { type LegalCustom, buildLegalPlan, contributedPreconditions } from '../legal/plan.ts';
import { mergeLegalTemplates } from '../legal/templates.ts';
import type { LegalSpec } from '../legal/types.ts';
import { loadEngineState } from '../state.ts';
import type { EngineState } from '../state.ts';
import type { StateJSON, SubStateJSON } from '../state.ts';
import { type ComponentSchema, type ValuePropDef, bindLegalChest, bindLegalMove } from './binding.ts';
import { newEnumRegistry } from './enums.ts';
import type { EnumRegistry } from './enums.ts';
import { sha256Hex } from './sha256.ts';
import { type ChoiceSchemaItem, type CompiledGame, type CompiledMove, type InputField, rulesOf, newRulesState } from './state.ts';
import type { RulesState } from './state.ts';
import { syncTimerRecords } from './timers.ts';
import type { GameRecord, GameSpec, InputCodec, MoveFieldSpec, MoveSpec, SubStateSpec, VariantKeySpec } from './types.ts';

export interface GameSpecError extends Error {
	readonly name: 'GameSpecError';
}

/**
 * Creates a `GameSpecError`. An ordinary `Error` named `GameSpecError`, not a subclass:
 * a Common Fabric pattern module may not declare a class (see README.md,
 * "No classes").
 */
export function gameSpecError(message: string): GameSpecError {
	return Object.assign(new Error(message), { name: 'GameSpecError' as const });
}

/** Whether `thrown` is a `GameSpecError`. */
export function isGameSpecError(thrown: unknown): thrown is GameSpecError {
	return thrown instanceof Error && thrown.name === 'GameSpecError';
}

// ----- the empty state ----------------------------------------------------

function emptyStack(prop: PropDef): Record<string, unknown> {
	const node: Record<string, unknown> = { Deck: prop.Deck ?? '', Indexes: [], IDs: [], IDsLastSeen: {}, ShuffleCount: 0 };
	if (prop.Type === 'sizedStack') {
		const size = prop.Size ?? 0;
		const indexes: number[] = [];
		const ids: string[] = [];
		for (let i = 0; i < size; i++) {
			indexes.push(-1);
			ids.push('');
		}
		node.Indexes = indexes;
		node.IDs = ids;
		if (size !== 0) node.Size = size;
	} else if (prop.MaxSize) {
		node.MaxSize = prop.MaxSize;
	}
	return node;
}

function emptySubState(spec: SubStateSpec, enums: Readonly<Record<string, EnumDef>>, index: number): SubStateJSON {
	const node: SubStateJSON = {};
	for (const name of Object.keys(spec.props)) {
		const prop = spec.props[name];
		switch (prop.Type) {
			case 'int':
			case 'playerIndex':
				node[name] = 0;
				break;
			case 'bool':
				node[name] = false;
				break;
			case 'string':
				node[name] = '';
				break;
			case 'enum': {
				const e = prop.Enum ? enums[prop.Enum] : undefined;
				if (!e) throw gameSpecError('property ' + name + ' names unknown enum ' + prop.Enum);
				node[name] = e.Values[String(e.DefaultValue)];
				break;
			}
			case 'timer':
				node[name] = { ID: '', IsTimer: true };
				break;
			case 'stack':
			case 'sizedStack':
				node[name] = emptyStack(prop);
				break;
			case 'mergedStack':
				// A view; EngineState.toJSON derives it from its parts.
				node[name] = { Deck: prop.Deck ?? '', Indexes: [], IDs: [], IDsLastSeen: {}, ShuffleCount: 0 };
				break;
			case 'board': {
				const spaces: unknown[] = [];
				for (let i = 0; i < (prop.Size ?? 0); i++) spaces.push(emptyStack({ ...prop, Type: 'stack' }));
				node[name] = { Spaces: spaces };
				break;
			}
			default:
				// A slice property starts as Go's nil slice, which serializes as null.
				node[name] = null;
		}
	}
	if (spec.initial) Object.assign(node, spec.initial(index));
	return node;
}

/** GameManager.emptyState: every property at its zero value, then the spec's initial values. */
export function emptyStateJSON(game: { spec: GameSpec; def: GameDefinition }, numPlayers: number): StateJSON {
	const spec = game.spec;
	const players: SubStateJSON[] = [];
	for (let i = 0; i < numPlayers; i++) players.push(emptySubState(spec.player, game.def.Enums, i));
	const components: Record<string, SubStateJSON[]> = {};
	const dynamic = spec.dynamicComponentValues ?? {};
	for (const deckName of Object.keys(dynamic)) {
		const values: SubStateJSON[] = [];
		for (let i = 0; i < game.def.Decks[deckName].Size; i++) values.push(emptySubState(dynamic[deckName], game.def.Enums, i));
		components[deckName] = values;
	}
	return { Game: emptySubState(spec.game, game.def.Enums, 0), Players: players, Components: components };
}

// ----- Go's JSON, for the fingerprints ------------------------------------

function goString(text: string): string {
	return JSON.stringify(text).replace(/</g, '\\u003c').replace(/>/g, '\\u003e').replace(/&/g, '\\u0026').replace(/\u2028/g, '\\u2028').replace(/\u2029/g, '\\u2029');
}

/** encoding/json for a value built from plain objects in struct-field order. */
function goJSON(value: unknown, indent: string, depth = 0): string {
	if (value === null || value === undefined) return 'null';
	if (typeof value === 'string') return goString(value);
	if (typeof value === 'number' || typeof value === 'boolean') return String(value);
	const open = indent === '' ? '' : '\n' + indent.repeat(depth + 1);
	const close = indent === '' ? '' : '\n' + indent.repeat(depth);
	const colon = indent === '' ? ':' : ': ';
	if (Array.isArray(value)) {
		if (value.length === 0) return '[]';
		return '[' + open + value.map(item => goJSON(item, indent, depth + 1)).join(',' + open) + close + ']';
	}
	const keys = Object.keys(value as object);
	if (keys.length === 0) return '{}';
	return '{' + open + keys.map(key => goString(key) + colon + goJSON((value as Record<string, unknown>)[key], indent, depth + 1)).join(',' + open) + close + '}';
}

// ----- moves --------------------------------------------------------------

const FIELD_TYPES: Record<MoveFieldSpec['type'], ValuePropDef['Type']> = {
	int: 'int',
	bool: 'bool',
	string: 'string',
	playerIndex: 'playerIndex',
	enum: 'enum',
};

const WIRE_TYPES: Record<MoveFieldSpec['type'], string> = { int: 'int', bool: 'bool', string: 'string', playerIndex: 'playerIndex', enum: 'enum' };

function inferCodec(field: MoveFieldSpec): InputCodec {
	switch (field.type) {
		case 'int':
			return 'integer';
		case 'bool':
			return 'boolean';
		case 'string':
			return 'string';
		case 'playerIndex':
			return 'player-index';
		default:
			return 'enum';
	}
}

/** moves.plainGameStackName: the property name when a stack path names a plain game stack. */
export function plainGameStackName(path: string | undefined): string | null {
	if (path === undefined || path === '') return null;
	if (path.indexOf('[') >= 0 || path.indexOf(']') >= 0) return null;
	if (path.startsWith('game.')) return path.slice(5) === '' ? null : path.slice(5);
	if (path.indexOf('.') >= 0) return null;
	return path;
}

function moveFields(spec: MoveSpec): Record<string, MoveFieldSpec> {
	const fields: Record<string, MoveFieldSpec> = {};
	if (spec.base === 'currentPlayer' || spec.base === 'anyPlayer') {
		fields.TargetPlayerIndex = { type: 'playerIndex', input: 'context-owned', codec: 'player-index' };
	}
	for (const name of Object.keys(spec.fields ?? {})) fields[name] = (spec.fields as Record<string, MoveFieldSpec>)[name];
	return fields;
}

function compileMove(
	spec: MoveSpec,
	index: number,
	shared: { enums: EnumRegistry; registry: CompiledGame['registry']; templates: Record<string, string>; example: RulesState },
): CompiledMove {
	const fields = moveFields(spec);
	const fieldNames = Object.keys(fields).sort();
	const schema: Record<string, ValuePropDef> = {};
	const fieldPolicies: Record<string, Record<string, PolicyName>> = {};
	const inputFields: InputField[] = [];
	for (const name of fieldNames) {
		const field = fields[name];
		schema[name] = field.enum ? { Type: FIELD_TYPES[field.type], Enum: field.enum } : { Type: FIELD_TYPES[field.type] };
		fieldPolicies[name] = parseSanitizeTag(field.sanitize ?? '', GROUP_ALL, 'hidden');
		for (const group of Object.keys(fieldPolicies[name])) {
			if (fieldPolicies[name][group] === 'invalid') {
				throw gameSpecError('Move ' + spec.name + ' property ' + name + ' had invalid sanitization policy for group ' + group);
			}
		}
		const disposition = field.input ?? 'required';
		inputFields.push({ name, disposition, codec: disposition === 'unsupported' ? (field.codec ?? '') : (field.codec ?? inferCodec(field)) });
	}

	let namePolicy: Record<string, PolicyName> = { [GROUP_ALL]: 'visible' };
	if (spec.nameSanitization !== undefined) {
		namePolicy = parseSanitizeTag(spec.nameSanitization, GROUP_ALL, 'hidden');
		for (const group of Object.keys(namePolicy)) {
			if (namePolicy[group] !== 'visible' && namePolicy[group] !== 'hidden') {
				throw gameSpecError('move name sanitization group ' + JSON.stringify(group) + ' must use visible or hidden');
			}
		}
	}

	const authored = spec.preconditions ?? [];
	const declared = authored.length > 0 || spec.planEnabled === true || spec.legalCustom !== undefined;
	const source = plainGameStackName(spec.sourceProperty);
	const destination = plainGameStackName(spec.destinationProperty);
	const base: LegalSpec[] = contributedPreconditions({
		moveName: spec.name,
		legalPhases: spec.legalPhases,
		hasMoveProgression: spec.progression !== undefined,
		stackConstraints: source !== null && destination !== null ? { source, destination } : null,
		currentPlayer: spec.base === 'currentPlayer',
	});
	const legalCustom = spec.legalCustom;
	const custom: LegalCustom | null =
		declared && legalCustom
			? ctx => {
					const { state, move } = rulesOf(ctx);
					return legalCustom(state, move, ctx.proposer);
				}
			: null;
	const exampleMove = shared.example.newMove(spec.name);
	const plan = buildLegalPlan({
		moveName: spec.name,
		contributed: declared && spec.contributes ? spec.contributes(base) : base,
		authored: declared ? authored : [],
		suppressions: declared ? (spec.suppress ?? []) : [],
		custom,
		registry: shared.registry,
		chest: bindLegalChest(shared.enums),
		exampleState: shared.example.legal,
		moveReader: bindLegalMove(exampleMove ? exampleMove.fields : {}, schema, shared.enums).reader(),
		templates: shared.templates,
	});
	return { spec, index, fields, fieldNames, schema, fieldPolicies, namePolicy, declared, plan, inputFields };
}

function inputSchema(moves: readonly CompiledMove[], enums: EnumRegistry): { name: string; fields: unknown[] | null }[] {
	const schema: { name: string; fields: unknown[] | null }[] = [];
	for (const move of moves) {
		if (move.spec.fixUp) continue;
		const fields: unknown[] = [];
		for (const input of move.inputFields) {
			const field = move.fields[input.name];
			const item: Record<string, unknown> = { name: input.name, wireType: WIRE_TYPES[field.type], disposition: input.disposition };
			if (input.codec !== '') item.codec = input.codec;
			if (field.type === 'enum') {
				const e = enums.get(field.enum);
				if (e !== null) {
					item.enumName = e.name;
					if (input.codec === 'enum') item.enumValues = e.keys().map(key => e.stringOf(key)).sort();
				}
			}
			fields.push(item);
		}
		schema.push({ name: move.spec.name, fields: fields.length === 0 ? null : fields });
	}
	return schema.sort((a, b) => (a.name < b.name ? -1 : a.name > b.name ? 1 : 0));
}

function choiceSchema(moves: readonly CompiledMove[], enums: EnumRegistry): ChoiceSchemaItem[] {
	const result: ChoiceSchemaItem[] = [];
	for (const move of moves) {
		const choice = move.spec.choices;
		if (!choice) continue;
		const name = move.spec.name;
		if (move.spec.fixUp) throw gameSpecError('fix-up move ' + JSON.stringify(name) + ' cannot publish projected choices');
		const fail = (reason: string): never => {
			throw gameSpecError('move ' + JSON.stringify(name) + ' choice projection: ' + reason);
		};
		const required = move.inputFields.filter(field => field.disposition === 'required');
		const input = move.inputFields.find(field => field.name === choice.field);
		if (!input) return fail('field ' + JSON.stringify(choice.field) + ' is not a configured creator input');
		if (required.length !== 1 || input.disposition !== 'required') {
			return fail('version one requires exactly one required creator field and it must be ' + JSON.stringify(choice.field) + '; got ' + required.length + ' required fields');
		}
		const exclude = choice.exclude ?? [];
		const item: ChoiceSchemaItem = { moveName: name, fieldName: choice.field, source: 'players', disclosure: 'actor-exact' };
		if (choice.stack) {
			if (input.codec !== 'integer') return fail('stack-slot source requires codec "integer", got ' + JSON.stringify(input.codec));
			if (exclude.length !== 0) return fail('stack-slot source does not support excluded values');
			item.source = 'stack-slots';
			item.stackSource = { scope: choice.stack.scope, property: choice.stack.property };
		} else if (input.codec === 'player-index') {
			if (exclude.length !== 0) return fail('player source does not support excluded values');
		} else if (input.codec === 'enum') {
			item.source = 'enum-values';
			const e = enums.get(move.fields[choice.field].enum);
			const values = e === null ? [] : e.keys().map(key => e.stringOf(key)).sort();
			for (const value of exclude) if (values.indexOf(value) < 0) return fail('excluded enum value ' + JSON.stringify(value) + ' is not canonical');
			item.candidateValues = values.filter(value => exclude.indexOf(value) < 0);
			if (item.candidateValues.length === 0) return fail('excluded enum values remove the entire candidate universe');
		} else {
			return fail('field ' + JSON.stringify(choice.field) + ' uses unsupported choice codec ' + JSON.stringify(input.codec) + '; choices require player-index or ordinary enum input');
		}
		result.push(item);
	}
	return result.sort((a, b) => (a.moveName < b.moveName ? -1 : a.moveName > b.moveName ? 1 : 0));
}

/** FingerprintMoveChoiceProjectionSchema. */
function choiceFingerprint(schema: readonly ChoiceSchemaItem[]): string {
	const projections = schema.map(item => {
		const out: Record<string, unknown> = { moveName: item.moveName, fieldName: item.fieldName, source: item.source };
		if (item.stackSource) out.stackSource = { scope: item.stackSource.scope, property: item.stackSource.property };
		if (item.candidateValues && item.candidateValues.length > 0) out.candidateValues = item.candidateValues.slice().sort();
		out.disclosure = item.disclosure;
		return out;
	});
	// Go fingerprints a copy made by appending to a nil slice, so a game with
	// no projections hashes `"projections":null`, not an empty list.
	return 'sha256:' + sha256Hex(goJSON({ version: 1, projections: projections.length === 0 ? null : projections }, ''));
}

// ----- variants -----------------------------------------------------------

/** strings.Title for the plain words variant names are. */
function title(text: string): string {
	return text.replace(/(^|[^A-Za-z0-9_'])([a-z])/g, (_all, before: string, letter: string) => before + letter.toUpperCase());
}

function displayName(name: string, given: string | undefined): string {
	if (given) return given;
	return title(name.replace(/-/g, ' ').replace(/_/g, ' ').split(/\s+/).filter(word => word !== '').join(' '));
}

/** GameManager.Variants(): the config with every name and display name filled in. */
export function variantConfig(spec: GameSpec): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	const variants = spec.variants ?? {};
	for (const key of Object.keys(variants)) {
		const config: VariantKeySpec = variants[key];
		const values: Record<string, unknown> = {};
		for (const valueName of Object.keys(config.values)) {
			const value = config.values[valueName];
			values[valueName] = { Description: value.description ?? '', DisplayName: displayName(valueName, value.displayName), Name: valueName };
		}
		out[key] = { Default: config.default ?? '', Description: config.description ?? '', DisplayName: displayName(key, config.displayName), Name: key, Values: values };
	}
	return out;
}

/** VariantConfig.NewVariant: the given values with defaults filled in; null when empty. Throws with Go's text. */
export function newVariant(spec: GameSpec, given: Readonly<Record<string, string>> | null | undefined): Record<string, string> | null {
	const result: Record<string, string> = { ...(given ?? {}) };
	const variants = spec.variants ?? {};
	for (const key of Object.keys(variants)) {
		if (result[key] !== undefined && result[key] !== '') continue;
		const fallback = variants[key].default ?? '';
		if (fallback === '') continue;
		result[key] = fallback;
	}
	for (const key of Object.keys(result).sort()) {
		const config = variants[key];
		if (!config) throw gameSpecError('configuration had a property called ' + key + " that isn't expected");
		if (!config.values[result[key]]) {
			throw gameSpecError("configuration's " + displayName(key, config.displayName) + " property had a value that wasn't allowed: " + result[key]);
		}
	}
	return Object.keys(result).length === 0 ? null : result;
}

// ----- the game -----------------------------------------------------------

function subStateDef(spec: SubStateSpec): SubStateDef {
	return { Props: spec.props };
}

/** base.GameDelegate.DisplayName: strings.Title of the name. */
export function gameDisplayName(spec: GameSpec): string {
	return spec.displayName ?? title(spec.name);
}

export function compileGame(spec: GameSpec): CompiledGame {
	const decks: Record<string, DeckDef> = {};
	const components: Record<string, Readonly<Record<string, ValuePropDef>>> = {};
	for (const deckName of Object.keys(spec.decks)) {
		decks[deckName] = { Size: spec.decks[deckName].components.length, Components: spec.decks[deckName].components };
		if (spec.decks[deckName].props) components[deckName] = spec.decks[deckName].props as Readonly<Record<string, ValuePropDef>>;
	}
	const chestEnums = spec.enums ?? {};
	const componentEnums = spec.componentEnums ?? {};
	const dynamic: Record<string, SubStateDef> = {};
	for (const deckName of Object.keys(spec.dynamicComponentValues ?? {})) {
		dynamic[deckName] = subStateDef((spec.dynamicComponentValues as Record<string, SubStateSpec>)[deckName]);
	}
	const def: GameDefinition = {
		Name: spec.name,
		Decks: decks,
		Enums: { ...componentEnums, ...chestEnums },
		GroupEnum: chestEnums.group ? 'group' : '',
		Game: subStateDef(spec.game),
		Player: subStateDef(spec.player),
		DynamicComponentValues: dynamic,
	};
	const enums = newEnumRegistry(chestEnums, componentEnums);
	const phaseEnum = enums.get(chestEnums.phase ? 'phase' : 'Phase');
	const templates = mergeLegalTemplates(spec.legalTemplates);
	const registry = defaultLegalRegistry();

	const names: Record<string, boolean> = {};
	const noProgressionMoves: Record<string, boolean> = {};
	for (const move of spec.moves) {
		if (move.name === '') throw gameSpecError('No name provided');
		if (names[move.name.toLowerCase()]) throw gameSpecError('a move named ' + move.name + ' is configured more than once');
		names[move.name.toLowerCase()] = true;
		if (!move.progression) noProgressionMoves[move.name] = true;
	}

	// A game under construction, enough to make the example state the plans
	// are validated against (Go: manager.ExampleState()).
	const moves: CompiledMove[] = [];
	const movesByName: Record<string, CompiledMove> = {};
	const canonicalMoveNames: Record<string, string> = {};
	for (const move of spec.moves) canonicalMoveNames[move.name.toLowerCase()] = move.name;
	const game: CompiledGame = {
		spec,
		def,
		enums,
		components: components as ComponentSchema,
		moves,
		movesByName,
		canonicalMoveNames,
		phaseEnum,
		templates,
		registry,
		noProgressionMoves,
		moveInputSchema: [],
		moveInputSchemaFingerprint: '',
		choiceSchema: [],
		choiceSchemaFingerprint: '',
	};
	const exampleRecord: GameRecord = { name: spec.name, id: '', secretSalt: '', numPlayers: spec.defaultNumPlayers, variant: null, agents: [], seating: false };
	const exampleEngine = loadEngineState(def, { gameId: '', secretSalt: '' }, 0, emptyStateJSON(game, spec.defaultNumPlayers), { skipConservation: true });
	syncTimerRecords(exampleEngine);
	const example = newRulesState(game, exampleEngine, exampleRecord, [], { pendingSeat: null });

	// Fields first, so a move's defaults can be computed while plans compile.
	for (let i = 0; i < spec.moves.length; i++) {
		const fields = moveFields(spec.moves[i]);
		// A whole record with nothing compiled yet: no schema, no policies and
		// an empty plan. The second pass replaces every one.
		movesByName[spec.moves[i].name] = {
			spec: spec.moves[i],
			index: i,
			fields,
			fieldNames: Object.keys(fields).sort(),
			schema: {},
			fieldPolicies: {},
			namePolicy: {},
			declared: false,
			plan: { moveName: spec.moves[i].name, steps: [], custom: null, allReads: [], specs: [] },
			inputFields: [],
		};
	}
	for (let i = 0; i < spec.moves.length; i++) {
		const compiled = compileMove(spec.moves[i], i, { enums, registry, templates, example });
		moves.push(compiled);
		movesByName[compiled.spec.name] = compiled;
	}

	const schema = inputSchema(moves, enums);
	const choices = choiceSchema(moves, enums);
	const mutable = game as { -readonly [K in keyof CompiledGame]: CompiledGame[K] };
	mutable.moveInputSchema = schema;
	mutable.moveInputSchemaFingerprint = 'sha256:' + sha256Hex(goJSON(schema, '  '));
	mutable.choiceSchema = choices;
	mutable.choiceSchemaFingerprint = choiceFingerprint(choices);
	return game;
}
