/**
 * Memory's rules and the table around them, with the clock and the
 * randomness supplied by the test. Run with Node 24:
 *
 *   node --test fabric/patterns/memory/rules.test.ts
 *
 * The pattern test beside this file (`multi-user.test.tsx`) plays the same
 * rules through the sealed cell and the runtime's own principals. This one
 * covers what needs a clock that can be moved: the hide-cards timer.
 */
import assert from "node:assert/strict";
import { test } from "node:test";

import {
  emptyTable,
  stepTable,
  type TableState,
  viewOf,
  WINDOW,
} from "../toolkit/table.ts";
import type { Proposal, TableView } from "../toolkit/view.ts";
import {
  HIDE_CARDS_MS,
  type MemoryGame,
  memoryRules,
  type MemoryView,
} from "./rules.ts";

const RULES = memoryRules();

/** A table driven by named people, with a clock the test moves. */
class Harness {
  table: TableState<MemoryGame> = emptyTable();
  now = 1_000_000;
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
    const result = stepTable(RULES, this.table, proposal, {
      proposer: who,
      now: this.now,
      entropy: this.entropy,
    });
    this.table = result.table;
    return result.error;
  }

  get view(): TableView<MemoryView> {
    return viewOf(RULES, this.table);
  }

  get head() {
    const entries = this.view.entries;
    return entries[entries.length - 1];
  }

  /** The unsanitized head, which only the test may look at. */
  get secret(): MemoryGame {
    return this.table.recent[this.table.recent.length - 1].game;
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

function assertNoHiddenFace(view: TableView<MemoryView>): void {
  for (const entry of view.entries) {
    const game = entry.state.Game;
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
    { principal: "alice", label: "Alice" },
    { principal: "bob", label: "Player 2" },
  ]);
  assert.equal(harness.propose("bob", { kind: "leave" }), "");
  assert.equal(harness.propose("bob", { kind: "sit", label: "Bob" }), "");
  assert.equal(harness.propose("bob", { kind: "start" }), "");
  assert.equal(harness.view.phase, "playing");
  assert.equal(harness.view.version, 0);
  assert.match(harness.propose("carol", { kind: "sit" }), /started/);
  assert.match(harness.propose("bob", { kind: "leave" }), /started/);
  assert.match(harness.propose("", { kind: "sit" }), /no one/);
});

test("the deal lays out whole pairs and depends only on the salt", () => {
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
  assert.deepEqual(
    head.state.Game.HiddenCards.Indexes,
    new Array(10).fill(-2),
  );
  assert.deepEqual(head.state.Game.HiddenCards.IDs, head.state.Game.Cards.IDs);
  assert.equal(head.viewers["0"].moveLegality["Reveal Card"].legalForPlayer, true);
  assert.equal(head.viewers["1"].moveLegality["Reveal Card"].legalForPlayer, false);
  assert.equal(head.viewers["1"].moveLegality["Reveal Card"].legalForAnyone, true);
  assert.equal(head.viewers["-1"].moveLegality["Reveal Card"].legalForPlayer, false);
  assert.equal(head.viewers["0"].projectedMoveChoices.Sets[0].Candidates.length, 10);
  assert.equal(head.viewers["1"].projectedMoveChoices.Sets.length, 0);
  // Nothing of the salt or the layout is in what readers get.
  const text = JSON.stringify(harness.view);
  assert.ok(!text.includes(harness.table.secretSalt));
  assert.ok(!text.includes("SecretMoveCount"));
});

test("illegal and out-of-turn moves are refused and change nothing", () => {
  const harness = Harness.dealt();
  const before = JSON.stringify(harness.table);
  assert.match(harness.reveal("bob", 0), /not your turn/);
  assert.match(harness.reveal("carol", 0), /no seat/);
  assert.match(harness.reveal("alice", 99), /no card/);
  assert.match(harness.reveal("alice", -1), /no card/);
  assert.match(
    harness.propose("alice", {
      kind: "play",
      name: "Reveal Card",
      arguments: { CardIndex: "1; drop" },
    }),
    /no card/,
  );
  assert.match(
    harness.propose("alice", { kind: "play", name: "Hide Cards" }),
    /reveal more/,
  );
  for (const name of ["Capture Cards", "Finish Turn", "Start Hide Cards Timer"]) {
    assert.match(
      harness.propose("alice", { kind: "play", name }),
      /cannot be proposed/,
    );
  }
  assert.match(harness.propose("alice", { kind: "play" }), /no name/);
  assert.match(harness.propose("alice", { kind: "restart" }), /not over/);
  assert.match(
    harness.propose("alice", { kind: "tick", timerId: "X" }),
    /no such timer/,
  );
  assert.equal(JSON.stringify(harness.table), before);
});

