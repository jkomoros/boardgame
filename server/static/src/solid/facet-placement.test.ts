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
  CORNER_INSET_MAX,
  CORNER_INSET_MIN,
  CORNER_MAX_SIZE,
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
    const unitsToEm = 0.5 / geometry.nominalRadius;
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
    const unitsToEm = 0.5 / geometry.nominalRadius;
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
    const unitsToEm = 0.5 / geometry.nominalRadius;
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
    const unitsToEm = 0.5 / geometry.nominalRadius;
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
 * THE ONE SCALE IN THE PIPELINE, and the direction its two radii differ in.
 *
 * Every solid renders at `0.5em` per NOMINAL radius whatever its own natural
 * scale, or a d20 (nominal 1.902) would draw nearly twice the size of a d8
 * (1.000) at the same `1em`.
 *
 * Nominal is not bounding, and the gap is what `boardgame-die.ts`'s `solidExtent`
 * divides the author's `--die-size` by to get this `1em`. For a closed-form
 * solid the two radii are the same number and the bounding sphere is `1em`
 * across. A BARREL is normalized by its short axis instead, so its length
 * deliberately overflows `1em` — and the die then scales the whole solid down by
 * that same overflow so it still fits the footprint the author asked for. The
 * two are inverses, which is why this test pins BOTH ends of the barrel case:
 * its width is exactly `1em` and its length is out of it by about the aspect
 * ratio. A change in either direction rescales every barrel on screen.
 */
test('solidFacets renders every solid at 0.5em per nominal radius', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    // The barrels in this list. A barrel's long axis is z, and `toScreen` is
    // `diag(1, -1, 1)`, so z stays the long axis in the rendered frame too.
    const barrel = geometry.capFaces.length > 0;
    let farthest = 0;
    let widest = 0;
    for (const facet of solidFacets(geometry)) {
      const parsed = parseStyle(facet.style);
      for (const [px, py] of parsed.clip) {
        const point = clipPointInParent(parsed, px, py);
        farthest = Math.max(farthest, magnitude(point));
        widest = Math.max(widest, Math.hypot(point[0], point[1]));
      }
    }

    // The scale itself, stated the one way that holds for every shape: the
    // rendered bounding sphere is the geometry's, times `0.5 / nominalRadius`.
    const expected = 0.5 * (geometry.boundingRadius / geometry.nominalRadius);
    assert.ok(Math.abs(farthest - expected) < 1e-4,
      `d${faceCount}: rendered at ${farthest}em per bounding radius, expected ${expected}em`);

    if (!barrel) {
      // Closed form: nominal IS the circumradius, so the whole solid fits 1em.
      assert.ok(Math.abs(farthest - 0.5) < 1e-4,
        `d${faceCount}: closed-form solid rendered ${farthest}em from centre, expected 0.5em`);
      continue;
    }
    // A barrel's WIDTH is the box: this is the assertion that says its marks
    // are sized by the short axis, and it is what fails if it is ever
    // normalized by its circumsphere again.
    assert.ok(Math.abs(widest - 0.5) < 1e-4,
      `d${faceCount}: barrel is ${2 * widest}em wide, expected exactly 1em`);
    // And its LENGTH is outside the box, by the aspect ratio: 2.13x from the d5
    // up (`BARREL_CAP_SAFETY`), so the bounding sphere clears 1em with room.
    assert.ok(farthest > 1.0,
      `d${faceCount}: barrel bounding sphere is only ${2 * farthest}em across; `
      + 'it has gone back to being sized by its circumsphere and its marks have halved');
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
    const unitsToEm = 0.5 / geometry.nominalRadius;
    for (const [faceIndex, face] of geometry.faces.entries()) {
      const parsed = parseStyle(facetPlacement(face, unitsToEm, null).style);
      const { corners, contentSize } = facetPlacement(
        face, unitsToEm, face.polygon.map((_, i) => i));
      // Distances in em, not in the box's percentages, which are anisotropic on
      // a barrel's 2.7:1 side face.
      const inEm = (px: number, py: number) =>
        [(px - 0.5) * parsed.width, (py - 0.5) * parsed.height] as const;
      // The content square is centred on the facet's centroid, so its own box
      // gives the centroid without re-deriving it.
      const content = {
        left: parsed.vars['--content-left'],
        top: parsed.vars['--content-top'],
        width: parsed.vars['--content-width'],
        height: parsed.vars['--content-height'],
      };
      const centroid = inEm(
        (content.left + content.width / 2) / 100, (content.top + content.height / 2) / 100);
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

        // HOW BIG, AND HOW FAR IN. Everything above this line is structure --
        // a mark exists, it is the right one, its centre is on the facet --
        // and a mutation pass found that structure was all this file ever
        // asserted: `Math.min(best.size, cap)` -> `Math.max` survived, as did
        // moving `CORNER_INSET_MAX` from 0.65 to 1.65 and `CORNER_MAX_SIZE`
        // from 0.6 to 1.6. Those constants carry a measurement log in their
        // doc comment (a 6.5px font on a 100px d7, marks that collide at the
        // next step up) and none of it was pinned by anything.
        assert.ok(
          corner.size <= contentSize * CORNER_MAX_SIZE + 1e-9,
          `d${faceCount} face ${faceIndex}: mark ${index} is ${corner.size}em, over the ${
            contentSize * CORNER_MAX_SIZE}em cap`,
        );
        // How far in from its vertex, as a fraction of the vertex-to-centroid
        // line -- which is exactly the quantity the inset scan walks.
        const [vx, vy] = inEm(...parsed.clip[index]);
        const inset = Math.hypot(cx - vx, cy - vy)
          / Math.hypot(centroid[0] - vx, centroid[1] - vy);
        assert.ok(
          inset >= CORNER_INSET_MIN - 1e-6 && inset <= CORNER_INSET_MAX + 1e-6,
          `d${faceCount} face ${faceIndex}: mark ${index} sits ${inset} of the way to the centroid`,
        );
      });

      // THE SCAN RANGE MUST NOT CLIP THE OPTIMUM. The inset the scan settles on
      // has to land strictly INSIDE [CORNER_INSET_MIN, CORNER_INSET_MAX]; an
      // endpoint means the best inset was outside the range and the mark is
      // wearing whatever the range would allow instead.
      //
      // This is the assertion that makes those two constants mean something,
      // and it took a measurement to find the right one. The obvious reading --
      // that widening the range lets marks slide down onto the centre numeral
      // -- is FALSE, measured: with CORNER_INSET_MAX at 0.95 instead of 0.65
      // the chosen inset moves 0.6108 -> 0.6292 (the step grid shifts) and mark
      // size is byte-identical at 0.1491/0.2539/0.1874em on d4/d5/d7. Raising
      // MIN to 0.55 is likewise inert. The optimum is interior at ~0.61, so
      // neither end BINDS, and a mutation widening either one is equivalent --
      // surviving is the correct outcome, not an escaped bug.
      //
      // Narrowing is the direction that hurts, and this catches it: MAX at 0.45
      // pins every mark to exactly 0.4500 and shrinks all three by 25%
      // (0.1491 -> 0.1118, 0.2539 -> 0.1904, 0.1874 -> 0.1406), which is
      // straight into the unreadable range the constants' doc comment exists to
      // keep marks out of.
      const chosen = corners.map((corner, index) => {
        const [mx, my] = inEm(
          (corner.left + corner.width / 2) / 100, (corner.top + corner.height / 2) / 100);
        const [vx, vy] = inEm(...parsed.clip[index]);
        return Math.hypot(mx - vx, my - vy) / Math.hypot(centroid[0] - vx, centroid[1] - vy);
      });
      for (const [index, inset] of chosen.entries()) {
        assert.ok(
          inset > CORNER_INSET_MIN + 1e-6 && inset < CORNER_INSET_MAX - 1e-6,
          `d${faceCount} face ${faceIndex}: mark ${index} settled at inset ${inset.toFixed(4)}, `
          + `on the edge of the [${CORNER_INSET_MIN}, ${CORNER_INSET_MAX}] scan range -- the range `
          + 'is clipping the best inset, so this mark is smaller than the facet can carry',
        );
      }

      // THE POSITIVE CONTROL for the size bound. On all three of these shapes
      // the inset scan finds a square bigger than the cap at every vertex, so
      // the cap is what decides the mark's size -- every mark comes out at
      // exactly `contentSize * CORNER_MAX_SIZE`. That is what makes the bound
      // above a bound rather than a ceiling nothing reaches, and it is what
      // makes `Math.min(best.size, cap)` -> `Math.max` observable at all.
      assert.ok(
        corners.every(
          (corner) => Math.abs(corner.size - contentSize * CORNER_MAX_SIZE) < 1e-9),
        `d${faceCount} face ${faceIndex}: the size cap never binds; sizes ${
          corners.map((corner) => corner.size / contentSize)}`,
      );
    }
  }
});

