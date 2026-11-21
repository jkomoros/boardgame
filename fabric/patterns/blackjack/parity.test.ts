/**
 * Blackjack on a table is blackjack in Go, and no seat is shown another
 * seat's hidden card.
 *
 *   node --test fabric/patterns/blackjack/parity.test.ts
 *
 * Every recorded Go game of blackjack in `conformance/` is replayed through
 * the functions this pattern's `move` handler and view lifts call, and at
 * every version the stored state, the move record and every viewer's whole
 * output are compared with the recording (`../toolkit/parity.ts`). The
 * recordings include a table that seats its players one by one, a game that
 * runs the deck out and reshuffles, and seven players.
 */
import assert from "node:assert/strict";
import { test } from "node:test";

import { type Recording, replayRecording } from "../toolkit/parity.ts";
import { asACellStoresIt, recordingsOf } from "../toolkit/recordings.ts";
import {
  dealTable,
  emptyTable,
  held,
  OBSERVER_SEAT,
  seatViewOf,
  stepTable,
  type TableState,
} from "../toolkit/table.ts";
import type { TableView } from "../toolkit/view.ts";
import { blackjackGame, blackjackInfo, MAX_SEATS } from "./game.ts";

const GAME = blackjackGame();
const recordings = recordingsOf("blackjack");

test("the corpus has blackjack recordings", () => {
  assert.ok(recordings.length >= 7, `found ${recordings.length}`);
  assert.equal(GAME.spec.maxNumPlayers, MAX_SEATS);
});

for (const recording of recordings) {
  test(`parity: ${recording.name}`, (context) => {
    const result = replayRecording(GAME, recording);
    assert.deepEqual(result.problems, []);
    assert.equal(result.versions, recording.versions.length);
    assert.equal(
      result.viewerVersions,
      recording.versions.reduce((sum, version) => sum + version.views.length, 0),
    );
    context.diagnostic(
      `${result.versions} versions, ${result.viewerVersions} viewer outputs, ${result.comparisons} comparisons`,
    );
  });
}

test("parity through a cell's copy", () => {
  // The shortest recording and the one that seats its players.
  for (const recording of recordings.filter((one) => one.versions.length < 45)) {
    const result = replayRecording(GAME, recording, asACellStoresIt);
    assert.deepEqual(result.problems, [], recording.name);
  }
});

//
// What a seat is shown
//

interface Hand {
  Indexes: number[];
  IDs: string[];
}
interface Shown {
  Game: { DrawStack: Hand; DiscardStack: Hand };
  Players: { HiddenHand: Hand; VisibleHand: Hand; Hand: Hand }[];
}
const shownState = (entry: TableView["entries"][number]): Shown =>
  entry.state as unknown as Shown;

/** A recording replayed to its end with every version still held. */
function replayed(recording: Recording): TableState {
  let table = emptyTable();
  for (let seat = 0; seat < recording.game.numPlayers; seat++) {
    table.seats.push({ principal: `p${seat}`, label: `P${seat}` });
  }
  // Hold every version, however many inputs the game takes.
  const inputsHeld = 1_000_000;
  table = dealTable(GAME, table, {
    id: recording.game.id,
    secretSalt: recording.game.secretSalt,
  }, {
    numPlayers: recording.game.numPlayers,
    variant: recording.game.variant,
    seating: false,
    now: 0,
    inputsHeld,
  }).table;
  // Play it out by proposal alone: every player stands when it is their turn.
  for (let guard = 0; guard < 400 && table.phase === "playing"; guard++) {
    const head = seatViewOf(GAME, table, OBSERVER_SEAT).entries.at(-1)!;
    const step = stepTable(GAME, table, {
      kind: "play",
      name: guard % 3 === 0 ? "Current Player Hit" : "Current Player Stand",
    }, {
      proposer: `p${head.currentPlayerIndex}`,
      now: 0,
      entropy: () => 0,
      inputsHeld,
    });
    const next = step.error === ""
      ? step.table
      : stepTable(GAME, table, { kind: "play", name: "Current Player Stand" }, {
        proposer: `p${head.currentPlayerIndex}`,
        now: 0,
        entropy: () => 0,
        inputsHeld,
      }).table;
    if (next === table) break;
    table = next;
  }
  return table;
}

