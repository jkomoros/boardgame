/**
 * Replays a recorded Go game through a table and compares, at every version,
 * the stored state, the move record and every viewer's whole output with
 * what the Go server recorded.
 *
 * The recording is a conformance document (`conformance/README.md`). The
 * replay goes through the same functions a pattern's `move` handler and its
 * view lift call: `stepTable` for a player's proposal, `advanceTable` for
 * what only the table itself may do, and `viewerOutputAt` for each view.
 *
 * It uses no platform API, so it runs under Node and inside a pattern.
 */
import {
  type CompiledGame,
  parseInstant,
} from "../../../server/static/src/engine/rules/index.ts";

import type { Json, JsonObject } from "./json.ts";
import {
  advanceTable,
  dealTable,
  emptyTable,
  held,
  stepTable,
  type TableState,
  type TableStep,
  viewerOutputAt,
} from "./table.ts";

/** The parts of a conformance document a replay reads. */
export interface Recording {
  name: string;
  game: {
    id: string;
    secretSalt: string;
    numPlayers: number;
    variant: Record<string, string>;
    seating: boolean;
  };
  engine: { clock: string };
  versions: RecordedVersion[];
}

export interface RecordedVersion {
  version: number;
  settled: boolean;
  finished: boolean;
  winners: number[] | null;
  move: null | {
    name: string;
    version: number;
    initiator: number;
    phase: number;
    proposer: number;
    arguments: JsonObject;
    trigger: string;
  };
  state: Json;
  views: {
    viewer: number;
    game: Json;
    move: Json;
    legality: Json;
    projectedMoveChoices?: Json;
  }[];
}

export interface ParityResult {
  name: string;
  /** Versions compared. */
  versions: number;
  /** Viewer outputs compared: one for each viewer of each version. */
  viewerVersions: number;
  /** Individual comparisons made. */
  comparisons: number;
  /** The first few differences; empty when the replay matched. */
  problems: string[];
}

function show(value: unknown): string {
  const text = JSON.stringify(value);
  return text === undefined
    ? "undefined"
    : text.length > 120
    ? text.slice(0, 120) + "…"
    : text;
}

/**
 * Where `got` first differs from `want`, or `""`. Structural: the order of
 * an object's keys does not matter, and an absent key equals `undefined`.
 */
export function difference(got: unknown, want: unknown, path = ""): string {
  if (got === want) return "";
  if (
    got === null || want === null || typeof got !== "object" ||
    typeof want !== "object"
  ) {
    return `${path || "(root)"}: got ${show(got)}, want ${show(want)}`;
  }
  if (Array.isArray(got) !== Array.isArray(want)) {
    return `${path || "(root)"}: got ${show(got)}, want ${show(want)}`;
  }
  if (Array.isArray(got) && Array.isArray(want)) {
    if (got.length !== want.length) {
      return `${path || "(root)"}: length ${got.length}, want ${want.length}`;
    }
    for (let i = 0; i < got.length; i++) {
      const found = difference(got[i], want[i], `${path}[${i}]`);
      if (found !== "") return found;
    }
    return "";
  }
  const g = got as Record<string, unknown>;
  const w = want as Record<string, unknown>;
  const keys: string[] = [];
  for (const key of [...Object.keys(g), ...Object.keys(w)]) {
    if (!keys.includes(key)) keys.push(key);
  }
  for (const key of keys.sort()) {
    if (g[key] === undefined && w[key] === undefined) continue;
    if (g[key] === undefined) {
      return `${path}.${key}: missing, want ${show(w[key])}`;
    }
    if (w[key] === undefined) {
      return `${path}.${key}: unexpected ${show(g[key])}`;
    }
    const found = difference(g[key], w[key], `${path}.${key}`);
    if (found !== "") return found;
  }
  return "";
}

/**
 * The one normalization the corpus needs and does not declare: the exporter
 * writes an unsettled version's winners as `null` (see
 * `fabric/status/engine-rules.md`).
 */
function winnersOrEmpty(value: Json | undefined): Json {
  return value === null || value === undefined ? [] : value;
}

function recordedGame(game: Json): Json {
  if (game === null || typeof game !== "object" || Array.isArray(game)) {
    return game;
  }
  return { ...game, Winners: winnersOrEmpty(game["Winners"]) };
}

/** The principal the replay seats at `seat`. */
export function replayPrincipal(seat: number): string {
  return `did:replay:seat-${seat}`;
}

/**
 * Replays `recording` and compares every version. `stored` is applied to
 * the table after each step, to stand in for whatever keeps it between
 * runs; by default the table is passed on as it is.
 */
