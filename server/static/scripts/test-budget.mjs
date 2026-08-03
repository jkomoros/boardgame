/**
 * A per-test wall-clock budget for `npm run test:unit`, registered globally by
 * `--import`.
 *
 * WHY THIS EXISTS RATHER THAN JUST `--test-timeout`. A mutation pass over
 * `src/motion` and `src/solid` found that 21% of all mutants did not make a
 * test *fail*, they made it *slow* -- and 210 of those were outright survivors:
 * the suite went green, just took an order of magnitude longer. The obvious fix
 * is a per-test timeout, and `--test-timeout` is set (see `test:unit`), but on
 * its own it would have been a check that cannot check:
 *
 *   `--test-timeout` is a timer. A timer cannot fire while the event loop is
 *   blocked, and essentially every test in this suite is synchronous,
 *   CPU-bound numerical code. Measured: with `--test-timeout=1000`, a test that
 *   busy-waits three seconds PASSES; only the `await`ing one fails.
 *
 * So the two mechanisms are complementary and both are wanted:
 *
 *   - `--test-timeout` catches an async test that never settles (a hook that
 *     never resolves, a promise that never rejects). The budget below cannot,
 *     because `afterEach` never runs for a test that never finishes.
 *   - This budget catches a synchronous test that finishes but takes far
 *     longer than it should -- the entire family the mutation pass found. It
 *     measures around each test and fails it by name, so the report says
 *     "this test got 12x slower", not "the job timed out".
 *
 * Neither can catch a synchronous infinite loop; nothing in-process can. The
 * CI job timeout is the backstop for that.
 */

import { beforeEach, afterEach } from 'node:test';

/**
 * The default budget. The slowest test in the suite that is not listed in
 * SLOW_TESTS runs in 2.6s on an idle machine, so this is roughly 4x headroom,
 * and it still turns a 10x slowdown of any test over a second into a named
 * failure. Raise it with UNIT_TEST_BUDGET_MS on a machine slow enough to need
 * it rather than by editing this constant -- the whole value of the number is
 * that it is tight.
 */
const DEFAULT_BUDGET_MS = 10_000;

/**
 * Tests that are legitimately slower than the default, with the budget each is
 * allowed and why. A test earns a line here by being measured, not by being
 * annoying: the point is that the exception is visible and costed, instead of
 * the ceiling being lifted for all 751 tests.
 *
 * Keyed by test name. A rename drops the test back to the default budget,
 * which fails loudly -- that is the intended direction to fail in.
 */
const SLOW_TESTS = new Map([
  // Throws twenty-five d20 into a tray they cannot settle in, eight times over
  // eight seeds, and asserts on the throw and step counts. ~18.5s idle; it is
  // the only test in the repo over three seconds.
  ['stops throwing a tray that is not getting any better', 60_000],
]);

const configured = Number.parseInt(process.env.UNIT_TEST_BUDGET_MS ?? '', 10);
const defaultBudget = Number.isFinite(configured) && configured > 0
  ? configured
  : DEFAULT_BUDGET_MS;

/**
 * Scale every budget, including the SLOW_TESTS entries, by this factor. Meant
 * for a machine under known load or an emulated architecture, where every test
 * is uniformly slower and lifting one number is the honest response.
 */
const scaleRaw = Number.parseFloat(process.env.UNIT_TEST_BUDGET_SCALE ?? '');
const scale = Number.isFinite(scaleRaw) && scaleRaw > 0 ? scaleRaw : 1;

let startedAt = 0;

beforeEach(() => {
  startedAt = performance.now();
});

afterEach((t) => {
  const elapsed = performance.now() - startedAt;
  const budget = (SLOW_TESTS.get(t.name) ?? defaultBudget) * scale;
  if (elapsed <= budget) return;
  throw new Error(
    `test "${t.name}" took ${elapsed.toFixed(0)}ms, over its ${budget.toFixed(0)}ms budget. `
    + 'A test that got much slower without failing is the shape a mutation pass found 210 of: '
    + 'either the code under it regressed, or the test genuinely needs a bigger budget -- '
    + 'in which case add it to SLOW_TESTS in scripts/test-budget.mjs with a measurement, '
    + 'or set UNIT_TEST_BUDGET_SCALE if the whole machine is slow.',
  );
});
