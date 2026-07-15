import { execFileSync } from 'node:child_process';
import { mkdtemp, mkdir, rm, symlink, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const staticRoot = resolve(here, '../..');
const repoRoot = resolve(staticRoot, '../..');
const assembled = await mkdtemp(resolve(tmpdir(), 'boardgame-facade-production-'));

try {
  await mkdir(resolve(assembled, 'game-src'), { recursive: true });
  // Junctions work without elevated symlink privileges on Windows and behave
  // as ordinary directory symlinks on the other supported platforms.
  await symlink(resolve(staticRoot, 'src'), resolve(assembled, 'src'), 'junction');
  await symlink(resolve(staticRoot, 'node_modules'), resolve(assembled, 'node_modules'), 'junction');
  await symlink(resolve(repoRoot, 'examples/pig/client'), resolve(assembled, 'game-src/pig'), 'junction');
  await symlink(resolve(repoRoot, 'examples/memory/client'), resolve(assembled, 'game-src/memory'), 'junction');

  await writeFile(resolve(assembled, 'entry.ts'),
    "import './game-src/pig/boardgame-render-game-pig.ts';\nimport './game-src/memory/boardgame-render-game-memory.ts';\n");
  await writeFile(resolve(assembled, 'vite.config.mjs'), `
    import { defineConfig } from 'vite';
    import { resolve } from 'node:path';
    export default defineConfig({
      root: ${JSON.stringify(assembled)},
      resolve: { preserveSymlinks: true },
      build: {
        outDir: 'dist',
        emptyOutDir: true,
        rollupOptions: { input: resolve(${JSON.stringify(assembled)}, 'entry.ts') },
      },
    });
  `);

  execFileSync(process.execPath, [
    resolve(staticRoot, 'node_modules/vite/bin/vite.js'),
    'build',
    '--config',
    resolve(assembled, 'vite.config.mjs'),
  ], { cwd: assembled, stdio: 'inherit' });
} finally {
  await rm(assembled, { recursive: true, force: true });
}
