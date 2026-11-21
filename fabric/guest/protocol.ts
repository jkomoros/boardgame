/**
 * The message protocol between a renderer host page ("the host") and the
 * boardgame guest document. See PROTOCOL.md for the prose contract.
 *
 * Every message is plain JSON. Nothing here imports from the client source,
 * so a host (a Common Fabric pattern, the test harness) can use these types
 * without pulling in Lit.
 */

export const GUEST_CHANNEL = 'boardgame-guest' as const;
export const GUEST_PROTOCOL_VERSION = 1 as const;

/** The seat a view was computed for. -1 is the observer. */
export type ViewerSeat = number;
export const OBSERVER_SEAT = -1;

export type JsonValue =
  | null
  | boolean
  | number
  | string
  | readonly JsonValue[]
  | { readonly [key: string]: JsonValue };

interface Envelope<Type extends string> {
  readonly channel: typeof GUEST_CHANNEL;
  readonly v: typeof GUEST_PROTOCOL_VERSION;
  readonly type: Type;
}

// ---------------------------------------------------------------------------
// Shared shapes
// ---------------------------------------------------------------------------

/** How one seat is named and coloured. `playerIndex` equals its position. */
export interface GuestPlayerPresentation {
  readonly playerIndex: number;
  readonly label: string;
  readonly color?: string;
}

/** The chest as the Go manager marshals it: decks, enums, constants. */
export interface GuestChest {
  readonly Decks: { readonly [deck: string]: readonly JsonValue[] };
  readonly Enums?: JsonValue;
  readonly Constants?: JsonValue;
  readonly LegalTemplates?: { readonly [key: string]: string };
}

/**
 * The sanitized record of the move that produced a version. The move's name
 * crosses as `AnimationKey`; there is deliberately no `Name`.
 */
export interface GuestMoveRecord {
  readonly AnimationKey: string;
  readonly Version: number;
  readonly Properties?: { readonly [key: string]: JsonValue };
}

export interface GuestMoveLegality {
  readonly legalForPlayer: boolean;
  readonly legalForAnyone: boolean;
  readonly error?: string;
}

/** Candidate legality for one move field, as the Go API projects it. */
export interface GuestProjectedChoices {
  readonly StateVersion: number;
  readonly Status: string;
  readonly MoveChoiceProjectionSchemaFingerprint?: string;
  readonly ProjectionSchemaVersion?: number;
  readonly Sets?: readonly {
    readonly MoveName: string;
    readonly FieldName: string;
    readonly Source?: string;
    readonly Candidates: readonly {
      readonly Value: JsonValue;
      readonly Available: boolean;
      readonly [extra: string]: JsonValue;
    }[];
    readonly [extra: string]: JsonValue | undefined;
  }[];
  readonly [extra: string]: JsonValue | undefined;
}

/** One state version, as one viewer is allowed to see it. */
export interface GuestStateEntry {
  /** The state version. Strictly increasing within one epoch. */
  readonly version: number;
  /**
   * The viewer's sanitized state, unexpanded: `Game`, `Players`,
   * `Components`, `Computed`, `Visibility`, exactly as
   * `Game.JSONForPlayer(...).CurrentState` marshals it.
   */
  readonly state: { readonly [key: string]: JsonValue };
  /** The move that produced this version; null for the first version. */
  readonly move: GuestMoveRecord | null;
  readonly currentPlayerIndex: number;
  /** Timers running at this version: milliseconds left when it was produced. */
  readonly timers?: { readonly [timerId: string]: { readonly TimeLeft: number } };
  readonly outcome: { readonly finished: boolean; readonly winners: readonly number[] };
  /**
   * Whether each move type is legal with its default arguments. Present only
   * for a version at which the engine waits for a player. An entry without it
   * is treated as "nothing is legal here".
   */
  readonly moveLegality?: { readonly [moveName: string]: GuestMoveLegality };
  /** Moves that need no argument from the player. */
  readonly zeroInputMoves?: readonly string[];
  /** Per-candidate legality for moves that take one argument. */
  readonly projectedMoveChoices?: GuestProjectedChoices | null;
}

// ---------------------------------------------------------------------------
// Host to guest
// ---------------------------------------------------------------------------

/**
 * Static facts about the table. Sending `init` again resets the guest: the
 * queue is dropped and the next `state` is shown without animation.
 */
