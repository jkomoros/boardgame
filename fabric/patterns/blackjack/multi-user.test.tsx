/**
 * Three players and a watcher at one table of blackjack, each in their own
 * runtime under their own principal.
 *
 * Alice, Bob and Carol take seats 0, 1 and 2. Dave never sits. The table has
 * one view, computed once for each principal who looks, so each participant
 * here reads their own instance of it and plays from it: hit below fifteen,
 * stand otherwise, counting their own hole card, which only their own
 * instance shows.
 *
 * What is checked:
 *
 * - A participant's view holds nobody else's hidden card. At every step each
 *   checks their own instance at every version it holds: every other seat's
 *   hole card is the hidden card, and the draw stack shows no face.
 * - A player's own view does show their own hole card, and a watcher's shows
 *   nobody's.
 * - Proposals the table must refuse change nothing: a move out of turn, a
 *   move whose payload names another player as its proposer, a move by
 *   someone with no seat, and an engine-owned move proposed by a player.
 * - The game plays out to a winner.
 *
 * A participant has no way to read another's instance, which is why nothing
 * here compares two views. That a hole card one seat is shown appears in no
 * other seat's view of the same version is checked under Node
 * (`parity.test.ts`), and what each principal is shown at a display boundary
 * in a browser (`tools/play-three-users.mjs`).
 *
 * Run with `EXPERIMENTAL_VIEWER_BOUND_RELEASE=true`.
 *
 * Pattern under test: ./table.tsx
 */
import {
  action,
  assert,
  multiUserTest,
  pattern,
  type Stream,
  TESTS,
  Writable,
} from "commonfabric";

import type { Proposal, TableView } from "../toolkit/view.ts";
import BlackjackTablePattern, {
  type BlackjackTableOutput,
} from "./table.tsx";

const PLAYERS = 3;

/**
 * Rounds enough to finish a one-round game: each player hits at most a
 * handful of times and stands once, and every round of this test gives each
 * player one step.
 */
const ROUNDS = 12;

interface Setup {
  table: BlackjackTableOutput;
}

export const setup = pattern<Record<string, never>, Setup>(() => ({
  table: BlackjackTablePattern(
    {} as Parameters<typeof BlackjackTablePattern>[0],
  ),
}));

//
// Reading a view
//

type Entry = TableView["entries"][number];

/** The part of blackjack's state these tests read, as a viewer is sent it. */
interface Shown {
  Game: { DrawStack: { Indexes: number[] } };
  Players: {
    HiddenHand: { Indexes: number[] };
    Hand: { Indexes: number[] };
  }[];
  Computed: { Players: { HandValue: number }[] };
}

function shown(entry: Entry): Shown {
  const state: unknown = entry.state;
  return state as Shown;
}

function head(view: TableView | undefined): Entry | undefined {
  const entries = view?.entries ?? [];
  return entries[entries.length - 1];
}

/**
 * Whether `view`, computed for the seat `viewer` (-1 for a watcher), shows no
 * other seat's hole card and no card of the draw stack, at any version it
 * holds.
 */
function keepsSecrets(view: TableView | undefined, viewer: number): boolean {
  if (view === undefined) return false;
  for (const entry of view.entries) {
    const state = shown(entry);
    if (state.Game.DrawStack.Indexes.some((index) => index >= 0)) return false;
    for (let seat = 0; seat < state.Players.length; seat++) {
      if (seat === viewer) continue;
      if (state.Players[seat].HiddenHand.Indexes.some((index) => index >= 0)) {
        return false;
      }
    }
  }
  return true;
}

/** Whether `view`, a seat's own, shows that seat its hole card whenever it holds one. */
function showsOwnHoleCard(view: TableView | undefined, seat: number): boolean {
  if (view === undefined) return false;
  return view.entries.every((entry) =>
    shown(entry).Players[seat].HiddenHand.Indexes.every((index) => index >= 0)
  );
}

