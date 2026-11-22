/**
 * The table around Memory, with the clock and the randomness supplied by the
 * test: seating and dealing as proposals, every kind of refusal, card ids
 * across a reveal, and the hide-cards timer against a clock the test moves.
 *
 *   node --test fabric/patterns/memory/table.test.ts
 *
 * That the game itself is Go's is `parity.test.ts`. That all of this holds
 * through the sealed cell and the runtime's own principals is
 * `multi-user.test.tsx`.
 */
import assert from "node:assert/strict";
import { test } from "node:test";

import {
  emptyTable,
  held,
  stepTable,
  type TableState,
  viewForPrincipal,
} from "../toolkit/table.ts";
import type { Proposal, TableView } from "../toolkit/view.ts";
import { memoryGame } from "./game.ts";

const GAME = memoryGame();
const HIDE_CARDS_MS = 4000;

interface Stack {
  Indexes: number[];
  IDs: string[];
  IDsLastSeen: Record<string, number>;
}
interface Shown {
  Game: {
    HiddenCards: Stack;
    VisibleCards: Stack;
    Cards: Stack;
    HideCardsTimer: { ID: string };
    CurrentPlayer: number;
  };
  Players: { WonCards: Stack; CardsLeftToReveal: number }[];
  Computed: { Players: { GameScore: number }[] };
}
const shownState = (entry: TableView["entries"][number]): Shown =>
  entry.state as unknown as Shown;

/** A table driven by named people, with a clock the test moves. */
class Harness {
  table: TableState = emptyTable();
  now = 1_700_000_000_000;
  private seed: number;

  constructor(seed = 1) {
    this.seed = seed;
  }

  /** A small deterministic generator, so a failure can be replayed. */
  private entropy = (): number => {
    this.seed = (this.seed * 1103515245 + 12345) % 2147483648;
    return this.seed / 2147483648;
  };

  /** Proposes as `who`; returns the refusal, or `""`. */
  propose(who: string, proposal: Proposal): string {
    const result = stepTable(GAME, this.table, proposal, {
      proposer: who,
      now: this.now,
      entropy: this.entropy,
    });
    this.table = result.table;
    return result.error;
  }

  /** What `who` is shown: their seat's view, or a watcher's. */
  viewAs(who: string | undefined): TableView {
    return viewForPrincipal(GAME, this.table, who);
  }

  /**
   * A watcher's view, with every seated person's own legality and choices
   * beside the watcher's. No one is shown this: each sees their own part.
   */
  get view(): TableView {
    const view = this.viewAs(undefined);
    for (const seat of this.table.seats) {
      this.viewAs(seat.principal).entries.forEach((entry, at) => {
        Object.assign(view.entries[at].viewers, entry.viewers);
      });
    }
    return view;
  }

  get head() {
    const entries = this.view.entries;
    return entries[entries.length - 1];
  }

  get shown(): Shown {
    return shownState(this.head);
  }

  /** The unsanitized head, which only the test may look at. */
  get secret(): Shown {
    return held(this.table).at(-1)!.state as unknown as Shown;
  }

  reveal(who: string, slot: number): string {
    return this.propose(who, {
      kind: "play",
      name: "Reveal Card",
      arguments: { CardIndex: String(slot) },
    });
  }

  static dealt(seed = 1): Harness {
    const harness = new Harness(seed);
    assert.equal(harness.propose("alice", { kind: "sit", label: "Alice" }), "");
    assert.equal(harness.propose("bob", { kind: "sit", label: "Bob" }), "");
    assert.equal(
      harness.propose("alice", {
        kind: "start",
        options: { numcards: "small" },
      }),
      "",
    );
    return harness;
  }

