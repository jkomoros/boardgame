import { test, expect } from '@playwright/test';
import { createOfflineGame } from '../helpers';

// Task 8: <boardgame-die> stops being a vertical reel of flat squares and
// becomes an actual solid -- one absolutely positioned element per polygon of
// the surface `die-geometry.ts` generates, placed by a transform derived from
// that polygon's own normal and centroid and cut to its outline with
// clip-path.
//
// These are FIXTURE assertions (no goldens): the die's *animation* is
// unchanged by this task, so nothing here samples motion curves. What it pins
// is the DOM/CSS contract that the later physics tasks build on:
//
//   1. one element per surface polygon, for shapes whose facets are
//      rectangles (d6), triangles (d20) and non-square rectangles + cap
//      triangles (the d7 barrel) -- the same code path for all three, each
//      drawn at the geometry's OWN in-plane extents (test 2: counting
//      clip-path vertices is not enough -- forcing every facet's box square
//      leaves every vertex count untouched while the solid falls apart);
//   1b. the resting pose shows more than one facet, so the die reads as a
//      solid rather than as the flat polygon it replaces;
//   2. `#inner` -- which `motionTrackTarget('visual')` returns and which a
//      later task animates the tumble on -- carries a LIVE preserve-3d
//      context, verified by a consequence (depth sorting) and not merely by
//      a computed-style string that Chromium reports from the specified
//      value even when a grouping property has flattened it;
//   3. every facet is cut to its polygon;
//   4. `--die-size` is a real custom property a caller can set;
//   5. the reel survives as the degenerate fallback for a die with no valid
//      geometry, rather than throwing during a render pass.
//
// Test 5 mounts the die in the REAL app (pig) rather than in an isolated
// fixture, because an isolated fixture cannot see an ancestor that flattens
// the 3D context.

const DIE_SIZE_DEFAULT_PX = 50;

// Mounts one <boardgame-die> in the served app page. Everything else in this
// spec reads through page.evaluate on the same element id.
//
// `SelectedFace` is an INDEX into `Faces`, not a value -- confusing the two is
// the silent bug this component invites, so the fixtures deliberately use face
// VALUES that never coincide with their own indices (10, 20, 30, ...).
async function mountDie(
  page: import('@playwright/test').Page,
  options: {
    faceCount: number;
    selectedFace?: number;
    dieSize?: string;
    /** Face VALUES. Defaults to 10, 20, 30, ... (never equal to their index). */
    faces?: number[];
    /** Face value -> name. What an enum would supply; also the a11y label. */
    faceNames?: Record<string, string>;
    /** Face name -> glyph. The author-supplied symbol set. */
    symbols?: Record<string, string>;
  },
): Promise<void> {
  await page.goto('/');
  await page.evaluate(async (opts) => {
    document.querySelectorAll('boardgame-die').forEach((el) => el.remove());
    await import('/src/components/boardgame-die.ts');
    const die = document.createElement('boardgame-die') as any;
    die.id = 'fixture-die';
    // z-index because the app shell paints over a bare fixed-position child
    // of <body>, and one test hit-tests the die's own centre.
    die.style.cssText = 'position:fixed;top:120px;left:120px;z-index:9999;';
    if (opts.dieSize) die.style.setProperty('--die-size', opts.dieSize);
    const faces = opts.faces ?? Array.from({ length: opts.faceCount }, (_, i) => (i + 1) * 10);
    if (opts.faceNames) die.faceNames = opts.faceNames;
    if (opts.symbols) die.symbols = opts.symbols;
    die.item = {
      Values: { Faces: faces },
      DynamicValues: { SelectedFace: opts.selectedFace ?? 0, Value: faces[opts.selectedFace ?? 0] },
    };
    document.body.appendChild(die);
    await die.updateComplete;
    // _itemChanged runs in updated(), which schedules a second render pass.
    await die.updateComplete;
  }, {
    faceCount: options.faceCount,
    selectedFace: options.selectedFace,
    dieSize: options.dieSize,
    faces: options.faces,
    faceNames: options.faceNames,
    symbols: options.symbols,
  } as any);
}

// Reads the facet inventory of the mounted fixture die, plus the surface the
// geometry module says it should have.
async function facetInventory(page: import('@playwright/test').Page, faceCount: number) {
  return await page.evaluate(async (count) => {
    const geometryModule: any = await import('/src/motion/die-geometry.ts');
    const geometry = geometryModule.dieGeometry(count);
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
    return {
      expectedReadable: geometry.faces.length,
      expectedTotal: geometry.faces.length + geometry.capFaces.length,
      expectedPolygonSizes: [...geometry.faces, ...geometry.capFaces]
        .map((f: any) => f.polygon.length)
        .sort((a: number, b: number) => a - b),
      readable: facets.filter((el) => el.dataset.faceIndex !== undefined).length,
      total: facets.length,
      clipPathPolygonSizes: facets
        .map((el) => {
          const clip = getComputedStyle(el).clipPath;
          if (!clip.startsWith('polygon(')) return -1;
          return clip.slice('polygon('.length, -1).split(',').length;
        })
        .sort((a, b) => a - b),
      faceValues: facets
        .filter((el) => el.dataset.faceIndex !== undefined)
        .map((el) => Number(el.dataset.faceValue)),
    };
  }, faceCount);
}

// Reads every facet's RENDERED box and compares it against the geometry's own
// polygons.
//
// The comparison is deliberately not "does the component's own facetStyle say
// what facetStyle says". For each facet it takes:
//
//   - from the DOM: the box's used width/height in px, and the 4x4 matrix the
//     browser resolved for it, whose first three columns are the axes the box
//     is rotated onto (u, v) and its normal (w), and whose last column is
//     where the box centre was moved to;
//   - from `die-geometry.ts`: the facet's polygon, scaled the one way the
//     component documents (`0.5 / circumradius` per em, `1em` = --die-size)
//     and turned into CSS space by the documented (x, -y, -z).
//
// and requires that the polygon, projected onto the rendered u/v axes about
// the rendered box centre, spans that box EXACTLY -- from -width/2 to +width/2
// and -height/2 to +height/2 -- and that every clip-path vertex lands on the
// polygon vertex it is cut from. Reading (u, v) off the render rather than
// recomputing them is what makes this basis-agnostic: the component may orient
// a facet's box however it likes inside the facet's plane, and the assertion
// still holds; what it may NOT do is choose the box's SIZE or CENTRE, which
// belong to the geometry. A square-facet assumption fails the extent check on
// the shorter axis of every non-square facet, and a d7's side facets are
// 2.7:1.
async function facetBoxes(page: import('@playwright/test').Page, faceCount: number) {
  return await page.evaluate(async (count) => {
    const geometryModule: any = await import('/src/motion/die-geometry.ts');
    const geometry = geometryModule.dieGeometry(count);
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
    // 1em is the die's size: #stage sets font-size, so this is the one number
    // that turns the geometry's own units into the px the browser laid out.
    const emPx = parseFloat(getComputedStyle(root.querySelector('#stage') as HTMLElement).fontSize);
    const unitsToPx = (0.5 / geometry.circumradius) * emPx;
    const surface = [...geometry.faces, ...geometry.capFaces];

    const dot = (a: number[], b: number[]) => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
    const rows: any[] = [];
    for (const [index, face] of surface.entries()) {
      const element = facets[index];
      if (!element) { rows.push({ index, missing: true }); continue; }
      const style = getComputedStyle(element);
      const matrix = new DOMMatrix(style.transform);
      const u = [matrix.m11, matrix.m12, matrix.m13];
      const v = [matrix.m21, matrix.m22, matrix.m23];
      const w = [matrix.m31, matrix.m32, matrix.m33];
      const centre = [matrix.m41, matrix.m42, matrix.m43];
      const width = parseFloat(style.width);
      const height = parseFloat(style.height);

      // The geometry's polygon, in the same frame the box centre is given in:
      // px, about the solid's centre, CSS axes (x, -y, -z).
      const projected: number[][] = face.polygon.map((point: number[]) => {
        const screen = [point[0] * unitsToPx, -point[1] * unitsToPx, -point[2] * unitsToPx];
        const offset = [0, 1, 2].map((i) => screen[i] - centre[i]);
        return [dot(offset, u), dot(offset, v)];
      });
      const spans = [0, 1].map((axis) => {
        const values = projected.map((p) => p[axis]);
        return [Math.min(...values), Math.max(...values)];
      });

      // Where the clip-path puts each vertex, in the same box-centred frame.
      const clip = style.clipPath;
      const clipPoints = clip.startsWith('polygon(')
        ? clip.slice('polygon('.length, -1).split(',').map((pair) => {
            const [x, y] = pair.trim().split(/\s+/).map((value) => parseFloat(value) / 100);
            return [x * width - width / 2, y * height - height / 2];
          })
        : null;

      rows.push({
        index,
        faceIndex: element.dataset.faceIndex ?? null,
        width,
        height,
        // The box must be exactly the polygon's extent along the axes it was
        // rotated onto, centred where the polygon's extent is centred.
        boxError: Math.max(
          Math.abs(spans[0][0] + width / 2), Math.abs(spans[0][1] - width / 2),
          Math.abs(spans[1][0] + height / 2), Math.abs(spans[1][1] - height / 2),
        ),
        clipError: clipPoints === null || clipPoints.length !== projected.length
          ? Number.POSITIVE_INFINITY
          : Math.max(...clipPoints.map((p, i) =>
              Math.max(Math.abs(p[0] - projected[i][0]), Math.abs(p[1] - projected[i][1])))),
        // The placement must be a rotation, never a reflection: a mirrored
        // facet draws its glyphs backwards.
        orthonormalError: Math.max(
          Math.abs(dot(u, u) - 1), Math.abs(dot(v, v) - 1), Math.abs(dot(w, w) - 1),
          Math.abs(dot(u, v)), Math.abs(dot(u, w)), Math.abs(dot(v, w)),
          Math.abs(dot([
            u[1] * v[2] - u[2] * v[1],
            u[2] * v[0] - u[0] * v[2],
            u[0] * v[1] - u[1] * v[0],
          ], w) - 1),
        ),
        // Reported so a failure says WHICH way the box is wrong, and so the
        // "these facets are not square in the first place" guard below has
        // something to measure.
        aspect: Math.max(width, height) / Math.min(width, height),
        geometryAspect: Math.max(spans[0][1] - spans[0][0], spans[1][1] - spans[1][0])
          / Math.min(spans[0][1] - spans[0][0], spans[1][1] - spans[1][0]),
      });
    }
    return { rows, emPx, facetCount: facets.length, surfaceCount: surface.length };
  }, faceCount);
}