export interface InitMessage extends Envelope<'init'> {
  readonly gameName: string;
  /** Opaque table identity. Used only to seed effects; never sent anywhere. */
  readonly gameId?: string;
  readonly chest: GuestChest;
  readonly viewer: ViewerSeat;
  readonly players: readonly GuestPlayerPresentation[];
  /**
   * The engine's move-input schema fingerprint. When absent the guest uses
   * the fingerprint compiled into its bundled renderer, which is only safe
   * when the host and the guest were built from the same game definition.
   */
  readonly moveInputSchemaFingerprint?: string;
}

/**
 * A window of consecutive versions, ascending, ending at the newest version
 * the host knows. The guest plays every version it has not yet queued and
 * ignores the rest, so re-sending an overlapping window is harmless.
 */
export interface StateMessage extends Envelope<'state'> {
  /** Must equal the viewer the entries were computed for. */
  readonly viewer: ViewerSeat;
  readonly entries: readonly GuestStateEntry[];
  /**
   * Changes when history is no longer comparable with what the guest has
   * shown (a new game on the same table, a rewind). A new epoch resets the
   * guest exactly as a viewer change does. Defaults to 0.
   */
  readonly epoch?: number;
}

/** The host's verdict on one `propose-move`. Optional; see PROTOCOL.md. */
export interface MoveResultMessage extends Envelope<'move-result'> {
  readonly requestId: string;
  readonly ok: boolean;
  readonly error?: string;
  /** The head version, when the move was refused for being stale. */
  readonly actualVersion?: number;
}

export type HostToGuestMessage = InitMessage | StateMessage | MoveResultMessage;

// ---------------------------------------------------------------------------
// Guest to host
// ---------------------------------------------------------------------------

/**
 * Sent twice: `loaded` once the guest is listening (the host may send `init`),
 * and `mounted` once `init` has been applied and the renderer exists.
 */
export interface ReadyMessage extends Envelope<'ready'> {
  readonly stage: 'loaded' | 'mounted';
  /** Games this bundle contains a renderer for. */
  readonly games: readonly string[];
}

/** One version has finished animating and the next (if any) is being shown. */
export interface AnimationsDoneMessage extends Envelope<'animations-done'> {
  readonly version: number;
  readonly viewer: ViewerSeat;
  readonly epoch: number;
  /** Versions still queued behind this one. */
  readonly remaining: number;
  /**
   * `played`: installed and animated. `initial`: the first version after a
   * reset, shown without animation. `skipped`: the renderer asked for this
   * intermediate version not to be shown.
   */
  readonly how: 'played' | 'initial' | 'skipped';
  /**
   * Present when this version was not the successor of the one shown before
   * it: the previous version. The versions between were never delivered, so
   * the guest animated straight from one to the other.
   */
  readonly jumpedFrom?: number;
}

/** The person made a move. The host decides whether it happens. */
export interface ProposeMoveMessage extends Envelope<'propose-move'> {
  readonly requestId: string;
  readonly name: string;
  /** Arguments serialized as strings, as the Go API's move forms take them. */
  readonly arguments: { readonly [field: string]: string };
  /** The version that was on screen when the gesture happened. */
  readonly snapshotVersion: number;
  readonly viewer: ViewerSeat;
}

/**
 * A timer the guest was shown has reached zero on the guest's clock. A
 * request to check deadlines, not a claim that anything is due: the host
 * compares the timer with its own clock.
 */
export interface TickMessage extends Envelope<'tick'> {
  readonly timerId: string;
  /** The version whose timers were on screen. */
  readonly version: number;
}

/** Something the host should surface while developing; never required. */
export interface GuestErrorMessage extends Envelope<'error'> {
  readonly message: string;
}

export type GuestToHostMessage =
  | ReadyMessage
  | AnimationsDoneMessage
  | ProposeMoveMessage
  | TickMessage
  | GuestErrorMessage;

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

export function isGuestMessage(value: unknown): value is { readonly type: string } {
  if (value === null || typeof value !== 'object') return false;
  const record = value as { readonly [key: string]: unknown };
  return record['channel'] === GUEST_CHANNEL
    && record['v'] === GUEST_PROTOCOL_VERSION
    && typeof record['type'] === 'string';
}

/** Body of a message: everything but the envelope fields. */
export type MessageBody<M extends { readonly type: string }> =
  M extends unknown ? Omit<M, 'channel' | 'v'> : never;

export function envelope<M extends HostToGuestMessage | GuestToHostMessage>(
  body: MessageBody<M>,
): M {
  return { channel: GUEST_CHANNEL, v: GUEST_PROTOCOL_VERSION, ...body } as unknown as M;
}
