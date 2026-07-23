import assert from 'node:assert/strict';
import { describe, test } from 'node:test';
import { compileMotionPresence, motionPresenceHostStyle } from './presence.ts';

describe('motion presence', () => {
  test('compiles immutable travel-only and scale-fade policies', () => {
    assert.deepEqual(motionPresenceHostStyle(compileMotionPresence('travel-only')), { transform: '', opacity: '1' });
    const facts = compileMotionPresence('scale-fade');
    const style = motionPresenceHostStyle(facts);
    assert.deepEqual(style, { transform: 'scale(0.6)', opacity: '0' });
    assert.deepEqual(facts, { scale: 0.6, opacity: 0 });
    assert.ok(Object.isFrozen(facts));
    assert.ok(Object.isFrozen(style));
  });

  test('rejects arbitrary policies', () => {
    assert.throws(() => compileMotionPresence('custom' as never), /unknown/);
  });
});
