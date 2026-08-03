import test from 'node:test';
import assert from 'node:assert/strict';

/**
 * The premise check for the per-test budget in `scripts/test-budget.mjs`.
 *
 * That budget is the only thing standing between this suite and the family of
 * regressions a mutation pass found 210 of: code that gets an order of
 * magnitude slower without any test failing. It is installed by a single
 * `--import` in a single npm script, so removing it -- deliberately, or by
 * rewriting `test:unit` for some unrelated reason -- would silently disarm
 * every budget in the repo and nothing would go red.
 *
 * This asserts the module was actually loaded, not that the flag appears in
 * package.json. A flag can be present in the text and still not be in effect:
 * a typo'd path, a `--import` that failed to resolve, or a different script
 * being the one that actually runs would all leave the text intact.
 */
test('the per-test budget is installed', () => {
  assert.equal(
    (globalThis as Record<string, unknown>).__unitTestBudgetInstalled,
    true,
    'scripts/test-budget.mjs did not run. The per-test wall-clock budget is '
    + 'installed by `--import=./scripts/test-budget.mjs` in the `test:unit` '
    + 'script in package.json; without it every test in this suite may take '
    + 'arbitrarily long and still pass. Restore the flag, or if the budget was '
    + 'removed on purpose, delete this test in the same commit so the removal '
    + 'is visible in review.',
  );
});