test("no seat's view holds another seat's hidden card, at any version", () => {
  let hiddenCardsSeen = 0;
  let viewsChecked = 0;
  for (const recording of recordings.filter((one) => !one.game.seating).slice(0, 4)) {
    const table = replayed(recording);
    const versions = held(table);
    assert.ok(versions.length > 20, `${recording.name}: ${versions.length} versions`);
    const players = recording.game.numPlayers;
    for (let viewer = OBSERVER_SEAT; viewer < players; viewer++) {
      const view = seatViewOf(GAME, table, viewer);
      assert.equal(view.entries.length, versions.length);
      assert.deepEqual(Object.keys(view.entries[0].viewers), [String(viewer)]);
      view.entries.forEach((entry, at) => {
        viewsChecked += 1;
        const shown = shownState(entry);
        const secret = versions[at].state as unknown as Shown;
        const where = `${recording.name} v${entry.version} viewer ${viewer}`;
        // The draw and discard stacks show nobody a face.
        assert.ok(shown.Game.DrawStack.Indexes.every((index) => index < 0), where);
        assert.ok(shown.Game.DiscardStack.Indexes.every((index) => index < 0), where);
        for (let seat = 0; seat < players; seat++) {
          const hidden = secret.Players[seat].HiddenHand.Indexes;
          const theirs = shown.Players[seat].HiddenHand.Indexes;
          assert.equal(theirs.length, hidden.length, where);
          if (seat === viewer) {
            // A player sees their own hole card.
            assert.deepEqual(theirs, hidden, where);
            continue;
          }
          for (const card of hidden) {
            hiddenCardsSeen += 1;
            assert.ok(theirs.every((index) => index < 0), `${where}: seat ${seat}'s hole card`);
            // Nor anywhere else in what this viewer is sent.
            assert.ok(
              !shown.Players[seat].Hand.Indexes.includes(card),
              `${where}: seat ${seat}'s merged hand shows the hole card`,
            );
          }
        }
      });
    }
  }
  assert.ok(hiddenCardsSeen > 500, `only ${hiddenCardsSeen} hidden cards checked`);
  assert.ok(viewsChecked > 300, `only ${viewsChecked} views checked`);
});

test("a seat's own view differs from the observer's exactly while it holds a hole card", () => {
  const recording = recordings.find((one) => !one.game.seating)!;
  const table = replayed(recording);
  const observer = seatViewOf(GAME, table, OBSERVER_SEAT);
  const mine = seatViewOf(GAME, table, 0);
  let differed = 0;
  mine.entries.forEach((entry, at) => {
    const holds = shownState(entry).Players[0].HiddenHand.Indexes.some((index) => index >= 0);
    const same = JSON.stringify(shownState(entry).Players[0]) ===
      JSON.stringify(shownState(observer.entries[at]).Players[0]);
    assert.equal(same, !holds, `v${entry.version}`);
    if (holds) differed += 1;
  });
  assert.ok(differed > 0);
});

test("the public facts are the recorded chest", () => {
  const info = blackjackInfo();
  const recorded = recordings[0] as unknown as {
    chest: { Decks: unknown; Enums: unknown; LegalTemplates: unknown };
    engine: { moveInputSchemaFingerprint: string };
  };
  assert.deepEqual(info.chest["Decks"], recorded.chest.Decks);
  assert.deepEqual(info.chest["LegalTemplates"], recorded.chest.LegalTemplates);
  assert.equal(info.moveInputSchemaFingerprint, recorded.engine.moveInputSchemaFingerprint);
});
