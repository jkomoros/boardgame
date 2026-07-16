import assert from 'node:assert/strict';
import test from 'node:test';
import {
  decodeJoinResponse,
  decodeJoinProblem,
  decodeJoinSeatResponse,
  decodeSeatOptionsResponse,
} from './join-response.ts';

test('join decoder enforces bounded player counts and copies the wire contract', () => {
  assert.deepEqual(decodeJoinResponse({
    gameID: 'GAME', gameName: 'pig', gameDisplayName: 'Pig',
    minPlayers: 2, maxPlayers: 4, currentPlayers: 1, availableSeats: 3, requiresSeatPicker: false, joinTicket: 'ticket',
    secret: 'discard',
  }), {
    gameID: 'GAME', gameName: 'pig', gameDisplayName: 'Pig',
    minPlayers: 2, maxPlayers: 4, currentPlayers: 1, availableSeats: 3, requiresSeatPicker: false, joinTicket: 'ticket',
  });
  assert.throws(() => decodeJoinResponse({
    gameID: 'GAME', gameName: 'pig', gameDisplayName: 'Pig',
    minPlayers: 4, maxPlayers: 2, currentPlayers: 1, availableSeats: 1, requiresSeatPicker: false, joinTicket: 'ticket',
  }), /minPlayers must not exceed/);
});

test('seat options require canonical contiguous indexes and exact slot values', () => {
  const decoded = decodeSeatOptionsResponse({
    gameID: 'GAME', gameName: 'werewolf', requiresSeatPicker: true,
    slots: [{ playerIndex: 0, label: 'Seat 1', status: 'human', filled: true, available: false, displayName: 'Ada' }],
  });
  assert.equal(decoded.slots[0]?.displayName, 'Ada');
  assert.throws(() => decodeSeatOptionsResponse({
    gameID: 'GAME', gameName: 'werewolf', requiresSeatPicker: true,
    slots: [{ playerIndex: 2, label: 'Seat 1', status: 'open', filled: false, available: true }],
  }), /playerIndex must equal 0/);
});

test('seat result decoder fails closed', () => {
  assert.deepEqual(decodeJoinSeatResponse({ gameID: 'GAME', gameName: 'pig', playerIndex: 0, resumed: true }), {
    gameID: 'GAME', gameName: 'pig', playerIndex: 0, resumed: true,
  });
  assert.throws(() => decodeJoinSeatResponse({ gameID: 'GAME', gameName: 'pig', playerIndex: -1, resumed: false }), /non-negative/);
});

test('typed join conflicts preserve canonical seat snapshots', () => {
  const problem = decodeJoinProblem({
    code: 'SEAT_TAKEN', error: 'That seat is no longer available',
    slots: [{ playerIndex: 0, label: 'Seat 1', status: 'agent', filled: true, available: false }],
  });
  assert.equal(problem.code, 'SEAT_TAKEN');
  assert.equal(problem.slots?.[0]?.status, 'agent');
});

test('seat status, filled, and availability cannot contradict each other', () => {
  for (const slot of [
    { playerIndex: 0, label: 'Seat 1', status: 'open', filled: true, available: true },
    { playerIndex: 0, label: 'Seat 1', status: 'human', filled: false, available: true },
    { playerIndex: 0, label: 'Seat 1', status: 'closed', filled: true, available: false },
  ]) {
    assert.throws(() => decodeSeatOptionsResponse({
      gameID: 'GAME', gameName: 'game', requiresSeatPicker: true, slots: [slot],
    }), /contradicted status/);
  }
});
