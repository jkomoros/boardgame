// Instrumentation counters for the animation pipeline, consumed by the
// Playwright regression gate (tests/animations/waapi-*.spec.ts). The counters
// always record: the cost is one array push per lifecycle event, and having
// them in production builds means bug reports can include the log.
//
// Publishing them on `window` is opt-in. This module is imported by layer-1
// elements, which must not write globals as an import side effect, so the
// renderer host decides: it calls `installAnimTestHooks()` when it wants the
// counters reachable from outside the module graph. The webapp does so when
// `localStorage[ANIM_TEST_HOOKS_STORAGE_KEY] === '1'` (see boardgame-app.ts);
// the Playwright config seeds that key.

export interface AnimHookEntry {
  t: number;
  ev: string;
  detail?: string;
  version?: number;
  targetAtMs?: number;
}

class AnimHooks {
  gateOpens = 0;
  gateCloses = 0;
  watchdogFirings = 0;
  plays = 0;
  settles = 0;
  log: AnimHookEntry[] = [];

  record(
    ev: 'gate-open' | 'gate-close' | 'watchdog' | 'play' | 'active' | 'install' | 'settle',
    detail?: string,
    timing?: { version?: number; targetAtMs?: number },
  ) {
    switch (ev) {
      case 'gate-open': this.gateOpens++; break;
      case 'gate-close': this.gateCloses++; break;
      case 'watchdog': this.watchdogFirings++; break;
      case 'play': this.plays++; break;
      case 'settle': this.settles++; break;
    }
    this.log.push({ t: performance.now(), ev, detail, ...timing });
    if (this.log.length > 5000) this.log.splice(0, 1000);
  }

  reset() {
    this.gateOpens = 0;
    this.gateCloses = 0;
    this.watchdogFirings = 0;
    this.plays = 0;
    this.settles = 0;
    this.log = [];
  }
}

export const animHooks = new AnimHooks();

/** The localStorage key the webapp host reads to decide whether to opt in. */
export const ANIM_TEST_HOOKS_STORAGE_KEY = 'boardgame:anim-test-hooks';

declare global {
  interface Window { __bgAnimTestHooks?: AnimHooks; }
}

/**
 * Publishes the counters as `target.__bgAnimTestHooks`. Idempotent. Nothing
 * calls this at import; a renderer host opts in.
 */
export function installAnimTestHooks(target: Window = window): AnimHooks {
  target.__bgAnimTestHooks = animHooks;
  return animHooks;
}
