import assert from 'node:assert/strict';
import test from 'node:test';
import {
  CAMERA_DEPTH_WIDTHS,
  CAMERA_LEAN_DEGREES,
  PRISM_SIDES,
  SHAPE_HEIGHT_RATIO,
  TOKEN_BASE_RED,
  TOKEN_COLOR_FILTERS,
  applyColorFilter,
  facetShade,
  fitScale,
  isTokenSolidShape,
  posedNormals,
  posedUp,
  restingPose,
  silhouetteExtent,
  tokenBaseColor,
  tokenSolid,
  tokenSurface,
  type Rgb,
  type TokenSolidShape,
} from './token-solid.ts';

const SHAPES: readonly TokenSolidShape[] = ['cube', 'token', 'chip', 'disc'];

const dot = (a: readonly number[], b: readonly number[]) =>
  a[0] * b[0] + a[1] * b[1] + a[2] * b[2];

/** The style's `background:rgb(r,g,b)`, as numbers. */
function fillOf(style: string): Rgb {
  const match = /background:rgb\((\d+),(\d+),(\d+)\)/.exec(style);
  assert.ok(match, `no background in ${style}`);
  return [Number(match[1]), Number(match[2]), Number(match[3])];
}

// ---------------------------------------------------------------------------
// The facet budget.
// ---------------------------------------------------------------------------

test('a prism token is 14 facets and a cube is 6', () => {
  assert.equal(PRISM_SIDES, 12);
  assert.equal(tokenSolid('cube', 'red').facets.length, 6);
  for (const shape of ['token', 'chip', 'disc'] as const) {
    assert.equal(tokenSolid(shape, 'red').facets.length, PRISM_SIDES + 2);
  }
});

test('55 tokens fit under the measured facet budget', () => {
  // The wall the design measured: ~800 clip-path'ed elements is free, ~1400 is
  // 42.8fps. `pass` is the worst real case at 55 tokens, all of them chips.
  const worst = Math.max(...SHAPES.map((shape) => tokenSolid(shape, 'red').facets.length));
  assert.equal(worst * 55, 770);
  assert.ok(worst * 55 < 800, `55 tokens is ${worst * 55} facet elements`);
});

// ---------------------------------------------------------------------------
// The pose.
// ---------------------------------------------------------------------------

test('every resting pose is a rotation, never a reflection', () => {
  for (const shape of SHAPES) {
    const m = restingPose(shape);
    // Orthonormal rows.
    for (let i = 0; i < 3; i++) {
      for (let j = 0; j < 3; j++) {
        const expected = i === j ? 1 : 0;
        const product = dot(m.slice(i * 3, i * 3 + 3), m.slice(j * 3, j * 3 + 3));
        assert.ok(Math.abs(product - expected) < 1e-12, `${shape} rows ${i},${j}: ${product}`);
      }
    }
    // A reflection here would mirror every facet's own content and leave the
    // silhouette alone -- the exact defect `screen-frame.ts` documents.
    const determinant =
      m[0] * (m[4] * m[8] - m[5] * m[7])
      - m[1] * (m[3] * m[8] - m[5] * m[6])
      + m[2] * (m[3] * m[7] - m[4] * m[6]);
    assert.ok(Math.abs(determinant - 1) < 1e-12, `${shape} determinant ${determinant}`);
  }
});

test('every shape stands the same way up under one camera', () => {
  // The two families are built about different axes -- a die's up is body +Y, a
  // prism's is body +Z -- so this is the assertion that `align` reconciles them
  // rather than merely existing. A sign error puts a cube's top face where a
  // disc's bottom is, which no other test here would notice.
  //
  // Two halves, and the second is the one with teeth: every shape's own up-axis
  // lands in the same place on screen, AND each shape actually HAS a facet
  // facing that way (the cube's +Y face, the prism's cap). Without the second,
  // an `align` that pointed a prism sideways would still satisfy the first.
  const lean = (CAMERA_LEAN_DEGREES * Math.PI) / 180;
  const expected = [0, -Math.cos(lean), Math.sin(lean)];
  for (const shape of SHAPES) {
    const up = posedUp(shape);
    for (let i = 0; i < 3; i++) {
      assert.ok(
        Math.abs(up[i] - expected[i]) < 1e-9,
        `${shape}'s up-axis points ${up} rather than ${expected}`,
      );
    }
    const facing = posedNormals(shape).some((normal) =>
      Math.abs(normal[0] - up[0]) < 1e-9
      && Math.abs(normal[1] - up[1]) < 1e-9
      && Math.abs(normal[2] - up[2]) < 1e-9);
    assert.ok(facing, `${shape} has no facet facing its own up-axis`);
  }
});

