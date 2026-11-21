/**
 * Blackjack at one table, for two to seven people.
 *
 * The pattern holds one sealed table (`table.tsx`) and hands a guest the
 * watcher's view, each seat's own view, the game's public facts, the
 * viewer's own principal, and the two verbs a browser may send.
 *
 * The guest that ships in `fabric/guest` cannot play this yet: it has no
 * blackjack renderer, and its bridge reads one shared `view`, where this
 * table has one a seat. `README.md` lists what the guest needs. Until then
 * this piece deploys and its guest shows the lobby and nothing after it.
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

import { OBSERVER_SEAT } from "../toolkit/table.ts";
import type { Proposal, TableInfo, TableView } from "../toolkit/view.ts";
import { blackjackInfo } from "./game.ts";
import { EMPTY_TABLE, move, type SealedTable, seatView } from "./table.tsx";

/** The guest's document and bundle, written by `npm run build:fabric`. */
const GUEST_HTML = dataFile("../guest/guest.html");
const GUEST_SCRIPT = dataFile("../guest/guest.js");

export interface BlackjackInput {
  /** The sealed table, which starts as the table before anyone sat down. */
  table: Writable<Default<SealedTable, typeof EMPTY_TABLE>>;
  /** The viewer's own principal, once `identify` has run for them. */
  me?: PerUser<Writable<Default<string, "">>>;
}

/** A table of blackjack. */
export interface BlackjackOutput {
  [NAME]: string;
  [UI]: VNode;
  /** What whoever only watches may see. */
  view: TableView | undefined;
  /** The deck and message texts, which are public. */
  info: TableInfo;
  /** Propose something as whoever sends it. */
  move: Stream<Proposal>;
  /** Record the sender's principal where only they read it. */
  identify: Stream<void>;
}

/**
 * Writes the acting principal into that person's own cell, so their guest
 * can tell which seat is theirs. Not a writer of the table.
 */
const identifyViewer = handler<void, { me: Writable<string> }>(
  (_, { me }) => {
    const principal = currentPrincipal();
    if (principal !== undefined && me.get() !== principal) me.set(principal);
  },
);

export default pattern<BlackjackInput, BlackjackOutput>(({ table, me }) => {
  const view = seatView({ table, seat: OBSERVER_SEAT });
  const seat0 = seatView({ table, seat: 0 });
  const seat1 = seatView({ table, seat: 1 });
  const seat2 = seatView({ table, seat: 2 });
  const seat3 = seatView({ table, seat: 3 });
  const seat4 = seatView({ table, seat: 4 });
  const seat5 = seatView({ table, seat: 5 });
  const seat6 = seatView({ table, seat: 6 });
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
          $context={{
            view,
            seat0,
            seat1,
            seat2,
            seat3,
            seat4,
            seat5,
            seat6,
            info,
            me,
            move: propose,
            identify,
          }}
          resourceKinds={{
            view: "readonly",
            seat0: "readonly",
            seat1: "readonly",
            seat2: "readonly",
            seat3: "readonly",
            seat4: "readonly",
            seat5: "readonly",
            seat6: "readonly",
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
