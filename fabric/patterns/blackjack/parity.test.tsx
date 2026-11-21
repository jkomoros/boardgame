/**
 * Blackjack is Go's blackjack inside the pattern sandbox too.
 *
 * `parity.test.ts` replays every recorded Go game of blackjack under Node.
 * This replays the opening of one of them, by the same function, inside a
 * lift: the deal, with its ten fix-up moves, and the first turns. It
 * compares the stored state, the move record and every viewer's whole
 * output at every version, each player's own view of their hole card
 * included.
 *
 * The recording is written from `conformance/` first:
 *
 *   node fabric/patterns/tools/recording-fixture.mjs blackjack/golden-714104226cb9a317 fabric/patterns/blackjack/testdata/recording.json 16
 */
import { assert, dataFile, lift, pattern, TESTS } from "commonfabric";

import { type Recording, replayRecording } from "../toolkit/parity.ts";
import { blackjackGame } from "./game.ts";

const RECORDING = dataFile("./testdata/recording.json");

/** What a replay found, as text a test can compare. */
export const replay = lift((text: string): string => {
  const recording: Recording = JSON.parse(text);
  const result = replayRecording(blackjackGame(), recording);
  return `${result.versions} versions, ${result.viewerVersions} viewer outputs, ${result.problems.length} problems${
    result.problems.length > 0 ? ": " + result.problems.join(" | ") : ""
  }`;
});

export default pattern(() => {
  const outcome = replay(RECORDING);
  const matches = assert(() =>
    outcome === "16 versions, 64 viewer outputs, 0 problems"
  );
  return { [TESTS]: [{ assertion: matches }], outcome };
});
