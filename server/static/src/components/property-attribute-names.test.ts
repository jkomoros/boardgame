import assert from 'node:assert/strict';
import test from 'node:test';
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

/**
 * THE ATTRIBUTE-NAME TRAP, LINTED AS A CLASS.
 *
 * Lit derives a reactive property's observed attribute by LOWERCASING the
 * property name -- it does not dash-case it. So
 *
 *     @property({ type: Number }) fauxComponents = 0;
 *
 * observes `fauxcomponents`, and markup that writes the dashed
 * `faux-components="5"` -- the spelling every author reaches for, and the
 * spelling this codebase uses everywhere it declares an attribute by hand --
 * is a silent no-op. Nothing throws, nothing warns, the attribute sits in the
 * DOM, and the property keeps its default forever.
 *
 * That is not a hypothetical: `fauxComponents` and `noDefaultSpacer` on
 * `boardgame-component-stack`, `autoMessage` on `boardgame-fading-text`, and
 * the `?is-agent` / `?is-empty` / `?game-open` / `?game-visible` / `?is-owner`
 * bindings on the roster and lobby components were all dead for exactly this
 * reason, some of them for years.
 *
 * Instance fixes do not close this. The next multi-word property someone adds
 * without thinking about it is the next silent binding, and no runtime test
 * can see a bug whose symptom is "nothing happened". So the rule is
 * mechanical and enforced here: every multi-word reactive property in
 * `src/` must SAY what its attribute is -- either the dash-case name, or
 * `false` if it is property-only. Single-word names are exempt because
 * lowercasing them is already the dash-case answer.
 *
 * Scope is the framework's own components (`src/`). Game renderers in
 * `examples/` and the games repo set their properties from their own
 * templates as `.prop=` bindings and are not linted here; if one of them ever
 * declares a dashed attribute, this file is the pattern to copy.
 */

const HERE = dirname(fileURLToPath(import.meta.url));
const SRC_ROOT = join(HERE, '..');

const dashCase = (name: string): string =>
  name.replace(/([a-z0-9])([A-Z])/g, '$1-$2').toLowerCase();

const isMultiWord = (name: string): boolean => /[a-z0-9][A-Z]/.test(name);

