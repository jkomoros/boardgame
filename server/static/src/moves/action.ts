import type { MoveInputIssue } from './input.js';
import {
  createTargetAction,
  type TargetAction,
  type TargetActionOptions,
  type TargetKey,
  type TargetPreviewTransport,
} from './target-action.ts';

const MOVE_ACTION_BRAND = Symbol('boardgame.move-action');
const CANCEL_PREVIEW = Symbol('boardgame.cancel-move-preview');
const MANAGED_PREVIEW = Symbol('boardgame.managed-move-preview');
const HYDRATE_PREVIEW = Symbol('boardgame.hydrate-move-preview');

export type MoveActionReasonCode =
  | 'move-not-possible'
  | 'not-legal-for-player'
  | 'animation-running'
  | 'schema-mismatch'
  | 'transport-unavailable'
  | 'preview-unchecked'
  | 'preview-illegal'
  | 'preview-failed'
  | 'submission-pending'
  | 'another-submission-pending'
  | 'snapshot-consumed'
  | 'stale-snapshot'
  | 'submission-rejected'
  | 'invalid-input';

export interface MoveActionReason {
  readonly code: MoveActionReasonCode;
  readonly message: string;
  /** Exact runtime input failures, when code is invalid-input. */
  readonly issues?: readonly MoveInputIssue[];
  /** Server-supplied declarative checks behind an availability verdict. */
  readonly preconditions?: readonly MoveActionPrecondition[];
}

/**
 * How LOUD a reason is, which is the one thing every consumer of `reason` has to
 * decide and none of them could.
 *
 * `reason` answers "why can this control not be used right now?", and its
 * answers are not the same kind of thing. A UI that paints all of them the same
 * way is wrong for most of them: a die that rolls for two seconds reports
 * `animation-running` and, because the client is still displaying the state
 * BEFORE the throw, `move-not-possible` — and a control that renders every
 * reason in the error colour then puts red text under the die for the whole of
 * every roll, saying nothing is wrong in the loudest available voice.
 *
 *   - `busy`: a state that clears itself with no one doing anything. Waiting for
 *     an animation, for a submission, or for a legality check. Never worth an
 *     error style, and usually not worth any visible text at all — the control
 *     is disabled, which already says "not now".
 *   - `unavailable`: the move is not on offer in this state. Legitimate,
 *     expected, and the normal condition of most controls most of the time; the
 *     disabled control is the message.
 *   - `error`: something is actually wrong and a person has to act — a schema
 *     that does not match the server, a transport that is not connected, input
 *     that does not validate, a rejection. This is what an error style is for.
 *
 * Stated here rather than in each component so that "is this worth shouting
 * about?" has ONE answer, and so that a new code has to be classified before it
 * can be reported: the table below is exhaustive over `MoveActionReasonCode`.
 */
export type MoveActionReasonSeverity = 'busy' | 'unavailable' | 'error';

const REASON_SEVERITY: Readonly<Record<MoveActionReasonCode, MoveActionReasonSeverity>> = {
  // Clears itself.
  'animation-running': 'busy',
  'submission-pending': 'busy',
  'another-submission-pending': 'busy',
  'snapshot-consumed': 'busy',
  'stale-snapshot': 'busy',
  'preview-unchecked': 'busy',
  // Not on offer in this state.
  'move-not-possible': 'unavailable',
  'not-legal-for-player': 'unavailable',
  'preview-illegal': 'unavailable',
  // Someone has to do something.
  'schema-mismatch': 'error',
  'transport-unavailable': 'error',
  'preview-failed': 'error',
  'submission-rejected': 'error',
  'invalid-input': 'error',
};

/**
 * How loud a reason is; see `MoveActionReasonSeverity`.
 *
 * An unrecognised code is an `error`, deliberately: a code this table has not
 * been taught about is more likely to be a real failure than a transient, and
 * the visible consequence of guessing wrong that way is a message that should
 * have been quiet rather than a failure that was silent.
 */
export function moveActionReasonSeverity(
  reason: MoveActionReason | MoveActionReasonCode | null | undefined,
): MoveActionReasonSeverity | null {
  if (reason === null || reason === undefined) return null;
  const code = typeof reason === 'string' ? reason : reason.code;
  return REASON_SEVERITY[code] ?? 'error';
}