/** What the player in `seat` does at the head of their own view, if anything. */
function nextProposal(
  view: TableView | undefined,
  seat: number,
): Proposal | undefined {
  const entry = head(view);
  if (view === undefined || entry === undefined) return undefined;
  if (view.phase !== "playing" || entry.currentPlayerIndex !== seat) {
    return undefined;
  }
  const legality = entry.viewers[String(seat)]?.moveLegality ?? {};
  const value = shown(entry).Computed.Players[seat]?.HandValue ?? 0;
  if (value < 15 && legality["Current Player Hit"]?.legalForPlayer) {
    return { kind: "play", name: "Current Player Hit", arguments: {} };
  }
  if (legality["Current Player Stand"]?.legalForPlayer) {
    return { kind: "play", name: "Current Player Stand", arguments: {} };
  }
  return undefined;
}

/** The steps of one player's rounds, each waiting for the player before. */
function rounds(
  step: Stream<void>,
  secret: ReturnType<typeof assert>,
  own: ReturnType<typeof assert>,
  me: string,
  before: string,
  first: boolean,
) {
  const steps = [];
  for (let round = 1; round <= ROUNDS; round++) {
    steps.push(
      { await: first ? `${before}-${round - 1}` : `${before}-${round}` },
      { action: step },
      { assertion: secret },
      { assertion: own },
      { label: `${me}-${round}` },
    );
  }
  return steps;
}

const finished = (view: TableView | undefined) => {
  const entry = head(view);
  return view?.phase === "finished" && entry?.outcome.finished === true &&
    (entry?.outcome.winners.length ?? 0) >= 1;
};

//
// Alice: seat 0
//

export const alice = pattern<{ setup: Setup }>(({ setup }) => {
  const table = setup.table;
  const SEAT = 0;
  const noteRevision = new Writable(-1);
  const sit = action(() => table.move.send({ kind: "sit", label: "Alice" }));
  const deal = action(() =>
    table.move.send({ kind: "start", options: { maxrounds: "1" } })
  );
  const note = action(() => noteRevision.set(table.view?.revision ?? -1));
  const propose_engine_move = action(() =>
    table.move.send({
      kind: "play",
      name: "Deal Initial Hidden Card",
      arguments: {},
    })
  );
  const step = action(() => {
    const proposal = nextProposal(table.view, 0);
    if (proposal !== undefined) table.move.send(proposal);
  });

  const three_seats = assert(() =>
    table.view?.seats.length === PLAYERS &&
    table.view?.seats[2]?.label === "Carol"
  );
  const dealt = assert(() =>
    table.view?.phase === "playing" &&
    (head(table.view)?.version ?? 0) > 5 &&
    head(table.view)?.currentPlayerIndex === 0
  );
  const unmoved = assert(() => table.view?.revision === noteRevision.get());
  const secret = assert(() => keepsSecrets(table.view, SEAT));
  const own = assert(() => showsOwnHoleCard(table.view, 0));
  const over = assert(() => finished(table.view));

  return {
    [TESTS]: [
      { action: sit },
      { label: "alice-sat" },
      { await: "carol-sat" },
      { assertion: three_seats },
      { action: deal },
      { assertion: dealt },
      { assertion: secret },
      { assertion: own },
      { action: note },
      { action: propose_engine_move },
      { assertion: unmoved },
      { label: "alice-dealt" },
      { await: "dave-refused" },
      { assertion: unmoved },
      { label: "carol-0" },
      ...rounds(step, secret, own, "alice", "carol", true),
      { await: `carol-${ROUNDS}` },
      { assertion: over },
      { assertion: secret },
    ],
  };
});

//
// Bob: seat 1
//

