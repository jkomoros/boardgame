import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { prismSurface } from './prism.ts';
// The consumer, imported to prove the surface is one it accepts — a `prism.ts`
// that produced a shape `facet-placement.ts` could not draw would pass every
// geometric assertion below and still be useless.
import { solidFacets } from './facet-placement.ts';
import type { Vec3 } from './screen-frame.ts';

const dot = (a: Vec3, b: Vec3): number => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
const sub = (a: Vec3, b: Vec3): Vec3 => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
const cross = (a: Vec3, b: Vec3): Vec3 => [
  a[1] * b[2] - a[2] * b[1],
  a[2] * b[0] - a[0] * b[2],
  a[0] * b[1] - a[1] * b[0],
];

function close(actual: number, expected: number, tolerance: number, message: string): void {
  assert.ok(
    Math.abs(actual - expected) <= tolerance,
    `${message}: expected ${expected}, got ${actual} (delta ${Math.abs(actual - expected)} > ${tolerance})`,
  );
}

/** The three token proportions plus the extremes of the side count. */
const CASES = [
  { name: 'a chunky token', sides: 12, heightRatio: 0.55 },
  { name: 'a poker chip', sides: 12, heightRatio: 0.13 },
  { name: 'a nearly flat disc', sides: 12, heightRatio: 0.04 },
  { name: 'a triangular prism', sides: 3, heightRatio: 1 },
  { name: 'a square prism', sides: 4, heightRatio: 0.5 },
  { name: 'a fine 24-gon', sides: 24, heightRatio: 0.2 },
] as const;