export interface MoveActionPrecondition {
  readonly name: string;
  readonly args?: readonly string[];
  readonly verdict: 'pass' | 'fail' | 'unknown';
  readonly evaluable: boolean;
  readonly provisional?: boolean;
  readonly message?: {
    readonly template: string;
    readonly bindings?: Readonly<Record<string, string | number | boolean>>;
  };
}

export type MoveActionAvailability =
  | { readonly kind: 'available' }
  | { readonly kind: 'unavailable'; readonly reason: MoveActionReason };

export type MoveActionPreview =
  | { readonly kind: 'not-needed' }
  | { readonly kind: 'unchecked' }
  | { readonly kind: 'checking'; readonly requestID: string }
  | { readonly kind: 'legal'; readonly preconditions?: readonly MoveActionPrecondition[] }
  | { readonly kind: 'illegal'; readonly reason: MoveActionReason }
  | { readonly kind: 'failed'; readonly reason: MoveActionReason; readonly retryable: boolean };

export interface MovePreviewRequest<MoveName extends string = string> {
  readonly requestID: string;
  readonly snapshotVersion: number;
  readonly viewingAsPlayer: number;
  readonly proposingAsPlayer: number;
  readonly proposingAsAdmin: boolean;
  readonly name: MoveName;
  readonly arguments: Readonly<Record<string, string>>;
  readonly candidateKey: string;
  readonly signal: AbortSignal;
}

export type MovePreviewTransportResult =
  | {
    readonly kind: 'success';
    readonly legal: boolean;
    readonly error?: string;
    readonly preconditions?: readonly MoveActionPrecondition[];
  }
  | { readonly kind: 'stale-snapshot'; readonly expectedVersion: number; readonly actualVersion: number }
  | { readonly kind: 'failure'; readonly error: string; readonly friendlyError?: string; readonly retryable: boolean };

export interface MovePreviewTransport {
  preview(request: MovePreviewRequest): Promise<MovePreviewTransportResult>;
}

export type MoveActionSubmission =
  | { readonly kind: 'idle' }
  | { readonly kind: 'pending'; readonly requestID: string }
  | {
    readonly kind: 'rejected';
    readonly requestID: string;
    readonly source: 'server' | 'network';
    readonly reason: string;
    readonly retryable: boolean;
  };

export interface MoveSubmissionRequest<MoveName extends string = string> {
  readonly requestID: string;
  readonly snapshotVersion: number;
  readonly viewingAsPlayer: number;
  readonly proposingAsPlayer: number;
  readonly proposingAsAdmin: boolean;
  readonly name: MoveName;
  readonly arguments: Readonly<Record<string, string>>;
}

export type MoveTransportResult =
  | { readonly kind: 'success' }
  | {
    readonly kind: 'server-rejection';
    readonly error: string;
    readonly friendlyError?: string;
    readonly code?: string;
    readonly expectedVersion?: number;
    readonly actualVersion?: number;
  }
  | { readonly kind: 'network-failure'; readonly error: string; readonly friendlyError?: string };

export interface MoveTransport {
  submit(request: MoveSubmissionRequest): Promise<MoveTransportResult>;
}

export type MoveProposalResult =
  | { readonly kind: 'success'; readonly requestID: string }
  | { readonly kind: 'server-rejection'; readonly requestID: string; readonly error: string; readonly friendlyError?: string }
  | { readonly kind: 'network-failure'; readonly requestID: string; readonly error: string; readonly friendlyError?: string }
  | { readonly kind: 'blocked'; readonly requestID: string; readonly reason: MoveActionReason }
  | { readonly kind: 'stale-snapshot'; readonly requestID: string; readonly expectedVersion: number; readonly actualVersion: number };

export interface MoveActionTelemetryEvent<MoveName extends string = string> {
  readonly name: MoveName;
  readonly requestID: string;
  readonly result: MoveProposalResult;
}

export interface MoveActionLegality {
  readonly legalForPlayer: boolean;
  readonly legalForAnyone: boolean;
  readonly error?: string;
  readonly preconditions?: readonly MoveActionPrecondition[];
}

/** Stable per-renderer dependencies shared by every action and snapshot. */
export interface MoveActionService {
  readonly currentClientSchemaFingerprint: () => string;
  readonly currentServerSchemaFingerprint: () => string | null;
  readonly currentTransport: () => MoveTransport | null;
  readonly currentPreviewTransport: () => MovePreviewTransport | null;
  readonly currentTargetPreviewTransport: () => TargetPreviewTransport | null;
  readonly currentGate: () => MoveSubmissionGate;
  readonly nextRequestID: () => string;
  readonly validate: (moveName: string, input: unknown) => readonly MoveInputIssue[];
  readonly serialize: (moveName: string, input: unknown) => Readonly<Record<string, string>>;
  readonly changed?: () => void;
  readonly telemetry?: (event: MoveActionTelemetryEvent) => void;
  readonly actionCache?: Map<string, BoundMoveAction<string, object>>;
  readonly targetActionCache?: Map<string, TargetAction<TargetKey, string, object>>;
}