// Which facets the die actually SHOWS, by hit-testing a grid over its box.
//
// Counted by share of the hit area, not by "was hit at all": a facet a few
// degrees off edge-on is a sliver a pixel or two wide, which is not a facet a
// player can see. Before the resting pose was fixed, a d4's second facet
// covered 0% to 1.7% of the die (it was a back-face, culled outright, or the
// sliver at the silhouette); after, 16% to 33%.
async function visibleFacets(
  page: import('@playwright/test').Page,
  options: { minShare: number },
) {
  return await page.evaluate(async (opts) => {
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const box = (root.querySelector('#main') as HTMLElement).getBoundingClientRect();
    const hits = new Map<string, number>();
    const N = 60;
    for (let i = 0; i < N; i++) {
      for (let j = 0; j < N; j++) {
        const hit = (root as any).elementFromPoint(
          box.left + ((i + 0.5) / N) * box.width,
          box.top + ((j + 0.5) / N) * box.height,
        ) as Element | null;
        const facet = hit?.closest?.('.facet') as HTMLElement | null;
        if (!facet) continue;
        const key = facet.dataset.faceIndex ?? `cap${Array.from(root.querySelectorAll('.facet')).indexOf(facet)}`;
        hits.set(key, (hits.get(key) ?? 0) + 1);
      }
    }
    const total = [...hits.values()].reduce((a, b) => a + b, 0);
    const shares = Object.fromEntries(
      [...hits.entries()].map(([key, value]) => [key, Number((value / total).toFixed(3))]),
    );
    return {
      distinctFacetsVisible: [...hits.values()].filter((value) => value / total >= opts.minShare).length,
      presentedShare: total ? (hits.get(String(die.selectedFace)) ?? 0) / total : 0,
      shares,
    };
  }, options);
}

// How the PRESENTED facet ends up on screen, once the pose has been applied.
//
// `facetBoxes` above reads a facet's own transform, which is in the solid's
// body frame; what a player sees is that composed with everything between the
// facet and #stage (#orient's resting pose, and #inner, which the animation
// kernel owns). So this walks the chain and multiplies it out.
//
// `rollDegrees` is the angle the facet's local +y -- the direction its content
// reads DOWNWARDS in, since CSS y points down -- makes with screen-down after
// that composition. Zero means the numeral on the presented face is upright;
// 180 means it is upside down.
async function presentedFacetPose(page: import('@playwright/test').Page) {
  return await page.evaluate(() => {
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
    const element = facets[die.selectedFace] as HTMLElement;
    const chain: HTMLElement[] = [];
    for (let node: HTMLElement | null = element; node && node.id !== 'stage'; node = node.parentElement) {
      chain.unshift(node);
    }
    let matrix = new DOMMatrix();
    for (const node of chain) matrix = matrix.multiply(new DOMMatrix(getComputedStyle(node).transform));
    const v = [matrix.m21, matrix.m22, matrix.m23];
    const w = [matrix.m31, matrix.m32, matrix.m33];
    return {
      chain: chain.map((node) => node.id || node.className),
      // atan2(x, y): measured from screen-down (0, +1), positive clockwise.
      rollDegrees: (Math.atan2(v[0], v[1]) * 180) / Math.PI,
      projectedLength: Math.hypot(v[0], v[1]),
      towardsCamera: w[2],
    };
  });
}

