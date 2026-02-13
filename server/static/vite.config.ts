import { defineConfig } from 'vite';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export default defineConfig({
  root: '.',
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
    port: 8080,
    open: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8888',
        changeOrigin: true
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