/** Immutable identity plus live gates for one renderer snapshot. */
export interface MoveActionSnapshot {
  /** Route + connection epoch + version + viewer/admin perspective + schema. */
  readonly snapshotKey: string;
  readonly currentSnapshotKey: () => string;
  readonly snapshotVersion: number;
  readonly currentSnapshotVersion: () => number;
  readonly viewingAsPlayer: number;
  readonly proposingAsPlayer: number;
  readonly proposingAsAdmin: boolean;
  readonly currentLegality: () => MoveActionLegality | undefined;
  readonly currentAnimating: () => boolean;
  /** False when proposal/admin perspective differs from the baseline form. */
  readonly baselineLegalityApplies: boolean;
  readonly actionCacheKey?: string;
}

export interface MoveSnapshotIdentity {
  readonly gameName: string;
  readonly gameID: string;
  readonly epoch: number;
  readonly version: number;
  readonly viewingAsPlayer: number;
  readonly proposingAsPlayer: number;
  readonly proposingAsAdmin: boolean;
  readonly serverSchemaFingerprint: string | null;
}

export function moveSnapshotKey(identity: MoveSnapshotIdentity): string {
  return JSON.stringify([
    identity.gameName,
    identity.gameID,
    identity.epoch,
    identity.version,
    identity.viewingAsPlayer,
    identity.proposingAsPlayer,
    identity.proposingAsAdmin,
    identity.serverSchemaFingerprint,
  ]);
}

export class MoveSubmissionGate {
  #requestID: string | null = null;
  #consumedSnapshotKey: string | null = null;
  readonly #listeners = new Set<() => void>();

  get pendingRequestID(): string | null {
    return this.#requestID;
  }

