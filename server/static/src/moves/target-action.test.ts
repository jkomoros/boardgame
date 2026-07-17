import assert from 'node:assert/strict';
import test from 'node:test';
import {
  MoveSubmissionGate,
  createMoveAction,
  type MoveActionLegality,
  type MoveActionService,
  type MoveActionSnapshot,
} from './action.ts';
import { serializeCreatorMoveInput, validateCreatorMoveInput } from './input.ts';
import {
  notifyTargetActionLiveStateChanged,
  type TargetPreviewRequest,
  type TargetPreviewTransportResult,
} from './target-action.ts';

type Names = 'Place';
type Inputs = { Place: { Slot: number } };
const schema = [{
  name: 'Place',
  fields: [{ name: 'Slot', wireType: 'int', disposition: 'required', codec: 'integer' }],
}] as const;

function fixture(
  preview: (request: TargetPreviewRequest) => Promise<TargetPreviewTransportResult>,
  legality: MoveActionLegality = { legalForAnyone: true, legalForPlayer: true },
) {
  const requests: TargetPreviewRequest[] = [];
  const submissions: Readonly<Record<string, string>>[] = [];
  let sequence = 0;
  const gate = new MoveSubmissionGate();
  let currentSnapshotKey = 'game:g1:v4';
  const service: MoveActionService = {
    currentClientSchemaFingerprint: () => 'sha256:test',
    currentServerSchemaFingerprint: () => 'sha256:test',
    currentTransport: () => ({ submit: async request => {
      submissions.push(request.arguments);
      return { kind: 'success' };
    } }),
    currentPreviewTransport: () => ({ preview: async () => {
      throw new Error('Target candidates must not use individual preview');
    } }),
    currentTargetPreviewTransport: () => ({ previewTargets: async request => {
      requests.push(request);
      return preview(request);
    } }),
    currentGate: () => gate,
    nextRequestID: () => `request-${++sequence}`,
    validate: (name, input) => validateCreatorMoveInput(schema, name, input).errors,
    serialize: (name, input) => serializeCreatorMoveInput(schema, name, input),
    actionCache: new Map(),
    targetActionCache: new Map(),
  };
  const snapshot: MoveActionSnapshot = {
    snapshotKey: 'game:g1:v4',
    currentSnapshotKey: () => currentSnapshotKey,
    snapshotVersion: 4,
    currentSnapshotVersion: () => 4,
    viewingAsPlayer: 0,
    proposingAsPlayer: 0,
    proposingAsAdmin: false,
    currentLegality: () => legality,
    currentAnimating: () => false,
    baselineLegalityApplies: true,
  };
  const builder = createMoveAction<'Place', Names, Inputs>('Place', service, snapshot);
  return { builder, requests, submissions, service, setSnapshotKey: (key: string) => { currentSnapshotKey = key; } };
}

test('candidate subscriptions coalesce into one correlated batch and activation uses the canonical action', async () => {
  const value = fixture(async request => ({
    kind: 'success',
    results: [...request.candidates].reverse().map(candidate => ({
      id: candidate.id,
      legal: candidate.arguments['Slot'] !== '1',
      ...(candidate.arguments['Slot'] === '1' ? { error: 'occupied' } : {}),
    })),
  }));
  const targets = value.builder.targets([0, 1, 2] as const, Slot => ({ Slot }));
  const unsubscribes = targets.candidates.map(candidate => candidate.action.subscribe(() => undefined));
  const preview = await targets.ensurePreview();
  assert.equal(preview.kind, 'ready');
  assert.equal(value.requests.length, 1);
  assert.deepEqual(value.requests[0]?.candidates.map(candidate => candidate.id), [
    '["number",0]', '["number",1]', '["number",2]',
  ]);
  assert.equal(targets.get(0)?.action.canActivate, true);
  assert.equal(targets.get(1)?.action.reason?.message, 'occupied');
  assert.equal((await targets.get(1)?.action.activate())?.kind, 'blocked');
  assert.equal((await targets.get(2)?.action.activate())?.kind, 'success');
  assert.deepEqual(value.submissions, [{ Slot: '2' }]);
  for (const unsubscribe of unsubscribes) unsubscribe();
});

