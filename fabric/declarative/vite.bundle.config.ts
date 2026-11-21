/**
 * Builds `dist/bg-elements.js`: the declarative element set as one ES module
 * with no imports, for a host page outside this repository (the Common
 * Fabric shell spike loads it by path).
 */
import { defineConfig } from 'vite';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';

const here = fileURLToPath(new URL('.', import.meta.url));

export default defineConfig({
  root: here,
  resolve: { dedupe: ['lit'], preserveSymlinks: true },
  build: {
    outDir: resolve(here, 'dist'),
    emptyOutDir: true,
    target: 'es2022',
    lib: {
      entry: resolve(here, 'src/bundle.ts'),
      formats: ['es'],
      fileName: () => 'bg-elements.js',
    },
    rollupOptions: { output: { inlineDynamicImports: true } },
  },
  esbuild: {
    tsconfigRaw: {
      compilerOptions: { experimentalDecorators: true, useDefineForClassFields: false },
    },
  },
});
