/**
 * `node --test` support: resolves `@boardgame/*` specifiers through each
 * package's `exports`, the way Vite does for the browser (see packages.mjs).
 * Registered with `--import` by `npm run test:unit`.
 */
import { registerHooks } from 'node:module';
import { pathToFileURL } from 'node:url';
import { clientPackages, resolvePackageSpecifier } from './packages.mjs';

const packages = clientPackages();

registerHooks({
  resolve(specifier, context, nextResolve) {
    if (specifier.startsWith('@boardgame/')) {
      const file = resolvePackageSpecifier(packages, specifier);
      if (file) return { url: pathToFileURL(file).href, shortCircuit: true };
    }
    return nextResolve(specifier, context);
  },
});