test('the camera leans towards the viewer, so the up face is visible', () => {
  // Sign check, stated as the thing a person would see: the face the pose puts
  // on top of the piece has to face the CAMERA (+z in the CSS frame), not away
  // from it. Leaning the wrong way renders every piece from underneath, and
  // `backface-visibility` then culls the top face out of every one of them.
  for (const shape of SHAPES) {
    assert.ok(posedUp(shape)[2] > 0, `${shape}'s top face points away from the camera`);
    assert.ok(posedUp(shape)[1] < 0, `${shape}'s top face is not at the top`);
  }
});

test('the pose depends on nothing but the shape', () => {
  for (const shape of SHAPES) {
    assert.deepEqual(restingPose(shape), restingPose(shape));
    assert.deepEqual(tokenSolid(shape, 'red').pose, tokenSolid(shape, 'blue').pose);
  }
});

// ---------------------------------------------------------------------------
// Sizing.
// ---------------------------------------------------------------------------

test('the fit scale is the exact fixed point of its own silhouette', () => {
  // `fitScale` iterates, because the silhouette depends on the camera distance
  // and the camera distance depends on the size. This says the answer is right
  // without saying how many iterations it took to get there.
  for (const shape of SHAPES) {
    const fit = fitScale(shape);
    const filled = fit * silhouetteExtent(shape, fit);
    assert.ok(Math.abs(filled - 1) < 1e-12, `${shape} fills ${filled} of its box`);
  }
});

test('the perspective actually magnifies, so the fit is not the orthographic one', () => {
  // If this ever came out equal, the camera would have quietly become
  // orthographic and the fixed point above would be vacuously satisfied.
  for (const shape of SHAPES) {
    const orthographic = 1 / silhouetteExtent(shape, 0);
    assert.ok(
      orthographic > fitScale(shape),
      `${shape}: perspective must widen the silhouette, so the fit must shrink`,
    );
    assert.ok(orthographic / fitScale(shape) < 1.2, `${shape}: ${CAMERA_DEPTH_WIDTHS} widths is not a fisheye`);
  }
});

test('a cube is sized by its silhouette, so its own face is smaller than the box', () => {
  // What "size by drawn extent" MEANS, said about the shape that makes it
  // clearest. The silhouette is one box across; a cube's face spans only
  // 1/sqrt(3) = 58% of its circumsphere, so the face is much less than the box
  // -- which is exactly what the isometric `token_cube.svg` draws, and exactly
  // what makes the solid a drop-in for it. Sizing so the FACE filled the box
  // instead would draw the piece 73% larger and hang it out of its own slot.
  const fit = fitScale('cube');
  const surface = tokenSurface('cube');
  const pose = restingPose('cube');
  // The widest pair of vertices across the posed cube, orthographically.
  const unitsToEm = 0.5 / surface.nominalRadius;
  let widest = 0;
  for (const face of surface.faces) {
    for (const vertex of face.polygon) {
      const screen = [vertex[0], -vertex[1], vertex[2]].map((v) => v * unitsToEm);
      const x = dot(pose.slice(0, 3), screen);
      const y = dot(pose.slice(3, 6), screen);
      widest = Math.max(widest, Math.abs(x), Math.abs(y));
    }
  }
  // Drawn at `fit`, the silhouette is one box across (within the perspective
  // term the exact test above pins), and the cube's own face is much less.
  assert.ok(Math.abs(2 * widest * fit - 1) < 0.02, `cube silhouette ${2 * widest * fit}`);
  const faceSpan = (1 / Math.sqrt(3)) * fit;
  assert.ok(faceSpan < 0.6, `a cube's face spans ${faceSpan} of the box, as it must`);
});

