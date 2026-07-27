import assert from 'node:assert/strict';
import test from 'node:test';
import {
  CORNER_GLYPH_HEIGHT,
  GLYPH_HEIGHT,
  MAX_PIP_VALUE,
  PIP_DIAMETER,
  glyphScale,
  isPipValue,
  pipCells,
} from './die-face-marks.ts';

const key = ([col, row]: readonly [number, number]) => `${col},${row}`;

test('a pip layout has exactly as many dots as the value', () => {
  for (let value = 0; value <= MAX_PIP_VALUE; value++) {
    const cells = pipCells(value);
    assert.equal(cells.length, value, `value ${value}`);
    assert.equal(new Set(cells.map(key)).size, value, `value ${value}: no cell used twice`);
    for (const [col, row] of cells) {
      assert.ok(Number.isInteger(col) && col >= 0 && col <= 2, `value ${value}: col ${col}`);
      assert.ok(Number.isInteger(row) && row >= 0 && row <= 2, `value ${value}: row ${row}`);
    }
  }
});

/**
 * Every familiar die and domino face is symmetric under a half turn, and that
 * is the property that makes a pipped face read the same whichever way up the
 * facet is seen. It is also what the "opposite pairs plus a centre" scheme is
 * FOR — an implementation that added cells in some other order could still
 * produce the right counts and fail here.
 */
test('every pip layout is symmetric under a half turn', () => {
  for (let value = 0; value <= MAX_PIP_VALUE; value++) {
    const cells = pipCells(value);
    const present = new Set(cells.map(key));
    for (const [col, row] of cells) {
      assert.ok(present.has(key([2 - col, 2 - row])),
        `value ${value}: (${col}, ${row}) has no opposite`);
    }
  }
});

test('the canonical faces come out canonical', () => {
  const sorted = (value: number) => pipCells(value).map(key).sort().join(' ');
  assert.equal(sorted(0), '');
  assert.equal(sorted(1), '1,1');
  assert.equal(sorted(2), '0,0 2,2');
  assert.equal(sorted(3), '0,0 1,1 2,2');
  assert.equal(sorted(4), '0,0 0,2 2,0 2,2');
  assert.equal(sorted(5), '0,0 0,2 1,1 2,0 2,2');
  // Six is the corners plus the SIDE midpoints, which is what a physical d6
  // draws, and not the top/bottom midpoints.
  assert.equal(sorted(6), '0,0 0,1 0,2 2,0 2,1 2,2');
  assert.equal(sorted(7), '0,0 0,1 0,2 1,1 2,0 2,1 2,2');
  assert.equal(sorted(8), '0,0 0,1 0,2 1,0 1,2 2,0 2,1 2,2');
  assert.equal(sorted(9), '0,0 0,1 0,2 1,0 1,1 1,2 2,0 2,1 2,2');
});

test('isPipValue admits exactly the values with a lattice pattern', () => {
  for (let value = 0; value <= MAX_PIP_VALUE; value++) assert.ok(isPipValue(value), `${value}`);
  assert.ok(!isPipValue(MAX_PIP_VALUE + 1), 'a tenth pip needs a fourth row');
  assert.ok(!isPipValue(-1));
  assert.ok(!isPipValue(2.5));
  assert.ok(!isPipValue(Number.NaN));
  assert.ok(!isPipValue(Number.POSITIVE_INFINITY));
});

test('the pip lattice fits inside the content square', () => {
  // Cell centres sit at a sixth, a half and five sixths of the square, so a dot
  // of diameter PIP_DIAMETER must not reach past either edge.
  assert.ok(PIP_DIAMETER / 2 < 1 / 6, 'a dot at the first cell would overflow the square');
  // And two dots in adjacent cells (a third of the square apart) must not touch.
  assert.ok(PIP_DIAMETER < 1 / 3, 'adjacent dots would merge');
});

/**
 * A numeral is sized so that the widest label a face can carry still fits the
 * square. Without the width budget a "20" set at the one-character height runs
 * off both ends of a d20's triangle.
 */
test('glyphScale trades height for width as the text gets longer', () => {
  assert.equal(glyphScale('5', GLYPH_HEIGHT), GLYPH_HEIGHT, 'one character is height-bound');
  assert.equal(glyphScale('20', GLYPH_HEIGHT), GLYPH_HEIGHT, 'two still are');
  const three = glyphScale('100', GLYPH_HEIGHT);
  assert.ok(three < GLYPH_HEIGHT, 'three characters are width-bound');
  assert.ok(Math.abs(three - 1.6 / 3) < 1e-12);
  // Monotonic: a longer label is never drawn larger.
  let previous = Infinity;
  for (const text of ['1', '12', '123', '1234', '12345']) {
    const scale = glyphScale(text, GLYPH_HEIGHT);
    assert.ok(scale <= previous, `${text} is not smaller than the shorter label`);
    previous = scale;
  }
  // Empty text must not divide by zero or blow the square up.
  assert.equal(glyphScale('', GLYPH_HEIGHT), GLYPH_HEIGHT);
});

test('a corner mark is drawn taller in its own smaller square', () => {
  assert.ok(CORNER_GLYPH_HEIGHT > GLYPH_HEIGHT);
  assert.ok(CORNER_GLYPH_HEIGHT < 1, 'and still inside it');
  assert.equal(glyphScale('4', CORNER_GLYPH_HEIGHT), CORNER_GLYPH_HEIGHT);
});
