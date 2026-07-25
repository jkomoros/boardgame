import assert from 'node:assert/strict';
import test from 'node:test';
import { AnimatableRegistry, type RegistrableAnimatableItem } from './animatable-registry.ts';

function makeItem(): RegistrableAnimatableItem & { finishCalls: number } {
  return {
    finishCalls: 0,
    animationContext: null,
    finishAllAnimations() { this.finishCalls++; },
  };
}

test('register makes an item show up in items()', () => {
  const registry = new AnimatableRegistry();
  const item = makeItem();
  registry.register(item);
  assert.deepEqual(registry.items(), [item]);
});

test('register is idempotent: registering the same item twice does not duplicate it', () => {
  const registry = new AnimatableRegistry();
  const item = makeItem();
  registry.register(item);
  registry.register(item);
  assert.equal(registry.items().length, 1);
  assert.deepEqual(registry.items(), [item]);
});

test('unregister removes an item', () => {
  const registry = new AnimatableRegistry();
  const item = makeItem();
  registry.register(item);
  registry.unregister(item);
  assert.deepEqual(registry.items(), []);
});

test('unregister is idempotent: unregistering an absent item is a no-op', () => {
  const registry = new AnimatableRegistry();
  const item = makeItem();
  // Never registered -- must not throw.
  assert.doesNotThrow(() => registry.unregister(item));
  assert.deepEqual(registry.items(), []);

  // Registered, unregistered, then unregistered again.
  registry.register(item);
  registry.unregister(item);
  assert.doesNotThrow(() => registry.unregister(item));
  assert.deepEqual(registry.items(), []);
});

test('items() with several registered items returns all of them', () => {
  const registry = new AnimatableRegistry();
  const a = makeItem();
  const b = makeItem();
  const c = makeItem();
  registry.register(a);
  registry.register(b);
  registry.register(c);
  const items = registry.items();
  assert.equal(items.length, 3);
  assert.ok(items.includes(a) && items.includes(b) && items.includes(c));
});

// items() must return a snapshot, not a live view: render-game's cycle-start
// reset iterates the returned array while calling into each item's
// finishAllAnimations(), which is arbitrary caller code that could itself
// register/unregister items as a side effect (e.g. an item's finish
// triggering a synchronous DOM removal -> disconnectedCallback ->
// unregister). Mutating the registry mid-iteration of a snapshot array must
// be safe -- no throw, and the array being iterated must be unaffected.
test('items() returns a snapshot: mutating the registry after calling items() does not affect the previously-returned array', () => {
  const registry = new AnimatableRegistry();
  const a = makeItem();
  const b = makeItem();
  registry.register(a);
  registry.register(b);

  const snapshot = registry.items();
  assert.equal(snapshot.length, 2);

  const c = makeItem();
  registry.register(c);
  registry.unregister(a);

  // The already-taken snapshot must be untouched by the mutation.
  assert.equal(snapshot.length, 2);
  assert.ok(snapshot.includes(a) && snapshot.includes(b));

  // A fresh call reflects the mutation.
  const after = registry.items();
  assert.equal(after.length, 2);
  assert.ok(after.includes(b) && after.includes(c) && !after.includes(a));
});

test('iterating a snapshot while mutating the registry (register/unregister from within the loop) does not throw and visits every originally-snapshotted item exactly once', () => {
  const registry = new AnimatableRegistry();
  const a = makeItem();
  const b = makeItem();
  const c = makeItem();
  registry.register(a);
  registry.register(b);
  registry.register(c);

  const visited: RegistrableAnimatableItem[] = [];
  const extra = makeItem();
  assert.doesNotThrow(() => {
    for (const item of registry.items()) {
      visited.push(item);
      item.finishAllAnimations();
      // Simulate finishAllAnimations() triggering a synchronous
      // register/unregister elsewhere (e.g. another item detaching).
      registry.unregister(item);
      registry.register(extra);
    }
  });

  assert.equal(visited.length, 3);
  assert.ok(visited.includes(a) && visited.includes(b) && visited.includes(c));
  assert.equal(a.finishCalls, 1);
  assert.equal(b.finishCalls, 1);
  assert.equal(c.finishCalls, 1);

  // a/b/c were all unregistered during the loop; extra was (repeatedly)
  // registered -- idempotent, so it appears exactly once now.
  assert.deepEqual(registry.items(), [extra]);
});
