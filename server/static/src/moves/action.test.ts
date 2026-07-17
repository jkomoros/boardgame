import assert from 'node:assert/strict';
import test from 'node:test';
import {
  MoveSubmissionGate,
  cancelMoveActionPreview,
  createMoveAction,
  notifyMoveActionLiveStateChanged,
  type MoveActionFor,
  type MoveActionLegality,
  type MoveActionService,
  type MoveActionSnapshot,
  type MovePreviewRequest,
  type MovePreviewTransport,
  type MoveSubmissionRequest,
  type MoveTransportResult,
} from './action.ts';
import { serializeCreatorMoveInput, validateCreatorMoveInput } from './input.ts';

type Names = 'Roll' | 'Place';
type Inputs = { Roll: Record<string, never>; Place: { Slot: number } };

const schema = [
  { name: 'Roll', fields: [] },
  { name: 'Place', fields: [{ name: 'Slot', wireType: 'int', disposition: 'required', codec: 'integer' }] },
] as const;

interface ContextOverrides {
  serverSchemaFingerprint?: string | null;
  currentSnapshotKey?: () => string;
  currentSnapshotVersion?: () => number;
  legality?: MoveActionLegality;
  animating?: boolean;
  currentAnimating?: () => boolean;
  transport?: MoveActionService['currentTransport'] extends () => infer T ? T : never;
  previewTransport?: MovePreviewTransport | null;
  gate?: MoveSubmissionGate;
  baselineLegalityApplies?: boolean;
}

function context(overrides: ContextOverrides = {}): {
  service: MoveActionService;
  snapshot: MoveActionSnapshot;
} {
  const gate = overrides.gate ?? new MoveSubmissionGate();
  const service: MoveActionService = {
    currentClientSchemaFingerprint: () => 'sha256:client',
    currentServerSchemaFingerprint: () => overrides.serverSchemaFingerprint === undefined
      ? 'sha256:client'
      : overrides.serverSchemaFingerprint,
    currentTransport: () => overrides.transport ?? { submit: async () => ({ kind: 'success' }) },
    currentPreviewTransport: () => overrides.previewTransport === undefined
      ? { preview: async () => ({ kind: 'success', legal: true }) }
      : overrides.previewTransport,
    currentTargetPreviewTransport: () => null,
    currentGate: () => gate,
    nextRequestID: sequence(),
    validate: (name, input) => validateCreatorMoveInput(schema, name, input).errors,
    serialize: (name, input) => serializeCreatorMoveInput(schema, name, input),
    actionCache: new Map(),
  };
  const snapshot: MoveActionSnapshot = {
    snapshotKey: 'game:g1:epoch:1:version:4:viewer:0:admin:0:schema:client',
    currentSnapshotKey: overrides.currentSnapshotKey
      ?? (() => 'game:g1:epoch:1:version:4:viewer:0:admin:0:schema:client'),
    snapshotVersion: 4,
    currentSnapshotVersion: overrides.currentSnapshotVersion ?? (() => 4),
    viewingAsPlayer: 0,
    proposingAsPlayer: 0,
    proposingAsAdmin: false,
    currentLegality: () => overrides.legality ?? { legalForPlayer: true, legalForAnyone: true },
    currentAnimating: overrides.currentAnimating ?? (() => overrides.animating ?? false),
    baselineLegalityApplies: overrides.baselineLegalityApplies ?? true,
  };
  return { service, snapshot };
}

function action<K extends Names>(
  name: K,
  value = context(),
): MoveActionFor<K, Inputs[K]> {
  return createMoveAction<K, Names, Inputs>(name, value.service, value.snapshot);
}

test('zero-input actions propose successfully and keep bound methods', async () => {
  const requests: MoveSubmissionRequest[] = [];
  const move = action('Roll', context({
    transport: { submit: async request => { requests.push(request); return { kind: 'success' }; } },
  }));
  const propose = move.propose;
  assert.deepEqual(await propose(), { kind: 'success', requestID: 'request-1' });
  assert.deepEqual(requests, [{
    requestID: 'request-1', snapshotVersion: 4, viewingAsPlayer: 0, name: 'Roll', arguments: {},
    proposingAsPlayer: 0, proposingAsAdmin: false,
  }]);
  assert.equal(move.submission.kind, 'idle');
});

