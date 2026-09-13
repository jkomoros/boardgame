import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('a generated card carrier preserves shared face regions, appearance, and leaf values', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      await import('/src/components/boardgame-stat.ts');
      await import('/src/components/boardgame-pips.ts');
      const { cardView, html } = await import('/src/client.ts');

      const cards = cardView<any>({
        render: ({ kind, component }) => kind === 'visible'
          ? html`
              <boardgame-stat id="private-source-id"
                .value=${component.Values.Value}></boardgame-stat>
              <span slot="art">art</span>
              <small slot="footer">FORGE <boardgame-pips
                .count=${component.Values.Pips}
                glyph="✦"></boardgame-pips></small>
              <span slot="corner">C</span>
            `
          : null,
        properties: ({ kind }) => kind === 'visible'
          ? {
              frontColor: '#fff5dc',
              inkColor: '#18324a',
              tall: true,
              aspectRatio: 0.7,
            }
          : {},
      });
      const animator = document.createElement('boardgame-component-animator') as any;
      const makeStack = (left: number) => {
        const stack = document.createElement('boardgame-component-stack') as any;
        Object.assign(stack.style, { position: 'fixed', left: `${left}px`, top: '40px' });
        stack.style.setProperty('--animation-length', '40ms');
        stack.style.setProperty('--card-face-shadow', 'inset 0 0 0 4px rgb(227, 185, 87)');
        stack.componentView = cards;
        return stack;
      };
      const source = makeStack(20);
      const destination = makeStack(320);
      const visible = {
        ID: 'forge-8', Index: 0, Deck: 'cards', GameName: 'presentation-test',
        Values: { Value: 8, Pips: 3 },
      };
      const data = (
        components: readonly unknown[],
        ids: readonly string[],
        lastSeen: Record<string, number>,
      ) => ({
        Deck: 'cards', Indexes: components.map((_item, index) => index), IDs: ids,
        IDsLastSeen: lastSeen, ShuffleCount: 0, Size: components.length,
        GameName: 'presentation-test', Components: components,
      });

      source.stack = data([visible], ['forge-8'], { 'forge-8': 1 });
      destination.stack = data([], [], {});
      document.body.append(animator, source, destination);
      await Promise.all([animator.updateComplete, source.updateComplete, destination.updateComplete]);
      const live = source.Components[0] as any;
      await live.updateComplete;
      // Exercise relative units on the visible source and a different
      // intrinsic basis on the fresh destination carrier. Historical capture
      // must preserve proportions rather than parsing "6rem" as six pixels.
      live.style.setProperty('--default-component-width', '6rem');
      live.style.setProperty('--card-face-font-size', '1.75rem');
      const liveRoot = live.shadowRoot!;
      const liveInner = liveRoot.querySelector<HTMLElement>('#inner')!;
      const liveCenter = liveRoot.querySelector<HTMLElement>('#center')!;
      const liveFooter = liveRoot.querySelector<HTMLElement>('#footer')!;
      const liveFront = liveRoot.querySelector<HTMLElement>('#front')!;
      const liveFace = liveRoot.querySelector<HTMLElement>('#face')!;
      const liveCenterRect = liveCenter.getBoundingClientRect();
      const liveInnerRect = liveInner.getBoundingClientRect();
      const liveBasis = liveInnerRect.height;
      const sourcePresentation = {
        centerColor: getComputedStyle(liveCenter).color,
        centerFontSize: getComputedStyle(liveCenter).fontSize,
        basis: liveBasis,
        centerWidthRatio: liveCenterRect.width / liveBasis,
        centerHeightRatio: liveCenterRect.height / liveBasis,
        fontRatio: parseFloat(getComputedStyle(liveCenter).fontSize) / liveBasis,
        footerDisplay: getComputedStyle(liveFooter).display,
        frontBackground: getComputedStyle(liveFront).backgroundColor,
        faceShadow: getComputedStyle(liveFace).boxShadow,
      };

      const cardPrototype = customElements.get('boardgame-card')!.prototype as any;
      const originalPlayAnimation = cardPrototype.playAnimation;
      let carrierAtPlayback: any = null;
      cardPrototype.playAnimation = function(record: unknown) {
        if (this.inert) {
          const root = this.shadowRoot!;
          const center = root.querySelector<HTMLElement>('#center')!;
          const inner = root.querySelector<HTMLElement>('#inner')!;
          const footer = root.querySelector<HTMLElement>('#footer')!;
          const front = root.querySelector<HTMLElement>('#front')!;
          const face = root.querySelector<HTMLElement>('#face')!;
          const centerRect = center.getBoundingClientRect();
          const innerRect = inner.getBoundingClientRect();
          const basis = innerRect.height;
          const stat = this.querySelector('boardgame-stat') as any;
          const pips = this.querySelector('boardgame-pips') as any;
          carrierAtPlayback = {
            noContent: this.noContent,
            centerColor: getComputedStyle(center).color,
            centerFontSize: getComputedStyle(center).fontSize,
            basis,
            centerWidthRatio: centerRect.width / basis,
            centerHeightRatio: centerRect.height / basis,
            fontRatio: parseFloat(getComputedStyle(center).fontSize) / basis,
            footerDisplay: getComputedStyle(footer).display,
            frontBackground: getComputedStyle(front).backgroundColor,
            faceShadow: getComputedStyle(face).boxShadow,
            bodySlot: stat?.getAttribute('slot'),
            footerSlot: pips?.parentElement?.getAttribute('slot'),
            artSlot: this.querySelector('[slot="motion-history-art"]')?.getAttribute('slot'),
            cornerSlot: this.querySelector('[slot="motion-history-corner"]')?.getAttribute('slot'),
            descendantIds: this.querySelectorAll('[id]').length,
            statValue: stat?.value,
            pipCount: pips?.count,
            publicFrontColor: this.frontColor,
            publicInkColor: this.inkColor,
          };
        }
        return originalPlayAnimation.call(this, record);
      };

      try {
        animator.prepare();
        source.stack = data([], [], { 'forge-8': 1 });
        destination.stack = data([], [], { 'forge-8': 2 });
        await Promise.all([source.updateComplete, destination.updateComplete]);
        await animator.animateFlip();
        const carrier = animator._animatingComponents[0]?.component as HTMLElement | undefined;
        return {
          sourcePresentation,
          carrierAtPlayback,
          historyAfterSettle: carrier?.querySelectorAll('[slot^="motion-history"]').length ?? -1,
        };
      } finally {
        cardPrototype.playAnimation = originalPlayAnimation;
      }
    });

    expect(result.carrierAtPlayback).toMatchObject({
      noContent: true,
      centerColor: result.sourcePresentation.centerColor,
      footerDisplay: result.sourcePresentation.footerDisplay,
      frontBackground: result.sourcePresentation.frontBackground,
      faceShadow: result.sourcePresentation.faceShadow,
      bodySlot: 'motion-history-center',
      footerSlot: 'motion-history-footer',
      artSlot: 'motion-history-art',
      cornerSlot: 'motion-history-corner',
      descendantIds: 0,
      statValue: 8,
      pipCount: 3,
      // Presentation overrides must not become authored public card state.
      publicFrontColor: '',
      publicInkColor: '',
    });
    expect(Math.abs(result.carrierAtPlayback.basis - result.sourcePresentation.basis))
      .toBeGreaterThan(2);
    expect(Math.abs(
      result.carrierAtPlayback.centerWidthRatio - result.sourcePresentation.centerWidthRatio,
    )).toBeLessThan(0.01);
    expect(Math.abs(
      result.carrierAtPlayback.centerHeightRatio - result.sourcePresentation.centerHeightRatio,
    )).toBeLessThan(0.01);
    expect(Math.abs(
      result.carrierAtPlayback.fontRatio - result.sourcePresentation.fontRatio,
    )).toBeLessThan(0.001);
    expect(result.historyAfterSettle).toBe(0);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('retained visible-to-hidden cards use temporary history without mutating live state or retaining stale footer', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView, html } = await import('/src/client.ts');
      const cards = cardView<any>({
        render: ({ kind, component }) => kind === 'visible'
          ? html`<span>${component.Values.Value}</span>${component.Values.Footer
            ? html`<small slot="footer">FORGE</small>`
            : null}`
          : null,
        properties: ({ kind }) => kind === 'visible'
          ? { frontColor: '#fff5dc', inkColor: '#18324a' }
          : { frontColor: '#abcdef', inkColor: '#654321' },
      });
      const animator = document.createElement('boardgame-component-animator') as any;
      const stack = document.createElement('boardgame-component-stack') as any;
      stack.style.setProperty('--animation-length', '30ms');
      stack.componentView = cards;
      const visible = (footer: boolean) => ({
        ID: 'retained-card', Index: 0, Deck: 'cards', GameName: 'retained-test',
        Values: { Value: 8, Footer: footer },
      });
      const data = (component: unknown, seen: number) => ({
        Deck: 'cards', Indexes: [0], IDs: ['retained-card'],
        IDsLastSeen: { 'retained-card': seen }, ShuffleCount: 0, Size: 1,
        GameName: 'retained-test', Components: [component],
      });
      stack.stack = data(visible(true), 1);
      document.body.append(animator, stack);
      await Promise.all([animator.updateComplete, stack.updateComplete]);
      await stack.Components[0].updateComplete;

      const cardPrototype = customElements.get('boardgame-card')!.prototype as any;
      const originalPlayAnimation = cardPrototype.playAnimation;
      const captures: any[] = [];
      cardPrototype.playAnimation = function(record: unknown) {
        if (this.id === 'retained-card' && this.noContent) {
          const root = this.shadowRoot!;
          captures.push({
            frontBackground: getComputedStyle(root.querySelector('#front')!).backgroundColor,
            centerColor: getComputedStyle(root.querySelector('#center')!).color,
            historyText: [...this.querySelectorAll('[slot^="motion-history"]')]
              .map((node: Element) => node.textContent?.trim()).join(' '),
            publicFrontColor: this.frontColor,
            publicInkColor: this.inkColor,
          });
        }
        return originalPlayAnimation.call(this, record);
      };

      const transition = async (next: unknown, seen: number) => {
        animator.prepare();
        stack.stack = data(next, seen);
        await stack.updateComplete;
        await animator.animateFlip();
      };
      try {
        await transition({}, 2);
        const afterFirst = stack.Components[0] as any;
        const firstSettled = {
          historyCount: afterFirst.querySelectorAll('[slot^="motion-history"]').length,
          publicFrontColor: afterFirst.frontColor,
          publicInkColor: afterFirst.inkColor,
        };
        await transition(visible(false), 3);
        await transition({}, 4);
        const afterSecond = stack.Components[0] as any;
        return {
          captures,
          firstSettled,
          secondSettledHistoryCount:
            afterSecond.querySelectorAll('[slot^="motion-history"]').length,
        };
      } finally {
        cardPrototype.playAnimation = originalPlayAnimation;
      }
    });

    expect(result.captures).toHaveLength(2);
    expect(result.captures[0]).toEqual({
      frontBackground: 'rgb(255, 245, 220)',
      centerColor: 'rgb(24, 50, 74)',
      historyText: '8 FORGE',
      publicFrontColor: '#abcdef',
      publicInkColor: '#654321',
    });
    expect(result.firstSettled).toEqual({
      historyCount: 0,
      publicFrontColor: '#abcdef',
      publicInkColor: '#654321',
    });
    expect(result.captures[1]).toEqual({
      frontBackground: 'rgb(255, 245, 220)',
      centerColor: 'rgb(24, 50, 74)',
      historyText: '8',
      publicFrontColor: '#abcdef',
      publicInkColor: '#654321',
    });
    expect(result.secondSettledHistoryCount).toBe(0);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('property-only classic faces survive a faux transfer and hidden reappearance cannot replay cached ink', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');
      const cards = cardView<any>({
        properties: ({ kind }) => kind === 'visible'
          ? { suit: '♥', rank: 'A', inkColor: '#b3362b' }
          : {},
      });
      const animator = document.createElement('boardgame-component-animator') as any;
      const makeStack = () => {
        const stack = document.createElement('boardgame-component-stack') as any;
        stack.style.setProperty('--animation-length', '30ms');
        stack.componentView = cards;
        return stack;
      };
      const source = makeStack();
      const destination = makeStack();
      const visible = {
        ID: 'classic-card', Index: 0, Deck: 'cards', GameName: 'classic-test', Values: {},
      };
      const data = (
        components: readonly unknown[], ids: readonly string[], lastSeen: Record<string, number>,
      ) => ({
        Deck: 'cards', Indexes: components.map((_item, index) => index), IDs: ids,
        IDsLastSeen: lastSeen, ShuffleCount: 0, Size: components.length,
        GameName: 'classic-test', Components: components,
      });
      source.stack = data([visible], ['classic-card'], { 'classic-card': 1 });
      destination.stack = data([], [], {});
      document.body.append(animator, source, destination);
      await Promise.all([animator.updateComplete, source.updateComplete, destination.updateComplete]);

      const cardPrototype = customElements.get('boardgame-card')!.prototype as any;
      const originalPlayAnimation = cardPrototype.playAnimation;
      const captures: any[] = [];
      cardPrototype.playAnimation = function(record: unknown) {
        const root = this.shadowRoot!;
        captures.push({
          inert: this.inert,
          noContent: this.noContent,
          rank: root.querySelector('#top-rank')?.textContent,
          historyCount: this.querySelectorAll('[slot^="motion-history"]').length,
        });
        return originalPlayAnimation.call(this, record);
      };

      try {
        animator.prepare();
        source.stack = data([], [], { 'classic-card': 1 });
        destination.stack = data([], [], { 'classic-card': 2 });
        await Promise.all([source.updateComplete, destination.updateComplete]);
        await animator.animateFlip();

        // The cached visible face now exists only as history. A later exact
        // hidden appearance must revoke it before any appearing animation.
        animator.clearAnimatingComponents();
        animator.prepare();
        source.stack = data([], [], { 'classic-card': 2 });
        destination.stack = data([{}], ['classic-card'], { 'classic-card': 3 });
        await Promise.all([source.updateComplete, destination.updateComplete]);
        await animator.animateFlip();
        return captures;
      } finally {
        cardPrototype.playAnimation = originalPlayAnimation;
      }
    });

    expect(result[0]).toMatchObject({
      inert: true,
      noContent: true,
      rank: '♥A',
      historyCount: 0,
    });
    const hiddenAppearance = result.find((capture: any, index: number) =>
      index > 0 && !capture.inert && capture.noContent);
    expect(hiddenAppearance).toMatchObject({
      noContent: true,
      rank: '',
      historyCount: 0,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