test('a taller prism is drawn smaller across, because its silhouette is taller', () => {
  // The token is 0.55 as thick as it is wide and the disc a tenth, so leaning
  // them by the same camera makes the token's silhouette the taller one -- and a
  // fit that ignored the pose would give all three prisms the same number.
  assert.ok(SHAPE_HEIGHT_RATIO.token > SHAPE_HEIGHT_RATIO.chip);
  assert.ok(SHAPE_HEIGHT_RATIO.chip > SHAPE_HEIGHT_RATIO.disc);
  assert.ok(fitScale('token') < fitScale('chip'));
  assert.ok(fitScale('chip') < fitScale('disc'));
});

// ---------------------------------------------------------------------------
// Lighting.
// ---------------------------------------------------------------------------

test('the light comes from above and from the left', () => {
  // Stated as three comparisons rather than as the light vector, so it is a fact
  // about the PICTURE: up beats down, left beats right, and the range matches
  // the bevel the authored disc art draws (1.03 down to 0.58 of its face).
  const up = facetShade([0, -1, 0]);
  const down = facetShade([0, 1, 0]);
  const left = facetShade([-1, 0, 0]);
  const right = facetShade([1, 0, 0]);
  assert.ok(up > down, `up ${up} must be brighter than down ${down}`);
  assert.ok(left > right, `left ${left} must be brighter than right ${right}`);
  assert.ok(up <= 1.08 && down >= 0.5, `the shade range is ${down}..${up}`);
});

test('a facing cap is drawn at the flat art\'s own brightness', () => {
  // The cap of a resting prism is what replaces the flat SVG's face, so it has
  // to come out at essentially 1.0 -- otherwise every 3D token on a board is a
  // different shade of red from every meeple beside it.
  for (const shape of ['token', 'chip', 'disc'] as const) {
    const shade = facetShade(posedUp(shape));
    assert.ok(Math.abs(shade - 1) < 0.05, `${shape}'s cap is drawn at ${shade}`);
  }
});

test('a shape is drawn with more than one brightness', () => {
  // A flat-shaded solid is a polygon. This is the cheapest possible statement
  // that the shading is doing anything at all.
  for (const shape of SHAPES) {
    const fills = tokenSolid(shape, 'red').facets.map((facet) => fillOf(facet.style)[0]);
    assert.ok(new Set(fills).size >= 3, `${shape} draws ${new Set(fills).size} distinct fills`);
  }
});

// ---------------------------------------------------------------------------
// Colour.
// ---------------------------------------------------------------------------

test('an unnamed colour is the red the art is drawn in', () => {
  // What a stack passes for a component whose colour is hidden, and what the
  // flat art shows in the same case: no class matches, so no filter applies.
  assert.deepEqual(tokenBaseColor(''), TOKEN_BASE_RED);
  assert.deepEqual(tokenBaseColor('red'), TOKEN_BASE_RED);
  assert.deepEqual(tokenBaseColor('not-a-colour'), TOKEN_BASE_RED);
});

test('every legal colour is its own colour', () => {
  const seen = new Map<string, string>();
  for (const name of Object.keys(TOKEN_COLOR_FILTERS)) {
    const key = tokenBaseColor(name).join(',');
    assert.ok(!seen.has(key), `${name} and ${seen.get(key)} are the same colour`);
    seen.set(key, name);
  }
  assert.equal(seen.size, Object.keys(TOKEN_COLOR_FILTERS).length);
});

test('the colour name is case-insensitive, as the class is', () => {
  assert.deepEqual(tokenBaseColor('Blue'), tokenBaseColor('blue'));
  assert.deepEqual(
    fillOf(tokenSolid('disc', 'Blue').facets[0].style),
    fillOf(tokenSolid('disc', 'blue').facets[0].style),
  );
});