test('host live-state invalidation notifies a cached action without replacing it', async () => {
  let animating = true;
  const actionContext = context({ currentAnimating: () => animating });
  const move = action('Roll', actionContext);
  const cached = action('Roll', actionContext);
  assert.equal(cached, move);
  assert.equal(move.reason?.code, 'animation-running');
  let notifications = 0;
  const unsubscribe = move.subscribe(() => { notifications++; });
  animating = false;
  notifyMoveActionLiveStateChanged(move);
  assert.equal(move.reason, null);
  assert.equal(notifications, 1);
  unsubscribe();
});

test('with(args) is immutable and invalid native input fails safely', async () => {
  const actionContext = context();
  const builder = action('Place', actionContext);
  const source = { Slot: 3 };
  const bound = builder.with(source);
  source.Slot = 8;
  assert.deepEqual(bound.input, { Slot: 3 });
  assert.equal(builder.with({ Slot: 3 }), bound);
  const invalid = builder.with({ Slot: 1.5 });
  assert.equal(invalid.canPropose, false);
  assert.equal(invalid.reason?.code, 'invalid-input');
  const result = await invalid.propose();
  assert.equal(result.kind, 'blocked');
  if (result.kind === 'blocked') {
    assert.equal(result.reason.code, 'invalid-input');
    assert.deepEqual(result.reason.issues?.map(issue => issue.code), ['invalid-type']);
  }
});

test('exact bound previews are idempotent, self-contained, and gate proposal', async () => {
  let finish: ((result: { kind: 'success'; legal: boolean }) => void) | undefined;
  const requests: MovePreviewRequest[] = [];
  const value = context({
    previewTransport: {
      preview: request => {
        requests.push(request);
        return new Promise(resolve => { finish = resolve; });
      },
    },
  });
  const move = action('Place', value).with({ Slot: 3 });
  assert.equal(move.preview.kind, 'unchecked');
  const first = move.ensurePreview();
  const second = move.ensurePreview();
  assert.equal(first, second);
  assert.equal(move.preview.kind, 'checking');
  assert.equal(requests.length, 1);
  assert.deepEqual({
    version: requests[0]?.snapshotVersion,
    viewer: requests[0]?.viewingAsPlayer,
    proposer: requests[0]?.proposingAsPlayer,
    admin: requests[0]?.proposingAsAdmin,
    args: requests[0]?.arguments,
  }, { version: 4, viewer: 0, proposer: 0, admin: false, args: { Slot: '3' } });
  finish?.({ kind: 'success', legal: true });
  assert.equal((await first).kind, 'legal');
  assert.equal(move.canPropose, true);
});

test('exact bound legality supersedes an illegal default-form baseline', async () => {
  const submissions: MoveSubmissionRequest[] = [];
  const legal = action('Place', context({
    legality: { legalForPlayer: false, legalForAnyone: false },
    previewTransport: { preview: async () => ({ kind: 'success', legal: true }) },
    transport: { submit: async request => { submissions.push(request); return { kind: 'success' }; } },
  })).with({ Slot: 3 });
  assert.equal(legal.reason?.code, 'preview-unchecked');
  assert.equal((await legal.ensurePreview()).kind, 'legal');
  assert.equal(legal.canActivate, true);
  assert.equal((await legal.activate()).kind, 'success');
  assert.deepEqual(submissions.map(request => request.arguments), [{ Slot: '3' }]);

  const illegal = action('Place', context({
    legality: { legalForPlayer: false, legalForAnyone: false },
    previewTransport: {
      preview: async () => ({ kind: 'success', legal: false, error: 'That slot is occupied' }),
    },
  })).with({ Slot: 4 });
  assert.equal((await illegal.ensurePreview()).kind, 'illegal');
  assert.equal(illegal.reason?.code, 'preview-illegal');
  assert.equal(illegal.canActivate, false);
  assert.equal((await illegal.propose()).kind, 'blocked');
});

