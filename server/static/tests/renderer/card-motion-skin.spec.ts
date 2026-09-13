import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('a generated cardView keeps its shared face skin on a historical motion carrier', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView, html } = await import('/src/client.ts');

      const cards = cardView<any>({
        render: ({ kind, component }) => kind === 'visible'
          ? html`<h2>${(component.Values as { Value: number }).Value}</h2>`
          : null,
      });
      const animator = document.createElement('boardgame-component-animator') as any;
      const makeStack = (left: number) => {
        const stack = document.createElement('boardgame-component-stack') as any;
        Object.assign(stack.style, { position: 'fixed', left: `${left}px`, top: '40px' });
        stack.style.setProperty('--animation-length', '40ms');
        stack.style.setProperty('--card-front-color', 'rgb(212, 232, 218)');
        stack.style.setProperty('--card-face-shadow', 'inset 0 0 0 4px rgb(175, 124, 48)');
        stack.componentView = cards;
        return stack;
      };
      const source = makeStack(20);
      const destination = makeStack(320);
      const card = {
        ID: 'card-8', Index: 0, Deck: 'cards', GameName: 'skin-test', Values: { Value: 8 },
      };
      const data = (
        components: readonly unknown[],
        ids: readonly string[],
        lastSeen: Record<string, number>,
      ) => ({
        Deck: 'cards', Indexes: components.map((_item, index) => index), IDs: ids,
        IDsLastSeen: lastSeen, ShuffleCount: 0, Size: components.length,
        GameName: 'skin-test', Components: components,
      });

      source.stack = data([card], ['card-8'], { 'card-8': 1 });
      destination.stack = data([], [], {});
      document.body.append(animator, source, destination);
      await Promise.all([animator.updateComplete, source.updateComplete, destination.updateComplete]);
      await source.Components[0].updateComplete;

      const cardPrototype = customElements.get('boardgame-card')!.prototype as any;
      const originalPlayAnimation = cardPrototype.playAnimation;
      let carrierAtPlayback: unknown = null;
      cardPrototype.playAnimation = function(record: unknown) {
        if (this.inert) {
          const root = this.shadowRoot!;
          const outer = root.querySelector<HTMLElement>('#outer')!;
          const front = root.querySelector<HTMLElement>('#front')!;
          const face = root.querySelector<HTMLElement>('#face')!;
          const fallback = root.querySelector<HTMLElement>('.fallback')!;
          carrierAtPlayback = {
            noContent: this.noContent,
            outerNoContent: outer.classList.contains('no-content'),
            faceDisplay: getComputedStyle(face).display,
            fallbackDisplay: getComputedStyle(fallback).display,
            frontBackground: getComputedStyle(front).backgroundColor,
            faceShadow: getComputedStyle(face).boxShadow,
            text: this.textContent?.trim(),
          };
        }
        return originalPlayAnimation.call(this, record);
      };

      try {
        animator.prepare();
        source.stack = data([], [], { 'card-8': 1 });
        destination.stack = data([], [], { 'card-8': 2 });
        await Promise.all([source.updateComplete, destination.updateComplete]);
        await animator.animateFlip();
        return {
          carrierAtPlayback,
          segmentStatus: animator._solvedMotionPlan?.segments[0]?.execution.status,
        };
      } finally {
        cardPrototype.playAnimation = originalPlayAnimation;
      }
    });

    expect(result).toEqual({
      carrierAtPlayback: {
        noContent: true,
        outerNoContent: true,
        faceDisplay: 'grid',
        fallbackDisplay: 'block',
        frontBackground: 'rgb(212, 232, 218)',
        faceShadow: 'rgb(175, 124, 48) 0px 0px 0px 4px inset',
        text: '8',
      },
      segmentStatus: 'finished',
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