test.describe('boardgame-die solid', () => {
  for (const faceCount of [6, 20, 7]) {
    test(`d${faceCount} renders one element per surface polygon, each cut to its outline`, async ({ page }) => {
      await mountDie(page, { faceCount });
      const inventory = await facetInventory(page, faceCount);

      expect(inventory.readable).toBe(faceCount);
      expect(inventory.readable).toBe(inventory.expectedReadable);
      expect(inventory.total).toBe(inventory.expectedTotal);
      // Every facet is clipped, and clipped to the SAME polygon arity the
      // geometry reports -- a triangle stays a triangle and a pentagon stays
      // a pentagon, which is what "one general routine" has to mean.
      expect(inventory.clipPathPolygonSizes).toEqual(inventory.expectedPolygonSizes);
      // Face VALUES, not indices: a d7 fixture's faces are 10..70.
      expect(inventory.faceValues.slice().sort((a, b) => a - b))
        .toEqual(Array.from({ length: faceCount }, (_, i) => (i + 1) * 10));
    });
  }

  // The assertion the vertex-count check above cannot make. Forcing every
  // facet's box square -- `const sq = Math.max(width, height)` in facetStyle,
  // the exact square-facet assumption this component exists to avoid -- leaves
  // every clip-path vertex COUNT untouched, so the test above stays green
  // while a d7 renders as disconnected slabs and a d6/d12/d20 opens seam gaps.
  // This one measures the box itself against the geometry's own extents.
  for (const faceCount of [6, 20, 12, 10, 7]) {
    test(`d${faceCount} draws each facet at the geometry's own in-plane extents`, async ({ page }) => {
      // 200px so a facet's shorter side is tens of px: the tolerances below
      // are absolute px, and a 50px die would put them near the noise floor.
      await mountDie(page, { faceCount, dieSize: '200px' });
      const { rows, emPx, facetCount, surfaceCount } = await facetBoxes(page, faceCount);

      expect(emPx).toBe(200);
      expect(facetCount).toBe(surfaceCount);
      // DOM order is surface order -- `[...faces, ...capFaces]` -- and the
      // readable prefix of it carries the face indices. Everything below pairs
      // facet i with surface polygon i, so pin that pairing first.
      expect(rows.map((row: any) => row.faceIndex)).toEqual(
        rows.map((_: any, index: number) => (index < faceCount ? String(index) : null)),
      );

      // Sub-pixel on a 200px die. Not exact-equal because the browser reports
      // used px for an em length and the component rounds its own output to
      // five decimals.
      const worst = (key: string) => rows.reduce(
        (best: any, row: any) => (row[key] > best[key] ? row : best), rows[0]);
      const detail = (row: any) =>
        `facet ${row.index}: ${row.width.toFixed(2)}x${row.height.toFixed(2)}px,`
        + ` rendered aspect ${row.aspect.toFixed(3)} vs geometry ${row.geometryAspect.toFixed(3)}`;
      const worstBox = worst('boxError');
      const worstClip = worst('clipError');
      expect(worstBox.boxError, `box vs geometry extents -- ${detail(worstBox)}`)
        .toBeLessThanOrEqual(0.05);
      expect(worstClip.clipError, `clip vertices vs geometry vertices -- ${detail(worstClip)}`)
        .toBeLessThanOrEqual(0.05);
      // Loose because Chromium serializes a computed matrix to about six
      // significant figures (measured worst case 9e-6), and tight enough by a
      // wide margin for what it is looking for: a reflection puts the
      // determinant at -1, i.e. an error of 2.
      expect(Math.max(...rows.map((row: any) => row.orthonormalError)),
        'facet placement is a rotation, not a reflection or a scale')
        .toBeLessThan(1e-3);

      // And the guard that keeps the above from passing vacuously: it only
      // says anything if the geometry's facets are NOT square to begin with.
      // Measured from the geometry, so it moves with the geometry: a d7's
      // side facets are 2.73:1 today, a d10's kites 1.27:1, a d20's triangles
      // 1.07 and 1.16:1, a d12's pentagons 1.05:1 -- and a d6's squares 1:1,
      // which is why the d6 case asserts nothing here.
      if (faceCount !== 6) {
        expect(Math.max(...rows.map((row: any) => row.geometryAspect))).toBeGreaterThan(1.05);
      }
      for (const row of rows) {
        expect(row.aspect, detail(row)).toBeCloseTo(row.geometryAspect, 2);
      }
    });
  }

  // A solid whose faces are far apart in normal angle -- a tetrahedron's are
  // 109.47 degrees apart -- puts every non-presented face behind the camera at
  // the fixed resting tilt, where backface-visibility culls it, so a d4 used to
  // render as ONE flat triangle: the 2D die this component replaces. The pose
  // now tilts as far as it takes for the most face-on of the others to come
  // into view (and not one degree further: shapes with closer normals, which is
  // every other shape, are untouched).
  for (const faceCount of [4, 6, 8, 12, 20]) {
    test(`d${faceCount} rests showing more than one facet`, async ({ page }) => {
      // Every face of the d4 in turn: which of the other three is best placed
      // depends on the presented face's own orientation, and the pose has to
      // work for all of them.
      const selectedFaces = faceCount === 4 ? [0, 1, 2, 3] : [1];
      for (const selectedFace of selectedFaces) {
        await mountDie(page, { faceCount, selectedFace, dieSize: '200px' });
        const visible = await visibleFacets(page, { minShare: 0.05 });
        expect(visible.distinctFacetsVisible,
          `d${faceCount} presenting face ${selectedFace}, hit-area share per facet: ${JSON.stringify(visible.shares)}`)
          .toBeGreaterThanOrEqual(2);
        // The presented face is the value the player reads, so it stays the
        // dominant one. Asserted for the d4 only: it is the shape the tilt
        // touches, and it is the one whose facets are big enough that
        // Chromium's hit testing of a preserve-3d subtree is dependable --
        // on a d12 or d20 it misses whole facets in some orientations.
        if (faceCount === 4) expect(visible.presentedShare).toBeGreaterThan(0.5);
      }
    });
  }

  // The pose has to leave the presented face's NUMBER the right way up, and
  // pointing its normal at the camera by the shortest path does not: the
  // minimal turn carries the facet's own +y wherever it happens to land. Before
  // the presentation roll, measured here, a d4 presenting face 1 was 122
  // degrees out and a d10 presenting face 2 was 116 -- upside-down numbers on
  // the one face the player is meant to read -- while a d20 was within 16, which
  // is exactly why eyeballing one shape proves nothing.
  //
  // Only the RESTING pose is pinned. After a physics roll lands the die, the
  // content roll is whatever the simulation stopped at, the same as a real
  // die's; this is the pre-roll pose that has to read like the flat 2D die the
  // solid replaces.
  for (const faceCount of [4, 6, 7, 8, 10, 12, 20]) {
    test(`d${faceCount} rests with the presented face's content upright`, async ({ page }) => {
      const worst: string[] = [];
      // Several faces per shape: the roll depends on the presented facet's own
      // orientation, so one sample says nothing about the next.
      for (const selectedFace of [0, 1, 2, 3, faceCount - 1]) {
        if (selectedFace >= faceCount) continue;
        await mountDie(page, { faceCount, selectedFace, dieSize: '200px' });
        const pose = await presentedFacetPose(page);
        // Facing the viewer at all: a roll measured on a facet that is edge-on
        // or turned away would be measuring nothing.
        expect(pose.towardsCamera,
          `d${faceCount} face ${selectedFace} presented normal .z`).toBeGreaterThan(0.5);
        expect(pose.projectedLength,
          `d${faceCount} face ${selectedFace} local +y has a screen direction`).toBeGreaterThan(0.5);
        worst.push(`face ${selectedFace}: ${pose.rollDegrees.toFixed(2)}`);
        // Not "within 35 degrees of upright" -- the pose CHOOSES this roll, so
        // anything but zero is a bug, and the slack is Chromium's matrix
        // serialization (about six significant figures).
        expect(Math.abs(pose.rollDegrees),
          `d${faceCount} presented-face content roll off screen-up, per face: ${worst.join(', ')}`)
          .toBeLessThan(0.5);
      }
    });
  }

  test('#inner carries a live preserve-3d context that depth-sorts the facets', async ({ page }) => {
    // selectedFace 3 is deliberately neither the first nor the last facet in
    // DOM order: under a FLATTENED context the facets paint in DOM order and
    // the last one wins the centre of the die, so this assertion fails.
    // Mounted large so the perspective probe below has room: the projection
    // shift it measures is a fraction of the die's size.
    await mountDie(page, { faceCount: 6, selectedFace: 3, dieSize: '160px' });
    const result = await page.evaluate(() => {
      const die = document.getElementById('fixture-die') as any;
      const root = die.shadowRoot as ShadowRoot;
      const inner = root.querySelector('#inner') as HTMLElement;
      const main = root.querySelector('#main') as HTMLElement;
      const box = main.getBoundingClientRect();
      const hit = (root as any).elementFromPoint(
        box.left + box.width / 2,
        box.top + box.height / 2,
      ) as Element | null;
      const facet = hit?.closest?.('.facet') as HTMLElement | null;

      // Second, occlusion-independent probe that the 3D context is LIVE all
      // the way from the perspective wrapper down to the facets: switch the
      // perspective off and the projection must change. It cannot, if
      // anything between #stage and a facet has flattened -- a flattened
      // context ignores the perspective entirely. Concretely, on a cube
      // without perspective the three pairs of opposite faces project to
      // pairwise-IDENTICAL rectangles; with it, the nearer of each pair is
      // measurably bigger.
      const stage = root.querySelector('#stage') as HTMLElement;
      const sizes = () => (Array.from(root.querySelectorAll('.facet')) as HTMLElement[])
        .map((el) => { const r = el.getBoundingClientRect(); return [r.width, r.height]; });
      const projected = sizes();
      stage.style.perspective = 'none';
      const flattened = sizes();
      stage.style.perspective = '';
      const worstShift = Math.max(...projected.flatMap((pair, i) =>
        pair.map((value, j) => Math.abs(value - flattened[i][j]))));

      return {
        transformStyle: getComputedStyle(inner).transformStyle,
        frontFaceIndex: facet?.dataset.faceIndex ?? null,
        frontFaceValue: facet?.dataset.faceValue ?? null,
        hitDescription: hit ? `${hit.tagName.toLowerCase()}.${hit.className}` : 'nothing',
        domOrderLastFaceIndex: (Array.from(root.querySelectorAll('.facet')) as HTMLElement[])
          .filter((el) => el.dataset.faceIndex !== undefined)
          .at(-1)?.dataset.faceIndex ?? null,
        worstShift,
      };
    });

    expect(result.transformStyle).toBe('preserve-3d');
    // Measured 7.0px on a 160px d6; 0 exactly if anything has flattened.
    expect(result.worstShift).toBeGreaterThan(4);
    // The die presents the SELECTED face, and selectedFace is an index: face
    // index 3 carries value 40 in this fixture.
    expect(`${result.frontFaceIndex} (hit ${result.hitDescription})`)
      .toBe(`3 (hit ${result.hitDescription})`);
    expect(result.frontFaceValue).toBe('40');
    expect(result.domOrderLastFaceIndex).not.toBe('3');
  });

  test('--die-size is a real custom property a caller can set', async ({ page }) => {
    await mountDie(page, { faceCount: 6 });
    const defaults = await page.evaluate(() => {
      const root = (document.getElementById('fixture-die') as any).shadowRoot as ShadowRoot;
      const main = root.querySelector('#main') as HTMLElement;
      const facet = root.querySelector('.facet') as HTMLElement;
      return { box: main.offsetWidth, facet: facet.offsetWidth };
    });
    expect(defaults.box).toBe(DIE_SIZE_DEFAULT_PX);

    await mountDie(page, { faceCount: 6, dieSize: '160px' });
    const scaled = await page.evaluate(() => {
      const root = (document.getElementById('fixture-die') as any).shadowRoot as ShadowRoot;
      const main = root.querySelector('#main') as HTMLElement;
      const facet = root.querySelector('.facet') as HTMLElement;
      return { box: main.offsetWidth, facet: facet.offsetWidth };
    });
    expect(scaled.box).toBe(160);
    // The SOLID scales with the property, not just the box around it.
    expect(scaled.facet / defaults.facet).toBeCloseTo(160 / DIE_SIZE_DEFAULT_PX, 1);
  });

  // The reel's face-change spin is UNCHANGED by this task -- same track, same
  // hooks, so the trace goldens cannot move -- but a solid has no reel to
  // scroll, and letting that translateY through would slide the die by a
  // multiple of its own size and back on every roll. #inner.solid zeroes
  // --reel-step instead, which both the CSS resting rule and the keyframes
  // read, and which is a variable of its own precisely so that zeroing it
  // cannot zero anything else the solid measures against the die's size.
  test('the face-change spin still plays, and does not displace the solid', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0 });
    const result = await page.evaluate(async () => {
      const die = document.getElementById('fixture-die') as any;
      const root = die.shadowRoot as ShadowRoot;
      const inner = root.querySelector('#inner') as HTMLElement;
      const at = () => getComputedStyle(inner).transform;
      const before = at();
      die.item = {
        Values: { Faces: die.faces },
        DynamicValues: { SelectedFace: 5, Value: die.faces[5] },
      };
      await die.updateComplete;
      await die.updateComplete;
      const animations = inner.getAnimations();
      const keyframes = animations[0]?.effect instanceof KeyframeEffect
        ? animations[0].effect.getKeyframes().map((frame: any) => frame.transform)
        : [];
      await new Promise((resolve) => setTimeout(resolve, 80));
      const during = at();
      await Promise.all(animations.map((a) => a.finished.catch(() => undefined)));
      return { before, during, after: at(), count: animations.length, keyframes };
    });

    expect(result.count).toBe(1);
    expect(result.keyframes).toEqual([
      'translateY(calc(-1 * var(--reel-step) * 0))',
      'translateY(calc(-1 * var(--reel-step) * 5))',
    ]);
    expect([result.before, result.during, result.after])
      .toEqual(['matrix(1, 0, 0, 1, 0, 0)', 'matrix(1, 0, 0, 1, 0, 0)', 'matrix(1, 0, 0, 1, 0, 0)']);
  });

  test('falls back to the reel for a die with fewer than three faces', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-die.ts');
      const errors: string[] = [];
      const onError = (e: ErrorEvent) => errors.push(String(e.message));
      window.addEventListener('error', onError);
      const mount = async (faces: number[] | null) => {
        const die = document.createElement('boardgame-die') as any;
        die.item = faces === null ? null : {
          Values: { Faces: faces },
          DynamicValues: { SelectedFace: 0, Value: faces[0] },
        };
        document.body.appendChild(die);
        await die.updateComplete;
        await die.updateComplete;
        const root = die.shadowRoot as ShadowRoot;
        const faceEls = Array.from(root.querySelectorAll('#inner.reel .face')) as HTMLElement[];
        const out = {
          facets: root.querySelectorAll('.facet').length,
          reelFaces: faceEls.length,
          hasInner: !!root.querySelector('#inner'),
          // The reel goes through the SAME content resolution as the solid --
          // it is a fallback, not a second implementation -- so a regression
          // in the shared path has to show up here too.
          pipsPerFace: faceEls.map((el) => el.querySelectorAll('.pip').length),
          // A dot's diameter and its offset from the face's centre, in px, on
          // the default 50px die. These are the flat die's original numbers
          // (a 7px dot 10.5px off centre), which is what --content-size: 63%
          // of a reel face is chosen to reproduce.
          pipGeometry: faceEls.length === 0 ? null : (() => {
            const content = faceEls[1].querySelector('.content') as HTMLElement;
            const pip = content.querySelector('.pip') as HTMLElement;
            const box = content.getBoundingClientRect();
            const dot = pip.getBoundingClientRect();
            return {
              diameter: Math.round(dot.width * 10) / 10,
              offsetX: Math.round((dot.left + dot.width / 2 - (box.left + box.width / 2)) * 10) / 10,
            };
          })(),
        };
        die.remove();
        return out;
      };
      const twoFaced = await mount([1, 2]);
      const empty = await mount(null);
      window.removeEventListener('error', onError);
      return { twoFaced, empty, errors };
    });

    expect(result.errors).toEqual([]);
    expect(result.twoFaced).toEqual({
      facets: 0,
      reelFaces: 2,
      hasInner: true,
      // Face value 1 draws one dot, value 2 draws two: computed, in the reel
      // as much as on the solid.
      pipsPerFace: [1, 2],
      pipGeometry: { diameter: 6.3, offsetX: -10.5 },
    });
    expect(result.empty).toEqual({
      facets: 0, reelFaces: 0, hasInner: true, pipsPerFace: [], pipGeometry: null,
    });
  });

  test('the solid survives mounted in the real app (pig)', async ({ page }) => {
    test.setTimeout(120000);
    await createOfflineGame(page, 'pig');
    const result = await page.evaluate(() => {
      const deepQueryFirst = (root: Document | ShadowRoot | Element, selector: string): Element | null => {
        const direct = root.querySelector(selector);
        if (direct) return direct;
        for (const el of Array.from(root.querySelectorAll('*'))) {
          if ((el as any).shadowRoot) {
            const found = deepQueryFirst((el as any).shadowRoot, selector);
            if (found) return found;
          }
        }
        return null;
      };
      const die = deepQueryFirst(document, 'boardgame-die') as HTMLElement | null;
      if (!die) return { found: false } as any;
      // The game view mounts scrolled to the admin panel, so the board can be
      // outside the viewport, where elementFromPoint answers null.
      die.scrollIntoView({ block: 'center' });
      const root = die.shadowRoot as ShadowRoot;
      const inner = root.querySelector('#inner') as HTMLElement;
      const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
      const main = root.querySelector('#main') as HTMLElement;
      const box = main.getBoundingClientRect();
      const hit = (root as any).elementFromPoint(
        box.left + box.width / 2,
        box.top + box.height / 2,
      ) as Element | null;
      const facet = hit?.closest?.('.facet') as HTMLElement | null;
      const selectedFace = (die as any).selectedFace;
      // Ancestors that flatten a 3D context do so by carrying a grouping
      // property; report the chain so a failure names the culprit.
      const flatteners: string[] = [];
      let node: Element | null = inner;
      while (node) {
        const style = getComputedStyle(node);
        const reasons: string[] = [];
        if (style.overflow !== 'visible') reasons.push(`overflow:${style.overflow}`);
        if (style.filter !== 'none') reasons.push(`filter:${style.filter}`);
        if (Number(style.opacity) < 1) reasons.push(`opacity:${style.opacity}`);
        if (style.clipPath !== 'none' && node !== inner) reasons.push(`clip-path:${style.clipPath}`);
        if (reasons.length) flatteners.push(`${node.tagName.toLowerCase()}#${node.id || ''}: ${reasons.join(',')}`);
        node = node.parentElement
          ?? ((node.getRootNode() as ShadowRoot).host as Element | null)
          ?? null;
      }
      return {
        found: true,
        transformStyle: getComputedStyle(inner).transformStyle,
        facetCount: facets.length,
        clipped: facets.every((el) => getComputedStyle(el).clipPath.startsWith('polygon(')),
        frontFaceIndex: facet?.dataset.faceIndex ?? null,
        hitDescription: hit ? `${hit.tagName.toLowerCase()}.${hit.className}` : 'nothing',
        selectedFace: String(selectedFace),
        innerFlatteners: flatteners,
      };
    });

    expect(result.found).toBe(true);
    expect(result.transformStyle).toBe('preserve-3d');
    // pig's die is a d6: six readable faces, no cap triangles.
    expect(result.facetCount).toBe(6);
    expect(result.clipped).toBe(true);
    // Depth sorting works in situ: the presented face is the one at the
    // centre of the die, not whichever facet paints last.
    expect(`${result.frontFaceIndex} (hit ${result.hitDescription}; flatteners above #inner: ${JSON.stringify(result.innerFlatteners)})`)
      .toBe(`${result.selectedFace} (hit ${result.hitDescription}; flatteners above #inner: ${JSON.stringify(result.innerFlatteners)})`);
  });
});

