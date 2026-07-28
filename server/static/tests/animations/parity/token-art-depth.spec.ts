import { test, expect, type Page } from '@playwright/test';

/**
 * DEPTH FOR THE TWO SHAPES THAT CANNOT BE SOLIDS.
 *
 * `meeple` and `pawn` keep their authored SVG, because a prism over a
 * non-convex outline paints its own back surface through its front and CSS has
 * no z-buffer to stop it (measured: 2.8% of a meeple's silhouette wrong at 75
 * degrees, a comb-shaped control at 12.1%, and explicit `z-index` sorting
 * measured WORSE than doing nothing). So they are presented rather than
 * modelled — and the bar for the presentation is that a board mixing a meeple
 * with a cube reads as one scene rather than two art styles.
 *
 * "Reads as one scene" is a judgment, but it is not only a judgment, and this
 * spec pins the part that is measurable.
 *
 * 1. THE LIGHT COMES FROM THE SAME PLACE. Both assets are drawn lit from the
 *    upper RIGHT and every solid is lit from the upper LEFT. Sampled at 200px
 *    on white, mean luma of the left half against the right half of the drawn
 *    silhouette's top band: a meeple was 8.2 units the WRONG way and a pawn
 *    17.1. Both are mirror-symmetric in silhouette, so `scaleX(-1)` moves the
 *    light across without touching the shape, and the same measurement now
 *    reads +8.2 and +17.1. This is measured off PIXELS on purpose: a CSS
 *    assertion would pass on a mirror applied to the wrong element.
 *
 * 2. EVERY SHADOW POINTS THE WAY THE LIGHT SAYS. The edge shadow and the
 *    contact shadow are both offset along `SHADOW_DIRECTION`, which is derived
 *    from the same `LIGHT` the solids are shaded by. Asserted against the
 *    module's own numbers, so moving the light and not the shadows fails here.
 *
 * 3. NOTHING IT ADDS IS 3D. A `rotateX` would have been the obvious way to
 *    tilt a standing piece, and a 3D transform is a composited layer per
 *    element the moment an ancestor animates — the exact cost the flattening
 *    of the solids exists to avoid. The tilt is a 2D `scaleY`, and 55 animating
 *    meeples promote no more than the flat art did.
 *
 * 4. IT IS A PURE FUNCTION OF STATE. Stacks pool and reparent hosts, so a node
 *    that was a cube and becomes a meeple must be a meeple, and one that was a
 *    meeple and becomes a disc must lose the whole treatment.
 */

const ART_SHAPES = ['meeple', 'pawn'];
const SOLID_SHAPES = ['cube', 'token', 'chip', 'disc'];

/** Mean luma of the left half against the right half of the drawn top band. */
async function lightAsymmetry(page: Page, shape: string): Promise<number> {
  await page.evaluate(async (type: string) => {
    await import('/src/components/boardgame-token.ts');
    document.body.innerHTML = '';
    document.body.style.cssText = 'margin:0;background:#fff';
    const el = document.createElement('boardgame-token') as any;
    el.type = type;
    el.item = { ID: type };
    // The elevation drop-shadow is not part of the piece and would tint the
    // background around it; the measurement is of the art itself.
    el.noShadow = true;
    el.style.cssText = 'position:fixed;left:0;top:0;display:block';
    el.style.setProperty('--component-width', '200px');
    document.body.appendChild(el);
    await el.updateComplete;
    await el.updateComplete;
    await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
  }, shape);

  const shot = (await page.screenshot({
    clip: { x: 0, y: 0, width: 220, height: 220 },
  })).toString('base64');

  return page.evaluate(async (base64: string) => {
    const image = new Image();
    image.src = 'data:image/png;base64,' + base64;
    await image.decode();
    const canvas = document.createElement('canvas');
    canvas.width = image.width;
    canvas.height = image.height;
    const context = canvas.getContext('2d', { willReadFrequently: true })!;
    context.drawImage(image, 0, 0);
    const data = context.getImageData(0, 0, canvas.width, canvas.height).data;
    const pixels: { x: number; y: number; luma: number }[] = [];
    let minX = Infinity; let maxX = -Infinity; let minY = Infinity; let maxY = -Infinity;
    for (let y = 0; y < canvas.height; y++) {
      for (let x = 0; x < canvas.width; x++) {
        const p = (y * canvas.width + x) * 4;
        // Decisively red pixels only: the piece, never a grey contact shadow.
        if (!(data[p] > 90 && data[p] - data[p + 1] > 50 && data[p] - data[p + 2] > 50)) continue;
        pixels.push({ x, y, luma: 0.213 * data[p] + 0.715 * data[p + 1] + 0.072 * data[p + 2] });
        minX = Math.min(minX, x); maxX = Math.max(maxX, x);
        minY = Math.min(minY, y); maxY = Math.max(maxY, y);
      }
    }
    if (pixels.length < 1000) return NaN;
    const centreX = (minX + maxX) / 2;
    // The TOP band, because that is where a piece lit from above shows the
    // faces whose brightness differs left to right; the bottom of a leaned
    // solid is nearly all rim.
    const band = pixels.filter((q) => q.y <= (minY + maxY) / 2);
    const mean = (side: typeof band) => side.reduce((sum, q) => sum + q.luma, 0) / side.length;
    return mean(band.filter((q) => q.x < centreX)) - mean(band.filter((q) => q.x >= centreX));
  }, shot);
}

