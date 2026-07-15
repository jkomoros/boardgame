import assert from 'node:assert/strict';
import { execFile } from 'node:child_process';
import { mkdtemp, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { promisify } from 'node:util';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const execFileAsync = promisify(execFile);
const runner = fileURLToPath(new URL('./check-client.mjs', import.meta.url));

async function run(project) {
  const arguments_ = project === undefined
    ? [runner]
    : [runner, '--project', project];
  try {
    const result = await execFileAsync(process.execPath, arguments_);
    return { exitCode: 0, stdout: result.stdout, stderr: result.stderr };
  } catch (error) {
    return {
      exitCode: error.code,
      stdout: error.stdout,
      stderr: error.stderr,
    };
  }
}

async function fixture(source, compilerOptions = {}) {
  const directory = await mkdtemp(path.join(tmpdir(), 'boardgame-check-client-'));
  const project = path.join(directory, 'tsconfig.json');
  await writeFile(path.join(directory, 'index.ts'), source);
  await writeFile(project, JSON.stringify({
    compilerOptions: {
      skipLibCheck: true,
      ...compilerOptions,
    },
    files: ['index.ts'],
  }));
  return project;
}

test('green project emits one empty diagnostic document', async () => {
  const result = await run(await fixture('const answer: number = 42;\n'));
  assert.equal(result.exitCode, 0);
  assert.equal(result.stderr, '');
  const document = JSON.parse(result.stdout);
  assert.equal(document.version, 1);
  assert.deepEqual(document.diagnostics, []);
  assert.equal(result.stdout.trim().split('\n').length, 1);
});

test('red project reports deterministic native TypeScript diagnostics', async () => {
  const project = await fixture([
    'const second: number = "also wrong";',
    'const first: number = "wrong";',
    '',
  ].join('\n'));
  const firstRun = await run(project);
  const secondRun = await run(project);

  assert.equal(firstRun.exitCode, 1);
  assert.equal(firstRun.stderr, '');
  assert.equal(firstRun.stdout, secondRun.stdout);
  const document = JSON.parse(firstRun.stdout);
  assert.deepEqual(document.diagnostics.map(({ code, file, line }) => ({ code, file, line })), [
    { code: 'TS2322', file: 'index.ts', line: 1 },
    { code: 'TS2322', file: 'index.ts', line: 2 },
  ]);
});

test('bad invocation is an infrastructure failure with valid JSON', async () => {
  const result = await run();
  assert.equal(result.exitCode, 2);
  const document = JSON.parse(result.stdout);
  assert.deepEqual(document.diagnostics, []);
  assert.match(document.infrastructureError, /--project/);
});

test('a project cannot weaken the required authoring checks', async () => {
  const result = await run(await fixture(
    'const read = (values: string[]) => values[0].toUpperCase();\n',
    { strict: false, noUncheckedIndexedAccess: false },
  ));
  assert.equal(result.exitCode, 1);
  const document = JSON.parse(result.stdout);
  assert.ok(document.diagnostics.some(({ code }) => code === 'TS2532'));
});
