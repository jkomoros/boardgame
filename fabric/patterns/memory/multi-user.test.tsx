/**
 * Two players and a watcher at one table of Memory, each in their own
 * runtime under their own principal.
 *
 * Alice takes seat 0 and Bob seat 1. Carol never sits. Before play, each
 * kind of proposal the table must refuse is sent and the table's revision is
 * checked not to have moved: a deal by someone not seated, a move out of
 * turn, a move whose payload names another player as its proposer, a move by
 * someone with no seat, a fix-up move proposed by a player, and reveals of a
 * slot that does not exist or is already face-up. Then the two play the game
 * out.
 *
 * Neither player can read the layout, so neither is scripted: each plays
 * from the released view and from notes of the faces it has shown. A round
 * gives each player one step, in which the player whose turn it is makes one
 * move and the other makes none.
 *
 * Every step checks that no view shows the face of a hidden card. The two
 * players' views are checked against each other through the moves: before a
 * player's step each of them works out, from their own view and notes, which
 * card that player will turn, and afterwards checks that it was the one
 * turned. They agree only while their views do.
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

import type { Proposal } from "../toolkit/view.ts";
import MemoryTablePattern, {
  type MemoryTableOutput,
  type MemoryTableView,
} from "./table.tsx";

/** `small` deals ten cards, five pairs. */
const PAIRS = 5;

/**
 * Rounds enough to finish any game of `PAIRS` pairs. A turn is at most three
 * proposals (two reveals and a hide), and with a perfect memory either
 * captures a pair or turns a card nobody had seen, so there are at most
 * `3 * PAIRS` turns.
 */
const ROUNDS = 9 * PAIRS;

interface Setup {
  table: MemoryTableOutput;
}

export const setup = pattern<Record<string, never>, Setup>(() => ({
  table: MemoryTablePattern(
    {} as Parameters<typeof MemoryTablePattern>[0],
  ),
}));

//
// Reading a view
//

type Entry = MemoryTableView["entries"][number];

/** The part of Memory's state these tests read, as a viewer is sent it. */
interface Shown {
  Game: {
    HiddenCards: { Indexes: number[] };
    VisibleCards: { Indexes: number[] };
    Cards: { Indexes: number[] };
  };
  Players: { WonCards: { Indexes: number[] } }[];
}

/** A view carries the engine's state as JSON; this names what is read. */
function shown(entry: Entry): Shown {
  const state: unknown = entry.state;
  return state as Shown;
}

function head(view: MemoryTableView | undefined): Entry | undefined {
  const entries = view?.entries ?? [];
  return entries[entries.length - 1];
}

/**
 * Whether no entry of `view` shows a hidden card's face: every card in the
 * hidden stack is the hidden card, and the board shows a face only at a slot
 * whose card is in the visible stack.
 */
function showsNoHiddenFace(view: MemoryTableView | undefined): boolean {
  if (view === undefined) return false;
  for (const entry of view.entries) {
    const game = shown(entry).Game;
    if (game.HiddenCards.Indexes.some((index) => index >= 0)) return false;
    for (let slot = 0; slot < game.Cards.Indexes.length; slot++) {
      const shown = game.Cards.Indexes[slot] >= 0;
      if (shown !== (game.VisibleCards.Indexes[slot] >= 0)) return false;
    }
  }
  return true;
}

/** The pair a deck index belongs to: two cards of a type sit side by side. */
const pairOf = (deckIndex: number) => Math.floor(deckIndex / 2);

/** `seen` with the pair of every face-up card in `view` noted at its slot. */
function noted(seen: readonly number[], view: MemoryTableView): number[] {
  const next = seen.slice();
  for (const entry of view.entries) {
    shown(entry).Game.VisibleCards.Indexes.forEach((deckIndex, slot) => {
      if (deckIndex >= 0) next[slot] = pairOf(deckIndex);
    });
  }
  return next;
}

/**
 * The proposal a player with a perfect memory makes at the head of `view`,
 * given the pairs it has `seen` by slot, or `undefined` when it is not
 * `seat`'s turn or the game is over.
 */
