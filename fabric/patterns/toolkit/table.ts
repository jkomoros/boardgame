/**
 * A table: seats, a game's record, its state at its most recent versions and
 * its move history, with the one function that advances it (`stepTable`) and
 * the ones that say what a viewer may see (`viewOf`, `seatViewOf`).
 *
 * The game itself is the TypeScript engine's
 * (`server/static/src/engine/rules`): a table holds exactly what that engine
 * asks a host to hold (the game record, the move history, the stored state)
 * and passes it back on every call. Go's shuffle, component ids, fix-up moves
 * and sanitization all come from there.
 *
 * Nothing here knows about Common Fabric. A game's pattern stores a
 * `TableState` in a sealed cell, calls `stepTable` from its one handler and a
 * view function from its released lift. Everything is plain data and pure
 * functions: no clock, no randomness and no identity but what the caller
 * passes in.
 */
import {
  type CompiledGame,
  createGame,
  type GameRecord,
  type Input,
  type MoveRecord,
  propose,
  type StepResult,
  viewerOutput,
  type ViewerOutput,
} from "../../../server/static/src/engine/rules/index.ts";
import type { StateJSON } from "../../../server/static/src/engine/state.ts";

import { type Json, type JsonObject, toJson, toJsonObject } from "./json.ts";
import type {
  MoveLegality,
  Proposal,
  SeatView,
  TableInfo,
  TableView,
  TimerView,
  ViewEntry,
  ViewerPart,
} from "./view.ts";

/** The viewer who holds no seat. */
export const OBSERVER_SEAT = -1;

/**
 * How many inputs' worth of versions a table holds. An input is one
 * proposal, tick or deal; it produces one version and then one for each
 * fix-up move, however many. A view shows every held version.
 *
 * One, because the runtime's cost for a move grows with the size of what is
 * stored and released, and a Go-faithful state is large (Memory's is 19 KB
 * a version). With one input held, a guest that receives every value of the
 * view animates every step; one that misses a value jumps to the newer
 * state without animating what it missed. With two held it would not jump,
 * and in Memory a move took 5.9 s to arrive where it now takes 3.7 s.
 */
export const INPUTS_HELD = 1;

/** The game record the engine was created with; `id` is empty before the deal. */
export interface StoredRecord {
  name: string;
  id: string;
  /** The game's only secret besides its state. */
  secretSalt: string;
  numPlayers: number;
  variant: Record<string, string>;
  agents: string[];
  seating: boolean;
}

/** One version the table still holds. */
export interface Recorded {
  version: number;
  /** The engine's storage record of the state, unsanitized. */
  state: JsonObject;
  /** True for the last version an input produced. */
  settled: boolean;
  finished: boolean;
  winners: number[];
  /** The table's clock when the version was produced, in milliseconds. */
  at: number;
}

/** Everything a table knows. No person reads this. */
export interface TableState {
  phase: "seating" | "playing" | "finished";
  seats: SeatView[];
  record: StoredRecord;
  /** Changes with every deal; a guest keys what it has shown by it. */
  epoch: number;
  /** How many proposals have been accepted, seating included. */
  revision: number;
  /** The head game version, or -1 before the deal. */
  version: number;
  /**
   * The versions still held, by version (`"v12"`): those of the last
   * `INPUTS_HELD` inputs, and the one before them, since a version's move is
   * judged against the state before it.
   *
   * A record, not a list, so that a version keeps its place while it is
   * held: a store that compares what it is given with what it has then
   * writes only the versions that are new.
   */
  recent: Record<string, Recorded>;
  /** The engine's move record of every version from 1, in order. */
  history: JsonObject[];
}

/** What a step needs that only the runtime can supply. */
export interface StepContext {
  /** The principal the runtime says is acting. Never from the payload. */
  proposer: string;
  /** The runtime's clock, in milliseconds. */
  now: number;
  /** A source of fresh randomness, drawn from only when dealing. */
  entropy: () => number;
  /** How many inputs' versions to hold; `INPUTS_HELD` unless a test wants more. */
  inputsHeld?: number;
}

export interface TableStep {
  /** Why the proposal was refused; empty when it was applied. */
  error: string;
  /** The table after the proposal; the same table when it was refused. */
  table: TableState;
}

function noRecord(): StoredRecord {
  return {
    name: "",
    id: "",
    secretSalt: "",
    numPlayers: 0,
    variant: {},
    agents: [],
    seating: false,
  };
}

/** The table before anyone sat down. */
export function emptyTable(): TableState {
  return {
    phase: "seating",
    seats: [],
    record: noRecord(),
    epoch: 0,
    revision: 0,
    version: -1,
    recent: {},
    history: [],
  };
}

