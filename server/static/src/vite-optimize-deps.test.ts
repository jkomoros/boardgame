import test from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

/**
 * `vite.config.ts`'s `optimizeDeps.include` must list every bare specifier the
 * CLIENT imports, because Vite's initial crawl follows static imports from
 * `index.html` and game renderers are discovered dynamically. Anything the
 * crawl misses is optimized on first game load, and Vite answers that with a
 * full page reload -- which, mid-test, erases the animation counters the parity
 * suite is collecting, and mid-game loses whatever the player was doing.
 *
 * The list was hand-maintained, along with a hand-written grep for regenerating
 * it that was hardcoded to four package prefixes. Deriving the set mechanically
 * immediately found three specifiers both had missed: `reselect`,
 * `reselect-tools/src` and `firebase/compat/app`.
 */

/**
 * Specifiers that belong in `optimizeDeps.include` even though no client file
 * imports them by name -- a dependency Vite only resolves transitively, and so
 * only discovers once a page is already open, which is exactly the late
 * discovery that triggers the reload.
 *
 * Empty today, and it should stay a short list with a reason per line: an entry
 * here is a claim that the scan below cannot see something real, which is also
 * what a stale entry looks like. Anything added should say what pulls it in.
 */
const TRANSITIVE_EXTRAS = new Set<string>([]);

const here = dirname(fileURLToPath(import.meta.url));
const staticRoot = resolve(here, '..');
const repoRoot = resolve(staticRoot, '../..');

/**
 * Only files that become part of the browser bundle count. Test files import
 * `node:*` and `typescript`, which never reach a browser and must not be handed
 * to Vite to pre-bundle. `.d.ts` files declare types and emit no runtime import.
 */
const isClientSource = (path: string) =>
  (path.endsWith('.ts') || path.endsWith('.js'))
  && !path.endsWith('.d.ts')
  && !path.endsWith('.test.ts')
  && !path.endsWith('.test.js');

function sourceFiles(root: string): string[] {
  let found: string[] = [];
  let entries: string[];
  try {
    entries = readdirSync(root);
  } catch {
    return [];
  }
  for (const entry of entries) {
    if (entry === 'node_modules' || entry === 'dist') continue;
    const path = join(root, entry);
    if (statSync(path).isDirectory()) {
      found = found.concat(sourceFiles(path));
    } else if (isClientSource(path)) {
      found.push(path);
    }
  }
  return found;
}

/** Client roots, matching what the config's comment claims to cover. */
function clientRoots(): string[] {
  const roots = [join(staticRoot, 'src')];
  const examples = join(repoRoot, 'examples');
  let games: string[] = [];
  try {
    games = readdirSync(examples);
  } catch {
    return roots;
  }
  for (const game of games) {
    const client = join(examples, game, 'client');
    try {
      if (statSync(client).isDirectory()) roots.push(client);
    } catch {
      // A game without a client directory is fine.
    }
  }
  return roots;
}

/**
 * Bare specifiers: not relative, not absolute, not a node builtin.
 *
 * Both patterns are anchored to the start of a line, which is where every
 * module-level import sits and where prose never does. An unanchored
 * `(?:from|import)\s*['"]([^'"]+)['"]` matches inside comments -- the
 * apostrophe in a word like "don't" happily pairs with a quote further down the
 * file -- and the first run of this test duly reported a sentence fragment as a
 * missing dependency.
 */
function bareSpecifiers(files: string[]): Set<string> {
  const found = new Set<string>();
  const patterns = [
    // import x from 'y'  /  export { a } from 'y'. The gap between the keyword
    // and `from` deliberately allows newlines: a braced import list spanning
    // several lines is the normal style here, and requiring one line missed
    // `lit/async-directive.js` and `reselect-tools/src` -- the first of which
    // is the very specifier whose late discovery caused the reload this whole
    // config block exists to prevent.
    /^[ \t]*(?:import|export)\b[^'"]*?\bfrom\s*['"]([^'"\n]+)['"]/gm,
    // import 'y'  (side effect only -- how the @material/web components load,
    // and invisible to any search for `from`)
    /^[ \t]*import\s*['"]([^'"\n]+)['"]/gm,
  ];
  for (const file of files) {
    const text = readFileSync(file, 'utf8');
    for (const pattern of patterns) {
      for (const match of text.matchAll(pattern)) {
        const specifier = match[1];
        if (specifier.startsWith('.') || specifier.startsWith('/')) continue;
        if (specifier.startsWith('node:')) continue;
        found.add(specifier);
      }
    }
  }
  return found;
}

function configuredIncludes(): string[] {
  const config = readFileSync(join(staticRoot, 'vite.config.ts'), 'utf8');
  const block = /optimizeDeps:\s*\{\s*include:\s*\[([\s\S]*?)\]/.exec(config);
  assert.ok(block, 'could not find optimizeDeps.include in vite.config.ts -- this test is blind');
  return [...block[1].matchAll(/['"]([^'"]+)['"]/g)].map((m) => m[1]);
}

test('optimizeDeps.include lists every bare specifier the client imports', () => {
  const roots = clientRoots();
  //Premise guard: if the scan finds nothing, it is broken, and an empty set
  //trivially satisfies every assertion below.
  assert.ok(roots.length > 1,
    `found only ${roots.length} client root(s) -- expected src/ plus examples/*/client/`);
  const files = roots.flatMap((root) => sourceFiles(root));
  assert.ok(files.length > 50, `found only ${files.length} client source files -- the scan is broken`);

  const imported = bareSpecifiers(files);
  assert.ok(imported.has('lit'), 'the scan did not even find lit -- it is not reading imports');

  const configured = new Set(configuredIncludes());

  const missing = [...imported].filter((s) => !configured.has(s)).sort();
  assert.deepEqual(missing, [],
    `these bare specifiers are imported by client code but not pre-bundled, so Vite will `
    + `discover them on first game load and reload the page: ${missing.join(', ')}. `
    + 'Add them to optimizeDeps.include in vite.config.ts.');

  const stale = [...configured]
    .filter((s) => !imported.has(s) && !TRANSITIVE_EXTRAS.has(s)).sort();
  assert.deepEqual(stale, [],
    `these are pre-bundled but no client code imports them any more: ${stale.join(', ')}. `
    + 'Remove them from optimizeDeps.include so the list keeps saying something true.');
});
