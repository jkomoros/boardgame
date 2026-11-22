/**
 * The wire shapes a table hands its guest. They are the boardgame client's
 * own: a state entry here is what `fabric/guest/PROTOCOL.md` calls a `state`
 * entry, with the parts that differ by viewer held under `viewers`.
 *
 * This module holds types only, so a pattern and the guest's bridge can
 * share it.
 */
import type { Json, JsonObject } from "./json.ts";

/** The sanitized record of the move that produced a version. */
export interface MoveRecord {
  AnimationKey: string;
  Version: number;
  /** The move's arguments, as far as this viewer may see them. */
  Properties?: JsonObject;
}

export interface MoveLegality {
  legalForPlayer: boolean;
  legalForAnyone: boolean;
  error?: string;
}

export interface ChoiceCandidate {
  Value: Json;
  Available: boolean;
}

export interface ChoiceSet {
  MoveName: string;
  FieldName: string;
  Source: string;
  Candidates: ChoiceCandidate[];
}

/** Per-candidate legality for single-argument moves, in the Go API's shape. */
export interface ProjectedChoices {
  StateVersion: number;
  Status: string;
  MoveChoiceProjectionSchemaFingerprint: string;
  ProjectionSchemaVersion: number;
  Sets: ChoiceSet[];
}

/** What of a version depends on who is looking. */
export interface ViewerPart {
  moveLegality: Record<string, MoveLegality>;
  zeroInputMoves: string[];
  /** `Sets` is empty for a viewer with nothing to choose. */
  projectedMoveChoices: ProjectedChoices;
}

export interface TimerView {
  /** Milliseconds left when the version was produced. */
  TimeLeft: number;
}

export interface Outcome {
  finished: boolean;
  winners: number[];
}

/**
 * One state version, as the view's viewer may see it. `S` is the state's
 * type for a consumer that wants a narrower one than JSON; the guest's
 * bridge names it.
 */
export interface ViewEntry<S = JsonObject> {
  version: number;
  /** The sanitized state: what `Game.JSONForPlayer(...).CurrentState` holds. */
  state: S;
  /** The move that produced this version; absent for the first. */
  move?: MoveRecord;
  currentPlayerIndex: number;
  timers: Record<string, TimerView>;
  outcome: Outcome;
  /** False for a version a fix-up move replaced at once. */
  settled: boolean;
  /**
   * The viewer's own part, under their seat as a string; `"-1"` is whoever
   * only watches. A view is computed for one viewer and holds no other's.
   * Empty for a version that is not settled.
   */
  viewers: Record<string, ViewerPart>;
}

/** A seat as a view names it: by the name its occupant chose, not by who they are. */
export interface SeatView {
  label: string;
}

/**
 * What a table releases to one principal: who is seated, and the versions
 * the last input produced with the move that made each, as that principal
 * may see them.
 */
export interface TableView<S = JsonObject> {
  gameName: string;
  phase: "seating" | "playing" | "finished";
  /** Changes whenever history stops being comparable: each new deal. */
  epoch: number;
  /** How many proposals the table has accepted, seating included. */
  revision: number;
  /** The head game version, or -1 before the deal. */
  version: number;
  /**
   * The seat of the principal this view was computed for, or -1 when they
   * hold none. Every entry is sanitized for this seat.
   */
  viewer: number;
  seats: SeatView[];
  minPlayers: number;
  maxPlayers: number;
  /**
   * The versions the last input produced, ascending: the move a proposal
   * made and each fix-up after it. Empty before the deal.
   */
  entries: ViewEntry<S>[];
}

/** One value a setup option can take. */
export interface SetupValue {
  name: string;
  displayName: string;
}

/**
 * One choice a table offers before the deal: a variant of the game, as the
 * engine's `variantConfig` describes it. `start` takes `{ [name]: value }`.
 */
export interface SetupOption {
  name: string;
  displayName: string;
  /** The value used when `start` names none. */
  default: string;
  values: SetupValue[];
}

/** Static, public facts about a game that its renderer and its lobby need once. */
export interface TableInfo<V = Json> {
  gameName: string;
  /** The game's name for people. */
  displayName: string;
  /** What `start` may be given, in the game's own order. Empty for none. */
  setupOptions: SetupOption[];
  /**
   * Decks, enums, constants and message templates, as Go marshals them. A
   * deck is a list of `{ Index, Values }`, with `Values` of type `V`.
   */
  chest: { [key: string]: Json | V };
  /** The engine's fingerprint of its move inputs. */
  moveInputSchemaFingerprint: string;
  /** Colours for seats, in seat order. */
  seatColors: string[];
}

/**
 * What a browser sends the table. It names no proposer: the table takes that
 * from the runtime.
 */
export interface Proposal {
  /**
   * `sit` takes the next seat and `leave` gives one up, before the deal.
   * `start` deals. `play` proposes a move of the game. `tick` asks the table
   * to check a timer's deadline against its own clock. `restart` clears a
   * finished table for a new deal.
   *
   * TODO(cfc): `tick` exists because nothing wakes a pattern at a time. A
   * guest that was shown a timer sends it when its own clock runs out, and
   * the table fires the timer only if its own clock agrees. With a wake at
   * an instant a pattern's state names (labs
   * `docs/plans/scheduled-work-in-the-server.md` proposes one on an
   * interval of minutes, which is too coarse), the table would fire its own
   * timers and no client would be involved.
   */
  kind: "sit" | "leave" | "start" | "play" | "tick" | "restart";
  /** `sit`: the name to show for the seat. */
  label?: string;
  /** `start`: the game's own setup options. */
  options?: Record<string, string>;
  /** `play`: the move's name. */
  name?: string;
  /** `play`: the move's arguments, as the Go API's move forms take them. */
  arguments?: Record<string, string>;
  /** `tick`: the timer the guest watched run out. */
  timerId?: string;
}
