/**
 * A table: seats, a game's record, its state after each move of the last
 * input, and its move history, with the one function that advances it
 * (`stepTable`) and the one that says what a principal may see
 * (`viewForPrincipal`).
 *
 * The game itself is the TypeScript engine's
 * (`server/static/src/engine/rules`): a table holds exactly what that engine
 * asks a host to hold (the game record, the move history, the stored state)
 * and passes it back on every call. Go's shuffle, component ids, fix-up moves
 * and sanitization all come from there.
 *
 * Nothing here knows about Common Fabric. A game's pattern stores a
 * `TableState` in a sealed cell, calls `stepTable` from its one handler and
 * `viewForPrincipal` from its one released lift. Everything is plain data
 * and pure functions: no clock, no randomness and no identity but what the
 * caller passes in.
 */
import {
  type CompiledGame,
  createGame,
  gameDisplayName,
  type GameRecord,
  type Input,
  type MoveRecord,
  propose,
  type StepResult,
  variantConfig,
  viewerOutput,
  type ViewerOutput,
} from "../../../server/static/src/engine/rules/index.ts";
import type { StateJSON } from "../../../server/static/src/engine/state.ts";

import { type Json, type JsonObject, toJson, toJsonObject } from "./json.ts";
import type {
  MoveLegality,
  Proposal,
  SetupOption,
  TableInfo,
  TableView,
  TimerView,
  ViewEntry,
  ViewerPart,
} from "./view.ts";

/** The viewer who holds no seat. */
export const OBSERVER_SEAT = -1;

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

/** A seat and who holds it. */
export interface Seat {
  /** The DID of whoever holds the seat. */
  principal: string;
  label: string;
}

/** Everything a table knows. No person reads this. */
export interface TableState {
  phase: "seating" | "playing" | "finished";
  seats: Seat[];
  record: StoredRecord;
  /** Changes with every deal; a guest keys what it has shown by it. */
  epoch: number;
  /** How many proposals have been accepted, seating included. */
  revision: number;
  /** The head game version, or -1 before the deal. */
  version: number;
  /**
   * The engine's storage record of the state after each move of the last
   * input, in order: the move a proposal, a tick or the deal made, then each
   * fix-up move the engine applied after it. The last is the head, which
   * the next input is applied to; the others are what an animator plays on
   * the way there. Unsanitized.
   */
  states: JsonObject[];
  /**
   * The state the last input was applied to, which its first move is
   * sanitized against; `null` after a deal, which starts from nothing.
   */
  before: JsonObject | null;
  /** The table's clock when the last input was applied, in milliseconds. */
  at: number;
  /** The engine's move record of every version from 1, in order. */
  history: JsonObject[];
}

/** One version a table holds: one of `states`, with its place among them. */
export interface Held {
  version: number;
  /** The engine's storage record of the state, unsanitized. */
  state: JsonObject;
  /** The state before this version's move; absent for the deal's first. */
  before: JsonObject | undefined;
  /** True for the head; false for a version a fix-up replaced at once. */
  settled: boolean;
}

/** What a step needs that only the runtime can supply. */
export interface StepContext {
  /** The principal the runtime says is acting. Never from the payload. */
  proposer: string;
  /** The runtime's clock, in milliseconds. */
  now: number;
  /** A source of fresh randomness, drawn from only when dealing. */
  entropy: () => number;
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
    states: [],
    before: null,
    at: 0,
    history: [],
  };
}

/**
 * A plain copy of a stored table, with nothing shared with what was read;
 * the table before anyone sat down when nothing is stored yet.
 *
 * A handler steps the copy and sets the result, so every value it stores is
 * one it supplied, and none is a reference to what it found in place.
 */
export function tableOf(stored: TableState | null | undefined): TableState {
  if (stored === null || stored === undefined) return emptyTable();
  const copy: unknown = toJson(stored);
  return copy as TableState;
}

/** The versions a table holds, ascending. The last one is the head. */
export function held(table: TableState): Held[] {
  const states = Array.isArray(table.states) ? table.states : [];
  const first = table.version - (states.length - 1);
  return states.map((state, index) => ({
    version: first + index,
    state,
    before: index > 0 ? states[index - 1] : table.before ?? undefined,
    settled: index === states.length - 1,
  }));
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

/**
 * The table after an input: the versions it produced replace the last
 * input's, and the head they were applied to becomes `before`.
 */
function withVersions(
  table: TableState,
  result: StepResult,
  now: number,
): Partial<TableState> {
  const history = table.history.slice();
  for (const produced of result.versions) {
    if (produced.move !== null) history.push(toJsonObject(produced.move));
  }
  const head = result.versions[result.versions.length - 1];
  const previous = table.states[table.states.length - 1];
  return {
    phase: head.finished ? "finished" : "playing",
    version: head.version,
    states: result.versions.map((produced) => toJsonObject(produced.state)),
    before: previous ?? null,
    at: now,
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
      { ...table, states: [], before: null, history: [] },
      created,
      options.now,
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
): TableStep {
  const head = table.states[table.states.length - 1];
  if (head === undefined) return refused(table, "no game is being played");
  const result = propose(game, {
    record: engineRecord(table.record),
    history: engineHistory(table.history),
    state: engineState(head),
  }, input, { now });
  if (result.versions.length === 0) {
    return refused(table, result.error ?? "the input changed nothing");
  }
  // The engine returns versions and an error together only at the fix-up
  // cap, where Go has already committed them.
  return accepted(table, withVersions(table, result, now));
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
        states: [],
        before: null,
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
      }, context.now);
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
      );
    }
    default:
      return refused(table, "unknown kind of proposal");
  }
}

