// stockMove(type, config) on the games: every stock move of the seven
// generated games (games/<game>/definition.gen.ts, Go's exported
// definitions) is built by the registry from the configuration Go exports
// for it, and the loader runs it (rules/data.ts); the games' corpora
// (games/<game>/<game>.gen.test.ts) then hold those parts to Go at every
// version for every viewer. The moves the hand ports wrote out as stock
// moves (stockOf) must be the types and configuration Go exports.
//
// Until the hand ports were replaced, this file rebuilt each hand port with
// its stock moves swapped for the registry's parts and replayed the corpus;
// a generated game is built from the registry's parts already, so what that
// rebuild proved is now the generated games' own corpus test
// (fabric/transpile/status/replacement.md).
//
// Also: the registry's lookups and refusals.
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { definition as blackjackDefinition } from '../games/blackjack/definition.gen.ts';
import { definition as checkersDefinition } from '../games/checkers/definition.gen.ts';
import { definition as debuganimationsDefinition } from '../games/debuganimations/definition.gen.ts';
import { definition as memoryDefinition } from '../games/memory/definition.gen.ts';
import { definition as pigDefinition } from '../games/pig/definition.gen.ts';
import { definition as tictactoeDefinition } from '../games/tictactoe/definition.gen.ts';
import { definition as werewolfDefinition } from '../games/werewolf/definition.gen.ts';
import type { GameData } from './data.ts';
import { CONFIG_PREFIX, configValue, stockMove } from './stockRegistry.ts';

/** A move the hand ports wrote out as a Go stock move, by name: the stock type and the configuration they read in the Go source. */
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

const GENERATED: Record<string, () => GameData> = {
	memory: memoryDefinition, blackjack: blackjackDefinition, tictactoe: tictactoeDefinition, pig: pigDefinition,
	debuganimations: debuganimationsDefinition, werewolf: werewolfDefinition, checkers: checkersDefinition,
};
const totals: Record<string, number> = {};

/** `config` holds every key of `want` with the same value. */
function holds(config: Readonly<Record<string, unknown>>, want: Readonly<Record<string, unknown>>): boolean {
	return Object.keys(want).every(key => JSON.stringify(config[key]) === JSON.stringify(want[key]));
}

for (const game of Object.keys(GENERATED)) {
	test('the registry builds every stock move of the generated ' + game + ' from Go\'s exported configuration', () => {
		let built = 0;
		for (const move of GENERATED[game]().Moves) {
			if (move.Stock === null) continue;
			const parts = stockMove(move.Stock.Type, move.Stock.Config);
			assert.ok(parts !== undefined, game + ': ' + move.Name + ' (' + move.Stock.Type + ') is not in the registry');
			// The stock moves the hand ports wrote out are the types and the
			// configuration Go exports for them.
			const read = stockOf(game, move.Name);
			if (read !== null) {
				assert.equal('moves.' + move.Stock.Type.replace(/^moves\./, ''), read.type, game + ': ' + move.Name);
				assert.ok(holds(move.Stock.Config, read.config), game + ': ' + move.Name + ': ' + JSON.stringify(read.config) + ' not in ' + JSON.stringify(move.Stock.Config));
				assert.ok(parts.spec && parts.apply && parts.legal, game + ': ' + move.Name + ' has no stock apply and legal');
			}
			if (parts.apply !== undefined) built++;
		}
		totals[game] = built;
	});
}

test('every generated game runs stock moves from the registry', () => {
	// Blackjack's 13 and more: shuffle, seat, the three gathering moves, Begin Game, Start Round Cleanup, Finish Turn, Force Finish Turn, the default round setup's three, two deals, two start phases.
	assert.ok(totals.blackjack >= 13, 'blackjack: ' + totals.blackjack);
	for (const game of Object.keys(GENERATED)) assert.ok(totals[game] >= 1, game + ': ' + totals[game]);
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