test.describe('meeple and pawn, given depth without a mesh', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('are lit from the same side as the solids', async ({ page }) => {
    const measured: Record<string, number> = {};
    for (const shape of [...SOLID_SHAPES, ...ART_SHAPES]) {
      measured[shape] = await lightAsymmetry(page, shape);
    }
    for (const shape of [...SOLID_SHAPES, ...ART_SHAPES]) {
      expect(Number.isNaN(measured[shape]), `${shape} drew nothing to measure`).toBe(false);
    }
    // The art shapes must be DECISIVELY left-lit. Unmirrored they measure
    // -8.2 (meeple) and -17.1 (pawn), so this is a sign flip, not a nudge.
    for (const shape of ART_SHAPES) {
      expect(measured[shape], `${shape} must be lit from the left, like every solid`)
        .toBeGreaterThan(4);
    }
    // And no shape may be lit from the RIGHT. The solids are nearly symmetric
    // by construction -- a 12-gon rim under a mostly-overhead light measures
    // within 0.1 -- so this is a direction check, not a magnitude one.
    for (const shape of SOLID_SHAPES) {
      expect(measured[shape], `${shape} must not be lit from the right`)
        .toBeGreaterThan(-0.5);
    }
  });

  test('offset every shadow along the light\'s own direction', async ({ page }) => {
    const result = await page.evaluate(async (shapes: string[]) => {
      const solid = await import('/src/components/token-solid.ts');
      await import('/src/components/boardgame-token.ts');
      document.body.innerHTML = '';
      const out: Record<string, any> = {};
      const SIZE = 200;
      for (const type of shapes) {
        const el = document.createElement('boardgame-token') as any;
        el.type = type;
        el.item = { ID: type };
        el.style.cssText = 'display:block';
        el.style.setProperty('--component-width', `${SIZE}px`);
        document.body.appendChild(el);
        await el.updateComplete;
        await el.updateComplete;
        const art = el.renderRoot.querySelector('#art');
        const inner = el.renderRoot.querySelector('#inner');
        out[type] = {
          artFilter: getComputedStyle(art).filter,
          artTransform: getComputedStyle(art).transform,
          imgTransform: getComputedStyle(art.querySelector('img')).transform,
          groundTransform: getComputedStyle(inner, '::after').transform,
          groundWidth: getComputedStyle(inner, '::after').width,
          groundBackground: getComputedStyle(inner, '::after').backgroundImage,
        };
        el.remove();
      }
      return {
        out,
        size: SIZE,
        direction: [...solid.SHADOW_DIRECTION],
        depth: { ...solid.ART_DEPTH },
        drawnWidth: Object.fromEntries(shapes.map((s) => [s, solid.artDrawnWidth(s)])),
      };
    }, ART_SHAPES);

    const [dirX, dirY] = result.direction;
    // The light is above and to the left, so a shadow falls down and to the
    // right. Anything else here and the rest of the spec is measuring nothing.
    expect(dirX, 'shadows fall to the right').toBeGreaterThan(0);
    expect(dirY, 'and downwards, further than they fall sideways')
      .toBeGreaterThan(dirX);

    for (const shape of ART_SHAPES) {
      const shot = result.out[shape];

      // The edge: a hard-edged drop-shadow hugging the silhouette.
      const edge = /drop-shadow\(rgba?\([^)]*\)\s+(-?[\d.]+)px\s+(-?[\d.]+)px\s+(-?[\d.]+)px\)/
        .exec(shot.artFilter)
        ?? /drop-shadow\((-?[\d.]+)px\s+(-?[\d.]+)px\s+(-?[\d.]+)px/.exec(shot.artFilter);
      expect(edge, `${shape}: no edge drop-shadow in "${shot.artFilter}"`).not.toBeNull();
      const edgeX = Number(edge![1]);
      const edgeY = Number(edge![2]);
      expect(edgeX / result.size, `${shape}: the edge offset is derived from the light`)
        .toBeCloseTo(dirX * result.depth.edgeEm, 4);
      expect(edgeY / result.size).toBeCloseTo(dirY * result.depth.edgeEm, 4);
      expect(Number(edge![3]), 'the edge is hard, not a blur').toBeCloseTo(0, 4);

      // The tilt: a 2D scale about the foot. matrix(), never matrix3d().
      expect(shot.artTransform, `${shape}: the tilt must not be a 3D transform`)
        .toMatch(/^matrix\(/);
      const tilt = shot.artTransform.slice('matrix('.length, -1).split(',').map(Number);
      expect(tilt[0], `${shape}: the tilt does not touch the width`).toBeCloseTo(1, 6);
      expect(tilt[3], `${shape}: the tilt is ART_DEPTH.lean`)
        .toBeCloseTo(result.depth.lean, 6);

      // The mirror.
      expect(shot.imgTransform, `${shape}: the art is mirrored`)
        .toBe('matrix(-1, 0, 0, 1, 0, 0)');

      // The contact shadow: sized off the PIECE's drawn width, not the box,
      // and translated along the same direction.
      const expectedWidth = result.drawnWidth[shape] * result.depth.groundWidth * result.size;
      expect(Number.parseFloat(shot.groundWidth),
        `${shape}: the contact shadow is as wide as the piece, not as the box`)
        .toBeCloseTo(expectedWidth, 1);
      expect(shot.groundBackground, `${shape}: the contact shadow is a soft ellipse`)
        .toContain('radial-gradient');
      const ground = shot.groundTransform.slice('matrix('.length, -1).split(',').map(Number);
      expect(ground[4], `${shape}: the contact shadow is centred, then offset by the light`)
        .toBeCloseTo(-expectedWidth / 2 + dirX * result.depth.groundEm * result.size, 1);
      expect(ground[5]).toBeCloseTo(dirY * result.depth.groundEm * result.size, 1);
    }
  });

  test('give a solid neither an #art wrapper nor a contact shadow', async ({ page }) => {
    const seen = await page.evaluate(async (shapes: string[]) => {
      await import('/src/components/boardgame-token.ts');
      document.body.innerHTML = '';
      const out: Record<string, any> = {};
      for (const type of shapes) {
        const el = document.createElement('boardgame-token') as any;
        el.type = type;
        el.item = { ID: type };
        document.body.appendChild(el);
        await el.updateComplete;
        await el.updateComplete;
        out[type] = {
          art: el.renderRoot.querySelectorAll('#art').length,
          // A pseudo-element that no rule matches computes `content: none`.
          groundContent: getComputedStyle(el.renderRoot.querySelector('#inner'), '::after').content,
        };
        el.remove();
      }
      return out;
    }, SOLID_SHAPES);

    for (const shape of SOLID_SHAPES) {
      expect(seen[shape].art, `${shape} is a solid; it has no authored art to wrap`).toBe(0);
      expect(seen[shape].groundContent,
        `${shape} lies on the board and meets it along its own dark rim`).toBe('none');
    }
  });

  /**
   * The same hazard `token-3d.spec.ts` opens with, on the other path. A stack
   * pools hosts, and the treatment here is CSS keyed off a class that
   * `_computeClasses` recomputes from `this.type` on every render -- so it is
   * structurally safe. That is exactly the kind of claim that stops being true
   * when someone writes one style imperatively, so it is checked BOTH ways:
   * a cube that becomes a meeple gains the treatment, and a meeple that becomes
   * a disc loses all of it.
   */
  test('follow a recycled host in both directions', async ({ page }) => {
    const result = await page.evaluate(async () => {
      const view = await import('/src/components/component-view.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const stack = document.createElement('boardgame-component-stack') as any;
      document.body.appendChild(stack);
      stack.componentView = view.tokenView({
        properties: ({ kind, component }: any) => kind === 'visible'
          ? { type: component.Values.Type, color: 'red' }
          : { type: 'token', color: 'red' },
      });
      const component = (id: string, Type: string) => ({
        ID: id, Index: 0, Deck: 'd', GameName: 'g',
        Values: { Type }, DynamicValues: {},
      });
      const setMembership = async (components: unknown[]) => {
        stack.stack = {
          Deck: 'd',
          Size: components.length,
          Components: components,
          IDs: components.map((c: any) => c?.ID ?? ''),
        };
        await stack.updateComplete;
        await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
        for (const child of [...stack.children]) await (child as any).updateComplete;
        for (const child of [...stack.children]) await (child as any).updateComplete;
      };
      const snapshot = (el: any) => {
        const inner = el.renderRoot.querySelector('#inner');
        const art = el.renderRoot.querySelector('#art');
        return {
          type: el.type,
          art: !!art,
          facets: el.renderRoot.querySelectorAll('.facet').length,
          imgTransform: art ? getComputedStyle(art.querySelector('img')).transform : null,
          artTransform: art ? getComputedStyle(art).transform : null,
          groundContent: getComputedStyle(inner, '::after').content,
        };
      };

      await setMembership([component('a', 'cube'), component('b', 'cube')]);
      const host = [...stack.querySelectorAll('boardgame-token')][1] as any;
      const asCube = snapshot(host);

      await setMembership([component('a', 'cube')]);
      await setMembership([component('a', 'cube'), component('c', 'meeple')]);
      const hosts = [...stack.querySelectorAll('boardgame-token')];
      const asMeeple = { recycled: hosts[1] === host, ...snapshot(host) };

      await setMembership([component('a', 'cube')]);
      await setMembership([component('a', 'cube'), component('d', 'disc')]);
      const asDisc = {
        recycled: [...stack.querySelectorAll('boardgame-token')][1] === host,
        ...snapshot(host),
      };

      stack.remove();
      return { asCube, asMeeple, asDisc };
    });

    expect(result.asCube.art, 'a cube is a solid').toBe(false);
    expect(result.asCube.groundContent).toBe('none');

    expect(result.asMeeple.recycled, 'the stack must have reused the pooled host').toBe(true);
    expect(result.asMeeple.type).toBe('meeple');
    expect(result.asMeeple.art, 'a recycled host that became a meeple has the art wrapper')
      .toBe(true);
    expect(result.asMeeple.imgTransform, 'and it is mirrored')
      .toBe('matrix(-1, 0, 0, 1, 0, 0)');
    expect(result.asMeeple.artTransform, 'and leaned').toMatch(/^matrix\(/);
    expect(result.asMeeple.groundContent, 'and stands on a contact shadow')
      .not.toBe('none');
    expect(result.asMeeple.facets, 'and builds no mesh').toBe(0);

    expect(result.asDisc.recycled, 'the same host again').toBe(true);
    expect(result.asDisc.type).toBe('disc');
    expect(result.asDisc.art, 'a host that stopped being a meeple loses the wrapper')
      .toBe(false);
    expect(result.asDisc.groundContent, 'and the contact shadow with it').toBe('none');
    expect(result.asDisc.facets, 'and is a solid again').toBe(6);
  });

  /**
   * The tripwire the solids' flattening exists for, pointed at the other path.
   * A `rotateX` tilt or a `perspective` here would put every one of these back
   * inside a live 3D context, and Chromium promotes every element in one to its
   * own composited layer the instant an ancestor transform animates -- which is
   * what a stack's FLIP does on every move.
   */
  test('promote no layers while an ancestor transform animates', async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== 'chromium', 'CDP LayerTree is Chromium-only');

    const TOKENS = 55;
    await page.evaluate(async (count) => {
      await import('/src/components/boardgame-token.ts');
      document.body.style.cssText = 'margin:0;background:#eee;height:100vh;overflow:hidden';
      document.body.innerHTML = '';
      const tokens: any[] = [];
      for (let i = 0; i < count; i++) {
        const el = document.createElement('boardgame-token') as any;
        el.type = i % 2 ? 'meeple' : 'pawn';
        el.color = ['red', 'blue', 'green', 'yellow'][i % 4];
        el.item = { ID: `t${i}` };
        el.style.cssText = `position:absolute;left:${(i % 10) * 95 + 10}px;`
          + `top:${Math.floor(i / 10) * 95 + 10}px;--component-width:60px`;
        document.body.appendChild(el);
        tokens.push(el);
      }
      await Promise.all(tokens.map((t) => t.updateComplete));
      await Promise.all(tokens.map((t) => t.updateComplete));
      (window as any).__tokens = tokens;
      (window as any).__drive = () => tokens.map((t) => t.animate(
        [{ transform: 'translate(0px,0px)' }, { transform: 'translate(8px,5px)' }],
        { duration: 900, iterations: Infinity, direction: 'alternate', composite: 'add' },
      ));
    }, TOKENS);

    const wrappers = await page.evaluate(() => (window as any).__tokens
      .reduce((n: number, t: any) => n + t.renderRoot.querySelectorAll('#art').length, 0));
    expect(wrappers, 'the scene under test must actually be 55 art tokens').toBe(TOKENS);

    const client = await page.context().newCDPSession(page);
    const snapshots: any[][] = [];
    client.on('LayerTree.layerTreeDidChange', (event: any) => {
      if (event.layers) snapshots.push(event.layers);
    });
    await client.send('LayerTree.enable');
    await page.waitForTimeout(500);
    await page.evaluate(() => (window as any).__drive());
    await page.waitForTimeout(1200);
    await client.send('LayerTree.disable');
    await client.detach();

    const layers = snapshots[snapshots.length - 1] ?? [];
    const painted = layers.filter((layer) => layer.drawsContent);
    const megapixels = painted.reduce((sum, l) => sum + l.width * l.height, 0) / 1e6;

    expect(layers.length, `the layer tree did not arrive (${snapshots.length} snapshots)`)
      .toBeGreaterThan(0);
    expect(painted.length, '55 animating art tokens promote about one layer each')
      .toBeLessThan(TOKENS * 2);
    expect(megapixels, 'and those layers stay token-sized').toBeLessThan(10);

    const styles = await page.evaluate(() => {
      const el = (window as any).__tokens[0];
      const art = el.renderRoot.querySelector('#art');
      return {
        outerPerspective: getComputedStyle(el.renderRoot.querySelector('#outer')).perspective,
        innerStyle: getComputedStyle(el.renderRoot.querySelector('#inner')).transformStyle,
        artStyle: getComputedStyle(art).transformStyle,
        artWillChange: getComputedStyle(art).willChange,
        innerTransform: getComputedStyle(el.renderRoot.querySelector('#inner')).transform,
      };
    });
    expect(styles.outerPerspective, 'no camera in CSS').toBe('none');
    expect(styles.innerStyle, '#inner must not open a 3D sorting context').toBe('flat');
    expect(styles.artStyle, 'nor may #art').toBe('flat');
    expect(styles.artWillChange, 'and nothing asks to be promoted').toBe('auto');
    // #inner's transform belongs to the animation kernel; the treatment lives
    // on #art and on the img, never here.
    expect(styles.innerTransform).toBe('none');
  });
});
