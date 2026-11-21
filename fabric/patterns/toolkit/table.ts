/**
 * A table: seats, a game's state at its most recent versions, the move
 * history, and running timers, with the one function that advances it
 * (`stepTable`) and the one that says what its readers may see (`viewOf`).
 *
 * Nothing here knows about Common Fabric. A game's pattern stores a
 * `TableState` in a sealed cell, calls `stepTable` from its one handler and
 * `viewOf` from its one released lift. Everything is plain data and pure
 * functions: no clock, no randomness and no identity but what the caller
 * passes in.
 */
import type {
  MoveRecord,
  Outcome,
  Proposal,
  SeatView,
  TableView,
  TimerView,
  ViewEntry,
  ViewerPart,
} from "./view.ts";

/** The seat a timer's move is proposed from, as in the Go core. */
export const ADMIN_SEAT = -2;
/** The viewer who holds no seat. */
export const OBSERVER_SEAT = -1;

/** How many versions a table keeps, and so how many a view can carry. */
export const WINDOW = 6;

/** The only secrets of a game: ids and every draw derive from them. */
export interface GameIdentity {
  gameId: string;
  secretSalt: string;
}

/** A timer a move asks the table to start. */
export interface TimerRequest {
  id: string;
  durationMs: number;
  /** The move the table proposes, as admin, once the deadline passes. */
  moveName: string;
}

/** One version a step produced. */
export interface SteppedVersion<G> {
  game: G;
  /** The name of the move that produced it. */
  moveName: string;
  startTimer?: TimerRequest;
  /** Ids of timers this version cancelled. */
  cancelTimers: string[];
}

/** What a game's rules say to a proposed move. */
export interface StepOutcome<G> {
  /** Why the move was refused; empty when it was applied. */
  error: string;
  /** The proposed move's version, then one for each fix-up move after it. */
  versions: SteppedVersion<G>[];
}

/**
 * A game, as a table needs it. `G` is the game's whole state, unsanitized;
 * `S` is the state a viewer is shown.
 *
 * This is the seam the TypeScript engine's rules layer replaces: an engine
 * definition would supply all of these from one `GameDefinition`.
 */
export interface GameRules<G, S> {
  name: string;
  minPlayers: number;
  maxPlayers: number;
  /** The state at version 0. Every draw must derive from `identity`. */
  setUp(
    identity: GameIdentity,
    numPlayers: number,
    options: Record<string, string>,
  ): G;
  /**
   * Applies `moveName` as `seat` to `game`, whose version is `version`, and
   * then every fix-up move that follows. Returns an error and no versions
   * when the move is not legal for that seat.
   */
  step(
    game: G,
    version: number,
    identity: GameIdentity,
    moveName: string,
    args: Record<string, string>,
    seat: number,
  ): StepOutcome<G>;
  currentPlayer(game: G): number;
  outcome(game: G): Outcome;
  /** `game` as `viewer` may see it. Must be total. */
  sanitize(game: G, viewer: number): S;
  /** What `viewer` may do at `game`. Must be total. */
  viewerPart(game: G, version: number, viewer: number): ViewerPart;
}

/** One version the table still holds. */
export interface Recorded<G> {
  version: number;
  game: G;
  /** Empty for version 0. */
  moveName: string;
  timers: Record<string, TimerView>;
  settled: boolean;
}

/** One accepted move of the game, for the history. */
export interface HistoryEntry {
  version: number;
  name: string;
  /** The seat that proposed it; `ADMIN_SEAT` for a timer or a fix-up. */
  seat: number;
  arguments: Record<string, string>;
  /** The table's clock when it was applied, in milliseconds. */
  at: number;
}

export interface ActiveTimer {
  id: string;
  /** The table's clock at which the timer's move may be proposed. */
  deadline: number;
  moveName: string;
}

