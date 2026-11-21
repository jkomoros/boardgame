import { defineConfig } from 'vite';
import { fileURLToPath } from 'node:url';

const here = fileURLToPath(new URL('.', import.meta.url));
const repo = fileURLToPath(new URL('../..', import.meta.url));

// The experiment imports the client source IN PLACE from server/static/src.
// That tree has no node_modules of its own in a fresh worktree, so `lit` is
// resolved from this package for every importer (and stays a single copy).
export default defineConfig({
  root: here,
  resolve: { dedupe: ['lit'], preserveSymlinks: true },
  optimizeDeps: {
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
