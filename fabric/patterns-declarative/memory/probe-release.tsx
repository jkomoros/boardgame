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
  NAME,
  pattern,
  type Stream,
  UI,
  type VNode,
  type Writable,
} from "commonfabric";

import "../elements.ts";
import type { Proposal } from "../../patterns/toolkit/view.ts";
import { memoryInfo } from "../../patterns/memory/rules.ts";
import {
  EMPTY_TABLE,
  move,
  type SealedTable,
  seatView,
} from "../../patterns/memory/table.tsx";

const sit = handler<unknown, { move: Stream<Proposal> }>((_, { move }) => {
  move.send({ kind: "sit", label: "Probe" });
});
const deal = handler<unknown, { move: Stream<Proposal> }>((_, { move }) => {
  move.send({ kind: "start" });
});

export default pattern<
  { table: Writable<Default<SealedTable, typeof EMPTY_TABLE>> },
  { [NAME]: string; [UI]: VNode }
>(({ table }) => {
  const view = seatView(table);
  const propose = move({ table });
  const info = computed(() => memoryInfo());
  const derived = computed(() => view?.revision ?? -1);
  return {
    [NAME]: "Release probe",
    [UI]: (
      <div>
        <cf-button id="sit" onClick={sit({ move: propose })}>Sit</cf-button>
        <cf-button id="deal" onClick={deal({ move: propose })}>Deal</cf-button>
        <div>direct: <span id="direct">{view?.revision}</span></div>
        <div>derived: <span id="derived">{derived}</span></div>
        <div id="seats">
          {view?.seats.map((seat) => <span className="seat">{seat.label}</span>)}
        </div>
        <bg-table name="probe" gameName="memory" chest={info.chest}>
          <bg-stack
            name="direct"
            stack={view?.entries[0]?.state.Game.Cards}
            layout="grid"
            face="Type"
          />
        </bg-table>
      </div>
    ),
  };
});
