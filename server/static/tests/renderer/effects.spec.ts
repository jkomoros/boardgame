import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('burst effects are capped, deterministic, and clean themselves up', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-effect-layer.ts');
      const anchor = document.createElement('button');
      anchor.id = 'score-anchor';
      anchor.textContent = 'Score';
      document.body.append(anchor);
      const layer = document.createElement('boardgame-effect-layer') as HTMLElement & {
        updateComplete: Promise<unknown>;
        shadowRoot: ShadowRoot;
        burst(anchor: HTMLElement, options: object): { finished: Promise<void> };
        cancelAll(): void;
      };
      document.body.append(layer);
      await layer.updateComplete;
      const handles = ['one', 'two', 'three'].map(seed => layer.burst(anchor, {
        preset: 'score',
        count: 100,
        duration: 120,
        seed: `fixture-score-${seed}`,
      }));
      const particles = [...layer.shadowRoot.querySelectorAll<HTMLElement>('.particle')];
      const initialStyles = particles.map(particle => ({
        size: particle.style.getPropertyValue('--particle-size'),
        color: particle.style.getPropertyValue('--particle-color'),
      }));
      await Promise.all(handles.map(handle => handle.finished));
      const cancelled = layer.burst(anchor, { duration: 1200, seed: 'route-change' });
      const countBeforeCancel = layer.shadowRoot.querySelectorAll('.particle').length;
      layer.cancelAll();
      await cancelled.finished;
      return {
        initialCount: particles.length,
        initialStyles,
        finalCount: layer.shadowRoot.querySelectorAll('.particle').length,
        countBeforeCancel,
      };
    });

    expect(result.initialCount).toBe(60);
    expect(result.initialStyles.every(style => style.size && style.color)).toBe(true);
    expect(result.countBeforeCancel).toBeGreaterThan(0);
    expect(result.finalCount).toBe(0);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('reduced motion skips decorative traveling particles', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const count = await page.evaluate(async () => {
      await import('/src/components/boardgame-effect-layer.ts');
      const anchor = document.createElement('div');
      document.body.append(anchor);
      const layer = document.createElement('boardgame-effect-layer') as HTMLElement & {
        updateComplete: Promise<unknown>;
        shadowRoot: ShadowRoot;
        burst(anchor: HTMLElement): { finished: Promise<void> };
      };
      document.body.append(layer);
      await layer.updateComplete;
      await layer.burst(anchor).finished;
      return layer.shadowRoot.querySelectorAll('.particle').length;
    });
    expect(count).toBe(0);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
