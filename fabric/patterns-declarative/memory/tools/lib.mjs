// Shared by the scripts here: a browser identity on the shell's login screen,
// and reads of the board that cross the shell's shadow roots.
import { execFileSync } from 'node:child_process';
import { dirname, resolve } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const playwrightPath = process.env.PLAYWRIGHT ??
  resolve(here, '../../../declarative/node_modules/@playwright/test/index.mjs');
export const { chromium } = await import(pathToFileURL(playwrightPath).href);

export const sleep = ms => new Promise(done => setTimeout(done, ms));

export function requireEnv() {
  for (const name of ['CF_LABS_ROOT', 'CF_IDENTITY', 'CF_SPACE', 'CF_API_URL']) {
    if (!process.env[name]) {
      console.error(`${name} is not set`);
      process.exit(2);
    }
  }
}

export function grant(did) {
  execFileSync(resolve(process.env.CF_LABS_ROOT, 'bin/cf'), ['acl', 'set', did, 'WRITE'], {
    env: process.env,
    stdio: ['ignore', 'ignore', 'inherit'],
  });
}

export async function waitFor(check, what, timeout = 30_000) {
  const started = Date.now();
  for (;;) {
    const value = await check();
    if (value) return value;
    if (Date.now() - started > timeout) throw new Error(`timed out waiting for ${what}`);
    await sleep(40);
  }
}

/** Registers a new identity on the shell's login screen; returns its DID. */
export async function register(page, url) {
  await page.goto(url);
  await page.getByText('Register', { exact: false }).first().click();
  await page.getByText('Use Passphrase', { exact: false }).click();
  await page.getByText('Generate Passphrase', { exact: false }).click();
  await page.getByText("I've Saved It", { exact: false }).click();
  await page.waitForFunction(() => {
    try {
      return typeof globalThis.app?.state?.().identity?.did?.() === 'string';
    } catch {
      return false;
    }
  }, null, { timeout: 30_000 });
  return page.evaluate(() => globalThis.app.state().identity.did());
}

/**
 * Installed in every page: `__deep(selector)` finds elements through every
 * open shadow root, and the animator's counters are opted into.
 */
export const INIT = () => {
  try { localStorage.setItem('boardgame:anim-test-hooks', '1'); } catch { /* denied */ }
  globalThis.__deep = (selector, root = document) => {
    const found = [];
    const walk = node => {
      for (const child of node.querySelectorAll('*')) {
        if (child.matches(selector)) found.push(child);
        if (child.shadowRoot) walk(child.shadowRoot);
      }
    };
    walk(root);
    return found;
  };
  // What the reconciler does to the elements, recorded by the elements'
  // own lifecycle: how often each was connected (a recreated element is a new
  // connection under the same key), and which property assignments landed
  // before the same microtask checkpoint (the table coalesces exactly those).
  const trace = globalThis.__bgTrace = { connects: {}, sets: [] };
  let checkpoint = 0, open = false;
  const mark = () => {
    if (!open) {
      open = true;
      checkpoint++;
      queueMicrotask(() => { open = false; });
    }
    return checkpoint;
  };
  // Lifecycle callbacks are read off the class when it is defined, so the
  // class is wrapped on its way into the registry.
  const watched = new Map();
  const define = customElements.define.bind(customElements);
  customElements.define = (tag, constructor, options) => {
    watched.get(tag)?.(constructor.prototype);
    return define(tag, constructor, options);
  };
  const watch = (tag, props) => watched.set(tag, proto => {
    const connected = proto.connectedCallback;
    proto.connectedCallback = function () {
      connected.call(this);
      const key = `${tag}:${this.key ?? this.name ?? ''}`;
      trace.connects[key] = (trace.connects[key] ?? 0) + 1;
    };
    for (const prop of props) {
      const descriptor = Object.getOwnPropertyDescriptor(proto, prop);
      if (!descriptor?.set) continue;
      Object.defineProperty(proto, prop, {
        ...descriptor,
        set(value) {
          trace.sets.push({
            checkpoint: mark(), tag, prop, key: this.key ?? this.name ?? '',
            version: prop === 'tableView' ? value?.version : prop === 'version' ? value : undefined,
            at: Math.round(performance.now()),
          });
          descriptor.set.call(this, value);
        },
      });
    }
  });
  watch('bg-table', ['tableView', 'principal']);
  watch('bg-stack', ['stack', 'version']);
  // Where every card is, frame by frame, between `start()` and `stop()`:
  // the box the browser reports, which includes the animator's transform.
  // Times are milliseconds since the epoch, to line up with screencast frames.
  let sampling = false, samples = [];
  const sample = () => {
    if (!sampling) return;
    const at = performance.timeOrigin + performance.now();
    for (const card of globalThis.__deep('boardgame-card')) {
      if (!card.id || card.hasAttribute('spacer')) continue;
      const box = card.getBoundingClientRect();
      samples.push([card.id, at, Math.round(box.left + box.width / 2), Math.round(box.top + box.height / 2)]);
    }
    requestAnimationFrame(sample);
  };
  globalThis.__flight = {
    start() { samples = []; sampling = true; requestAnimationFrame(sample); },
    stop() { sampling = false; return samples; },
  };
};

