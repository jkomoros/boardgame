// Instrumentation counters for the animation pipeline, consumed by the
// Playwright regression gate (tests/animations/waapi-*.spec.ts). Always
// installed: the cost is one array push per lifecycle event, and having
// it in production builds means bug reports can include the log.

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

declare global {
  interface Window { __bgAnimTestHooks: AnimHooks; }
}
window.__bgAnimTestHooks = animHooks;
