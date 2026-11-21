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
import { clientPackages } from './packages.mjs';

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

// The package boundary. Layers 1 and 2 are packages (packages/elements,
// packages/renderer-kit). Whatever is outside a package reaches it only by
// its name and only through a subpath its package.json exports; a package
// reaches another only if it declares the dependency; no relative import
// crosses a package edge in either direction; and a package holds exactly the
// files of its layer.
test('every cross-package import goes through declared dependencies and explicit exports', () => {
  assert.deepEqual(result.boundary, []);
});

test('the client packages exist and export something', () => {
  const packages = clientPackages();
  assert.deepEqual(packages.map(pkg => pkg.name), ['@boardgame/elements', '@boardgame/renderer-kit']);
  for (const pkg of packages) assert.ok(pkg.exports.size > 0, `${pkg.name} exports nothing`);
  const kit = packages[1];
  assert.deepEqual(kit.dependencies.filter(name => name.startsWith('@boardgame/')), ['@boardgame/elements']);
  assert.deepEqual(packages[0].dependencies.filter(name => name.startsWith('@boardgame/')), [],
    'the elements package depends on no other client package');
  assert.ok(kit.exports.has('@boardgame/renderer-kit/client.js'));
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

test('boardgame-render-game imports nothing from the host layer at runtime', () => {
  assert.ok(result.edges.some(edge => edge.from === 'packages/renderer-kit/src/components/boardgame-render-game.ts'),
    'the scan did not find boardgame-render-game, so this check is blind');
  const edges = result.edges.filter(edge => edge.from === 'packages/renderer-kit/src/components/boardgame-render-game.ts'
    && edge.toLayer === 'host' && !edge.typeOnly);
  assert.deepEqual(edges.map(edge => edge.to), []);
});

test('the checker reports each kind of package-boundary breach', () => {
  const root = mkdtempSync(path.join(tmpdir(), 'layers-packages-'));
  try {
    const write = (name, text) => {
      mkdirSync(path.dirname(path.join(root, name)), { recursive: true });
      writeFileSync(path.join(root, name), text);
    };
    write('packages/low/package.json', JSON.stringify({
      name: '@fixture/low',
      exports: { './public.js': './src/public.ts', './types.js': { types: './src/types.d.ts' }, './gone.js': './src/gone.ts' },
    }));
    write('packages/low/src/public.ts', "import './private.js';\nimport '../../../src/host.js';\nimport '@fixture/high/thing.js';\n");
    write('packages/low/src/private.ts', 'export {};\n');
    write('packages/low/src/types.d.ts', 'export type T = number;\n');
    write('packages/high/package.json', JSON.stringify({
      name: '@fixture/high', dependencies: { '@fixture/low': 'file:../low' }, exports: { './thing.js': './src/thing.ts' },
    }));
    write('packages/high/src/thing.ts', "import '@fixture/low/public.js';\nimport '@fixture/low/private.js';\nimport '@fixture/low/types.js';\nimport '../../low/src/public.js';\n");
    write('src/host.ts', "import '@fixture/low/public.js';\nimport type { T } from '@fixture/low/types.js';\nimport '../packages/high/src/thing.js';\n");
    write('src/stray.ts', 'export {};\n');
    write('src/fixture.compile.ts', "import '@fixture/low/private.js';\n");
    const fixture = analyze({
      order: ['low', 'high', 'host'],
      unranked: ['test-support'],
      packages: { 'packages/low': 'low', 'packages/high': 'high' },
      rules: [{ prefix: 'src/', suffix: '.compile.ts', layer: 'test-support' }],
      files: {
        'packages/low/src/public.ts': 'low',
        'packages/low/src/private.ts': 'high',
        'packages/high/src/thing.ts': 'high',
        'src/host.ts': 'host',
        'src/stray.ts': 'low',
      },
      allow: { runtime: {}, typeOnly: {} },
    }, root);
    assert.deepEqual(fixture.unresolved, []);
    assert.deepEqual(fixture.boundary, [
      "@fixture/low: export \"@fixture/low/gone.js\" points at a file that does not exist",
      "packages/high/src/thing.ts -> ../../low/src/public.js: a relative import crosses a package boundary; use @fixture/low",
      "packages/high/src/thing.ts -> @fixture/low/private.js: not in @fixture/low's exports",
      "packages/high/src/thing.ts -> @fixture/low/types.js: a types-only export imported at runtime",
      "packages/low/src/private.ts: is in @fixture/low (layer low) but the table says high",
      "packages/low/src/public.ts -> ../../../src/host.js: a relative import crosses a package boundary; use no import at all",
      "packages/low/src/public.ts -> @fixture/high/thing.js: @fixture/low does not declare @fixture/high as a dependency",
      "src/host.ts -> ../packages/high/src/thing.js: a relative import crosses a package boundary; use @fixture/high",
      "src/stray.ts: layer low lives in packages/low, not in src/",
    ]);
    // The upward imports are still reported as layer violations too.
    assert.ok(fixture.violations.some(violation => violation.key === 'packages/low/src/public.ts -> src/host.ts'));
    assert.ok(fixture.violations.some(violation => violation.key === 'packages/low/src/public.ts -> packages/high/src/thing.ts'));
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
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
