// The stock moves by name, for the loader (rules/data.ts): GameData names a
// move's nearest stock ancestor and its options (MoveData.Stock), and the
// loader asks this module for the parts that ancestor contributes. It is a
// `switch` over the stock type's Go name, because a pattern module may not
// hold a record of functions.
//
// stock-inventory.json beside this file lists every stock move type and
// behavior in Go, whether it is ported, under which name, and what proves it.

import {
	activateEmptySeat, activateFilledSeat, activateInactivePlayer, anyPlayerLegal, baseLegal, closeAllSeats, closeEmptySeat, currentPlayerMove,
	dealComponents, defaultComponent, defaultMove, done, drawToPlayer, finishTurn, fixUp, fixUpMulti, forceFinishTurn, inactivateEmptySeat,
	moveComponents, moveComponentToSlot, noOp, seatPlayer, shuffleDiscardIntoDraw, shuffleStack, startPhase, stockParts, waitForEnoughPlayers,
	type DealKind, type MoveComponentsKind,
} from './stock.ts';
import type { LegalResult, MoveInstance, MoveSpec } from './types.ts';
import type { RulesState } from './state.ts';

/**
 * What one stock move contributes to a move built from data. A game move
 * that declares no method of its own is `spec` alone; one that overrides a
 * method calls the matching part where its Go calls the embedded stock
 * method (`stockApply`, `base()`).
 */
export interface StockParts {
	/** The stock move's configuration as MoveSpec fields: base, fields, contributions, names, flags. */
	spec?: Partial<MoveSpec>;
	/** The stock Apply (Go: m.<Stock>.Apply(state)). */
	apply?: (state: RulesState, move: MoveInstance) => void;
	/** The stock Legal, given the base checks (Go: m.<Stock>.Legal(state, proposer)). */
	legal?: (state: RulesState, move: MoveInstance, proposer: number, base: () => LegalResult) => LegalResult;
	/** The stock DefaultsForState. */
	defaults?: (state: RulesState, move: MoveInstance) => void;
}

/** Go's fullyQualifiedPackageName, the prefix of every moves.With* configuration key. */
export const CONFIG_PREFIX = 'github.com/jkomoros/boardgame/moves.';

/**
 * A configuration value by its key's short name ("TargetCount" for
 * moves.configPropTargetCount), stored fully qualified as Go does or bare.
 */
export function configValue(config: Readonly<Record<string, unknown>>, key: string): unknown {
	if (Object.prototype.hasOwnProperty.call(config, CONFIG_PREFIX + key)) return config[CONFIG_PREFIX + key];
	if (Object.prototype.hasOwnProperty.call(config, key)) return config[key];
	return undefined;
}

function configString(config: Readonly<Record<string, unknown>>, key: string): string | undefined {
	const value = configValue(config, key);
	return typeof value === 'string' ? value : undefined;
}

function configInt(config: Readonly<Record<string, unknown>>, key: string): number | undefined {
	const value = configValue(config, key);
	return typeof value === 'number' ? value : undefined;
}

function configBool(config: Readonly<Record<string, unknown>>, key: string): boolean | undefined {
	const value = configValue(config, key);
	return typeof value === 'boolean' ? value : undefined;
}

/**
 * The name of an enum carried in a configuration value (WithPhaseToStart's
 * and WithLegalType's optional enum): a string, or an object with a Name or
 * Values. Undefined when absent.
 */
function enumValues(value: unknown): Readonly<Record<string, string>> | undefined {
	if (value !== null && typeof value === 'object') {
		const values = (value as { Values?: unknown }).Values;
		if (values !== null && typeof values === 'object') return values as Record<string, string>;
	}
	return undefined;
}

function enumName(value: unknown): string | undefined {
	if (typeof value === 'string') return value;
	if (value !== null && typeof value === 'object') {
		const name = (value as { Name?: unknown }).Name;
		if (typeof name === 'string') return name;
	}
	return undefined;
}

function required(config: Readonly<Record<string, unknown>>, key: string, type: string): string {
	const value = configString(config, key);
	if (value === undefined) throw new Error('stockMove: ' + type + ' needs configuration ' + key);
	return value;
}