// ---------------------------------------------------------------------------
// Task 9: what is PAINTED on those facets.
//
// The die used to top out at six because its pip layouts were six hard-coded
// CSS classes (`.face.one` ... `.face.six`). Content is now resolved per face
// -- author symbol set, then generated pips, then numerals -- and the pip
// layout is COMPUTED from the value on a 3x3 lattice rather than enumerated.
//
// The assertions below are deliberately geometric rather than structural. A
// test that merely counts `.pip` elements passes a layout that stacks every
// dot in one corner, and a test that merely finds a numeral passes a numeral
// drawn half outside a d7's 2.7:1 side facet. So every one of these measures
// where the content actually landed, in the facet's OWN plane, against the
// polygon `die-geometry.ts` reports for that facet.
// ---------------------------------------------------------------------------

/**
 * Every readable facet's painted content, measured in the facet's own plane.
 *
 * The frame is the one `facetStyle` works in: px, origin at the polygon's
 * centroid, +a along the axis the box's local x was rotated onto and +b along
 * its local y. The polygon comes from `die-geometry.ts` and the content
 * positions from the laid-out DOM, so the two are independent measurements
 * that have to agree.
 */
async function faceContent(page: import('@playwright/test').Page, faceCount: number) {
  return await page.evaluate(async (count) => {
    const geometryModule: any = await import('/src/motion/die-geometry.ts');
    const faceModule: any = await import('/src/motion/die-faces.ts');
    const geometry = geometryModule.dieGeometry(count);
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const stage = root.querySelector('#stage') as HTMLElement;
    const emPx = parseFloat(getComputedStyle(stage).fontSize);
    const unitsToPx = (0.5 / geometry.circumradius) * emPx;
    const dot = (a: number[], b: number[]) => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];

    // Which face a die in this orientation is READ from is a property of the
    // solid, not of the renderer: `die-faces.ts` owns it.
    const readingRule = faceModule.resolveReadingRule(geometry);

    // For each VERTEX of the solid, the face that is read when that vertex is
    // the topmost point -- the face whose normal is most opposed to it. That
    // is the value a corner-printed die has to show at that corner.
    const faceReadAtVertex = geometry.vertices.map((vertex: number[]) => {
      let best = 0;
      let bestScore = Infinity;
      for (let i = 0; i < geometry.faces.length; i++) {
        const score = dot(geometry.faces[i].normal, vertex);
        if (score < bestScore) { bestScore = score; best = i; }
      }
      return best;
    });

    const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
    const px = (el: HTMLElement, prop: string) => parseFloat(getComputedStyle(el).getPropertyValue(prop));

    const rows = geometry.faces.map((face: any, index: number) => {
      const element = facets[index];
      const style = getComputedStyle(element);
      const matrix = new DOMMatrix(style.transform);
      const u = [matrix.m11, matrix.m12, matrix.m13];
      const v = [matrix.m21, matrix.m22, matrix.m23];
      const width = parseFloat(style.width);
      const height = parseFloat(style.height);

      // The polygon in the facet's plane, about its centroid.
      const centroid = [face.centroid[0], -face.centroid[1], -face.centroid[2]]
        .map((value: number) => value * unitsToPx);
      const polygon = face.polygon.map((point: number[]) => {
        const screen = [point[0] * unitsToPx, -point[1] * unitsToPx, -point[2] * unitsToPx];
        const offset = [0, 1, 2].map((i) => screen[i] - centroid[i]);
        return [dot(offset, u), dot(offset, v)];
      });
      const boxA = (Math.min(...polygon.map((p: number[]) => p[0]))
        + Math.max(...polygon.map((p: number[]) => p[0]))) / 2;
      const boxB = (Math.min(...polygon.map((p: number[]) => p[1]))
        + Math.max(...polygon.map((p: number[]) => p[1]))) / 2;
      // DOM x measured from the box's left edge -> the facet's own (a, b).
      const toLocal = (x: number, y: number) => [x - width / 2 + boxA, y - height / 2 + boxB];

      const contentEl = element.querySelector('.content') as HTMLElement | null;
      let content: any = null;
      let pips: any[] = [];
      let text: any = null;
      if (contentEl) {
        const cw = px(contentEl, 'width');
        const ch = px(contentEl, 'height');
        const cl = px(contentEl, 'left');
        const ct = px(contentEl, 'top');
        const [ccx, ccy] = toLocal(cl + cw / 2, ct + ch / 2);
        content = { cx: ccx, cy: ccy, width: cw, height: ch, left: cl, top: ct };
        pips = (Array.from(contentEl.querySelectorAll('.pip')) as HTMLElement[]).map((pip) => {
          const pw = px(pip, 'width');
          const [cx, cy] = toLocal(cl + px(pip, 'left') + pw / 2, ct + px(pip, 'top') + px(pip, 'height') / 2);
          // Where the dot sits on the 3x3 lattice of the content square,
          // measured from the laid-out boxes -- not read off any class name.
          return {
            cx, cy, r: pw / 2,
            col: Math.round(((cx - ccx) / cw) * 3 + 1),
            row: Math.round(((cy - ccy) / ch) * 3 + 1),
          };
        });
        const span = contentEl.querySelector('span') as HTMLElement | null;
        if (span && span.textContent) {
          const [cx, cy] = toLocal(cl + span.offsetLeft + span.offsetWidth / 2,
            ct + span.offsetTop + span.offsetHeight / 2);
          text = {
            text: span.textContent, cx, cy,
            width: span.offsetWidth, height: span.offsetHeight,
          };
        }
      }

      // Measured on the TEXT, not on the box that holds it: a box that is
      // inside the polygon while its glyph spills out of it is exactly the
      // failure this is looking for.
      const corners = (Array.from(element.querySelectorAll('.corner')) as HTMLElement[]).map((corner) => {
        const span = corner.querySelector('span') as HTMLElement;
        const [cx, cy] = toLocal(px(corner, 'left') + span.offsetLeft + span.offsetWidth / 2,
          px(corner, 'top') + span.offsetTop + span.offsetHeight / 2);
        return {
          text: corner.textContent,
          faceIndex: Number(corner.dataset.cornerFaceIndex),
          cx, cy, width: span.offsetWidth, height: span.offsetHeight,
        };
      });

      return {
        faceIndex: Number(element.dataset.faceIndex),
        faceValue: Number(element.dataset.faceValue),
        label: element.dataset.faceLabel ?? null,
        polygon, width, height,
        // Every vertex of this facet, paired with the face that is read when
        // that vertex is uppermost.
        vertexFaces: face.polygon.map((point: number[]) => {
          let best = -1;
          let bestDistance = Infinity;
          geometry.vertices.forEach((vertex: number[], i: number) => {
            const d = Math.hypot(vertex[0] - point[0], vertex[1] - point[1], vertex[2] - point[2]);
            if (d < bestDistance) { bestDistance = d; best = i; }
          });
          return faceReadAtVertex[best];
        }),
        content, pips, text, corners,
      };
    });

    const button = root.querySelector('#main') as HTMLElement;
    return {
      readingRule,
      ariaLabel: button.getAttribute('aria-label'),
      selectedFace: die.selectedFace,
      capsWithContent: facets.slice(geometry.faces.length)
        .filter((el) => el.querySelector('.content, .corner')).length,
      rows,
    };
  }, faceCount);
}