describe('prism', () => {
  for (const { name, sides, heightRatio } of CASES) {
    describe(`${name}: ${sides} sides at ${heightRatio} h/d`, () => {
      const prism = prismSurface(sides, heightRatio);
      // Circumradius 1, so the cross-section is 2 across and the height is
      // twice the ratio. Stated here rather than imported: the analytic side of
      // every check below has to be derived independently of the module.
      const halfHeight = heightRatio;
      const area = (sides / 2) * Math.sin((2 * Math.PI) / sides);

      it('is one polygon per wall plus one per cap', () => {
        assert.equal(prism.faces.length, sides + 2);
        // Flat caps carry the value a token would print, so they are real
        // faces; the field for value-less surface stays empty, as it is for
        // every closed-form solid.
        assert.deepEqual(prism.capFaces, []);
        assert.equal(prism.vertices.length, 2 * sides);
      });

      it('is a closed 2-manifold: every directed edge once, and its reverse present', () => {
        // Keyed on vertex IDENTITY, which the module guarantees by having every
        // polygon reference the same frozen tuples. A key built from
        // coordinates would quietly pass a surface whose walls happened to
        // agree numerically without sharing a corner.
        const index = new Map<Vec3, number>();
        prism.vertices.forEach((vertex, i) => index.set(vertex, i));
        const directed = new Set<string>();
        for (const face of [...prism.faces, ...prism.capFaces]) {
          for (let i = 0; i < face.polygon.length; i++) {
            const from = index.get(face.polygon[i]);
            const to = index.get(face.polygon[(i + 1) % face.polygon.length]);
            assert.ok(from !== undefined && to !== undefined, 'a polygon uses a vertex off the list');
            const key = `${from}->${to}`;
            assert.ok(!directed.has(key), `directed edge ${key} is used twice`);
            directed.add(key);
          }
        }
        // 3N undirected edges (N top, N bottom, N vertical), each used twice.
        assert.equal(directed.size, 6 * sides, 'a prism has 3N edges, so 6N directed uses');
        for (const key of directed) {
          const [from, to] = key.split('->');
          assert.ok(directed.has(`${to}->${from}`), `edge ${key} has no face on its other side`);
        }
      });

      it('points every normal outward, at the angle the closed form says', () => {
        // Independent oracle: the caps face straight along the axis, and wall k
        // faces the direction of its own edge's midpoint.
        const expected: Vec3[] = [[0, 0, 1], [0, 0, -1]];
        for (let k = 0; k < sides; k++) {
          const mid = (2 * Math.PI * (k + 0.5)) / sides;
          expected.push([Math.cos(mid), Math.sin(mid), 0]);
        }
        prism.faces.forEach((face, i) => {
          close(Math.hypot(...face.normal), 1, 1e-15, `face ${i} normal length`);
          for (let axis = 0; axis < 3; axis++) {
            close(face.normal[axis], expected[i][axis], 1e-12, `face ${i} normal axis ${axis}`);
          }
          // The solid is centred on the origin, so outward is also away from it.
          assert.ok(dot(face.normal, face.centroid) > 0, `face ${i} normal points inward`);
        });
      });

      it('keeps every polygon planar and every cap perpendicular to the axis', () => {
        prism.faces.forEach((face, i) => {
          for (const vertex of face.polygon) {
            close(dot(face.normal, sub(vertex, face.centroid)), 0, 1e-12, `face ${i} planarity`);
          }
        });
        const [top, bottom] = prism.faces;
        assert.equal(top.polygon.length, sides);
        assert.equal(bottom.polygon.length, sides);
        for (const vertex of top.polygon) close(vertex[2], halfHeight, 1e-15, 'top cap height');
        for (const vertex of bottom.polygon) close(vertex[2], -halfHeight, 1e-15, 'bottom cap height');
        for (let i = 2; i < prism.faces.length; i++) {
          assert.equal(prism.faces[i].polygon.length, 4, `wall ${i - 2} is not a quad`);
        }
      });

      it('encloses the analytic prism volume', () => {
        // Divergence theorem over the surface as returned. The analytic value
        // is the regular polygon's area times the height, with no pi/2 sleight
        // of hand: (N/2) R^2 sin(2 pi / N) times 2 * halfHeight.
        let volume = 0;
        for (const face of [...prism.faces, ...prism.capFaces]) {
          for (let i = 1; i + 1 < face.polygon.length; i++) {
            volume += dot(face.polygon[0], cross(face.polygon[i], face.polygon[i + 1])) / 6;
          }
        }
        const expected = area * 2 * halfHeight;
        close(volume, expected, expected * 1e-14, `${name} volume`);
      });

      it('normalizes by its width and reports the diagonal it overflows by', () => {
        // `1em` is the cross-section, which is what "size by drawn extent"
        // needs: a chip seen face-on fills the box exactly. The bounding sphere
        // is bigger, always, and a caller that spins one has to know by how much.
        assert.equal(prism.nominalRadius, 1);
        close(prism.boundingRadius, Math.hypot(1, halfHeight), 1e-15, 'bounding radius');
        assert.ok(prism.boundingRadius > prism.nominalRadius, 'a prism is longer on the diagonal');
        for (const vertex of prism.vertices) {
          assert.ok(
            Math.hypot(...vertex) <= prism.boundingRadius + 1e-12,
            'a vertex lies outside the bounding radius',
          );
          close(Math.hypot(vertex[0], vertex[1]), 1, 1e-15, 'a corner is off the circumcircle');
        }
      });

      it('is a surface facet-placement can draw', () => {
        const facets = solidFacets(prism);
        assert.equal(facets.length, sides + 2);
        for (const facet of facets) {
          assert.match(facet.style, /transform:translate3d\(/);
          assert.match(facet.style, /clip-path:polygon\(/);
          assert.ok(facet.contentSize > 0, 'a facet has no room for content');
        }
      });
    });
  }

  it('draws a 12-side prism as exactly 14 facets', () => {
    // Pinned on its own rather than left implicit in `sides + 2`, because the
    // whole facet budget is this number: 55 tokens on screen at 24 sides is
    // 1,430 facet elements and 42.8fps at rest, and at 12 sides it is 770
    // elements and 60fps. ~800 elements is free, ~1,400 is a cliff. A change
    // that makes this 15 costs 55 elements a board.
    const prism = prismSurface(12, 0.55);
    assert.equal(prism.faces.length + prism.capFaces.length, 14);
    assert.equal(solidFacets(prism).length, 14);
  });

  it('rejects a side count or a height that is not a prism', () => {
    for (const sides of [2, 1, 0, -12, 3.5, NaN, Infinity]) {
      assert.throws(() => prismSurface(sides, 0.5), /side count/, `sides ${sides}`);
    }
    for (const ratio of [0, -0.5, NaN, Infinity]) {
      // Zero height is the tempting reading of "a disc is flat" and it is a
      // degenerate solid: no side walls with any area, and no enclosed volume.
      assert.throws(() => prismSurface(12, ratio), /height ratio/, `ratio ${ratio}`);
    }
  });
});
