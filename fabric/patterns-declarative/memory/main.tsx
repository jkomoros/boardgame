/**
 * Memory at one table, drawn by the pattern's own JSX with the boardgame
 * declarative elements, in the shell's document: no `cf-iframe`, no guest.
 *
 * The table, its one writer `move` and its one released lift `viewFor` are
 * those of `fabric/patterns/memory/table.tsx`, imported, not copied. This
 * file is the view only.
 *
 * `viewFor` is computed for whoever is looking and released to them alone
 * (labs' `viewerBoundRelease`). It is handed to `<bg-table>` exactly as
 * released. The element reads whose view it is from the view, queues every
 * version it has not shown, and owns the animation. The board below it is markup whose props never
 * change, so nothing the reconciler does between two states touches it.
 */
import {
  computed,
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

import "../elements.ts";
import type {
  ElementEvent,
  ProposeMoveDetail,
  TimerElapsedDetail,
} from "../elements.ts";
import type {
  Proposal,
  TableInfo,
  TableView,
} from "../../patterns/toolkit/view.ts";
import { memoryInfo } from "../../patterns/memory/game.ts";
import {
  move,
  type SealedTable,
  viewFor,
} from "../../patterns/memory/table.tsx";

export interface MemoryInput {
  /** The sealed table, which starts as the table before anyone sat down. */
  table: Writable<Default<SealedTable, null>>;
  /** What the viewer typed as their name. */
  seatName?: PerUser<Writable<Default<string, "">>>;
}

export interface MemoryOutput {
  [NAME]: string;
  [UI]: VNode;
  view: TableView;
  info: TableInfo;
  move: Stream<Proposal>;
}

/**
 * `propose-move` from an element, as a proposal to the table. The element
 * says which move and which arguments; who proposes is the runtime's to say,
 * inside `move`.
 */
const play = handler<
  ElementEvent<ProposeMoveDetail>,
  { move: Stream<Proposal> }
>((event, { move }) => {
  const args: Record<string, string> = {};
  for (const [name, value] of Object.entries(event.detail.args)) {
    args[name] = String(value);
  }
  move.send({ kind: "play", name: event.detail.move, arguments: args });
});

/** `timer-elapsed` from the table element: ask the table to look at its clock. */
const tick = handler<
  ElementEvent<TimerElapsedDetail>,
  { move: Stream<Proposal> }
>((event, { move }) => {
  move.send({ kind: "tick", timerId: event.detail.timerId });
});

/** Take a seat. Who is sitting is the runtime's to say, inside `move`. */
const sit = handler<
  unknown,
  {
    move: Stream<Proposal>;
    seatName: Writable<string>;
  }
>((_, { move, seatName }) => {
  move.send({ kind: "sit", label: seatName.get() || "Player" });
});

const deal = handler<unknown, { move: Stream<Proposal> }>((_, { move }) => {
  move.send({ kind: "start", options: { numcards: "small" } });
});

const PILE = {
  display: "inline-block",
  width: "120px",
  minHeight: "150px",
  margin: "0 12px",
  textAlign: "center",
};

export default pattern<MemoryInput, MemoryOutput>(
  ({ table, seatName }) => {
    const view = viewFor({ table });
    const propose = move({ table });
    const info = computed(() => memoryInfo());
    return {
      [NAME]: "Memory (declarative)",
      [UI]: (
        <div style={{ padding: "12px", fontFamily: "system-ui, sans-serif" }}>
          <div id="lobby" style={{ marginBottom: "12px" }}>
            <cf-input $value={seatName} placeholder="Your name" />
            <cf-button id="sit" onClick={sit({ move: propose, seatName })}>
              Sit down
            </cf-button>
            <cf-button id="deal" onClick={deal({ move: propose })}>
              Deal
            </cf-button>
          </div>
          <bg-table
            name="memory"
            gameName="memory"
            chest={info.chest}
            tableView={view}
            animationLength={400}
            holdBefore={{ "Capture Cards": 600 }}
            onpropose-move={play({ move: propose })}
            ontimer-elapsed={tick({ move: propose })}
          >
            <div style={{ position: "relative", minHeight: "320px" }}>
              <bg-stack
                path="Game.Cards"
                layout="grid"
                messy
                face="Type"
                pick={{ move: "Reveal Card", arg: "CardIndex", slots: "hidden" }}
              />
              <bg-callout message="Match" path="Game.Cards.$count" />
            </div>
            <div style={{ margin: "16px 0" }}>
              <div style={PILE}>
                <bg-stack
                  path="Players.0.WonCards"
                  layout="stack"
                  messy
                  face="Type"
                />
              </div>
              <div style={PILE}>
                <bg-stack
                  path="Players.1.WonCards"
                  layout="stack"
                  messy
                  face="Type"
                />
              </div>
            </div>
            <bg-turn-status />
          </bg-table>
        </div>
      ),
      view,
      info,
      move: propose,
    };
  },
);