/**
 * How far inside its polygon a point is, in px: positive inside, negative out.
 * Convex polygons only, which is every facet `die-geometry.ts` produces.
 */
function insideBy(polygon: number[][], point: [number, number]): number {
  let area2 = 0;
  for (let i = 0; i < polygon.length; i++) {
    const p = polygon[i];
    const q = polygon[(i + 1) % polygon.length];
    area2 += p[0] * q[1] - q[0] * p[1];
  }
  const sign = area2 >= 0 ? 1 : -1;
  let worst = Infinity;
  for (let i = 0; i < polygon.length; i++) {
    const p = polygon[i];
    const q = polygon[(i + 1) % polygon.length];
    const ex = q[0] - p[0];
    const ey = q[1] - p[1];
    const length = Math.hypot(ex, ey);
    if (!(length > 0)) continue;
    // Inward normal of edge p->q for this winding.
    const nx = (-ey / length) * sign;
    const ny = (ex / length) * sign;
    worst = Math.min(worst, nx * (point[0] - p[0]) + ny * (point[1] - p[1]));
  }
  return worst;
}

/**
 * Every readable facet's in-plane axes as the browser resolved them, paired
 * with the facet's own outward normal from `die-geometry.ts`.
 *
 * `facetBoxes` reads (u, v) off the render on purpose, so that its box and
 * clip assertions hold whichever way the component turns a facet's box inside
 * the facet's plane. That freedom is real for a BOX and false for CONTENT: the
 * marks are laid out axis-aligned in the box, so the box's roll IS the
 * numeral's roll, and a quarter turn of it renders every numeral on the die
 * ninety degrees off while leaving every box, clip-path and pip lattice
 * measurement identical. This is the measurement that sees it.
 */
