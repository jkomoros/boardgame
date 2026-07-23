import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('automatic Hand arrival remains visible inside a transformed layout', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');

      const animator = document.createElement('boardgame-component-animator') as any;
      const transformed = document.createElement('div');
      transformed.style.transform = 'scale(0.9)';
      const stack = document.createElement('boardgame-component-stack') as any;
      stack.style.setProperty('--animation-length', '600ms');
      stack.componentView = cardView({});
      const beforeStack = {
        Deck: 'cards', Indexes: [], IDs: [], IDsLastSeen: {},
        ShuffleCount: 0, Size: 0, GameName: 'compatibility', Components: [],
      };
      const afterStack = {
        Deck: 'cards', Indexes: [0], IDs: ['transformed-deal'], IDsLastSeen: {},
        ShuffleCount: 0, Size: 1, GameName: 'compatibility', Components: [{
          ID: 'transformed-deal', Index: 0, Deck: 'cards', GameName: 'compatibility',
          Values: { rank: 'A' },
        }],
      };
      stack.stack = beforeStack;
      const anchor = document.createElement('div');
      anchor.id = 'hand-top-edge';
      Object.assign(anchor.style, {
        position: 'fixed', left: '20px', top: '10px', width: '10px', height: '10px',
      });
      Object.assign(stack.style, { position: 'fixed', left: '300px', top: '200px' });
      transformed.append(stack);
      document.body.append(animator, transformed, anchor);
      await Promise.all([animator.updateComplete, stack.updateComplete]);

      stack.stack = afterStack;
      await stack.updateComplete;
      const card = stack.Components[0] as HTMLElement;
      let settled = false;
      const settlement = animator.animateBetween(
        card, anchor, 600, { timing: 'immediate' },
      ).then(() => { settled = true; });
      for (let frame = 0; frame < 20 && card.getAnimations().length === 0 && !settled; frame++) {
        await new Promise(requestAnimationFrame);
      }
      const visibleAnimations = card.getAnimations().filter(animation => animation.playState !== 'idle');
      for (const animation of visibleAnimations) animation.finish();
      await settlement;
      return {
        visibleAnimationCount: visibleAnimations.length,
        settled,
      };
    });

    // Compatibility contract: transformed responsive layouts may alter the
    // geometry, but automatic dealing must not disappear entirely.
    expect(result.visibleAnimationCount).toBeGreaterThan(0);
    expect(result.settled).toBe(true);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('multi-card automatic Hand arrivals start together in their final visual pose', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');

      const ids = ['simultaneous-a', 'simultaneous-b'];

      const animator = document.createElement('boardgame-component-animator') as any;
      const stack = document.createElement('boardgame-component-stack') as any;
      stack.stagger = 0.5;
      stack.componentView = cardView({
        properties: ({ component }: any) => ({ faceUp: component.Values.faceUp }),
      });
      const beforeStack = {
        Deck: 'cards', Indexes: [], IDs: [], IDsLastSeen: {},
        ShuffleCount: 0, Size: 0, GameName: 'compatibility', Components: [],
      };
      const afterStack = {
        Deck: 'cards', Indexes: [0, 1], IDs: ids, IDsLastSeen: {},
        ShuffleCount: 0, Size: 2, GameName: 'compatibility', Components: ids.map((ID, Index) => ({
          ID, Index, Deck: 'cards', GameName: 'compatibility',
          Values: { faceUp: Index === 0 },
        })),
      };
      stack.stack = beforeStack;
      Object.assign(stack.style, { position: 'fixed', left: '300px', top: '200px' });
      const anchor = document.createElement('div');
      anchor.id = 'hand-top-edge';
      Object.assign(anchor.style, {
        position: 'fixed', left: '20px', top: '10px', width: '10px', height: '10px',
      });
      document.body.append(animator, stack, anchor);
      await Promise.all([animator.updateComplete, stack.updateComplete]);

      stack.stack = afterStack;
      await stack.updateComplete;
      let settled = false;
      const settlement = Promise.all((stack.Components as HTMLElement[]).map(card => (
        animator.animateBetween(card, anchor, 600, { timing: 'immediate' })
      ))).then(() => { settled = true; });
      for (let frame = 0; frame < 20
        && (stack.Components as HTMLElement[]).every(card => card.getAnimations({ subtree: true }).length === 0)
        && !settled; frame++) {
        await new Promise(requestAnimationFrame);
      }
      const observations = (stack.Components as HTMLElement[]).map(card => {
        const animations = card.getAnimations({ subtree: true });
        return {
          hostDelays: card.getAnimations().map(animation => (
            animation.effect instanceof KeyframeEffect
              ? animation.effect.getTiming().delay
              : null
          )),
          visualAnimations: animations.filter(animation => (
            animation.effect instanceof KeyframeEffect
              && animation.effect.target !== card
          )).length,
        };
      });
      for (const card of stack.Components as HTMLElement[]) {
        for (const animation of card.getAnimations({ subtree: true })) animation.finish();
      }
      await settlement;
      return observations;
    });

    // Legacy autoFlyIncoming launched each final-state carrier in the same
    // callback; stack stagger and destination defaults did not rewrite it.
    expect(result).toEqual([
      { hostDelays: [0], visualAnimations: 0 },
      { hostDelays: [0], visualAnimations: 0 },
    ]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('automatic Table stub flight remains decorative and does not gate structural settlement', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');

      const animator = document.createElement('boardgame-component-animator') as any;
      // One registered stack establishes the renderer root used by endpoint lookup.
      const registryStack = document.createElement('boardgame-component-stack') as any;
      registryStack.componentView = cardView({});
      registryStack.stack = {
        Deck: 'cards', Indexes: [], IDs: [], IDsLastSeen: {}, ShuffleCount: 0,
        Size: 0, GameName: 'compatibility', Components: [],
      };
      const source = document.createElement('div');
      source.id = 'deal-source';
      Object.assign(source.style, {
        position: 'fixed', left: '20px', top: '20px', width: '40px', height: '60px',
      });
      const stub = document.createElement('div');
      stub.id = 'stub:p0:hand';
      Object.assign(stub.style, {
        position: 'fixed', left: '350px', top: '250px', width: '40px', height: '60px',
      });
      document.body.append(animator, registryStack, source, stub);
      await Promise.all([animator.updateComplete, registryStack.updateComplete]);

      const flight = animator.animateBetween(stub, source, 600, { timing: 'immediate' });
      let settled = false;
      const settlement = animator.animateFlip().then(() => { settled = true; });
      for (let frame = 0; frame < 20 && stub.getAnimations().length === 0 && !settled; frame++) {
        await new Promise(requestAnimationFrame);
      }
      const animations = stub.getAnimations();
      await Promise.race([
        settlement,
        new Promise(resolve => setTimeout(resolve, 50)),
      ]);
      const result = { settledWhileFlying: settled, visibleAnimationCount: animations.length };
      for (const animation of animations) animation.finish();
      await Promise.all([settlement, flight]);
      return result;
    });

    expect(result.visibleAnimationCount).toBe(1);
    expect(result.settledWhileFlying).toBe(true);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('resolving a Hand viewer identity does not replay long-held cards as incoming', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { BoardgameHandViewBase } = await import('/src/components/boardgame-hand-view-base.ts');
      class CompatibilityHand extends BoardgameHandViewBase<any, any, any, any> {
        calls: string[] = [];
        protected override get animator(): any {
          return {
            animateBetween: (carrier: string) => { this.calls.push(carrier); },
          };
        }
      }
      customElements.define('compatibility-viewer-reset-hand', CompatibilityHand);
      const hand = document.createElement('compatibility-viewer-reset-hand') as CompatibilityHand;
      // The previous snapshot was installed while the renderer had no resolved
      // private player. The next snapshot exposes an already-held hand when the
      // seat identity resolves; this is not an authoritative card arrival.
      document.body.append(hand);
      hand.viewingAsPlayer = -1;
      hand.state = { Players: [] } as any;
      await hand.updateComplete;
      hand.viewingAsPlayer = 1;
      await hand.updateComplete;
      hand.state = { Players: [
        {},
        { Hand: { IDs: ['held-before-identity', 'also-held-before-identity'] } },
      ] } as any;
      await hand.updateComplete;
      hand.state = { Players: [
        {},
        { Hand: { IDs: [
          'held-before-identity', 'also-held-before-identity', 'actually-new',
        ] } },
      ] } as any;
      await hand.updateComplete;
      await new Promise(requestAnimationFrame);
      await new Promise(requestAnimationFrame);
      return hand.calls;
    });

    expect(result).toEqual(['actually-new']);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
