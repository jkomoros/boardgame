import { defineConfig } from 'vite';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';
import { existsSync } from 'node:fs';
import type { Plugin } from 'vite';
import { clientPackages, packageAliases } from './scripts/packages.mjs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const staticPort = Number(process.env.BOARDGAME_STATIC_PORT || 8080);
const apiPort = Number(process.env.BOARDGAME_API_PORT || 8888);

// The directory Vite serves: server/static itself, or the build directory the
// Go tooling assembles from it (where `src`, `packages` and `scripts` are links).
const servedRoot = process.cwd();

/**
 * Keeps every pre-package module URL working.
 *
 * Layers 1 and 2 moved from `src/` into `packages/<name>/src/` with their paths
 * inside `src/` unchanged. The Playwright suites import several hundred
 * modules by URL from inside the page (`await import('/src/motion/timing.ts')`),
 * which the type checker never sees. For any `/src/<path>` that no longer
 * exists but does exist in a package, this serves a module that re-exports
 * the package's file, so the page gets the one instance the app itself uses.
 * It is test-path indirection only: it bypasses `exports`, and no source file
 * may rely on it.
 */
function legacySrcPaths(): Plugin {
  const packages = clientPackages(servedRoot);
  const redirects = new Map<string, string>();
  return {
    name: 'boardgame-legacy-src-paths',
    enforce: 'pre',
    resolveId(source) {
      const url = source.split('?')[0]!;
      if (!url.startsWith('/src/')) return null;
      const here = resolve(servedRoot, url.slice(1));
      const candidates = here.endsWith('.js') ? [here, here.replace(/\.js$/, '.ts')] : [here];
      if (candidates.some(candidate => existsSync(candidate))) return null;
      for (const pkg of packages) {
        for (const candidate of candidates) {
          const moved = resolve(pkg.directory, 'src', candidate.slice(resolve(servedRoot, 'src').length + 1));
          if (existsSync(moved)) {
            const id = here.replace(/\.js$/, '.ts');
            redirects.set(id, moved);
            return id;
          }
        }
      }
      return null;
    },
    load(id) {
      const moved = redirects.get(id.split('?')[0]!);
      if (!moved) return null;
      return `export * from ${JSON.stringify(moved)};\n`;
    },
  };
}

export default defineConfig({
  root: '.',
  plugins: [legacySrcPaths()],
  // Game renderers are discovered dynamically, so Vite's initial crawl (which
  // follows STATIC imports from index.html) cannot see their dependencies.
  // Whatever it misses is discovered when a game page first loads, and Vite
  // answers that with `optimized dependencies changed. reloading` -- a full
  // page reload, mid-test, that erases every animation counter the parity
  // suite is in the middle of collecting.
  //
  // That is not hypothetical and it is not only a test problem: measured
  // against a freshly started `boardgame-util serve --offline-dev-mode`, the
  // debuganimations trace scenario died during its setup drain with
  //   [vite] ✨ new dependencies optimized: lit/async-directive.js
  //   [vite] ✨ optimized dependencies changed. reloading
  // and the test reported `Cannot read properties of undefined (reading
  // 'gateCloses')` -- the hooks object had gone with the page. One entry
  // (style-map) was already listed here for exactly this reason; the list was
  // just incomplete, which is the failure mode of any hand-maintained
  // allowlist.
  //
  // So this is the WHOLE set of bare specifiers imported by client code under
  // `src/` or `examples/*/client/`, not the subset someone happened to hit.
  // Listing a specifier the crawl would have found anyway costs nothing.
  //
  // DO NOT hand-maintain this. `src/vite-optimize-deps.test.ts` derives the set
  // from the source and fails naming any specifier that is missing or stale --
  // which is how the last three below were found. The grep recipe that used to
  // live here was hardcoded to four package prefixes, so `reselect`,
  // `reselect-tools/src` and `firebase/compat/app` were invisible to the very
  // recipe offered for regenerating the list. A hand-maintained allowlist plus
  // a hand-maintained way to regenerate it is two copies of the same mistake.
  optimizeDeps: {
    include: [
      'lit',
      'lit/async-directive.js',
      'lit/static-html.js',
      'lit/decorators.js',
      'lit/directives/class-map.js',
      'lit/directives/repeat.js',
      'lit/directives/style-map.js',
      'lit/directives/when.js',
      '@material/web/button/filled-button.js',
      '@material/web/button/outlined-button.js',
      '@material/web/checkbox/checkbox.js',
      '@material/web/chips/assist-chip.js',
      '@material/web/dialog/dialog.js',
      '@material/web/icon/icon.js',
      '@material/web/iconbutton/icon-button.js',
      '@material/web/progress/linear-progress.js',
      '@material/web/radio/radio.js',
      '@material/web/select/filled-select.js',
      '@material/web/select/select-option.js',
      '@material/web/slider/slider.js',
      '@material/web/switch/switch.js',
      '@material/web/textfield/filled-text-field.js',
      'pwa-helpers/connect-mixin.js',
      'pwa-helpers/lazy-reducer-enhancer.js',
      'pwa-helpers/router.js',
      'redux',
      'redux-thunk',
      'firebase/compat/app',
      'firebase/compat/auth',
      'reselect',
      'reselect-tools/src',
    ],
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html')
      }
    }
  },
  server: {
    port: staticPort,
    strictPort: true,
    open: false,
    proxy: {
      '/api': {
        target: `http://127.0.0.1:${apiPort}`,
        changeOrigin: true,
        ws: true  // Enable WebSocket proxying
      }
    },
    fs: {
      // Allow serving files from the root and follow symlinks
      strict: false,
      allow: ['..']
    }
  },
  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx', '.json'],
    // Preserve symlinks to ensure correct path resolution for game renderers
    preserveSymlinks: true,
    alias: [
      // Allow game renderers to import from a clean path
      { find: '/@server-static', replacement: resolve(__dirname, '.') },
      // The client packages, one exact alias per exported subpath.
      ...packageAliases(servedRoot),
    ]
  }
});
