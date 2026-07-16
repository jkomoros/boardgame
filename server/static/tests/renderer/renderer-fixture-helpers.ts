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
  await page.goto('/');
  await page.evaluate(() => document.body.replaceChildren());
  return captureRendererDiagnostics(page);
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
