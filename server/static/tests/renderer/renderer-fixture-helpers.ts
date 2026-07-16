import { expect, type Locator, type Page } from '@playwright/test';

export const RENDERER_VIEWPORTS = {
  phone: { width: 320, height: 640 },
  tablet: { width: 768, height: 1024 },
  desktop: { width: 1280, height: 800 },
} as const;

export interface RendererDiagnosticCapture {
  assertEmpty(): void;
  stop(): void;
}

export function captureRendererDiagnostics(page: Page): RendererDiagnosticCapture {
  const diagnostics: string[] = [];
  const onConsole = (message: { type(): string; text(): string }): void => {
    if (message.type() === 'error') {
      diagnostics.push(`console.error: ${message.text()}`);
    }
  };
  const onPageError = (error: Error): void => {
    diagnostics.push(`pageerror: ${error.message}`);
  };
  page.on('console', onConsole);
  page.on('pageerror', onPageError);
  return {
    assertEmpty(): void {
      expect(diagnostics, 'renderer emitted browser diagnostics').toEqual([]);
    },
    stop(): void {
      page.off('console', onConsole);
      page.off('pageerror', onPageError);
    },
  };
}

export async function prepareRendererFixturePage(page: Page): Promise<RendererDiagnosticCapture> {
  // Start from a same-origin response that Vite does not transform. An HTML
  // entrypoint receives Vite's HMR client and can reload several seconds after
  // `serve` regenerates contracts, destroying a fixture evaluation mid-test.
  await resetRendererFixturePage(page);
  return captureRendererDiagnostics(page);
}

export async function resetRendererFixturePage(page: Page): Promise<void> {
  await page.goto('/client_config.js');
  await page.evaluate(() => {
    document.open();
    document.write('<!doctype html><html lang="en"><head><meta charset="utf-8"><title>Renderer contract fixture</title></head><body></body></html>');
    document.close();
  });
}

export async function retryRendererEvaluation<T>(page: Page, operation: () => Promise<T>): Promise<T> {
  for (let attempt = 0; attempt < 2; attempt++) {
    try {
      return await operation();
    } catch (error) {
      if (attempt > 0 || !(error instanceof Error)
        || !error.message.includes('Execution context was destroyed')) throw error;
      // Importing a newly-added renderer dependency can make Vite optimize once
      // and reload every open page. Reset the inert fixture document and retry;
      // the dependency graph is stable on the second attempt.
      await resetRendererFixturePage(page);
    }
  }
  throw new Error('Renderer evaluation retry exhausted');
}

export async function focusWithKeyboard(
  page: Page,
  target: Locator,
  maximumTabs = 12,
): Promise<void> {
  await page.evaluate(() => {
    if (document.activeElement instanceof HTMLElement) document.activeElement.blur();
  });
  for (let index = 0; index < maximumTabs; index += 1) {
    await page.keyboard.press('Tab');
    if (await target.evaluate((element) => (
      element.matches(':focus') || (element.shadowRoot?.activeElement ?? null) !== null
    ))) {
      return;
    }
  }
  throw new Error(`Element was not keyboard reachable after ${maximumTabs} Tab presses`);
}