/** The versions a table holds, ascending. The last one is the head. */
export function held(table: TableState): Recorded[] {
  const recent = table.recent ?? {};
  return Object.keys(recent).map((key) => recent[key]).sort((a, b) =>
    a.version - b.version
  );
}

function refused(table: TableState, error: string): TableStep {
  return { error, table };
}

function accepted(table: TableState, change: Partial<TableState>): TableStep {
  return {
    error: "",
    table: { ...table, ...change, revision: table.revision + 1 },
  };
}

const HEX = "0123456789abcdef";

/** `length` hex digits drawn from `entropy`. */
function randomHex(entropy: () => number, length: number): string {
  let out = "";
  for (let i = 0; i < length; i++) {
    out += HEX[Math.floor(entropy() * 16) % 16];
  }
  return out;
}

function cleanLabel(label: string | undefined, seat: number): string {
  const text = typeof label === "string" ? label.trim().slice(0, 24) : "";
  return text === "" ? `Player ${seat + 1}` : text;
}

//
// The engine's side of the boundary
//

function engineRecord(record: StoredRecord): GameRecord {
  return {
    name: record.name,
    id: record.id,
    secretSalt: record.secretSalt,
    numPlayers: record.numPlayers,
    variant: Object.keys(record.variant).length === 0
      ? null
      : { ...record.variant },
    agents: record.agents.slice(),
    seating: record.seating,
  };
}

function storedRecord(record: GameRecord): StoredRecord {
  return {
    name: record.name,
    id: record.id,
    secretSalt: record.secretSalt,
    numPlayers: record.numPlayers,
    variant: { ...(record.variant ?? {}) },
    agents: record.agents.slice(),
    seating: record.seating,
  };
}

/** A stored state as the engine takes one. The engine copies what it loads. */
function engineState(state: JsonObject): StateJSON {
  const value: unknown = state;
  return value as StateJSON;
}

function engineHistory(history: readonly JsonObject[]): MoveRecord[] {
  const value: unknown = history;
  return value as MoveRecord[];
}

/** The versions a step produced, added to the table. */
function withVersions(
  table: TableState,
  result: StepResult,
  now: number,
  inputsHeld: number,
): Partial<TableState> {
  const all = held(table);
  const history = table.history.slice();
  result.versions.forEach((produced, index) => {
    if (produced.move !== null) history.push(toJsonObject(produced.move));
    all.push({
      version: produced.version,
      state: toJsonObject(produced.state),
      settled: index === result.versions.length - 1,
      finished: produced.finished,
      winners: produced.winners.slice(),
      at: now,
    });
  });
  // Keep back to the settled version that precedes the last `inputsHeld`
  // inputs: each input ends in exactly one settled version.
  let first = 0;
  let settledSeen = 0;
  for (let i = all.length - 1; i >= 0; i--) {
    if (!all[i].settled) continue;
    settledSeen += 1;
    if (settledSeen > inputsHeld) {
      first = i;
      break;
    }
  }
  const recent: Record<string, Recorded> = {};
  for (const kept of all.slice(first)) recent[`v${kept.version}`] = kept;
  const head = all[all.length - 1];
  return {
    phase: head.finished ? "finished" : "playing",
    version: head.version,
    recent,
    history,
  };
}

/**
 * Deals: creates the game with the engine. `id` and `secretSalt` are the
 * game's identity; everything random about it derives from them.
 *
 * `stepTable` calls this with a freshly drawn identity. A test that replays
 * a recording calls it with the recorded one.
 */
export function dealTable(
  game: CompiledGame,
  table: TableState,
  identity: { id: string; secretSalt: string },
  options: {
    numPlayers: number;
    variant: Record<string, string> | null;
    seating: boolean;
    now: number;
    inputsHeld?: number;
  },
): TableStep {
  const created = createGame(game, {
    id: identity.id,
    secretSalt: identity.secretSalt,
    numPlayers: options.numPlayers,
    variant: options.variant,
    seating: options.seating,
    now: options.now,
  });
  if (created.record === null || created.versions.length === 0) {
    return refused(table, created.error ?? "the game could not be created");
  }
  return accepted(table, {
    record: storedRecord(created.record),
    // A deal a client speculated and the deal the server made have different
    // ids, so a guest never mixes the two.
    epoch: parseInt(identity.id.slice(0, 8), 16) || 1,
    ...withVersions(
      { ...table, recent: {}, history: [] },
      created,
      options.now,
      options.inputsHeld ?? INPUTS_HELD,
    ),
  });
}

