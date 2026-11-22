/**
 * Memory on a table is Memory in Go. Every recorded Go game of Memory in
 * `conformance/` is replayed through the functions this pattern's `move`
 * handler and view lift call, and at every version the stored state, the
 * move record and every viewer's whole output are compared with the
 * recording (`../toolkit/parity.ts`).
 *
 *   node --test fabric/patterns/memory/parity.test.ts
 *
 * Each recording is replayed twice: once passing the table straight on, and
 * once passing it through a copy with every object's keys reordered, the way
 * a Common Fabric cell hands a stored value back.
 */
import assert from "node:assert/strict";
import { test } from "node:test";

import { replayRecording } from "../toolkit/parity.ts";
import { asACellStoresIt, recordingsOf } from "../toolkit/recordings.ts";
import { memoryGame, memoryInfo } from "./game.ts";

const recordings = recordingsOf("memory");

test("the corpus has Memory recordings", () => {
  assert.ok(recordings.length >= 6, `found ${recordings.length}`);
});

for (const recording of recordings) {
  test(`parity: ${recording.name}`, (context) => {
    const result = replayRecording(memoryGame(), recording);
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

  test(`parity through a cell's copy: ${recording.name}`, () => {
    const result = replayRecording(memoryGame(), recording, asACellStoresIt);
    assert.deepEqual(result.problems, []);
    assert.equal(result.versions, recording.versions.length);
  });
}

test("every seat of Memory is shown the observer's state", () => {
  // Memory has no hands: a seat's view differs from a watcher's only in
  // what the seat may do. The recordings confirm it.
  for (const recording of recordings) {
    for (const version of recording.versions) {
      const observer = version.views.find((view) => view.viewer === -1);
      assert.ok(observer, `${recording.name} v${version.version}`);
      for (const view of version.views) {
        if (view.viewer < 0) continue;
        const mine = view.game as { CurrentState: unknown };
        const theirs = observer.game as { CurrentState: unknown };
        assert.deepEqual(
          mine.CurrentState,
          theirs.CurrentState,
          `${recording.name} v${version.version} seat ${view.viewer}`,
        );
      }
    }
  }
});

test("the public facts are the recorded chest", () => {
  const info = memoryInfo();
  const recorded = recordings[0] as unknown as {
    chest: { Decks: unknown; LegalTemplates: unknown };
    engine: { moveInputSchemaFingerprint: string };
  };
  assert.deepEqual(info.chest["Decks"], recorded.chest.Decks);
  assert.deepEqual(info.chest["LegalTemplates"], recorded.chest.LegalTemplates);
  assert.equal(
    info.moveInputSchemaFingerprint,
    recorded.engine.moveInputSchemaFingerprint,
  );
});
