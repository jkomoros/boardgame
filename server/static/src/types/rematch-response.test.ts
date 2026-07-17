import assert from 'node:assert/strict';
import test from 'node:test';
import { decodeRematchResponse } from './rematch-response.ts';

test('rematch decoder copies the exact ready successor contract', () => {
  assert.deepEqual(decodeRematchResponse({
    ok: true,
    gameID: '0123456789abcdef',
    gameName: 'blackjack',
    roomCode: 'ABCD',
    secret: 'must not escape',
  }), {
    ok: true,
    gameID: '0123456789abcdef',
    gameName: 'blackjack',
    roomCode: 'ABCD',
  });
});

test('rematch decoder rejects partial, failed, and malformed successors', () => {
  assert.throws(() => decodeRematchResponse(null), /must be an object/);
  assert.throws(() => decodeRematchResponse({ ok: false }), /ok must be true/);
  assert.throws(
    () => decodeRematchResponse({ ok: true, gameID: '', gameName: 'pig', roomCode: 'ABCD' }),
    /gameID must be a non-empty string/,
  );
  assert.throws(
    () => decodeRematchResponse({ ok: true, gameID: 'GAME', gameName: 'pig', roomCode: 'I234' }),
    /canonical room code/,
  );
});
