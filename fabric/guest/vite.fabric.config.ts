/**
 * Builds the guest for a Common Fabric `cf-iframe`: one script
 * (`fabric/patterns/guest/guest.js`) and the document it runs in
 * (`guest.html`), which a pattern reads with `dataFile()`.
 *
 * It is the standalone build (`vite.config.ts`) with another entry point and
 * one more alias: Common Fabric's guest client, which must first be bundled
 * out of a labs checkout by `tools/vendor-fabric-client.sh`.
 */
import { defineConfig, type Plugin } from 'vite';
import { existsSync, mkdirSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import standalone from './vite.config.js';

const here = dirname(fileURLToPath(import.meta.url));
const outDir = resolve(here, '../patterns/guest');
const fabricClient = resolve(here, 'vendor/commonfabric-guest.js');

if (!existsSync(fabricClient)) {
  throw new Error(
    `${fabricClient} is missing. Run tools/vendor-fabric-client.sh <path to a labs checkout> first.`,
  );
}

const GUEST_HTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>boardgame guest</title>
  </head>
  <body></body>
</html>
`;

/** Writes the document beside the script, and a size report. */
function emitFabricGuest(): Plugin {
  return {
    name: 'guest-emit-fabric',
    generateBundle(_options, bundle) {
      const chunks = Object.values(bundle).filter(item => item.type === 'chunk');
      if (chunks.length !== 1) {
        this.error(`expected exactly one chunk, found ${chunks.length}`);
      }
      const chunk = chunks[0]!;
      if (chunk.type !== 'chunk') return;
      mkdirSync(outDir, { recursive: true });
      writeFileSync(resolve(outDir, 'guest.html'), GUEST_HTML);
      const modules = Object.entries(chunk.modules)
        .map(([id, info]) => ({ id: id.split('?')[0]!.replace(/^\0/, '').replace(resolve(here, '../..') + '/', ''), renderedBytes: info.renderedLength }))
        .filter(module => module.renderedBytes > 0)
        .sort((a, b) => b.renderedBytes - a.renderedBytes);
      writeFileSync(resolve(outDir, 'bundle-report.json'), JSON.stringify({
        minifiedBytes: chunk.code.length,
        modules,
      }, null, 2));
    },
  };
}

// Not `mergeConfig`: it joins arrays, and this build replaces the standalone
// build's entry, format and output plugin rather than adding to them.
const plugins = (standalone.plugins ?? []).filter(
  plugin => !(plugin && typeof plugin === 'object' && 'name' in plugin && plugin.name === 'guest-emit-single-file'),
);

export default defineConfig({
  ...standalone,
  plugins: [...plugins, emitFabricGuest()],
  resolve: {
    ...standalone.resolve,
    alias: {
      '@commonfabric/iframe-sandbox/guest': fabricClient,
    },
  },
  build: {
    ...standalone.build,
    outDir,
    emptyOutDir: false,
    lib: {
      entry: resolve(here, 'src/fabric-main.ts'),
      // `cf-iframe` runs the script as a module, and Common Fabric's client
      // awaits at its top level, which only a module may do.
      formats: ['es'],
      fileName: () => 'guest.js',
    },
  },
});
