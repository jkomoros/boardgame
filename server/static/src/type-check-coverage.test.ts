// Guards the premise that every test file is actually type-checked.
//
// tsconfig.json excludes `src/**/*.test.ts` because it drives the
// declaration/outDir build and .d.ts files for tests are not wanted. For a long
// time nothing else checked them, so `npm run type-check` was green while test
// files called functions with obsolete signatures. tsconfig.test.json now covers
// them, but that is only true as long as someone keeps it true -- hence this
// test, which asks the TypeScript compiler itself which files each project in
// the `type-check` script would see, and fails by name for any test file that
// falls outside the union.
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';
import ts from 'typescript';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');

/** Every `*.test.ts` that exists on disk under src/, as absolute paths. */
function testFilesOnDisk(directory: string): string[] {
  const found: string[] = [];
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const absolute = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      found.push(...testFilesOnDisk(absolute));
    } else if (entry.name.endsWith('.test.ts')) {
      found.push(absolute);
    }
  }
  return found;
}

/**
 * The tsconfig projects `npm run type-check` compiles. A bare `tsc` means the
 * default tsconfig.json; `-p X` / `--project X` names one explicitly.
 */
function projectsInTypeCheckScript(): string[] {
  const manifest = JSON.parse(readFileSync(path.join(root, 'package.json'), 'utf8')) as {
    scripts?: Record<string, string>;
  };
  const script = manifest.scripts?.['type-check'];
  assert.ok(script, 'package.json has no "type-check" script for this guard to inspect');
  const projects: string[] = [];
  for (const command of script.split('&&')) {
    const words = command.trim().split(/\s+/);
    if (!words.includes('tsc')) continue;
    const flag = words.findIndex(word => word === '-p' || word === '--project');
    const project = flag === -1 ? 'tsconfig.json' : words[flag + 1];
    assert.ok(project, `"${command.trim()}" passes a project flag with no project`);
    projects.push(path.resolve(root, project));
  }
  assert.ok(projects.length > 0,
    `the "type-check" script (${script}) runs no tsc invocation this guard can read`);
  return projects;
}

/** The exact file set a tsconfig project compiles, per the compiler itself. */
function filesInProject(project: string): string[] {
  const read = ts.readConfigFile(project, ts.sys.readFile);
  assert.equal(read.error, undefined,
    `${path.relative(root, project)} could not be read as a tsconfig`);
  const parsed = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(project));
  assert.deepEqual(parsed.errors, [],
    `${path.relative(root, project)} did not parse cleanly`);
  return parsed.fileNames.map(file => path.resolve(file));
}

test('every test file is type-checked by npm run type-check', () => {
  const onDisk = testFilesOnDisk(path.join(root, 'src'));
  // An empty scan would make every assertion below vacuously true, which is
  // exactly the failure mode this guard exists to prevent. There are well over
  // a hundred test files in src/; a scan that finds a handful means the walk
  // broke, not that the tests went away.
  assert.ok(onDisk.length >= 50,
    `found only ${onDisk.length} *.test.ts files under src/ -- the scan is broken, `
    + 'so this guard is no longer checking anything');
  assert.ok(onDisk.includes(path.resolve(fileURLToPath(import.meta.url))),
    'the scan did not even find this file, so it is not finding test files correctly');

  const checked = new Set(projectsInTypeCheckScript().flatMap(filesInProject));
  const unchecked = onDisk.filter(file => !checked.has(file))
    .map(file => path.relative(root, file))
    .sort();
  assert.deepEqual(unchecked, [],
    `these test files are not type-checked by "npm run type-check":\n  ${unchecked.join('\n  ')}\n`
    + 'Add them to tsconfig.test.json (or another project the script compiles).');
});

test('the test project never emits', () => {
  // The whole reason tests live in their own project is that tsconfig.json
  // emits declarations into dist/. If the test project ever started emitting it
  // would put .d.ts files for test files there.
  const project = path.join(root, 'tsconfig.test.json');
  const read = ts.readConfigFile(project, ts.sys.readFile);
  const parsed = ts.parseJsonConfigFileContent(read.config, ts.sys, root);
  assert.equal(parsed.options.noEmit, true, 'tsconfig.test.json must set noEmit');
  assert.notEqual(parsed.options.declaration, true, 'tsconfig.test.json must not emit declarations');
});
