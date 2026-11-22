// The stock moves and behaviors of rules/stock.ts against real Go: the
// conformance documents in testdata/stock/ (from testdata/gen/stockvec) are
// games of testdata/gen/stockgame and testdata/gen/seatgame played on the Go
// engine. Each is replayed here on a GameSpec whose stock moves come from
// stockMove(type, config) with the configuration keys exactly as Go stores
// them, so the registry is what is proven. At every version the stored
// state, the move record and every viewer's view, legality (Go's error
// texts) and choices must equal Go's.
//
// Covered: every deal and collect move, every move-components move,
// ShuffleStack, StartPhase, DefaultComponent (with and without a legal
// type), FinishTurn, Done, NoOp, SeatPlayer, InactivateEmptySeat,
// ActivateFilledSeat, ActivateEmptySeat, CloseEmptySeat, CloseAllSeats, the
// PlayerColor and ComponentColor behaviors and the RoundRobin properties.
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import { type ConformanceDocument, FIELDS, passTable, runDocument, type DocumentResult } from '../testing/conformance.ts';
import { ADMIN_PLAYER_INDEX } from '../legal/types.ts';
import type { StateJSON } from '../state.ts';
import { compileGame } from './compile.ts';
import { loadState } from './engine.ts';
import type { RulesState } from './state.ts';
import {
	addForPhase, addOrderedForPhase, combine, add, componentColorProps, currentPlayerProps, inactivePlayerProps, phaseProps, playerColorProps,
	roundRobinProps, seatProps,
} from './stock.ts';
import { CONFIG_PREFIX, stockMove } from './stockRegistry.ts';
import type { GameRecord, GameSpec, LegalResult, MoveInstance, MoveRecord, MoveSpec } from './types.ts';

const STOCK = fileURLToPath(new URL('../testdata/stock/', import.meta.url));

/** A configuration map with Go's fully qualified keys. */
function config(short: Record<string, unknown>): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	for (const key of Object.keys(short)) out[CONFIG_PREFIX + key] = short[key];
	return out;
}

/** The stock move of that Go type, as the loader builds it, with the game's own parts laid over it. */
function stock(type: string, short: Record<string, unknown> = {}, own: Partial<MoveSpec> = {}): MoveSpec {
	const parts = stockMove('moves.' + type, config(short));
	assert.ok(parts !== undefined, type + ' is not in the registry');
	const spec = { ...parts.spec, ...own } as MoveSpec;
	if (own.apply === undefined) {
		assert.ok(parts.apply !== undefined, type + ' has no stock apply');
		spec.apply = parts.apply;
	}
	return spec;
}

// ----- stockgame ----------------------------------------------------------

const PHASE_DEAL = 0;
const PHASE_PLAY = 1;
const PHASE_CLEANUP = 2;
const COLORS = ['Red', 'Blue', 'Green'];

/** The board slot holding card `index`, or -1 (stockgame's boardSlot). */
function boardSlot(state: RulesState, index: number): number {
	const board = state.gameView('Board');
	for (let i = 0; i < board.len(); i++) if (board.componentAt(i) === index) return i;
	return -1;
}

/** stockgame's card.Legal. */
function cardLegal(state: RulesState, index: number, legalType: number | null): string | null {
	const slot = boardSlot(state, index);
	if (slot < 0) return 'not on the board';
	const dyn = state.dynamic('cards', index);
	if (legalType !== null && legalType === 1) {
		if (slot !== 0) return 'not in the corner';
		if (dyn.Corner === true) return 'already marked as the corner';
		return null;
	}
	if (slot % 2 !== 0) return 'on an odd slot';
	if (dyn.Crowned === true) return 'already crowned';
	return null;
}

function markApply(prop: 'Crowned' | 'Corner') {
	return (state: RulesState, move: MoveInstance): void => {
		const index = state.gameStack('Board').componentAt(move.fields.ComponentIndex as number);
		if (index === null) throw new Error('no card there');
		state.dynamic('cards', index)[prop] = true;
	};
}

