import test from 'node:test';
import assert from 'node:assert/strict';
import {
  clearPendingTableTransfer,
  decodeTableTransferCancel, decodeTableTransferInspection, decodeTableTransferOffer, decodeTableTransferRedemption,
  rememberPendingTableTransfer, restorePendingTableTransfer,
  TableTransferScope, transferTokenFromFragment,
} from './table-transfer.ts';

test('strictly decodes transfer protocol successes', () => {
  const pairingID = '0123456789abcdef0123456789abcdef';
  const token = `v1.Z2FtZQ.${pairingID}.${'a'.repeat(64)}`;
  assert.equal(decodeTableTransferOffer({ ok: true, pairingID, token, manualCode: '12345ABCDE', claimURL: `/table#transfer=${token}`, qrDataURL: 'data:image/png;base64,eA==', expiresAtMs: 2, serverNowMs: 1 }).token, token);
  assert.equal(decodeTableTransferInspection({ ok: true, pairingID, gameID: 'g', gameName: 'memory', gameDisplayName: 'Memory', expiresAtMs: 2, serverNowMs: 1, alreadyRedeemed: false }).gameID, 'g');
  assert.equal(decodeTableTransferRedemption({ ok: true, gameID: 'g', gameName: 'memory', gameURL: '/game/memory/g' }).gameURL, '/game/memory/g');
  assert.deepEqual(decodeTableTransferCancel({ ok: true }), { ok: true });
});

test('rejects malformed and contradictory transfer values', () => {
  assert.throws(() => decodeTableTransferOffer({ ok: false }));
  assert.throws(() => decodeTableTransferOffer({ ok: true, pairingID: 'p', token: 't', manualCode: 'c', claimURL: 'u', qrDataURL: 'javascript:x', expiresAtMs: 2, serverNowMs: 1 }));
  assert.throws(() => decodeTableTransferInspection({ ok: true, pairingID: 'p', gameID: '', gameName: 'm', gameDisplayName: 'M', expiresAtMs: 2, serverNowMs: 1 }));
  assert.throws(() => decodeTableTransferRedemption({ ok: true, gameID: 'g', gameName: 'm' }));
  assert.throws(() => decodeTableTransferRedemption({ ok: true, gameID: 'g', gameName: 'm', gameURL: '/\\evil.example' }));
  const pairingID = '0123456789abcdef0123456789abcdef';
  const token = `v1.Z2FtZQ.${pairingID}.${'a'.repeat(64)}`;
  assert.throws(() => decodeTableTransferOffer({ ok: true, pairingID, token: token.replace(pairingID, 'b'.repeat(32)), manualCode: '12345ABCDE', claimURL: `/table#transfer=${token}`, qrDataURL: 'data:image/png;base64,eA==', expiresAtMs: 2, serverNowMs: 1 }));
  assert.throws(() => decodeTableTransferOffer({ ok: true, pairingID, token, manualCode: '12345ABCDE', claimURL: `/table#transfer=other`, qrDataURL: 'data:image/png;base64,eA==', expiresAtMs: 2, serverNowMs: 1 }));
  assert.throws(() => decodeTableTransferOffer({ ok: true, pairingID, token, manualCode: '12345ABCDE', claimURL: `/table#transfer=${token}`, qrDataURL: 'data:image/png;base64,eA==', expiresAtMs: 1, serverNowMs: 1 }));
});

test('pending confirmed claims survive a receiver reload and clear explicitly', () => {
  const values = new Map<string, string>();
  const storage = {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => { values.set(key, value); },
    removeItem: (key: string) => { values.delete(key); },
  } as Storage;
  Object.defineProperty(globalThis, 'sessionStorage', { configurable: true, value: storage });
  const input = { kind: 'token' as const, token: `v1.Z2FtZQ.${'a'.repeat(32)}.${'b'.repeat(64)}` };
  rememberPendingTableTransfer(input, true);
  assert.deepEqual(restorePendingTableTransfer(), { input, confirmed: true });
  clearPendingTableTransfer();
  assert.equal(restorePendingTableTransfer(), null);
  Reflect.deleteProperty(globalThis, 'sessionStorage');
});

test('fragment parser is exact and route scope invalidates stale operations', () => {
  assert.equal(transferTokenFromFragment('#transfer=secret'), 'secret');
  assert.equal(transferTokenFromFragment('#other=x'), null);
  assert.equal(transferTokenFromFragment('#transfer=a&transfer=b'), null);
  const scope = new TableTransferScope();
  const first = scope.begin();
  const second = scope.begin();
  assert.equal(first.signal.aborted, true);
  assert.equal(first.isCurrent(), false);
  assert.equal(second.isCurrent(), true);
  scope.invalidate();
  assert.equal(second.isCurrent(), false);
});