/** Applies the configuration every stock move honors: MoveName, HelpText, IsFixUp. */
function configured(spec: MoveSpec, config: Readonly<Record<string, unknown>>): StockParts {
	const name = configString(config, 'MoveName');
	const helpText = configString(config, 'HelpText');
	const isFixUp = configValue(config, 'IsFixUp');
	const out: MoveSpec = { ...spec };
	if (name !== undefined) out.name = name;
	if (helpText !== undefined) out.helpText = helpText;
	// overrideIsFixUp: a value that is not a bool means false.
	if (isFixUp !== undefined) out.fixUp = isFixUp === true;
	const parts = stockParts(out);
	const result: StockParts = { spec: out, apply: parts.apply, legal: parts.legal };
	if (parts.defaults) result.defaults = parts.defaults;
	return result;
}

/** A stock move whose Apply is the game's: the spec and legality only. */
function withoutApply(parts: StockParts): StockParts {
	const { apply: _apply, ...rest } = parts;
	return rest;
}

function nothing(): void {
	// A base move's Apply: the game supplies the real one.
}

/**
 * The parts of the stock move whose Go type name in package moves is
 * `type` ("CurrentPlayer", "StartPhase", "DealCountComponents", ..., with or
 * without a "moves." prefix), configured with `config`
 * (MoveData.Stock.Config: CustomConfiguration() keys as Go stores them,
 * fully qualified; bare short names are accepted too). Undefined for a stock
 * type the engine has not ported: the loader reports that as an engine gap.
 *
 * For a base type whose Apply the game must supply (Default, FixUp,
 * FixUpMulti, CurrentPlayer, DefaultComponent) there is no `apply` part.
 */
