/**
 * A table of blackjack as a sealed state machine, with a view for each seat.
 *
 * - `SealedTable` is the whole table: seats, the deck's order, every hand,
 *   scores, the round, timers and the move history. It is confidential to
 *   this module's policy and to nothing else, so no person reads it.
 * - `move` is its only writer. Sitting down and dealing are moves. The
 *   proposer is whoever the runtime says is acting, never a field of the
 *   event.
 * - `seatView` is the only function whose output the policy releases. It
 *   computes one seat's view: the game as the engine sanitizes it for that
 *   seat, so a seat sees its own hole card and nobody else's. The table has
 *   one such view for each seat and one for whoever only watches.
 *
 * Unlike Memory, the seats see different things, so who a view is released
 * to matters. That is the second of the two interim seams below.
 *
 * This module is the policy's identity; the UI lives in `main.tsx`.
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

import {
  OBSERVER_SEAT,
  seatViewOf,
  stepTable,
  type TableState,
} from "../toolkit/table.ts";
import type { Proposal, TableView } from "../toolkit/view.ts";
import { blackjackGame } from "./game.ts";

//
// The policy, and the interim seams
//
// Everything about how a view is labeled and released today is between this
// heading and the next.
//
// AS DESIGNED there are two releases, each guarded by `TransformedBy` for
// `seatView` with an input witness naming `move`:
//
//   1. The watcher's view (`seat: -1`) gains `Space(table)`, and the
//      runtime's display rule admits a reader of the table's space. This is
//      Memory's release.
//   2. Seat i's view gains the seat's audience instead: the people on seat
//      i's occupant list, which `move` keeps inside the sealed table
//      (fabric/RESEARCH.md, "Dealing to a player: seats are audience
//      lists"). A reader of the space who is not on that list is not
//      admitted, so a modified client cannot read another seat's hole card.
//
// TODO(cfc), the input witness: restore `inputWitness` when a record its
// writer reads and then sets keeps the writer's stamp (Memory's FINDINGS.md
// F2; gap A of the labs spike). Without it the policy releases `seatView`'s
// view of whatever is in the cell, not only of what `move` wrote. This is
// the same seam as Memory's.
//
// TODO(cfc), the seat's audience: release a seat's view to that seat's
// occupants once a rule can name a list held in a cell (specs#59; labs#8608
// is an unlanded draft). No shipped atom says "whoever this value names",
// so the rule below gives EVERY seat's view `Space(table)`, like the
// watcher's.
//
// AS SHIPPED, then, what is not held: every reader of the table's space may
// read every seat's view, hole cards included. A seat's own view goes to
// its own guest only because the guest asks for that one. That protects
// against nothing but an honest client. The table still keeps the views
// apart, one value a seat, so that only this rule changes when the release
// can be narrowed.
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
export type BlackjackTable = TableState;

/** The stored table: sealed, and written only by `move`. */
export type SealedTable = WriteAuthorizedBy<Sealed<BlackjackTable>, typeof move>;

/** The table before anyone sat down, as the cell's default. */
export const EMPTY_TABLE: BlackjackTable = {
  phase: "seating",
  seats: [],
  record: {
    name: "",
    id: "",
    secretSalt: "",
    numPlayers: 0,
    variant: {},
    agents: [],
    seating: false,
  },
  epoch: 0,
  revision: 0,
  version: -1,
  recent: {},
  history: [],
};

/** What `move` is bound to. */
export interface MoveBinding {
  table: Writable<SealedTable>;
}

/**
 * The table's one writer. A proposal the rules refuse writes nothing and
 * does not throw. It writes the sealed table and nothing else.
 */
export const move = handler<Proposal, MoveBinding>((proposal, { table }) => {
  const proposer = currentPrincipal();
  if (proposer === undefined) return;
  const result = stepTable(blackjackGame(), table.get(), proposal, {
    proposer,
    now: Date.now(),
    entropy: Math.random,
  });
  if (result.error === "") table.set(result.table);
});

//
// The released projection
//

/** Which seat's view to compute. */
export interface SeatQuery {
  table: BlackjackTable | undefined;
  /** A seat, or -1 for whoever only watches. Fixed by the pattern below. */
  seat: number;
}

/**
 * The table's one released function: one seat's view of the table. Before
 * the first accepted proposal it returns nothing.
 *
 * `seat` is not a parameter a viewer steers: the pattern below asks for
 * each seat once, by a literal. What stops one seat's view from reaching
 * another seat is the release, which today does not (see above).
 */
export const seatView = lift((
  { table, seat }: SeatQuery,
): TableView | undefined =>
  // Nothing before the first proposal, and nothing for a seat nobody holds.
  table !== undefined && table.revision > 0 && seat < table.seats.length
    ? seatViewOf(blackjackGame(), table, seat)
    : undefined
);

//
// The table
//

export interface BlackjackTableInput {
  /** The sealed table, which starts as the table before anyone sat down. */
  table: Writable<Default<SealedTable, typeof EMPTY_TABLE>>;
}

export interface BlackjackTableOutput {
  /** What whoever only watches may see. */
  observer: TableView | undefined;
  /** Each seat's own view, in seat order. */
  seat0: TableView | undefined;
  seat1: TableView | undefined;
  seat2: TableView | undefined;
  seat3: TableView | undefined;
  seat4: TableView | undefined;
  seat5: TableView | undefined;
  seat6: TableView | undefined;
  /** Propose something as whoever sends it. */
  move: Stream<Proposal>;
}

/**
 * The table with no screen: its sealed state, its writer, and a released
 * view for the watcher and for each of the seven seats a table can have.
 */
export default pattern<BlackjackTableInput, BlackjackTableOutput>(
  ({ table }) => ({
    observer: seatView({ table, seat: OBSERVER_SEAT }),
    seat0: seatView({ table, seat: 0 }),
    seat1: seatView({ table, seat: 1 }),
    seat2: seatView({ table, seat: 2 }),
    seat3: seatView({ table, seat: 3 }),
    seat4: seatView({ table, seat: 4 }),
    seat5: seatView({ table, seat: 5 }),
    seat6: seatView({ table, seat: 6 }),
    move: move({ table }),
  }),
);
