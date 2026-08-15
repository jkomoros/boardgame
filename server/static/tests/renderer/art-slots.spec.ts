import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

/**
 * Art on the token, the board and the table.
 *
 * The load-bearing case is the token's, because the token already had a
 * recolouring pipeline: nine `#outer.<colour> img { filter: … }` rules
 * generated from `TOKEN_COLOR_FILTERS`, every one of them calibrated against
 * the one red-family SVG the shipped assets are drawn in. Art that composes
 * with that is a feature; art that is silently run through it is a bug that
 * looks like a design decision, so both directions are pinned here.
 */

async function setup(page: import('@playwright/test').Page) {
  const diagnostics = await prepareRendererFixturePage(page);
  await page.evaluate(async () => {
    await import('/src/client.ts');
    await customElements.whenDefined('boardgame-token');
    await customElements.whenDefined('boardgame-game-surface');
  });
  return diagnostics;
}

const ART = '/src/assets/token_disc.svg';

test('a game\'s own token art is drawn, and is NOT run through the colour filters', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async (art) => {
      const make = async (options: { art?: string; recolor?: boolean }) => {
        const token = document.createElement('boardgame-token');
        token.type = 'chip';
        token.color = 'green';
        if (options.art) token.art = art;
        if (options.recolor) token.recolorArt = true;
        document.body.append(token);
        await token.updateComplete;
        const root = token.shadowRoot!;
        const img = root.querySelector('#art img');
        return {
          src: img?.getAttribute('src') ?? null,
          filter: img ? getComputedStyle(img).filter : null,
          facets: root.querySelectorAll('.facet').length,
        };
      };
      return {
        shipped: await make({}),
        custom: await make({ art }),
        recoloured: await make({ art, recolor: true }),
      };
    }, ART);

    // A `chip` is a SOLID by default: no img at all, facets instead.
    expect(result.shipped.src).toBe(null);
    expect(result.shipped.facets).toBeGreaterThan(0);

    // Art wins over the generated solid -- otherwise `art` on a chip would
    // silently do nothing.
    expect(result.custom.src).toBe(ART);
    expect(result.custom.facets).toBe(0);
    // 'green' has a filter in TOKEN_COLOR_FILTERS. It must not reach this img.
    expect(result.custom.filter).toBe('none');

    // …unless the author says their art is in the red family the table expects.
    expect(result.recoloured.src).toBe(ART);
    expect(result.recoloured.filter).toContain('hue-rotate');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a custom pawn keeps the standing-piece depth but loses the shipped assets\' mirror', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async (art) => {
      const make = async (custom: boolean) => {
        const token = document.createElement('boardgame-token');
        token.type = 'pawn';
        token.color = 'red';
        if (custom) token.art = art;
        document.body.append(token);
        await token.updateComplete;
        const root = token.shadowRoot!;
        const wrapper = root.querySelector('#art')!;
        const img = root.querySelector('#art img')!;
        return {
          // The lean and the edge shadow follow from the TYPE meaning "this
          // stands up", so they survive somebody else's art.
          wrapperTransform: getComputedStyle(wrapper).transform,
          wrapperFilter: getComputedStyle(wrapper).filter,
          // The mirror is a correction for two specific FILES and must not.
          imgTransform: getComputedStyle(img).transform,
          objectFit: getComputedStyle(img).objectFit,
        };
      };
      return { shipped: await make(false), custom: await make(true) };
    }, ART);

    // matrix(1, 0, 0, lean, 0, 0) on both -- the lean is unchanged.
    expect(result.shipped.wrapperTransform).toBe(result.custom.wrapperTransform);
    expect(result.custom.wrapperTransform).not.toBe('none');
    expect(result.custom.wrapperFilter).toContain('drop-shadow');

    // matrix(-1, …) for the shipped art, identity for somebody else's.
    expect(result.shipped.imgTransform).toContain('-1');
    expect(result.custom.imgTransform).toBe('matrix(1, 0, 0, 1, 0, 0)');
    // A PNG in a square box is stretched by the default `fill`; an SVG is not.
    expect(result.custom.objectFit).toBe('contain');
    expect(result.shipped.objectFit).toBe('fill');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a token still holds its square layout box when it is wearing art', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async (art) => {
      const token = document.createElement('boardgame-token');
      token.type = 'chip';
      token.art = art;
      token.style.setProperty('--component-width', '64px');
      document.body.append(token);
      await token.updateComplete;
      const inner = token.shadowRoot!.querySelector('#inner')!.getBoundingClientRect();
      return { width: Math.round(inner.width), height: Math.round(inner.height) };
    }, ART);

    // The box is the stack's layout contract; art may not renegotiate it.
    expect(result).toEqual({ width: 64, height: 64 });

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('the game surface becomes a table when it is given a mat, and only then', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async (art) => {
      const plain = document.createElement('boardgame-game-surface');
      plain.heading = 'Plain';
      document.body.append(plain);
      await plain.updateComplete;

      const matted = document.createElement('boardgame-game-surface');
      matted.heading = 'Darwin';
      matted.art = art;
      matted.artWash = '#f4eddde8';
      document.body.append(matted);
      await matted.updateComplete;

      const read = (surface: Element) => {
        const style = getComputedStyle(surface);
        return {
          image: style.backgroundImage,
          size: style.backgroundSize,
          repeat: style.backgroundRepeat,
          radius: style.borderTopLeftRadius,
          borderWidth: style.borderTopWidth,
        };
      };
      return {
        plain: read(plain.shadowRoot!.querySelector('#surface')!),
        matted: read(matted.shadowRoot!.querySelector('#surface')!),
      };
    }, ART);

    // A surface with no art is exactly what it was before this existed.
    expect(result.plain.image).toBe('none');
    expect(result.plain.borderWidth).toBe('0px');

    // The wash is the FIRST layer, so it is painted over the photograph. The
    // other order leaves it invisible underneath, which is the bug that makes
    // text on a mat unreadable.
    expect(result.matted.image.indexOf('linear-gradient')).toBeLessThan(
      result.matted.image.indexOf('token_disc.svg'));
    expect(result.matted.size).toBe('cover, cover');
    expect(result.matted.repeat).toBe('no-repeat, no-repeat');
    // Art implies the frame: a picture with no edge does not read as a table.
    expect(result.matted.radius).not.toBe('0px');
    expect(result.matted.borderWidth).not.toBe('0px');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a grid board takes art in place of its flat surface colour', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async (art) => {
      await customElements.whenDefined('boardgame-game-board');
      const make = async (withArt: boolean) => {
        const board = document.createElement('boardgame-game-board');
        board.rows = 2;
        board.cols = 2;
        if (withArt) {
          board.art = art;
          board.artWash = 'rgba(255, 255, 255, 0.2)';
        }
        document.body.append(board);
        await board.updateComplete;
        const surface = board.shadowRoot!.querySelector('.board-surface')!;
        const style = getComputedStyle(surface);
        return { image: style.backgroundImage, color: style.backgroundColor };
      };
      return { plain: await make(false), arted: await make(true) };
    }, ART);

    expect(result.plain.image).toBe('none');
    expect(result.plain.color).toBe('rgb(45, 80, 22)');
    expect(result.arted.image).toContain('token_disc.svg');
    expect(result.arted.image).toContain('linear-gradient');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
