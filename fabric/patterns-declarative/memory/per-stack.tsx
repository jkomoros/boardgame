/**
 * EXPERIMENT, not the recommended shape. The same table as `main.tsx`, drawn
 * the way a pattern author would write it without knowing the findings: the
 * pattern picks the newest version out of the released view and hands each
 * stack its own value, and the won piles are a `.map` over the players.
 *
 * It exists to see what the real reconciler does with it: whether a derived
 * value is still released, whether the elements are recreated, whether the
 * stacks of one version arrive together, and whether a card still flies.
 * `tools/play-two-users.mjs` reports all four.
 */
import {
  computed,
  currentPrincipal,
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
import type { ElementEvent, ProposeMoveDetail } from "../elements.ts";
import type { Proposal } from "../../patterns/toolkit/view.ts";
import { memoryInfo } from "../../patterns/memory/game.ts";
import {
  EMPTY_TABLE,
  type MemoryTableView,
  move,
  type SealedTable,
  seatView,
} from "../../patterns/memory/table.tsx";
import type { MemoryState } from "./state.ts";

export interface MemoryInput {
  table: Writable<Default<SealedTable, typeof EMPTY_TABLE>>;
  me?: PerUser<Writable<Default<string, "">>>;
  seatName?: PerUser<Writable<Default<string, "">>>;
}

export interface MemoryOutput {
  [NAME]: string;
  [UI]: VNode;
  view: MemoryTableView | undefined;
  move: Stream<Proposal>;
}

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

const sit = handler<
  unknown,
  {
    move: Stream<Proposal>;
    me: Writable<string>;
    seatName: Writable<string>;
  }
>((_, { move, me, seatName }) => {
  const principal = currentPrincipal();
  if (principal !== undefined && me.get() !== principal) me.set(principal);
  move.send({ kind: "sit", label: seatName.get() || "Player" });
});

const deal = handler<unknown, { move: Stream<Proposal> }>((_, { move }) => {
  move.send({ kind: "start", options: { numcards: "small" } });
});

/** Nobody clicks to hide a mismatch here, so a button asks the table to. */
const hide = handler<unknown, { move: Stream<Proposal> }>((_, { move }) => {
  move.send({ kind: "play", name: "Hide Cards", arguments: {} });
});

const PILE = {
  display: "inline-block",
  width: "120px",
  minHeight: "150px",
  margin: "0 12px",
};

export default pattern<MemoryInput, MemoryOutput>(
  ({ table, me, seatName }) => {
    const view = seatView(table);
    const propose = move({ table });
    const info = computed(() => memoryInfo());
    // The newest version, picked by the pattern.
    const head = computed(() => {
      const entries = view?.entries ?? [];
      return entries.length > 0 ? entries[entries.length - 1] : undefined;
    });
    const version = computed(() => head?.version ?? -1);
    const cards = computed(() =>
      (head?.state as MemoryState | undefined)?.Game.Cards
    );
    const players = computed(() =>
      (head?.state as MemoryState | undefined)?.Players ?? []
    );
    return {
      [NAME]: "Memory (per-stack experiment)",
      [UI]: (
        <div style={{ padding: "12px", fontFamily: "system-ui, sans-serif" }}>
          <div id="lobby" style={{ marginBottom: "12px" }}>
            <cf-input $value={seatName} placeholder="Your name" />
            <cf-button id="sit" onClick={sit({ move: propose, me, seatName })}>
              Sit down
            </cf-button>
            <cf-button id="deal" onClick={deal({ move: propose })}>
              Deal
            </cf-button>
            <cf-button id="hide" onClick={hide({ move: propose })}>
              Hide cards
            </cf-button>
            <span id="version">{version}</span>
          </div>
          <bg-table
            name="memory-per-stack"
            gameName="memory"
            chest={info.chest}
            animationLength={400}
            onpropose-move={play({ move: propose })}
          >
            <div style={{ position: "relative", minHeight: "320px" }}>
              <bg-stack
                name="Game.Cards"
                stack={cards}
                version={version}
                layout="grid"
                messy
                face="Type"
                pick={{ move: "Reveal Card", arg: "CardIndex", slots: "hidden" }}
              />
            </div>
            <div style={{ margin: "16px 0" }}>
              {players.map((player, index) => (
                <div style={PILE}>
                  <bg-stack
                    name={`Players.${index}.WonCards`}
                    stack={player.WonCards}
                    layout="stack"
                    messy
                    face="Type"
                  />
                </div>
              ))}
            </div>
          </bg-table>
        </div>
      ),
      view,
      move: propose,
    };
  },
);