function nextProposal(
  view: MemoryTableView,
  seen: readonly number[],
  seat: number,
): Proposal | undefined {
  const entry = head(view);
  if (entry === undefined || view.phase !== "playing") return undefined;
  if (entry.currentPlayerIndex !== seat) return undefined;
  const part = entry.viewers[String(seat)];
  if (part?.moveLegality["Hide Cards"]?.legalForPlayer) {
    return { kind: "play", name: "Hide Cards", arguments: {} };
  }
  if (!part?.moveLegality["Reveal Card"]?.legalForPlayer) return undefined;
  const game = shown(entry).Game;
  const hidden: number[] = [];
  game.HiddenCards.Indexes.forEach((index, slot) => {
    if (index !== -1) hidden.push(slot);
  });
  const known = (slot: number) => seen[slot] !== undefined && seen[slot] >= 0;
  const unseen = hidden.filter((slot) => !known(slot));
  const turned = game.VisibleCards.Indexes.findIndex((index) => index >= 0);
  let slot: number | undefined;
  if (turned >= 0) {
    const want = pairOf(game.VisibleCards.Indexes[turned]);
    slot = hidden.find((other) => known(other) && seen[other] === want);
  } else {
    slot = hidden.find((one) =>
      known(one) &&
      hidden.some((other) => other !== one && seen[other] === seen[one])
    );
  }
  slot = slot ?? unseen[0] ?? hidden[0];
  if (slot === undefined) return undefined;
  return {
    kind: "play",
    name: "Reveal Card",
    arguments: { CardIndex: String(slot) },
  };
}

/**
 * The slots turned face-up by reveals after version `since`, in order, or
 * `undefined` when the view does not hold every version after `since` and
 * so cannot say. `visibleAtSince` is what was face-up at `since`.
 */
function revealedSince(
  view: MemoryTableView,
  since: number,
  visibleAtSince: readonly number[],
): number[] | undefined {
  const slots: number[] = [];
  let before: readonly number[] = visibleAtSince;
  let expected = since + 1;
  for (const entry of view.entries) {
    if (entry.version <= since) continue;
    if (entry.version !== expected) return undefined;
    expected += 1;
    const visible = shown(entry).Game.VisibleCards.Indexes;
    if (entry.move?.AnimationKey === "Reveal Card") {
      const previous = before;
      visible.forEach((index, slot) => {
        if (index >= 0 && (previous[slot] ?? -1) < 0) slots.push(slot);
      });
    }
    before = visible;
  }
  return slots;
}

//
// A player
//

/** What a player remembers between steps. */
interface Notes {
  /** The pair seen at each slot, by slot. */
  seen: number[];
  /** The head version at the last step. */
  lastVersion: number;
  /** What was face-up at that version, by slot. */
  lastVisible: number[];
  /** The slot this player worked out would be turned next, or -1. */
  expected: number;
  /** Whether every card turned so far was the one worked out. */
  agreed: boolean;
  /** How many turned cards were compared. */
  checked: number;
}

const NO_NOTES: Notes = {
  seen: [],
  lastVersion: 0,
  lastVisible: [],
  expected: -1,
  agreed: true,
  checked: 0,
};

/**
 * One step of the player in `seat`: check the card turned since the last
 * step against the one worked out then, note what is face-up, work out what
 * whoever's turn it is will do, and propose it when the turn is this seat's.
 */
function stepped(
  view: MemoryTableView | undefined,
  seat: number,
  notes: Notes,
): { notes: Notes; proposal?: Proposal } {
  const entry = head(view);
  if (view === undefined || entry === undefined) return { notes };
  let { agreed, checked } = notes;
  const turned = revealedSince(view, notes.lastVersion, notes.lastVisible);
  if (notes.expected >= 0 && turned !== undefined && turned.length > 0) {
    if (turned[0] !== notes.expected) agreed = false;
    checked += 1;
  }
  const seen = noted(notes.seen, view);
  const mover = entry.currentPlayerIndex;
  const proposal = nextProposal(view, seen, mover);
  const next: Notes = {
    seen,
    lastVersion: entry.version,
    lastVisible: shown(entry).Game.VisibleCards.Indexes.slice(),
    expected: proposal?.name === "Reveal Card"
      ? Number(proposal.arguments?.["CardIndex"])
      : -1,
    agreed,
    checked,
  };
  return mover === seat && proposal !== undefined
    ? { notes: next, proposal }
    : { notes: next };
}

/** The steps of one player's rounds, each waiting for the other's step. */
function rounds(
  step: Stream<void>,
  sound: ReturnType<typeof assert>,
  agrees: ReturnType<typeof assert>,
  me: string,
  other: string,
  first: boolean,
) {
  const steps = [];
  for (let round = 1; round <= ROUNDS; round++) {
    steps.push(
      { await: first ? `${other}-${round - 1}` : `${other}-${round}` },
      { action: step },
      { assertion: sound },
      { assertion: agrees },
      { label: `${me}-${round}` },
    );
  }
  return steps;
}