test('batch-hydrated candidates supersede an illegal default-form baseline', async () => {
  const value = fixture(async request => ({
    kind: 'success',
    results: request.candidates.map(candidate => ({
      id: candidate.id,
      legal: candidate.arguments['Slot'] === '2',
      ...(candidate.arguments['Slot'] === '2' ? {} : { error: 'occupied' }),
    })),
  }), { legalForAnyone: false, legalForPlayer: false });
  const targets = value.builder.targets([1, 2] as const, Slot => ({ Slot }));
  assert.equal((await targets.ensurePreview()).kind, 'ready');
  assert.equal(targets.get(1)?.action.reason?.code, 'preview-illegal');
  assert.equal(targets.get(1)?.action.canActivate, false);
  assert.equal(targets.get(2)?.action.reason, null);
  assert.equal(targets.get(2)?.action.canActivate, true);
  assert.equal((await targets.get(2)?.action.activate())?.kind, 'success');
  assert.deepEqual(value.submissions, [{ Slot: '2' }]);
});

test('host live-state invalidation reaches cached target candidates', async () => {
  const value = fixture(async request => ({
    kind: 'success', results: request.candidates.map(candidate => ({ id: candidate.id, legal: true })),
  }));
  const targets = value.builder.targets([0], Slot => ({ Slot }));
  const candidate = targets.get(0)!.action;
  let candidateNotifications = 0;
  let targetNotifications = 0;
  const unsubscribe = candidate.subscribe(() => { candidateNotifications++; });
  const unsubscribeTarget = targets.subscribe(() => { targetNotifications++; });
  await targets.ensurePreview();
  candidateNotifications = 0;
  targetNotifications = 0;
  notifyTargetActionLiveStateChanged(targets);
  assert.equal(candidateNotifications, 1);
  assert.equal(targetNotifications, 1);
  unsubscribe();
  unsubscribeTarget();
});

test('recreation is cached and malformed correlation fails every candidate closed', async () => {
  const value = fixture(async request => ({
    kind: 'success',
    results: request.candidates.map(candidate => ({ id: `${candidate.id}-unknown`, legal: true })),
  }));
  const first = value.builder.targets(['a.b', 'c#d'] as const, (key, index) => ({ Slot: index }));
  const second = value.builder.targets(['a.b', 'c#d'] as const, (key, index) => ({ Slot: index }));
  assert.equal(first, second);
  const result = await first.ensurePreview();
  assert.equal(result.kind, 'failed');
  assert.match(first.get('a.b')?.action.reason?.message ?? '', /unknown candidate/);
  assert.equal(first.get('a.b')?.action.canActivate, false);
});

test('transient batch failure retries as a batch when a candidate activates', async () => {
  let calls = 0;
  const value = fixture(async request => {
    calls++;
    if (calls === 1) return { kind: 'failure', error: 'offline', retryable: true };
    return { kind: 'success', results: request.candidates.map(candidate => ({ id: candidate.id, legal: true })) };
  });
  const targets = value.builder.targets([0, 1], Slot => ({ Slot }));
  assert.equal((await targets.ensurePreview()).kind, 'failed');
  assert.equal((await targets.get(0)?.action.activate())?.kind, 'success');
  assert.equal(calls, 2);
  assert.equal(value.requests.length, 2);
});

test('configuration rejects duplicate keys, duplicate arguments, invalid input, and mapper failures', () => {
  const value = fixture(async () => ({ kind: 'success', results: [] }));
  assert.throws(() => value.builder.targets([1, 1], Slot => ({ Slot })), /Duplicate target key/);
  assert.throws(() => value.builder.targets([Number.NaN], Slot => ({ Slot })), /must be finite/);
  assert.throws(() => value.builder.targets(Array.from({ length: 1025 }, (_, index) => index), Slot => ({ Slot })), /maximum is 1024/);
  assert.throws(() => value.builder.targets(['left', 'right'], () => ({ Slot: 1 })), /identical move arguments/);
  const aliases = value.builder.targets(
    ['left', 'right'],
    () => ({ Slot: 1 }),
    { allowDuplicateInputs: true },
  );
  assert.equal(aliases.candidates.length, 2);
  assert.throws(() => value.builder.targets([0], () => ({ Slot: Number.NaN })), /Invalid target input/);
  assert.throws(() => value.builder.targets([0], () => { throw new Error('boom'); }), /index 0: boom/);
  const bound = value.builder.with({ Slot: 0 }) as unknown as {
    targets(keys: readonly number[], mapper: (key: number) => { Slot: number }): unknown;
  };
  assert.throws(() => bound.targets([0], Slot => ({ Slot })), /before \.with/);

  const keys = [0, 1];
  const copied = value.builder.targets(keys, Slot => ({ Slot }));
  keys[0] = 9;
  assert.deepEqual(copied.candidates.map(candidate => candidate.key), [0, 1]);
});

