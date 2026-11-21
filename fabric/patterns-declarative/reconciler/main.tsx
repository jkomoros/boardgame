/**
 * EXPERIMENT. What the real reconciler does to stack elements whose values
 * the pattern computes, with nothing confidential involved, so that policy
 * cannot be what decides the result.
 *
 * Six face-up cards start in a grid. Each press of "Next" moves one more of
 * them to one of two piles. The grid's stack is a computed value bound to a
 * prop; the piles are a `.map` over a computed list, as an author would
 * write a variable number of seats.
 *
 * It is the real counterpart of `fabric/declarative/memory-recreate.html`,
 * which simulated a renderer that recreates children.
 * `tools/probe-reconciler.mjs` drives it and reports.
 */
import {
  computed,
  type Default,
  handler,
  NAME,
  pattern,
  UI,
  type VNode,
  type Writable,
} from "commonfabric";

import "../elements.ts";

const FACES = ["A", "B", "C", "D", "E", "F"];
const CHEST = {
  Decks: {
    cards: FACES.map((face, index) => ({ Index: index, Values: { Type: face } })),
  },
};

interface Wire {
  Deck: string;
  Indexes: number[];
  IDs: string[];
}

/** The grid after `step` cards have left it: a sized stack with gaps. */
function gridAt(step: number): Wire {
  return {
    Deck: "cards",
    Indexes: FACES.map((_, index) => (index < step ? -1 : index)),
    IDs: FACES.map((_, index) => (index < step ? "" : `card-${index}`)),
  };
}

/** Pile `pile` after `step` cards have left the grid, alternating piles. */
function pileAt(step: number, pile: number): Wire {
  const held = FACES.map((_, index) => index)
    .filter((index) => index < step && index % 2 === pile);
  return {
    Deck: "cards",
    Indexes: held,
    IDs: held.map((index) => `card-${index}`),
  };
}

const next = handler<unknown, { step: Writable<number> }>((_, { step }) => {
  step.set(Math.min(FACES.length, step.get() + 1));
});

const PILE = { display: "inline-block", width: "120px", minHeight: "150px", margin: "0 12px" };

export default pattern<
  { step: Writable<Default<number, 0>> },
  { [NAME]: string; [UI]: VNode }
>(({ step }) => {
  const grid = computed(() => gridAt(step.get()));
  const piles = computed(() => [0, 1].map((pile) => ({ pile, stack: pileAt(step.get(), pile) })));
  return {
    [NAME]: "Reconciler probe",
    [UI]: (
      <div style={{ padding: "12px" }}>
        <cf-button id="next" onClick={next({ step })}>Next</cf-button>
        <span id="step">{step}</span>
        <bg-table name="reconciler-probe" gameName="probe" chest={CHEST} animationLength={400}>
          <div style={{ minHeight: "180px" }}>
            <bg-stack name="grid" stack={grid} layout="grid" face="Type" />
          </div>
          <div>
            {piles.map((entry) => (
              <div style={PILE}>
                <bg-stack name={`pile-${entry.pile}`} stack={entry.stack} layout="stack" messy face="Type" />
              </div>
            ))}
          </div>
        </bg-table>
      </div>
    ),
  };
});
