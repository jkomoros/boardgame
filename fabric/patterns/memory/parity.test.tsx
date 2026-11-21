/**
 * Memory is Go's Memory inside the pattern sandbox too.
 *
 * `parity.test.ts` replays every recorded Go game of Memory under Node. This
 * replays one of them, by the same function, inside a lift: under SES, with
 * the engine loaded as a pattern loads it and its module data frozen. It
 * compares the stored state, the move record and every viewer's whole output
 * at every version.
 *
 * The recording is written from `conformance/` first:
 *
 *   node fabric/patterns/tools/recording-fixture.mjs memory/scenario-mismatch fabric/patterns/memory/testdata/recording.json
 */
import { assert, dataFile, lift, pattern, TESTS } from "commonfabric";

import { type Recording, replayRecording } from "../toolkit/parity.ts";
import { memoryGame } from "./game.ts";

const RECORDING = dataFile("./testdata/recording.json");

/** What a replay found, as text a test can compare. */
export const replay = lift((text: string): string => {
  const recording: Recording = JSON.parse(text);
  const result = replayRecording(memoryGame(), recording);
  return `${result.versions} versions, ${result.viewerVersions} viewer outputs, ${result.problems.length} problems${
    result.problems.length > 0 ? ": " + result.problems.join(" | ") : ""
  }`;
});

export default pattern(() => {
  const outcome = replay(RECORDING);
  const matches = assert(() =>
    outcome === "6 versions, 24 viewer outputs, 0 problems"
  );
  return { [TESTS]: [{ assertion: matches }], outcome };
});
