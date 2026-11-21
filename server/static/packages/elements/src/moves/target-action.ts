import type {
  BoundMoveAction,
  MoveActionPreview,
  MoveActionReason,
  MoveActionService,
  MoveActionSnapshot,
} from './action.js';

export type TargetKey = string | number;
export const MAX_TARGET_ACTION_CANDIDATES = 1024;
const MAX_CACHED_TARGET_ACTIONS = 32;

export interface TargetActionOptions {
  /** Permit distinct UI targets that intentionally submit identical move input. */
  readonly allowDuplicateInputs?: boolean;
}

export interface TargetPreviewCandidate {
  readonly id: string;
  readonly arguments: Readonly<Record<string, string>>;
}

export interface TargetPreviewRequest<MoveName extends string = string> {
  readonly requestID: string;
  readonly snapshotVersion: number;
  readonly proposingAsPlayer: number;
  readonly proposingAsAdmin: boolean;
  readonly name: MoveName;
  readonly candidates: readonly TargetPreviewCandidate[];
  readonly signal: AbortSignal;
}

export interface TargetPreviewResult {
  readonly id: string;
  readonly legal: boolean;
  readonly error?: string;
}

export type TargetPreviewTransportResult =
  | { readonly kind: 'success'; readonly results: readonly TargetPreviewResult[] }
  | { readonly kind: 'stale-snapshot'; readonly expectedVersion: number; readonly actualVersion: number }
  | { readonly kind: 'failure'; readonly error: string; readonly friendlyError?: string; readonly retryable: boolean };

export interface TargetPreviewTransport {
  previewTargets(request: TargetPreviewRequest): Promise<TargetPreviewTransportResult>;
}

export type TargetActionPreview =
  | { readonly kind: 'unchecked' }
  | { readonly kind: 'checking'; readonly requestID: string }
  | { readonly kind: 'ready' }
  | { readonly kind: 'failed'; readonly reason: MoveActionReason; readonly retryable: boolean };

export interface TargetCandidate<
  Key extends TargetKey,
  MoveName extends string = string,
  Input extends object = object,
> {
  readonly key: Key;
  readonly action: BoundMoveAction<MoveName, Input>;
}

export interface TargetAction<
  Key extends TargetKey,
  MoveName extends string = string,
  Input extends object = object,
> {
  readonly candidates: readonly TargetCandidate<Key, MoveName, Input>[];
  readonly preview: TargetActionPreview;
  get(key: Key): TargetCandidate<Key, MoveName, Input> | undefined;
  ensurePreview(): Promise<TargetActionPreview>;
  refreshPreview(): Promise<TargetActionPreview>;
  subscribe(listener: () => void): () => void;
}

export interface TargetActionHost<Input extends object, MoveName extends string> {
  readonly service: MoveActionService;
  readonly snapshot: MoveActionSnapshot;
  readonly makeAction: (input: Input) => BoundMoveAction<MoveName, Input>;
  readonly manage: (
    action: BoundMoveAction<MoveName, Input>,
    preview: (force: boolean) => Promise<void>,
    changed: () => void,
  ) => void;
  readonly hydrate: (action: BoundMoveAction<MoveName, Input>, preview: MoveActionPreview) => void;
  readonly notify: (action: BoundMoveAction<MoveName, Input>, notifyManaged: boolean) => void;
}

interface TargetRecord<Key extends TargetKey, MoveName extends string, Input extends object>
  extends TargetCandidate<Key, MoveName, Input> {
  readonly id: string;
  readonly wire: Readonly<Record<string, string>>;
}

export function createTargetAction<
  Key extends TargetKey,
  MoveName extends string,
  Input extends object,
  Actual extends Input,