  /** Two hidden slots holding different types, and two holding the same. */
  slots(): { mismatch: [number, number]; match: [number, number] } {
    const hidden = this.secret.Game.HiddenCards.Indexes;
    const type = (slot: number) => Math.floor(hidden[slot] / 2);
    const present = hidden.map((_, slot) => slot).filter((slot) =>
      hidden[slot] >= 0
    );
    const first = present[0];
    const partner = present.find((slot) =>
      slot !== first && type(slot) === type(first)
    )!;
    const other = present.find((slot) => type(slot) !== type(first))!;
    return { mismatch: [first, other], match: [first, partner] };
  }
}

function assertNoHiddenFace(view: TableView): void {
  for (const entry of view.entries) {
    const game = shownState(entry).Game;
    assert.ok(
      game.HiddenCards.Indexes.every((index) => index === -1 || index === -2),
      `version ${entry.version} shows a hidden card's face`,
    );
    game.Cards.Indexes.forEach((index, slot) => {
      assert.equal(index >= 0, game.VisibleCards.Indexes[slot] >= 0);
    });
  }
}

test("seating and the deal are proposals like any other", () => {
  const harness = new Harness();
  assert.equal(harness.view.phase, "seating");
  assert.match(harness.propose("alice", { kind: "start" }), /seated/);
  assert.equal(harness.propose("alice", { kind: "sit", label: " Alice " }), "");
  assert.match(harness.propose("alice", { kind: "sit" }), /already/);
  assert.match(harness.propose("alice", { kind: "start" }), /not enough/);
  assert.equal(harness.propose("bob", { kind: "sit" }), "");
  assert.match(harness.propose("carol", { kind: "start" }), /seated/);
  assert.deepEqual(harness.view.seats, [
    { label: "Alice" },
    { label: "Player 2" },
  ]);
  assert.equal(harness.propose("bob", { kind: "leave" }), "");
  assert.equal(harness.propose("bob", { kind: "sit", label: "Bob" }), "");
  assert.equal(
    harness.propose("bob", { kind: "start", options: { numcards: "small" } }),
    "",
  );
  assert.equal(harness.view.phase, "playing");
  assert.equal(harness.view.version, 0);
  assert.match(harness.propose("carol", { kind: "sit" }), /started/);
  assert.match(harness.propose("bob", { kind: "leave" }), /started/);
  assert.match(harness.propose("", { kind: "sit" }), /no one/);
});

test("the deal depends only on the drawn identity", () => {
  const one = Harness.dealt(1);
  const hidden = one.secret.Game.HiddenCards;
  assert.equal(hidden.Indexes.length, 10);
  const types = hidden.Indexes.map((index) => Math.floor(index / 2)).sort();
  for (let i = 0; i < types.length; i += 2) {
    assert.equal(types[i], types[i + 1]);
  }
  assert.equal(new Set(hidden.IDs).size, 10);
  assert.ok(hidden.IDs.every((id) => /^[0-9a-f]{40}$/.test(id)));

  const again = Harness.dealt(1);
  assert.deepEqual(again.secret, one.secret);
  const other = Harness.dealt(2);
  assert.notDeepEqual(other.secret.Game.HiddenCards.IDs, hidden.IDs);
  assert.notEqual(other.view.epoch, one.view.epoch);
});

test("the view of a fresh deal shows no face and offers every slot", () => {
  const harness = Harness.dealt();
  assertNoHiddenFace(harness.view);
  const head = harness.head;
  assert.equal(head.move, undefined);
  assert.equal(head.currentPlayerIndex, 0);
  assert.deepEqual(harness.shown.Game.HiddenCards.Indexes, new Array(10).fill(-2));
  assert.deepEqual(harness.shown.Game.HiddenCards.IDs, harness.shown.Game.Cards.IDs);
  assert.equal(head.viewers["0"].moveLegality["Reveal Card"].legalForPlayer, true);
  assert.equal(head.viewers["1"].moveLegality["Reveal Card"].legalForPlayer, false);
  assert.equal(head.viewers["1"].moveLegality["Reveal Card"].legalForAnyone, true);
  assert.equal(head.viewers["-1"].moveLegality["Reveal Card"].legalForPlayer, false);
  assert.deepEqual(head.viewers["0"].zeroInputMoves, ["Hide Cards"]);
  assert.equal(head.viewers["0"].projectedMoveChoices.Sets[0].Candidates.length, 10);
  assert.equal(head.viewers["1"].projectedMoveChoices.Sets.length, 0);
  // Nothing of the salt or of the layout is in what readers get.
  const text = JSON.stringify(harness.view);
  assert.ok(!text.includes(harness.table.record.secretSalt));
  assert.ok(!text.includes("SecretMoveCount"));
  assert.ok(!text.includes('"Timers"'));
});

