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
 *   `move`; today neither that condition nor the reader's role is checked (see "The interim
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
// AS DESIGNED, the one release reads: a reader of the table's space may see
// what `seatView` computed, when every confidential location it read was
// written by `move`:
//
//   {
//     type: ".../TransformedBy",
//     identity: { kind: "verified", moduleIdentity: THIS_POLICY.moduleIdentity,
//                 symbol: "seatView" },
//     inputWitness: {
//       type: ".../TransformedBy",
//       identity: { kind: "verified", moduleIdentity: THIS_POLICY.moduleIdentity,
//                   symbol: "move" },
//     },
//   },
//   cfcPattern.hasRole(v("reader"), THIS_POLICY.subject, "reader"),
//
// with `post: { addAlternatives: [cfcPattern.user(v("reader"))] }`.
//
// TODO(cfc): restore both conditions (FINDINGS.md F1 to F3). With them, the
// runtime releases nothing of this table to a guest:
//
//   F1. A handler run by the server, which is the default, leaves no input
//       witness at all, whatever shape the state has.
//   F2. Where the handler runs on a client, the witness survives only when
//       the state is a single scalar. `move` reads the record and sets it,
//       and a destination the transaction read is not stamped.
//   F3. Under a rule that asks for a role, the display boundary releases
//       each scalar of the view read alone and refuses one read of the
//       whole view, which is the read a `cf-iframe` guest makes.
//
// AS SHIPPED, the rule below names `seatView` alone and drops the clause.
// What that gives up:
//
//   - Without the witness, the policy releases `seatView`'s view of whatever
//     is in the cell, not only of what `move` wrote. `WriteAuthorizedBy`
//     still refuses a write to the table by any other code, so this matters
//     where that check does not reach: a modified client with write access
//     to the space, and a second instance of this module pointed at this
//     table's cell.
//   - Without the role, the released view carries no clause at all, so it is
//     readable by whoever the bytes reach rather than by readers of the
//     table's space. Today those are the same people, because a space's
//     documents are delivered to its readers and to nobody else. Once
//     delivery checks labels (specs#95) they are not, and the role has to
//     be back by then.
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
  post: { dropClause: true },
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
