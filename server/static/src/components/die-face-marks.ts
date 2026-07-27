/**
 * What a DIE paints inside a facet's content square.
 *
 * The deliberately die-specific half of `boardgame-die.ts`'s geometry. Its
 * counterpart, `src/solid/`, knows how to place a polygon and how big a square
 * fits inside it, and knows nothing about pips or numerals; this module knows
 * about pips and numerals and nothing about polygons. A 3D `boardgame-token`
 * will import the former and none of this.
 *
 * Content resolves in one order, per face: an author-supplied SYMBOL SET first,
 * then generated PIPS, then a NUMERAL. Nothing here enumerates a layout: the die
 * used to stop at six because its pip patterns were six hard-coded CSS classes,
 * and the replacement computes the pattern from the value on a 3x3 lattice.
 *
 * Every length below is a FRACTION of the content square the facet publishes as
 * `--content-size`, never an absolute size, which is what lets one set of
 * numbers draw a legible face on a cube's square, a d20's triangle, a d10's kite
 * and a barrel's 2.7:1 rectangle alike.
 */

/** A pip's cell on the 3x3 lattice: [col, row], each 0..2, +row downward. */
export type PipCell = readonly [number, number];

/**
 * The lattice cells a pip layout is built from, IN THE ORDER THEY ARE ADDED,
 * as opposite pairs. A layout for `n` is the centre cell when `n` is odd
 * followed by the first `floor(n / 2)` pairs -- which reproduces every
 * familiar die and domino face from 0 to 9 without naming any of them:
 *
 *   0 blank; 1 centre; 2 a diagonal; 3 diagonal + centre; 4 the corners;
 *   5 corners + centre; 6 corners + the side midpoints; 7 six + centre;
 *   8 six + top and bottom midpoints; 9 the full lattice.
 */
const PIP_PAIRS: readonly (readonly [PipCell, PipCell])[] = [
  [[0, 0], [2, 2]],
  [[2, 0], [0, 2]],
  [[0, 1], [2, 1]],
  [[1, 0], [1, 2]],
];

const PIP_CENTRE: PipCell = [1, 1];

/**
 * The largest value still drawn as pips.
 *
 * NINE: the 3x3 lattice that physical dice and dominoes use holds exactly
 * nine, and every count up to it has a canonical symmetric pattern on it. A
 * tenth pip needs a fourth row, which both breaks those familiar patterns and
 * shrinks each dot below what reads at the size a facet actually gets (a d10's
 * kite gives its content square about a third of the die's width). Past nine
 * a numeral is smaller to draw AND faster to read — nobody counts ten dots at
 * a glance — so the die switches over.
 */
export const MAX_PIP_VALUE = 9;

/** Pip diameter as a fraction of the content square's side (one lattice cell is a third). */
export const PIP_DIAMETER = 0.2;

/** Numeral/glyph height as a fraction of the content square's side. */
export const GLYPH_HEIGHT = 0.66;

/**
 * How wide the text may run, as a multiple of the content square, divided by
 * its character count: a two-digit numeral is drawn smaller than a one-digit
 * one so that "20" fits the same square "5" does.
 */
const GLYPH_WIDTH_BUDGET = 1.6;

/**
 * Corner marks are drawn a little taller in their (smaller) square, and can
 * afford to be: that square is itself already shrunk by the content margin and
 * tucked against the facet's outline, so the air a face needs is there whether
 * or not the glyph fills its own box.
 */
export const CORNER_GLYPH_HEIGHT = 0.82;

/** The lattice cells for a pip count, computed rather than enumerated. */
export function pipCells(count: number): readonly PipCell[] {
  const cells: PipCell[] = [];
  if (count % 2 === 1) cells.push(PIP_CENTRE);
  for (let index = 0; index < Math.floor(count / 2); index++) {
    cells.push(...PIP_PAIRS[index]);
  }
  return cells;
}

/** True for the values `pipCells` has a canonical lattice pattern for. */
export function isPipValue(value: number): boolean {
  return Number.isInteger(value) && value >= 0 && value <= MAX_PIP_VALUE;
}

/**
 * Font size for a mark, as a fraction of the square it is drawn in: capped by
 * the square's height, and by a width budget that shrinks with the text's
 * length so a three-character label still fits.
 */
export function glyphScale(text: string, heightFraction: number): number {
  return Math.min(heightFraction, GLYPH_WIDTH_BUDGET / Math.max(1, text.length));
}
