/**
 * A table of blackjack as a sealed state machine, with one view that each
 * principal is shown their own instance of.
 *
 * - `SealedTable` is the whole table: seats, the deck's order, every hand,
 *   scores, the round, timers and the move history. It is confidential to
 *   this module's policy and to nothing else, so no person reads it, the
 *   table's owner included.
 * - `move` is its only writer. Sitting down and dealing are moves. The
 *   proposer is whoever the runtime says is acting, never a field of the
 *   event.
 * - `viewFor` is the only function whose output the policy releases. It
 *   reads the table and asks the runtime who it is computed for, and returns
 *   what that principal may see: the game as the engine sanitizes it for
 *   their seat, so their own hole card and nobody else's, or a watcher's
 *   view if they hold no seat. The runtime computes it once for each
 *   principal who is looking, and the one rule below releases each instance
 *   to the principal it was computed for.
 *
 * Nothing here is per seat. Who holds a seat is a fact inside the sealed
 * table, and the release follows it.
 *
 * The release needs labs' `viewerBoundRelease` experimental option
 * (`EXPERIMENTAL_VIEWER_BOUND_RELEASE=true`); the shape follows labs'
 * `patterns/examples/cfc-exchange-rules/card-table.tsx`.
 *
 * This module is the policy's identity: the rule names `move` and `viewFor`
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
  v,
} from "commonfabric/cfc";

import {
  stepTable,
  tableOf,
  type TableState,
  viewForPrincipal,
} from "../toolkit/table.ts";
import type { Proposal, TableView } from "../toolkit/view.ts";
import { blackjackGame } from "./game.ts";

//
// The policy
//

// The one release. What `viewFor` computed is released to the principal it
// was computed for, when every confidential location it read was written by
// `move`. The rule names no person and no seat: the principal comes from the
// stamp the runtime put on the value, and which seat is theirs is `viewFor`'s
// to decide from the table.
export const releaseOwnView = exchangeRule({
  appliesTo: THIS_POLICY,
  pre: {
    integrity: [{
      type: "https://commonfabric.org/cfc/atom/TransformedBy",
      identity: {
        kind: "verified",
        moduleIdentity: THIS_POLICY.moduleIdentity,
        symbol: "viewFor",
      },
      inputWitness: {
        type: "https://commonfabric.org/cfc/atom/TransformedBy",
        identity: {
          kind: "verified",
          moduleIdentity: THIS_POLICY.moduleIdentity,
          symbol: "move",
        },
      },
      viewer: v("viewer"),
    }],
  },
  post: {
    addAlternatives: [{
      type: "https://commonfabric.org/cfc/atom/User",
      subject: v("viewer"),
    }],
  },
});

export const tableRules = exchangeRules([releaseOwnView]);

/** A value only this module's policy may release. */
export type Sealed<T> = Confidential<T, readonly [PolicyOf<typeof tableRules>]>;

//
// The sealed state
//

/** Everything the table knows. */
export type BlackjackTable = TableState;

/**
 * The stored table: sealed, and written only by `move`. It is `null` until
 * the first accepted proposal.
 *
 * TODO(cfc): the table starts as `null` and not as the table before anyone
 * sat down. The runtime names `move` as the writer of the whole record only
 * when every value a write left in place is one `move` wrote before, and a
 * default is nobody's: a record default whose `history: []` the first move
 * left as it found it would keep every view sealed (labs
 * `docs/specs/cfc-transformed-by-input-witnesses.md`, "What the endorsed
 * writer's stamp covers"). Nothing of `null` outlives the first move.
 */
export type SealedTable = WriteAuthorizedBy<
  Sealed<BlackjackTable | null>,
  typeof move
>;

/** What `move` is bound to: the table, and nothing else. */
export interface MoveBinding {
  table: Writable<SealedTable>;
}

/**
 * The table's one writer. A proposal the rules refuse writes nothing and
 * does not throw, so a refusal tells its sender only that the table's
 * revision did not move.
 *
 * Everything it works from but the proposer and the clock is the table
 * itself, so binding it anywhere else runs the same rules over the same
 * table for the same proposer.
 */
export const move = handler<Proposal, MoveBinding>((proposal, { table }) => {
  const proposer = currentPrincipal();
  if (proposer === undefined) return;
  const result = stepTable(blackjackGame(), tableOf(table.get()), proposal, {
    proposer,
    now: Date.now(),
    entropy: Math.random,
  });
  if (result.error === "") table.set(result.table);
});

//
// The released projection
//

/** What `viewFor` reads: the table, and nothing a caller can choose. */
export interface Viewed {
  table: BlackjackTable | null | undefined;
}

/**
 * The table's one released function: what the principal it is computed for
 * may see. It is a module-scope `lift` because `TransformedBy` names a
 * function by its module and its export name.
 */
export const viewFor = lift(({ table }: Viewed): TableView =>
  viewForPrincipal(blackjackGame(), table, currentPrincipal())
);

/**
 * The game's secret salt, from which the deck's order derives. Defined beside `viewFor` and named by no rule, so
 * what it computes stays sealed. It exists for the tests.
 */
export const peek = lift(({ table }: Viewed): string =>
  table?.record?.secretSalt ?? ""
);

//
// The table
//

export interface BlackjackTableInput {
  /** The sealed table, which starts as nothing. */
  table: Writable<Default<SealedTable, null>>;
}

export interface BlackjackTableOutput {
  /** What the principal looking may see of the table. */
  view: TableView;
  /** The game's secret salt, which nobody is shown. */
  peek: string;
  /** Propose something as whoever sends it. */
  move: Stream<Proposal>;
}

/**
 * The table with no screen: its sealed state, its writer and its released
 * view. The table holds its own state: a view of a table reached through a
 * link in another pattern's argument carries no input witness, and is shown
 * to nobody.
 */
export default pattern<BlackjackTableInput, BlackjackTableOutput>(({ table }) => ({
  view: viewFor({ table }),
  peek: peek({ table }),
  move: move({ table }),
}));