export async function player(browser, name, url) {
  const context = await browser.newContext({ viewport: { width: 900, height: 820 } });
  await context.addInitScript(INIT);
  const page = await context.newPage();
  const problems = [];
  page.on('console', message => {
    if (message.type() === 'error' || /policy|ceiling/i.test(message.text())) {
      problems.push(message.text().slice(0, 400));
    }
  });
  page.on('pageerror', error => problems.push(String(error).slice(0, 400)));
  const did = await register(page, url);
  grant(did);
  await page.reload();
  await waitFor(() => page.evaluate(() => globalThis.__deep('bg-table').length > 0),
    `${name}'s table element`, 60_000);
  return { name, page, did, problems };
}

/** What one player's table element shows, read from the elements themselves. */
export const tableOf = who => who.page.evaluate(() => {
  const table = globalThis.__deep('bg-table')[0];
  if (!table) return null;
  const stacks = globalThis.__deep('bg-stack', table);
  const grid = stacks.find(stack => stack.key === 'Game.Cards');
  const cards = grid ? [...grid.renderRoot.querySelectorAll('boardgame-card')] : [];
  const status = globalThis.__deep('boardgame-turn-status', table)[0];
  return {
    defined: !!customElements.get('bg-table'),
    upgraded: typeof table.whenSettled === 'function',
    phase: table.getAttribute('phase'),
    viewer: table.viewer,
    version: table.view?.Version ?? null,
    turn: table.view?.CurrentPlayerIndex ?? null,
    queued: table.tableView?.version ?? null,
    seats: table.seatLabels ?? [],
    animating: table.animating,
    stats: { ...table.stats },
    status: (status?.shadowRoot?.textContent ?? '').trim(),
    finished: !!table.view?.Finished,
    winners: table.view?.Winners ?? [],
    board: cards.map(card => ({
      face: (card.innerText || '').trim(),
      present: !card.hasAttribute('spacer'),
      enabled: card.getAttribute('aria-disabled') === 'false',
      id: card.id,
    })),
    won: stacks.filter(stack => /WonCards/.test(stack.key)).map(stack =>
      stack.renderRoot.querySelectorAll('boardgame-card:not([spacer])').length),
    blocked: globalThis.__deep('cf-cfc-blocked').length,
  };
});

export const traceOf = who => who.page.evaluate(() => {
  const trace = globalThis.__bgTrace;
  // Assignments per checkpoint: more than one stack in a checkpoint means the
  // table saw them as one change.
  const groups = new Map();
  for (const set of trace.sets) {
    if (set.prop !== 'stack') continue;
    groups.set(set.checkpoint, (groups.get(set.checkpoint) ?? 0) + 1);
  }
  const sizes = {};
  for (const size of groups.values()) sizes[size] = (sizes[size] ?? 0) + 1;
  return {
    connects: trace.connects,
    tableViewSets: trace.sets.filter(set => set.prop === 'tableView').length,
    stackSets: trace.sets.filter(set => set.prop === 'stack').length,
    stackSetsPerCheckpoint: sizes,
    // [stack, milliseconds since page load] for the last assignments.
    lastStackSets: trace.sets.filter(set => set.prop === 'stack').slice(-8).map(set => [set.key, set.at]),
  };
});