//
// What a viewer may see
//

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
  at: Held,
  viewer: number,
): ViewerOutput {
  return viewerOutput(game, {
    record: engineRecord(table.record),
    history: engineHistory(table.history.slice(0, at.version)),
    state: engineState(at.state),
    previousState: at.before === undefined ? undefined : engineState(at.before),
    settled: at.settled,
    // A lift has no clock. A timer's time left is as of when the version was
    // produced, which is what a guest counts down from.
    now: table.at,
  }, viewer);
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
  at: Held,
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
  viewer: number,
): Omit<TableView, "entries"> {
  const seats = Array.isArray(table.seats) ? table.seats : [];
  return {
    gameName: game.spec.name,
    phase: table.phase,
    epoch: table.epoch,
    revision: table.revision,
    version: table.version,
    viewer,
    // A seat's name, and not who holds it: a view tells its viewer which
    // seat is theirs, and nobody needs anyone's identity.
    seats: seats.map((seat) => ({ label: seat.label })),
    minPlayers: game.spec.minNumPlayers,
    maxPlayers: game.spec.maxNumPlayers,
  };
}

/**
 * One seat's view: who is seated, and each version the last input produced,
 * sanitized as the engine sanitizes it for `viewer`. `viewer` is a seat, or
 * `OBSERVER_SEAT` for whoever only watches.
 *
 * Every entry has the state and the move that made it, which is what an
 * animator plays. Only the head has what the viewer may do there: nobody
 * acts on a version a fix-up replaced at once.
 *
 * It is total, and every object and array in the result is new (`toJson`):
 * a part of the table returned as it is would reach the store as a
 * reference to the sealed table.
 */
export function seatViewOf(
  game: CompiledGame,
  table: TableState,
  viewer: number,
): TableView {
  const entries = held(table).map((at) => {
    const output = viewerOutputAt(game, table, at, viewer);
    return entryOf(
      at,
      output,
      at.settled ? { [String(viewer)]: viewerPart(game, output) } : {},
    );
  });
  return { ...frame(game, table, viewer), entries };
}

/** The seat `principal` holds at `table`, or `OBSERVER_SEAT`. */
export function seatOf(
  table: TableState,
  principal: string | undefined,
): number {
  if (typeof principal !== "string" || principal === "") return OBSERVER_SEAT;
  const seats = Array.isArray(table.seats) ? table.seats : [];
  const seat = seats.findIndex((held) => held.principal === principal);
  return seat < 0 ? OBSERVER_SEAT : seat;
}

/**
 * What `principal` may see of a stored table: their own seat's view if they
 * hold a seat, and a watcher's if they hold none or are nobody. Which seat
 * is theirs is a fact of the table, so sitting down and leaving change what
 * a principal sees with no label changing anywhere.
 *
 * This is the whole read side of a table. A game's pattern calls it from its
 * one released lift, with the principal the runtime says the lift is
 * computed for.
 */
export function viewForPrincipal(
  game: CompiledGame,
  stored: TableState | null | undefined,
  principal: string | undefined,
): TableView {
  const table = stored ?? emptyTable();
  return seatViewOf(game, table, seatOf(table, principal));
}

/** A game's variants, as what a lobby offers before the deal. */
function setupOptionsOf(game: CompiledGame): SetupOption[] {
  // The engine's own description, with every display name filled in as Go's
  // GameManager.Variants() fills them.
  const config = variantConfig(game.spec) as Record<string, {
    Name: string;
    DisplayName: string;
    Default: string;
    Values: Record<string, { Name: string; DisplayName: string }>;
  }>;
  return Object.keys(config).map((key): SetupOption => {
    const variant = config[key];
    return {
      name: variant.Name,
      displayName: variant.DisplayName,
      default: variant.Default,
      values: Object.keys(variant.Values).map((value) => ({
        name: variant.Values[value].Name,
        displayName: variant.Values[value].DisplayName,
      })),
    };
  });
}

/**
 * The public facts a renderer and a lobby need once, from the game's
 * definition.
 */
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
    displayName: gameDisplayName(game.spec),
    setupOptions: setupOptionsOf(game),
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
