/**
 * A table of Memory as a sealed state machine.
 *
 * - `SealedTable` is the whole table: seats, layout, face-up slots, captures,
 *   turn, version, timers and move history. It is confidential to this
 *   module's policy and to nothing else, so no person reads it, the table's
 *   owner included.
 * - `move` is its only writer. Sitting down and dealing are moves. The
 *   proposer is whoever the runtime says is acting, never a field of the
 *   event.
 * - `seatView` is the only function whose output the policy releases. As
 *   designed, only when everything confidential it read was written by
 *   `move`; today that last condition is not checked (see "The interim
 *   seam" below).
 *
 * This module is the policy's identity: the rule names `move` and `seatView`
 * by this module and their export names, and an edit here or to anything it
 * imports is a different policy. The UI lives in `main.tsx` so that changing
 * it does not strand a table in flight.
 */
import {
  type Confidential,
  currentPrincipal,
  type Default,
  handler,
  lift,
  pattern,
  type Stream,
  type Writable,
  type WriteAuthorizedBy,
} from "commonfabric";
import {
  exchangeRule,
  exchangeRules,
  type PolicyOf,
  THIS_POLICY,
} from "commonfabric/cfc";

import { stepTable, type TableState, viewOf } from "../toolkit/table.ts";
import type { Proposal, TableView } from "../toolkit/view.ts";
import { type MemoryGame, memoryRules, type MemoryView } from "./rules.ts";

//
// The policy, and the interim seam
//
// Everything about how the view is labeled and released today is between
// this heading and the next. It is the one place this game is weaker than
// its design, and the one place to change as the runtime catches up.
//
// The release is in two steps. This module's rule turns the policy's clause
// into "the table's space" for what `seatView` computed; the runtime's own
// display rule then admits a reader of that space. A rule that asked for the
// reader's role itself would be evaluated over a whole read at once, and
// would refuse the one read of the whole view a `cf-iframe` guest makes
// (FINDINGS.md F3). `cfcPattern` has no constructor for a space, so the atom
// is written out.
//
// AS DESIGNED, the guard also carries an input witness: the view is released
// only when every confidential location `seatView` read was written by
// `move`.
//
//   inputWitness: {
//     type: "https://commonfabric.org/cfc/atom/TransformedBy",
//     identity: { kind: "verified", moduleIdentity: THIS_POLICY.moduleIdentity,
//                 symbol: "move" },
//   },
//
// TODO(cfc): restore `inputWitness` when a record its writer reads and then
// sets keeps the writer's stamp (FINDINGS.md F2; gap A of the labs spike).
// `move` reads the table and sets it, and a destination the transaction read
// is not stamped, so with the witness nothing of this table is ever
// released. The alternative today is to store the table as one JSON string,
// which this pattern declines to do.
//
// The witness also needs a served handler's writes to name the handler,
// which labs main does not do (FINDINGS.md F1). That has a fix on a labs
// branch, and this pattern assumes it; without the witness the rule below
// does not depend on it.
//
// AS SHIPPED, without the witness, the policy releases `seatView`'s view of
// whatever is in the cell, not only of what `move` wrote. `WriteAuthorizedBy`
// still refuses a write to the table by any other code, so this matters
// where that check does not reach: a modified client with write access to
// the space, and a second instance of this module pointed at this table's
// cell.
//
export const releaseSeatView = exchangeRule({
  appliesTo: THIS_POLICY,
  pre: {
    integrity: [{
      type: "https://commonfabric.org/cfc/atom/TransformedBy",
      identity: {
        kind: "verified",
        moduleIdentity: THIS_POLICY.moduleIdentity,
        symbol: "seatView",
      },
    }],
  },
  post: {
    addAlternatives: [{
      type: "https://commonfabric.org/cfc/atom/Space",
      id: THIS_POLICY.subject,
    }],
  },
});

export const tableRules = exchangeRules([releaseSeatView]);

/** A value only this module's policy may release. */
export type Sealed<T> = Confidential<T, readonly [PolicyOf<typeof tableRules>]>;

//
// The sealed state
//

/** Everything the table knows. */
export type MemoryTable = TableState<MemoryGame>;

/** The stored table: sealed, and written only by `move`. */
export type SealedTable = WriteAuthorizedBy<Sealed<MemoryTable>, typeof move>;

/** The table before anyone sat down, as the cell's default. */
export const EMPTY_TABLE: MemoryTable = {
  phase: "seating",
  seats: [],
  gameId: "",
  secretSalt: "",
  epoch: 0,
  revision: 0,
  version: -1,
  recent: [],
  history: [],
  timers: [],
};

/** What `move` is bound to. */
export interface MoveBinding {
  table: Writable<SealedTable>;
}

/**
 * The table's one writer. A proposal the rules refuse writes nothing and
 * does not throw, so a refusal tells its sender only that the table's
 * revision did not move.
 *
 * It writes the sealed table and nothing else: one transaction carries one
 * flow label, so anything public about the table is a field of the view.
 */
export const move = handler<Proposal, MoveBinding>((proposal, { table }) => {
  const proposer = currentPrincipal();
  if (proposer === undefined) return;
  const result = stepTable(memoryRules(), table.get(), proposal, {
    proposer,
    now: Date.now(),
    entropy: Math.random,
  });
  if (result.error === "") table.set(result.table);
});

//
// The released projection
//

/** What a reader of the table's space may see. */
export type MemoryTableView = TableView<MemoryView>;

/**
 * The table's one released function. It is a module-scope `lift` because
 * `TransformedBy` names a function by its module and its export name.
 *
 * Before the first accepted proposal it returns nothing. The table then
 * holds its default, which `move` did not write, so under the rule as
 * designed a view of it would carry no input witness, and a field the first
 * move left unchanged would keep that unwitnessed stamp.
 */
export const seatView = lift((
  table: MemoryTable | undefined,
): MemoryTableView | undefined =>
  table !== undefined && table.revision > 0 ? viewOf(memoryRules(), table) : undefined
);

/**
 * Every card's face, in slot order. Defined beside `seatView` and named by
 * no rule, so what it computes stays sealed. It exists for the tests.
 */
export const peek = lift((table: MemoryTable | undefined): number[] => {
  const head = table?.recent?.[table.recent.length - 1];
  return head === undefined ? [] : head.game.Game.HiddenCards.Indexes.slice();
});

//
// The table
//

export interface MemoryTableInput {
  /** The sealed table, which starts as the table before anyone sat down. */
  table: Writable<Default<SealedTable, typeof EMPTY_TABLE>>;
}

export interface MemoryTableOutput {
  /** What the policy releases; `undefined` until the first proposal. */
  view: MemoryTableView | undefined;
  /** Propose something as whoever sends it. */
  move: Stream<Proposal>;
}

/**
 * The table with no screen: its sealed state, its writer and its released
 * view. The table holds its own state. A table lent its state by the pattern
 * around it would lose the release's input witness.
 */
export default pattern<MemoryTableInput, MemoryTableOutput>(({ table }) => ({
  view: seatView(table),
  move: move({ table }),
}));