export const hooksOf = who => who.page.evaluate(() => {
  const hooks = globalThis.__bgAnimTestHooks;
  return hooks ? { plays: hooks.plays, settles: hooks.settles, watchdogFirings: hooks.watchdogFirings } : null;
});

/** Where one grid slot's card is on the page, for a real click. */
export const cardPoint = (who, slot) => who.page.evaluate(index => {
  const table = globalThis.__deep('bg-table')[0];
  const grid = globalThis.__deep('bg-stack', table).find(stack => stack.key === 'Game.Cards');
  const card = grid?.renderRoot.querySelectorAll('boardgame-card')[index];
  if (!card) return null;
  card.scrollIntoView({ block: 'center' });
  const box = card.getBoundingClientRect();
  return { x: box.left + box.width / 2, y: box.top + box.height / 2 };
}, slot);

/** Web Animations running on cards right now. */
export const runningOf = who => who.page.evaluate(() => {
  // The document's list reaches into shadow trees, where a card's flip runs.
  return document.getAnimations().filter(animation => {
    if (animation.playState !== 'running') return false;
    for (let node = animation.effect?.target; node; node = node.parentNode ?? node.host) {
      if (node.localName === 'boardgame-card') return true;
    }
    return false;
  }).length;
});

/**
 * Records one player's page from now until `stop()`: the browser's own
 * screencast frames, and where each card was on every animation frame.
 * `stop()` returns, for each card that travelled more than 40px, its path and
 * the screencast frames taken while it was between a quarter and three
 * quarters of the way: a card seen in flight, not counted.
 */
export async function recordFlight(who) {
  const cdp = await who.page.context().newCDPSession(who.page);
  const frames = [];
  cdp.on('Page.screencastFrame', frame => {
    frames.push({ at: frame.metadata.timestamp * 1000, data: frame.data });
    cdp.send('Page.screencastFrameAck', { sessionId: frame.sessionId }).catch(() => {});
  });
  await cdp.send('Page.startScreencast', { format: 'png', everyNthFrame: 1 });
  await who.page.evaluate(() => globalThis.__flight.start());
  return {
    async stop() {
      const samples = await who.page.evaluate(() => globalThis.__flight.stop());
      await cdp.send('Page.stopScreencast').catch(() => {});
      await cdp.detach().catch(() => {});
      const paths = new Map();
      for (const [id, at, x, y] of samples) {
        if (!paths.has(id)) paths.set(id, []);
        paths.get(id).push({ at, x, y });
      }
      const flights = [];
      for (const [id, path] of paths) {
        const from = path[0], to = path[path.length - 1];
        const distance = Math.hypot(to.x - from.x, to.y - from.y);
        if (distance < 40) continue;
        const progress = point => Math.hypot(point.x - from.x, point.y - from.y) / distance;
        const between = path.filter(point => progress(point) > 0.02 && progress(point) < 0.98);
        const middle = path.filter(point => progress(point) > 0.25 && progress(point) < 0.75);
        flights.push({
          id, from: [from.x, from.y], to: [to.x, to.y], distancePx: Math.round(distance),
          framesInFlight: between.length,
          flightMs: between.length ? Math.round(between[between.length - 1].at - between[0].at) : 0,
          path: between.map(point => [Math.round(point.at - between[0].at), point.x, point.y]),
          midFrames: middle.length
            ? frames.filter(frame => frame.at >= middle[0].at && frame.at <= middle[middle.length - 1].at)
            : [],
        });
      }
      return { screencastFrames: frames.length, animationFrames: new Set(samples.map(s => s[1])).size, flights };
    },
  };
}
