import assert from 'node:assert/strict';
import test from 'node:test';
import {
  fallbackPlayerPresentation,
  playerPresentations,
  validatePlayerPresentations,
} from './player-presentation.ts';

test('player presentations normalize public labels and remain immutable', () => {
  const result = playerPresentations([
    { DisplayName: ' Alice ' },
    {},
  ], ['#123456', '']);
  assert.deepEqual(result, [
    { playerIndex: 0, label: 'Alice', color: '#123456' },
    { playerIndex: 1, label: 'Player 2' },
  ]);
  assert.equal(Object.isFrozen(result), true);
  assert.equal(Object.isFrozen(result[0]), true);
});

test('player presentation normalization is bounded and indexes fail loudly', () => {
  assert.throws(
    () => playerPresentations(Array.from({ length: 129 }, () => ({})), []),
    /maximum is 128/,
  );
  assert.throws(
    () => playerPresentations([{ DisplayName: 'x'.repeat(201) }], []),
    /exceeds 200 characters/,
  );
  assert.throws(() => fallbackPlayerPresentation(-1), /non-negative safe integer/);
});

test('validatePlayerPresentations copies canonical contiguous presentations', () => {
  const source = [{ playerIndex: 0, label: '  Ada  ', color: ' red ' }];
  const result = validatePlayerPresentations(source);
  assert.deepEqual(result, [{ playerIndex: 0, label: 'Ada', color: 'red' }]);
  assert.notEqual(result, source);
  assert.ok(Object.isFrozen(result));
  assert.ok(Object.isFrozen(result[0]));
  assert.throws(
    () => validatePlayerPresentations([{ playerIndex: 1, label: 'Ada' }]),
    /must have playerIndex 0/,
  );
  assert.throws(
    () => validatePlayerPresentations([{ playerIndex: 0, label: '   ' }]),
    /non-empty string/,
  );
});