export function replayRecording(
  game: CompiledGame,
  recording: Recording,
  stored: (table: TableState) => TableState = (table) => table,
): ParityResult {
  const result: ParityResult = {
    name: recording.name,
    versions: 0,
    viewerVersions: 0,
    comparisons: 0,
    problems: [],
  };
  const check = (where: string, got: unknown, want: unknown): void => {
    result.comparisons += 1;
    const found = difference(got, want);
    if (found !== "" && result.problems.length < 5) {
      result.problems.push(`${where}: ${found}`);
    }
  };
  const fail = (message: string): ParityResult => {
    result.problems.push(message);
    return result;
  };

  const now = parseInstant(recording.engine.clock);
  if (now === null) return fail(`unreadable clock ${recording.engine.clock}`);
  // Hold every version, so each is compared in turn.
  const inputsHeld = recording.versions.length + 1;

  let table = emptyTable();
  for (let seat = 0; seat < recording.game.numPlayers; seat++) {
    table.seats.push({
      principal: replayPrincipal(seat),
      label: `Player ${seat + 1}`,
    });
  }

  let compared = -1;
  /** Compares every version the table holds that has not been compared. */
  const compareNew = (): void => {
    const all = held(table);
    all.forEach((at, index) => {
      if (at.version <= compared) return;
      compared = at.version;
      const want = recording.versions[at.version];
      const where = `v${at.version}`;
      if (want === undefined) {
        result.problems.push(`${where}: past the end of the recording`);
        return;
      }
      result.versions += 1;
      check(`${where} state`, at.state, want.state);
      const record = at.version === 0 ? null : table.history[at.version - 1];
      check(
        `${where} move`,
        record === null || record === undefined ? null : {
          name: record["Name"],
          version: record["Version"],
          initiator: record["Initiator"],
          phase: record["Phase"],
          proposer: record["Proposer"],
          arguments: record["Blob"],
        },
        want.move === null ? null : {
          name: want.move.name,
          version: want.move.version,
          initiator: want.move.initiator,
          phase: want.move.phase,
          proposer: want.move.proposer,
          arguments: want.move.arguments,
        },
      );
      check(`${where} status`, {
        settled: at.settled,
        finished: at.finished,
        winners: at.finished ? at.winners : [],
      }, {
        settled: want.settled,
        finished: want.finished,
        winners: winnersOrEmpty(want.winners),
      });
      const before = all[index - 1];
      for (const recorded of want.views) {
        const output = viewerOutputAt(game, table, at, before, recorded.viewer);
        const who = `${where} viewer ${recorded.viewer}`;
        result.viewerVersions += 1;
        check(`${who} game`, output.game, recordedGame(recorded.game));
        check(`${who} move`, output.move, recorded.move);
        check(`${who} legality`, output.legality, recorded.legality);
        check(
          `${who} choices`,
          output.projectedMoveChoices,
          recorded.projectedMoveChoices,
        );
      }
    });
  };

  const dealt = dealTable(game, table, {
    id: recording.game.id,
    secretSalt: recording.game.secretSalt,
  }, {
    numPlayers: recording.game.numPlayers,
    variant: recording.game.variant,
    seating: recording.game.seating,
    now,
    inputsHeld,
  });
  if (dealt.error !== "") return fail(`the deal was refused: ${dealt.error}`);
  table = stored(dealt.table);
  compareNew();

  while (table.version + 1 < recording.versions.length) {
    const next = recording.versions[table.version + 1];
    const move = next.move;
    if (move === null) return fail(`v${next.version} records no move`);
    let step: TableStep;
    if (move.trigger === "propose" && move.proposer >= 0) {
      // A player's proposal, as a browser sends one: by name, with only the
      // fields a proposer supplies, each as text.
      const supplied: Record<string, string> = {};
      const compiled = game.movesByName[move.name];
      for (const field of compiled?.inputFields ?? []) {
        if (
          field.disposition !== "required" &&
          field.disposition !== "server-defaulted"
        ) continue;
        const value = move.arguments[field.name];
        if (value !== undefined) supplied[field.name] = String(value);
      }
      step = stepTable(game, table, {
        kind: "play",
        name: move.name,
        arguments: supplied,
      }, {
        proposer: replayPrincipal(move.proposer),
        now,
        entropy: () => 0,
        inputsHeld,
      });
    } else if (move.trigger === "propose") {
      // The admin's proposal: no seat makes one, so it goes to the engine.
      step = advanceTable(game, table, {
        kind: "propose",
        name: move.name,
        input: {},
        proposer: move.proposer,
      }, now, inputsHeld);
    } else if (move.trigger === "seat") {
      step = advanceTable(game, table, {
        kind: "seat",
        player: Number(move.arguments["TargetPlayerIndex"]),
      }, now, inputsHeld);
    } else if (move.trigger === "timer") {
      // The recording's clock never advances, so its timers are forced.
      step = advanceTable(
        game,
        table,
        { kind: "timer", force: true },
        now,
        inputsHeld,
      );
    } else {
      return fail(
        `v${next.version} (${move.name}, ${move.trigger}) should have followed v${table.version}, but the table stopped there`,
      );
    }
    if (step.error !== "") {
      return fail(
        `v${next.version} (${move.name}): the table refused the recorded ${move.trigger}: ${step.error}`,
      );
    }
    table = stored(step.table);
    compareNew();
  }
  return result;
}
