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
import { PlacementDraftController, type DraftPlacement } from './placement-draft.ts';
import { gameSnapshotKey, type GameSnapshotHost } from './snapshot-controller.ts';

class TestHost implements GameSnapshotHost {
  readonly controllers: ReactiveController[] = [];
  updates = 0;
  state: object | null = {};
  gameName = 'words';
  gameId = 'game';
  gameVersion = 1;
  snapshotEpoch = 1;
  viewingAsPlayer = 0;
  proposingAsPlayer = 0;
  proposingAsAdmin = false;
  addController(controller: ReactiveController): void { this.controllers.push(controller); }
  removeController(controller: ReactiveController): void {
    const index = this.controllers.indexOf(controller);
    if (index >= 0) this.controllers.splice(index, 1);
  }
  requestUpdate(): void { this.updates++; }
  readonly updateComplete = Promise.resolve(true);
}

type Inputs = { Play: { Placements: string } };

function actionFor(
  host: TestHost,
  placements: readonly DraftPlacement<string, number>[],
): BoundMoveAction<'Play', Inputs['Play']> {
  const gate = new MoveSubmissionGate();
  const service: MoveActionService = {
    currentClientSchemaFingerprint: () => 'schema',
    currentServerSchemaFingerprint: () => 'schema',
    currentTransport: () => ({ submit: async () => ({ kind: 'success' }) }),
    currentPreviewTransport: () => ({ preview: async () => ({ kind: 'success', legal: true }) }),
    currentTargetPreviewTransport: () => null,
    currentGate: () => gate,
    nextRequestID: () => 'request-1',
    validate: (_name, input) => typeof (input as Partial<Inputs['Play']>).Placements === 'string'
      ? []
      : [{ field: 'Placements', code: 'invalid-type', message: 'Placements must be a string' }],
    serialize: (_name, input) => ({ Placements: (input as Inputs['Play']).Placements }),
  };
  const snapshotKey = gameSnapshotKey(host);
  const snapshot: MoveActionSnapshot = {
    snapshotKey,
    currentSnapshotKey: () => gameSnapshotKey(host),
    snapshotVersion: host.gameVersion,
    currentSnapshotVersion: () => host.gameVersion,
    viewingAsPlayer: host.viewingAsPlayer,
    proposingAsPlayer: host.proposingAsPlayer,
    proposingAsAdmin: host.proposingAsAdmin,
    currentLegality: () => ({ legalForPlayer: true, legalForAnyone: true }),
    currentAnimating: () => false,
    baselineLegalityApplies: true,
  };
  return createMoveAction<'Play', 'Play', Inputs>('Play', service, snapshot)
    .with({ Placements: JSON.stringify(placements) });
}

test('placement draft supports select/place and direct assignment through one immutable history', () => {
  const host = new TestHost();
  const controller = new PlacementDraftController<string, number>(host);
  const options = {
    items: ['a', 'b', 'c'], targets: [0, 1, 2], minPlacements: 2,
    action: (placements: readonly DraftPlacement<string, number>[]) => actionFor(host, placements),
  };

  let draft = controller.bind(options);
  assert.equal(draft.action, null);
  assert.equal(draft.item('a').selected, false);
  assert.equal(draft.target(0).canPlace, false);
  assert.equal(draft.target(0).reason, 'Select an item first');
  draft.selectItem('a');
  draft = controller.bind(options);
  assert.equal(draft.selectedItem, 'a');
  assert.equal(draft.item('a').selected, true);
  assert.equal(draft.target(0).canPlace, true);
  draft.place(0);
  draft = controller.bind(options);
  assert.deepEqual(draft.placements, [{ item: 'a', target: 0 }]);
  assert.equal(draft.item('a').placedAt, 0);
  assert.equal(draft.target(0).occupiedBy, 'a');
  assert.equal(draft.selectedItem, null);
  assert.equal(draft.action, null);
  assert.throws(() => { (draft.placements as DraftPlacement<string, number>[]).push({ item: 'x', target: 9 }); });

  draft.assign('b', 1);
  draft = controller.bind(options);
  assert.deepEqual(draft.placements, [{ item: 'a', target: 0 }, { item: 'b', target: 1 }]);
  assert.equal(draft.targetFor('b'), 1);
  assert.equal(draft.itemAt(0), 'a');
  assert.deepEqual(draft.action?.input, { Placements: '[{"item":"a","target":0},{"item":"b","target":1}]' });

  draft.assign('a', 2);
  assert.deepEqual(controller.bind(options).placements, [{ item: 'a', target: 2 }, { item: 'b', target: 1 }]);
  controller.bind(options).undo();
  assert.deepEqual(controller.bind(options).placements, [{ item: 'a', target: 0 }, { item: 'b', target: 1 }]);
  controller.bind(options).clear();
  assert.deepEqual(controller.bind(options).placements, []);
  controller.bind(options).undo();
  assert.equal(controller.bind(options).placements.length, 2);
});

