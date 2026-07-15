import { defineConfig } from 'vite';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const staticPort = Number(process.env.BOARDGAME_STATIC_PORT || 8080);
const apiPort = Number(process.env.BOARDGAME_API_PORT || 8888);

export default defineConfig({
  root: '.',
  // Game renderers are discovered dynamically, so Vite's initial crawl cannot
  // see all of their dependencies. Prebundle the known dynamic Lit directive
  // to prevent a mid-test dependency-optimization reload from erasing runtime
  // animation evidence.
  optimizeDeps: {
    include: ['lit/directives/style-map.js'],
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