test('refresh aborts the prior batch and late completion cannot overwrite the new result', async () => {
  const resolvers: ((result: TargetPreviewTransportResult) => void)[] = [];
  const value = fixture(request => new Promise(resolve => { resolvers.push(resolve); }));
  const targets = value.builder.targets([0, 1], Slot => ({ Slot }));
  const first = targets.ensurePreview();
  const refresh = targets.refreshPreview();
  assert.equal(value.requests.length, 2);
  assert.equal(value.requests[0]?.signal.aborted, true);
  const success = (request: TargetPreviewRequest): TargetPreviewTransportResult => ({
    kind: 'success',
    results: request.candidates.map(candidate => ({ id: candidate.id, legal: true })),
  });
  resolvers[1]?.(success(value.requests[1]!));
  assert.equal((await refresh).kind, 'ready');
  resolvers[0]?.({
    kind: 'success',
    results: value.requests[0]!.candidates.map(candidate => ({ id: candidate.id, legal: false, error: 'late' })),
  });
  await first;
  assert.equal(targets.get(0)?.action.canPropose, true);
});

test('a collection belonging to an old snapshot fails before transport', async () => {
  const value = fixture(async request => ({
    kind: 'success', results: request.candidates.map(candidate => ({ id: candidate.id, legal: true })),
  }));
  const targets = value.builder.targets([0], Slot => ({ Slot }));
  value.setSnapshotKey('game:g1:v5');
  const preview = await targets.ensurePreview();
  assert.equal(preview.kind, 'failed');
  assert.equal(value.requests.length, 0);
  assert.equal(targets.get(0)?.action.reason?.code, 'stale-snapshot');
});

test('a persistent retryable failure performs exactly one new batch per activation', async () => {
  const value = fixture(async () => ({ kind: 'failure', error: 'offline', retryable: true }));
  const targets = value.builder.targets([0], Slot => ({ Slot }));
  assert.equal((await targets.ensurePreview()).kind, 'failed');
  assert.equal((await targets.get(0)?.action.activate())?.kind, 'blocked');
  assert.equal(value.requests.length, 2);
});

test('malformed result values fail closed instead of relying on JavaScript truthiness', async () => {
  const value = fixture(async request => ({
    kind: 'success',
    results: [{ id: request.candidates[0]!.id, legal: 'false' }],
  } as unknown as TargetPreviewTransportResult));
  const targets = value.builder.targets([0], Slot => ({ Slot }));
  const preview = await targets.ensurePreview();
  assert.equal(preview.kind, 'failed');
  assert.match(targets.get(0)?.action.reason?.message ?? '', /malformed/);
});

test('the maximum collection has constant coordinator notifications and bounded caches', async () => {
  const value = fixture(async request => ({
    kind: 'success',
    results: request.candidates.map(candidate => ({ id: candidate.id, legal: true })),
  }));
  const targets = value.builder.targets(
    Array.from({ length: 1024 }, (_, index) => index),
    Slot => ({ Slot }),
  );
  let notifications = 0;
  const unsubscribe = targets.subscribe(() => { notifications++; });
  await targets.ensurePreview();
  unsubscribe();
  assert.ok(notifications <= 2, `received ${notifications} notifications for one batch`);
  assert.equal(value.service.actionCache?.size, 1, 'candidate actions must not occupy the global action cache');

  for (let collection = 0; collection < 40; collection++) {
    value.builder.targets([collection + 2000], Slot => ({ Slot }));
  }
  assert.equal(value.service.targetActionCache?.size, 32);
});
