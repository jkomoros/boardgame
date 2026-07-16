import assert from 'node:assert/strict';
import test from 'node:test';

import { isVisibleComponent } from './boardgame-types.ts';

test('isVisibleComponent accepts visible components with and without static or dynamic values', () => {
  const visible = {
    Index: 2,
    Values: { Suit: 'Hearts' },
    Deck: 'cards',
    GameName: 'blackjack',
    ID: 'card-2',
  };
  assert.equal(isVisibleComponent(visible), true);
  assert.equal(isVisibleComponent({ ...visible, Values: {} }), true);
  assert.equal(isVisibleComponent({ ...visible, DynamicValues: { FaceUp: true } }), true);
  assert.equal(isVisibleComponent({ ...visible, Values: Object.create(null) }), true);
});

test('isVisibleComponent rejects opaque, empty, null, and malformed entries', () => {
  const visible = {
    Index: 2,
    Values: { Suit: 'Hearts' },
    Deck: 'cards',
    GameName: 'blackjack',
    ID: 'card-2',
  };
  assert.equal(isVisibleComponent({}), false);
  assert.equal(isVisibleComponent(null), false);
  assert.equal(isVisibleComponent({ ...visible, Index: Number.NaN }), false);
  assert.equal(isVisibleComponent({ ...visible, Index: Number.POSITIVE_INFINITY }), false);
  assert.equal(isVisibleComponent({ ...visible, Index: 1.5 }), false);
  assert.equal(isVisibleComponent({ ...visible, Index: -1 }), false);
  assert.equal(isVisibleComponent({ ...visible, Values: [] }), false);
  assert.equal(isVisibleComponent({ ...visible, Values: new Date() }), false);
  assert.equal(isVisibleComponent({ ...visible, DynamicValues: [] }), false);
  assert.equal(isVisibleComponent({ ...visible, DynamicValues: null }), false);
  assert.equal(isVisibleComponent({ ...visible, Values: undefined }), false);
  assert.equal(isVisibleComponent({ ...visible, ID: undefined }), false);
});