>(
  name: MoveName,
  keys: readonly Key[],
  inputFor: (key: Key, index: number) => Actual & Record<Exclude<keyof Actual, keyof Input>, never>,
  options: TargetActionOptions | undefined,
  host: TargetActionHost<Input, MoveName>,
): TargetAction<Key, MoveName, Input> {
  if (keys.length > MAX_TARGET_ACTION_CANDIDATES) {
    throw new Error(`Target collection has ${keys.length} candidates; maximum is ${MAX_TARGET_ACTION_CANDIDATES}`);
  }
  const copiedKeys = Object.freeze([...keys]);
  const seenKeys = new Map<TargetKey, number>();
  const seenArguments = new Map<string, number>();
  const mapped = copiedKeys.map((key, index) => {
    if (typeof key !== 'string' && typeof key !== 'number') {
      throw new Error(`Target key at index ${index} must be a string or number`);
    }
    if (typeof key === 'number' && !Number.isFinite(key)) {
      throw new Error(`Target key at index ${index} must be finite`);
    }
    const duplicateKey = seenKeys.get(key);
    if (duplicateKey !== undefined) {
      throw new Error(`Duplicate target key ${JSON.stringify(key)} at indexes ${duplicateKey} and ${index}`);
    }
    seenKeys.set(key, index);
    let input: Actual;
    try {
      input = inputFor(key, index);
    } catch (error) {
      const detail = error instanceof Error ? `: ${error.message}` : '';
      throw new Error(`Target input mapper failed for ${JSON.stringify(key)} at index ${index}${detail}`);
    }
    const issues = host.service.validate(name, input);
    if (issues.length) {
      throw new Error(`Invalid target input for ${JSON.stringify(key)} at index ${index}: ${issues.map(issue => issue.message).join('; ')}`);
    }
    let wire: Readonly<Record<string, string>>;
    try {
      wire = host.service.serialize(name, input);
    } catch (error) {
      const detail = error instanceof Error ? `: ${error.message}` : '';
      throw new Error(`Target input serialization failed for ${JSON.stringify(key)} at index ${index}${detail}`);
    }
    const canonical = canonicalArguments(wire);
    const duplicateArguments = seenArguments.get(canonical);
    if (duplicateArguments !== undefined && !options?.allowDuplicateInputs) {
      throw new Error(`Targets at indexes ${duplicateArguments} and ${index} map to identical move arguments`);
    }
    seenArguments.set(canonical, index);
    return Object.freeze({ key, id: targetID(key), input: Object.freeze({ ...input }) as Input, wire });
  });
  const cacheKey = `${host.snapshot.snapshotKey}\u0000${name}\u0000${mapped.map(item => `${item.id}=${canonicalArguments(item.wire)}`).join('|')}`;
  const cached = host.service.targetActionCache?.get(cacheKey);
  if (cached) return cached as TargetAction<Key, MoveName, Input>;

  const records: readonly TargetRecord<Key, MoveName, Input>[] = Object.freeze(mapped.map(item => Object.freeze({
    key: item.key,
    id: item.id,
    wire: item.wire,
    action: host.makeAction(item.input),
  })));
  const result = new TargetActionImplementation(name, records, host);
  const cache = host.service.targetActionCache;
  cache?.set(cacheKey, result as TargetAction<TargetKey, string, object>);
  if (cache && cache.size > MAX_CACHED_TARGET_ACTIONS) {
    const oldest = cache.keys().next().value;
    if (oldest !== undefined) cache.delete(oldest);
  }
  return result;
}

/** Host lifecycle hook; intentionally absent from the creator-facing interface. */
export function cancelTargetActionPreview(action: TargetAction<TargetKey>): void {
  if (action instanceof TargetActionImplementation) action.cancel();
}

/** Host lifecycle hook for live action dependencies outside snapshot identity. */
export function notifyTargetActionLiveStateChanged(action: TargetAction<TargetKey>): void {
  if (action instanceof TargetActionImplementation) action.notifyLiveStateChanged();
}