/**
 * Hands the engine one input for a dealt table and records what it
 * produced. This is the only place a move is applied or a timer fires.
 */
export function advanceTable(
  game: CompiledGame,
  table: TableState,
  input: Input,
  now: number,
  inputsHeld: number = INPUTS_HELD,
): TableStep {
  const all = held(table);
  const head = all[all.length - 1];
  if (head === undefined) return refused(table, "no game is being played");
  const result = propose(game, {
    record: engineRecord(table.record),
    history: engineHistory(table.history),
    state: engineState(head.state),
  }, input, { now });
  if (result.versions.length === 0) {
    return refused(table, result.error ?? "the input changed nothing");
  }
  // The engine returns versions and an error together only at the fix-up
  // cap, where Go has already committed them.
  return accepted(table, withVersions(table, result, now, inputsHeld));
}

/** The engine's schema for a move's input fields, as far as this needs it. */
interface InputFieldSchema {
  name: string;
  wireType: string;
  disposition: string;
}

function inputFields(game: CompiledGame, moveName: string): InputFieldSchema[] {
  const canonical = game.canonicalMoveNames[moveName.toLowerCase()] ??
    moveName;
  for (const item of game.moveInputSchema) {
    const schema = item as { name?: string; fields?: InputFieldSchema[] };
    if (schema.name === canonical) return schema.fields ?? [];
  }
  return [];
}

/**
 * A move's arguments as the engine binds them. A browser sends every
 * argument as text, the way the Go API's move forms do; the engine wants
 * each in its own type and refuses anything else, in Go's words.
 */
export function boundInput(
  game: CompiledGame,
  moveName: string,
  args: Readonly<Record<string, string>>,
): Record<string, unknown> {
  const fields = inputFields(game, moveName);
  const input: Record<string, unknown> = {};
  for (const name of Object.keys(args)) {
    const text = args[name];
    const field = fields.find((candidate) => candidate.name === name);
    const type = field?.wireType ?? "string";
    if (typeof text !== "string") {
      input[name] = text;
    } else if (type === "int" || type === "playerIndex") {
      input[name] = /^-?\d+$/.test(text) ? Number(text) : text;
    } else if (type === "bool") {
      input[name] = text === "true" || text === "1" || text === "on";
    } else {
      input[name] = text;
    }
  }
  return input;
}

/**
 * The table after `proposal`, made by `context.proposer` at `context.now`.
 * A refused proposal returns the table unchanged with the reason.
 *
 * This is the whole write side of a table. It is the only place a seat is
 * taken, cards are dealt, a move is proposed or a timer is checked.
 */
export function stepTable(
  game: CompiledGame,
  table: TableState,
  proposal: Proposal,
  context: StepContext,
): TableStep {
  const proposer = context.proposer;
  if (typeof proposer !== "string" || proposer === "") {
    return refused(table, "no one is acting");
  }
  const seat = table.seats.findIndex((held) => held.principal === proposer);
  const inputsHeld = context.inputsHeld ?? INPUTS_HELD;
  switch (proposal?.kind) {
    case "sit": {
      if (table.phase !== "seating") {
        return refused(table, "the game has started");
      }
      if (seat >= 0) return refused(table, "you already hold a seat");
      if (table.seats.length >= game.spec.maxNumPlayers) {
        return refused(table, "every seat is taken");
      }
      return accepted(table, {
        seats: [...table.seats, {
          principal: proposer,
          label: cleanLabel(proposal.label, table.seats.length),
        }],
      });
    }
    case "leave": {
      if (table.phase !== "seating") {
        return refused(table, "the game has started");
      }
      if (seat < 0) return refused(table, "you hold no seat");
      return accepted(table, {
        seats: table.seats.filter((_, index) => index !== seat),
      });
    }
    case "start": {
      if (table.phase !== "seating") {
        return refused(table, "the game has started");
      }
      if (seat < 0) return refused(table, "only a seated player may deal");
      if (table.seats.length < game.spec.minNumPlayers) {
        return refused(table, "not enough players are seated");
      }
      // The one draw of fresh randomness. Everything else about the game
      // derives from these two values.
      const id = randomHex(context.entropy, 16);
      const secretSalt = randomHex(context.entropy, 32);
      // Every seat here is a person already, so the engine seats nobody.
      return dealTable(game, table, { id, secretSalt }, {
        numPlayers: table.seats.length,
        variant: proposal.options ?? null,
        seating: false,
        now: context.now,
        inputsHeld,
      });
    }
    case "restart": {
      if (table.phase !== "finished") {
        return refused(table, "the game is not over");
      }
      if (seat < 0) return refused(table, "only a seated player may reset");
      return accepted(table, {
        phase: "seating",
        record: noRecord(),
        version: -1,
        recent: {},
        history: [],
      });
    }
    case "play": {
      if (table.phase !== "playing") {
        return refused(table, "no game is being played");
      }
      if (seat < 0) return refused(table, "you hold no seat");
      const name = proposal.name;
      if (typeof name !== "string" || name === "") {
        return refused(table, "the move has no name");
      }
      return advanceTable(game, table, {
        kind: "propose",
        name,
        input: boundInput(game, name, proposal.arguments ?? {}),
        proposer: seat,
      }, context.now, inputsHeld);
    }
    case "tick": {
      if (table.phase !== "playing") {
        return refused(table, "no game is being played");
      }
      if (typeof proposal.timerId !== "string" || proposal.timerId === "") {
        return refused(table, "the tick names no timer");
      }
      // A tick is a request to look at the clock, not a claim about it: the
      // engine fires the timer only if its deadline is at or before `now`.
      return advanceTable(
        game,
        table,
        { kind: "timer", id: proposal.timerId },
        context.now,
        inputsHeld,
      );
    }
    default:
      return refused(table, "unknown kind of proposal");
  }
}

