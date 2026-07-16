import assert from 'node:assert/strict';
import test from 'node:test';
import type { ReactiveController, ReactiveControllerHost } from 'lit';
import { SourceDestinationController } from './source-destination.ts';
import type { TargetAction } from './target-action.ts';

class TestHost implements ReactiveControllerHost {
  readonly controllers: ReactiveController[] = [];
  updates = 0;
  state: object | null = {};
  gameName = 'test';
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

function action(keys: readonly number[]): TargetAction<number, 'Move', { From: number; To: number }> {
  return {
    candidates: keys.map(key => ({ key, action: {} as never })),
    preview: { kind: 'unchecked' },
    get: key => keys.includes(key) ? { key, action: {} as never } : undefined,
    ensurePreview: async () => ({ kind: 'unchecked' }),
    refreshPreview: async () => ({ kind: 'unchecked' }),
    subscribe: () => () => undefined,
  };
}

test('source selection toggles, resets with the renderer snapshot, and derives destinations', () => {
  const host = new TestHost();
  const controller = new SourceDestinationController<number>(host);
  const options = {
    sources: [1, 3],
    destinations: (source: number) => action(source === 1 ? [2] : [4]),
  };

  let binding = controller.bind(options);
  assert.equal(binding.selectedSource, null);
  assert.equal(binding.action, null);
  binding.selectSource(1);
  assert.equal(host.updates, 1);

  binding = controller.bind(options);
  assert.equal(binding.selectedSource, 1);
  assert.deepEqual(binding.action?.candidates.map(candidate => candidate.key), [2]);
  binding.selectSource(1);
  assert.equal(controller.bind(options).selectedSource, null);

  controller.bind(options).selectSource(3);
  host.gameVersion++;
  assert.equal(controller.bind(options).selectedSource, null);
});

test('source configuration and ambiguous workflows fail loudly', () => {
  const controller = new SourceDestinationController<number>(new TestHost());
  assert.throws(() => controller.bind({ sources: [1, 1], destinations: () => action([2]) }), /duplicated/);
  assert.throws(() => controller.bind({ sources: [Number.NaN], destinations: () => action([2]) }), /finite/);

  let binding = controller.bind({ sources: [1], destinations: () => action([1]) });
  assert.throws(() => binding.selectSource(2), /unknown source/);
  binding.selectSource(1);
  assert.throws(
    () => controller.bind({ sources: [1], destinations: () => action([1]) }),
    /ambiguously both a source and destination/,
  );

  const secondHost = new TestHost();
  const secondController = new SourceDestinationController<number>(secondHost);
  binding = secondController.bind({ sources: [2], destinations: () => action([3]) });
  binding.selectSource(2);
  assert.throws(
    () => secondController.bind({ sources: [2], destinations: () => { throw new Error('boom'); } }),
    /destinations failed.*boom/,
  );
});
