/**
 * Layer check for `server/static/src` (fabric/PROPOSAL.md, stream A step 1).
 *
 * `scripts/layers.json` assigns every non-test source file to a layer. A file
 * may import from its own layer or a lower one. This module computes the
 * import graph with the TypeScript parser (so type-only imports, re-exports,
 * dynamic `import()` and `import('...')` type references are all seen) and
 * reports every upward edge. `scripts/check-layers.test.mjs` turns the report
 * into a `node --test` gate with a shrinking allow-list.
 *
 * Run `node scripts/check-layers.mjs` for a human-readable report, or
 * `node scripts/check-layers.mjs --graph` to dump every cross-layer edge.
 */
import { readFileSync, readdirSync, statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import ts from 'typescript';
import { clientPackages, packageOfFile, packageOfSpecifier } from './packages.mjs';

export const staticRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
export const tablePath = path.join(staticRoot, 'scripts', 'layers.json');

/** @typedef {{order: string[], unranked: string[], rules: {prefix?: string, suffix?: string, layer: string}[], files: Record<string,string>, allow: {runtime: Record<string,string>, typeOnly: Record<string,string>}}} LayerTable */

/** @returns {LayerTable} */
export function readLayerTable(file = tablePath) {
  return JSON.parse(readFileSync(file, 'utf8'));
}

const SOURCE_EXTENSIONS = ['.ts', '.js', '.mjs'];

function isTestFile(relative) {
  return /\.test\.(ts|js|mjs)$/.test(relative);
}

/** Every non-test source file under `src/` and `packages/<name>/src/`, as posix paths relative to `server/static`. */
export function sourceFiles(root = staticRoot) {
  const found = [];
  const walk = (directory) => {
    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      const absolute = path.join(directory, entry.name);
      if (entry.isDirectory()) {
        walk(absolute);
      } else if (SOURCE_EXTENSIONS.some(extension => entry.name.endsWith(extension))) {
        const relative = path.relative(root, absolute).split(path.sep).join('/');
        if (!isTestFile(relative)) found.push(relative);
      }
    }
  };
  walk(path.join(root, 'src'));
  // The client packages: packages/<name>/src.
  let packageNames = [];
  try {
    packageNames = readdirSync(path.join(root, 'packages'));
  } catch {
    // A tree with no packages directory is only src/.
  }
  for (const name of packageNames) {
    const sourceDir = path.join(root, 'packages', name, 'src');
    try {
      if (statSync(sourceDir).isDirectory()) walk(sourceDir);
    } catch {
      // Not a package directory.
    }
  }
  return found.sort();
}

/** The layer a file belongs to, or undefined when the table does not cover it. */
export function layerOf(table, file) {
  if (Object.hasOwn(table.files, file)) return table.files[file];
  for (const rule of table.rules) {
    if (rule.prefix !== undefined && !file.startsWith(rule.prefix)) continue;
    if (rule.suffix !== undefined && !file.endsWith(rule.suffix)) continue;
    return rule.layer;
  }
  return undefined;
}

function isFile(candidate) {
  try {
    return statSync(candidate).isFile();
  } catch {
    return false;
  }
}

function resolveRelative(root, fromFile, specifier) {
  const base = path.resolve(root, path.dirname(fromFile), specifier);
  const candidates = [base];
  if (base.endsWith('.js')) candidates.push(base.replace(/\.js$/, '.ts'), base.replace(/\.js$/, '.d.ts'));
  if (!path.extname(base)) candidates.push(`${base}.ts`, `${base}.d.ts`, path.join(base, 'index.ts'));
  const found = candidates.find(isFile);
  return found === undefined ? undefined : path.relative(root, found).split(path.sep).join('/');
}

/**
 * Imports of one file: `{specifier, typeOnly}`. An import is type-only when
 * TypeScript erases it: `import type`, `export type ... from`, an import whose
 * every named binding is `type`-qualified, an `import('...')` type node, or
 * anything in a `.d.ts`.
 */
