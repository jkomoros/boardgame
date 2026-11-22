// The stock moves by name, for the loader (rules/data.ts): GameData names a
// move's nearest stock ancestor and its options (MoveData.Stock), and the
// loader asks this module for the parts that ancestor contributes. It is a
// `switch` over the stock type's Go name, because a pattern module may not
// hold a record of functions.

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

/**
 * The parts of the stock move whose Go type name in package moves is
 * `type` ("CurrentPlayer", "StartPhase", "DealCountComponents", ...),
 * configured with `config` (MoveData.Stock.Config: CustomConfiguration()
 * keys as stored). Undefined for a stock type the engine has not ported:
 * the loader reports that as an engine gap.
 *
 * Not implemented: workstream 6 owns the registry and workstream 9 the
 * parts it returns.
 */
export function stockMove(type: string, config: Readonly<Record<string, unknown>>): StockParts | undefined {
	void type;
	void config;
	throw new Error('stockMove: not implemented');
}
