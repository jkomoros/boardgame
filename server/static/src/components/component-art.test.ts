import assert from 'node:assert/strict';
import test from 'node:test';
import { artLayerStyle, cssUrl, isArtFit } from './component-art.ts';

/**
 * The escaping and the layer order, which is all of this module that can be
 * wrong without a browser.
 *
 * What is checked here is deliberately the part the four hand-rolled
 * `background:` shorthands got away with because their filenames happened to be
 * boring.
 */

test('a plain URL becomes a quoted url() token', () => {
  assert.equal(cssUrl('./assets/card-back.png'), 'url("./assets/card-back.png")');
});

test('a quote in the URL cannot end the CSS string', () => {
  // Unescaped, this closes the string and the rest of the declaration is
  // garbage -- which is exactly how a background silently stops painting.
  assert.equal(cssUrl('a"b.png'), 'url("a\\"b.png")');
  assert.equal(cssUrl('a\\b.png'), 'url("a\\\\b.png")');
});

test('a newline cannot smuggle a second declaration in', () => {
  assert.equal(cssUrl('a\nb.png'), 'url("ab.png")');
  assert.equal(cssUrl('a\r\nb.png'), 'url("ab.png")');
});

test('a dev-server cache-buster survives intact', () => {
  // The shape every Vite-served asset actually has.
  assert.equal(
    cssUrl('/game-src/darwin/assets/card-back.png?t=1723200000'),
    'url("/game-src/darwin/assets/card-back.png?t=1723200000")',
  );
});

test('the default layer is the art, centred, covering and never tiled', () => {
  assert.equal(
    artLayerStyle('mat.png'),
    'background: url("mat.png") center / cover no-repeat;',
  );
});

test('contain is honoured and is the only other fit', () => {
  assert.equal(
    artLayerStyle('chip.png', 'contain'),
    'background: url("chip.png") center / contain no-repeat;',
  );
  assert.ok(isArtFit('cover'));
  assert.ok(isArtFit('contain'));
  assert.ok(!isArtFit('fill'));
  assert.ok(!isArtFit(''));
  assert.ok(!isArtFit(undefined));
});

test('the wash is a flat sheet ABOVE the art, not below it', () => {
  // Two things at once. Order: CSS paints the first background layer on top, so
  // a wash listed second would be underneath the photograph and invisible.
  // Placement: the shorthand applies `center / cover no-repeat` only to the
  // layer it is written on, so BOTH layers have to carry it.
  assert.equal(
    artLayerStyle('habitat-mat.png', 'cover', '#f4eddde8'),
    'background: linear-gradient(#f4eddde8, #f4eddde8) center / cover no-repeat, '
    + 'url("habitat-mat.png") center / cover no-repeat;',
  );
});

test('an empty or non-string art is refused rather than emitting a broken shorthand', () => {
  assert.throws(() => artLayerStyle(''), /non-empty URL/);
  assert.throws(() => artLayerStyle(undefined as unknown as string), /non-empty URL/);
});

test('an unknown fit is refused by name', () => {
  assert.throws(
    () => artLayerStyle('a.png', 'fill' as never),
    /must be "cover" or "contain", not "fill"/,
  );
});