export const bob = pattern<{ setup: Setup }>(({ setup }) => {
  const table = setup.table;
  const SEAT = 1;
  const noteRevision = new Writable(-1);
  const sit = action(() => table.move.send({ kind: "sit", label: "Bob" }));
  const note = action(() => noteRevision.set(table.view?.revision ?? -1));
  const hit_out_of_turn = action(() =>
    table.move.send({ kind: "play", name: "Current Player Hit", arguments: {} })
  );
  // The payload names Alice, whose turn it is, as the proposer. The table
  // reads no such field.
  const hit_as_alice = action(() => {
    const forged = {
      kind: "play" as const,
      name: "Current Player Hit",
      arguments: {},
      proposer: table.view?.seats[0]?.principal,
      acting: table.view?.seats[0]?.principal,
      viewer: 0,
      seat: 0,
    };
    table.move.send(forged as Proposal);
  });
  const step = action(() => {
    const proposal = nextProposal(table.view, 1);
    if (proposal !== undefined) table.move.send(proposal);
  });

  const unmoved = assert(() =>
    table.view?.phase === "playing" &&
    table.view?.revision === noteRevision.get()
  );
  const secret = assert(() => keepsSecrets(table.view, SEAT));
  const own = assert(() => showsOwnHoleCard(table.view, 1));
  const over = assert(() => finished(table.view));

  return {
    [TESTS]: [
      { await: "alice-sat" },
      { action: sit },
      { label: "bob-sat" },
      { await: "alice-dealt" },
      { assertion: secret },
      { assertion: own },
      { action: note },
      { action: hit_out_of_turn },
      { action: hit_as_alice },
      { assertion: unmoved },
      { label: "bob-refused" },
      ...rounds(step, secret, own, "bob", "alice", false),
      { await: `carol-${ROUNDS}` },
      { assertion: over },
    ],
  };
});

//
// Carol: seat 2
//

export const carol = pattern<{ setup: Setup }>(({ setup }) => {
  const table = setup.table;
  const SEAT = 2;
  const sit = action(() => table.move.send({ kind: "sit", label: "Carol" }));
  const step = action(() => {
    const proposal = nextProposal(table.view, 2);
    if (proposal !== undefined) table.move.send(proposal);
  });
  const secret = assert(() => keepsSecrets(table.view, SEAT));
  const own = assert(() => showsOwnHoleCard(table.view, 2));
  const over = assert(() => finished(table.view));

  return {
    [TESTS]: [
      { await: "bob-sat" },
      { action: sit },
      { label: "carol-sat" },
      { await: "alice-dealt" },
      { assertion: secret },
      { assertion: own },
      ...rounds(step, secret, own, "carol", "bob", false),
      { assertion: over },
      { assertion: secret },
    ],
  };
});

//
// Dave: no seat
//

export const dave = pattern<{ setup: Setup }>(({ setup }) => {
  const table = setup.table;
  const noteRevision = new Writable(-1);
  const note = action(() => noteRevision.set(table.view?.revision ?? -1));
  const sit_late = action(() => table.move.send({ kind: "sit" }));
  const hit_unseated = action(() =>
    table.move.send({ kind: "play", name: "Current Player Hit", arguments: {} })
  );
  const unmoved = assert(() =>
    table.view?.phase === "playing" &&
    table.view?.seats.length === PLAYERS &&
    table.view?.revision === noteRevision.get()
  );
  // A watcher's view shows nobody's hole card.
  const secret = assert(() => keepsSecrets(table.view, -1));
  const over = assert(() => finished(table.view));

  return {
    [TESTS]: [
      { await: "alice-dealt" },
      { await: "bob-refused" },
      { action: note },
      { action: sit_late },
      { action: hit_unseated },
      { assertion: unmoved },
      { assertion: secret },
      { label: "dave-refused" },
      { await: `carol-${ROUNDS}` },
      { assertion: over },
      { assertion: secret },
    ],
  };
});

export default multiUserTest({
  setup,
  participants: {
    alice,
    bob,
    carol,
    dave: { pattern: dave, access: "READ" },
  },
});
