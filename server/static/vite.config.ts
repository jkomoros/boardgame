import { defineConfig } from 'vite';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const staticPort = Number(process.env.BOARDGAME_STATIC_PORT || 8080);
const apiPort = Number(process.env.BOARDGAME_API_PORT || 8888);

export default defineConfig({
  root: '.',
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
  // So this is now the WHOLE set of bare specifiers imported anywhere under
  // `src/` or `examples/*/client/`, not the subset someone happened to hit.
  // Listing a specifier the crawl would have found anyway costs nothing.
  // Regenerate with:
  //   grep -rhoE "from '(lit[^']*|@material/web[^']*|redux[^']*|pwa-helpers[^']*)'" \
  //     examples/*/client/ server/static/src/ | sort -u
  optimizeDeps: {
    include: [
      'lit',
      'lit/async-directive.js',
      'lit/decorators.js',
      'lit/directives/class-map.js',
      'lit/directives/repeat.js',
      'lit/directives/style-map.js',
      'lit/directives/when.js',
      '@material/web/checkbox/checkbox.js',
      '@material/web/dialog/dialog.js',
      '@material/web/radio/radio.js',
      '@material/web/select/filled-select.js',
      '@material/web/slider/slider.js',
      '@material/web/switch/switch.js',
      'pwa-helpers/connect-mixin.js',
      'pwa-helpers/lazy-reducer-enhancer.js',
      'pwa-helpers/router.js',
      'redux',
      'redux-thunk',
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
    alias: {
      // Allow game renderers to import from a clean path
      '/@server-static': resolve(__dirname, '.')
    }
  }
});