export function importsOf(root, file) {
  const absolute = path.join(root, file);
  const text = readFileSync(absolute, 'utf8');
  const source = ts.createSourceFile(absolute, text, ts.ScriptTarget.Latest, true);
  const declarationFile = file.endsWith('.d.ts');
  const out = [];
  const add = (specifier, typeOnly) => out.push({ specifier, typeOnly: typeOnly || declarationFile });
  const namedAllType = (named) => named.elements.length > 0 && named.elements.every(element => element.isTypeOnly);
  const visit = (node) => {
    if (ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)) {
      const clause = node.importClause;
      let typeOnly = false;
      if (clause) {
        typeOnly = clause.isTypeOnly
          || (!clause.name && !!clause.namedBindings && ts.isNamedImports(clause.namedBindings) && namedAllType(clause.namedBindings));
      }
      add(node.moduleSpecifier.text, typeOnly);
    } else if (ts.isExportDeclaration(node) && node.moduleSpecifier && ts.isStringLiteral(node.moduleSpecifier)) {
      const typeOnly = node.isTypeOnly
        || (!!node.exportClause && ts.isNamedExports(node.exportClause) && namedAllType(node.exportClause));
      add(node.moduleSpecifier.text, typeOnly);
    } else if (ts.isImportTypeNode(node) && ts.isLiteralTypeNode(node.argument) && ts.isStringLiteral(node.argument.literal)) {
      add(node.argument.literal.text, true);
    } else if (ts.isCallExpression(node) && node.expression.kind === ts.SyntaxKind.ImportKeyword
      && node.arguments.length > 0 && ts.isStringLiteralLike(node.arguments[0])) {
      add(node.arguments[0].text, false);
    }
    ts.forEachChild(node, visit);
  };
  visit(source);
  return out;
}

/**
 * @returns {{
 *   unassigned: string[],
 *   stale: string[],
 *   unknownLayers: string[],
 *   unresolved: string[],
 *   boundary: string[],
 *   violations: {from: string, to: string, fromLayer: string, toLayer: string, typeOnly: boolean, key: string}[],
 *   edges: {from: string, to: string, fromLayer: string, toLayer: string, typeOnly: boolean}[],
 * }}
 */