  acquire(requestID: string, snapshotKey: string): 'acquired' | 'busy' | 'consumed' {
    if (this.#consumedSnapshotKey === snapshotKey) return 'consumed';
    if (this.#requestID !== null) return 'busy';
    this.#requestID = requestID;
    this.#notify();
    return 'acquired';
  }

  release(requestID: string): void {
    if (this.#requestID !== requestID) return;
    this.#requestID = null;
    this.#notify();
  }

  consume(snapshotKey: string): void {
    this.#consumedSnapshotKey = snapshotKey;
    this.#notify();
  }

  isConsumed(snapshotKey: string): boolean {
    return this.#consumedSnapshotKey === snapshotKey;
  }

  subscribe(listener: () => void): () => void {
    this.#listeners.add(listener);
    return () => this.#listeners.delete(listener);
  }

  #notify(): void {
    for (const listener of this.#listeners) notifyListener(listener);
  }
}

type InputFor<K extends string, Inputs extends Record<string, object>> =
  K extends keyof Inputs ? Inputs[K] : Record<string, unknown>;

type ExactInput<Expected extends object, Actual extends Expected> = Actual &
  Record<Exclude<keyof Actual, keyof Expected>, never>;

export interface MoveActionState<MoveName extends string> {
  readonly name: MoveName;
  readonly availability: MoveActionAvailability;
  readonly preview: MoveActionPreview;
  readonly submission: MoveActionSubmission;
  readonly canPropose: boolean;
  readonly reason: MoveActionReason | null;
}

export interface MoveActionBuilder<MoveName extends string, Input extends object>
  extends MoveActionState<MoveName> {
  with<Actual extends Input>(input: ExactInput<Input, Actual>): BoundMoveAction<MoveName, Input>;
  targets<Key extends TargetKey, Actual extends Input>(
    keys: readonly Key[],
    inputFor: (key: Key, index: number) => ExactInput<Input, Actual>,
    options?: TargetActionOptions,
  ): TargetAction<Key, MoveName, Input>;
}

export interface BoundMoveAction<MoveName extends string, Input extends object>
  extends MoveActionState<MoveName> {
  readonly input: Readonly<Input>;
  /** True when activation can propose now or retry a transient preview failure. */
  readonly canActivate: boolean;
  ensurePreview(): Promise<MoveActionPreview>;
  refreshPreview(): Promise<MoveActionPreview>;
  /** UI-safe activation: retries a transient preview failure, then proposes. */
  activate(): Promise<MoveProposalResult>;
  propose(): Promise<MoveProposalResult>;
  subscribe(listener: () => void): () => void;
}

export function isBoundMoveAction(
  value: unknown,
): value is BoundMoveAction<string, object> {
  if (typeof value !== 'object' || value === null) return false;
  const candidate = value as Partial<BoundMoveAction<string, object>> & {
    readonly [MOVE_ACTION_BRAND]?: boolean;
  };
  return candidate[MOVE_ACTION_BRAND] === true
    && typeof candidate.propose === 'function'
    && typeof candidate.subscribe === 'function';
}

/** Host lifecycle hook; intentionally absent from the creator action interface. */
export function cancelMoveActionPreview(action: BoundMoveAction<string, object>): void {
  const internal = action as BoundMoveAction<string, object> & { [CANCEL_PREVIEW]?: () => void };
  internal[CANCEL_PREVIEW]?.();
}

/**
 * Host lifecycle hook for live dependencies that intentionally do not change
 * an action's snapshot identity (for example, the animation gate). Cached
 * actions keep stable identity, so their subscribed controls need an explicit
 * notification when one of those dependencies changes.
 */
export function notifyMoveActionLiveStateChanged(action: BoundMoveAction<string, object>): void {
  const internal = action as MoveActionImplementation<string, object>;
  internal.notifyLiveStateChanged();
}

/**
 * Hydrate an ordinary bound action with an authoritative legality projection.
 * This is a host/runtime hook: game code receives the resulting action and
 * uses it exactly like one produced by move(...).with(...).
 */
export function hydrateMoveActionPreview(
  action: BoundMoveAction<string, object>,
  preview: Extract<MoveActionPreview, { readonly kind: 'legal' | 'illegal' }>,
): void {
  const internal = action as MoveActionImplementation<string, object>;
  internal[HYDRATE_PREVIEW](preview);
}

export type MoveActionFor<MoveName extends string, Input extends object> =
  [Input[keyof Input]] extends [never]
    ? BoundMoveAction<MoveName, Input>
    : MoveActionBuilder<MoveName, Input>;

export function createMoveAction<
  K extends MoveName,
  MoveName extends string,
  Inputs extends Record<string, object>,
>(
  name: K,
  service: MoveActionService,
  snapshot: MoveActionSnapshot,
): MoveActionFor<K, InputFor<K, Inputs>> {
  const action = cachedAction<K, InputFor<K, Inputs>>(
    name,
    service,
    snapshot,
    {} as InputFor<K, Inputs>,
  );
  return action as MoveActionFor<K, InputFor<K, Inputs>>;
}

class MoveActionImplementation<MoveName extends string, Input extends object>
implements BoundMoveAction<MoveName, Input> {
  readonly name: MoveName;
  readonly input: Readonly<Input>;
  readonly [MOVE_ACTION_BRAND]: boolean;
  readonly #service: MoveActionService;
  readonly #snapshot: MoveActionSnapshot;
  readonly #inputWasBound: boolean;
  readonly #inputReason: MoveActionReason | null;
  #submission: MoveActionSubmission = { kind: 'idle' };
  #preview: MoveActionPreview;
  #previewPromise: Promise<MoveActionPreview> | null = null;
  #previewAbort: AbortController | null = null;
  readonly #listeners = new Set<() => void>();
  #gateUnsubscribe: (() => void) | null = null;
  #managedPreview: ((force: boolean) => Promise<void>) | null = null;
  #managedChanged: (() => void) | null = null;

  constructor(
    name: MoveName,
    service: MoveActionService,
    snapshot: MoveActionSnapshot,
    input: Input = {} as Input,
    inputWasBound = false,
  ) {
    this.name = name;
    this.#service = service;
    this.#snapshot = snapshot;
    this.#inputWasBound = inputWasBound;
    this.input = Object.freeze({ ...input });
    const issues = service.validate(name, this.input);
    this[MOVE_ACTION_BRAND] = inputWasBound || issues.length === 0;
    this.#inputReason = issues.length
      ? { code: 'invalid-input', message: issues.map(issue => issue.message).join('; '), issues }
      : null;
    this.#preview = Object.keys(this.input).length === 0 && snapshot.baselineLegalityApplies
      ? { kind: 'not-needed' }
      : { kind: 'unchecked' };
    this.propose = this.propose.bind(this);
    this.activate = this.activate.bind(this);
    this.ensurePreview = this.ensurePreview.bind(this);
    this.refreshPreview = this.refreshPreview.bind(this);
    this.targets = this.targets.bind(this);
  }

  get availability(): MoveActionAvailability {
    // A bound action represents a different concrete move from the server's
    // default form. Its exact preview is the legality authority; applying the
    // default form's booleans here can permanently block a legal non-default
    // binding. Unbound builders and zero-input actions still use the baseline.
    if (this.#inputWasBound) return { kind: 'available' };
    const legality = this.#snapshot.currentLegality();
    if (!legality?.legalForAnyone) {
      return unavailable(
        'move-not-possible',
        legality?.error ?? `${this.name} is not possible right now`,
        legality?.preconditions,
      );
    }
    if (this.#snapshot.baselineLegalityApplies && !legality.legalForPlayer) {
      return unavailable(
        'not-legal-for-player',
        legality.error ?? `${this.name} is not legal for this player`,
        legality.preconditions,
      );
    }
    return { kind: 'available' };
  }

  get preview(): MoveActionPreview {
    return this.#preview;
  }

  get submission(): MoveActionSubmission {
    return this.#submission;
  }

  get reason(): MoveActionReason | null {
    if (this.availability.kind === 'unavailable') return this.availability.reason;
    if (this.#snapshot.currentSnapshotKey() !== this.#snapshot.snapshotKey) {
      return { code: 'stale-snapshot', message: 'This action belongs to an older game snapshot' };
    }
    if (this.#service.currentGate().isConsumed(this.#snapshot.snapshotKey)) {
      return { code: 'snapshot-consumed', message: 'Waiting for the accepted move to update the game' };
    }
    if (this.#snapshot.currentAnimating()) {
      return { code: 'animation-running', message: 'Wait for the current animation to finish' };
    }
    const clientFingerprint = this.#service.currentClientSchemaFingerprint();
    const serverFingerprint = this.#service.currentServerSchemaFingerprint();
    if (!clientFingerprint || !serverFingerprint || serverFingerprint !== clientFingerprint) {
      return { code: 'schema-mismatch', message: 'Refresh: this renderer does not match the server move schema' };
    }
    if (this.#inputReason) return this.#inputReason;
    if (!this.#service.currentTransport()) {
      return { code: 'transport-unavailable', message: 'Move submission is not connected' };
    }
    if (this.#submission.kind === 'pending') {
      return { code: 'submission-pending', message: 'Move submission is pending' };
    }
    if (this.#service.currentGate().pendingRequestID !== null) {
      return { code: 'another-submission-pending', message: 'Another move submission is pending' };
    }
    if (this.preview.kind === 'unchecked' || this.preview.kind === 'checking') {
      return { code: 'preview-unchecked', message: 'Move legality is still being checked' };
    }
    if (this.preview.kind === 'illegal' || this.preview.kind === 'failed') return this.preview.reason;
    if (this.#submission.kind === 'rejected' && !this.#submission.retryable) {
      return { code: 'submission-rejected', message: this.#submission.reason };
    }
    return null;
  }

  get canPropose(): boolean {
    return this.reason === null;
  }

  get canActivate(): boolean {
    return this.canPropose
      || (this.reason?.code === 'preview-failed'
        && this.#preview.kind === 'failed'
        && this.#preview.retryable);
  }

  with<Actual extends Input>(input: ExactInput<Input, Actual>): BoundMoveAction<MoveName, Input> {
    return cachedAction<MoveName, Input>(this.name, this.#service, this.#snapshot, input, true);
  }

  targets<Key extends TargetKey, Actual extends Input>(
    keys: readonly Key[],
    inputFor: (key: Key, index: number) => ExactInput<Input, Actual>,
    options?: TargetActionOptions,
  ): TargetAction<Key, MoveName, Input> {
    if (this.#inputWasBound) {
      throw new Error('Call .targets() directly on this.move(name), before .with(input)');
    }
    return createTargetAction(this.name, keys, inputFor, options, {
      service: this.#service,
      snapshot: this.#snapshot,
      makeAction: input => new MoveActionImplementation(
        this.name, this.#service, this.#snapshot, input, true,
      ),
      manage: (action, refresh, changed) => {
        const internal = action as MoveActionImplementation<MoveName, Input>;
        internal[MANAGED_PREVIEW](refresh, changed);
      },
      hydrate: (action, preview) => {
        const internal = action as MoveActionImplementation<MoveName, Input>;
        internal[HYDRATE_PREVIEW](preview);
      },
      notify: (action, notifyManaged) => {
        const internal = action as MoveActionImplementation<MoveName, Input>;
        internal.notifyLiveStateChanged(notifyManaged);
      },
    });
  }

  subscribe(listener: () => void): () => void {
    this.#listeners.add(listener);
    this.#gateUnsubscribe ??= this.#service.currentGate().subscribe(() => this.notifyListeners());
    void this.ensurePreview();
    return () => {
      this.#listeners.delete(listener);
      if (this.#listeners.size === 0) {
        this.#gateUnsubscribe?.();
        this.#gateUnsubscribe = null;
      }
    };
  }

  ensurePreview(): Promise<MoveActionPreview> {
    if (this.#preview.kind === 'not-needed'
      || this.#preview.kind === 'legal'
      || this.#preview.kind === 'illegal'
      || this.#preview.kind === 'failed') {
      return Promise.resolve(this.#preview);
    }
    if (this.#previewPromise) return this.#previewPromise;
    if (this.#managedPreview) {
      return this.#managedPreview(false).then(() => this.#preview);
    }
    return this.startPreview();
  }

  refreshPreview(): Promise<MoveActionPreview> {
    if (this.#preview.kind === 'not-needed') return Promise.resolve(this.#preview);
    if (this.#managedPreview) {
      return this.#managedPreview(true).then(() => this.#preview);
    }
    this.#previewAbort?.abort();
    this.#previewPromise = null;
    this.#preview = { kind: 'unchecked' };
    return this.startPreview();
  }

  [CANCEL_PREVIEW](): void {
    this.#previewAbort?.abort();
    this.#previewAbort = null;
    this.#previewPromise = null;
  }

  [MANAGED_PREVIEW](preview: (force: boolean) => Promise<void>, changed: () => void): void {
    this.#managedPreview = preview;
    this.#managedChanged = changed;
    this.#previewAbort?.abort();
    this.#previewAbort = null;
    this.#previewPromise = null;
  }

  [HYDRATE_PREVIEW](preview: MoveActionPreview): void {
    this.#preview = preview;
    this.notifyListeners();
  }

  async activate(): Promise<MoveProposalResult> {
    if (this.reason?.code === 'preview-failed'
      && this.#preview.kind === 'failed'
      && this.#preview.retryable) {
      await this.refreshPreview();
    }
    return this.propose();
  }

  private startPreview(): Promise<MoveActionPreview> {
    const requestID = this.#service.nextRequestID();
    if (this.#snapshot.currentSnapshotKey() !== this.#snapshot.snapshotKey) {
      return Promise.resolve(this.setPreview({
        kind: 'failed', retryable: false,
        reason: { code: 'stale-snapshot', message: 'This action belongs to an older game snapshot' },
      }));
    }
    const transport = this.#service.currentPreviewTransport();
    if (!transport) {
      return Promise.resolve(this.setPreview({
        kind: 'failed', retryable: false,
        reason: { code: 'preview-failed', message: 'Move preview is not connected' },
      }));
    }
    let wireArguments: Readonly<Record<string, string>>;
    try {
      wireArguments = this.#service.serialize(this.name, this.input);
    } catch (error) {
      return Promise.resolve(this.setPreview({
        kind: 'failed', retryable: false, reason: inputFailureReason(error),
      }));
    }
    const controller = new AbortController();
    this.#previewAbort = controller;
    this.#preview = { kind: 'checking', requestID };
    this.changed();
    const request: MovePreviewRequest<MoveName> = {
      requestID,
      snapshotVersion: this.#snapshot.snapshotVersion,
      viewingAsPlayer: this.#snapshot.viewingAsPlayer,
      proposingAsPlayer: this.#snapshot.proposingAsPlayer,
      proposingAsAdmin: this.#snapshot.proposingAsAdmin,
      name: this.name,
      arguments: wireArguments,
      candidateKey: canonicalInput(wireArguments),
      signal: controller.signal,
    };
    let previewCall: Promise<MovePreviewTransportResult>;
    try {
      previewCall = Promise.resolve(transport.preview(request));
    } catch (error) {
      previewCall = Promise.reject(error);
    }
    const operation = previewCall.then(result => {
      if (controller.signal.aborted
        || this.#snapshot.currentSnapshotKey() !== this.#snapshot.snapshotKey) {
        return this.#preview;
      }
      if (result.kind === 'success') {
        this.#preview = result.legal
          ? (result.preconditions
            ? { kind: 'legal', preconditions: result.preconditions }
            : { kind: 'legal' })
          : {
            kind: 'illegal',
            reason: {
              code: 'preview-illegal',
              message: result.error ?? `${this.name} is not legal for these values`,
              ...(result.preconditions ? { preconditions: result.preconditions } : {}),
            },
          };
      } else if (result.kind === 'stale-snapshot') {
        this.#preview = {
          kind: 'failed', retryable: false,
          reason: { code: 'stale-snapshot', message: 'The game changed while checking this move' },
        };
      } else {
        this.#preview = {
          kind: 'failed', retryable: result.retryable,
          reason: { code: 'preview-failed', message: result.friendlyError ?? result.error },
        };
      }
      return this.#preview;
    }).catch(error => {
      if (!controller.signal.aborted) {
        this.#preview = {
          kind: 'failed', retryable: true,
          reason: {
            code: 'preview-failed',
            message: error instanceof Error ? error.message : 'Move preview failed',
          },
        };
      }
      return this.#preview;
    }).finally(() => {
      if (this.#previewPromise === operation) this.#previewPromise = null;
      if (this.#previewAbort === controller) this.#previewAbort = null;
      this.changed();
    });
    this.#previewPromise = operation;
    return operation;
  }

  private setPreview(preview: MoveActionPreview): MoveActionPreview {
    this.#preview = preview;
    this.changed();
    return preview;
  }

  async propose(): Promise<MoveProposalResult> {
    const requestID = this.#service.nextRequestID();
    const currentVersion = this.#snapshot.currentSnapshotVersion();
    if (this.#snapshot.currentSnapshotKey() !== this.#snapshot.snapshotKey) {
      return this.report({
        kind: 'stale-snapshot', requestID,
        expectedVersion: this.#snapshot.snapshotVersion, actualVersion: currentVersion,
      });
    }
    if (this.#preview.kind === 'unchecked' || this.#preview.kind === 'checking') {
      await this.ensurePreview();
    }
    const reason = this.reason;
    if (reason) return this.report({ kind: 'blocked', requestID, reason });
    const gate = this.#service.currentGate();
    const acquisition = gate.acquire(requestID, this.#snapshot.snapshotKey);
    if (acquisition !== 'acquired') {
      return this.report({
        kind: 'blocked', requestID,
        reason: acquisition === 'busy'
          ? { code: 'another-submission-pending', message: 'Another move submission is pending' }
          : { code: 'snapshot-consumed', message: 'Waiting for the accepted move to update the game' },
      });
    }

    let wireArguments: Readonly<Record<string, string>>;
    try {
      wireArguments = this.#service.serialize(this.name, this.input);
    } catch (error) {
      gate.release(requestID);
      const reason = inputFailureReason(error);
      return this.report({ kind: 'blocked', requestID, reason });
    }

    this.#submission = { kind: 'pending', requestID };
    this.changed();
    try {
      const transport = this.#service.currentTransport();
      if (!transport) {
        return this.report({
          kind: 'blocked', requestID,
          reason: { code: 'transport-unavailable', message: 'Move submission is not connected' },
        });
      }
      const result = await transport.submit({
        requestID,
        snapshotVersion: this.#snapshot.snapshotVersion,
        viewingAsPlayer: this.#snapshot.viewingAsPlayer,
        proposingAsPlayer: this.#snapshot.proposingAsPlayer,
        proposingAsAdmin: this.#snapshot.proposingAsAdmin,
        name: this.name,
        arguments: wireArguments,
      });
      if (result.kind === 'success') {
        this.#submission = { kind: 'idle' };
        gate.consume(this.#snapshot.snapshotKey);
        return this.report({ kind: 'success', requestID });
      }
      if (result.kind === 'server-rejection' && result.code === 'STALE_SNAPSHOT') {
        this.#submission = {
          kind: 'rejected', requestID, source: 'server',
          reason: result.friendlyError ?? result.error, retryable: false,
        };
        gate.consume(this.#snapshot.snapshotKey);
        return this.report({
          kind: 'stale-snapshot',
          requestID,
          expectedVersion: result.expectedVersion ?? this.#snapshot.snapshotVersion,
          actualVersion: result.actualVersion ?? this.#snapshot.currentSnapshotVersion(),
        });
      }
      if (result.kind === 'server-rejection'
        && (result.code === 'CLIENT_SUBMISSION_BUSY' || result.code === 'CLIENT_SNAPSHOT_CONSUMED')) {
        const reason: MoveActionReason = result.code === 'CLIENT_SUBMISSION_BUSY'
          ? { code: 'another-submission-pending', message: result.friendlyError ?? result.error }
          : { code: 'snapshot-consumed', message: result.friendlyError ?? result.error };
        if (result.code === 'CLIENT_SNAPSHOT_CONSUMED') {
          gate.consume(this.#snapshot.snapshotKey);
        }
        this.#submission = { kind: 'idle' };
        return this.report({ kind: 'blocked', requestID, reason });
      }
      this.#submission = {
        kind: 'rejected', requestID,
        source: result.kind === 'server-rejection' ? 'server' : 'network',
        reason: result.friendlyError ?? result.error,
        retryable: result.kind === 'network-failure',
      };
      return this.report({ ...result, requestID });
    } catch (error) {
      const result: MoveProposalResult = {
        kind: 'network-failure', requestID,
        error: error instanceof Error ? error.message : 'Move transport failed',
        friendlyError: 'Unable to submit the move',
      };
      this.#submission = {
        kind: 'rejected', requestID, source: 'network',
        reason: result.friendlyError ?? result.error, retryable: true,
      };
      return this.report(result);
    } finally {
      gate.release(requestID);
      this.changed();
    }
  }

  private report(result: MoveProposalResult): MoveProposalResult {
    this.#service.telemetry?.({ name: this.name, requestID: result.requestID, result });
    return result;
  }

  private changed(): void {
    this.#managedChanged?.();
    this.notifyListeners();
    if (!this.#managedChanged) this.#service.changed?.();
  }

  private notifyListeners(): void {
    for (const listener of this.#listeners) notifyListener(listener);
  }

  notifyLiveStateChanged(notifyManaged = true): void {
    if (notifyManaged) this.#managedChanged?.();
    this.notifyListeners();
  }
}

function cachedAction<MoveName extends string, Input extends object>(
  name: MoveName,
  service: MoveActionService,
  snapshot: MoveActionSnapshot,
  input: Input,
  inputWasBound = false,
  cacheNamespace = '',
): BoundMoveAction<MoveName, Input> {
  const key = `${snapshot.actionCacheKey ?? snapshot.snapshotKey}\u0000${cacheNamespace}\u0000${name}\u0000${canonicalInput(input)}`;
  const existing = service.actionCache?.get(key);
  if (existing) return existing as BoundMoveAction<MoveName, Input>;
  const action = new MoveActionImplementation<MoveName, Input>(
    name, service, snapshot, input, inputWasBound,
  );
  service.actionCache?.set(key, action as BoundMoveAction<string, object>);
  return action;
}

function canonicalInput(input: object): string {
  return JSON.stringify(Object.entries(input).sort(([left], [right]) => left.localeCompare(right)));
}

function unavailable(
  code: MoveActionReasonCode,
  message: string,
  preconditions?: readonly MoveActionPrecondition[],
): MoveActionAvailability {
  return {
    kind: 'unavailable',
    reason: preconditions ? { code, message, preconditions } : { code, message },
  };
}

function inputFailureReason(error: unknown): MoveActionReason {
  if (isErrorWithCode(error, 'BOARDGAME_STALE_MOVE_INPUT_SCHEMA')) {
    return { code: 'schema-mismatch', message: error.message };
  }
  return {
    code: 'invalid-input',
    message: error instanceof Error ? error.message : 'Move input could not be serialized',
    ...(isInputValidationError(error) ? { issues: error.errors } : {}),
  };
}

function isErrorWithCode(error: unknown, code: string): error is Error & { readonly code: string } {
  return error instanceof Error && 'code' in error && error.code === code;
}

function isInputValidationError(
  error: unknown,
): error is Error & { readonly errors: readonly MoveInputIssue[] } {
  return error instanceof Error && 'errors' in error && Array.isArray(error.errors);
}

function notifyListener(listener: () => void): void {
  try {
    listener();
  } catch (error) {
    const report = (globalThis as { reportError?: (value: unknown) => void }).reportError;
    try {
      if (report) report(error);
      else console.error('Move action subscriber failed', error);
    } catch {
      // Diagnostic hooks are userland too; they must never wedge the gate.
    }
  }
}
