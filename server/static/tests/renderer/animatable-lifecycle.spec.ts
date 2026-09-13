import { test, expect } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

for (const mode of ['disabled', 'missing-target', 'playback-error'] as const) {
  test(`owned tracks preserve their resting presentation after ${mode}`, async ({ page }) => {
    await prepareRendererFixturePage(page);
    const result = await page.evaluate(async mode => {
      const { BoardgameAnimatableItem } = await import('/src/components/boardgame-animatable-item.ts');
      const { componentMotionTracks } = await import('/src/motion/component-track.ts');
      class OwnedTracks extends BoardgameAnimatableItem {
        visual = document.createElement('div');
        protected override motionTrackTarget(target: 'host' | 'visual') {
          return target === 'visual' ? this.visual : mode === 'missing-target' ? null : this;
        }
        run() {
          return this.playMotionTracks(componentMotionTracks([
            { target: 'host', property: 'transform', from: 'translateX(10px)', to: 'none' },
            { target: 'visual', property: 'opacity', curve: p => String(0.25 + p / 2), resting: '0.75' },
          ]));
        }
      }
      customElements.define('owned-tracks', OwnedTracks);
      const item = document.createElement('owned-tracks') as OwnedTracks;
      item.append(item.visual); document.body.append(item); await item.updateComplete;
      item.noAnimate = mode === 'disabled';
      if (mode === 'playback-error') item.animate = () => { throw new Error('injected playback failure'); };
      return { result: item.run(), opacity: item.visual.style.opacity, live: item.getAnimations({ subtree: true }).length };
    }, mode);
    expect(result.result.status).toBe('skipped');
    expect(result.opacity).toBe('0.75'); expect(result.live).toBe(0);
  });
}

test('ungated loops survive synchronous reparenting and stop on removal', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await prepareRendererFixturePage(page);
  const result = await page.evaluate(async () => {
    const { BoardgameAnimatableItem } = await import('/src/components/boardgame-animatable-item.ts');
    const item = new BoardgameAnimatableItem();
    const destination = document.createElement('div'); document.body.append(item, destination);
    await item.updateComplete;
    const animation = item.play(item, [{ opacity: 0.4 }, { opacity: 1 }], { duration: 1000, iterations: Infinity }, { gated: false, timing: 'immediate' })!;
    destination.append(item); await Promise.resolve();
    const reparented = animation.playState;
    item.remove(); await Promise.resolve(); await Promise.resolve();
    return { reparented, removed: animation.playState, gated: item.isAnimating };
  });
  expect(result).toEqual({ reparented: 'running', removed: 'idle', gated: false });
});
