import { test, expect, type Page } from '@playwright/test';

/**
 * THE SAME PICTURE, DRAWN WITHOUT A 3D CONTEXT.
 *
 * A token used to be a live `preserve-3d` scene: a `perspective` on #outer, a
 * `matrix3d` pose on #solid, and one 3D-transformed, `backface-visibility:
 * hidden` element per facet. It is now flat, already-projected polygons —
 * `src/solid/flat-facets.ts` carries the frame-rate measurements that forced
 * that, and `token-3d.spec.ts` is the regression net for them.
 *
 * This file answers the only other question that matters: IS IT THE SAME
 * PICTURE? Not "does it look about right", and not "does the arithmetic agree
 * with itself" (`token-solid.test.ts` already covers that, and would agree with
 * itself just as happily if the projection were wrong). Both renderings are put
 * on screen, at 400px, and differenced.
 *
 * The reference is built here, in the test, out of the SAME parts the component
 * used to use — `tokenSurface`, `restingPose`, `fitScale` and
 * `facet-placement.ts`'s `solidFacets`, with the fills read straight out of
 * `tokenSolid()` so the only thing that can differ is the GEOMETRY. That is
 * deliberate: the question is whether the perspective divide done once in
 * JavaScript lands where the browser's own would have.
 *
 * ## Why the difference is not required to be zero
 *
 * It is required to be a one-pixel band and nothing else. A clipped, 3D-
 * transformed element and a clip-path'ed flat one anti-alias their edges
 * differently by a fraction of a pixel, everywhere two facets meet and around
 * the silhouette — an unavoidable, invisible difference that no amount of
 * correct arithmetic removes. So the wrong-pixel mask is ERODED: a wrong pixel
 * survives only if all EIGHT of its neighbours are wrong too. An AA seam erodes
 * to nothing, including the diagonal wisps where three seams meet at a vertex
 * (a four-neighbour erosion left two such pixels on a cube, measured). A facet drawn in the wrong place, at the wrong size, or not
 * culled when it should have been, leaves a region several pixels thick, and
 * that must be zero.
 *
 * The positive control at the bottom is what makes that threshold mean
 * anything: it perturbs the projection by one part in fifty and requires the
 * erosion to survive it.
 */

const SIZE = 400;
const ORIGIN = 40;
const SHAPES = ['cube', 'token', 'chip', 'disc'] as const;

/**
 * Renders one token both ways and returns the two screenshots.
 *
 * `perturb` scales the projected polygons about the solid's centre, and exists
 * only for the positive control.
 */
async function renderBoth(
  page: Page,
  shape: string,
  perturb = 1,
): Promise<{ flat: string; reference: string; facets: number }> {
  const facets = await page.evaluate(async ({ shape: type, size, origin, perturb: scale }) => {
    const tokenSolid = await import('/src/components/token-solid.ts') as any;
    const placement = await import('/src/solid/facet-placement.ts') as any;
    const flat = await import('/src/solid/flat-facets.ts') as any;

    document.body.style.cssText = 'margin:0;background:#fff';
    document.body.innerHTML = '';

    const solid = tokenSolid.tokenSolid(type, 'red');
    const surface = tokenSolid.tokenSurface(type);
    const pose = tokenSolid.restingPose(type);
    // The fills come from the component's own output, so a difference here can
    // only ever be geometry.
    const fill = new Map<number, string>(solid.facets.map((facet: any) =>
      [facet.key, /background:(rgb\([^)]*\))/.exec(facet.style)![1]]));
    const anyFill = fill.values().next().value as string;

    const box = (left: number) => {
      const el = document.createElement('div');
      el.style.cssText = `position:fixed;left:${left}px;top:${origin}px;`
        + `width:${size}px;height:${size}px`;
      document.body.appendChild(el);
      return el;
    };

    // 1. THE COMPONENT'S OWN RENDERING, rebuilt from its own output so the
    //    screenshot is not at the mercy of the host element's shadow, elevation
    //    filter or hover transition.
    const flatHost = box(origin);
    const flatSolid = document.createElement('div');
    flatSolid.style.cssText =
      `position:relative;width:100%;height:100%;font-size:${size * solid.fit}px`;
    for (const facet of solid.facets) {
      const el = document.createElement('div');
      el.style.cssText = facet.style;
      if (scale !== 1) {
        // Re-emit the polygon, scaled about the centre.
        const points = tokenSolid.visibleFacetPolygons(type, solid.fit)
          .find((p: any) => p.key === facet.key)!.points
          .map((p: any) => ({ x: p.x * scale, y: p.y * scale }));
        el.style.cssText = `${flat.flatFacetStyle(points)};background:${fill.get(facet.key)}`;
      }
      flatSolid.appendChild(el);
    }
    flatHost.appendChild(flatSolid);

    // 2. THE 3D RENDERING IT REPLACED, exactly: camera on the outer box,
    //    preserve-3d carrier, matrix3d pose, one placed facet per polygon,
    //    culled by backface-visibility.
    const refOuter = box(origin + size + origin);
    refOuter.style.perspective = `${size * tokenSolid.CAMERA_DEPTH_WIDTHS}px`;
    const refInner = document.createElement('div');
    refInner.style.cssText = 'width:100%;height:100%;transform-style:preserve-3d';
    const refSolid = document.createElement('div');
    const columns = [0, 1, 2]
      .map((col) => [pose[col], pose[3 + col], pose[6 + col], 0].join(','))
      .join(',');
    refSolid.style.cssText = `position:relative;width:100%;height:100%;`
      + `font-size:${size * solid.fit}px;transform-style:preserve-3d;`
      + `transform:matrix3d(${columns},0,0,0,1)`;
    const placements = placement.solidFacets(surface);
    for (const p of placements) {
      const el = document.createElement('div');
      el.style.cssText = `position:absolute;left:50%;top:50%;backface-visibility:hidden;`
        + `${p.style};background:${fill.get(p.key) ?? anyFill}`;
      refSolid.appendChild(el);
    }
    refInner.appendChild(refSolid);
    refOuter.appendChild(refInner);

    await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
    return solid.facets.length;
  }, { shape, size: SIZE, origin: ORIGIN, perturb });

  const clip = (x: number) => page.screenshot({
    clip: { x, y: ORIGIN, width: SIZE, height: SIZE },
  }).then((buffer) => buffer.toString('base64'));

  return {
    flat: await clip(ORIGIN),
    reference: await clip(ORIGIN + SIZE + ORIGIN),
    facets,
  };
}

