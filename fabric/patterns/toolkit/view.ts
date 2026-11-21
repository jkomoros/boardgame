/**
 * The wire shapes a table hands its guest. They are the boardgame client's
 * own: a state entry here is what `fabric/guest/PROTOCOL.md` calls a `state`
 * entry, with the parts that differ by viewer held under `viewers`.
 *
 * This module holds types only and imports nothing, so a pattern, the rules
 * of a game and the guest's bridge can all share it.
 */

/** A stack as the Go server marshals one, and as the client expands one. */
export interface StackJSON {
  Deck: string;
  /** A deck index a slot, `-1` for an empty slot, `-2` for a hidden card. */
  Indexes: number[];
  /** A component id a slot, stable across reveals; `""` for an empty slot. */
  IDs: string[];
  /** The version each id was last put into this stack. */
  IDsLastSeen: Record<string, number>;
  ShuffleCount: number;
  /** Present on a sized stack. */
  Size?: number;
}

/** The sanitized record of the move that produced a version. */
export interface MoveRecord {
  AnimationKey: string;
  Version: number;
}

export interface MoveLegality {
  legalForPlayer: boolean;
  legalForAnyone: boolean;
  error?: string;
}

export interface ChoiceCandidate {
  Value: number;
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
  Status: "ready" | "failed";
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

/** One state version, as the table's readers may see it. */
export interface ViewEntry<S> {
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
  /** Keyed by seat as a string; `"-1"` is the observer. */
  viewers: Record<string, ViewerPart>;
}

export interface SeatView {
  /** The DID of whoever holds the seat. */
  principal: string;
  label: string;
}

/**
 * Everything a table releases to its readers: who is seated, and a window of
 * the most recent versions with the move that produced each.
 */
export interface TableView<S> {
  gameName: string;
  phase: "seating" | "playing" | "finished";
  /** Changes whenever history stops being comparable: each new deal. */
  epoch: number;
  /** How many proposals the table has accepted, seating included. */
  revision: number;
  /** The head game version, or -1 before the deal. */
  version: number;
  seats: SeatView[];
  minPlayers: number;
  maxPlayers: number;
  /** Ascending by version; empty before the deal. */
  entries: ViewEntry<S>[];
}

/** One card (or other component) of a deck, as the chest lists it. */
export interface ChestComponent<V> {
  Index: number;
  Values: V;
}

/** Static, public facts about a game that its renderer needs once. */
export interface TableInfo<V> {
  gameName: string;
  chest: {
    Decks: Record<string, ChestComponent<V>[]>;
    Enums: Record<string, never>;
    LegalTemplates: Record<string, string>;
  };
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