/**
 * The two bounds of the inset scan, stated in absolute terms rather than
 * against themselves. `t` is the fraction of the way from a vertex to the
 * centroid, so `t >= 1` puts a corner mark AT or PAST the centroid -- on the
 * far side of the facet from the vertex whose value it carries, which makes
 * `each corner mark sits nearest the vertex whose value it carries` above
 * unsatisfiable. A mutation pass moved `CORNER_INSET_MAX` to 1.65 and nothing
 * noticed; the assertion above would move with it, this one does not.
 */
test('the corner inset scan stays on the vertex side of the centroid', () => {
  assert.ok(CORNER_INSET_MIN > 0, `${CORNER_INSET_MIN}: a mark sitting on its vertex has no room`);
  assert.ok(CORNER_INSET_MIN < CORNER_INSET_MAX, `${CORNER_INSET_MIN} >= ${CORNER_INSET_MAX}`);
  assert.ok(CORNER_INSET_MAX < 1, `${CORNER_INSET_MAX}: the scan walks past the centroid`);
  // A corner mark is a fraction of the CENTRE content, never a multiple of it:
  // at 1.0 the mark is the whole content square and there is nothing left to
  // put in the middle of the facet.
  assert.ok(CORNER_MAX_SIZE > 0 && CORNER_MAX_SIZE < 1, `${CORNER_MAX_SIZE}`);
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
    nominalRadius: Math.hypot(half, half, half),
  };
  const facets = solidFacets(slab);
  assert.equal(facets.length, 2);
  for (const facet of facets) {
    assert.ok(facet.style.includes('clip-path:polygon('));
    assert.ok(facet.style.includes('--content-size:'));
  }
});
