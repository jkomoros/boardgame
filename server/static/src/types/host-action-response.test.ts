import assert from 'node:assert/strict';
import test from 'node:test';
import { decodeHostActionResponse } from './host-action-response.ts';

test('host action decoder copies the exact success contract', () => {
  assert.deepEqual(decodeHostActionResponse({ ok: true, locked: false, secret: 'discard' }), {
    ok: true,
    locked: false,
  });
});

test('host action decoder rejects false success and malformed lock state', () => {
  assert.throws(() => decodeHostActionResponse({ ok: false }), /ok must be true/);
  assert.throws(() => decodeHostActionResponse({ ok: true, locked: 'yes' }), /locked must be a boolean/);
});
