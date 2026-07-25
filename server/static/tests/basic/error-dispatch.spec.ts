import { test, expect } from '@playwright/test';
import { createOfflineGame } from '../animations/helpers.js';

// Regression: a single game error used to fan out into a deep chain of
// RE-ENTRANT Redux dispatches, spamming "Maximum call stack size exceeded"
// on game pages.
//
// Mechanism: every component that watches state.game.error
// (boardgame-configure-game-properties -- 13+ instances live on a game page
// -- plus boardgame-player-roster and boardgame-move-form) dispatches a
// `show-error` DOM event from inside its stateChanged. boardgame-app's
// handler answered that event with a SYNCHRONOUS store.dispatch, i.e. a
// dispatch from inside a subscriber notification, and updateAndShowError is
// a thunk that fires two more actions -- so each watcher instance pushed two
// more full notification passes onto the stack while the first was still
// unwinding. Measured before the fix: dispatch depth 16-17 and 31-33
// notifications from ONE error, growing linearly with the number of watcher
// instances until the stack blew.
//
// Each watcher's own _lastError guard bounds it to one dispatch per distinct
// error, which is why the depth was large-but-finite rather than truly
// infinite -- it does nothing about the cross-instance amplification.
test('a single game error does not re-enter the dispatch cycle', async ({ page }) => {
  test.setTimeout(180_000);

  const pageErrors: string[] = [];
  page.on('pageerror', (e) => pageErrors.push(String(e).slice(0, 200)));

  await createOfflineGame(page, 'memory', { adminMode: false });
  await page.waitForTimeout(2000);

  const result = await page.evaluate(async () => {
    const { store } = await import('/src/store.ts');
    let depth = 0;
    let maxDepth = 0;
    let notifications = 0;
    const unsubscribe = store.subscribe(() => { notifications++; });
    const origDispatch = store.dispatch.bind(store);
    (store as unknown as { dispatch: unknown }).dispatch = (action: unknown) => {
      depth++;
      maxDepth = Math.max(maxDepth, depth);
      try {
        return origDispatch(action as never);
      } finally {
        depth--;
      }
    };
    // Exactly the shape a failed move produces.
    store.dispatch({
      type: 'SUBMIT_MOVE_FAILURE',
      error: 'recursion probe',
      friendlyError: 'recursion probe',
    } as never);
    await new Promise<void>((resolve) => setTimeout(resolve, 400));
    (store as unknown as { dispatch: unknown }).dispatch = origDispatch;
    unsubscribe();

    const state = store.getState() as { error?: { showing?: boolean } };
    return {
      maxDepth,
      notifications,
      errorShowing: state.error?.showing ?? false,
    };
  });

  // The error must still reach the dialog -- this fix breaks the re-entrancy,
  // it does not swallow errors.
  expect(result.errorShowing, 'the error must still be shown to the user').toBe(true);
  // One nesting level (the thunk's own inner dispatches) is inherent; the
  // watcher-driven cascade is not. Pre-fix this measured 16-17.
  expect(result.maxDepth,
    'a game error must not re-enter the dispatch cycle once per error watcher')
    .toBeLessThanOrEqual(3);
  expect(pageErrors.join(' ')).not.toContain('Maximum call stack size exceeded');
  // Every watcher answering the same error with its own dispatch is the other
  // half of the defect: flat, but ~38 app-wide re-render passes for one
  // error. Deduping identical content while the dialog is up cut this to 9.
  expect(result.notifications,
    'one error must not fan out into a dispatch per error watcher')
    .toBeLessThanOrEqual(12);
});

// Regression: after ANY error dialog was dismissed, the next error never
// displayed. boardgame-app synced dismissal state from md-dialog's `closed`
// event, which fires only after the close ANIMATION finishes — an error
// arriving in that window set showing=true and was immediately clobbered
// back to false by the late handler. Now synced from `close`, which fires
// synchronously at dismissal intent, leaving no window to lose an error in.
test('a different error still shows after one is dismissed', async ({ page }) => {
  test.setTimeout(180_000);
  await createOfflineGame(page, 'memory', { adminMode: false });
  await page.waitForTimeout(2000);

  const shown = await page.evaluate(async () => {
    const { store } = await import('/src/store.ts');
    const { hideError } = await import('/src/actions/error.ts');
    const showing = () => (store.getState() as { error?: { showing?: boolean } }).error?.showing ?? false;
    const fire = async (message: string) => {
      store.dispatch({
        type: 'SUBMIT_MOVE_FAILURE', error: message, friendlyError: message,
      } as never);
      await new Promise<void>((r) => setTimeout(r, 300));
    };
    await fire('first probe error');
    const afterFirst = showing();
    store.dispatch(hideError());
    await new Promise<void>((r) => setTimeout(r, 100));
    const afterDismiss = showing();
    await fire('a genuinely different probe error');
    return { afterFirst, afterDismiss, afterSecond: showing() };
  });

  expect(shown.afterFirst, 'the first error must show').toBe(true);
  expect(shown.afterDismiss, 'dismissing must hide the dialog').toBe(false);
  expect(shown.afterSecond, 'a different error must still show after a dismissal').toBe(true);
});
