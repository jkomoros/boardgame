import { defineConfig, type Plugin } from 'vite';
import { resolve, dirname, relative } from 'node:path';
import { fileURLToPath } from 'node:url';
import { mkdirSync, writeFileSync } from 'node:fs';

const here = dirname(fileURLToPath(import.meta.url));
const repo = resolve(here, '../..');
/**
 * Emit the one deliverable: dist/guest.html, with the whole bundle inlined
 * as a classic script, plus a JSON report of what the bundle is made of.
 */
function emitSingleFile(): Plugin {
  return {
    name: 'guest-emit-single-file',
    generateBundle(_options, bundle) {
      const chunks = Object.values(bundle).filter(item => item.type === 'chunk');
      if (chunks.length !== 1) {
        this.error(`expected exactly one chunk, found ${chunks.length}: a dynamic import() split the bundle`);
      }
      const chunk = chunks[0]!;
      if (chunk.type !== 'chunk') return;
      // A literal "</script" inside the code would end the inline element.
      const code = chunk.code.replace(/<\/(script)/gi, '<\\/$1');
      const html = [
        '<!doctype html>',
        '<html lang="en">',
        '<head>',
        '<meta charset="utf-8">',
        '<meta name="viewport" content="width=device-width, initial-scale=1">',
        '<title>boardgame guest</title>',
        '</head>',
        '<body>',
        `<script>${code}</script>`,
        '</body>',
        '</html>',
        '',
      ].join('\n');
      this.emitFile({ type: 'asset', fileName: 'guest.html', source: html });

      const modules = Object.entries(chunk.modules)
        .map(([id, info]) => ({
          id: relative(repo, id.split('?')[0]!.replace(/^\0/, '')),
          renderedBytes: info.renderedLength,
        }))
        .filter(module => module.renderedBytes > 0)
        .sort((a, b) => b.renderedBytes - a.renderedBytes);
      mkdirSync(resolve(here, 'dist'), { recursive: true });
      writeFileSync(resolve(here, 'dist/bundle-report.json'), JSON.stringify({
        minifiedBytes: chunk.code.length,
        modules,
      }, null, 2));
    },
  };
}

export default defineConfig({
  root: here,
  plugins: [emitSingleFile()],
  resolve: {
    // Game renderers are reached through the links under game-src/ so that
    // their `../../src/client.js` imports land on this package's src/client.ts.
    // The client packages are `file:` dependencies, linked into node_modules;
    // keeping the link paths also keeps their `lit` import inside this
    // package's node_modules.
    preserveSymlinks: true,
    // One copy of Lit and of each client package, whoever imports them.
    dedupe: ['lit', 'lit-html', 'lit-element', '@lit/reactive-element', '@boardgame/elements', '@boardgame/renderer-kit'],
  },
  esbuild: {
    // Lit's legacy decorators need these two flags. Vite would otherwise take
    // them from whichever tsconfig.json is nearest each source file, and the
    // renderers under examples/ have none: compiled with the defaults, class
    // fields shadow Lit's accessors and reactivity breaks silently.
    tsconfigRaw: {
      compilerOptions: {
        target: 'ES2022',
        experimentalDecorators: true,
        useDefineForClassFields: false,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    target: 'es2022',
    minify: 'esbuild',
    sourcemap: false,
    // Nothing may be left to fetch: no asset files, no preload helper.
    assetsInlineLimit: Number.MAX_SAFE_INTEGER,
    cssCodeSplit: false,
    modulePreload: false,
    lib: {
      entry: resolve(here, 'src/guest-main.ts'),
      formats: ['iife'],
      name: 'BoardgameGuest',
      fileName: () => 'guest.js',
    },
    rollupOptions: {
      output: { inlineDynamicImports: true },
    },
  },
});