function placeCard(): MoveSpec {
	return stock('CurrentPlayer', { MoveName: 'Place Card', HelpText: 'Places the first card of your hand on the board' }, {
		fields: { Slot: { type: 'int' } },
		defaults(state, move) {
			const board = state.gameView('Board');
			for (let i = 0; i < board.len(); i++) {
				if (board.componentAt(i) === null) {
					move.fields.Slot = i;
					return;
				}
			}
		},
		legal(state, move, _proposer, base): LegalResult {
			const err = base();
			if (err !== null) return err;
			const target = move.fields.TargetPlayerIndex as number;
			if (state.players[target].FinishedTurn === true) return 'you already placed a card this turn';
			const hand = state.playerStack(target, 'Hand');
			if (hand.numComponents() === 0) return 'your hand is empty';
			return hand.moveComponentError(hand.firstComponentIndex(), state.gameStack('Board'), move.fields.Slot as number);
		},
		apply(state, move) {
			const target = move.fields.TargetPlayerIndex as number;
			const hand = state.playerStack(target, 'Hand');
			hand.moveComponent(hand.firstComponentIndex(), state.gameStack('Board'), move.fields.Slot as number);
			state.players[target].FinishedTurn = true;
		},
	});
}

function callCleanup(): MoveSpec {
	return stock('CurrentPlayer', { MoveName: 'Call Cleanup', HelpText: 'Ends play once the board is full' }, {
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			if (state.gameStack('Board').slotsRemaining() > 0) return 'the board is not full';
			return null;
		},
		apply: state => state.setPhase(PHASE_CLEANUP),
	});
}

export function stockgame(): GameSpec {
	const cards: Record<string, unknown>[] = [];
	for (let i = 0; i < 16; i++) cards.push({ Color: COLORS[i % 3], Rank: i + 1 });
	const deal = { GameProperty: 'DrawStack', PlayerPropety: 'Hand' };
	return {
		name: 'stockgame',
		displayName: 'Stock Game',
		minNumPlayers: 2,
		maxNumPlayers: 3,
		defaultNumPlayers: 2,
		decks: { cards: { components: cards, props: { ...componentColorProps, Rank: { Type: 'int' } }, componentLegal: cardLegal } },
		enums: {
			phase: { Values: { 0: 'Deal', 1: 'Play', 2: 'Cleanup' }, DefaultValue: 0 },
			color: { Values: { 0: 'Red', 1: 'Blue', 2: 'Green' }, DefaultValue: 0 },
			legaltype: { Values: { 0: 'None', 1: 'Corner' }, DefaultValue: 0 },
		},
		game: {
			props: {
				...currentPlayerProps,
				...phaseProps,
				...roundRobinProps,
				DrawStack: { Type: 'stack', Deck: 'cards' },
				DiscardStack: { Type: 'stack', Deck: 'cards' },
				Spare: { Type: 'stack', Deck: 'cards' },
				Board: { Type: 'sizedStack', Deck: 'cards', Size: 6 },
			},
		},
		player: {
			props: { ...playerColorProps, Hand: { Type: 'stack', Deck: 'cards' }, FinishedTurn: { Type: 'bool' } },
			traits: {
				color: true,
				turn: {
					turnDone: (state, player) => (state.players[player].FinishedTurn === true ? null : 'the player has not placed a card'),
					resetForTurnStart(state, player) {
						state.players[player].FinishedTurn = false;
					},
					resetForTurnEnd: () => undefined,
				},
			},
		},
		dynamicComponentValues: { cards: { props: { Crowned: { Type: 'bool' }, Corner: { Type: 'bool' } } } },
		moves: combine(
			addOrderedForPhase(
				PHASE_DEAL,
				stock('DealCountComponents', { ...deal, TargetCount: 2 }),
				stock('DealComponentsUntilPlayerCountReached', { ...deal, TargetCount: 3 }),
				stock('CollectCountComponents', { GameProperty: 'DiscardStack', PlayerPropety: 'Hand' }),
				stock('CollectComponentsUntilPlayerCountLeft', { GameProperty: 'DiscardStack', PlayerPropety: 'Hand', TargetCount: 1 }),
				stock('MoveCountComponents', { SourceProperty: 'DrawStack', DestinationProperty: 'Spare', TargetCount: 2 }),
				stock('MoveComponentsUntilCountReached', { SourceProperty: 'DrawStack', DestinationProperty: 'Spare', TargetCount: 4 }),
				stock('MoveComponentsUntilCountLeft', { SourceProperty: 'Spare', DestinationProperty: 'DiscardStack', TargetCount: 1 }),
				stock('ShuffleStack', { SourceProperty: 'DiscardStack' }),
				stock('MoveAllComponents', { SourceProperty: 'DiscardStack', DestinationProperty: 'DrawStack' }),
				stock('DealComponentsUntilGameCountLeft', { ...deal, TargetCount: 7 }),
				stock('CollectComponentsUntilGameCountReached', { ...deal, TargetCount: 9 }),
				stock('StartPhase', { StartPhase: PHASE_PLAY, StartPhaseEnum: { Values: { 0: 'Deal', 1: 'Play', 2: 'Cleanup' } } }),
			),
			addForPhase(
				PHASE_PLAY,
				stock('DefaultComponent', { SourceProperty: 'Board', MoveName: 'Crown Card', HelpText: 'Crowns a card on an even slot' }, { apply: markApply('Crowned') }),
				stock(
					'DefaultComponent',
					{ SourceProperty: 'Board', LegalType: 1, LegalTypeEnum: 'legaltype', MoveName: 'Mark Corner', HelpText: 'Marks the card in the corner' },
					{ apply: markApply('Corner') },
				),
				stock('FinishTurn'),
				placeCard(),
				callCleanup(),
				stock('Done', { MoveName: 'Pass', HelpText: 'Does nothing' }),
			),
			addOrderedForPhase(
				PHASE_CLEANUP,
				stock('DealAllComponents', deal),
				stock('CollectAllComponents', { GameProperty: 'DiscardStack', PlayerPropety: 'Hand' }),
				stock('NoOp'),
			),
		),
		setUp: {
			distribute: state => state.gameStack('DrawStack'),
			finish(state) {
				for (let i = 0; i < state.numPlayers(); i++) state.players[i].Color = COLORS[(i + 1) % 3];
				state.gameStack('DrawStack').shuffle();
			},
		},
		end: {
			conditionMet(state) {
				if (state.currentPhaseKey() !== PHASE_CLEANUP) return false;
				if (state.gameStack('DrawStack').numComponents() > 0) return false;
				for (let i = 0; i < state.numPlayers(); i++) if (state.playerStack(i, 'Hand').numComponents() > 0) return false;
				return true;
			},
		},
	};
}

