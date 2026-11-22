// stockMove(type, config) against the conformance corpus: every move of the
// six hand ports that is a Go stock move used as is (SeatPlayer, FinishTurn,
// StartPhase, DealCountComponents, ...) is replaced by the registry's parts,
// built from the configuration Go stores for it, keeping only what the
// installers and the game add (name, help text, phases, progression,
// authored preconditions). The rebuilt games must still pass every corpus
// document at every version for every viewer. Blackjack's two deals then
// run on the generic dealComponents, not on the hand ports'
// dealCountComponents.
//
// Also: the registry's lookups and refusals.
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import { blackjack } from '../games/blackjack.ts';
import { debuganimations } from '../games/debuganimations.ts';
import { memory } from '../games/memory.ts';
import { pig } from '../games/pig.ts';
import { tictactoe } from '../games/tictactoe.ts';
import { werewolf } from '../games/werewolf.ts';
import { type ConformanceDocument, FIELDS, runDocument } from '../testing/conformance.ts';
import { compileGame } from './compile.ts';
import { CONFIG_PREFIX, configValue, stockMove } from './stockRegistry.ts';
import type { GameSpec, MoveSpec } from './types.ts';

/** A hand port's move by name: the Go stock type it is and the configuration Go stores for it. */
function stockOf(game: string, name: string): { type: string; config: Record<string, unknown> } | null {
	const q = (short: Record<string, unknown>): Record<string, unknown> => {
		const out: Record<string, unknown> = {};
		for (const key of Object.keys(short)) out[CONFIG_PREFIX + key] = short[key];
		return out;
	};
	switch (name) {
		case 'Seat Player':
			return { type: 'moves.SeatPlayer', config: {} };
		case 'Activate Inactive Player':
			return { type: 'moves.ActivateInactivePlayer', config: {} };
		case 'Activate Inactive Player - Gathering':
			return { type: 'moves.ActivateInactivePlayer', config: q({ MoveNameSuffix: 'Gathering' }) };
		case 'Wait For Enough Players':
			return { type: 'moves.WaitForEnoughPlayers', config: {} };
		case 'Wait For Enough Players - Gathering':
			return { type: 'moves.WaitForEnoughPlayers', config: q({ MoveNameSuffix: 'Gathering' }) };
		case 'Inactivate Empty Seat':
			return { type: 'moves.InactivateEmptySeat', config: {} };
		case 'Inactivate Empty Seat - Gathering':
			return { type: 'moves.InactivateEmptySeat', config: q({ MoveNameSuffix: 'Gathering' }) };
		case 'Finish Turn':
			return { type: 'moves.FinishTurn', config: {} };
		case 'Force Finish Turn':
			return { type: 'moves.ForceFinishTurn', config: q({ MoveName: 'Force Finish Turn' }) };
		case 'Shuffle Discard Into Draw':
			return { type: 'moves.ShuffleDiscardIntoDraw', config: {} };
	}
	if (game !== 'blackjack') return null;
	switch (name) {
		case 'Begin Game':
		case 'Start Next Round':
			return { type: 'moves.StartPhase', config: q({ StartPhase: 1 }) };
		case 'Start Phase Normal Play':
			return { type: 'moves.StartPhase', config: q({ StartPhase: 2 }) };
		case 'Start Round Cleanup':
			return { type: 'moves.StartPhase', config: q({ StartPhase: 3 }) };
		case 'Deal Initial Hidden Card':
			return { type: 'moves.DealCountComponents', config: q({ GameProperty: 'DrawStack', PlayerPropety: 'HiddenHand' }) };
		case 'Deal Initial Visible Card':
			return { type: 'moves.DealCountComponents', config: q({ GameProperty: 'DrawStack', PlayerPropety: 'VisibleHand' }) };
	}
	return null;
}

/** The game with every stock move rebuilt from the registry. */
function rebuilt(game: string, spec: GameSpec): { spec: GameSpec; replaced: string[] } {
	const replaced: string[] = [];
	const moves = spec.moves.map((original): MoveSpec => {
		const stock = stockOf(game, original.name);
		if (stock === null) return original;
		const parts = stockMove(stock.type, stock.config);
		assert.ok(parts?.spec && parts.apply && parts.legal, stock.type);
		replaced.push(original.name);
		const move: MoveSpec = { ...(parts.spec as MoveSpec), name: original.name, helpText: original.helpText, apply: parts.apply, legal: parts.legal };
		if (parts.defaults) move.defaults = parts.defaults;
		else delete move.defaults;
		// What the installers and the game add over the stock move.
		if (original.legalPhases) move.legalPhases = original.legalPhases;
		if (original.progression) move.progression = original.progression;
		if (original.preconditions) move.preconditions = original.preconditions;
		return move;
	});
	return { spec: { ...spec, moves }, replaced };
}

