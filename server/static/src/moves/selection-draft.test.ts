import assert from 'node:assert/strict';
import test from 'node:test';
import type { ReactiveController } from 'lit';
import {
  MoveSubmissionGate,
  createMoveAction,
  type BoundMoveAction,
  type MoveActionService,
  type MoveActionSnapshot,
} from './action.ts';
import { SelectionDraftController } from './selection-draft.ts';
import { gameSnapshotKey, type GameSnapshotHost } from './snapshot-controller.ts';

class TestHost implements GameSnapshotHost {
  state: object | null = {};
  gameName = 'cards'; gameId = 'game'; gameVersion = 1; snapshotEpoch = 1;
  viewingAsPlayer = 0; proposingAsPlayer = 0; proposingAsAdmin = false;
  updates = 0;
  addController(_controller: ReactiveController): void {}
  removeController(_controller: ReactiveController): void {}
  requestUpdate(): void { this.updates++; }
  readonly updateComplete = Promise.resolve(true);
}

type Inputs = { Trade: { Cards: string } };

function actionFor(host: TestHost, selected: readonly string[]): BoundMoveAction<'Trade', Inputs['Trade']> {
  const service: MoveActionService = {
    currentClientSchemaFingerprint: () => 'schema', currentServerSchemaFingerprint: () => 'schema',
    currentTransport: () => ({ submit: async () => ({ kind: 'success' }) }),
    currentPreviewTransport: () => ({ preview: async () => ({ kind: 'success', legal: true }) }),
    currentTargetPreviewTransport: () => null, currentGate: () => new MoveSubmissionGate(),
    nextRequestID: () => 'request-1', validate: () => [],
    serialize: (_name, input) => ({ Cards: (input as Inputs['Trade']).Cards }),
  };
  const snapshotKey = gameSnapshotKey(host);
  const snapshot: MoveActionSnapshot = {
    snapshotKey, currentSnapshotKey: () => gameSnapshotKey(host), snapshotVersion: host.gameVersion,
    currentSnapshotVersion: () => host.gameVersion, viewingAsPlayer: 0, proposingAsPlayer: 0,
    proposingAsAdmin: false, currentLegality: () => ({ legalForPlayer: true, legalForAnyone: true }),
    currentAnimating: () => false, baselineLegalityApplies: true,
  };
  return createMoveAction<'Trade', 'Trade', Inputs>('Trade', service, snapshot)
    .with({ Cards: selected.join(',') });
}

test('selection draft toggles immutable choices, undoes, and builds one exact action', () => {
  const host = new TestHost();
  const controller = new SelectionDraftController<string>(host);
  const options = {
    candidates: ['clay', 'ore', 'wool'], minSelected: 2, maxSelected: 2,
    action: (selected: readonly string[]) => actionFor(host, selected),
  };
  let draft = controller.bind(options);
  assert.equal(draft.action, null);
  draft.toggle('clay');
  draft = controller.bind(options);
  assert.deepEqual(draft.selected, ['clay']);
  assert.equal(Object.isFrozen(draft.selected), true);
  assert.deepEqual(
    { choice: draft.option('clay').choice, selected: draft.option('clay').selected,
      capacityBlocked: draft.option('ore').capacityBlocked },
    { choice: 'clay', selected: true, capacityBlocked: false },
  );
  draft.select('ore');
  draft = controller.bind(options);
  assert.equal(draft.option('wool').capacityBlocked, true);
  assert.equal(draft.option('clay').capacityBlocked, false);
  assert.deepEqual(draft.action?.input, { Cards: 'clay,ore' });
  assert.equal(draft.isSelected('ore'), true);
  assert.throws(() => draft.select('wool'), /cannot exceed 2 selections/);
  draft.toggle('clay');
  assert.deepEqual(controller.bind(options).selected, ['ore']);
  controller.bind(options).undo();
  assert.deepEqual(controller.bind(options).selected, ['clay', 'ore']);
  controller.bind(options).clear();
  assert.deepEqual(controller.bind(options).selected, []);
  controller.bind(options).undo();
  assert.deepEqual(controller.bind(options).selected, ['clay', 'ore']);
});

test('selection draft clears safely, can keep valid stable keys, and leaves stale actions closed', async () => {
  const host = new TestHost();
  const controller = new SelectionDraftController<string>(host);
  const action = (selected: readonly string[]) => actionFor(host, selected);
  let draft = controller.bind({ candidates: ['clay', 'ore'], action });
  draft.select('clay');
  const stale = controller.bind({ candidates: ['clay', 'ore'], action }).action!;
  host.state = {}; host.gameVersion++;
  draft = controller.bind({ candidates: ['clay', 'ore'], action });
  assert.deepEqual(draft.selected, []);
  assert.equal(draft.notice?.kind, 'cleared');
  assert.equal((await stale.propose()).kind, 'stale-snapshot');

  const keepHost = new TestHost();
  const keep = new SelectionDraftController<string>(keepHost);
  const keepAction = (selected: readonly string[]) => actionFor(keepHost, selected);
  const options = { candidates: ['clay', 'ore'], rebase: 'keep-valid' as const, action: keepAction };
  keep.bind(options).select('clay'); keep.bind(options).select('ore');
  keepHost.state = {}; keepHost.gameVersion++;
  const rebased = keep.bind({ ...options, candidates: ['ore'] });
  assert.deepEqual(rebased.selected, ['ore']);
  assert.deepEqual(rebased.notice?.removed, ['clay']);
  assert.equal(rebased.canUndo, false);
});

test('selection draft validates bounds, candidates, mutations, and adapters loudly', () => {
  const host = new TestHost();
  const controller = new SelectionDraftController<string>(host);
  const action = (selected: readonly string[]) => actionFor(host, selected);
  assert.throws(() => controller.bind({ candidates: ['a', 'a'], action }), /duplicated/);
  assert.throws(() => controller.bind({ candidates: [Number.NaN as never], action }), /finite string or number/);
  assert.throws(() => controller.bind({ candidates: ['a'], minSelected: 0, action }), /minSelected/);
  assert.throws(() => controller.bind({ candidates: ['a'], maxSelected: 1.5, action }), /maxSelected/);
  assert.throws(() => controller.bind({ candidates: ['a'], rebase: 'merge' as never, action }), /unknown rebase/);
  const draft = controller.bind({ candidates: ['a'], action });
  assert.throws(() => draft.select('missing'), /unknown candidate/);
  assert.throws(() => draft.option('missing'), /bind unknown candidate/);

  const invalid = new SelectionDraftController<string>(host);
  invalid.bind({ candidates: ['a'], action: () => ({} as BoundMoveAction<'Trade', Inputs['Trade']>) }).select('a');
  assert.throws(() => invalid.bind({
    candidates: ['a'], action: () => ({} as BoundMoveAction<'Trade', Inputs['Trade']>),
  }), /must return a bound move action/);
  assert.throws(() => controller.bind({
    candidates: Array.from({ length: 4097 }, (_, index) => String(index)), action,
  }), /maximum is 4096/);
});