const dealtCards = (view: MemoryTableView | undefined) => {
  const entry = view?.entries[0];
  return entry === undefined ? -1 : shown(entry).Game.HiddenCards.Indexes.length;
};

const finishedWithEveryCardWon = (view: MemoryTableView | undefined) => {
  const entry = head(view);
  if (view === undefined || entry === undefined) return false;
  const won = shown(entry).Players.reduce(
    (sum, seat) => sum + seat.WonCards.Indexes.length,
    0,
  );
  return view.phase === "finished" && entry.outcome.finished &&
    won === 2 * PAIRS && entry.outcome.winners.length >= 1;
};

//
// Alice: seat 0
//

export const alice = pattern<{ setup: Setup }>(({ setup }) => {
  const table = setup.table;
  const notes = new Writable<Notes>(NO_NOTES);
  const step = action(() => {
    const result = stepped(table.view, 0, notes.get());
    notes.set(result.notes);
    if (result.proposal !== undefined) table.move.send(result.proposal);
  });
  const sound = assert(() => showsNoHiddenFace(table.view));
  const agrees = assert(() => notes.get().agreed === true);
  const checkedOften = assert(() => notes.get().checked >= PAIRS);
  const noteRevision = new Writable(-1);

  const sit = action(() => table.move.send({ kind: "sit", label: "Alice" }));
  const deal_alone = action(() => table.move.send({ kind: "start" }));
  const sit_again = action(() => table.move.send({ kind: "sit" }));
  const deal = action(() =>
    table.move.send({ kind: "start", options: { numcards: "small" } })
  );
  const note = action(() => noteRevision.set(table.view?.revision ?? -1));
  const propose_fix_up = action(() =>
    table.move.send({ kind: "play", name: "Capture Cards", arguments: {} })
  );
  const reveal_nowhere = action(() =>
    table.move.send({
      kind: "play",
      name: "Reveal Card",
      arguments: { CardIndex: "99" },
    })
  );
  const hide_too_soon = action(() =>
    table.move.send({ kind: "play", name: "Hide Cards", arguments: {} })
  );
  const unknown_kind = action(() => {
    const cheat = { kind: "cheat" };
    table.move.send(cheat as Proposal);
  });

  const one_seat = assert(() =>
    table.view?.seats.length === 1 && table.view?.seats[0]?.label === "Alice" &&
    table.view?.phase === "seating" && table.view?.revision === 1
  );
  const two_seats = assert(() =>
    table.view?.seats.length === 2 && table.view?.seats[1]?.label === "Bob" &&
    table.view?.phase === "seating"
  );
  const dealt = assert(() =>
    table.view?.phase === "playing" && table.view?.version === 0 &&
    table.view?.entries.length === 1 &&
    table.view?.entries[0]?.currentPlayerIndex === 0 &&
    dealtCards(table.view) === 2 * PAIRS
  );
  const unmoved = assert(() => table.view?.revision === noteRevision.get());
  const over = assert(() => finishedWithEveryCardWon(table.view));

  return {
    [TESTS]: [
      { action: sit },
      { assertion: one_seat },
      // One player is not enough, and a seat cannot be taken twice.
      { action: deal_alone },
      { action: sit_again },
      { assertion: one_seat },
      { label: "alice-sat" },
      { await: "bob-sat" },
      { assertion: two_seats },
      { await: "carol-tried-to-deal" },
      { assertion: two_seats },
      { action: deal },
      { assertion: dealt },
      { assertion: sound },
      { action: note },
      // Refused: a fix-up move, a slot off the board, a hide with nothing
      // face-up, and a proposal of no known kind.
      { action: propose_fix_up },
      { action: reveal_nowhere },
      { action: hide_too_soon },
      { action: unknown_kind },
      { assertion: unmoved },
      { label: "alice-dealt" },
      { await: "bob-refused" },
      { await: "carol-refused" },
      { assertion: unmoved },
      { label: "alice-0" },
      ...rounds(
        step,
        sound,
        agrees,
        "alice",
        "bob",
        false,
      ),
      { await: `bob-${ROUNDS}` },
      { assertion: over },
      { assertion: agrees },
      { assertion: checkedOften },
    ],
  };
});

//
// Bob: seat 1
//

