// The engine has to load inside a Common Fabric pattern, whose sandbox
// verifies every module-scope statement before a module may run (labs
// docs/specs/sandboxing/SES_SANDBOXING_SPEC.md, section 4.2). This test
// reads every engine source, the rules layer, the legal evaluator and the
// games included, and fails by name on each form that sandbox refuses:
//
//   - a class at module scope (it would hold mutable prototype state);
//   - `let` or `var` at module scope;
//   - a module-scope constant that is not a function and not plain data: a
//     record holding functions, a `new` of anything, a stateful regular
//     expression. A call is allowed, since its result is checked when the
//     module loads; games/*.ts are loaded by sandbox tests for that;
//   - a generator at module scope;
//   - a double cast through `unknown`, which the pattern compiler rejects;
//   - an import that leaves the engine directory.
//
// It is the cheap, labs-free half of the check. The other half is loading
// the engine in a real pattern: fabric/patterns/*/parity.test.tsx.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join, relative, resolve, dirname } from 'node:path';
import ts from 'typescript';

const HERE = fileURLToPath(new URL('.', import.meta.url));

function sources(dir: string, out: string[] = []): string[] {
	for (const entry of readdirSync(dir, { withFileTypes: true })) {
		const path = join(dir, entry.name);
		if (entry.isDirectory()) {
			if (entry.name === 'testing' || entry.name === 'testdata') continue;
			sources(path, out);
		} else if (entry.name.endsWith('.ts') && !entry.name.endsWith('.test.ts')) {
			out.push(path);
		}
	}
	return out;
}

/** Whether an initializer is data the sandbox can freeze: no functions in it. */
function plainData(node: ts.Expression): boolean {
	if (ts.isStringLiteralLike(node) || ts.isNumericLiteral(node) || ts.isBigIntLiteral(node) || ts.isTemplateExpression(node) || ts.isIdentifier(node)) return true;
	if (node.kind === ts.SyntaxKind.TrueKeyword || node.kind === ts.SyntaxKind.FalseKeyword || node.kind === ts.SyntaxKind.NullKeyword) return true;
	if (ts.isPrefixUnaryExpression(node)) return plainData(node.operand);
	if (ts.isAsExpression(node) || ts.isParenthesizedExpression(node) || ts.isSatisfiesExpression(node)) return plainData(node.expression);
	if (ts.isBinaryExpression(node)) return plainData(node.left) && plainData(node.right);
	if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) return true;
	if (ts.isArrayLiteralExpression(node)) return node.elements.every(element => !ts.isSpreadElement(element) && !ts.isOmittedExpression(element) && plainData(element));
	if (ts.isObjectLiteralExpression(node)) {
		return node.properties.every(property => ts.isPropertyAssignment(property) && plainData(property.initializer));
	}
	if (ts.isRegularExpressionLiteral(node)) return !/[gy][a-z]*$/.test(node.text.slice(node.text.lastIndexOf('/')));
	// A call's result is validated when the module loads.
	if (ts.isCallExpression(node)) return true;
	return false;
}

const files = sources(HERE);

test('every engine source is scanned', () => {
	assert.ok(files.length >= 49, `only ${files.length} sources found`);
	for (const dir of ['rules', 'legal', 'games']) {
		assert.ok(files.some(file => relative(HERE, file).startsWith(dir + '/')), `${dir}/ is not scanned`);
	}
});

test('no engine module holds anything at module scope that the pattern sandbox refuses', () => {
	const problems: string[] = [];
	for (const file of files) {
		const name = relative(HERE, file);
		const text = readFileSync(file, 'utf8');
		const source = ts.createSourceFile(file, text, ts.ScriptTarget.ES2022, true, ts.ScriptKind.TS);
		const at = (node: ts.Node) => `${name}:${source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1}`;
		for (const statement of source.statements) {
			if (ts.isClassDeclaration(statement)) problems.push(`${at(statement)} declares a class at module scope`);
			if (ts.isFunctionDeclaration(statement) && statement.asteriskToken) problems.push(`${at(statement)} declares a generator at module scope`);
			if (ts.isVariableStatement(statement)) {
				if (!(statement.declarationList.flags & ts.NodeFlags.Const)) problems.push(`${at(statement)} declares let or var at module scope`);
				for (const declaration of statement.declarationList.declarations) {
					const init = declaration.initializer;
					if (init === undefined || ts.isArrowFunction(init) || ts.isFunctionExpression(init)) continue;
					if (!plainData(init)) problems.push(`${at(declaration)} holds something other than a function or plain data at module scope`);
				}
			}
			if (ts.isImportDeclaration(statement) || (ts.isExportDeclaration(statement) && statement.moduleSpecifier)) {
				const specifier = (statement.moduleSpecifier as ts.StringLiteral).text;
				const target = resolve(dirname(file), specifier);
				if (!specifier.startsWith('.') || relative(HERE, target).startsWith('..')) problems.push(`${at(statement)} imports ${specifier}, outside the engine`);
				if (!specifier.endsWith('.ts')) problems.push(`${at(statement)} imports ${specifier} without its extension`);
			}
		}
		(function walk(node: ts.Node): void {
			if (ts.isAsExpression(node) && ts.isAsExpression(node.expression) && node.expression.type.kind === ts.SyntaxKind.UnknownKeyword) {
				problems.push(`${at(node)} casts through unknown`);
			}
			if (ts.isClassExpression(node)) problems.push(`${at(node)} holds a class expression`);
			ts.forEachChild(node, walk);
		})(source);
	}
	assert.deepEqual(problems, []);
});
