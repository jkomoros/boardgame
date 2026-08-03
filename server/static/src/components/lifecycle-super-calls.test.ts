import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, statSync, readFileSync } from 'node:fs';
import { join, relative, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

/**
 * Lit's `connectedCallback` and `disconnectedCallback` are not hooks the base
 * class merely offers -- they are where `ReactiveElement` attaches and detaches
 * the update lifecycle and every reactive controller. An override that forgets
 * `super` does not throw, does not warn, and does not render: the element
 * simply never performs its first update, and the failure surfaces somewhere
 * else entirely, as an empty component.
 *
 * `boardgame-animatable-item.ts` says so in prose on both of its overrides
 * ("Always call super first"), and prose about a required invariant is not
 * enforcement. This lints it instead.
 *
 * # What is enforced, and what deliberately is not
 *
 * `connectedCallback` must super-call FIRST. Everything the codebase does after
 * connecting -- ambient-registry discovery, resize observers, store
 * subscriptions -- assumes the base class has already run, and all 20-odd
 * overrides in `src/` already comply.
 *
 * `disconnectedCallback` must super-call, but this does NOT pin where. The
 * animatable-item comment used to claim super belongs last there, "mirroring
 * other overrides in this codebase"; a scan showed the opposite -- ten of the
 * thirteen overrides call it first. Release-before-yield versus yield-first is
 * a real per-component judgement, and a lint that picked a side would be
 * inventing a convention rather than enforcing one. Omitting super entirely is
 * the actual bug, and that is what this catches.
 *
 * A source scan rather than a runtime walk, for the same reasons
 * property-attribute-names.test.ts gives: these components need a DOM, this
 * suite has none, and what is being checked is a declaration, so the failure
 * can name a file and a line.
 */

const HERE = dirname(fileURLToPath(import.meta.url));
const SRC_ROOT = join(HERE, '..');

type CallbackName = 'connectedCallback' | 'disconnectedCallback';

interface Override {
  file: string;
  line: number;
  name: CallbackName;
  /** Body statements, comments and blank lines removed, trimmed. */
  statements: string[];
}

function tsFiles(dir: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir).sort()) {
    const path = join(dir, entry);
    if (statSync(path).isDirectory()) {
      tsFiles(path, out);
    } else if (path.endsWith('.ts') && !path.endsWith('.d.ts')
      && !path.endsWith('.test.ts') && !path.endsWith('.compile.ts')) {
      out.push(path);
    }
  }
  return out;
}

/**
 * Every `connectedCallback`/`disconnectedCallback` method body in `src/`.
 *
 * The body runs from the opening line to the first line whose closing brace is
 * at the method's own indentation, which is what a formatted class method looks
 * like. Declarations in `.d.ts` files are excluded by tsFiles, so an ambient
 * signature with no body cannot masquerade as an empty override.
 */
function lifecycleOverrides(): Override[] {
  const overrides: Override[] = [];

  for (const file of tsFiles(SRC_ROOT)) {
    const lines = readFileSync(file, 'utf8').split('\n');
    for (let i = 0; i < lines.length; i++) {
      const opened = lines[i].match(
        /^(\s*)(?:override\s+)?(connectedCallback|disconnectedCallback)\s*\(\s*\)\s*(?::\s*void\s*)?\{/,
      );
      if (!opened) continue;

      const indent = opened[1];
      const closing = new RegExp(`^${indent}\\}`);
      const body: string[] = [];
      let cursor = i + 1;
      while (cursor < lines.length && !closing.test(lines[cursor])) {
        body.push(lines[cursor]);
        cursor++;
      }

      overrides.push({
        file: relative(SRC_ROOT, file),
        line: i + 1,
        name: opened[2] as CallbackName,
        statements: body
          .filter((l) => l.trim() !== '' && !/^\s*(\/\/|\/\*|\*)/.test(l))
          .map((l) => l.trim()),
      });
    }
  }

  return overrides;
}

const callsSuper = (statement: string, name: CallbackName): boolean =>
  statement.startsWith(`super.${name}(`);

test('the scan finds the lifecycle overrides it is meant to lint', () => {
  const overrides = lifecycleOverrides();

  // A premise guard. If a refactor renames the components or reformats the
  // method signatures out from under this regex, the lint must go loud rather
  // than pass over an empty list -- a vacuous lint is worse than none.
  assert.ok(
    overrides.length > 20,
    `expected the source scan to find the lifecycle overrides, found ${overrides.length}`,
  );
  assert.ok(
    overrides.some((o) => o.file === 'components/boardgame-animatable-item.ts'
      && o.name === 'connectedCallback'),
    'the scan must see boardgame-animatable-item\'s connectedCallback, the override whose doc comment this lint replaces',
  );
  assert.ok(
    overrides.filter((o) => o.name === 'disconnectedCallback').length > 8,
    'the scan must see both callbacks, or half the rule below is vacuous',
  );
  assert.ok(
    overrides.every((o) => o.statements.length > 0),
    'no override body should scan as empty; an empty body means the brace matching is wrong, not that the method does nothing',
  );
});

test('every connectedCallback override super-calls first', () => {
  const offenders = lifecycleOverrides()
    .filter((o) => o.name === 'connectedCallback')
    .filter((o) => !callsSuper(o.statements[0] ?? '', 'connectedCallback'));

  assert.deepEqual(
    offenders.map((o) => `${o.file}:${o.line}`),
    [],
    'a connectedCallback override must open with `super.connectedCallback();`. '
    + 'Lit performs the element\'s first update there, so an override that skips '
    + 'it -- or does work before it -- gets an element that never renders, with '
    + 'no error anywhere to say why.',
  );
});

test('every disconnectedCallback override super-calls somewhere', () => {
  const offenders = lifecycleOverrides()
    .filter((o) => o.name === 'disconnectedCallback')
    .filter((o) => !o.statements.some((s) => callsSuper(s, 'disconnectedCallback')));

  assert.deepEqual(
    offenders.map((o) => `${o.file}:${o.line}`),
    [],
    'a disconnectedCallback override must call `super.disconnectedCallback()`. '
    + 'That is where Lit detaches the update lifecycle and every reactive '
    + 'controller; skipping it leaks them. Where in the body it goes is your '
    + 'call -- this codebase does both, and neither is wrong.',
  );
});
