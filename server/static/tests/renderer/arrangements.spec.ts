import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage, RENDERER_VIEWPORTS } from './renderer-fixture-helpers.js';

test('stack geometry follows the rendered component and drives overlap and capped depth', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { cardView } = await import('/src/client.ts');
      const stack = document.createElement('boardgame-component-stack');
      stack.style.setProperty('--component-scale', '0.5');
      stack.layout = 'fan';
      stack.componentView = cardView({});
      stack.stack = {
        Deck: 'cards', Indexes: [0, 1, 2], IDs: ['a', 'b', 'c'], IDsLastSeen: {},
        ShuffleCount: 0, Size: 3, GameName: 'fixture',
        Components: [0, 1, 2].map((Index) => ({
          ID: ['a', 'b', 'c'][Index], Index, Deck: 'cards', GameName: 'fixture', Values: {},
        })),
      } as never;
      document.body.append(stack);
      await stack.updateComplete;
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
      const cards = [...stack.querySelectorAll<HTMLElement>('boardgame-card')];
      const first = {
        geometry: stack.slotGeometry,
        width: cards[0].offsetWidth,
        marginRight: getComputedStyle(cards[0]).marginRight,
      };

      stack.style.setProperty('--component-scale', '0.805');
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
      const resized = {
        geometry: stack.slotGeometry,
        layoutWidth: parseFloat(getComputedStyle(cards[0].shadowRoot!.querySelector('#outer')!).width),
        marginRight: getComputedStyle(cards[0]).marginRight,
      };

      // The component border box is unchanged here. Only the stack-authored 1em
      // margin changes, so observing the component alone cannot catch this.
      stack.style.fontSize = '25px';
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
      const remargined = {
        geometry: stack.slotGeometry,
        marginLeft: getComputedStyle(cards[0]).marginLeft,
        marginRight: getComputedStyle(cards[0]).marginRight,
      };

      stack.layout = 'stack';
      stack.stack = {
        ...stack.stack,
        Indexes: Array.from({ length: 9 }, (_value, index) => index),
        IDs: Array.from({ length: 9 }, (_value, index) => `depth-${index}`),
        Components: Array.from({ length: 9 }, (_value, Index) => ({
          ID: `depth-${Index}`, Index, Deck: 'cards', GameName: 'fixture', Values: {},
        })),
      } as never;
      await stack.updateComplete;
      await new Promise(resolve => requestAnimationFrame(resolve));
      const depthCards = [...stack.querySelectorAll<HTMLElement>('boardgame-card')];
      const depths = depthCards.map(card => ({
        depth: card.style.getPropertyValue('--boardgame-stack-depth'),
        top: getComputedStyle(card).top,
        zIndex: getComputedStyle(card).zIndex,
      }));
      stack.remove();
      return { first, resized, remargined, depths };
    });

    expect(result.first.width).toBe(50);
    expect(result.first.geometry?.componentInlineSize).toBe(50);
    expect(result.first.geometry?.inlineSize).toBeCloseTo(41, 0);
    expect(parseFloat(result.first.marginRight)).toBeCloseTo(-25, 0);
    expect(result.resized.layoutWidth).toBeCloseTo(80.5, 1);
    expect(result.resized.geometry?.componentInlineSize).toBeCloseTo(result.resized.layoutWidth, 3);
    expect(parseFloat(result.resized.marginRight)).toBeCloseTo(-40.25, 2);
    expect(parseFloat(result.remargined.marginLeft)).toBe(25);
    expect(parseFloat(result.remargined.marginRight)).toBeCloseTo(-40.25, 2);
    expect(result.remargined.geometry?.inlineSize).toBeCloseTo(65.25, 2);
    expect(result.depths.map(item => item.depth)).toEqual(['0', '1', '2', '3', '4', '5', '6', '6', '6']);
    expect(result.depths[0]).toMatchObject({ top: '0px', zIndex: '10' });
    expect(result.depths[8]).toMatchObject({ top: '6px', zIndex: '4' });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('deck and market preserve hidden sources, stable slot identity, attachments, and narrow scrolling', async ({ page }) => {
  await page.setViewportSize(RENDERER_VIEWPORTS.desktop);
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { cardView, tokenView } = await import('/src/client.ts');
      const makeStack = (
        deck: string,
        ids: string[],
        components: Array<Record<string, unknown> | null>,
        maxSize = 0,
      ) => ({
        Deck: deck,
        Indexes: components.map((_value, index) => index),
        IDs: ids,
        IDsLastSeen: Object.fromEntries(ids.map((id, index) => [id, index])),
        ShuffleCount: 0,
        ...(maxSize > 0 ? { MaxSize: maxSize } : {}),
        GameName: 'fixture',
        Components: components.map((component, index) => component === null ? null : ({
          ID: ids[index], Index: index, Deck: deck, GameName: 'fixture', Values: component,
        })),
      });
      const market = document.createElement('boardgame-market');
      market.label = 'Contracts';
      market.sourceLabel = 'Contract deck';
      market.displayLabel = 'Face-up contracts';
      market.attachmentLabel = 'Payments';
      market.attachmentPosition = 'before';
      market.style.width = '100%';
      market.style.setProperty('--component-scale', '0.8');
      market.sourceView = cardView({}).withProperties({ faceUp: false });
      market.componentView = cardView({}).withProperties({ faceUp: true });
      market.attachmentView = tokenView({
        properties: ({ index }) => ({ type: 'cube', cssColor: '#b87333', title: `payment-slot-${index}` }),
      });
      market.sourceStack = {
        ...makeStack('cards', ['hidden'], [{}]),
        Components: [{}],
      } as never;
      market.stack = makeStack('cards', ['one', 'two', 'three'], [{}, {}, {}], 3) as never;
      market.attachmentStack = {
        ...makeStack('payments', ['pay-one', '', 'pay-three'], [{}, null, {}]),
        Size: 3,
      } as never;
      const coins = document.createElement('boardgame-component-stack');
      coins.slot = 'attachment-0';
      coins.componentView = tokenView({}).withProperties({ type: 'chip', cssColor: 'gold' });
      coins.stack = makeStack('coins', ['coin'], [{}]) as never;
      coins.componentsDisabled = true;
      coins.style.setProperty('--component-scale', '0.5');
      market.append(coins);
      document.body.append(market);
      await market.updateComplete;
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));

      const root = market.shadowRoot!;
      const sourceDeck = root.querySelector('boardgame-deck')!;
      const sourceZone = sourceDeck.shadowRoot!.querySelector('boardgame-component-zone')!;
      const sourceStack = sourceZone.shadowRoot!.querySelector('boardgame-component-stack')!;
      const displayZone = root.querySelector<HTMLElement>('#display-zone')!;
      const displayStack = displayZone.shadowRoot!.querySelector('boardgame-component-stack')!;
      const cells = [...root.querySelectorAll<HTMLElement>('.attachment-cell')];
      const cards = [...displayStack.querySelectorAll<HTMLElement>('boardgame-card')];
      const splitStacks = cells.map(cell => cell.querySelector('boardgame-component-stack'));
      const paymentIDs = splitStacks.map(stack => stack?.stack?.IDs[0] ?? null);
      const paymentIndexes = splitStacks.map(stack => stack?.stack?.Indexes[0] ?? null);
      const paymentTokens = splitStacks.map(stack => stack?.querySelector<HTMLElement>('boardgame-token') ?? null);
      const paymentTitles = paymentTokens.map(token => token?.title ?? null);
      const center = (element: HTMLElement) => {
        const rect = element.getBoundingClientRect();
        return rect.left + rect.width / 2;
      };
      const cardCenters = cards.map(center);
      const paymentCenters = paymentTokens.map(token => token ? center(token) : null);
      const coinToken = coins.querySelector<HTMLElement>('boardgame-token')!;
      const hiddenCard = sourceStack.querySelector<HTMLElement>('boardgame-card')!;
      const marketSection = root.querySelector<HTMLElement>('#market')!;
      const slotSizeBeforeNested = marketSection.style.getPropertyValue('--boardgame-market-slot-inline-size');

      // A component recipe may itself contain a stack. Its composed geometry
      // event must not replace the market's top-level card measurement.
      const nested = document.createElement('boardgame-component-stack');
      nested.componentView = tokenView({}).withProperties({ type: 'cube' });
      nested.stack = makeStack('nested', ['nested'], [{}]) as never;
      nested.componentsDisabled = true;
      nested.style.setProperty('--component-width', '18px');
      cards[0].append(nested);
      await nested.updateComplete;
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
      const slotSizeAfterNested = marketSection.style.getPropertyValue('--boardgame-market-slot-inline-size');
      nested.remove();
      const originalCells = cells.slice();
      const wide = {
        sourceFaceUp: (hiddenCard as HTMLElement & { faceUp: boolean }).faceUp,
        sourceItemKeys: Object.keys((hiddenCard as HTMLElement & { item: object }).item),
        directStack: market.stackElement === displayStack,
        cellCount: cells.length,
        paymentIDs,
        paymentIndexes,
        paymentTitles,
        cardCenters,
        paymentCenters,
        cardWidth: cards[0].getBoundingClientRect().width,
        paymentWidth: paymentTokens[0]!.getBoundingClientRect().width,
        coinCenter: center(coinToken),
        coinWidth: coinToken.getBoundingClientRect().width,
        slotSizeBeforeNested,
        slotSizeAfterNested,
      };

      market.sourceStack = makeStack('cards', [], []) as never;
      market.stack = makeStack('cards', ['hidden', 'one'], [{}, {}], 3) as never;
      await market.updateComplete;
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
      const refilledIDs = [...displayStack.querySelectorAll<HTMLElement>('boardgame-card')].map(card => card.id);
      const depletedCells = [...root.querySelectorAll<HTMLElement>('.attachment-cell')];
      const exhausted = sourceZone.shadowRoot!.querySelector<HTMLElement>('#empty')?.textContent?.trim();
      return {
        wide,
        refilledIDs,
        depletedCellCount: depletedCells.length,
        sameCells: depletedCells.every((cell, index) => cell === originalCells[index]),
        exhausted,
      };
    });

    expect(result.wide.sourceFaceUp).toBe(false);
    expect(result.wide.sourceItemKeys).toEqual(['ID']);
    expect(result.wide.directStack).toBe(true);
    expect(result.wide.cellCount).toBe(3);
    expect(result.wide.paymentIDs).toEqual(['pay-one', '', 'pay-three']);
    expect(result.wide.paymentIndexes).toEqual([0, 1, 2]);
    expect(result.wide.paymentTitles).toEqual(['payment-slot-0', 'payment-slot-1', 'payment-slot-2']);
    expect(result.wide.paymentCenters[0]).toBeCloseTo(result.wide.cardCenters[0], 1);
    expect(result.wide.paymentCenters[2]).toBeCloseTo(result.wide.cardCenters[2], 1);
    expect(result.wide.coinCenter).toBeCloseTo(result.wide.cardCenters[0], 1);
    expect(result.wide.paymentWidth).toBeLessThan(result.wide.cardWidth / 2);
    expect(result.wide.coinWidth).toBeLessThan(result.wide.paymentWidth);
    expect(result.wide.slotSizeAfterNested).toBe(result.wide.slotSizeBeforeNested);
    expect(result.refilledIDs).toEqual(['hidden', 'one']);
    expect(result.depletedCellCount).toBe(3);
    expect(result.sameCells).toBe(true);
    expect(result.exhausted).toBe('Exhausted');

    await page.setViewportSize(RENDERER_VIEWPORTS.phone);
    await page.waitForTimeout(50);
    const narrow = await page.evaluate(() => {
      const market = document.querySelector('boardgame-market')!;
      const scroller = market.shadowRoot!.querySelector<HTMLElement>('#display-scroll')!;
      const pageWidth = document.documentElement.scrollWidth;
      const viewportWidth = document.documentElement.clientWidth;
      scroller.scrollLeft = 80;
      const displayZone = market.shadowRoot!.querySelector<HTMLElement>('#display-zone')!;
      const displayStack = displayZone.shadowRoot!.querySelector('boardgame-component-stack')!;
      const firstCard = displayStack.querySelector<HTMLElement>('boardgame-card')!;
      const firstCell = market.shadowRoot!.querySelector<HTMLElement>('.attachment-cell')!;
      const firstToken = firstCell.querySelector<HTMLElement>('boardgame-token')!;
      const center = (element: HTMLElement) => {
        const rect = element.getBoundingClientRect();
        return rect.left + rect.width / 2;
      };
      return {
        pageWidth,
        viewportWidth,
        scrollable: scroller.scrollWidth > scroller.clientWidth,
        cardCenter: center(firstCard),
        cellContentCenter: center(firstToken),
      };
    });
    expect(narrow.pageWidth).toBeLessThanOrEqual(narrow.viewportWidth + 1);
    expect(narrow.scrollable).toBe(true);
    expect(narrow.cellContentCenter).toBeCloseTo(narrow.cardCenter, 1);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('CSS colours tint generated solids without changing named materials or authored art', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/client.ts');
      const token = document.createElement('boardgame-token');
      token.type = 'cube';
      token.color = 'orange';
      document.body.append(token);
      await token.updateComplete;
      const fills = () => [...token.shadowRoot!.querySelectorAll<HTMLElement>('.facet')]
        .map(facet => getComputedStyle(facet).backgroundColor);
      const legacyBefore = fills();
      for (let index = 0; index < 120; index++) {
        token.cssColor = `hsl(${index * 3} 70% 45%)`;
        await token.updateComplete;
      }
      token.cssColor = '#b87333';
      await token.updateComplete;
      const custom = {
        fills: fills(),
        variable: token.shadowRoot!.querySelector<HTMLElement>('#solid')!.style
          .getPropertyValue('--boardgame-token-css-color'),
        facetStyles: [...token.shadowRoot!.querySelectorAll<HTMLElement>('.facet')]
          .map(facet => facet.style.background),
      };
      token.cssColor = '';
      await token.updateComplete;
      const legacyAfter = fills();

      token.art = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg"/>';
      token.cssColor = 'lime';
      await token.updateComplete;
      const art = {
        solid: token.shadowRoot!.querySelector('#solid') !== null,
        filter: getComputedStyle(token.shadowRoot!.querySelector('img')!).filter,
      };
      token.remove();
      return { legacyBefore, custom, legacyAfter, art };
    });

    expect(result.legacyBefore).toEqual(result.legacyAfter);
    expect(result.custom.variable).toBe('#b87333');
    expect(result.custom.fills).not.toEqual(result.legacyBefore);
    expect(new Set(result.custom.fills).size).toBeGreaterThan(1);
    expect(result.custom.facetStyles.every(style => style.includes('var(--boardgame-token-css-color)'))).toBe(true);
    expect(result.art).toEqual({ solid: false, filter: 'none' });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
