import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';

test('axe detects the deliberately inaccessible control fixture', async ({ page }) => {
  await page.setContent(`
    <main>
      <button data-testid="unnamed-action"></button>
    </main>
  `);

  const result = await new AxeBuilder({ page })
    .withRules(['button-name'])
    .analyze();

  expect(result.violations.map(({ id }) => id)).toContain('button-name');
  const violation = result.violations.find(({ id }) => id === 'button-name');
  expect(violation?.nodes.some(({ html }) => (
    html.includes('data-testid="unnamed-action"')
  ))).toBe(true);
});