class TargetActionImplementation<Key extends TargetKey, MoveName extends string, Input extends object>
implements TargetAction<Key, MoveName, Input> {
  readonly candidates: readonly TargetCandidate<Key, MoveName, Input>[];
  readonly #records: readonly TargetRecord<Key, MoveName, Input>[];
  readonly #byKey = new Map<Key, TargetCandidate<Key, MoveName, Input>>();
  readonly #name: MoveName;
  readonly #host: TargetActionHost<Input, MoveName>;
  readonly #listeners = new Set<() => void>();
  #gateUnsubscribe: (() => void) | null = null;
  #preview: TargetActionPreview = { kind: 'unchecked' };
  #operation: Promise<TargetActionPreview> | null = null;
  #abort: AbortController | null = null;

  constructor(name: MoveName, records: readonly TargetRecord<Key, MoveName, Input>[], host: TargetActionHost<Input, MoveName>) {
    this.#name = name;
    this.#records = records;
    this.candidates = Object.freeze(records.map(({ key, action }) => Object.freeze({ key, action })));
    this.#host = host;
    for (const candidate of this.candidates) this.#byKey.set(candidate.key, candidate);
    if (records.length === 0) this.#preview = { kind: 'ready' };
    for (const record of records) {
      host.manage(record.action, async force => {
        if (force) await this.refreshPreview();
        else await this.ensurePreview();
      }, () => this.changed());
    }
  }

  get preview(): TargetActionPreview { return this.#preview; }
  get(key: Key): TargetCandidate<Key, MoveName, Input> | undefined { return this.#byKey.get(key); }

  subscribe(listener: () => void): () => void {
    this.#listeners.add(listener);
    if (this.#listeners.size === 1) {
      this.#gateUnsubscribe = this.#host.service.currentGate().subscribe(() => this.changed());
    }
    void this.ensurePreview();
    return () => {
      this.#listeners.delete(listener);
      if (this.#listeners.size === 0) {
        this.#gateUnsubscribe?.();
        this.#gateUnsubscribe = null;
      }
    };
  }

  ensurePreview(): Promise<TargetActionPreview> {
    if (this.#preview.kind === 'ready' || this.#preview.kind === 'failed') return Promise.resolve(this.#preview);
    return this.startPreview();
  }

  refreshPreview(): Promise<TargetActionPreview> {
    this.#abort?.abort();
    this.#operation = null;
    this.#preview = this.#records.length ? { kind: 'unchecked' } : { kind: 'ready' };
    return this.startPreview();
  }

  cancel(): void {
    this.#abort?.abort();
    this.#abort = null;
    this.#operation = null;
  }

  notifyLiveStateChanged(): void {
    for (const record of this.#records) {
      // Wake controls subscribed directly to each candidate, but suppress the
      // candidate's managed callback: that would notify this coordinator once
      // per candidate. One coalesced target notification below is sufficient.
      this.#host.notify(record.action, false);
    }
    this.changed();
  }

  private startPreview(): Promise<TargetActionPreview> {
    if (this.#operation) return this.#operation;
    if (!this.#records.length) return Promise.resolve(this.#preview);
    if (this.#host.snapshot.currentSnapshotKey() !== this.#host.snapshot.snapshotKey) {
      return Promise.resolve(this.fail('stale-snapshot', 'This target collection belongs to an older game snapshot', false));
    }
    const transport = this.#host.service.currentTargetPreviewTransport();
    if (!transport) return Promise.resolve(this.fail('preview-failed', 'Target preview is not connected', false));
    const requestID = this.#host.service.nextRequestID();
    const controller = new AbortController();
    this.#abort = controller;
    this.#preview = { kind: 'checking', requestID };
    for (const record of this.#records) this.#host.hydrate(record.action, { kind: 'checking', requestID });
    this.changed();
    let transportCall: Promise<TargetPreviewTransportResult>;
    try {
      transportCall = Promise.resolve(transport.previewTargets({
        requestID,
        snapshotVersion: this.#host.snapshot.snapshotVersion,
        proposingAsPlayer: this.#host.snapshot.proposingAsPlayer,
        proposingAsAdmin: this.#host.snapshot.proposingAsAdmin,
        name: this.#name,
        candidates: this.#records.map(record => ({ id: record.id, arguments: record.wire })),
        signal: controller.signal,
      }));
    } catch (error) {
      transportCall = Promise.reject(error);
    }
    const operation = transportCall.then(response => {
      if (controller.signal.aborted || this.#host.snapshot.currentSnapshotKey() !== this.#host.snapshot.snapshotKey) return this.#preview;
      if (response.kind === 'stale-snapshot') return this.fail('stale-snapshot', 'The game changed while checking these targets', false);
      if (response.kind === 'failure') return this.fail('preview-failed', response.friendlyError ?? response.error, response.retryable);
      if (!Array.isArray(response.results)) return this.fail('preview-failed', 'Target preview results must be an array', false);
      const correlated = correlateResults(response.results, this.#records);
      if (correlated instanceof Error) return this.fail('preview-failed', correlated.message, false);
      for (const record of this.#records) {
        const item = correlated.get(record.id);
        if (!item) return this.fail('preview-failed', `Target preview omitted candidate ${record.id}`, false);
        this.#host.hydrate(record.action, item.legal ? { kind: 'legal' } : {
          kind: 'illegal',
          reason: { code: 'preview-illegal', message: item.error ?? `${this.#name} is not legal for this target` },
        });
      }
      this.#preview = { kind: 'ready' };
      this.changed();
      return this.#preview;
    }).catch(error => {
      if (controller.signal.aborted) return this.#preview;
      return this.fail('preview-failed', error instanceof Error ? error.message : 'Target preview failed', true);
    }).finally(() => {
      if (this.#operation === operation) this.#operation = null;
      if (this.#abort === controller) this.#abort = null;
    });
    this.#operation = operation;
    return operation;
  }

  private fail(code: 'preview-failed' | 'stale-snapshot', message: string, retryable: boolean): TargetActionPreview {
    const reason: MoveActionReason = { code, message };
    this.#preview = { kind: 'failed', reason, retryable };
    for (const record of this.#records) {
      this.#host.hydrate(record.action, { kind: 'failed', reason, retryable });
    }
    this.changed();
    return this.#preview;
  }

  private changed(): void {
    for (const listener of this.#listeners) {
      try { listener(); } catch (error) { safeDiagnostic('Target action subscriber failed', error); }
    }
    this.#host.service.changed?.();
  }
}

function correlateResults<Key extends TargetKey, MoveName extends string, Input extends object>(
  results: readonly unknown[],
  records: readonly TargetRecord<Key, MoveName, Input>[],
): ReadonlyMap<string, TargetPreviewResult> | Error {
  const expected = new Set(records.map(record => record.id));
  const correlated = new Map<string, TargetPreviewResult>();
  for (const value of results) {
    if (!isTargetPreviewResult(value)) return new Error('Target preview returned a malformed result');
    const result = value;
    if (!expected.has(result.id)) return new Error(`Target preview returned unknown candidate ${JSON.stringify(result.id)}`);
    if (correlated.has(result.id)) return new Error(`Target preview returned duplicate candidate ${JSON.stringify(result.id)}`);
    correlated.set(result.id, result);
  }
  if (correlated.size !== expected.size) {
    return new Error(`Target preview returned ${correlated.size} results for ${expected.size} candidates`);
  }
  return correlated;
}

function isTargetPreviewResult(value: unknown): value is TargetPreviewResult {
  if (typeof value !== 'object' || value === null) return false;
  const result = value as Partial<TargetPreviewResult>;
  return typeof result.id === 'string'
    && result.id.length > 0
    && typeof result.legal === 'boolean'
    && (result.error === undefined || typeof result.error === 'string');
}

function targetID(key: TargetKey): string {
  return JSON.stringify([typeof key, key]);
}

function canonicalArguments(arguments_: Readonly<Record<string, string>>): string {
  return JSON.stringify(Object.entries(arguments_).sort(([left], [right]) => left.localeCompare(right)));
}

function safeDiagnostic(message: string, error: unknown): void {
  try { console.error(message, error); } catch { /* Diagnostics must not break state propagation. */ }
}
