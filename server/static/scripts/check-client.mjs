#!/usr/bin/env node

import process from 'node:process';
import path from 'node:path';
import ts from 'typescript';

const VERSION = 1;
const REQUIRED_COMPILER_OPTIONS = Object.freeze({
  strict: true,
  noUncheckedIndexedAccess: true,
  exactOptionalPropertyTypes: true,
  noPropertyAccessFromIndexSignature: true,
  noImplicitOverride: true,
  useUnknownInCatchVariables: true,
  noImplicitReturns: true,
  noFallthroughCasesInSwitch: true,
  noUnusedLocals: true,
  noUnusedParameters: true,
  noEmit: true,
});

function posixPath(value) {
  return value.split(path.sep).join('/');
}

function displayPath(value, baseDirectory) {
  const relative = path.relative(baseDirectory, value);
  return posixPath(relative || path.basename(value));
}

function categoryName(category) {
  return ts.DiagnosticCategory[category]?.toLowerCase() ?? 'error';
}

function serializeDiagnostic(diagnostic, baseDirectory) {
  const result = {
    source: 'typescript',
    code: `TS${diagnostic.code}`,
    category: categoryName(diagnostic.category),
    message: ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'),
  };

  if (diagnostic.file) {
    result.file = displayPath(diagnostic.file.fileName, baseDirectory);
    if (diagnostic.start !== undefined) {
      const location = diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start);
      result.line = location.line + 1;
      result.column = location.character + 1;
    }
  }

  return result;
}

function compareDiagnostics(left, right) {
  const leftFile = left.file ?? '';
  const rightFile = right.file ?? '';
  return (leftFile < rightFile ? -1 : leftFile > rightFile ? 1 : 0)
    || (left.line ?? 0) - (right.line ?? 0)
    || (left.column ?? 0) - (right.column ?? 0)
    || (left.code < right.code ? -1 : left.code > right.code ? 1 : 0)
    || (left.message < right.message ? -1 : left.message > right.message ? 1 : 0);
}

function parseArguments(argv) {
  let project;
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === '--project') {
      project = argv[index + 1];
      index += 1;
    } else if (argument.startsWith('--project=')) {
      project = argument.slice('--project='.length);
    } else {
      throw new Error(`unknown argument: ${argument}`);
    }
  }
  if (!project) {
    throw new Error('missing required --project <tsconfig.json>');
  }
  return { project: path.resolve(project) };
}

function readProject(projectPath) {
  const readResult = ts.readConfigFile(projectPath, ts.sys.readFile);
  if (readResult.error) {
    return { diagnostics: [readResult.error] };
  }

  const parsed = ts.parseJsonConfigFileContent(
    readResult.config,
    ts.sys,
    path.dirname(projectPath),
    undefined,
    projectPath,
  );
  if (parsed.errors.length > 0) {
    return { diagnostics: parsed.errors };
  }
  return { parsed };
}

function emit(document) {
  process.stdout.write(`${JSON.stringify(document)}\n`);
}

function main() {
  let projectPath;
  try {
    ({ project: projectPath } = parseArguments(process.argv.slice(2)));
  } catch (error) {
    emit({
      version: VERSION,
      project: null,
      diagnostics: [],
      infrastructureError: error instanceof Error ? error.message : String(error),
    });
    return 2;
  }

  const baseDirectory = path.dirname(projectPath);
  const project = displayPath(projectPath, process.cwd());
  try {
    const loaded = readProject(projectPath);
    if (loaded.diagnostics) {
      const diagnostics = loaded.diagnostics
        .map((diagnostic) => serializeDiagnostic(diagnostic, baseDirectory))
        .sort(compareDiagnostics);
      emit({ version: VERSION, project, diagnostics });
      return 1;
    }

    const program = ts.createProgram({
      rootNames: loaded.parsed.fileNames,
      options: {
        ...loaded.parsed.options,
        ...REQUIRED_COMPILER_OPTIONS,
      },
      projectReferences: loaded.parsed.projectReferences,
    });
    const diagnostics = ts.getPreEmitDiagnostics(program)
      .map((diagnostic) => serializeDiagnostic(diagnostic, baseDirectory))
      .sort(compareDiagnostics);
    emit({ version: VERSION, project, diagnostics });
    return diagnostics.length === 0 ? 0 : 1;
  } catch (error) {
    emit({
      version: VERSION,
      project,
      diagnostics: [],
      infrastructureError: error instanceof Error ? error.message : String(error),
    });
    return 2;
  }
}

process.exitCode = main();
