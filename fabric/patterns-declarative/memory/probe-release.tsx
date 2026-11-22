/**
 * EXPERIMENT. Which reads of the released view reach the screen: a path read
 * written straight into JSX, the same value through a `computed`, a `.map`
 * over a list in the view, and a stack handed to an element by path read.
 * `tools/probe-release.mjs` reads the result.
 */
import {
  computed,
  type Default,
  handler,
  lift,
  NAME,
  pattern,
  type Stream,
  UI,
  type VNode,
  type Writable,
} from "commonfabric";

import "../elements.ts";
import type { Proposal, TableView } from "../../patterns/toolkit/view.ts";
import { memoryInfo } from "../../patterns/memory/game.ts";
import {
  move,
  type SealedTable,
  viewFor,
} from "../../patterns/memory/table.tsx";
import type { MemoryState } from "./state.ts";

/** A module-scope lift over the view: named by no rule of the table. */
const revisionOf = lift((view: TableView | undefined): number =>
  view?.revision ?? -1
);

const sit = handler<unknown, { move: Stream<Proposal> }>((_, { move }) => {
  move.send({ kind: "sit", label: "Probe" });
});
const deal = handler<unknown, { move: Stream<Proposal> }>((_, { move }) => {
  move.send({ kind: "start", options: { numcards: "small" } });
});

export default pattern<
  { table: Writable<Default<SealedTable, null>> },
  { [NAME]: string; [UI]: VNode }
>(({ table }) => {
  const view = viewFor({ table });
  // The same released value, read with the state's shape named: the table
  // releases its state as plain JSON, and a path read needs the path's type.
  const typed = view as TableView<MemoryState> | undefined;
  const propose = move({ table });
  const info = computed(() => memoryInfo());
  const derived = computed(() => view?.revision ?? -1);
  const lifted = revisionOf(view);
  return {
    [NAME]: "Release probe",
    [UI]: (
      <div>
        <cf-button id="sit" onClick={sit({ move: propose })}>Sit</cf-button>
        <cf-button id="deal" onClick={deal({ move: propose })}>Deal</cf-button>
        <div>direct: <span id="direct">{view?.revision}</span></div>
        <div>viewer: <span id="viewer">{view?.viewer}</span></div>
        <div>derived: <span id="derived">{derived}</span></div>
        <div>lifted: <span id="lifted">{lifted}</span></div>
        <div id="seats">
          {view?.seats.map((seat) => <span className="seat">{seat.label}</span>)}
        </div>
        <bg-table name="probe" gameName="memory" chest={info.chest}>
          <bg-stack
            name="direct"
            stack={typed?.entries[0]?.state.Game.Cards}
            layout="grid"
            face="Type"
          />
        </bg-table>
      </div>
    ),
  };
});