function corpusDir(): string {
	let dir = path.dirname(fileURLToPath(import.meta.url));
	for (;;) {
		if (existsSync(path.join(dir, 'conformance', 'index.json'))) return path.join(dir, 'conformance');
		const parent = path.dirname(dir);
		assert.notEqual(parent, dir, 'could not find conformance/');
		dir = parent;
	}
}

const dir = corpusDir();
const index = JSON.parse(readFileSync(path.join(dir, 'index.json'), 'utf8')) as { file: string; game: string }[];
const ports: Record<string, () => GameSpec> = { memory, blackjack, tictactoe, pig, debuganimations, werewolf };
const totals: Record<string, number> = {};

for (const game of Object.keys(ports)) {
	test('registry parts reproduce ' + game + ' on its corpus', () => {
		const { spec, replaced } = rebuilt(game, ports[game]());
		totals[game] = replaced.length;
		assert.ok(replaced.length >= 1, game + ' has no stock move');
		const compiled = compileGame(spec);
		const entries = index.filter(entry => entry.game === game);
		assert.ok(entries.length >= 1);
		for (const entry of entries) {
			const doc = JSON.parse(gunzipSync(readFileSync(path.join(dir, entry.file))).toString('utf8')) as ConformanceDocument;
			const result = runDocument(compiled, doc);
			const problems: string[] = [];
			if (result.fatal) problems.push('replay stopped: ' + result.fatal);
			for (const field of FIELDS) {
				const f = result.fields[field];
				if (f.mismatches > 0) problems.push(field + ': ' + f.mismatches + ' of ' + f.compared + ' differ, e.g. ' + f.examples.join('; '));
			}
			assert.equal(problems.length, 0, entry.file + ' with ' + replaced.join(', ') + ':\n  ' + problems.join('\n  '));
		}
	});
}

test('every hand port had stock moves replaced', () => {
	// Blackjack's 13: shuffle, seat, the three gathering moves, Begin Game, Start Round Cleanup, Finish Turn, Force Finish Turn, the default round setup's three, two deals, two start phases.
	assert.ok(totals.blackjack >= 13, 'blackjack: ' + totals.blackjack);
	for (const game of Object.keys(ports)) assert.ok(totals[game] >= 1, game);
});

test('registry: names with and without moves., keys qualified or bare, unknown types', () => {
	assert.deepEqual(stockMove('moves.FinishTurn', {})?.spec?.name, stockMove('FinishTurn', {})?.spec?.name);
	assert.equal(configValue({ [CONFIG_PREFIX + 'TargetCount']: 3 }, 'TargetCount'), 3);
	assert.equal(configValue({ TargetCount: 4 }, 'TargetCount'), 4);
	assert.equal(stockMove('moves.NoSuchMove', {}), undefined);
	assert.equal(stockMove('moves.HopAlongPath', {}), undefined);
	assert.equal(stockMove('moves.WaitForEnoughPlayers', { [CONFIG_PREFIX + 'RequireExplicitStart']: true }), undefined);
	assert.throws(() => stockMove('moves.ShuffleStack', {}), /needs configuration SourceProperty/);
	assert.throws(() => stockMove('moves.StartPhase', {}), /needs configuration StartPhase/);
	const named = stockMove('moves.SeatPlayer', { [CONFIG_PREFIX + 'MoveName']: 'Join', [CONFIG_PREFIX + 'IsFixUp']: false });
	assert.equal(named?.spec?.name, 'Join');
	assert.equal(named?.spec?.fixUp, false);
	// A base whose Apply the game supplies has none.
	assert.equal(stockMove('moves.CurrentPlayer', {})?.apply, undefined);
	assert.equal(stockMove('moves.DefaultComponent', { [CONFIG_PREFIX + 'SourceProperty']: 'Board' })?.apply, undefined);
	assert.equal(stockMove('moves.CurrentPlayer', {})?.spec?.base, 'currentPlayer');
});

test('every stock type the inventory calls ported is in the registry', () => {
	const inventory = JSON.parse(readFileSync(fileURLToPath(new URL('./stock-inventory.json', import.meta.url)), 'utf8')) as {
		goType: string; kind: string; ported: boolean; registryName: string | null; verifiedBy: string[];
	}[];
	const config: Record<string, unknown> = {};
	for (const key of ['SourceProperty', 'DestinationProperty', 'GameProperty', 'PlayerPropety', 'SlotField']) config[CONFIG_PREFIX + key] = 'X';
	config[CONFIG_PREFIX + 'StartPhase'] = 0;
	for (const entry of inventory) {
		if (entry.kind !== 'move') continue;
		const parts = stockMove(entry.goType, config);
		if (entry.ported && entry.registryName !== null) assert.ok(parts !== undefined, entry.goType + ' is ported but not in the registry');
		if (!entry.ported) assert.equal(parts, undefined, entry.goType + ' is in the registry but the inventory says unported');
		if (entry.ported) assert.ok(entry.verifiedBy.length > 0, entry.goType + ' is ported with nothing verifying it');
	}
});