//
// What a viewer may see
//

/** A held version and, when it is held too, the one before it. */
interface Shown {
  at: Recorded;
  before: Recorded | undefined;
}

/** The versions a view shows: every one whose predecessor is still held. */
function shown(table: TableState): Shown[] {
  const recent = held(table);
  const out: Shown[] = [];
  recent.forEach((at, index) => {
    const before = recent[index - 1];
    if (before !== undefined) out.push({ at, before });
    else if (at.version === 0) out.push({ at, before: undefined });
  });
  return out;
}

/**
 * The engine's whole output for one viewer at one held version: the game
 * object with its sanitized state, the move as that viewer may see it,
 * legality and projected choices. `viewer` is a seat, or `OBSERVER_SEAT`.
 *
 * Every view below is made from this. A parity test compares it with what
 * the Go server recorded.
 */
export function viewerOutputAt(
  game: CompiledGame,
  table: TableState,
  at: Recorded,
  before: Recorded | undefined,
  viewer: number,
): ViewerOutput {
  return viewerOutput(game, {
    record: engineRecord(table.record),
    history: engineHistory(table.history.slice(0, at.version)),
    state: engineState(at.state),
    previousState: before === undefined ? undefined : engineState(before.state),
    settled: at.settled,
    // A lift has no clock. A timer's time left is as of when the version was
    // produced, which is what a guest counts down from.
    now: at.at,
  }, viewer);
}

/** Every held version's engine output for one viewer, ascending. */
export function viewerOutputs(
  game: CompiledGame,
  table: TableState,
  viewer: number,
): { version: number; output: ViewerOutput }[] {
  return shown(table).map(({ at, before }) => ({
    version: at.version,
    output: viewerOutputAt(game, table, at, before, viewer),
  }));
}

/** Moves a player may propose that take no argument from them. */
function zeroInputMoves(game: CompiledGame, names: string[]): string[] {
  return names.filter((name) =>
    !inputFields(game, name).some((field) => field.disposition === "required")
  );
}

/** What of a version depends on who is looking, in the guest's shape. */
function viewerPart(game: CompiledGame, output: ViewerOutput): ViewerPart {
  const moveLegality: Record<string, MoveLegality> = {};
  for (const item of output.legality) {
    const legality: MoveLegality = {
      legalForPlayer: item.legalForPlayer,
      legalForAnyone: item.legalForAnyone,
    };
    if (item.legalForPlayerError !== undefined) {
      legality.error = item.legalForPlayerError;
    }
    moveLegality[item.name] = legality;
  }
  const choices = output.projectedMoveChoices;
  return {
    moveLegality,
    zeroInputMoves: zeroInputMoves(game, Object.keys(moveLegality)),
    projectedMoveChoices: {
      StateVersion: output.game.Version,
      Status: choices?.Status ?? "ready",
      MoveChoiceProjectionSchemaFingerprint:
        choices?.MoveChoiceProjectionSchemaFingerprint ??
          game.choiceSchemaFingerprint,
      ProjectionSchemaVersion: choices?.ProjectionSchemaVersion ?? 1,
      Sets: (choices?.Sets ?? []).map((set) => ({
        MoveName: set.MoveName,
        FieldName: set.FieldName,
        Source: set.Source,
        Candidates: set.Candidates.map((candidate) => ({
          Value: toJson(candidate.Value),
          Available: candidate.Available,
        })),
      })),
    },
  };
}