test('placement draft rejects ambiguous mutations and stale actions fail closed', async () => {
  const host = new TestHost();
  const controller = new PlacementDraftController<string, number>(host);
  const options = {
    items: ['a', 'b'], targets: [0, 1], maxPlacements: 1,
    action: (placements: readonly DraftPlacement<string, number>[]) => actionFor(host, placements),
  };
  let draft = controller.bind(options);
  assert.throws(() => draft.place(0), /before selecting/);
  assert.throws(() => draft.selectItem('missing'), /unknown item/);
  assert.throws(() => draft.assign('a', 8), /unknown target/);
  assert.throws(() => draft.item('missing'), /bind unknown item/);
  assert.throws(() => draft.target(8), /bind unknown target/);
  draft.assign('a', 0);
  draft = controller.bind(options);
  assert.equal(draft.item('b').capacityBlocked, true);
  assert.equal(draft.item('a').capacityBlocked, false);
  draft.selectItem('b');
  draft = controller.bind(options);
  assert.equal(draft.target(0).reason, 'Destination is occupied');
  assert.equal(draft.target(1).reason, 'Maximum placements reached');
  assert.equal(draft.target(1).canPlace, false);
  draft.selectItem('b');
  assert.throws(() => draft.assign('b', 0), /already occupied/);
  assert.throws(() => draft.assign('b', 1), /cannot exceed 1 placements/);

  const exactAction = draft.action;
  assert.ok(exactAction);
  host.gameVersion++;
  const result = await exactAction.propose();
  assert.equal(result.kind, 'stale-snapshot');
});

test('snapshot rebasing clears by default or explicitly retains only valid placements', () => {
  const host = new TestHost();
  const clearController = new PlacementDraftController<string, number>(host);
  const action = (placements: readonly DraftPlacement<string, number>[]) => actionFor(host, placements);
  let clear = clearController.bind({ items: ['a'], targets: [0], action });
  clear.assign('a', 0);
  host.state = {};
  host.gameVersion++;
  clear = clearController.bind({ items: ['a'], targets: [0], action });
  assert.deepEqual(clear.placements, []);
  assert.equal(clear.notice?.kind, 'cleared');
  assert.deepEqual(clear.notice?.removed, [{ item: 'a', target: 0 }]);
  assert.equal(clear.canUndo, false);

  const keepHost = new TestHost();
  const keepController = new PlacementDraftController<string, number>(keepHost);
  const keepAction = (placements: readonly DraftPlacement<string, number>[]) => actionFor(keepHost, placements);
  const keepOptions = { items: ['a', 'b'], targets: [0, 1], maxPlacements: 2, rebase: 'keep-valid' as const, action: keepAction };
  let keep = keepController.bind(keepOptions);
  keep.assign('a', 0);
  keep.assign('b', 1);
  keepHost.state = {};
  keepHost.gameVersion++;
  keep = keepController.bind({ ...keepOptions, items: ['a'], targets: [0, 1] });
  assert.deepEqual(keep.placements, [{ item: 'a', target: 0 }]);
  assert.equal(keep.notice?.kind, 'pruned');
  assert.deepEqual(keep.notice?.removed, [{ item: 'b', target: 1 }]);
  keep.dismissNotice();
  assert.equal(keepController.bind({ ...keepOptions, items: ['a'], targets: [0, 1] }).notice, null);
});

test('placement draft validates configuration and commit adapters loudly', () => {
  const host = new TestHost();
  const controller = new PlacementDraftController<string, number>(host);
  const action = (placements: readonly DraftPlacement<string, number>[]) => actionFor(host, placements);
  assert.throws(() => controller.bind({ items: ['a', 'a'], targets: [0], action }), /item "a" is duplicated/);
  assert.throws(() => controller.bind({
    items: Array.from({ length: 1025 }, (_, index) => `item-${index}`), targets: [0], action,
  }), /maximum is 1024/);
  assert.throws(() => controller.bind({
    items: ['a'], targets: Array.from({ length: 4097 }, (_, index) => index), action,
  }), /maximum is 4096/);
  assert.throws(() => controller.bind({ items: ['a'], targets: [Number.NaN], action }), /finite number/);
  assert.throws(() => controller.bind({ items: ['a'], targets: [0], minPlacements: 0, action }), /minPlacements/);
  assert.throws(() => controller.bind({ items: ['a'], targets: [0], maxPlacements: 1.5, action }), /maxPlacements/);
  assert.throws(() => controller.bind({
    items: ['a'], targets: [0], rebase: 'merge' as never, action,
  }), /unknown rebase policy/);
  assert.doesNotThrow(() => controller.bind({
    items: [], targets: [], minPlacements: 2, action,
  }));

  const invalidAction = new PlacementDraftController<string, number>(host);
  invalidAction.bind({ items: ['a'], targets: [0], action: () => ({} as BoundMoveAction<'Play', Inputs['Play']>) })
    .assign('a', 0);
  assert.throws(() => invalidAction.bind({
    items: ['a'], targets: [0], action: () => ({} as BoundMoveAction<'Play', Inputs['Play']>),
  }), /must return a bound move action/);
});