test("illegal and out-of-turn moves are refused, in Go's words, and change nothing", () => {
  const harness = Harness.dealt();
  const before = JSON.stringify(harness.table);
  assert.equal(harness.reveal("bob", 0), "it's not your turn");
  assert.match(harness.reveal("carol", 0), /no seat/);
  assert.match(harness.reveal("alice", 99), /not an occupied slot|no card/);
  assert.match(
    harness.propose("alice", {
      kind: "play",
      name: "Reveal Card",
      arguments: { CardIndex: "1; drop" },
    }),
    /cannot unmarshal string/,
  );
  assert.equal(
    harness.propose("alice", { kind: "play", name: "Hide Cards" }),
    "You still have to reveal more cards before your turn is over",
  );
  for (const name of ["Capture Cards", "Finish Turn", "Start Hide Cards Timer"]) {
    assert.match(
      harness.propose("alice", { kind: "play", name }),
      /engine-owned/,
    );
  }
  assert.match(
    harness.propose("alice", { kind: "play", name: "No Such Move" }),
    /unknown move/,
  );
  assert.match(
    harness.propose("alice", {
      kind: "play",
      name: "Reveal Card",
      arguments: { CardIndex: "0", TargetPlayerIndex: "1" },
    }),
    /does not accept/,
  );
  assert.match(harness.propose("alice", { kind: "play" }), /no name/);
  assert.match(harness.propose("alice", { kind: "restart" }), /not over/);
  assert.match(
    harness.propose("alice", { kind: "tick", timerId: "X" }),
    /stale|already handled/,
  );
  assert.match(harness.propose("alice", { kind: "tick" }), /names no timer/);
  assert.equal(JSON.stringify(harness.table), before);
});

test("a card keeps its id when it is turned over and back", () => {
  const harness = Harness.dealt();
  const { mismatch } = harness.slots();
  const id = harness.shown.Game.HiddenCards.IDs[mismatch[0]];
  assert.equal(harness.reveal("alice", mismatch[0]), "");
  let game = harness.shown.Game;
  assert.equal(game.VisibleCards.IDs[mismatch[0]], id);
  assert.equal(game.HiddenCards.IDs[mismatch[0]], "");
  assert.equal(game.Cards.IDs[mismatch[0]], id);
  assert.ok(game.VisibleCards.Indexes[mismatch[0]] >= 0);
  assert.equal(game.VisibleCards.IDsLastSeen[id], 1);
  assert.match(harness.reveal("alice", mismatch[0]), /not an occupied slot/);
  assert.equal(harness.reveal("alice", mismatch[1]), "");
  assert.match(harness.reveal("alice", 5), /no cards left/);
  assert.equal(harness.propose("alice", { kind: "play", name: "Hide Cards" }), "");
  game = harness.shown.Game;
  assert.equal(game.HiddenCards.IDs[mismatch[0]], id);
  assert.equal(game.HiddenCards.Indexes[mismatch[0]], -2);
  assertNoHiddenFace(harness.view);
});

