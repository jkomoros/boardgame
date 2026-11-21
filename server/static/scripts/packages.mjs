/**
 * The client packages under `server/static/packages/` and how the webapp
 * resolves them.
 *
 * Each package has a `package.json` whose `exports` lists every public
 * subpath explicitly. A host outside this directory (the Fabric guest)
 * installs a package with a `file:` dependency and resolves it through
 * `exports` in the ordinary way. The webapp itself does not go through
 * `node_modules`: the Go build keeps one `node_modules` in a per-user cache
 * that no checkout's packages can be linked into. Instead this module turns
 * the same `exports` maps into
 *
 *   - Vite aliases (`vite.config.ts`, `tests/renderer/facade-production.mjs`),
 *   - a Node resolve hook for `node --test` (`scripts/package-resolve.mjs`),
 *
 * so a subpath that is not exported does not resolve for the webapp either.
 * TypeScript uses `paths` in `tsconfig.json`, which cannot read `exports`;
 * `scripts/check-layers.test.mjs` closes that gap by checking every
 * cross-package import against the `exports` map.
 */
import { readFileSync, readdirSync, statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

export const staticRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');

/**
 * @typedef {{ name: string, dir: string, directory: string, dependencies: string[], exports: Map<string, { file: string, typesOnly: boolean }> }} ClientPackage
 * `dir` is relative to `server/static` (posix); `directory` is absolute.
 * `exports` maps a full specifier (`@boardgame/elements/motion/timing.js`)
 * to the absolute file behind it.
 */

/** @returns {ClientPackage[]} */
export function clientPackages(root = staticRoot) {
  const packagesDir = path.join(root, 'packages');
  let names;
  try {
    names = readdirSync(packagesDir);
  } catch {
    return [];
  }
  const found = [];
  for (const entry of names.sort()) {
    const directory = path.join(packagesDir, entry);
    let manifest;
    try {
      if (!statSync(directory).isDirectory()) continue;
      manifest = JSON.parse(readFileSync(path.join(directory, 'package.json'), 'utf8'));
    } catch {
      continue;
    }
    const exportsMap = new Map();
    for (const [subpath, target] of Object.entries(manifest.exports ?? {})) {
      if (subpath.includes('*')) {
        throw new Error(`${manifest.name}: exports must be explicit; "${subpath}" is a pattern`);
      }
      const typesOnly = typeof target === 'object' && target !== null && !('default' in target) && !('import' in target);
      const relative = typeof target === 'string' ? target : (target.default ?? target.import ?? target.types);
      if (typeof relative !== 'string' || !relative.startsWith('./')) {
        throw new Error(`${manifest.name}: export "${subpath}" has no usable target`);
      }
      const specifier = subpath === '.' ? manifest.name : `${manifest.name}/${subpath.slice(2)}`;
      exportsMap.set(specifier, { file: path.join(directory, relative), typesOnly });
    }
    found.push({
      name: manifest.name,
      dir: `packages/${entry}`,
      directory,
      dependencies: Object.keys({ ...manifest.dependencies, ...manifest.peerDependencies }),
      exports: exportsMap,
    });
  }
  return found;
}

/** The package a `server/static`-relative posix path belongs to, or undefined. */
export function packageOfFile(packages, file) {
  return packages.find(pkg => file.startsWith(`${pkg.dir}/`));
}

/** The package a bare specifier names, or undefined when it is not one of ours. */
export function packageOfSpecifier(packages, specifier) {
  return packages.find(pkg => specifier === pkg.name || specifier.startsWith(`${pkg.name}/`));
}

/** Vite `resolve.alias` entries: one exact match per exported runtime subpath. */
export function packageAliases(root = staticRoot) {
  const aliases = [];
  for (const pkg of clientPackages(root)) {
    for (const [specifier, target] of pkg.exports) {
      if (target.typesOnly) continue;
      aliases.push({
        find: new RegExp(`^${specifier.replace(/[.*+?^${}()|[\]\\/]/g, '\\$&')}$`),
        replacement: target.file,
      });
    }
  }
  return aliases;
}

/** Resolves one of our package specifiers to a file, or undefined. Throws for an unexported subpath. */
export function resolvePackageSpecifier(packages, specifier) {
  const pkg = packageOfSpecifier(packages, specifier);
  if (!pkg) return undefined;
  const target = pkg.exports.get(specifier);
  if (!target) {
    throw new Error(`"${specifier}" is not exported by ${pkg.name} (see ${pkg.dir}/package.json)`);
  }
  return target.file;
}