test('host cancellation aborts an exact preview without submitting', async () => {
  let signal: AbortSignal | undefined;
  const value = context({
    previewTransport: {
      preview: request => {
        signal = request.signal;
        return new Promise(() => {});
      },
    },
  });
  const move = action('Place', value).with({ Slot: 3 });
  void move.ensurePreview();
  assert.equal(signal?.aborted, false);
  cancelMoveActionPreview(move);
  assert.equal(signal?.aborted, true);
});

test('UI activation recovers from a transient preview failure', async () => {
  let previewCalls = 0;
  let submissionCalls = 0;
  const value = context({
    previewTransport: {
      preview: async () => (++previewCalls === 1
        ? { kind: 'failure', error: 'offline', retryable: true }
        : { kind: 'success', legal: true }),
    },
    transport: {
      submit: async () => { submissionCalls++; return { kind: 'success' }; },
    },
  });
  const move = action('Place', value).with({ Slot: 3 });
  assert.equal((await move.ensurePreview()).kind, 'failed');
  assert.equal(move.canPropose, false);
  assert.equal(move.canActivate, true);
  assert.equal((await move.activate()).kind, 'success');
  assert.equal(previewCalls, 2);
  assert.equal(submissionCalls, 1);
});

test('UI activation stays blocked when a preview retry is still illegal', async () => {
  let previewCalls = 0;
  let submissionCalls = 0;
  const value = context({
    previewTransport: {
      preview: async () => (++previewCalls === 1
        ? { kind: 'failure', error: 'offline', retryable: true }
        : { kind: 'success', legal: false, error: 'That slot is occupied' }),
    },
    transport: {
      submit: async () => { submissionCalls++; return { kind: 'success' }; },
    },
  });
  const move = action('Place', value).with({ Slot: 3 });
  await move.ensurePreview();
  const result = await move.activate();
  assert.equal(result.kind, 'blocked');
  if (result.kind === 'blocked') assert.equal(result.reason.code, 'preview-illegal');
  assert.equal(previewCalls, 2);
  assert.equal(submissionCalls, 0);
  assert.equal(move.canActivate, false);
});

test('live animation and global gate changes immediately disable retained actions', async () => {
  let animating = false;
  let finish: ((result: MoveTransportResult) => void) | undefined;
  const gate = new MoveSubmissionGate();
  const value = context({
    gate,
    currentAnimating: () => animating,
    transport: { submit: () => new Promise(resolve => { finish = resolve; }) },
  });
  const first = action('Roll', value);
  const second = action('Place', value).with({ Slot: 2 });
  animating = true;
  assert.equal(first.reason?.code, 'animation-running');
  animating = false;
  const pending = first.propose();
  assert.equal(second.reason?.code, 'another-submission-pending');
  finish?.({ kind: 'success' });
  await pending;
});

test('throwing subscribers cannot wedge the submission gate', async () => {
  const reported: unknown[] = [];
  const original = globalThis.reportError;
  globalThis.reportError = value => { reported.push(value); };
  try {
    const move = action('Roll');
    let healthyCalls = 0;
    move.subscribe(() => { throw new Error('observer failure'); });
    move.subscribe(() => { healthyCalls++; });
    assert.equal((await move.propose()).kind, 'success');
    assert.ok(healthyCalls > 0);
    assert.ok(reported.length > 0);
  } finally {
    globalThis.reportError = original;
  }
});

test('a throwing diagnostic hook cannot wedge subscriber notification', async () => {
  const original = globalThis.reportError;
  globalThis.reportError = () => { throw new Error('broken diagnostics'); };
  try {
    const move = action('Roll');
    let healthyCalls = 0;
    move.subscribe(() => { throw new Error('observer failure'); });
    move.subscribe(() => { healthyCalls++; });
    assert.equal((await move.propose()).kind, 'success');
    assert.ok(healthyCalls > 0);
  } finally {
    globalThis.reportError = original;
  }
});

