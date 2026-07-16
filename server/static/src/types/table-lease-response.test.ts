import assert from 'node:assert/strict';
import test from 'node:test';
import {
  decodeTableLeaseAcquireResponse,
  isTableLeaseFailureCode,
  tableLeaseFailureMessage,
} from './table-lease-response.ts';

test('table lease decoder copies the exact success contract', () => {
  assert.deepEqual(decodeTableLeaseAcquireResponse({
    ok: true,
    alreadyHeld: false,
    expiresAtMs: 1_800_000_000_000,
    capability: 'must not escape the decoder',
  }), {
    ok: true,
    alreadyHeld: false,
    expiresAtMs: 1_800_000_000_000,
  });
});

test('table lease failures have stable user-facing explanations', () => {
  assert.match(tableLeaseFailureMessage('TABLE_LEASE_ACTIVE'), /still active/);
  assert.match(tableLeaseFailureMessage('TABLE_LEASE_NOT_ELIGIBLE'), /seated player/);
  assert.match(tableLeaseFailureMessage('GAME_NOT_COMPANION'), /companion mode/);
  assert.match(tableLeaseFailureMessage('GAME_FINISHED'), /finished/);
});

test('table lease decoder rejects malformed success values', () => {
  assert.throws(() => decodeTableLeaseAcquireResponse(null), /must be an object/);
  assert.throws(() => decodeTableLeaseAcquireResponse({ ok: false }), /ok must be true/);
  assert.throws(
    () => decodeTableLeaseAcquireResponse({ ok: true, alreadyHeld: 'no', expiresAtMs: 1 }),
    /alreadyHeld must be a boolean/,
  );
  for (const expiresAtMs of [0, -1, 1.5, Number.MAX_SAFE_INTEGER + 1]) {
    assert.throws(
      () => decodeTableLeaseAcquireResponse({ ok: true, alreadyHeld: false, expiresAtMs }),
      /expiresAtMs must be a positive safe integer/,
    );
  }
});

test('table lease failure code guard recognizes only the public protocol', () => {
  assert.equal(isTableLeaseFailureCode('TABLE_LEASE_ACTIVE'), true);
  assert.equal(isTableLeaseFailureCode('GAME_FINISHED'), true);
  assert.equal(isTableLeaseFailureCode('INTERNAL_DETAIL'), false);
  assert.equal(isTableLeaseFailureCode(undefined), false);
});