/** Everything a table knows. No person reads this. */
export interface TableState<G> {
  phase: "seating" | "playing" | "finished";
  seats: SeatView[];
  gameId: string;
  secretSalt: string;
  epoch: number;
  /** How many proposals have been accepted, seating included. */
  revision: number;
  /** The head game version, or -1 before the deal. */
  version: number;
  /** The last `WINDOW` versions, ascending; the last one is the head. */
  recent: Recorded<G>[];
  history: HistoryEntry[];
  timers: ActiveTimer[];
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

export interface TableStep<G> {
  /** Why the proposal was refused; empty when it was applied. */
  error: string;
  /** The table after the proposal; the same table when it was refused. */
  table: TableState<G>;
}

/** The table before anyone sat down. */
export function emptyTable<G>(): TableState<G> {
  return {
    phase: "seating",
    seats: [],
    gameId: "",
    secretSalt: "",
    epoch: 0,
    revision: 0,
    version: -1,
    recent: [],
    history: [],
    timers: [],
  };
}

function refused<G>(table: TableState<G>, error: string): TableStep<G> {
  return { error, table };
}

function accepted<G>(
  table: TableState<G>,
  change: Partial<TableState<G>>,
): TableStep<G> {
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

/** Applies the versions a game step produced to the table. */
function advance<G, S>(
  rules: GameRules<G, S>,
  table: TableState<G>,
  outcome: StepOutcome<G>,
  seat: number,
  args: Record<string, string>,
  now: number,
): TableStep<G> {
  let version = table.version;
  let timers = table.timers;
  const recent = table.recent.slice();
  const history = table.history.slice();
  outcome.versions.forEach((stepped, index) => {
    version += 1;
    if (stepped.cancelTimers.length > 0) {
      timers = timers.filter((timer) =>
        !stepped.cancelTimers.includes(timer.id)
      );
    }
    const shown: Record<string, TimerView> = {};
    const start = stepped.startTimer;
    if (start !== undefined) {
      timers = [...timers, {
        id: start.id,
        deadline: now + start.durationMs,
        moveName: start.moveName,
      }];
      shown[start.id] = { TimeLeft: start.durationMs };
    }
    recent.push({
      version,
      game: stepped.game,
      moveName: stepped.moveName,
      timers: shown,
      settled: index === outcome.versions.length - 1,
    });
    history.push({
      version,
      name: stepped.moveName,
      seat: index === 0 ? seat : ADMIN_SEAT,
      arguments: index === 0 ? args : {},
      at: now,
    });
  });
  const head = recent[recent.length - 1].game;
  return accepted(table, {
    phase: rules.outcome(head).finished ? "finished" : "playing",
    version,
    recent: recent.slice(-WINDOW),
    history,
    timers: rules.outcome(head).finished ? [] : timers,
  });
}

/**
 * The table after `proposal`, made by `context.proposer` at `context.now`.
 * A refused proposal returns the table unchanged with the reason.
 *
 * This is the whole write side of a table. It is the only place a seat is
 * taken, cards are dealt, a move is applied or a timer fires.
 */
export function stepTable<G, S>(
  rules: GameRules<G, S>,
  table: TableState<G>,
  proposal: Proposal,
  context: StepContext,
): TableStep<G> {
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
      if (table.seats.length >= rules.maxPlayers) {
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
      if (table.seats.length < rules.minPlayers) {
        return refused(table, "not enough players are seated");
      }
      // The one draw of fresh randomness. Everything else about the game
      // derives from these two values.
      const gameId = randomHex(context.entropy, 16);
      const secretSalt = randomHex(context.entropy, 32);
      const game = rules.setUp(
        { gameId, secretSalt },
        table.seats.length,
        proposal.options ?? {},
      );
      return accepted(table, {
        phase: "playing",
        gameId,
        secretSalt,
        // A guest keys what it has shown by epoch, so a deal it speculated
        // locally and the deal the server made never mix.
        epoch: parseInt(gameId.slice(0, 8), 16),
        version: 0,
        recent: [{
          version: 0,
          game,
          moveName: "",
          timers: {},
          settled: true,
        }],
        history: [],
        timers: [],
      });
    }
    case "restart": {
      if (table.phase !== "finished") {
        return refused(table, "the game is not over");
      }
      if (seat < 0) return refused(table, "only a seated player may reset");
      return accepted(table, {
        phase: "seating",
        gameId: "",
        secretSalt: "",
        version: -1,
        recent: [],
        history: [],
        timers: [],
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
      const args = proposal.arguments ?? {};
      const head = table.recent[table.recent.length - 1];
      const outcome = rules.step(
        head.game,
        table.version,
        { gameId: table.gameId, secretSalt: table.secretSalt },
        name,
        args,
        seat,
      );
      if (outcome.error !== "" || outcome.versions.length === 0) {
        return refused(table, outcome.error || "the move changed nothing");
      }
      return advance(rules, table, outcome, seat, args, context.now);
    }
    case "tick": {
      if (table.phase !== "playing") {
        return refused(table, "no game is being played");
      }
      const timer = table.timers.find((running) =>
        running.id === proposal.timerId
      );
      if (timer === undefined) return refused(table, "no such timer");
      // A tick is a request to look at the clock, not a claim about it.
      if (context.now < timer.deadline) {
        return refused(table, "the timer has not run out");
      }
      const head = table.recent[table.recent.length - 1];
      const outcome = rules.step(
        head.game,
        table.version,
        { gameId: table.gameId, secretSalt: table.secretSalt },
        timer.moveName,
        {},
        ADMIN_SEAT,
      );
      const without = {
        ...table,
        timers: table.timers.filter((running) => running.id !== timer.id),
      };
      if (outcome.error !== "" || outcome.versions.length === 0) {
        // The timer's move is no longer legal; the timer is spent.
        return accepted(without, {});
      }
      return advance(
        rules,
        without,
        outcome,
        ADMIN_SEAT,
        {},
        context.now,
      );
    }
    default:
      return refused(table, "unknown kind of proposal");
  }
}

function copyTimers(
  timers: Record<string, TimerView>,
): Record<string, TimerView> {
  const copy: Record<string, TimerView> = {};
  for (const id of Object.keys(timers ?? {})) {
    copy[id] = { TimeLeft: timers[id].TimeLeft };
  }
  return copy;
}

/**
 * What the table's readers may see: who is seated, and every version the
 * table still holds, sanitized, with the move that produced it.
 *
 * It is total: it reads only fields a `TableState` has, and a table that is
 * missing one is shown as empty. Nothing in it depends on who asks; the
 * parts that differ by viewer are all present, keyed by seat.
 *
 * Every object and array in the result is new. A part of the table returned
 * as it is would reach the store as a reference to the sealed table, which
 * carries the table's label and not this function's, and the whole view
 * would be withheld for it (FINDINGS.md of the Memory pattern, F4). The same
 * goes for what `rules.sanitize` and `rules.viewerPart` return.
 */
export function viewOf<G, S>(
  rules: GameRules<G, S>,
  table: TableState<G>,
): TableView<S> {
  const seats = Array.isArray(table.seats) ? table.seats : [];
  const recent = Array.isArray(table.recent) ? table.recent : [];
  const entries: ViewEntry<S>[] = recent.map((recorded) => {
    const viewers: Record<string, ViewerPart> = {};
    for (let viewer = OBSERVER_SEAT; viewer < seats.length; viewer++) {
      viewers[String(viewer)] = rules.viewerPart(
        recorded.game,
        recorded.version,
        viewer,
      );
    }
    const move: MoveRecord | undefined = recorded.moveName === ""
      ? undefined
      : { AnimationKey: recorded.moveName, Version: recorded.version };
    const entry: ViewEntry<S> = {
      version: recorded.version,
      // Memory shows every viewer the same state. A game with per-seat
      // secrets needs one released view a seat instead (RESEARCH.md,
      // "Dealing to a player"); the observer's is what is shared here.
      state: rules.sanitize(recorded.game, OBSERVER_SEAT),
      currentPlayerIndex: rules.currentPlayer(recorded.game),
      timers: copyTimers(recorded.timers),
      outcome: rules.outcome(recorded.game),
      settled: recorded.settled,
      viewers,
    };
    if (move !== undefined) entry.move = move;
    return entry;
  });
  return {
    gameName: rules.name,
    phase: table.phase,
    epoch: table.epoch,
    revision: table.revision,
    version: table.version,
    seats: seats.map((seat) => ({
      principal: seat.principal,
      label: seat.label,
    })),
    minPlayers: rules.minPlayers,
    maxPlayers: rules.maxPlayers,
    entries,
  };
}
