import assert from 'node:assert/strict';
import test from 'node:test';
import {
  add,
  dieGeometry,
  dot,
  magnitude,
  normalize,
  scale as scaleVec,
  subtract,
  vec3,
  type Vec3,
} from '../motion/die-geometry.ts';
import { toScreen } from './screen-frame.ts';
import {
  facetPlacement,
  inscribedSquareHalfSide,
  solidFacets,
  type PlanePoint,
  type SolidFace,
} from './facet-placement.ts';

/** Every face count the die pipeline builds a closed-form or barrel solid for. */
const FACE_COUNTS = [4, 5, 6, 7, 8, 10, 12, 20] as const;

interface ParsedStyle {
  readonly width: number;
  readonly height: number;
  readonly translate: Vec3;
  /** The matrix3d columns: the facet's local x, local y and outward normal. */
  readonly u: Vec3;
  readonly v: Vec3;
  readonly w: Vec3;
  /** clip-path vertices, as fractions of the box in [0, 1]. */
  readonly clip: readonly (readonly [number, number])[];
  readonly vars: Readonly<Record<string, number>>;
}

/**
 * Read a generated facet style back apart.
 *
 * The point of parsing rather than snapshotting is that every assertion below is
 * then a GEOMETRIC one — "this percentage lands on that vertex" — which survives
 * a change of formatting and fails on a change of meaning.
 */
function parseStyle(style: string): ParsedStyle {
  const declarations = new Map<string, string>();
  for (const piece of style.split(';')) {
    const colon = piece.indexOf(':');
    declarations.set(piece.slice(0, colon), piece.slice(colon + 1));
  }
  const em = (name: string) => {
    const raw = declarations.get(name);
    assert.ok(raw !== undefined && raw.endsWith('em'), `${name} must be an em length, got ${raw}`);
    return Number.parseFloat(raw);
  };
  const transform = declarations.get('transform') ?? '';
  const translate = /translate3d\(([^)]*)\)/.exec(transform);
  const matrix = /matrix3d\(([^)]*)\)/.exec(transform);
  assert.ok(translate && matrix, `transform must be translate3d + matrix3d, got ${transform}`);
  const t = translate[1].split(',').map((piece) => {
    assert.ok(piece.trim().endsWith('em'), `translate3d components are em, got ${piece}`);
    return Number.parseFloat(piece);
  });
  const m = matrix[1].split(',').map(Number);
  assert.equal(m.length, 16);
  const clipText = /clip-path:polygon\(([^)]*)\)/.exec(style);
  assert.ok(clipText, 'a facet must carry a clip-path polygon');
  const clip = clipText[1].split(',').map((pair) => {
    const [a, b] = pair.trim().split(/\s+/);
    assert.ok(a.endsWith('%') && b.endsWith('%'), `clip vertices are percentages, got ${pair}`);
    return [Number.parseFloat(a) / 100, Number.parseFloat(b) / 100] as const;
  });
  const vars: Record<string, number> = {};
  for (const [name, value] of declarations) {
    if (name.startsWith('--content-')) vars[name] = Number.parseFloat(value);
  }
  return {
    width: em('width'),
    height: em('height'),
    translate: vec3(t[0], t[1], t[2]),
    u: vec3(m[0], m[1], m[2]),
    v: vec3(m[4], m[5], m[6]),
    w: vec3(m[8], m[9], m[10]),
    clip,
    vars,
  };
}

/**
 * Where the point at clip-path fraction `(px, py)` of the box actually lands in
 * the parent's frame, by doing what the browser does with the emitted style: the
 * box's centre is at `translate`, its local axes are the matrix columns, and the
 * box spans `width x height` about that centre (the negative margins move its
 * top-left corner off the 50%/50% anchor by half its size).
 */
function clipPointInParent(parsed: ParsedStyle, px: number, py: number): Vec3 {
  return add(
    add(parsed.translate, scaleVec(parsed.u, (px - 0.5) * parsed.width)),
    scaleVec(parsed.v, (py - 0.5) * parsed.height),
  );
}

/**
 * THE CORRESPONDENCE THAT MAKES ONE CODE PATH ENOUGH.
 *
 * A facet is a rectangular box cut down by a `clip-path`. Nothing checks that
 * the cut lands on the actual polygon except this: the i-th clip-path vertex,
 * transformed by the very `transform` the same call emitted, must land on the
 * i-th polygon vertex of the source face, in CSS space and at the solid's
 * rendered scale. A basis mixed up, a clip axis swapped, a winding reversed, an
 * off-by-one in the vertex loop — each of those keeps the silhouette plausible
 * and fails here.
 */