export const bob = pattern<{ setup: Setup }>(({ setup }) => {
  const table = setup.table;
  const notes = new Writable<Notes>(NO_NOTES);
  const step = action(() => {
    const result = stepped(table.view, 1, notes.get());
    notes.set(result.notes);
    if (result.proposal !== undefined) table.move.send(result.proposal);
  });
  const sound = assert(() => showsNoHiddenFace(table.view));
  const agrees = assert(() => notes.get().agreed === true);
  const checkedOften = assert(() => notes.get().checked >= PAIRS);
  const noteRevision = new Writable(-1);

  const sit = action(() => table.move.send({ kind: "sit", label: " Bob " }));
  const note = action(() => noteRevision.set(table.view?.revision ?? -1));
  const reveal_out_of_turn = action(() =>
    table.move.send({
      kind: "play",
      name: "Reveal Card",
      arguments: { CardIndex: "0" },
    })
  );
  // The payload names Alice, whose turn it is, as the proposer. The table
  // reads no such field.
  const reveal_as_alice = action(() => {
    const forged = {
      kind: "play" as const,
      name: "Reveal Card",
      arguments: { CardIndex: "0" },
      proposer: table.view?.seats[0]?.principal,
      acting: table.view?.seats[0]?.principal,
      seat: 0,
    };
    table.move.send(forged as Proposal);
  });
  const leave_mid_game = action(() => table.move.send({ kind: "leave" }));
  const tick_no_timer = action(() =>
    table.move.send({ kind: "tick", timerId: "NOSUCHTIMER" })
  );

  const seated = assert(() =>
    table.view?.seats.length === 2 && table.view?.seats[1]?.label === "Bob"
  );
  const unmoved = assert(() =>
    table.view?.phase === "playing" &&
    table.view?.revision === noteRevision.get()
  );
  const over = assert(() => finishedWithEveryCardWon(table.view));

  return {
    [TESTS]: [
      { await: "alice-sat" },
      { action: sit },
      { assertion: seated },
      { label: "bob-sat" },
      { await: "alice-dealt" },
      { assertion: sound },
      { action: note },
      { action: reveal_out_of_turn },
      { action: reveal_as_alice },
      { action: leave_mid_game },
      { action: tick_no_timer },
      { assertion: unmoved },
      { label: "bob-refused" },
      ...rounds(
        step,
        sound,
        agrees,
        "bob",
        "alice",
        true,
      ),
      { assertion: over },
      { assertion: agrees },
      { assertion: checkedOften },
    ],
  };
});

//
// Carol: no seat
//

export const carol = pattern<{ setup: Setup }>(({ setup }) => {
  const table = setup.table;
  const noteRevision = new Writable(-1);

  const note = action(() => noteRevision.set(table.view?.revision ?? -1));
  const deal_unseated = action(() => table.move.send({ kind: "start" }));
  const sit_late = action(() => table.move.send({ kind: "sit" }));
  const reveal_unseated = action(() =>
    table.move.send({
      kind: "play",
      name: "Reveal Card",
      arguments: { CardIndex: "0" },
    })
  );
  const reveal_as_alice = action(() => {
    const forged = {
      kind: "play" as const,
      name: "Reveal Card",
      arguments: { CardIndex: "1" },
      proposer: table.view?.seats[0]?.principal,
    };
    table.move.send(forged as Proposal);
  });

  const still_seating = assert(() =>
    table.view?.phase === "seating" &&
    table.view?.revision === noteRevision.get()
  );
  const unmoved = assert(() =>
    table.view?.phase === "playing" && table.view?.seats.length === 2 &&
    table.view?.revision === noteRevision.get()
  );
  const sound = assert(() => showsNoHiddenFace(table.view));
  const over = assert(() => finishedWithEveryCardWon(table.view));

  return {
    [TESTS]: [
      { await: "bob-sat" },
      { action: note },
      { action: deal_unseated },
      { assertion: still_seating },
      { label: "carol-tried-to-deal" },
      { await: "alice-dealt" },
      { await: "bob-refused" },
      { action: note },
      { action: sit_late },
      { action: reveal_unseated },
      { action: reveal_as_alice },
      { assertion: unmoved },
      { assertion: sound },
      { label: "carol-refused" },
      { await: `bob-${ROUNDS}` },
      { assertion: over },
      { assertion: sound },
    ],
  };
});

export default multiUserTest({
  setup,
  participants: { alice, bob, carol: { pattern: carol, access: "READ" } },
});
