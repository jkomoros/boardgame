import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('legacy component and stack motion hooks remain functional adapters', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { BoardgameComponent } = await import('/src/components/boardgame-component.ts');
      const { BoardgameComponentStack } = await import('/src/components/boardgame-component-stack.ts');
      const { componentView } = await import('/src/components/component-view.ts');

      class LegacyComponent extends BoardgameComponent {
        legacyPropertyCalls = 0;
        preparedBy = '';
        override get animatingProperties() { return ['turned']; }
        override playPropertyAnimation(_before: any, _after: any, delayMs = 0): void {
          this.legacyPropertyCalls++;
          this.play(this.shadowRoot!.querySelector<HTMLElement>('#inner')!, [
            { opacity: '0.2' }, { opacity: '1' },
          ], { duration: 40, delay: delayMs }, { timing: 'immediate' });
        }
        override prepareForBeingAnimatingComponent(stack: HTMLElement): void {
          this.preparedBy = stack.id;
        }
        override get cloneContent(): boolean { return true; }
        override animationRotates(before: any, after: any): boolean {
          return before.turned !== after.turned;
        }
      }
      customElements.define('legacy-motion-component', LegacyComponent);

      class LegacyStack extends BoardgameComponentStack {
        override setUnknownAnimationState(component: HTMLElement): void {
          component.style.transform = 'translateX(13px)';
          component.style.opacity = '0.25';
        }
      }
      customElements.define('legacy-motion-stack', LegacyStack);

      const standalone = document.createElement('legacy-motion-component') as LegacyComponent;
      document.body.append(standalone);
      await standalone.updateComplete;
      const animations = standalone.playAnimation({
        before: { turned: false }, after: { turned: true },
        invertedTransform: '', finalTransform: '',
        beforeOpacity: '1', finalOpacity: '1', needsHostTransition: false,
      });
      const innerAnimations = standalone.shadowRoot!.querySelector<HTMLElement>('#inner')!.getAnimations();
      const legacyAnimationCount = innerAnimations.length;
      for (const animation of innerAnimations) animation.finish();
      await standalone.settled();

      const stack = document.createElement('legacy-motion-stack') as LegacyStack;
      stack.id = 'legacy-stack';
      stack.componentView = componentView(
        () => document.createElement('legacy-motion-component'),
        {},
      );
      document.body.append(stack);
      await stack.updateComplete;
      const carrier = stack.newMotionCarrier();

      const live = document.createElement('legacy-motion-component') as LegacyComponent;
      live.style.transform = 'rotate(7deg)';
      live.style.opacity = '0.8';
      const sampled = stack.motionPresenceStyleFor(live);

      return {
        legacyPropertyCalls: standalone.legacyPropertyCalls,
        returnedPlannedAnimations: animations.length,
        legacyAnimationCount,
        clonePolicy: standalone.historicalPresentationPolicy,
        stackId: stack.id,
        carrierPreparedBy: carrier.component.preparedBy,
        carrierPresenceStyle: carrier.presenceStyle,
        sampled,
        liveRestingStyle: {
          transform: live.style.transform,
          opacity: live.style.opacity,
        },
        legacyAxesDiffer: standalone.animationRotates(
          { turned: false }, { turned: true },
        ),
      };
    });

    expect(result).toEqual({
      legacyPropertyCalls: 1,
      returnedPlannedAnimations: 0,
      legacyAnimationCount: 1,
      clonePolicy: 'clone-default-slot',
      stackId: result.stackId,
      carrierPreparedBy: result.stackId,
      carrierPresenceStyle: {
        transform: 'translateX(13px)',
        opacity: '0.25',
      },
      sampled: {
        transform: 'translateX(13px)',
        opacity: '0.25',
      },
      liveRestingStyle: {
        transform: 'rotate(7deg)',
        opacity: '0.8',
      },
      legacyAxesDiffer: true,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('legacy void playAnimation does not abort structural settlement', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { BoardgameComponent } = await import('/src/components/boardgame-component.ts');
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { componentView } = await import('/src/components/component-view.ts');

      class VoidPlaybackComponent extends BoardgameComponent {
        turned = false;
        legacyCalls = 0;
        override get animatingProperties() { return ['turned']; }
        override playPropertyAnimation(): void {
          this.legacyCalls++;
          this.play(this.shadowRoot!.querySelector<HTMLElement>('#inner')!, [
            { opacity: '0.2' }, { opacity: '1' },
          ], { duration: 40 }, { timing: 'immediate' });
        }
        override playAnimation(record: any): any {
          super.playAnimation(record);
          return undefined;
        }
      }
      customElements.define('legacy-void-playback', VoidPlaybackComponent);
      const animator = document.createElement('boardgame-component-animator') as any;
      const stack = document.createElement('boardgame-component-stack') as any;
      stack.style.setProperty('--animation-length', '40ms');
      stack.componentView = componentView(
        () => document.createElement('legacy-void-playback'),
        {
          properties: ({ component }: any) => ({
            turned: !!component?.Values?.turned,
          }),
        },
      );
      stack.stack = {
        Deck: 'pieces', Indexes: [0], IDs: ['legacy-piece'], IDsLastSeen: {},
        ShuffleCount: 0, Size: 1, GameName: 'legacy',
        Components: [{
          ID: 'legacy-piece', Index: 0, Deck: 'pieces', GameName: 'legacy',
          Values: { turned: false },
        }],
      };
      document.body.append(animator, stack);
      await Promise.all([animator.updateComplete, stack.updateComplete]);
      animator.prepare();
      stack.stack = {
        Deck: 'pieces', Indexes: [0], IDs: ['legacy-piece'], IDsLastSeen: {},
        ShuffleCount: 0, Size: 1, GameName: 'legacy',
        Components: [{
          ID: 'legacy-piece', Index: 0, Deck: 'pieces', GameName: 'legacy',
          Values: { turned: true },
        }],
      };
      await stack.updateComplete;
      await animator.animateFlip();
      return {
        phase: animator._solvedMotionPlan?.phase,
        execution: animator._solvedMotionPlan?.segments[0]?.execution,
        legacyCalls: stack.Components[0].legacyCalls,
      };
    });

    expect(result.phase).toBe('settled');
    expect(result.execution).toMatchObject({ status: 'skipped', reason: 'not-started' });
    expect(result.legacyCalls).toBe(1);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
