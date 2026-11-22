/**
 * Blackjack at one table, for two to seven people, drawn by the boardgame client
 * in a sandboxed `cf-iframe`.
 *
 * The pattern holds one sealed table (`table.tsx`) and hands the guest the
 * view computed for whoever is looking, the game's public facts, the
 * viewer's own principal, and the two verbs a browser may send. The guest is
 * the generic boardgame guest of `fabric/guest`; nothing in it is specific
 * to Blackjack but the renderer compiled into it.
 *
 * The view is handed over whole. A value this pattern derived from it would
 * be sealed again.
 */
import {
  computed,
  currentPrincipal,
  dataFile,
  type Default,
  handler,
  NAME,
  pattern,
  type PerUser,
  type Stream,
  UI,
  type VNode,
  type Writable,
} from "commonfabric";

import type { Proposal, TableInfo, TableView } from "../toolkit/view.ts";
import { blackjackInfo } from "./game.ts";
import { move, type SealedTable, viewFor } from "./table.tsx";

/** The guest's document and bundle, written by `npm run build:fabric`. */
const GUEST_HTML = dataFile("../guest/guest.html");
const GUEST_SCRIPT = dataFile("../guest/guest.js");

export interface BlackjackInput {
  /** The sealed table, which starts as nothing. */
  table: Writable<Default<SealedTable, null>>;
  /** The viewer's own principal, once `identify` has run for them. */
  me?: PerUser<Writable<Default<string, "">>>;
}

/** A table of Blackjack. */
export interface BlackjackOutput {
  [NAME]: string;
  [UI]: VNode;
  /** What the principal looking may see of the table. */
  view: TableView;
  /** The deck and message texts, which are public. */
  info: TableInfo;
  /** Propose something as whoever sends it. */
  move: Stream<Proposal>;
  /** Record the sender's principal where only they read it. */
  identify: Stream<void>;
}

/**
 * Writes the acting principal into that person's own cell, so their guest
 * can tell which seat in the view is theirs. Not a writer of the table.
 */
const identifyViewer = handler<void, { me: Writable<string> }>(
  (_, { me }) => {
    const principal = currentPrincipal();
    if (principal !== undefined && me.get() !== principal) me.set(principal);
  },
);

export default pattern<BlackjackInput, BlackjackOutput>(({ table, me }) => {
  const view = viewFor({ table });
  const propose = move({ table });
  const identify = identifyViewer({ me });
  const info = computed(() => blackjackInfo());
  return {
    [NAME]: "Blackjack",
    [UI]: (
      <div style={{ width: "100%", height: "100vh" }}>
        <cf-iframe
          src={GUEST_HTML}
          script={GUEST_SCRIPT}
          $context={{ view, info, me, move: propose, identify }}
          resourceKinds={{
            view: "readonly",
            info: "readonly",
            me: "readonly",
            move: "stream",
            identify: "stream",
          }}
        />
      </div>
    ),
    view,
    info,
    move: propose,
    identify,
  };
});