export function analyze(table = readLayerTable(), root = staticRoot) {
  const files = sourceFiles(root);
  const known = new Set([...table.order, ...table.unranked]);
  const rank = new Map(table.order.map((layer, index) => [layer, index]));
  const unranked = new Set(table.unranked);
  const unassigned = [];
  const unknownLayers = [];
  for (const file of files) {
    const layer = layerOf(table, file);
    if (layer === undefined) unassigned.push(file);
    else if (!known.has(layer)) unknownLayers.push(`${file}: ${layer}`);
  }
  const onDisk = new Set(files);
  const stale = Object.keys(table.files).filter(file => !onDisk.has(file)).sort();
  const unresolved = [];
  const violations = [];
  const edges = [];
  const boundary = [];
  const packages = clientPackages(root);
  const packageLayers = table.packages ?? {};
  const seen = new Map();
  // A file's package must agree with its layer, in both directions: a package
  // holds exactly one layer, and no file of that layer lives outside it.
  const layerPackages = new Map(Object.entries(packageLayers).map(([dir, layer]) => [layer, dir]));
  for (const file of files) {
    const layer = layerOf(table, file);
    if (layer === undefined) continue;
    const pkg = packageOfFile(packages, file);
    if (pkg) {
      const expected = packageLayers[pkg.dir];
      if (expected === undefined) boundary.push(`${file}: package ${pkg.dir} has no layer in the table's "packages"`);
      else if (layer !== expected) boundary.push(`${file}: is in ${pkg.name} (layer ${expected}) but the table says ${layer}`);
    } else if (layerPackages.has(layer)) {
      boundary.push(`${file}: layer ${layer} lives in ${layerPackages.get(layer)}, not in src/`);
    }
  }
  for (const pkg of packages) {
    for (const [specifier, target] of pkg.exports) {
      if (!isFile(target.file)) boundary.push(`${pkg.name}: export "${specifier}" points at a file that does not exist`);
    }
  }
  for (const from of files) {
    const fromLayer = layerOf(table, from);
    if (fromLayer === undefined) continue;
    const fromPackage = packageOfFile(packages, from);
    // Test support (compile-time contract tests, fixtures, demos) may reach a
    // package's internals by specifier; ranked code may not.
    const exempt = (table.unranked ?? []).includes(fromLayer);
    for (const { specifier, typeOnly } of importsOf(root, from)) {
      let to;
      const targetPackage = packageOfSpecifier(packages, specifier);
      if (targetPackage) {
        const exported = targetPackage.exports.get(specifier);
        if (exported) {
          to = path.relative(root, exported.file).split(path.sep).join('/');
          if (exported.typesOnly && !typeOnly) {
            boundary.push(`${from} -> ${specifier}: a types-only export imported at runtime`);
          }
        } else {
          const sub = specifier.slice(targetPackage.name.length + 1);
          to = resolveRelative(root, `${targetPackage.dir}/src/x`, `./${sub}`);
          if (to === undefined) {
            unresolved.push(`${from} -> ${specifier}`);
            continue;
          }
          if (!exempt) boundary.push(`${from} -> ${specifier}: not in ${targetPackage.name}'s exports`);
        }
        if (fromPackage === targetPackage) {
          boundary.push(`${from} -> ${specifier}: a package imports its own files by relative path`);
        } else if (fromPackage && !fromPackage.dependencies.includes(targetPackage.name)) {
          boundary.push(`${from} -> ${specifier}: ${fromPackage.name} does not declare ${targetPackage.name} as a dependency`);
        }
      } else if (specifier.startsWith('.')) {
        if (/\.(css|svg|json|png|jpg)(\?.*)?$/.test(specifier)) continue;
        to = resolveRelative(root, from, specifier);
        if (to === undefined) {
          unresolved.push(`${from} -> ${specifier}`);
          continue;
        }
        const toPackage = packageOfFile(packages, to);
        if (toPackage !== fromPackage) {
          boundary.push(`${from} -> ${specifier}: a relative import crosses a package boundary; use ${toPackage ? toPackage.name : 'no import at all'}`);
        }
      } else {
        continue;
      }
      if (isTestFile(to)) continue;
      const toLayer = layerOf(table, to);
      if (toLayer === undefined) continue;
      // One edge per (from, to); a runtime import outranks a type-only one.
      const edgeKey = `${from} -> ${to}`;
      const previous = seen.get(edgeKey);
      if (previous) {
        if (previous.typeOnly && !typeOnly) previous.typeOnly = false;
        continue;
      }
      const edge = { from, to, fromLayer, toLayer, typeOnly };
      seen.set(edgeKey, edge);
      edges.push(edge);
    }
  }
  for (const edge of edges) {
    // Unranked layers (test support) may import anything, and nothing ranked
    // may import them.
    let upward;
    if (unranked.has(edge.fromLayer)) upward = false;
    else if (unranked.has(edge.toLayer)) upward = true;
    else upward = rank.get(edge.toLayer) > rank.get(edge.fromLayer);
    if (upward) violations.push({ ...edge, key: `${edge.from} -> ${edge.to}` });
  }
  violations.sort((a, b) => a.key.localeCompare(b.key));
  boundary.sort();
  return { unassigned, stale, unknownLayers, unresolved, violations, edges, boundary };
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const table = readLayerTable();
  const result = analyze(table);
  if (process.argv.includes('--graph')) {
    for (const edge of result.edges) {
      if (edge.fromLayer !== edge.toLayer) {
        console.log(`${edge.fromLayer} -> ${edge.toLayer}\t${edge.typeOnly ? 'type' : 'runtime'}\t${edge.from} -> ${edge.to}`);
      }
    }
  } else {
    for (const file of result.unassigned) console.log(`UNASSIGNED ${file}`);
    for (const file of result.stale) console.log(`STALE ${file}`);
    for (const entry of result.unresolved) console.log(`UNRESOLVED ${entry}`);
    for (const entry of result.boundary) console.log(`BOUNDARY ${entry}`);
    for (const violation of result.violations) {
      const allowed = (violation.typeOnly ? table.allow.typeOnly : table.allow.runtime)[violation.key] !== undefined;
      console.log(`${allowed ? 'allowed' : 'VIOLATION'}\t${violation.typeOnly ? 'type' : 'runtime'}\t${violation.fromLayer} -> ${violation.toLayer}\t${violation.key}`);
    }
    console.log(`${result.violations.length} upward imports (${result.violations.filter(v => !v.typeOnly).length} runtime, ${result.violations.filter(v => v.typeOnly).length} type-only)`);
  }
}