export function stockMove(type: string, config: Readonly<Record<string, unknown>>): StockParts | undefined {
	const name = type.startsWith('moves.') ? type.slice('moves.'.length) : type;
	const suffix = configString(config, 'MoveNameSuffix');
	const nameSuffix = suffix === undefined ? {} : { nameSuffix: suffix };
	const targetCount = configInt(config, 'TargetCount');
	switch (name) {
		case 'Default':
			return withoutApply(configured(defaultMove({ name: 'Default Move', apply: nothing }), config));
		case 'FixUp':
			return withoutApply(configured(fixUp({ name: 'FixUp Move', apply: nothing }), config));
		case 'FixUpMulti':
			return withoutApply(configured(fixUpMulti({ name: 'FixUp Move', apply: nothing }), config));
		case 'CurrentPlayer':
			return withoutApply(configured(currentPlayerMove({ name: 'Current Player Move', apply: nothing }), config));
		case 'AnyPlayer': {
			const parts = configured({ name: 'Any Player Move', helpText: 'A move that any seated player can make for themselves.', fixUp: false, base: 'anyPlayer', legal: anyPlayerLegal, apply: nothing }, config);
			return withoutApply(parts);
		}
		case 'Done':
			return configured(done(), config);
		case 'NoOp':
			return configured(noOp(), config);
		case 'FinishTurn':
			return configured(finishTurn(), config);
		case 'ForceFinishTurn':
			// Go names it "Force Finish Turn" unless a MoveName is given.
			return configured(forceFinishTurn({ name: 'Force Finish Turn' }), config);
		case 'StartPhase': {
			const phase = configInt(config, 'StartPhase');
			if (phase === undefined) throw new Error('stockMove: StartPhase needs configuration StartPhase');
			const phaseEnum = configValue(config, 'StartPhaseEnum');
			const values = enumValues(phaseEnum);
			const phaseName = values?.[String(phase)] ?? String(phase);
			return configured(startPhase({ phase, phaseName }), config);
		}
		case 'SeatPlayer':
			return configured(seatPlayer(nameSuffix), config);
		case 'ActivateInactivePlayer':
			return configured(activateInactivePlayer(nameSuffix), config);
		case 'InactivateEmptySeat':
			return configured(inactivateEmptySeat(nameSuffix), config);
		case 'CloseEmptySeat':
			return configured(closeEmptySeat(nameSuffix), config);
		case 'ActivateEmptySeat':
			return configured(activateEmptySeat(nameSuffix), config);
		case 'ActivateFilledSeat':
			return configured(activateFilledSeat(nameSuffix), config);
		case 'WaitForEnoughPlayers':
			// WithRequireExplicitStart is not ported.
			if (configBool(config, 'RequireExplicitStart') === true) return undefined;
			return configured(waitForEnoughPlayers(targetCount === undefined ? nameSuffix : { ...nameSuffix, targetCount }), config);
		case 'CloseAllSeats': {
			const options: { targetCount?: number; requireAdmin?: boolean } = {};
			if (targetCount !== undefined) options.targetCount = targetCount;
			if (configBool(config, 'RequireAdmin') === true) options.requireAdmin = true;
			return configured(closeAllSeats(options), config);
		}
		case 'ShuffleDiscardIntoDraw':
			// A named DrawDiscardPair field (WithDrawDiscardPairField) is not ported.
			if (configValue(config, 'DrawDiscardPairField') !== undefined) return undefined;
			return configured(shuffleDiscardIntoDraw(), config);
		case 'ShuffleStack':
			return configured(shuffleStack({ source: required(config, 'SourceProperty', name) }), config);
		case 'DrawToPlayer': {
			const gameStack = configString(config, 'ResolvedDrawSource') ?? configString(config, 'GameProperty');
			const playerStack = configString(config, 'ResolvedDrawDestination') ?? configString(config, 'PlayerPropety');
			if (gameStack === undefined || playerStack === undefined) return undefined;
			const parts = drawToPlayer({ gameStack, playerStack });
			const { apply, legal, ...spec } = parts;
			return { spec, apply, legal };
		}
		case 'MoveComponentToSlot': {
			// Go discovers the slot field at boot when WithSlotField is absent
			// (the move's only int property) and stores nothing; the exporter
			// writes what Go resolved as ResolvedSlotField (gamedef
			// ResolvedSlotFieldKey). A configured SlotField is accepted alone.
			const options: { source: string; destination: string; slotField: string; sourceSlotField?: string } = {
				source: required(config, 'SourceProperty', name),
				destination: required(config, 'DestinationProperty', name),
				slotField: configString(config, 'ResolvedSlotField') ?? required(config, 'SlotField', name),
			};
			const sourceSlotField = configString(config, 'SourceSlotField');
			if (sourceSlotField !== undefined) options.sourceSlotField = sourceSlotField;
			const parts = moveComponentToSlot(options);
			const { apply, legal, ...spec } = parts;
			return { spec, apply, legal };
		}
		case 'DefaultComponent': {
			const options: { sourceStack: string; legalType?: number; legalTypeEnum?: string } = { sourceStack: required(config, 'SourceProperty', name) };
			const legalType = configInt(config, 'LegalType');
			if (legalType !== undefined) options.legalType = legalType;
			const legalTypeEnum = enumName(configValue(config, 'LegalTypeEnum'));
			if (legalTypeEnum !== undefined) options.legalTypeEnum = legalTypeEnum;
			return withoutApply(configured(defaultComponent(options), config));
		}
		case 'DealCountComponents':
		case 'DealComponentsUntilPlayerCountReached':
		case 'DealComponentsUntilGameCountLeft':
		case 'DealAllComponents':
		case 'CollectCountComponents':
		case 'CollectComponentsUntilPlayerCountLeft':
		case 'CollectComponentsUntilGameCountReached':
		case 'CollectAllComponents': {
			const options: { kind: DealKind; gameStack: string; playerStack: string; targetCount?: number } = {
				kind: name,
				gameStack: required(config, 'GameProperty', name),
				playerStack: required(config, 'PlayerPropety', name),
			};
			if (targetCount !== undefined) options.targetCount = targetCount;
			return configured(dealComponents(options), config);
		}
		case 'MoveCountComponents':
		case 'MoveComponentsUntilCountReached':
		case 'MoveComponentsUntilCountLeft':
		case 'MoveAllComponents': {
			const options: { kind: MoveComponentsKind; source: string; destination: string; targetCount?: number } = {
				kind: name,
				source: required(config, 'SourceProperty', name),
				destination: required(config, 'DestinationProperty', name),
			};
			if (targetCount !== undefined) options.targetCount = targetCount;
			return configured(moveComponents(options), config);
		}
	}
	return undefined;
}

export { baseLegal };