// ----- seatgame -----------------------------------------------------------

const PHASE_GATHERING = 0;
const PHASE_OPEN = 1;
const PHASE_DONE = 2;

function everySeatFilledOrClosed(state: RulesState): boolean {
	for (const p of state.players) if (p.SeatFilled !== true && p.SeatClosed !== true) return false;
	return true;
}

function seatFixUp(name: string, helpText: string, check: (state: RulesState) => string | null, apply: (state: RulesState) => void): MoveSpec {
	return stock('FixUp', { MoveName: name, HelpText: helpText }, {
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			return check(state);
		},
		apply,
	});
}

export function seatgame(): GameSpec {
	return {
		name: 'seatgame',
		displayName: 'Seat Game',
		minNumPlayers: 2,
		maxNumPlayers: 4,
		defaultNumPlayers: 4,
		decks: { tokens: { components: [{}, {}] } },
		enums: { phase: { Values: { 0: 'Gathering', 1: 'Open', 2: 'Done' }, DefaultValue: 0 } },
		game: { props: { ...currentPlayerProps, ...phaseProps, Tokens: { Type: 'stack', Deck: 'tokens' }, Reopened: { Type: 'bool' } } },
		player: { props: { ...seatProps, ...inactivePlayerProps }, traits: { seat: true, inactive: true } },
		moves: combine(
			add(stock('SeatPlayer'), stock('ActivateFilledSeat')),
			addForPhase(
				PHASE_GATHERING,
				stock('InactivateEmptySeat'),
				stock('CloseAllSeats', { TargetCount: 2 }),
				seatFixUp('Begin', 'Enters Open once no seat is open', state => (everySeatFilledOrClosed(state) ? null : 'a seat is still open'), state => state.setPhase(PHASE_OPEN)),
			),
			addForPhase(
				PHASE_OPEN,
				stock('ActivateEmptySeat'),
				stock('CloseEmptySeat'),
				seatFixUp(
					'Reopen',
					'Opens the empty seats again',
					state => {
						if (state.g.Reopened === true) return 'already reopened';
						for (const p of state.players) if (p.PlayerInactive === true) return 'a seat is still inactive';
						return null;
					},
					state => {
						state.g.Reopened = true;
						for (const p of state.players) if (p.SeatFilled !== true) p.SeatClosed = false;
					},
				),
				seatFixUp(
					'Finish',
					'Enters Done',
					state => {
						if (state.g.Reopened !== true) return 'not reopened yet';
						return everySeatFilledOrClosed(state) ? null : 'a seat is still open';
					},
					state => state.setPhase(PHASE_DONE),
				),
			),
		),
		setUp: { distribute: state => state.gameStack('Tokens') },
		end: { conditionMet: state => state.currentPhaseKey() === PHASE_DONE },
	};
}