/**
 * The two images, differenced and then eroded.
 *
 * `ink` is how much of the reference is not background, so a comparison of two
 * blank screenshots cannot pass by being empty.
 */
async function compare(page: Page, flat: string, reference: string) {
  return page.evaluate(async ({ a, b }) => {
    const load = async (data: string) => {
      const image = new Image();
      image.src = 'data:image/png;base64,' + data;
      await image.decode();
      const canvas = document.createElement('canvas');
      canvas.width = image.width;
      canvas.height = image.height;
      const context = canvas.getContext('2d', { willReadFrequently: true })!;
      context.drawImage(image, 0, 0);
      return context.getImageData(0, 0, canvas.width, canvas.height);
    };
    const A = await load(a);
    const B = await load(b);
    const { width, height } = A;
    const wrong = new Uint8Array(width * height);
    let ink = 0;
    let raw = 0;
    for (let i = 0; i < width * height; i++) {
      const p = i * 4;
      if (B.data[p] !== 255 || B.data[p + 1] !== 255 || B.data[p + 2] !== 255) ink++;
      const distance = Math.max(
        Math.abs(A.data[p] - B.data[p]),
        Math.abs(A.data[p + 1] - B.data[p + 1]),
        Math.abs(A.data[p + 2] - B.data[p + 2]),
      );
      // 8 of 255 is well under a shade step; anything real here is a whole
      // facet's fill against another's, which is tens of levels apart.
      if (distance > 8) { wrong[i] = 1; raw++; }
    }
    // Erode: a wrong pixel survives only if ALL EIGHT of its neighbours are
    // wrong too. One pass removes a one-pixel band and the little diagonal
    // wisps where three seams meet at a vertex; a misplaced facet is far
    // thicker than that -- the positive control below draws 4px of error and
    // leaves thousands.
    let thick = 0;
    for (let y = 1; y < height - 1; y++) {
      for (let x = 1; x < width - 1; x++) {
        const i = y * width + x;
        if (!wrong[i]) continue;
        let all = true;
        for (let dy = -1; dy <= 1 && all; dy++) {
          for (let dx = -1; dx <= 1; dx++) {
            if (!wrong[i + dy * width + dx]) { all = false; break; }
          }
        }
        if (all) thick++;
      }
    }
    return { ink, raw, thick, pixels: width * height };
  }, { a: flat, b: reference });
}

test.describe('a flattened token draws the picture the 3D one did', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  for (const shape of SHAPES) {
    test(`${shape}: every difference is a one-pixel seam`, async ({ page }) => {
      const { flat, reference, facets } = await renderBoth(page, shape);
      const result = await compare(page, flat, reference);

      expect(facets, `${shape} drew no facets`).toBeGreaterThan(0);
      // Both pictures have to BE pictures. Without this the comparison passes
      // on two blank screenshots.
      expect(result.ink, `${shape}: the 3D reference drew almost nothing`)
        .toBeGreaterThan(SIZE * SIZE * 0.3);
      // ...and the seams have to be seams: a few thousand edge pixels out of
      // 160,000 is the anti-aliasing band, not a redrawn solid.
      expect(result.raw, `${shape}: the difference is not a seam, it is the shape`)
        .toBeLessThan(result.ink * 0.06);
      // THE ASSERTION.
      expect(result.thick,
        `${shape}: ${result.thick} pixels differ by more than anti-aliasing`)
        .toBe(0);
    });
  }

  /**
   * THE POSITIVE CONTROL. Without it the four tests above are unfalsifiable:
   * they would pass just as well if the erosion ate everything, if both
   * renderings came from the same code path, or if the screenshots were of the
   * same box twice.
   *
   * The projection is scaled by 2% — about 4px on a 400px token, which is a
   * difference a person would struggle to see in isolation and which every
   * facet shares, so it is the smallest honest perturbation of the thing under
   * test. The erosion must survive it on every shape.
   */
  test('a 2% error in the projection is caught on every shape', async ({ page }) => {
    for (const shape of SHAPES) {
      const { flat, reference } = await renderBoth(page, shape, 1.02);
      const result = await compare(page, flat, reference);
      expect(result.thick, `${shape}: a 2% projection error survived the erosion undetected`)
        .toBeGreaterThan(100);
    }
  });
});