async function facetAxes(page: import('@playwright/test').Page, faceCount: number) {
  return await page.evaluate(async (count) => {
    const geometryModule: any = await import('/src/motion/die-geometry.ts');
    const geometry = geometryModule.dieGeometry(count);
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
    return geometry.faces.map((face: any, index: number) => {
      const matrix = new DOMMatrix(getComputedStyle(facets[index]).transform);
      return {
        faceIndex: index,
        // The box's local +y: the direction its content reads downwards in.
        v: [matrix.m21, matrix.m22, matrix.m23],
        // The facet's outward normal in CSS space, by the (x, -y, -z) the
        // component documents. From the geometry, not from the render.
        normal: [face.normal[0], -face.normal[1], -face.normal[2]],
      };
    });
  }, faceCount);
}

/**
 * The half-side of the largest axis-aligned square centred on the origin that
 * fits inside a convex polygon given about that origin. Independent of the
 * component: this is the geometry's own answer for how big the content square
 * is allowed to be.
 */
function inscribedSquareHalfSide(polygon: number[][]): number {
  let area2 = 0;
  for (let i = 0; i < polygon.length; i++) {
    const p = polygon[i];
    const q = polygon[(i + 1) % polygon.length];
    area2 += p[0] * q[1] - q[0] * p[1];
  }
  const sign = area2 >= 0 ? 1 : -1;
  let best = Infinity;
  for (let i = 0; i < polygon.length; i++) {
    const p = polygon[i];
    const q = polygon[(i + 1) % polygon.length];
    const ex = q[0] - p[0];
    const ey = q[1] - p[1];
    const length = Math.hypot(ex, ey);
    if (!(length > 0)) continue;
    const nx = (-ey / length) * sign;
    const ny = (ex / length) * sign;
    // The square's farthest corner in the direction of -n clears edge p->q
    // while half * (|nx| + |ny|) <= distance from the origin to the edge.
    best = Math.min(best, (nx * -p[0] + ny * -p[1]) / (Math.abs(nx) + Math.abs(ny)));
  }
  return Number.isFinite(best) ? best : 0;
}

/** The canonical 3x3 pip lattice for a count, as sorted "col,row" cells. */
function expectedPipCells(count: number): string[] {
  const pairs = [
    [[0, 0], [2, 2]],
    [[2, 0], [0, 2]],
    [[0, 1], [2, 1]],
    [[1, 0], [1, 2]],
  ];
  const cells: number[][] = [];
  if (count % 2 === 1) cells.push([1, 1]);
  for (let i = 0; i < Math.floor(count / 2); i++) cells.push(...pairs[i]);
  return cells.map(([c, r]) => `${c},${r}`).sort();
}

