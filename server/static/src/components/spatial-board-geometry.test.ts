import assert from 'node:assert/strict';
import test from 'node:test';
import type { ExpandedStack } from '../types/boardgame-types.ts';
import { piecesFromSizedStacks, rasterArtworkScene, rasterBoardArtwork } from './spatial-board-geometry.ts';

function stack(ids: readonly string[], components: ExpandedStack<object, object>['Components']): ExpandedStack<object, object> {
  return {
    Deck: 'tokens',
    Indexes: components.map((_, index) => index),
    IDs: ids,
    IDsLastSeen: {},
    ShuffleCount: 0,
    GameName: 'fixture',
    Components: components,
  };
}

test('piecesFromSizedStacks creates explicit stable piece-to-space projections', () => {
  const token = { Index: 2, Values: {}, Deck: 'tokens', GameName: 'fixture', ID: 'token-2' };
  const source = stack(['', 'token-2', ''], [null, token, null]);
  const pieces = piecesFromSizedStacks([source], ['hall', 'library', 'study'] as const);
  assert.deepEqual(pieces, [{
    id: 'token-2',
    space: 'library',
    stack: source,
    slot: 1,
    component: token,
  }]);
  assert.ok(Object.isFrozen(pieces));
  assert.ok(Object.isFrozen(pieces[0]));
});

test('piecesFromSizedStacks makes sentinel slots explicit without fake geometry', () => {
  const hidden = { Index: 0, Values: {}, Deck: 'tokens', GameName: 'fixture', ID: 'hidden' };
  const visible = { Index: 1, Values: {}, Deck: 'tokens', GameName: 'fixture', ID: 'visible' };
  const source = stack(['hidden', 'visible'], [hidden, visible]);
  assert.deepEqual(
    piecesFromSizedStacks([source], [null, 'library']),
    [{ id: 'visible', space: 'library', stack: source, slot: 1, component: visible }],
  );
});

test('piecesFromSizedStacks rejects cardinality and stable-ID mismatches loudly', () => {
  assert.throws(
    () => piecesFromSizedStacks([stack([''], [null])], ['hall', 'study']),
    /1 slots but 2 space keys/,
  );
  const token = { Index: 0, Values: {}, Deck: 'tokens', GameName: 'fixture', ID: 'catalog-id' };
  assert.throws(
    () => piecesFromSizedStacks([stack([''], [token])], ['hall']),
    /occupied slot 0 has no stable ID/,
  );
});

test('raster artwork freezes normalized geometry and rejects ambiguous inputs', () => {
  const artwork = rasterBoardArtwork({
    src: '/board.webp',
    fit: 'cover',
    viewportAspectRatio: 16 / 9,
    spaces: [
      {
        key: 'library',
        label: 'Library',
        group: 'rooms',
        order: 2,
        region: { shape: 'circle', center: { x: 0.25, y: 0.4 }, radius: 0.08 },
        pieceAnchor: { x: 0.3, y: 0.45 },
      },
      {
        key: 'hall',
        label: 'Hall',
        order: 1,
        region: { shape: 'rect', x: 0.5, y: 0.1, width: 0.3, height: 0.2 },
      },
    ],
  });
  assert.equal(artwork.kind, 'raster');
  assert.equal(artwork.fit, 'cover');
  assert.ok(Object.isFrozen(artwork));
  assert.ok(Object.isFrozen(artwork.spaces));
  assert.ok(Object.isFrozen(artwork.spaces[0]?.region));
  assert.equal(artwork.spaces[0]?.group, 'rooms');
  assert.throws(() => rasterBoardArtwork({ src: 'javascript:alert(1)', spaces: artwork.spaces }), /forbidden URL protocol/);
  assert.throws(() => rasterBoardArtwork({ src: 'data:image\/svg+xml,<svg\/>', spaces: artwork.spaces }), /supported raster image/);
  assert.throws(() => rasterBoardArtwork({ src: '/board.png', spaces: [{
    key: 'bad', label: 'Bad', region: { shape: 'rect', x: 0.8, y: 0, width: 0.3, height: 0.2 },
  }] }), /fit within normalized coordinates/);
  assert.throws(() => rasterBoardArtwork({ src: '/board.png', spaces: [{
    key: 'bad', label: 'Bad', region: { shape: 'polygon', points: [{ x: 0, y: 0 }, { x: 1, y: 0 }, { x: 0.5, y: 0 }] },
  }] }), /positive width and height/);
  assert.throws(() => rasterBoardArtwork({ src: '/board.png', spaces: [{
    key: 'bad', label: 'Bad', region: { shape: 'polygon', points: [{ x: 0, y: 0 }, { x: 0.5, y: 0.5 }, { x: 1, y: 1 }] },
  }] }), /positive area/);
  assert.throws(() => rasterBoardArtwork({ src: '/board.png', spaces: [artwork.spaces[0]!, artwork.spaces[0]!] }), /duplicate canonical/);
  assert.throws(() => rasterBoardArtwork({ src: '/board.png', spaces: [{
    ...artwork.spaces[0]!, group: ' rooms',
  }] }), /surrounding whitespace/);
  assert.doesNotThrow(() => rasterBoardArtwork({ src: '/board.png', spaces: [
    { ...artwork.spaces[0]!, key: 'tile', group: 'tiles', order: 0 },
    { ...artwork.spaces[0]!, key: 'vertex', group: 'vertices', order: 0 },
  ] }));
  assert.throws(() => rasterBoardArtwork({ src: '/board.png', spaces: [
    { ...artwork.spaces[0]!, key: 'tile-a', group: 'tiles', order: 0 },
    { ...artwork.spaces[0]!, key: 'tile-b', group: 'tiles', order: 0 },
  ] }), /duplicate keyboard order 0 in group "tiles"/);
  assert.throws(() => rasterArtworkScene(artwork, 20_000, 100), /dimension or .*area limit/);
  assert.throws(() => rasterArtworkScene(artwork, 10_001, 10_001), /dimension or .*area limit/);
  const overflowing = rasterBoardArtwork({
    src: '/board.png', spaces: artwork.spaces, viewportAspectRatio: Number.MAX_VALUE,
  });
  assert.throws(() => rasterArtworkScene(overflowing, 100, 100), /viewportAspectRatio overflows/);
});
