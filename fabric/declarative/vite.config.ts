import { defineConfig } from 'vite';
import { fileURLToPath } from 'node:url';

const here = fileURLToPath(new URL('.', import.meta.url));
const repo = fileURLToPath(new URL('../..', import.meta.url));

// The experiment is built on layer 1 alone: the `@boardgame/elements` package
// (server/static/packages/elements), installed as a `file:` dependency. The
// link is preserved so the package's `lit` import resolves from this
// package's node_modules, as one copy.
export default defineConfig({
  root: here,
  resolve: { dedupe: ['lit', '@boardgame/elements'], preserveSymlinks: true },
  // Lit's legacy decorators need these two flags, and the dev server does not
  // read a tsconfig for source it finds under node_modules.
  esbuild: {
    tsconfigRaw: {
      compilerOptions: { target: 'ES2022', experimentalDecorators: true, useDefineForClassFields: false },
    },
  },
  optimizeDeps: {
    // The elements package is TypeScript source, linked into node_modules.
    // Left to itself the dev server would pre-bundle each imported subpath
    // as a separate dependency: several copies of one module, compiled
    // without the decorator settings Lit needs. Serve it as source instead.
    exclude: ['@boardgame/elements'],
    include: [
      'lit', 'lit/decorators.js', 'lit/directives/class-map.js',
      'lit/directives/repeat.js', 'lit/directives/style-map.js',
      'lit/directives/when.js', 'lit/static-html.js', 'lit/async-directive.js',
    ],
  },
  server: {
    port: Number(process.env.DECLARATIVE_PORT || 5197),
    strictPort: true,
    fs: { strict: true, allow: [repo] },
  },
});