test('every clip-path vertex lands on the polygon vertex it came from', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const unitsToEm = 0.5 / geometry.circumradius;
    for (const face of [...geometry.faces, ...geometry.capFaces]) {
      const parsed = parseStyle(facetPlacement(face, unitsToEm, null).style);
      assert.equal(parsed.clip.length, face.polygon.length,
        `d${faceCount}: one clip vertex per polygon vertex`);
      face.polygon.forEach((vertex, index) => {
        const expected = scaleVec(toScreen(vertex), unitsToEm);
        const [px, py] = parsed.clip[index];
        const actual = clipPointInParent(parsed, px, py);
        // 5-decimal rounding on ~1em values; 1e-4 is two orders of slack.
        assert.ok(magnitude(subtract(actual, expected)) < 1e-4,
          `d${faceCount} vertex ${index}: ${JSON.stringify(actual)} vs ${JSON.stringify(expected)}`);
      });
    }
  }
});

test('the emitted matrix orients the facet along its own outward normal', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const unitsToEm = 0.5 / geometry.circumradius;
    for (const face of [...geometry.faces, ...geometry.capFaces]) {
      const parsed = parseStyle(facetPlacement(face, unitsToEm, null).style);
      const expected = normalize(toScreen(face.normal));
      assert.ok(magnitude(subtract(parsed.w, expected)) < 1e-4, `d${faceCount} normal`);
      // Still a rotation after rounding: a mirrored facet draws its content
      // backwards while leaving the outline identical.
      assert.ok(dot(parsed.u, cross3(parsed.v, parsed.w)) > 0.999, `d${faceCount} handedness`);
    }
  }
});

function cross3(a: Vec3, b: Vec3): Vec3 {
  return vec3(a[1] * b[2] - a[2] * b[1], a[2] * b[0] - a[0] * b[2], a[0] * b[1] - a[1] * b[0]);
}

/**
 * The box is the polygon's BOUNDING rectangle, not a square and not centred on
 * the centroid. If it were bigger the facet would have a transparent skirt that
 * still catches `backface-visibility` and hover; if it were smaller the
 * clip-path would run outside the box and the facet would be silently cropped.
 */
test('the box is exactly the polygon bounding box: the clip touches all four sides', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const unitsToEm = 0.5 / geometry.circumradius;
    for (const face of [...geometry.faces, ...geometry.capFaces]) {
      const { clip } = parseStyle(facetPlacement(face, unitsToEm, null).style);
      const xs = clip.map((p) => p[0]);
      const ys = clip.map((p) => p[1]);
      for (const value of [...xs, ...ys]) {
        assert.ok(value >= -1e-9 && value <= 1 + 1e-9, `d${faceCount}: clip stays inside the box`);
      }
      assert.ok(Math.min(...xs) < 1e-9 && Math.max(...xs) > 1 - 1e-9, `d${faceCount}: box hugs x`);
      assert.ok(Math.min(...ys) < 1e-9 && Math.max(...ys) > 1 - 1e-9, `d${faceCount}: box hugs y`);
    }
  }
});

/**
 * A mark sized to the content square must not be able to leave the facet. The
 * square is published as percentages of the box, so this checks the published
 * numbers rather than the internals that produced them.
 */
test('the content square is inside the polygon on every solid', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const unitsToEm = 0.5 / geometry.circumradius;
    for (const face of geometry.faces) {
      const parsed = parseStyle(facetPlacement(face, unitsToEm, null).style);
      const left = parsed.vars['--content-left'] / 100;
      const top = parsed.vars['--content-top'] / 100;
      const width = parsed.vars['--content-width'] / 100;
      const height = parsed.vars['--content-height'] / 100;
      assert.ok(width > 0 && height > 0, `d${faceCount}: content square has area`);
      // It is a SQUARE in the facet's plane even though its percentages differ:
      // width% of the box's width must equal height% of the box's height.
      assert.ok(Math.abs(width * parsed.width - height * parsed.height) < 1e-4,
        `d${faceCount}: content box is square in em`);
      const corners: readonly (readonly [number, number])[] = [
        [left, top], [left + width, top], [left + width, top + height], [left, top + height],
      ];
      for (const [px, py] of corners) {
        assert.ok(pointInPolygon(parsed.clip, px, py),
          `d${faceCount}: content corner (${px}, ${py}) escapes the facet`);
      }
    }
  }
});