interface Declaration {
  file: string;
  line: number;
  name: string;
  /** The decorator's options text, `''` when it was called with no arguments. */
  options: string;
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
 * Every `@property(...)` in `src/`, paired with the field it decorates.
 *
 * Deliberately a source scan rather than a runtime walk of
 * `elementProperties`: the components need a DOM to construct, this suite has
 * none, and the thing being checked is a DECLARATION, which is exactly what
 * the source says. It also means the failure names a file and a line, which
 * is what someone who just tripped the rule needs.
 */
function propertyDeclarations(): Declaration[] {
  const declarations: Declaration[] = [];
  for (const file of tsFiles(SRC_ROOT)) {
    const lines = readFileSync(file, 'utf8').split('\n');
    for (let i = 0; i < lines.length; i++) {
      const opened = lines[i].match(/^\s*@property\s*\((.*)$/);
      if (!opened) continue;

      // The decorator's arguments may wrap over several lines; balance parens.
      let options = opened[1];
      let cursor = i;
      const delta = (text: string) =>
        (text.match(/\(/g) || []).length - (text.match(/\)/g) || []).length;
      let depth = delta(options) + 1;
      while (depth > 0 && cursor + 1 < lines.length) {
        cursor++;
        options += '\n' + lines[cursor];
        depth += delta(lines[cursor]);
      }

      // The decorated field is whatever follows the closing paren: on the same
      // line for the one-liner form (`@property({...}) commitLabel = '';`),
      // otherwise the next line that is not blank, a comment, or another
      // decorator.
      const modifiers = '(?:(?:public|private|protected|override|declare|readonly|accessor|static)\\s+)*';
      const sameLine = lines[cursor].replace(/^[\s\S]*?\)/, '');
      let text = sameLine;
      if (!new RegExp(`^\\s*${modifiers}[A-Za-z_$]`).test(sameLine)) {
        let field = cursor + 1;
        while (field < lines.length
          && (lines[field].trim() === '' || /^\s*(\/\/|\/\*|\*|@)/.test(lines[field]))) {
          field++;
        }
        text = lines[field] ?? '';
      }
      const named = text.match(new RegExp(`^\\s*${modifiers}([A-Za-z_$][\\w$]*)`));
      if (!named) continue;

      declarations.push({
        file: relative(SRC_ROOT, file),
        line: i + 1,
        name: named[1],
        options: options.trim(),
      });
    }
  }
  return declarations;
}

const declaredAttribute = (options: string): string | false | null => {
  const explicit = options.match(/attribute\s*:\s*(?:'([^']*)'|"([^"]*)"|`([^`]*)`|(false|true))/);
  if (!explicit) return null;
  if (explicit[4] === 'false') return false;
  if (explicit[4] === 'true') return null; // `true` means "use the default" -- same trap.
  return explicit[1] ?? explicit[2] ?? explicit[3] ?? null;
};

test('the scan finds the reactive properties it is meant to lint', () => {
  const declarations = propertyDeclarations();
  // A premise guard. If a refactor moves the components or changes the
  // decorator spelling, this lint must go loud rather than quietly passing
  // over an empty list -- a vacuous lint is worse than none.
  assert.ok(
    declarations.length > 100,
    `expected the source scan to find the component properties, found ${declarations.length}`,
  );
  assert.ok(
    declarations.some((d) => d.name === 'fauxComponents'),
    'the scan must see boardgame-component-stack.fauxComponents',
  );
  assert.ok(
    declarations.filter((d) => isMultiWord(d.name)).length > 50,
    'the scan must see plenty of multi-word properties, or the rule below is vacuous',
  );
});

test('every multi-word reactive property declares its attribute name explicitly', () => {
  const offenders = propertyDeclarations()
    .filter((d) => isMultiWord(d.name))
    .filter((d) => declaredAttribute(d.options) === null)
    .map((d) => `${d.file}:${d.line}  ${d.name} would observe `
      + `"${d.name.toLowerCase()}", not "${dashCase(d.name)}"`);

  assert.deepEqual(
    offenders,
    [],
    'Lit lowercases a property name to derive its observed attribute, so a '
    + 'multi-word property silently observes a squashed spelling nobody '
    + 'writes. Declare it: `@property({ type: X, attribute: \'dash-case\' })`, '
    + 'or `attribute: false` if it is set as a property only.\n'
    + offenders.join('\n'),
  );
});

test('an explicitly declared attribute is the dash-case of its property', () => {
  const wrong = propertyDeclarations()
    .map((d) => ({ d, attribute: declaredAttribute(d.options) }))
    .filter(({ attribute }) => typeof attribute === 'string')
    .filter(({ d, attribute }) => attribute !== dashCase(d.name))
    .map(({ d, attribute }) =>
      `${d.file}:${d.line}  ${d.name} declares "${attribute}", expected "${dashCase(d.name)}"`);

  assert.deepEqual(
    wrong,
    [],
    'Declaring the attribute is only half the job -- declaring the WRONG name '
    + 'reproduces the bug with extra steps. Every attribute in this codebase '
    + 'is the dash-case of its property; keep it that way.\n' + wrong.join('\n'),
  );
});

test('private (underscore-prefixed) reactive properties are not attribute-observable', () => {
  // `_first-state-bundle` is not an attribute anyone would write, and internal
  // reactive state has no business being settable from markup. Its neighbours
  // in boardgame-game-view all say `attribute: false`; this makes that the
  // rule rather than a habit.
  const leaked = propertyDeclarations()
    .filter((d) => d.name.startsWith('_'))
    .filter((d) => declaredAttribute(d.options) !== false)
    .map((d) => `${d.file}:${d.line}  ${d.name}`);

  assert.deepEqual(leaked, [],
    'internal reactive state must declare `attribute: false`\n' + leaked.join('\n'));
});
