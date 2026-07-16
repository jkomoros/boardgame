import assert from 'node:assert/strict';
import { execFile } from 'node:child_process';
import { mkdir, mkdtemp, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { promisify } from 'node:util';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const execFileAsync = promisify(execFile);
const runner = fileURLToPath(new URL('./check-client.mjs', import.meta.url));

async function run(project, { creatorPolicy = false, lit = false } = {}) {
  const arguments_ = project === undefined
    ? [runner]
    : [runner, '--project', project];
  if (creatorPolicy) {
    arguments_.push('--creator-policy');
  }
  if (lit) {
    arguments_.push('--lit');
  }
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

async function fixture(source, compilerOptions = {}, fileName = 'index.ts') {
  const directory = await mkdtemp(path.join(tmpdir(), 'boardgame-check-client-'));
  const project = path.join(directory, 'tsconfig.json');
  await mkdir(path.dirname(path.join(directory, fileName)), { recursive: true });
  await writeFile(path.join(directory, fileName), source);
  await writeFile(project, JSON.stringify({
    compilerOptions: {
      skipLibCheck: true,
      ...compilerOptions,
    },
    files: [fileName],
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

test('creator policy reports unsafe TypeScript escape hatches', async () => {
  const project = await fixture([
    'declare const input: unknown;',
    'const explicit: any = input;',
    'const doubled = input as unknown as string;',
    '// @ts-ignore',
    'const ignored: number = "wrong";',
    '// @ts-expect-error',
    'const expected: number = "wrong";',
    'void explicit; void doubled; void ignored; void expected;',
    '',
  ].join('\n'));

  const withoutPolicy = await run(project);
  assert.equal(withoutPolicy.exitCode, 0);
  assert.deepEqual(JSON.parse(withoutPolicy.stdout).diagnostics, []);

  const result = await run(project, { creatorPolicy: true });
  assert.equal(result.exitCode, 1);
  const document = JSON.parse(result.stdout);
  assert.deepEqual(document.diagnostics.map(({ code, line }) => ({ code, line })), [
    { code: 'BGCLIENT0101', line: 2 },
    { code: 'BGCLIENT0104', line: 3 },
    { code: 'BGCLIENT0102', line: 4 },
    { code: 'BGCLIENT0103', line: 6 },
  ]);
});

test('creator policy accepts explained suppressions and single assertions', async () => {
  const project = await fixture([
    '// @ts-expect-error Deliberately prove the compiler rejects this assignment.',
    'const expected: number = "wrong";',
    'declare const input: unknown;',
    'const narrowed = input as string;',
    'void expected; void narrowed;',
    '',
  ].join('\n'));
  const result = await run(project, { creatorPolicy: true });
  assert.equal(result.exitCode, 0);
  assert.deepEqual(JSON.parse(result.stdout).diagnostics, []);
});

test('creator policy rejects deep imports for facade-registered components', async () => {
  const project = await fixture([
    "import '../../src/components/boardgame-card.js';",
    "import '../../src/components/boardgame-hand-view-base.js';",
    "import '../../src/components/companion-avatar-catalog.js';",
    '',
  ].join('\n'));
  const result = await run(project, { creatorPolicy: true });
  assert.equal(result.exitCode, 1);
  assert.deepEqual(JSON.parse(result.stdout).diagnostics
    .filter(({ source }) => source === 'boardgame-client')
    .map(({ code, line }) => ({ code, line })), [
      { code: 'BGCLIENT0105', line: 1 },
      { code: 'BGCLIENT0105', line: 2 },
      { code: 'BGCLIENT0105', line: 3 },
    ]);
});

test('creator policy skips generated root files', async () => {
  const project = await fixture([
    '/*',
    ' * Auto-generated by boardgame-util. DO NOT EDIT.',
    ' */',
    'declare const input: unknown;',
    'const escaped: any = input as unknown as string;',
    'void escaped;',
    '',
  ].join('\n'), {}, 'client/_types.ts');
  const result = await run(project, { creatorPolicy: true });
  assert.equal(result.exitCode, 0);
  assert.deepEqual(JSON.parse(result.stdout).diagnostics, []);
});

test('creator policy does not exempt a reserved basename without the canonical header', async () => {
  const project = await fixture([
    'declare const input: unknown;',
    'const escaped: any = input;',
    'void escaped;',
    '',
  ].join('\n'), {}, 'client/_types.ts');
  const result = await run(project, { creatorPolicy: true });
  assert.equal(result.exitCode, 1);
  assert.deepEqual(JSON.parse(result.stdout).diagnostics.map(({ code, line }) => ({ code, line })), [
    { code: 'BGCLIENT0101', line: 2 },
  ]);
});

test('creator policy does not exempt a nested reserved basename', async () => {
  const project = await fixture([
    '/*',
    ' * Auto-generated by boardgame-util. DO NOT EDIT.',
    ' */',
    'declare const input: unknown;',
    'const escaped: any = input;',
    'void escaped;',
    '',
  ].join('\n'), {}, 'client/helpers/_types.ts');
  const result = await run(project, { creatorPolicy: true });
  assert.equal(result.exitCode, 1);
  assert.deepEqual(JSON.parse(result.stdout).diagnostics.map(({ code, line }) => ({ code, line })), [
    { code: 'BGCLIENT0101', line: 5 },
  ]);
});

test('creator policy cannot be bypassed with generated-looking prose', async () => {
  const project = await fixture([
    '/* This is not auto-generated, but its documentation may say DO NOT EDIT. */',
    'const marker = "@generated";',
    'const escaped: any = marker;',
    'void escaped;',
    '',
  ].join('\n'));
  const result = await run(project, { creatorPolicy: true });
  assert.equal(result.exitCode, 1);
  assert.deepEqual(JSON.parse(result.stdout).diagnostics.map(({ code, line }) => ({ code, line })), [
    { code: 'BGCLIENT0101', line: 3 },
  ]);
});

const litElementDeclaration = [
  'declare function html(strings: TemplateStringsArray, ...values: readonly unknown[]): unknown;',
  'class FixtureElement extends HTMLElement {',
  '  count = 0;',
  '}',
  "customElements.define('fixture-element', FixtureElement);",
];

test('Lit analysis accepts well-typed custom element bindings', async () => {
  const project = await fixture([
    ...litElementDeclaration,
    'const listener = (): void => undefined;',
    'const view = html`<fixture-element .count=${1} @click=${listener}></fixture-element>`;',
    'void view;',
    '',
  ].join('\n'));
  const result = await run(project, { lit: true });
  assert.equal(result.exitCode, 0, result.stdout);
  assert.deepEqual(JSON.parse(result.stdout).diagnostics, []);
});

test('Lit analysis reports strict binding diagnostics with stable rule codes', async () => {
  const project = await fixture([
    ...litElementDeclaration,
    'declare const objectValue: { readonly nested: string };',
    'const view = html`<fixture-element',
    '  .missing=${1}',
    '  .count=${"wrong"}',
    '  title=${objectValue}',
    '  @click=${42}',
    '></fixture-element>`;',
    'void view;',
    '',
  ].join('\n'));
  const first = await run(project, { lit: true });
  const second = await run(project, { lit: true });
  assert.equal(first.exitCode, 1, first.stdout);
  assert.equal(first.stderr, '');
  assert.equal(first.stdout, second.stdout);
  const diagnostics = JSON.parse(first.stdout).diagnostics.filter(({ source }) => source === 'lit');
  assert.deepEqual(diagnostics.map(({ code, category, file, line, column }) => ({
    code, category, file, line, column,
  })), [
    {
      code: 'LIT-no-unknown-property', category: 'error', file: 'index.ts', line: 8, column: 4,
    },
    {
      code: 'LIT-no-incompatible-type-binding', category: 'error', file: 'index.ts', line: 9, column: 4,
    },
    {
      code: 'LIT-no-complex-attribute-binding', category: 'error', file: 'index.ts', line: 10, column: 3,
    },
    {
      code: 'LIT-no-noncallable-event-binding', category: 'error', file: 'index.ts', line: 11, column: 4,
    },
  ]);
});