test('structural legality never enables a viewer and blockers are explicit', () => {
  const structural = action('Roll', context({
    legality: { legalForPlayer: true, legalForAnyone: false },
  }));
  assert.equal(structural.canPropose, false);
  assert.equal(structural.reason?.code, 'move-not-possible');

  const viewer = action('Roll', context({
    legality: {
      legalForPlayer: false,
      legalForAnyone: true,
      preconditions: [{ name: 'player-bool', verdict: 'fail', evaluable: true }],
    },
  }));
  assert.equal(viewer.reason?.code, 'not-legal-for-player');
  assert.equal(viewer.reason?.preconditions?.[0]?.name, 'player-bool');
  const animating = action('Roll', context({ animating: true }));
  assert.equal(animating.reason?.code, 'animation-running');
  const skewed = action('Roll', context({ serverSchemaFingerprint: null }));
  assert.equal(skewed.reason?.code, 'schema-mismatch');
});

test('retained actions and concurrent actions fail closed', async () => {
  let version = 4;
  let snapshotKey = 'game:g1:epoch:1:version:4:viewer:0:admin:0:schema:client';
  const retained = action('Roll', context({
    currentSnapshotVersion: () => version,
    currentSnapshotKey: () => snapshotKey,
  }));
  version = 5;
  snapshotKey = 'game:g1:epoch:1:version:5:viewer:0:admin:0:schema:client';
  assert.deepEqual(await retained.propose(), {
    kind: 'stale-snapshot', requestID: 'request-1', expectedVersion: 4, actualVersion: 5,
  });

  let finish: ((result: MoveTransportResult) => void) | undefined;
  const gate = new MoveSubmissionGate();
  const shared = context({
    gate,
    transport: { submit: () => new Promise(resolve => { finish = resolve; }) },
  });
  const first = action('Roll', shared);
  const second = action('Place', shared).with({ Slot: 2 });
  const firstResult = first.propose();
  assert.equal(first.submission.kind, 'pending');
  const blocked = await second.propose();
  assert.equal(blocked.kind, 'blocked');
  if (blocked.kind === 'blocked') assert.equal(blocked.reason.code, 'another-submission-pending');
  finish?.({ kind: 'success' });
  assert.equal((await firstResult).kind, 'success');
  const consumed = await second.propose();
  assert.equal(consumed.kind, 'blocked');
  if (consumed.kind === 'blocked') assert.equal(consumed.reason.code, 'snapshot-consumed');
});

test('server and network failures remain discriminated and observable', async () => {
  const results: MoveTransportResult[] = [
    { kind: 'server-rejection', error: 'illegal', friendlyError: 'Choose another move' },
    { kind: 'network-failure', error: 'offline' },
  ];
  const actionContext = context({
    transport: { submit: async () => results.shift() ?? { kind: 'success' } },
  });
  const rejected = action('Roll', actionContext);
  assert.equal((await rejected.propose()).kind, 'server-rejection');
  assert.deepEqual(rejected.submission, {
    kind: 'rejected', requestID: 'request-1', source: 'server',
    reason: 'Choose another move', retryable: false,
  });
  const network = action('Roll', context({
    transport: { submit: async () => results.shift() ?? { kind: 'success' } },
  }));
  assert.equal((await network.propose()).kind, 'network-failure');
});

test('server expected-version rejection becomes stale and consumes the snapshot', async () => {
  const actionContext = context({
    transport: {
      submit: async () => ({
        kind: 'server-rejection', code: 'STALE_SNAPSHOT',
        error: 'stale', expectedVersion: 4, actualVersion: 5,
      }),
    },
  });
  const move = action('Roll', actionContext);
  assert.deepEqual(await move.propose(), {
    kind: 'stale-snapshot', requestID: 'request-1', expectedVersion: 4, actualVersion: 5,
  });
  assert.equal(move.reason?.code, 'snapshot-consumed');
});

function sequence(): () => string {
  let value = 0;
  return () => `request-${++value}`;
}