test.describe('boardgame-die face content', () => {
  // (a) Pips are GENERATED from the value. Counting dots is not enough -- a
  // layout that piles every dot in one corner counts the same -- so the cell
  // each dot lands on is pinned too, measured from the laid-out boxes.
  test('d6 draws generated pips, one dot per value, on the canonical lattice', async ({ page }) => {
    await mountDie(page, { faceCount: 6, faces: [1, 2, 3, 4, 5, 6], dieSize: '240px' });
    const { rows } = await faceContent(page, 6);
    expect(rows.map((row: any) => row.faceValue)).toEqual([1, 2, 3, 4, 5, 6]);
    for (const row of rows) {
      expect(row.pips.length, `face ${row.faceValue} dot count`).toBe(row.faceValue);
      expect(row.text, `face ${row.faceValue} draws pips, not a numeral`).toBe(null);
      expect(row.pips.map((pip: any) => `${pip.col},${pip.row}`).sort(),
        `face ${row.faceValue} pip cells`).toEqual(expectedPipCells(row.faceValue));
    }
  });

  // The whole point: the layout is not capped at six. A d10 labelled 0..9
  // walks the entire generated range, blank face included.
  test('a d10 labelled 0..9 draws every generated pip count', async ({ page }) => {
    const faces = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9];
    await mountDie(page, { faceCount: 10, faces, dieSize: '240px' });
    const { rows } = await faceContent(page, 10);
    expect(rows.map((row: any) => row.faceValue).sort((a: number, b: number) => a - b)).toEqual(faces);
    for (const row of rows) {
      expect(row.pips.length, `face ${row.faceValue} dot count`).toBe(row.faceValue);
      expect(row.pips.map((pip: any) => `${pip.col},${pip.row}`).sort(),
        `face ${row.faceValue} pip cells`).toEqual(expectedPipCells(row.faceValue));
    }
  });

  // (b) A d20 renders numerals. The threshold is a property of the DIE, not of
  // the individual face: a die that cannot draw all of its values as pips
  // draws all of them as numerals, so no die shows dots on one face and a
  // number on the next.
  test('a d20 renders numerals, not pips', async ({ page }) => {
    await mountDie(page, {
      faceCount: 20,
      faces: Array.from({ length: 20 }, (_, i) => i + 1),
      dieSize: '240px',
    });
    const { rows } = await faceContent(page, 20);
    for (const row of rows) {
      expect(row.pips.length, `face ${row.faceValue} draws no pips`).toBe(0);
      expect(row.text?.text, `face ${row.faceValue} numeral`).toBe(String(row.faceValue));
    }
  });

  test('the pip/numeral threshold is per die: 0..9 pips, one value past it numerals', async ({ page }) => {
    await mountDie(page, { faceCount: 10, faces: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9], dieSize: '240px' });
    const pipped = await faceContent(page, 10);
    expect(pipped.rows.every((row: any) => row.text === null)).toBe(true);
    expect(pipped.rows.reduce((sum: number, row: any) => sum + row.pips.length, 0)).toBe(45);

    // Exactly one value moved past the lattice's capacity; the WHOLE die
    // switches to numerals.
    await mountDie(page, { faceCount: 10, faces: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10], dieSize: '240px' });
    const numeralled = await faceContent(page, 10);
    expect(numeralled.rows.every((row: any) => row.pips.length === 0)).toBe(true);
    expect(numeralled.rows.map((row: any) => row.text?.text).sort())
      .toEqual(['1', '10', '2', '3', '4', '5', '6', '7', '8', '9']);
  });

  // (c) An author-supplied symbol set wins over both, is keyed by the face's
  // NAME (which an enum will supply), and that name is what gets announced.
  test('a supplied symbol set renders glyphs and announces the name', async ({ page }) => {
    await mountDie(page, {
      faceCount: 6,
      faces: [1, 2, 3, 4, 5, 6],
      selectedFace: 2,
      dieSize: '240px',
      faceNames: { 1: 'Sun', 2: 'Moon', 3: 'Star', 4: 'Cloud', 5: 'Rain', 6: 'Snow' },
      symbols: { Sun: '☀', Moon: '☽', Star: '★', Cloud: '☁', Rain: '☂', Snow: '❄' },
    });
    const result = await faceContent(page, 6);
    expect(result.rows.map((row: any) => row.text?.text))
      .toEqual(['☀', '☽', '★', '☁', '☂', '❄']);
    // The symbol set beats the pips these values would otherwise generate.
    expect(result.rows.every((row: any) => row.pips.length === 0)).toBe(true);
    expect(result.rows.map((row: any) => row.label))
      .toEqual(['Sun', 'Moon', 'Star', 'Cloud', 'Rain', 'Snow']);
    // selectedFace is an INDEX: index 2 carries value 3, named Star.
    expect(result.ariaLabel).toBe('Die showing Star');
    // ... and through the real accessibility tree, not just the attribute.
    await expect(page.getByRole('button', { name: 'Die showing Star', exact: true }))
      .toHaveCount(1);
  });

  // (d) The announced value matches what is DRAWN, in all three modes. This is
  // also the assertion that catches an index/value mix-up: the fixtures below
  // select a face whose index differs from its value.
  test('the announced value matches what is shown, in all three content modes', async ({ page }) => {
    // Pips: index 4 carries value 5, drawn as five dots.
    await mountDie(page, { faceCount: 6, faces: [1, 2, 3, 4, 5, 6], selectedFace: 4, dieSize: '240px' });
    const pips = await faceContent(page, 6);
    const pipFace = pips.rows.find((row: any) => row.faceIndex === pips.selectedFace);
    expect(pipFace.faceValue).toBe(5);
    expect(pipFace.pips.length).toBe(5);
    expect(pipFace.label).toBe('5');
    expect(pips.ariaLabel).toBe('Die showing 5');

    // Numerals: index 3 carries value 40.
    await mountDie(page, { faceCount: 20, selectedFace: 3, dieSize: '240px' });
    const numerals = await faceContent(page, 20);
    const numeralFace = numerals.rows.find((row: any) => row.faceIndex === numerals.selectedFace);
    expect(numeralFace.faceValue).toBe(40);
    expect(numeralFace.text.text).toBe('40');
    expect(numerals.ariaLabel).toBe('Die showing 40');

    // Glyphs: index 1 carries value 2, named Moon, drawn as its glyph.
    await mountDie(page, {
      faceCount: 6,
      faces: [1, 2, 3, 4, 5, 6],
      selectedFace: 1,
      dieSize: '240px',
      faceNames: { 2: 'Moon' },
      symbols: { Moon: '☽' },
    });
    const glyph = await faceContent(page, 6);
    const glyphFace = glyph.rows.find((row: any) => row.faceIndex === glyph.selectedFace);
    expect(glyphFace.text.text).toBe('☽');
    expect(glyph.ariaLabel).toBe('Die showing Moon');
    // A named face with NO glyph still announces something a player can tie
    // to what is drawn: the numeral is on the facet, the name is the meaning.
    const unnamed = glyph.rows.find((row: any) => row.faceIndex === 0);
    expect(unnamed.label).toBe('1');
  });

  // Trap 1. Facets are not square and often not rectangles: a d7 side face is
  // 2.37 x 0.87, a d12's are pentagons, a d20's triangles, a d10's kites. A
  // pip grid or a numeral sized to the facet's BOUNDING BOX spills outside the
  // polygon on every one of them. Nothing painted may leave its own facet.
  for (const faceCount of [6, 7, 8, 10, 12, 20, 4]) {
    test(`d${faceCount} keeps every painted mark inside its own facet`, async ({ page }) => {
      const faces = Array.from({ length: faceCount }, (_, i) => i + 1);
      await mountDie(page, { faceCount, faces, dieSize: '240px' });
      const { rows } = await faceContent(page, faceCount);
      for (const row of rows) {
        const marks: { what: string; points: [number, number][] }[] = [];
        // The content square itself, not only the marks in it. Sizing it to
        // the polygon but then CENTRING it on the facet's bounding box --
        // which is not the centroid, for a triangle or a kite -- leaves the
        // marks small enough to survive on their own; the square does not.
        marks.push({
          what: `content square ${row.content.width.toFixed(1)}px at (${row.content.cx.toFixed(1)}, ${row.content.cy.toFixed(1)})`,
          points: [
            [row.content.cx - row.content.width / 2, row.content.cy - row.content.height / 2],
            [row.content.cx + row.content.width / 2, row.content.cy - row.content.height / 2],
            [row.content.cx + row.content.width / 2, row.content.cy + row.content.height / 2],
            [row.content.cx - row.content.width / 2, row.content.cy + row.content.height / 2],
          ],
        });
        for (const pip of row.pips) {
          // Eight points around the dot's rim: a disc, not a point.
          const rim: [number, number][] = [];
          for (let i = 0; i < 8; i++) {
            const angle = (i / 8) * 2 * Math.PI;
            rim.push([pip.cx + Math.cos(angle) * pip.r, pip.cy + Math.sin(angle) * pip.r]);
          }
          marks.push({ what: `pip at (${pip.cx.toFixed(1)}, ${pip.cy.toFixed(1)}) r${pip.r.toFixed(1)}`, points: rim });
        }
        for (const box of [row.text, ...row.corners].filter(Boolean) as any[]) {
          marks.push({
            what: `text "${box.text}" ${box.width.toFixed(1)}x${box.height.toFixed(1)} at (${box.cx.toFixed(1)}, ${box.cy.toFixed(1)})`,
            points: [
              [box.cx - box.width / 2, box.cy - box.height / 2],
              [box.cx + box.width / 2, box.cy - box.height / 2],
              [box.cx + box.width / 2, box.cy + box.height / 2],
              [box.cx - box.width / 2, box.cy + box.height / 2],
            ],
          });
        }
        expect(marks.length, `face ${row.faceValue} paints something`).toBeGreaterThan(0);
        for (const mark of marks) {
          const worst = Math.min(...mark.points.map((point) => insideBy(row.polygon, point)));
          expect(worst,
            `face ${row.faceValue} (${row.width.toFixed(1)}x${row.height.toFixed(1)}px box): ${mark.what}`)
            .toBeGreaterThan(-0.5);
        }
      }
    });
  }

  // Trap 2. A d4 and every odd-sided barrel are read from a face NOBODY CAN
  // SEE -- the one they rest on -- so a die that paints its value only at each
  // face's centre lands showing nothing. `die-faces.ts` reports the rule; a
  // physical d4 answers it by printing the value at the CORNERS of the faces
  // that are visible, and so does this.
  //
  // The check is entirely geometric: for every corner mark, the vertex it sits
  // nearest must be the vertex whose "up" reading is the face whose value the
  // mark carries. Printing a face's own value at its own corners fails it, and
  // so does dropping the corner marks.
  for (const faceCount of [4, 7]) {
    test(`d${faceCount} is read from a hidden face, so it prints values at the corners`, async ({ page }) => {
      const faces = Array.from({ length: faceCount }, (_, i) => (i + 1) * 10);
      await mountDie(page, { faceCount, faces, dieSize: '240px' });
      const result = await faceContent(page, faceCount);
      expect(result.readingRule).not.toBe('up-face');
      for (const row of result.rows) {
        // One mark per vertex of the facet.
        expect(row.corners.length,
          `face ${row.faceValue} corner marks`).toBe(row.polygon.length);
        const matched = row.corners.map((corner: any) => {
          let nearest = -1;
          let best = Infinity;
          row.polygon.forEach((vertex: number[], i: number) => {
            const d = Math.hypot(vertex[0] - corner.cx, vertex[1] - corner.cy);
            if (d < best) { best = d; nearest = i; }
          });
          return { nearest, corner };
        });
        // Distinct vertices: one mark each, not three piled on one corner.
        expect(new Set(matched.map((m: any) => m.nearest)).size).toBe(row.polygon.length);
        for (const { nearest, corner } of matched) {
          const expectedFace = row.vertexFaces[nearest];
          expect(corner.faceIndex,
            `face ${row.faceIndex} corner at vertex ${nearest} must carry the face read from that vertex`)
            .toBe(expectedFace);
          expect(corner.text).toBe(String(faces[expectedFace]));
        }
      }
      // And the consequence that matters: the value of any face turns up on
      // OTHER faces, where a player can see it while the die rests on it.
      for (let index = 0; index < faceCount; index++) {
        const elsewhere = result.rows.filter((row: any) =>
          row.faceIndex !== index && row.corners.some((c: any) => c.faceIndex === index));
        expect(elsewhere.length,
          `face ${index}'s value must be readable from faces other than itself`).toBeGreaterThan(0);
      }
    });
  }

  // ... and a die that IS read from its up face gets no corner clutter.
  test('a d6 is read from its up face, so it prints no corner values', async ({ page }) => {
    await mountDie(page, { faceCount: 6, faces: [1, 2, 3, 4, 5, 6], dieSize: '240px' });
    const result = await faceContent(page, 6);
    expect(result.readingRule).toBe('up-face');
    expect(result.rows.every((row: any) => row.corners.length === 0)).toBe(true);
  });

  // A barrel's cap facets carry no value, so they must carry no content --
  // otherwise a d7 shows fourteen stray numerals on its cones.
  test('cap facets carry no content', async ({ page }) => {
    await mountDie(page, { faceCount: 7, dieSize: '240px' });
    const result = await faceContent(page, 7);
    expect(result.capsWithContent).toBe(0);
  });

  // Trap 3. WHICH WAY UP the content sits on each facet.
  //
  // Everything above measures where marks land and how big they are, in the
  // facet's own (u, v) frame read off the render -- so an in-plane quarter turn
  // of that frame moves the content and the frame together and changes not one
  // number. It is still a proper orthonormal rotation, the boxes still match
  // the geometry's extents, the pips still land on their canonical cells, and
  // every numeral on every die renders ninety degrees off. Nothing in this file
  // saw that until here.
  //
  // The invariant, stated without reference to the component's own routine: a
  // facet's content reads downwards along the SOLID's own down direction,
  // projected into that facet's plane. That is what makes the numerals on a
  // die agree with each other instead of each sitting at its own angle -- and
  // for the presented facet, composed with the pose, it is what the roll in
  // `presentationTransform` then turns to screen-down.
  //
  // Facets whose normal IS the down axis (a d6's top and bottom) have no such
  // direction and are skipped; the count of what was actually checked is
  // asserted so the skip cannot swallow the test.
  for (const [faceCount, minChecked] of [[6, 4], [7, 7], [8, 8], [10, 10], [12, 12], [20, 20], [4, 4]]) {
    test(`d${faceCount} orients every facet's content along the solid's own down axis`, async ({ page }) => {
      await mountDie(page, { faceCount, dieSize: '240px' });
      const rows = await facetAxes(page, faceCount);
      // CSS y points down, so the solid's down axis is +y in the body frame.
      const DOWN = [0, 1, 0];
      const dot = (a: number[], b: number[]) => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
      let checked = 0;
      const angles: string[] = [];
      for (const row of rows) {
        const length = Math.hypot(...row.normal);
        const w = row.normal.map((value: number) => value / length);
        // Down, with its out-of-plane part removed.
        const projected = [0, 1, 2].map((i) => DOWN[i] - w[i] * dot(DOWN, w));
        const scale = Math.hypot(...projected);
        if (scale < 0.05) continue;
        const expected = projected.map((value) => value / scale);
        const cosine = Math.min(1, Math.max(-1, dot(row.v, expected)));
        const degrees = (Math.acos(cosine) * 180) / Math.PI;
        angles.push(`face ${row.faceIndex}: ${degrees.toFixed(2)}`);
        checked++;
        expect(degrees,
          `d${faceCount} content roll away from the solid's down axis, per facet: ${angles.join(', ')}`)
          .toBeLessThan(0.5);
      }
      expect(checked,
        `d${faceCount} facets with a well-defined in-plane down direction`).toBeGreaterThanOrEqual(minChecked);
    });
  }

  // Trap 4. The content region is a SQUARE, sized to the largest square that
  // fits the polygon.
  //
  // Stretching it to the facet's bounding-box aspect instead -- still centred
  // on the centroid, with the glyph and pip sizing left alone -- is caught on a
  // d10, d12, d20 or d4 by the "nothing leaves its own facet" trap above, since
  // a stretched rectangle in a kite or a triangle crosses an edge. It is NOT
  // caught on a d6 or a d7, because a rectangle stretched inside a rectangular
  // facet never crosses anything: a d7 barrel's pip lattice would smear along
  // the barrel by a factor of 2.7 and every assertion in this file would pass.
  // The d7 is the worked example for that defect, and it is safe today only
  // because barrels are corner-printed and so never pipped -- an accident, not
  // a guarantee.
  //
  // Measured against the geometry's own inscribed square, not against the
  // component's `--content-size`.
  for (const faceCount of [6, 7, 8, 10, 12, 20, 4]) {
    test(`d${faceCount} sizes its content to a square inscribed in the facet, not to the facet's box`, async ({ page }) => {
      const faces = Array.from({ length: faceCount }, (_, i) => i + 1);
      await mountDie(page, { faceCount, faces, dieSize: '240px' });
      const { rows } = await faceContent(page, faceCount);
      const ratios: number[] = [];
      for (const row of rows) {
        const { width, height } = row.content;
        // Square to within the rounding a percentage of a non-square box
        // costs (measured worst case 0.13% on a d7).
        expect(Math.abs(width - height),
          `face ${row.faceValue} content region ${width.toFixed(2)}x${height.toFixed(2)}px`
          + ` in a ${row.width.toFixed(1)}x${row.height.toFixed(1)}px facet box`)
          .toBeLessThanOrEqual(0.02 * Math.max(width, height));
        const available = 2 * inscribedSquareHalfSide(row.polygon);
        expect(available, `face ${row.faceValue} has an inscribed square at all`).toBeGreaterThan(0);
        ratios.push(width / available);
      }
      // The one margin the component documents, applied uniformly: air around
      // the marks, and never more room than the polygon actually has. A
      // content region sized to the bounding box puts this past 1 on every
      // facet that is not square.
      const detail = `content side / inscribed square, per facet: ${ratios.map((r) => r.toFixed(3)).join(', ')}`;
      expect(Math.min(...ratios), detail).toBeGreaterThan(0.5);
      expect(Math.max(...ratios), detail).toBeLessThanOrEqual(0.95);
      expect(Math.max(...ratios) - Math.min(...ratios), detail).toBeLessThan(0.01);
    });
  }

  // Trap 5. The marks have to be BIG ENOUGH TO READ at the size a real board
  // draws a die at.
  //
  // Every other assertion here is an upper bound -- stay inside the facet, do
  // not overrun the inscribed square -- and all of them are satisfied perfectly
  // by drawing nothing at all. Shrinking the corner-mark cap by five times
  // leaves the corner values of a d4 and a barrel at about a pixel of dust and
  // leaves every one of those assertions green.
  //
  // 100px is pig's die, so it is the size that actually ships; the die is
  // scale-free (`1em` is `--die-size`), so a floor there is a floor
  // everywhere. 6px is about the smallest a digit can be and still be a digit;
  // a pip only has to be seen rather than read, so it gets 3.5. What binds is
  // the barrel: a d7's side facet is an 18-by-50px strip, and a square
  // inscribed anywhere in it is bounded by the 18.
  const LEGIBLE_DIE_PX = 100;
  const MIN_GLYPH_PX = 6;
  const MIN_PIP_PX = 3.5;
  for (const [label, faceCount, faces] of [
    ['d4 corner numerals', 4, [1, 2, 3, 4]],
    ['d6 pips', 6, [1, 2, 3, 4, 5, 6]],
    ['d7 corner numerals', 7, [1, 2, 3, 4, 5, 6, 7]],
    ['d8 pips', 8, [1, 2, 3, 4, 5, 6, 7, 8]],
    ['d10 pips', 10, [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]],
    ['d10 numerals', 10, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]],
    ['d12 numerals', 12, Array.from({ length: 12 }, (_, i) => i + 1)],
    ['d20 numerals', 20, Array.from({ length: 20 }, (_, i) => i + 1)],
  ] as [string, number, number[]][]) {
    test(`${label} stay legible on a ${LEGIBLE_DIE_PX}px die`, async ({ page }) => {
      await mountDie(page, { faceCount, faces, dieSize: `${LEGIBLE_DIE_PX}px` });
      const marks = await page.evaluate(() => {
        const die = document.getElementById('fixture-die') as any;
        const root = die.shadowRoot as ShadowRoot;
        const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
        const out: { what: string; kind: 'glyph' | 'pip'; px: number }[] = [];
        for (const facet of facets) {
          const faceValue = facet.dataset.faceValue;
          if (faceValue === undefined) continue;
          for (const span of Array.from(facet.querySelectorAll('.content > span')) as HTMLElement[]) {
            if (!span.textContent) continue;
            out.push({ what: `face ${faceValue} centre "${span.textContent}"`, kind: 'glyph',
              px: parseFloat(getComputedStyle(span).fontSize) });
          }
          for (const span of Array.from(facet.querySelectorAll('.corner > span')) as HTMLElement[]) {
            out.push({ what: `face ${faceValue} corner "${span.textContent}"`, kind: 'glyph',
              px: parseFloat(getComputedStyle(span).fontSize) });
          }
          for (const pip of Array.from(facet.querySelectorAll('.pip')) as HTMLElement[]) {
            out.push({ what: `face ${faceValue} pip`, kind: 'pip',
              px: parseFloat(getComputedStyle(pip).width) });
          }
        }
        return out;
      });
      expect(marks.length, `${label} draws something`).toBeGreaterThan(0);
      const smallest = (kind: string) => marks
        .filter((mark) => mark.kind === kind)
        .reduce((best, mark) => (best === null || mark.px < best.px ? mark : best), null as any);
      const glyph = smallest('glyph');
      const pip = smallest('pip');
      if (glyph) {
        expect(glyph.px, `smallest glyph -- ${glyph.what}`).toBeGreaterThanOrEqual(MIN_GLYPH_PX);
      }
      if (pip) {
        expect(pip.px, `smallest pip -- ${pip.what}`).toBeGreaterThanOrEqual(MIN_PIP_PX);
      }
    });
  }
});
