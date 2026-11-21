// The legal evaluator, run through the REAL engine state.
//
// ../legal's own tests bind the evaluator to a stand-in (legal/testing/
// fixtures.ts) that models stacks well enough for tests. Here the same Go
// answers are required with rules/binding.ts in its place: every fixture is
// loaded as an EngineState, so stack questions are answered by ../stack.ts.
//
// Proved, against Go:
//   - every case of legal/testdata/conformance (the cross-language corpus);
//   - every one of the Go predicate verdicts in legal/testdata/vectors.json,
//     in full (outcome, template, bindings, unknown reason).
import assert from 'node:assert/strict';
import { existsSync, readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import type { GameDefinition, PropDef, PropType, SubStateDef } from '../definition.ts';
import type { LegalContext, LegalFramework } from '../legal/access.ts';
import { defaultLegalRegistry } from '../legal/catalog.ts';
import { goQuote } from '../legal/goformat.ts';
import { type LegalPredicate, LegalSpecError, evalLegalPredicate, resolveLegalSpecs } from '../legal/predicate.ts';
import type { WireFixture, WireGame, WireSubState } from '../legal/testing/fixtures.ts';
import type { LegalOutcome, LegalSpec, LegalVerdict, PropertyType } from '../legal/types.ts';
import { EngineState } from '../state.ts';
import type { StateJSON } from '../state.ts';
import { type ComponentSchema, type ValuePropDef, bindLegalMove, bindLegalState } from './binding.ts';
import { EnumRegistry } from './enums.ts';

function repoRoot(): string {
	let dir = path.dirname(fileURLToPath(import.meta.url));
	for (;;) {
		if (existsSync(path.join(dir, 'legal', 'testdata', 'conformance'))) return dir;
		const parent = path.dirname(dir);
		assert.notEqual(parent, dir, 'could not find the repository root');
		dir = parent;
	}
}

const root = repoRoot();
const corpusDir = path.join(root, 'legal', 'testdata', 'conformance');
const testdata = path.join(root, 'server', 'static', 'src', 'engine', 'legal', 'testdata');

const exported = JSON.parse(readFileSync(path.join(testdata, 'fixtures.json'), 'utf8')) as {
	games: Record<string, WireGame>;
	fixtures: Record<string, WireFixture>;
};

const SCALARS: Partial<Record<PropertyType, PropType>> = {
	TypeInt: 'int',
	TypeBool: 'bool',
	TypeString: 'string',
	TypePlayerIndex: 'playerIndex',
	TypeEnum: 'enum',
	TypeIntSlice: 'intSlice',
	TypeBoolSlice: 'boolSlice',
	TypeStringSlice: 'stringSlice',
	TypePlayerIndexSlice: 'playerIndexSlice',
	TypeEnumSlice: 'enumSlice',
	TypeTimer: 'timer',
	TypeBoard: 'board',
};

// A state record does not say which stacks a merged stack views. These are
// the merged stacks of the fixture games (examples/*/state.go).
const MERGED: Record<string, { Overlap: boolean; Stacks: string[] }> = {
	Cards: { Overlap: true, Stacks: ['VisibleCards', 'HiddenCards'] },
	Hand: { Overlap: false, Stacks: ['HiddenHand', 'VisibleHand'] },
};

function subStateDef(meta: WireSubState, values: Record<string, unknown>): SubStateDef {
	const props: Record<string, PropDef> = {};
	for (const name of Object.keys(meta.props)) {
		const wire = meta.props[name];
		if (wire.type === 'TypeStack') {
			const json = values[name] as { Deck: string; Size?: number; MaxSize?: number };
			const kind = meta.stacks?.[name]?.kind ?? (json.Size !== undefined ? 'sized' : 'growable');
			if (kind === 'merged') {
				assert.ok(MERGED[name], 'unknown merged stack ' + name);
				props[name] = { Type: 'mergedStack', Deck: json.Deck, ...MERGED[name] };
			} else if (kind === 'sized') {
				props[name] = { Type: 'sizedStack', Deck: json.Deck, Size: json.Size ?? 0 };
			} else {
				props[name] = { Type: 'stack', Deck: json.Deck, MaxSize: json.MaxSize ?? 0 };
			}
			continue;
		}
		const type = SCALARS[wire.type];
		assert.ok(type, 'fixture property type ' + wire.type);
		props[name] = wire.enum ? { Type: type, Enum: wire.enum } : { Type: type };
	}
	return { Props: props };
}

function inferredDef(values: Record<string, unknown>): SubStateDef {
	const props: Record<string, PropDef> = {};
	for (const name of Object.keys(values)) {
		props[name] = { Type: typeof values[name] === 'boolean' ? 'bool' : typeof values[name] === 'string' ? 'string' : 'int' };
	}
	return { Props: props };
}

interface Bound {
	ctx: (proposer: number) => LegalContext;
}

const bound = new Map<string, Bound>();

/** Loads one exported fixture as a real EngineState and binds the evaluator to it. */
function bind(name: string): Bound {
	const cached = bound.get(name);
	if (cached) return cached;
	const fixture = exported.fixtures[name];
	assert.ok(fixture, `fixture ${name} is not in testdata/fixtures.json`);
	const game = exported.games[fixture.game];

	const enumDefs: Record<string, { Values: Record<string, string>; DefaultValue: number }> = {};
	for (const [enumName, e] of Object.entries(game.chest.Enums ?? {})) enumDefs[enumName] = { Values: e.Values, DefaultValue: 0 };
	const otherDefs: Record<string, { Values: Record<string, string>; DefaultValue: number }> = {};
	for (const [enumName, values] of Object.entries(game.otherEnums ?? {})) otherDefs[enumName] = { Values: values, DefaultValue: 0 };
	const enums = new EnumRegistry(enumDefs, otherDefs);

	const decks: GameDefinition['Decks'] = {};
	const components: Record<string, Record<string, ValuePropDef>> = {};
	for (const [deckName, entries] of Object.entries(game.chest.Decks ?? {})) {
		(decks as Record<string, unknown>)[deckName] = { Size: entries.length, Components: entries.map(entry => entry.Values) };
		const schema: Record<string, ValuePropDef> = {};
		for (const [propName, prop] of Object.entries(game.componentProps[deckName] ?? {})) {
			const type = SCALARS[prop.type];
			assert.ok(type, 'component property type ' + prop.type);
			schema[propName] = prop.enum ? { Type: type, Enum: prop.enum } : { Type: type };
		}
		components[deckName] = schema;
	}

	const players = fixture.state.Players ?? [];
	const stateJSON = { ...fixture.state, Players: players } as unknown as StateJSON;
	const dynamic: Record<string, SubStateDef> = {};
	for (const [deckName, values] of Object.entries(stateJSON.Components ?? {})) {
		if (values.length > 0) dynamic[deckName] = inferredDef(values[0]);
	}
	const def: GameDefinition = {
		Name: fixture.game,
		Decks: decks,
		Enums: { ...otherDefs, ...enumDefs },
		Game: subStateDef(fixture.gameState, fixture.state.Game),
		Player: players.length > 0 ? subStateDef((fixture.playerStates ?? [])[0], players[0]) : { Props: {} },
		DynamicComponentValues: dynamic,
	};

	const state = EngineState.load(def, { gameId: 'fixture', secretSalt: 'salt' }, 0, stateJSON, { skipConservation: true });
	const legalState = bindLegalState(
		state,
		{ enums, components: components as ComponentSchema },
		{
			currentPlayerIndex: () => fixture.currentPlayerIndex,
			playerMayBeActive: index => (fixture.playersMayBeActive ?? [])[index] ?? true,
			customPlayerOrder: () => fixture.customPlayerOrder,
			playerIsInactive: index => (fixture.playersInactive ?? [])[index] ?? false,
		},
	);

	let move: LegalContext['move'] = null;
	if (fixture.move !== null) {
		const fields: Record<string, unknown> = {};
		const schema: Record<string, ValuePropDef> = {};
		for (const [fieldName, prop] of Object.entries(fixture.move.props)) {
			fields[fieldName] = prop.value;
			const type = SCALARS[prop.type];
			assert.ok(type, 'move property type ' + prop.type);
			schema[fieldName] = prop.enum ? { Type: type, Enum: prop.enum } : { Type: type };
		}
		move = bindLegalMove(fields, schema, enums);
	}
	const framework: LegalFramework = { currentPhase: () => fixture.phase };
	const result: Bound = { ctx: proposer => ({ state: legalState, move, proposer, framework }) };
	bound.set(name, result);
	return result;
}

const registry = defaultLegalRegistry();

// ----- the conformance corpus ---------------------------------------------

interface ConformanceFile {
	predicate: string;
	cases: { spec: LegalSpec; fixture: string; proposer: number; expect: LegalOutcome; template?: string }[];
}

const corpusFiles = readdirSync(corpusDir)
	.filter(name => name.endsWith('.json'))
	.sort();

test('the legal corpus passes through the real engine binding', () => {
	assert.ok(corpusFiles.length >= 32, 'corpus files found: ' + corpusFiles.length);
	let cases = 0;
	for (const fileName of corpusFiles) {
		const file = JSON.parse(readFileSync(path.join(corpusDir, fileName), 'utf8')) as ConformanceFile;
		file.cases.forEach((c, i) => {
			const [predicate] = resolveLegalSpecs([c.spec], registry);
			const verdict = evalLegalPredicate(predicate, bind(c.fixture).ctx(c.proposer));
			const where = `${fileName} case ${i} (${c.fixture} as ${c.proposer}): ${JSON.stringify(verdict)}`;
			assert.equal(verdict.outcome, c.expect, where);
			if (c.template) assert.equal(verdict.message?.template, c.template, where);
			cases++;
		});
	}
	assert.ok(cases >= 130, 'corpus cases run: ' + cases);
});

// ----- the Go predicate vectors -------------------------------------------

interface GoVerdict {
	outcome: string;
	message?: { template: string; bindings?: Record<string, string | number | boolean> };
	reason?: string;
	panic?: string;
}

const vectors = JSON.parse(readFileSync(path.join(testdata, 'vectors.json'), 'utf8')) as {
	specs: LegalSpec[];
	verdicts: GoVerdict[];
	vectors: [source: string, spec: number, fixture: string, proposer: number, verdict: number][];
};

function resolveDirect(spec: LegalSpec): LegalPredicate {
	const ctor = registry.get(spec.name);
	if (ctor === undefined) throw new LegalSpecError(`legal: unknown predicate name ${goQuote(spec.name)}`);
	return ctor.construct(spec, null, resolveDirect);
}

function normalize(v: LegalVerdict): GoVerdict {
	const out: GoVerdict = { outcome: v.outcome };
	if (v.message) {
		out.message = { template: v.message.template };
		if (v.message.bindings && Object.keys(v.message.bindings).length > 0) out.message.bindings = v.message.bindings;
	}
	if (v.reason) out.reason = v.reason;
	return out;
}

test('every Go predicate verdict is reproduced through the real engine binding', () => {
	const predicates = vectors.specs.map(resolveDirect);
	assert.ok(vectors.vectors.length > 3000, `only ${vectors.vectors.length} vectors`);
	for (const [source, specIndex, fixtureName, proposer, verdictIndex] of vectors.vectors) {
		const want = vectors.verdicts[verdictIndex];
		const where = `${source}: ${JSON.stringify(vectors.specs[specIndex])} on ${fixtureName} as ${proposer}`;
		let got: GoVerdict;
		try {
			got = normalize(predicates[specIndex].evaluate(bind(fixtureName).ctx(proposer)));
		} catch (err) {
			got = { outcome: '', panic: err instanceof Error ? err.message : String(err) };
		}
		assert.deepEqual(got, want, where);
	}
});
