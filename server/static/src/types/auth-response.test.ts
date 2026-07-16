import assert from 'node:assert/strict';
import test from 'node:test';
import { decodeAuthResponse } from './auth-response.ts';

test('auth decoder distinguishes a valid logout from malformed identity data', () => {
  assert.deepEqual(
    decodeAuthResponse({ Status: 'Success', Message: 'signed out' }),
    { User: null, AdminAllowed: false, Message: 'signed out' },
  );
  assert.throws(
    () => decodeAuthResponse({ Status: 'Success', User: { ID: '' }, AdminAllowed: false }),
    /Auth response\.User\.ID must be a non-empty string/,
  );
});

test('auth decoder copies the exact public user shape and accepts Go nanosecond timestamps', () => {
  const decoded = decodeAuthResponse({
    Status: 'Success',
    User: {
      ID: 'ada@example.com',
      DisplayName: 'Ada & Grace',
      Email: 'ada+games@example.com',
      PhotoURL: 'https://example.com/a=b&c=d',
      Created: 1_784_000_000_000_000_000,
      LastSeen: 1_784_000_000_000_000_000,
      Secret: 'discard me',
    },
    AdminAllowed: true,
  });
  assert.equal(decoded.User?.PhotoURL, 'https://example.com/a=b&c=d');
  assert.equal('Secret' in (decoded.User ?? {}), false);
  assert.equal(decoded.AdminAllowed, true);
});