test("a card keeps its id when it is turned over and back", () => {
  const harness = Harness.dealt();
  const { mismatch } = harness.slots();
  const id = harness.head.state.Game.HiddenCards.IDs[mismatch[0]];
  assert.equal(harness.reveal("alice", mismatch[0]), "");
  let game = harness.head.state.Game;
  assert.equal(game.VisibleCards.IDs[mismatch[0]], id);
  assert.equal(game.HiddenCards.IDs[mismatch[0]], "");
  assert.equal(game.Cards.IDs[mismatch[0]], id);
  assert.ok(game.VisibleCards.Indexes[mismatch[0]] >= 0);
  assert.equal(game.VisibleCards.IDsLastSeen[id], 1);
  assert.match(harness.reveal("alice", mismatch[0]), /already been revealed/);
  assert.equal(harness.reveal("alice", mismatch[1]), "");
  assert.match(harness.reveal("alice", 5), /no cards left/);
  assert.equal(harness.propose("alice", { kind: "play", name: "Hide Cards" }), "");
  game = harness.head.state.Game;
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
    [
      [0, undefined, true],
      [1, "Reveal Card", true],
      [2, "Reveal Card", false],
      [3, "Start Hide Cards Timer", true],
    ],
  );
  const head = harness.head;
  const timerId = head.state.Game.HideCardsTimer.ID;
  assert.match(timerId, /^[0-9A-F]{16}$/);
  assert.deepEqual(head.timers, { [timerId]: { TimeLeft: HIDE_CARDS_MS } });
  assert.equal(head.viewers["0"].moveLegality["Hide Cards"].legalForPlayer, true);

  // Anyone may ask; nobody may make the clock run faster.
  const before = harness.table.revision;
  assert.match(
    harness.propose("bob", { kind: "tick", timerId }),
    /not run out/,
  );
  harness.now += HIDE_CARDS_MS - 1;
  assert.match(
    harness.propose("carol", { kind: "tick", timerId }),
    /not run out/,
  );
  assert.equal(harness.table.revision, before);

  harness.now += 1;
  assert.equal(harness.propose("carol", { kind: "tick", timerId }), "");
  assert.deepEqual(
    harness.view.entries.slice(-2).map((entry) => entry.move?.AnimationKey),
    ["Hide Cards", "Finish Turn"],
  );
  assert.equal(harness.head.currentPlayerIndex, 1);
  assert.equal(harness.head.state.Game.HideCardsTimer.ID, "");
  assert.deepEqual(harness.head.timers, {});
  assert.equal(harness.table.timers.length, 0);
  assert.equal(harness.table.history.at(-2)?.seat, -2);
  // The same tick again, as a second browser would send it.
  assert.match(
    harness.propose("bob", { kind: "tick", timerId }),
    /no such timer/,
  );
  assertNoHiddenFace(harness.view);
});

test("hiding by hand cancels the timer", () => {
  const harness = Harness.dealt();
  const { mismatch } = harness.slots();
  harness.reveal("alice", mismatch[0]);
  harness.reveal("alice", mismatch[1]);
  const timerId = harness.head.state.Game.HideCardsTimer.ID;
  assert.match(
    harness.propose("bob", { kind: "play", name: "Hide Cards" }),
    /not your turn/,
  );
  assert.equal(harness.propose("alice", { kind: "play", name: "Hide Cards" }), "");
  assert.equal(harness.table.timers.length, 0);
  harness.now += 2 * HIDE_CARDS_MS;
  assert.match(
    harness.propose("alice", { kind: "tick", timerId }),
    /no such timer/,
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
  const head = harness.head;
  assert.equal(head.state.Players[0].WonCards.Indexes.length, 2);
  assert.equal(head.state.Computed.Players[0].GameScore, 2);
  assert.equal(head.currentPlayerIndex, 1);
  assert.equal(head.state.Players[1].CardsLeftToReveal, 2);
  assert.equal(harness.table.timers.length, 0);
});

test("a whole game: bounded window, full history, a winner, a new deal", () => {
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
        head.state.Game.HideCardsTimer.ID;
      assert.equal(harness.propose("bob", { kind: "tick", timerId }), "");
    } else {
      const { match, mismatch } = harness.slots();
      const turned = head.state.Game.VisibleCards.Indexes.findIndex((index) =>
        index >= 0
      );
      // Alternate: complete a pair on even turns, miss on odd ones.
      const slot = turned >= 0
        ? harness.secret.Game.HiddenCards.Indexes.findIndex((index, at) =>
          index >= 0 &&
          (harness.table.version % 4 < 2) ===
            (Math.floor(index / 2) ===
              Math.floor(head.state.Game.VisibleCards.Indexes[turned] / 2)) &&
          at !== turned
        )
        : (harness.table.version % 4 < 2 ? match[0] : mismatch[0]);
      const pick = slot >= 0
        ? slot
        : harness.secret.Game.HiddenCards.Indexes.findIndex((index) =>
          index >= 0
        );
      assert.equal(harness.reveal(names[mover], pick), "");
    }
    assertNoHiddenFace(harness.view);
    assert.ok(harness.view.entries.length <= WINDOW);
    assert.equal(harness.table.history.length, harness.table.version);
  }
  const view = harness.view;
  assert.equal(view.phase, "finished");
  const head = harness.head;
  assert.equal(head.outcome.finished, true);
  const scores = head.state.Players.map((seat) => seat.WonCards.Indexes.length);
  assert.equal(scores[0] + scores[1], 10);
  assert.deepEqual(
    head.outcome.winners,
    scores[0] === scores[1] ? [0, 1] : [scores[0] > scores[1] ? 0 : 1],
  );
  assert.match(harness.reveal("alice", 0), /no game/);
  assert.deepEqual(
    harness.table.history.map((entry) => entry.version),
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