test('filter functions apply left to right and clamp as they go', () => {
  // Order matters, and WHY it matters is the clamp: brightness and the colour
  // matrices are both linear, so they would commute exactly if nothing ever
  // clipped. Brightening a saturated red past 255 and then desaturating is not
  // the same colour as desaturating first -- 175 against 255 -- and the CSS
  // shorthand is defined left to right.
  const forwards = applyColorFilter('brightness(3) saturate(0)', [200, 20, 20]);
  const backwards = applyColorFilter('saturate(0) brightness(3)', [200, 20, 20]);
  assert.notDeepEqual(forwards.map(Math.round), backwards.map(Math.round));
  // Clamped per function, which is what a filter primitive's result buffer does.
  assert.deepEqual(applyColorFilter('brightness(4)', [200, 200, 200]), [255, 255, 255]);
  assert.deepEqual(applyColorFilter('brightness(0)', [200, 200, 200]), [0, 0, 0]);
  // Identity cases, so a typo in the matrices shows up here rather than on screen.
  assert.deepEqual(applyColorFilter('brightness(1)', [12, 34, 56]), [12, 34, 56]);
  assert.deepEqual(applyColorFilter('hue-rotate(0deg)', [12, 34, 56]).map(Math.round), [12, 34, 56]);
  assert.deepEqual(applyColorFilter('saturate(1)', [12, 34, 56]).map(Math.round), [12, 34, 56]);
  // A grey has no hue to rotate.
  assert.deepEqual(applyColorFilter('hue-rotate(137deg)', [80, 80, 80]).map(Math.round), [80, 80, 80]);
});

test('an unimplemented filter function is refused rather than skipped', () => {
  // Skipping it silently would recolour the flat art and not the solid, which is
  // the exact drift the shared table exists to prevent.
  assert.throws(() => applyColorFilter('sepia(0.5)', TOKEN_BASE_RED), /unsupported function sepia/);
  assert.throws(() => applyColorFilter('nonsense', TOKEN_BASE_RED), /could not parse/);
});

test('a shaded facet keeps its colour\'s hue', () => {
  // Shading multiplies, so the hue survives and only the brightness moves --
  // which is why a 3D blue chip and a flat blue meeple read as the same blue.
  // The trap is the ceiling: orange is (255, 91, 0), so a highlight above 1.0
  // clips the red and lifts only the green, and the lit facet goes yellow.
  for (const name of Object.keys(TOKEN_COLOR_FILTERS).concat(['red'])) {
    const base = tokenBaseColor(name);
    const channels = [0, 1, 2].filter((i) => base[i] >= 16);
    for (const facet of tokenSolid('disc', name).facets) {
      const fill = fillOf(facet.style);
      for (const i of channels) {
        assert.ok(fill[i] <= 255, `${name} channel ${i} clipped at ${fill[i]}`);
        // Every channel scaled by the SAME factor, to within one part in 255.
        const drift = Math.abs(fill[i] / base[i] - fill[channels[0]] / base[channels[0]]);
        assert.ok(drift < 0.02, `${name} channel ${i} drifted by ${drift}`);
      }
    }
  }
});

// ---------------------------------------------------------------------------
// The whole thing.
// ---------------------------------------------------------------------------

test('a solid is a pure function of its type and colour', () => {
  // The property the whole design rests on: nothing here remembers a previous
  // caller, so a recycled host cannot inherit one.
  for (const shape of SHAPES) {
    assert.deepEqual(tokenSolid(shape, 'green'), tokenSolid(shape, 'green'));
    assert.notDeepEqual(
      tokenSolid(shape, 'green').facets[0].style,
      tokenSolid(shape, 'orange').facets[0].style,
    );
  }
  // ...and the geometry does not depend on the colour at all.
  const strip = (style: string) => style.replace(/;background:rgb\([\d,]+\)/, '');
  for (const shape of SHAPES) {
    assert.deepEqual(
      tokenSolid(shape, 'green').facets.map((f) => strip(f.style)),
      tokenSolid(shape, 'black').facets.map((f) => strip(f.style)),
    );
  }
});

test('every facet carries a clip path, a transform and a fill', () => {
  for (const shape of SHAPES) {
    const solid = tokenSolid(shape, 'teal');
    solid.facets.forEach((facet, index) => {
      assert.equal(facet.key, index, `${shape} facet keys must be the surface order`);
      assert.match(facet.style, /clip-path:polygon\(/);
      assert.match(facet.style, /transform:translate3d\(.*matrix3d\(/);
      assert.match(facet.style, /background:rgb\(/);
    });
  }
});

test('only the four convex types are solids', () => {
  for (const shape of SHAPES) assert.ok(isTokenSolidShape(shape));
  for (const other of ['meeple', 'pawn', '', 'Cube', 'sphere']) {
    assert.ok(!isTokenSolidShape(other), `${other} must not render as a solid`);
  }
});
