import test from 'node:test';
import assert from 'node:assert';

import { flatFacetStyle, type FlatPoint } from './flat-facets.ts';

/** Every declaration in a style string, as a map. */
function declarations(style: string): Record<string, string> {
  const out: Record<string, string> = {};
  // clip-path's own value contains no semicolons (only commas and spaces), so
  // a plain split is safe here and stays readable.
  for (const part of style.split(';')) {
    const colon = part.indexOf(':');
    out[part.slice(0, colon)] = part.slice(colon + 1);
  }
  return out;
}

/** The `clip-path: polygon(...)` points, as percentages. */
function polygonPoints(style: string): { x: number; y: number }[] {
  const match = /clip-path:polygon\(([^)]*)\)/.exec(style);
  assert.ok(match, `no clip path in ${style}`);
  return match[1].split(',').map((pair) => {
    const [x, y] = pair.trim().split(/\s+/);
    return { x: Number(x.replace('%', '')), y: Number(y.replace('%', '')) };
  });
}

const SQUARE: readonly FlatPoint[] = [
  { x: -0.5, y: -0.25 }, { x: 0.5, y: -0.25 }, { x: 0.5, y: 0.25 }, { x: -0.5, y: 0.25 },
];

test('the element is the polygon\'s own bounding box, placed from the centre', () => {
  const style = declarations(flatFacetStyle(SQUARE));
  assert.equal(style.position, 'absolute');
  assert.equal(style.left, '50%');
  assert.equal(style.top, '50%');
  assert.equal(style.width, '1em');
  assert.equal(style.height, '0.5em');
  // The negative margins put the box's own min corner at (minX, minY) from the
  // solid's centre, which is where the projection measured it.
  assert.equal(style['margin-left'], '-0.5em');
  assert.equal(style['margin-top'], '-0.25em');
});

test('an off-centre polygon keeps its offset in the margins, not in the box', () => {
  // A silhouette that is off-centre sticks out on one side only, and the box has
  // to follow it rather than being centred and grown -- a centred box would be
  // twice the size and would clip nothing.
  const style = declarations(flatFacetStyle([
    { x: 0.25, y: 0.5 }, { x: 0.75, y: 0.5 }, { x: 0.5, y: 1.5 },
  ]));
  assert.equal(style.width, '0.5em');
  assert.equal(style.height, '1em');
  assert.equal(style['margin-left'], '0.25em');
  assert.equal(style['margin-top'], '0.5em');
});

test('the outermost vertices land exactly on the box, so nothing is shaved', () => {
  // A clip-path is clipped by the element's own border box. If the polygon could
  // exceed the box the silhouette would be quietly trimmed on one side, which is
  // the failure a box that merely happens to be big enough eventually produces.
  for (const points of [SQUARE, [
    { x: -0.3, y: -0.9 }, { x: 0.8, y: -0.1 }, { x: 0.2, y: 0.7 }, { x: -0.7, y: 0.4 },
  ]]) {
    const percentages = polygonPoints(flatFacetStyle(points));
    const xs = percentages.map((p) => p.x);
    const ys = percentages.map((p) => p.y);
    assert.equal(Math.min(...xs), 0);
    assert.equal(Math.max(...xs), 100);
    assert.equal(Math.min(...ys), 0);
    assert.equal(Math.max(...ys), 100);
    for (const p of percentages) {
      assert.ok(p.x >= 0 && p.x <= 100 && p.y >= 0 && p.y <= 100, JSON.stringify(p));
    }
  }
});

test('the polygon keeps its vertex ORDER, and so its winding', () => {
  // clip-path uses a nonzero fill rule, so a reversed loop still fills -- but the
  // caller hands these over counter-clockwise-seen-from-outside and a shuffled
  // order would draw a self-intersecting bow tie rather than the facet.
  const points: FlatPoint[] = [
    { x: -1, y: -1 }, { x: 1, y: -1 }, { x: 1, y: 1 }, { x: -1, y: 1 },
  ];
  assert.deepEqual(polygonPoints(flatFacetStyle(points)), [
    { x: 0, y: 0 }, { x: 100, y: 0 }, { x: 100, y: 100 }, { x: 0, y: 100 },
  ]);
});

test('there is no transform, and nothing else that could take a layer', () => {
  // The whole reason this module exists. A transform, a perspective, a
  // will-change or a backface-visibility on a facet is a composited layer per
  // facet the moment an ancestor animates -- 1,047 of them for 55 tokens.
  const style = flatFacetStyle(SQUARE);
  for (const forbidden of ['transform', 'perspective', 'will-change', 'backface']) {
    assert.ok(!style.includes(forbidden), `${forbidden} in ${style}`);
  }
});

test('a degenerate polygon is refused rather than drawn as a hairline', () => {
  // An edge-on facet has no area to draw and is the caller's to cull. Emitting
  // a zero-width box would divide by zero building the percentages and produce
  // `NaN%`, which a browser drops silently -- so the facet would vanish and
  // nothing would say why.
  assert.throws(() => flatFacetStyle([{ x: 0, y: 0 }, { x: 1, y: 0 }, { x: 2, y: 0 }]),
    /collapsed to a line/);
  assert.throws(() => flatFacetStyle([{ x: 0, y: 0 }, { x: 0, y: 1 }]), /needs 3 points/);
  assert.throws(() => flatFacetStyle([
    { x: 0, y: 0 }, { x: NaN, y: 1 }, { x: 1, y: 1 },
  ]), /not finite/);
});
