// Test support for the facade's tests (no tests of its own): a RulesState
// over a bare definition, with only the parts of a CompiledGame the facade
// reads (the definition, the enums, the decks' static values). The facade
// is exercised on real Go states (testdata/defs/vecgame.def.json and the
// vectors' blobs) without a GameSpec for the vector game.
import { loadEngineState } from '../state.ts';
import type { StateJSON } from '../state.ts';
import type { EnumDef, GameDefinition, PropDef } from '../definition.ts';
import { newEnumRegistry } from '../rules/enums.ts';
import { newRulesState } from '../rules/state.ts';
import type { CompiledGame, RulesState } from '../rules/state.ts';
import type { GameRecord } from '../rules/types.ts';

/** A RulesState for `json` under `def`, as the facade sees one. */
export function rulesStateFor(def: GameDefinition, identity: { gameId: string; secretSalt: string }, version: number, json: StateJSON, variant: Record<string, string> | null = null): RulesState {
	const engine = loadEngineState(def, identity, version, json);
	const decks: Record<string, { components: Record<string, unknown>[] }> = {};
	for (const name of Object.keys(def.Decks)) decks[name] = { components: (def.Decks[name].Components ?? []) as Record<string, unknown>[] };
	const game = {
		spec: { name: def.Name, decks, game: { props: {} }, player: { props: {} } },
		def,
		enums: newEnumRegistry(def.Enums),
		components: {},
	} as unknown as CompiledGame;
	const record: GameRecord = {
		name: def.Name,
		id: identity.gameId,
		secretSalt: identity.secretSalt,
		numPlayers: json.Players.length,
		variant,
		agents: json.Players.map(() => ''),
		seating: false,
	};
	return newRulesState(game, engine, record, [], { pendingSeat: null });
}

/** A three-player state with no decks: game properties as given, players empty. */
export function stubRulesState(enums: Record<string, EnumDef> = {}, gameProps: Record<string, PropDef> = {}, variant: Record<string, string> | null = null): RulesState {
	const def: GameDefinition = { Name: 'stub', Decks: {}, Enums: enums, Game: { Props: gameProps }, Player: { Props: {} }, DynamicComponentValues: {} };
	const game: Record<string, unknown> = {};
	for (const name of Object.keys(gameProps)) game[name] = '';
	return rulesStateFor(def, { gameId: 'stub', secretSalt: 'salt' }, 0, { Game: game, Players: [{}, {}, {}], Components: {} }, variant);
}
