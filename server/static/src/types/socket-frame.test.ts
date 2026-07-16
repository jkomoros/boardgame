import assert from 'node:assert/strict';
import test from 'node:test';
import { decodeSocketFrame } from './socket-frame.ts';

test('socket decoder accepts exact JSON and legacy version frames', () => {
  assert.deepEqual(decodeSocketFrame('{"type":"version","data":4}'), {
    type: 'version', version: 4, transport: 'json',
  });
  assert.deepEqual(decodeSocketFrame('12'), {
    type: 'version', version: 12, transport: 'legacy',
  });
  assert.throws(() => decodeSocketFrame('12junk'), /exact version integer/);
});

test('socket decoder validates timing policy relationships', () => {
  const frame = JSON.stringify({
    type: 'version-timing',
    data: {
      version: 4,
      serverSentAt: 1_000,
      serverPlayAt: 1_500,
      slotDurationMs: 800,
      maxAnimationDurationMs: 600,
    },
  });
  assert.equal(decodeSocketFrame(frame).type, 'version-timing');
  assert.throws(() => decodeSocketFrame(frame.replace('600', '900')), /must not exceed/);
  assert.throws(() => decodeSocketFrame(frame.replace('1500', '500')), /must not precede/);
  assert.throws(() => decodeSocketFrame(frame.replace('800', '80000')), /no greater than/);
});

test('socket decoder fails closed for behavior-changing frames', () => {
  assert.deepEqual(decodeSocketFrame(JSON.stringify({
    type: 'mode-changed', data: { gameID: 'GAME', newMode: 'solo', ignored: true },
  })), { type: 'mode-changed', gameID: 'GAME', newMode: 'solo' });
  assert.throws(() => decodeSocketFrame(JSON.stringify({
    type: 'mode-changed', data: { gameID: 'GAME', newMode: 'table' },
  })), /newMode must be "solo"/);
  assert.throws(() => decodeSocketFrame(JSON.stringify({
    type: 'presence-changed', data: { gameID: '' },
  })), /gameID must be/);
	assert.deepEqual(decodeSocketFrame(JSON.stringify({
		type: 'table-session-changed', data: { gameID: 'GAME' },
	})), { type: 'table-session-changed', gameID: 'GAME' });
	assert.deepEqual(decodeSocketFrame(JSON.stringify({
		type: 'table-lease-lost', data: { gameID: 'GAME' },
	})), { type: 'table-lease-lost', gameID: 'GAME' });
});

test('socket decoder preserves forward compatibility without copying unknown data', () => {
  assert.deepEqual(decodeSocketFrame(JSON.stringify({
    type: 'future-frame', data: { dangerous: true },
  })), { type: 'unknown', wireType: 'future-frame' });
});