// ----- the replay ----------------------------------------------------------

const index = JSON.parse(readFileSync(STOCK + 'index.json', 'utf8')) as { file: string; game: string; name: string; versions: number }[];
const games = { stockgame: compileGame(stockgame()), seatgame: compileGame(seatgame()) };
const results: DocumentResult[] = [];

test('the stock vectors cover both games', () => {
	assert.ok(index.some(entry => entry.game === 'stockgame'));
	assert.ok(index.some(entry => entry.game === 'seatgame'));
	assert.ok(index.length >= 5);
});

for (const entry of index) {
	test('stock conformance: ' + entry.file, () => {
		const doc = JSON.parse(gunzipSync(readFileSync(STOCK + entry.file)).toString('utf8')) as ConformanceDocument;
		const result = runDocument(games[entry.game as keyof typeof games], doc);
		result.name = entry.file.replace(/\.json\.gz$/, '');
		results.push(result);
		const problems: string[] = [];
		if (result.fatal) problems.push('replay stopped: ' + result.fatal);
		for (const field of FIELDS) {
			const f = result.fields[field];
			if (f.mismatches > 0) problems.push(field + ': ' + f.mismatches + ' of ' + f.compared + ' differ, e.g.\n    ' + f.examples.join('\n    '));
		}
		assert.equal(problems.length, 0, '\n  ' + problems.join('\n  '));
		assert.equal(result.fields.state.compared, doc.versions.length, 'every version was compared');
		assert.equal(result.fields['view.game'].compared, doc.versions.length * doc.viewers.length, 'every viewer of every version was compared');
	});
}

test('stock conformance pass table', t => {
	t.diagnostic('\n' + passTable(results));
	assert.equal(results.length, index.length);
});

// ----- fix-up legality ------------------------------------------------------
//
// A fix-up is in no viewer's legality, so a conformance document shows only
// whether it applied, never its error text. testdata/stock/fixups.json
// records, at every version of every document, each fix-up bound to that
// state and Legal's text for the admin ('' for legal).

const fixups = JSON.parse(readFileSync(STOCK + 'fixups.json', 'utf8')) as Record<string, Record<string, string>[]>;

test('stock fix-up legality: every fix-up at every version, in Go words', () => {
	let compared = 0;
	for (const entry of index) {
		const doc = JSON.parse(gunzipSync(readFileSync(STOCK + entry.file)).toString('utf8')) as ConformanceDocument;
		const game = games[entry.game as keyof typeof games];
		const rows = fixups[entry.file];
		assert.equal(rows.length, doc.versions.length, entry.file);
		const record: GameRecord = {
			name: doc.game.name,
			id: doc.game.id,
			secretSalt: doc.game.secretSalt,
			numPlayers: doc.game.numPlayers,
			variant: Object.keys(doc.game.variant).length === 0 ? null : doc.game.variant,
			agents: doc.game.agents,
			seating: doc.game.seating,
		};
		const history: MoveRecord[] = [];
		for (const [v, want] of rows.entries()) {
			const version = doc.versions[v] as unknown as { move: null | { name: string; version: number; initiator: number; phase: number; proposer: number; arguments: Record<string, unknown> }; state: StateJSON };
			if (version.move !== null) {
				const m = version.move;
				history.push({ Name: m.name, Version: m.version, Initiator: m.initiator, Phase: m.phase, Proposer: m.proposer, Blob: m.arguments });
			}
			// Version 0 is judged against the game's whole history in Go
			// (Game.MoveRecords treats upToVersion < 1 as the head), the
			// engine's divergence 2 (fabric/status/engine-rules.md).
			if (v === 0) continue;
			const state = loadState(game, { record, history: history.slice(), state: version.state });
			for (const name of Object.keys(want)) {
				const move = state.newMove(name);
				assert.ok(move !== null, name);
				let got: string;
				try {
					got = state.legalError(move, ADMIN_PLAYER_INDEX) ?? '';
				} catch (thrown) {
					got = thrown instanceof Error ? thrown.message : String(thrown);
				}
				assert.equal(got, want[name], entry.file + ' v' + v + ' ' + name);
				compared++;
			}
		}
	}
	assert.ok(compared > 2000, 'compared ' + compared);
});
