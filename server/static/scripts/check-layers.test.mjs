// The layer gate for server/static/src. See scripts/check-layers.mjs and
// scripts/layers.json, and fabric/PROPOSAL.md ("The layers").
//
// Runtime and type-only upward imports are reported by separate tests, because
// they cost different things: a runtime edge drags a higher layer's code into
// a lower layer's bundle, a type-only edge only blocks packaging the types
// separately. Both have an allow-list in layers.json, and both lists may only
// shrink: an entry that no longer matches a real import fails too.
import assert from 'node:assert/strict';
import test from 'node:test';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { analyze, readLayerTable } from './check-layers.mjs';

const table = readLayerTable();
const result = analyze(table);

function describe(violation) {
  return `${violation.key}  (${violation.fromLayer} -> ${violation.toLayer})`;
}

test('every non-test source file under src/ is assigned to a known layer', () => {
  assert.deepEqual(result.unassigned, [],
    'Add these files to scripts/layers.json (files under src/engine/ are covered by a rule)');
  assert.deepEqual(result.unknownLayers, []);
});

test('the layer table names no file that has been removed', () => {
  assert.deepEqual(result.stale, []);
});

test('every relative import resolves to a file', () => {
  assert.deepEqual(result.unresolved, []);
});

test('no runtime import goes from a lower layer to a higher one', () => {
  const unexpected = result.violations
    .filter(violation => !violation.typeOnly && table.allow.runtime[violation.key] === undefined)
    .map(describe);
  assert.deepEqual(unexpected, []);
});

test('no type-only import goes from a lower layer to a higher one', () => {
  const unexpected = result.violations
    .filter(violation => violation.typeOnly && table.allow.typeOnly[violation.key] === undefined)
    .map(describe);
  assert.deepEqual(unexpected, []);
});

test('the allow-lists hold only real, documented violations', () => {
  const runtime = new Set(result.violations.filter(violation => !violation.typeOnly).map(violation => violation.key));
  const typeOnly = new Set(result.violations.filter(violation => violation.typeOnly).map(violation => violation.key));
  const staleRuntime = Object.keys(table.allow.runtime).filter(key => !runtime.has(key));
  const staleTypeOnly = Object.keys(table.allow.typeOnly).filter(key => !typeOnly.has(key));
  assert.deepEqual(staleRuntime, [], 'Remove fixed entries from allow.runtime (a runtime edge that became type-only moves to allow.typeOnly)');
  assert.deepEqual(staleTypeOnly, [], 'Remove fixed entries from allow.typeOnly');
  for (const [key, reason] of [...Object.entries(table.allow.runtime), ...Object.entries(table.allow.typeOnly)]) {
    assert.ok(typeof reason === 'string' && reason.trim().length > 10, `${key} needs a reason`);
  }
});

test('boardgame-render-game imports nothing from the host layer at runtime', { skip: Object.keys(table.allow.runtime).some(key => key.startsWith('src/components/boardgame-render-game.ts ->')) && 'still allow-listed' }, () => {
  const edges = result.edges.filter(edge => edge.from === 'src/components/boardgame-render-game.ts'
    && edge.toLayer === 'host' && !edge.typeOnly);
  assert.deepEqual(edges.map(edge => edge.to), []);
});

// The checker itself, against a throwaway tree, so a parser regression cannot
// make the real gate pass by seeing nothing.
test('the checker sees static, type-only, re-export, dynamic and import-type edges', () => {
  const root = mkdtempSync(path.join(tmpdir(), 'layers-'));
  try {
    mkdirSync(path.join(root, 'src', 'engine'), { recursive: true });
    const write = (name, text) => writeFileSync(path.join(root, name), text);
    write('src/low.ts', [
      "import { a } from './high.js';",
      "import type { B } from './high-types.js';",
      "import { type C } from './named-type.js';",
      "export { d } from './reexport.js';",
      "export type { E } from './reexport-type.js';",
      "const f = () => import('./dynamic.js');",
      "let g: import('./import-type.js').G;",
      "import './engine/core.js';",
    ].join('\n'));
    for (const name of ['high', 'high-types', 'named-type', 'reexport', 'reexport-type', 'dynamic', 'import-type']) {
      write(`src/${name}.ts`, 'export {};\n');
    }
    write('src/engine/core.ts', "import '../low.js';\n");
    write('src/unlisted.ts', 'export {};\n');
    write('src/skipped.test.ts', "import './high.js';\n");
    const files = { 'src/low.ts': 'elements', 'src/gone.ts': 'host' };
    for (const name of ['high', 'high-types', 'named-type', 'reexport', 'reexport-type', 'dynamic', 'import-type']) {
      files[`src/${name}.ts`] = 'host';
    }
    const fixture = analyze({
      order: ['engine', 'elements', 'kit', 'host'],
      unranked: ['test-support'],
      rules: [{ prefix: 'src/engine/', layer: 'engine' }],
      files,
      allow: { runtime: {}, typeOnly: {} },
    }, root);
    assert.deepEqual(fixture.unassigned, ['src/unlisted.ts']);
    assert.deepEqual(fixture.stale, ['src/gone.ts']);
    const kinds = Object.fromEntries(fixture.violations.map(violation => [violation.key, violation.typeOnly]));
    assert.deepEqual(kinds, {
      'src/engine/core.ts -> src/low.ts': false,
      'src/low.ts -> src/dynamic.ts': false,
      'src/low.ts -> src/high-types.ts': true,
      'src/low.ts -> src/high.ts': false,
      'src/low.ts -> src/import-type.ts': true,
      'src/low.ts -> src/named-type.ts': true,
      'src/low.ts -> src/reexport-type.ts': true,
      'src/low.ts -> src/reexport.ts': false,
    });
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});
