/**
 * Reads recorded Go games (`conformance/`) for the parity tests. Node only.
 */
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { gunzipSync } from "node:zlib";

import type { Recording } from "./parity.ts";
import type { TableState } from "./table.ts";

const CORPUS = resolve(
  dirname(fileURLToPath(import.meta.url)),
  "../../../conformance",
);

/** Every recording of `game` in the corpus. */
export function recordingsOf(game: string): Recording[] {
  const index = JSON.parse(
    readFileSync(resolve(CORPUS, "index.json"), "utf8"),
  ) as { file: string; game: string }[];
  return index.filter((entry) => entry.game === game).map((entry) =>
    JSON.parse(
      gunzipSync(readFileSync(resolve(CORPUS, entry.file))).toString("utf8"),
    ) as Recording
  );
}

function sorted(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(sorted);
  if (value !== null && typeof value === "object") {
    const record = value as Record<string, unknown>;
    const out: Record<string, unknown> = {};
    for (const key of Object.keys(record).sort()) out[key] = sorted(record[key]);
    return out;
  }
  return value;
}

/**
 * A table as a Common Fabric cell hands it back: a copy, with every
 * object's keys in sorted order and nothing shared with what was stored.
 */
export function asACellStoresIt(table: TableState): TableState {
  return sorted(JSON.parse(JSON.stringify(table))) as TableState;
}
