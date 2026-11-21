/**
 * Memory at one table, for two to six people, drawn by the boardgame client
 * in a sandboxed `cf-iframe`.
 *
 * The pattern holds one sealed table (`table.tsx`) and hands the guest four
 * things: the released view, the game's public facts, the viewer's own
 * principal, and the two verbs a browser may send. The guest is the generic
 * boardgame guest of `fabric/guest`; nothing in it is specific to Memory but
 * the renderer compiled into it.
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

import type { Proposal, TableInfo } from "../toolkit/view.ts";
import { memoryInfo } from "./game.ts";
import {
  EMPTY_TABLE,
  type MemoryTableView,
  move,
  type SealedTable,
  seatView,
} from "./table.tsx";

/** The guest's document and bundle, written by `npm run build:fabric`. */
const GUEST_HTML = dataFile("../guest/guest.html");
const GUEST_SCRIPT = dataFile("../guest/guest.js");

export interface MemoryInput {
  /** The sealed table, which starts as the table before anyone sat down. */
  table: Writable<Default<SealedTable, typeof EMPTY_TABLE>>;
  /** The viewer's own principal, once `identify` has run for them. */
  me?: PerUser<Writable<Default<string, "">>>;
}

/** A table of Memory. */
export interface MemoryOutput {
  [NAME]: string;
  [UI]: VNode;
  /** What the policy releases to the table's readers. */
  view: MemoryTableView | undefined;
  /** The deck and message texts, which are public. */
  info: TableInfo;
  /** Propose something as whoever sends it. */
  move: Stream<Proposal>;
  /** Record the sender's principal where only they read it. */
  identify: Stream<void>;
}

/**
 * Writes the acting principal into that person's own cell, so their guest
 * can tell which seat in the view is theirs. A lift cannot ask who is
 * looking, and the view is the same for every reader.
 *
 * This is not a writer of the table and reads nothing of it.
 */
const identifyViewer = handler<void, { me: Writable<string> }>(
  (_, { me }) => {
    const principal = currentPrincipal();
    if (principal !== undefined && me.get() !== principal) me.set(principal);
  },
);

export default pattern<MemoryInput, MemoryOutput>(({ table, me }) => {
  const view = seatView(table);
  const propose = move({ table });
  const identify = identifyViewer({ me });
  const info = computed(() => memoryInfo());
  return {
    [NAME]: "Memory",
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
