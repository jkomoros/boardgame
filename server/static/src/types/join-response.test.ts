import assert from 'node:assert/strict';
import test from 'node:test';
import {
  decodeJoinResponse,
  decodeJoinSeatResponse,
  decodeSeatOptionsResponse,
} from './join-response.ts';

test('join decoder enforces bounded player counts and copies the wire contract', () => {
  assert.deepEqual(decodeJoinResponse({
    gameID: 'GAME', gameName: 'pig', gameDisplayName: 'Pig',
    minPlayers: 2, maxPlayers: 4, currentPlayers: 1, requiresSeatPicker: false,
    secret: 'discard',
  }), {
    gameID: 'GAME', gameName: 'pig', gameDisplayName: 'Pig',
    minPlayers: 2, maxPlayers: 4, currentPlayers: 1, requiresSeatPicker: false,
  });
  assert.throws(() => decodeJoinResponse({
    gameID: 'GAME', gameName: 'pig', gameDisplayName: 'Pig',
    minPlayers: 4, maxPlayers: 2, currentPlayers: 1, requiresSeatPicker: false,
  }), /minPlayers must not exceed/);
});

test('seat options require canonical contiguous indexes and exact slot values', () => {
  const decoded = decodeSeatOptionsResponse({
    gameID: 'GAME', gameName: 'werewolf', requiresSeatPicker: true,
    slots: [{ playerIndex: 0, label: 'Seat 1', filled: true, displayName: 'Ada' }],
  });
  assert.equal(decoded.slots[0]?.displayName, 'Ada');
  assert.throws(() => decodeSeatOptionsResponse({
    gameID: 'GAME', gameName: 'werewolf', requiresSeatPicker: true,
    slots: [{ playerIndex: 2, label: 'Seat 1', filled: false }],
  }), /playerIndex must equal 0/);
});

test('seat result decoder fails closed', () => {
  assert.deepEqual(decodeJoinSeatResponse({ gameID: 'GAME', gameName: 'pig', playerIndex: 0 }), {
    gameID: 'GAME', gameName: 'pig', playerIndex: 0,
  });
  assert.throws(() => decodeJoinSeatResponse({ gameID: 'GAME', gameName: 'pig', playerIndex: -1 }), /non-negative/);
});
