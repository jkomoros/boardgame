import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

/**
 * `boardgame-card`'s face: the corner / centre / footer box model and the two
 * art slots.
 *
 * Every assertion here is a MEASUREMENT, deliberately. The three games that
 * hand-rolled a card face did not get the class names wrong, they got the box
 * wrong -- content that did not fill the card, a footer that floated in the
 * middle, an art band that stopped short of the edge -- and a test that only
 * asked "is there a footer element" would have passed against every one of
 * those broken layouts.
 */

async function setup(page: import('@playwright/test').Page) {
  const diagnostics = await prepareRendererFixturePage(page);
  await page.evaluate(async () => {
    await import('/src/client.ts');
    await customElements.whenDefined('boardgame-card');
  });
  return diagnostics;
}

test('the face fills the card: art at the top edge, footer at the bottom, centre between', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const host = document.createElement('div');
      host.innerHTML = `
        <boardgame-card face-up art="/src/assets/token_cube.svg">
          <span slot="corner" id="pip">+2</span>
          <strong id="title">Carnivore</strong>
          <span slot="footer" id="marks">sun snow</span>
        </boardgame-card>`;
      document.body.append(host);
      const card = host.querySelector('boardgame-card')!;
      await card.updateComplete;
      // Two frames: the art band's presence is driven by a slotchange, and the
      // corner/footer emptiness by two more.
      await new Promise(requestAnimationFrame);
      await card.updateComplete;
      const root = card.shadowRoot!;
      const box = (selector: string) => {
        const node = root.querySelector(selector);
        return node ? node.getBoundingClientRect().toJSON() : null;
      };
      return {
        face: box('#face'),
        front: box('#front'),
        art: box('#art-band'),
        centre: box('#center'),
        footer: box('#footer'),
        corner: box('#corner'),
        title: host.querySelector('#title')!.getBoundingClientRect().toJSON(),
        marks: host.querySelector('#marks')!.getBoundingClientRect().toJSON(),
        artEmptyClass: root.querySelector('#art-band')!.className,
        footerEmptyClass: root.querySelector('#footer')!.className,
        cornerEmptyClass: root.querySelector('#corner')!.className,
      };
    });

    const face = result.face!;
    // The face IS the card front, not a box inside it. This is the whole of
    // what the three hand-rolled versions were re-deriving.
    expect(Math.round(face.width)).toBe(Math.round(result.front!.width));
    expect(Math.round(face.height)).toBe(Math.round(result.front!.height));

    // The art band runs edge to edge and starts at the top edge. darwin got
    // here with `width: calc(100% + 1.1rem)` and a negative margin.
    expect(Math.round(result.art!.left)).toBe(Math.round(face.left));
    expect(Math.round(result.art!.right)).toBe(Math.round(face.right));
    expect(Math.round(result.art!.top)).toBe(Math.round(face.top));
    // 45% of the face, the documented default.
    expect(result.art!.height / face.height).toBeGreaterThan(0.4);
    expect(result.art!.height / face.height).toBeLessThan(0.5);

    // The footer sits on the bottom edge, and the centre occupies what is left.
    expect(Math.round(result.footer!.bottom)).toBe(Math.round(face.bottom));
    expect(Math.round(result.centre!.top)).toBe(Math.round(result.art!.bottom));
    expect(Math.round(result.centre!.bottom)).toBe(Math.round(result.footer!.top));

    // Slotted content lands where its region is, not wherever it fell.
    expect(result.title.top).toBeGreaterThanOrEqual(result.centre!.top - 1);
    expect(result.marks.top).toBeGreaterThanOrEqual(result.footer!.top - 1);
    // The corner is out of flow and in the corner: it overlaps the art rather
    // than shortening it.
    expect(result.corner!.top).toBeLessThan(result.art!.bottom);
    expect(result.corner!.right).toBeLessThanOrEqual(face.right);

    expect(result.artEmptyClass).toBe('');
    expect(result.footerEmptyClass).toBe('');
    expect(result.cornerEmptyClass).toBe('');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a card with no art or footer reserves no room for either', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const host = document.createElement('div');
      host.innerHTML = '<boardgame-card face-up><strong id="only">Solo</strong></boardgame-card>';
      document.body.append(host);
      const card = host.querySelector('boardgame-card')!;
      await card.updateComplete;
      await new Promise(requestAnimationFrame);
      await card.updateComplete;
      const root = card.shadowRoot!;
      const rect = (selector: string) => root.querySelector(selector)!.getBoundingClientRect().toJSON();
      return {
        art: rect('#art-band'),
        footer: rect('#footer'),
        centre: rect('#center'),
        face: rect('#face'),
      };
    });

    // display:none, so zero on every axis -- not a zero-height band that would
    // still show a seam.
    expect(result.art.height).toBe(0);
    expect(result.art.width).toBe(0);
    expect(result.footer.height).toBe(0);
    // The centre gets the whole card.
    expect(Math.round(result.centre.height)).toBe(Math.round(result.face.height));

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('the classic suit/rank card is untouched by the face regions', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const card = document.createElement('boardgame-card');
      card.faceUp = true;
      card.suit = '♥';
      card.rank = 'Q';
      document.body.append(card);
      await card.updateComplete;
      await new Promise(requestAnimationFrame);
      const root = card.shadowRoot!;
      const front = root.querySelector('#front')!.getBoundingClientRect();
      const face = root.querySelector('#face')!;
      const faceRect = face.getBoundingClientRect();
      const faceStyle = getComputedStyle(face);
      const centreEl = root.querySelector('#center-rank') as HTMLElement;
      const top = root.querySelector('#top-rank')!.getBoundingClientRect();
      return {
        /*
         * #top-rank, #bottom-rank and #center-rank are `position: absolute`, so
         * they resolve against the nearest POSITIONED ancestor -- which used to
         * be #front and is now #face. They therefore land in exactly the same
         * place if and only if #face's padding box is #front's content box:
         * same rect, and no padding. That pair is the invariant, and it is
         * strictly stronger than measuring one rotated glyph's bounding box,
         * which is a function of the text inside it.
         */
        sameBox: faceRect.top === front.top && faceRect.left === front.left
          && faceRect.width === front.width && faceRect.height === front.height,
        facePadding: [faceStyle.paddingTop, faceStyle.paddingRight,
          faceStyle.paddingBottom, faceStyle.paddingLeft].join(' '),
        facePosition: faceStyle.position,
        // The corner index still sits in the bottom-left quadrant of the card.
        topRankInBottomLeft: top.left < front.left + front.width / 2
          && top.bottom > front.top + front.height / 2,
        // The big centre pip still spans the whole card. Offset sizes, not the
        // bounding rect: the pip carries a rotate(90deg), so its RECT is the
        // card's box transposed.
        centreSpansCard: centreEl.offsetWidth === Math.round(front.width)
          && centreEl.offsetHeight === Math.round(front.height),
        redSuit: getComputedStyle(centreEl).color,
      };
    });

    expect(result.sameBox).toBe(true);
    expect(result.facePadding).toBe('0px 0px 0px 0px');
    expect(result.facePosition).toBe('absolute');
    expect(result.topRankInBottomLeft).toBe(true);
    expect(result.centreSpansCard).toBe(true);
    // #B3362B, the authored red-suit ink.
    expect(result.redSuit).toBe('rgb(179, 54, 43)');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('back-art paints the back and replaces the star fallback', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const plain = document.createElement('boardgame-card');
      document.body.append(plain);
      await plain.updateComplete;

      const arted = document.createElement('boardgame-card');
      arted.backArt = '/src/assets/token_disc.svg';
      document.body.append(arted);
      await arted.updateComplete;

      const backArt = arted.shadowRoot!.querySelector('#back-art')!;
      const backBox = arted.shadowRoot!.querySelector('#back')!.getBoundingClientRect();
      const artBox = backArt.getBoundingClientRect();
      return {
        plainHasStars: !!plain.shadowRoot!.querySelector('#default-back'),
        plainHasArt: !!plain.shadowRoot!.querySelector('#back-art'),
        artedHasStars: !!arted.shadowRoot!.querySelector('#default-back'),
        background: getComputedStyle(backArt).backgroundImage,
        size: getComputedStyle(backArt).backgroundSize,
        covers: Math.round(artBox.width) === Math.round(backBox.width)
          && Math.round(artBox.height) === Math.round(backBox.height),
      };
    });

    expect(result.plainHasStars).toBe(true);
    expect(result.plainHasArt).toBe(false);
    // The art is the slot's default content, so it REPLACES the stars rather
    // than sitting behind them.
    expect(result.artedHasStars).toBe(false);
    expect(result.background).toContain('token_disc.svg');
    expect(result.size).toBe('cover');
    expect(result.covers).toBe(true);

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('front-color and ink-color are per-card, which is why they exist', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const make = async (front: string, ink: string) => {
        const card = document.createElement('boardgame-card');
        card.faceUp = true;
        card.suit = '♠';
        card.frontColor = front;
        card.inkColor = ink;
        document.body.append(card);
        await card.updateComplete;
        const root = card.shadowRoot!;
        return {
          front: getComputedStyle(root.querySelector('#front')!).backgroundColor,
          ink: getComputedStyle(root.querySelector('#center-rank')!).color,
        };
      };
      const plain = document.createElement('boardgame-card');
      plain.faceUp = true;
      document.body.append(plain);
      await plain.updateComplete;
      return {
        flux: await make('rgb(101, 52, 119)', 'rgb(255, 245, 220)'),
        forge: await make('rgb(255, 245, 220)', 'rgb(24, 50, 74)'),
        // Unset means "whatever the stylesheet says", so the deck default holds.
        untouched: getComputedStyle(plain.shadowRoot!.querySelector('#front')!).backgroundColor,
      };
    });

    // Two cards of the same deck, two colours -- the thing a stylesheet cannot
    // do and sequenceforge built a whole extra div for.
    expect(result.flux.front).toBe('rgb(101, 52, 119)');
    expect(result.flux.ink).toBe('rgb(255, 245, 220)');
    expect(result.forge.front).toBe('rgb(255, 245, 220)');
    expect(result.forge.ink).toBe('rgb(24, 50, 74)');
    expect(result.untouched).toBe('rgb(212, 232, 218)');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('tall may be set by the author, and slotted content still overrides it', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const settle = async (card: HTMLElement & { updateComplete: Promise<unknown> }) => {
        await card.updateComplete;
        await new Promise(requestAnimationFrame);
        await card.updateComplete;
      };

      // 1. Authored `tall`, content that says nothing. This used to be reset to
      //    false by the first firstUpdated() scan, every time.
      const host = document.createElement('div');
      host.innerHTML = '<boardgame-card tall face-up><strong>Quiet</strong></boardgame-card>';
      document.body.append(host);
      const authored = host.querySelector('boardgame-card')!;
      await settle(authored);

      // 2. Authored `tall`, no content at all -- a deck back.
      const bare = document.createElement('boardgame-card');
      bare.tall = true;
      document.body.append(bare);
      await settle(bare);

      // 3. Content that declares `tall`, which is debuganimations' channel and
      //    must keep working -- including the clear when the content goes away.
      const scanned = document.createElement('div');
      scanned.innerHTML = '<boardgame-card face-up><div tall>Loud</div></boardgame-card>';
      document.body.append(scanned);
      const fromContent = scanned.querySelector('boardgame-card')!;
      await settle(fromContent);
      const declaredByContent = fromContent.tall;
      fromContent.textContent = '';
      await settle(fromContent);

      return {
        authored: authored.tall,
        bare: bare.tall,
        declaredByContent,
        clearedWhenContentLeft: fromContent.tall,
      };
    });

    expect(result.authored).toBe(true);
    expect(result.bare).toBe(true);
    expect(result.declaredByContent).toBe(true);
    // The scan still has the last word once it has ever had something to say.
    expect(result.clearedWhenContentLeft).toBe(false);

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('an unknown art-fit is refused by name rather than emitting a broken background', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const message = await page.evaluate(async () => {
      const card = document.createElement('boardgame-card');
      card.art = '/src/assets/token_disc.svg';
      (card as unknown as { artFit: string }).artFit = 'fill';
      document.body.append(card);
      try {
        await card.updateComplete;
      } catch (error) {
        return (error as Error).message;
      }
      return `NO THROW: ${card.shadowRoot?.querySelector('#art-image')?.getAttribute('style')}`;
    });

    expect(message).toContain('art-fit must be "cover" or "contain", not "fill"');

    diagnostics.stop();
  } finally {
    diagnostics.stop();
  }
});