test("a mismatch starts the timer; a tick is checked against the clock", () => {
  const harness = Harness.dealt();
  const { mismatch } = harness.slots();
  harness.reveal("alice", mismatch[0]);
  assert.equal(harness.reveal("alice", mismatch[1]), "");
  assert.deepEqual(
    harness.view.entries.map((entry) => [
      entry.version,
      entry.move?.AnimationKey,
      entry.settled,
    ]),
    // The versions of the last input: the reveal and its one fix-up.
    [
      [2, "Reveal Card", false],
      [3, "Start Hide Cards Timer", true],
    ],
  );
  const head = harness.head;
  const timerId = harness.shown.Game.HideCardsTimer.ID;
  assert.match(timerId, /^[0-9A-F]{16}$/);
  assert.deepEqual(head.timers, { [timerId]: { TimeLeft: HIDE_CARDS_MS } });
  assert.equal(head.viewers["0"].moveLegality["Hide Cards"].legalForPlayer, true);

  // Anyone may ask; nobody may make the clock run faster.
  const before = harness.table.revision;
  assert.equal(
    harness.propose("bob", { kind: "tick", timerId }),
    "timer deadline has not passed",
  );
  harness.now += HIDE_CARDS_MS - 1;
  assert.equal(
    harness.propose("carol", { kind: "tick", timerId }),
    "timer deadline has not passed",
  );
  assert.equal(harness.table.revision, before);

  harness.now += 1;
  assert.equal(harness.propose("carol", { kind: "tick", timerId }), "");
  assert.deepEqual(
    harness.view.entries.slice(-2).map((entry) => entry.move?.AnimationKey),
    ["Hide Cards", "Finish Turn"],
  );
  assert.equal(harness.head.currentPlayerIndex, 1);
  assert.equal(harness.shown.Game.HideCardsTimer.ID, "");
  assert.deepEqual(harness.head.timers, {});
  assert.equal(harness.table.history.at(-2)?.["Proposer"], -2);
  // The same tick again, as a second browser would send it.
  assert.match(
    harness.propose("bob", { kind: "tick", timerId }),
    /stale|already handled/,
  );
  assertNoHiddenFace(harness.view);
});

test("hiding by hand cancels the timer", () => {
  const harness = Harness.dealt();
  const { mismatch } = harness.slots();
  harness.reveal("alice", mismatch[0]);
  harness.reveal("alice", mismatch[1]);
  const timerId = harness.shown.Game.HideCardsTimer.ID;
  assert.equal(
    harness.propose("bob", { kind: "play", name: "Hide Cards" }),
    "it's not your turn",
  );
  assert.equal(harness.propose("alice", { kind: "play", name: "Hide Cards" }), "");
  harness.now += 2 * HIDE_CARDS_MS;
  assert.match(
    harness.propose("alice", { kind: "tick", timerId }),
    /stale|already handled/,
  );
});

test("a match is captured and the turn passes", () => {
  const harness = Harness.dealt();
  const { match } = harness.slots();
  harness.reveal("alice", match[0]);
  assert.equal(harness.reveal("alice", match[1]), "");
  assert.deepEqual(
    harness.view.entries.slice(-3).map((entry) => entry.move?.AnimationKey),
    ["Reveal Card", "Capture Cards", "Finish Turn"],
  );
  assert.equal(harness.shown.Players[0].WonCards.Indexes.length, 2);
  assert.equal(harness.shown.Computed.Players[0].GameScore, 2);
  assert.equal(harness.head.currentPlayerIndex, 1);
  assert.equal(harness.shown.Players[1].CardsLeftToReveal, 2);
});