/** Convex point-in-polygon in clip space, with slack for the emitted rounding. */
function pointInPolygon(polygon: readonly (readonly [number, number])[], px: number, py: number): boolean {
  let sign = 0;
  for (let i = 0; i < polygon.length; i++) {
    const [ax, ay] = polygon[i];
    const [bx, by] = polygon[(i + 1) % polygon.length];
    const side = (bx - ax) * (py - ay) - (by - ay) * (px - ax);
    if (Math.abs(side) < 1e-6) continue;
    const current = side > 0 ? 1 : -1;
    if (sign === 0) sign = current;
    else if (sign !== current) return false;
  }
  return true;
}

test('inscribedSquareHalfSide is exact on shapes with a known answer', () => {
  const unitSquare: PlanePoint[] = [
    { a: -1, b: -1 }, { a: 1, b: -1 }, { a: 1, b: 1 }, { a: -1, b: 1 },
  ];
  assert.ok(Math.abs(inscribedSquareHalfSide(unitSquare, 0, 0) - 1) < 1e-12);
  // Off centre by 0.25: the nearest edge is 0.75 away and that is the bound.
  assert.ok(Math.abs(inscribedSquareHalfSide(unitSquare, 0.25, 0) - 0.75) < 1e-12);
  // Outside the polygon there is no square at all.
  assert.equal(inscribedSquareHalfSide(unitSquare, 2, 0), 0);
  assert.equal(inscribedSquareHalfSide(unitSquare, 0, -1.5), 0);
  // A 45-degree diamond of "radius" 1: the inscribed axis-aligned square has
  // half-side 0.5, since each edge normal is (±1, ±1)/sqrt(2) at distance
  // 1/sqrt(2) and s * (|na| + |nb|) = s * sqrt(2) must not exceed it.
  const diamond: PlanePoint[] = [{ a: 1, b: 0 }, { a: 0, b: 1 }, { a: -1, b: 0 }, { a: 0, b: -1 }];
  assert.ok(Math.abs(inscribedSquareHalfSide(diamond, 0, 0) - 0.5) < 1e-12);
});

test('inscribedSquareHalfSide does not care which way the polygon winds', () => {
  // The projection into a facet's plane produces either winding depending on
  // which way the facet faces, and the routine takes the winding from the
  // polygon rather than assuming one.
  const clockwise: PlanePoint[] = [{ a: -2, b: -1 }, { a: -2, b: 1 }, { a: 2, b: 1 }, { a: 2, b: -1 }];
  const counter = [...clockwise].reverse();
  assert.equal(inscribedSquareHalfSide(clockwise, 0, 0), inscribedSquareHalfSide(counter, 0, 0));
  assert.ok(Math.abs(inscribedSquareHalfSide(clockwise, 0, 0) - 1) < 1e-12);
});

test('a degenerate facet is refused rather than emitted as a zero-sized box', () => {
  const edgeOn: SolidFace = {
    normal: vec3(0, 0, 1),
    centroid: vec3(0, 0, 0),
    polygon: [vec3(-1, 0, 0), vec3(1, 0, 0), vec3(0.5, 0, 0)],
  };
  assert.throws(() => facetPlacement(edgeOn, 1, null), /degenerate facet/);
});

test('solidFacets covers the whole surface and marks caps with faceIndex -1', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const facets = solidFacets(geometry);
    assert.equal(facets.length, geometry.faces.length + geometry.capFaces.length);
    facets.forEach((facet, index) => {
      assert.equal(facet.key, index, 'key is the surface index');
      assert.equal(facet.faceIndex, index < geometry.faces.length ? index : -1);
      assert.deepEqual(facet.corners, [], 'no corner marks without a cornerOwner');
    });
  }
});

/**
 * Every solid must be exactly 1em across its bounding sphere whatever its own
 * natural scale, or a d20 (circumradius 1.902) would render nearly twice the
 * size of a d8 (1.000) at the same `--die-size`, and a tumbling die would leave
 * the box a game laid out for it.
 */
test('solidFacets normalizes every solid to a 1em bounding sphere', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    let farthest = 0;
    for (const facet of solidFacets(geometry)) {
      const parsed = parseStyle(facet.style);
      for (const [px, py] of parsed.clip) {
        farthest = Math.max(farthest, magnitude(clipPointInParent(parsed, px, py)));
      }
    }
    assert.ok(Math.abs(farthest - 0.5) < 1e-4,
      `d${faceCount}: circumradius rendered at ${farthest}em, expected 0.5em`);
  }
});

