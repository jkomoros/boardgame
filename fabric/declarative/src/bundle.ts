/**
 * Entry point of the one-file build (`npm run build:bundle`): every `bg-`
 * element and the layer-1 elements they wrap, with Lit inside, for a page
 * that is not built with this repository's toolchain.
 *
 * Loading it registers the elements and does nothing else. The animator's
 * counters are published on `window` only where the page has opted in with
 * the same `localStorage` key the webapp reads.
 */
import './index.js';
import {
  ANIM_TEST_HOOKS_STORAGE_KEY, installAnimTestHooks,
} from '../../../server/static/src/utils/anim-test-hooks.js';

try {
  if (globalThis.localStorage?.getItem(ANIM_TEST_HOOKS_STORAGE_KEY) === '1') installAnimTestHooks();
} catch {
  // Storage can be denied; the counters then stay unpublished.
}

/** Tag names this bundle defines that a page author writes. */
export const BG_ELEMENTS = [
  'bg-table', 'bg-stack', 'bg-callout', 'bg-turn-status', 'bg-action', 'bg-seat', 'bg-value',
] as const;
