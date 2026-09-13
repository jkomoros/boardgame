import assert from 'node:assert/strict';
import test from 'node:test';
import { DeferredMotion } from './deferred-motion.ts';

class FakeFrames {
  private _nextHandle = 1;
  private readonly _callbacks = new Map<number, FrameRequestCallback>();

  request = (callback: FrameRequestCallback): number => {
    const handle = this._nextHandle++;
    this._callbacks.set(handle, callback);
    return handle;
  };

  cancel = (handle: number): void => {
    this._callbacks.delete(handle);
  };

  advance(): void {
    const callbacks = [...this._callbacks.values()];
    this._callbacks.clear();
    for (const callback of callbacks) callback(0);
  }
}

test('deferred motion only starts current work for a connected owner', () => {
  const frames = new FakeFrames();
  const deferred = new DeferredMotion(frames.request, frames.cancel);
  const owner = { isConnected: true };
  const calls: string[] = [];

  deferred.schedule(owner as Element, () => calls.push('superseded'));
  deferred.schedule(owner as Element, () => calls.push('current'));
  frames.advance();
  frames.advance();
  assert.deepEqual(calls, ['current']);

  deferred.schedule(owner as Element, () => calls.push('detached'));
  frames.advance();
  owner.isConnected = false;
  frames.advance();
  assert.deepEqual(calls, ['current']);
});