test('cornerOwner turns on corner marks, on readable faces only', () => {
  const geometry = dieGeometry(4);
  const seen: Vec3[] = [];
  const facets = solidFacets(geometry, {
    cornerOwner: (vertex) => {
      seen.push(vertex);
      return 0;
    },
  });
  assert.ok(seen.length > 0, 'cornerOwner is consulted');
  facets.forEach((facet) => {
    if (facet.faceIndex < 0) {
      assert.deepEqual(facet.corners, [], 'caps carry nothing');
      return;
    }
    assert.equal(facet.corners.length, geometry.faces[facet.faceIndex].polygon.length);
    for (const corner of facet.corners) {
      assert.equal(corner.faceIndex, 0);
      assert.ok(corner.size > 0, 'a corner mark has a readable square');
      // Inside its own facet's box, so it cannot spill onto a neighbour.
      assert.ok(corner.left >= -1e-9 && corner.top >= -1e-9);
      assert.ok(corner.left + corner.width <= 100 + 1e-9);
      assert.ok(corner.top + corner.height <= 100 + 1e-9);
    }
  });
});

/**
 * A corner mark carries the value read when THAT corner is the top of the die,
 * so which vertex a mark belongs to is the whole meaning of the mark. Nothing
 * downstream re-derives it — `faceIndex` is taken on trust from the index the
 * mark was produced at — so the ownership has to be pinned here.
 */
test('each corner mark sits nearest the vertex whose value it carries', () => {
  for (const faceCount of [4, 5, 7] as const) {
    const geometry = dieGeometry(faceCount);
    const unitsToEm = 0.5 / geometry.circumradius;
    for (const [faceIndex, face] of geometry.faces.entries()) {
      const parsed = parseStyle(facetPlacement(face, unitsToEm, null).style);
      const { corners } = facetPlacement(face, unitsToEm, face.polygon.map((_, i) => i));
      // Distances in em, not in the box's percentages, which are anisotropic on
      // a barrel's 2.7:1 side face.
      const inEm = (px: number, py: number) =>
        [(px - 0.5) * parsed.width, (py - 0.5) * parsed.height] as const;
      corners.forEach((corner, index) => {
        assert.equal(corner.faceIndex, index, 'the mark keeps its vertex index');
        const [cx, cy] = inEm(
          (corner.left + corner.width / 2) / 100, (corner.top + corner.height / 2) / 100);
        const distances = parsed.clip.map(([px, py]) => {
          const [vx, vy] = inEm(px, py);
          return Math.hypot(cx - vx, cy - vy);
        });
        const nearest = distances.indexOf(Math.min(...distances));
        assert.equal(nearest, index,
          `d${faceCount} face ${faceIndex}: mark ${index} is nearest vertex ${nearest}`);
        assert.ok(pointInPolygon(
          parsed.clip,
          (corner.left + corner.width / 2) / 100,
          (corner.top + corner.height / 2) / 100,
        ), `d${faceCount} face ${faceIndex}: mark ${index} escaped the facet`);
      });
    }
  }
});

test('solidFacets accepts any structurally compatible surface, not just a die', () => {
  // A 3D token is meant to build one of these by hand. Nothing die-shaped is
  // required: three fields, and the caller keeps its own vocabulary.
  const half = 0.5;
  const quad = (normal: Vec3, polygon: readonly Vec3[]): SolidFace => ({
    normal,
    centroid: scaleVec(polygon.reduce((a, b) => add(a, b), vec3(0, 0, 0)), 1 / polygon.length),
    polygon,
  });
  const slab = {
    faces: [
      quad(vec3(0, 0, 1), [
        vec3(-half, -half, half), vec3(half, -half, half), vec3(half, half, half), vec3(-half, half, half),
      ]),
      quad(vec3(0, 0, -1), [
        vec3(half, -half, -half), vec3(-half, -half, -half), vec3(-half, half, -half), vec3(half, half, -half),
      ]),
    ],
    capFaces: [],
    circumradius: Math.hypot(half, half, half),
  };
  const facets = solidFacets(slab);
  assert.equal(facets.length, 2);
  for (const facet of facets) {
    assert.ok(facet.style.includes('clip-path:polygon('));
    assert.ok(facet.style.includes('--content-size:'));
  }
});