/** One version in the guest's shape, with the state `stateOf` saw. */
function entryOf(
  at: Recorded,
  stateOf: ViewerOutput,
  viewers: Record<string, ViewerPart>,
): ViewEntry {
  const timers: Record<string, TimerView> = {};
  for (const id of Object.keys(stateOf.game.ActiveTimers)) {
    timers[id] = { TimeLeft: stateOf.game.ActiveTimers[id].TimeLeft };
  }
  const entry: ViewEntry = {
    version: at.version,
    state: toJsonObject(stateOf.game.CurrentState),
    currentPlayerIndex: stateOf.game.CurrentPlayerIndex,
    timers,
    outcome: {
      finished: stateOf.game.Finished,
      winners: stateOf.game.Winners.slice(),
    },
    settled: at.settled,
    viewers,
  };
  if (stateOf.move !== null) {
    entry.move = {
      AnimationKey: stateOf.move.AnimationKey,
      Version: stateOf.move.Version,
    };
    if (stateOf.move.Properties !== undefined) {
      entry.move.Properties = toJsonObject(stateOf.move.Properties);
    }
  }
  return entry;
}

function frame(
  game: CompiledGame,
  table: TableState,
): Omit<TableView, "entries"> {
  const seats = Array.isArray(table.seats) ? table.seats : [];
  return {
    gameName: game.spec.name,
    phase: table.phase,
    epoch: table.epoch,
    revision: table.revision,
    version: table.version,
    seats: seats.map((seat) => ({
      principal: seat.principal,
      label: seat.label,
    })),
    minPlayers: game.spec.minNumPlayers,
    maxPlayers: game.spec.maxNumPlayers,
  };
}

/**
 * One view for everyone at the table: who is seated, and every version the
 * table still holds, sanitized as an observer sees it, with each seat's
 * legality and choices beside it.
 *
 * For a game whose seats all see the same state, such as Memory. A game
 * with hands must not use it: an observer's state is what is shown.
 *
 * It is total, and every object and array in the result is new (`toJson`):
 * a part of the table returned as it is would reach the store as a
 * reference to the sealed table and withhold the whole view.
 */
export function viewOf(game: CompiledGame, table: TableState): TableView {
  const seats = Array.isArray(table.seats) ? table.seats : [];
  const entries = shown(table).map(({ at, before }) => {
    const observer = viewerOutputAt(game, table, at, before, OBSERVER_SEAT);
    const viewers: Record<string, ViewerPart> = {
      [String(OBSERVER_SEAT)]: viewerPart(game, observer),
    };
    for (let seat = 0; seat < seats.length; seat++) {
      viewers[String(seat)] = viewerPart(
        game,
        viewerOutputAt(game, table, at, before, seat),
      );
    }
    return entryOf(at, observer, viewers);
  });
  return { ...frame(game, table), entries };
}

/**
 * One viewer's own view: the same shape as `viewOf`, with every version
 * sanitized for `viewer` and only that viewer's legality in it. `viewer` is
 * a seat, or `OBSERVER_SEAT`.
 *
 * For a game with per-seat secrets, such as blackjack: each seat's view is
 * a separate value, to be released to that seat's audience alone.
 */
export function seatViewOf(
  game: CompiledGame,
  table: TableState,
  viewer: number,
): TableView {
  const entries = shown(table).map(({ at, before }) => {
    const output = viewerOutputAt(game, table, at, before, viewer);
    return entryOf(at, output, { [String(viewer)]: viewerPart(game, output) });
  });
  return { ...frame(game, table), entries };
}

/** The public facts a renderer needs once, from the game's definition. */
export function infoOf(game: CompiledGame, seatColors: string[]): TableInfo {
  const decks: JsonObject = {};
  for (const deckName of Object.keys(game.spec.decks)) {
    decks[deckName] = game.spec.decks[deckName].components.map((
      values,
      index,
    ): Json => ({ Index: index, Values: toJson(values) }));
  }
  const enums: JsonObject = {};
  for (const bound of game.enums.chestEnums()) {
    enums[bound.name] = toJson(bound.def);
  }
  return {
    gameName: game.spec.name,
    chest: {
      Decks: decks,
      Enums: enums,
      Constants: toJson(game.spec.constants ?? null),
      LegalTemplates: toJson(game.templates),
    },
    moveInputSchemaFingerprint: game.moveInputSchemaFingerprint,
    seatColors: seatColors.slice(),
  };
}