test("a whole game: one input held, full history, a winner, a new deal", () => {
  const harness = Harness.dealt(7);
  const names = ["alice", "bob"];
  let guard = 0;
  while (harness.view.phase === "playing" && guard++ < 500) {
    const head = harness.head;
    const mover = head.currentPlayerIndex;
    const part = head.viewers[String(mover)];
    if (part.moveLegality["Hide Cards"].legalForPlayer) {
      harness.now += HIDE_CARDS_MS;
      const timerId = Object.keys(head.timers)[0] ??
        harness.shown.Game.HideCardsTimer.ID;
      assert.equal(harness.propose("bob", { kind: "tick", timerId }), "");
    } else {
      const { match, mismatch } = harness.slots();
      const visible = harness.shown.Game.VisibleCards.Indexes;
      const turned = visible.findIndex((index) => index >= 0);
      const wantMatch = harness.table.version % 4 < 2;
      const hidden = harness.secret.Game.HiddenCards.Indexes;
      // Alternate: complete a pair on some turns, miss on others.
      const slot = turned >= 0
        ? hidden.findIndex((index, at) =>
          index >= 0 && at !== turned &&
          wantMatch === (Math.floor(index / 2) === Math.floor(visible[turned] / 2))
        )
        : (wantMatch ? match[0] : mismatch[0]);
      const pick = slot >= 0 ? slot : hidden.findIndex((index) => index >= 0);
      assert.equal(harness.reveal(names[mover], pick), "");
    }
    assertNoHiddenFace(harness.view);
    // One input's versions at most: a reveal with its two fix-ups.
    assert.ok(harness.view.entries.length <= 3);
    assert.equal(harness.table.history.length, harness.table.version);
  }
  const view = harness.view;
  assert.equal(view.phase, "finished");
  const head = harness.head;
  assert.equal(head.outcome.finished, true);
  const scores = harness.shown.Players.map((seat) => seat.WonCards.Indexes.length);
  assert.equal(scores[0] + scores[1], 10);
  assert.deepEqual(
    head.outcome.winners,
    scores[0] === scores[1] ? [0, 1] : [scores[0] > scores[1] ? 0 : 1],
  );
  assert.match(harness.reveal("alice", 0), /no game/);
  assert.deepEqual(
    harness.table.history.map((entry) => entry["Version"]),
    harness.table.history.map((_, index) => index + 1),
  );

  const epoch = view.epoch;
  assert.match(harness.propose("carol", { kind: "restart" }), /seated/);
  assert.equal(harness.propose("bob", { kind: "restart" }), "");
  assert.equal(harness.view.phase, "seating");
  assert.equal(harness.view.seats.length, 2);
  assert.equal(harness.propose("bob", { kind: "start" }), "");
  assert.notEqual(harness.view.epoch, epoch);
  assert.equal(harness.view.version, 0);
});

test("a principal is shown their own seat's view, and anyone else a watcher's", () => {
  // Nothing stored yet: the table before anyone sat down, for everyone.
  const before = viewForPrincipal(GAME, null, "alice");
  assert.equal(before.phase, "seating");
  assert.deepEqual(before.seats, []);
  assert.equal(before.revision, 0);
  assert.equal(before.viewer, -1);
  assert.deepEqual(viewForPrincipal(GAME, undefined, undefined), before);

  const harness = Harness.dealt();
  const keys = (who: string | undefined) =>
    Object.keys(harness.viewAs(who).entries.at(-1)!.viewers);
  // A view says which seat it was computed for, and names no principal.
  assert.equal(harness.viewAs("alice").viewer, 0);
  assert.equal(harness.viewAs("bob").viewer, 1);
  assert.equal(harness.viewAs("carol").viewer, -1);
  assert.ok(!JSON.stringify(harness.viewAs("alice")).includes("alice"));
  assert.deepEqual(keys("alice"), ["0"]);
  assert.deepEqual(keys("bob"), ["1"]);
  assert.deepEqual(keys("carol"), ["-1"]);
  assert.deepEqual(keys(undefined), ["-1"]);
  assert.deepEqual(keys(""), ["-1"]);
  // A view holds its own viewer's legality and nobody else's.
  const alice = harness.viewAs("alice").entries.at(-1)!;
  assert.equal(alice.viewers["0"].moveLegality["Reveal Card"].legalForPlayer, true);
  const bob = harness.viewAs("bob").entries.at(-1)!;
  assert.equal(bob.viewers["1"].moveLegality["Reveal Card"].legalForPlayer, false);
});
