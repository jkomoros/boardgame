// The source rules every engine module is held to, as functions over one
// file's text, so that more than one checker can apply them:
//
// - sandboxProblems: what the Common Fabric pattern sandbox refuses at
//   module scope (../sandbox.test.ts runs it over every engine source);
// - platformUses: the platform APIs, clock, timers, ambient randomness and
//   asynchrony the rules layer and the games may not use
//   (../rules/purity.test.ts runs it over rules/ and games/).
//
// `boardgame-util port check` (scripts/port-check.mjs) runs both on an
// agent's translation of one function. The TypeScript compiler module is
// passed in, not imported: this file is not a test, and a static import
// would count as a client dependency (src/vite-optimize-deps.test.ts).
type TSModule = typeof import('typescript');
type TSExpression = import('typescript').Expression;
type TSNode = import('typescript').Node;
type TSStringLiteral = import('typescript').StringLiteral;

/** A literal, or arithmetic over literals. */
function literal(ts: TSModule, node: TSExpression): boolean {
	if (ts.isStringLiteralLike(node) || ts.isNumericLiteral(node) || ts.isBigIntLiteral(node)) return true;
	if (ts.isPrefixUnaryExpression(node)) return literal(ts, node.operand);
	if (ts.isParenthesizedExpression(node)) return literal(ts, node.expression);
	if (ts.isBinaryExpression(node)) return literal(ts, node.left) && literal(ts, node.right);
	return false;
}

/** Whether an initializer is data the sandbox can freeze: no functions in it. */
function plainData(ts: TSModule, node: TSExpression): boolean {
	if (ts.isStringLiteralLike(node) || ts.isNumericLiteral(node) || ts.isBigIntLiteral(node) || ts.isIdentifier(node)) return true;
	if (node.kind === ts.SyntaxKind.TrueKeyword || node.kind === ts.SyntaxKind.FalseKeyword || node.kind === ts.SyntaxKind.NullKeyword) return true;
	if (ts.isPrefixUnaryExpression(node)) return plainData(ts, node.operand);
	if (ts.isAsExpression(node) || ts.isParenthesizedExpression(node) || ts.isSatisfiesExpression(node)) return plainData(ts, node.expression);
	// Another constant may be named, alone or inside a record, but not
	// computed with: `DIM * DIM` is refused where `{ size: DIM }` is not.
	if (ts.isBinaryExpression(node)) return literal(ts, node.left) && literal(ts, node.right);
	if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) return true;
	if (ts.isArrayLiteralExpression(node)) return node.elements.every(element => !ts.isSpreadElement(element) && !ts.isOmittedExpression(element) && plainData(ts, element));
	if (ts.isObjectLiteralExpression(node)) {
		return node.properties.every(property => ts.isPropertyAssignment(property) && plainData(ts, property.initializer));
	}
	if (ts.isRegularExpressionLiteral(node)) return !/[gy][a-z]*$/.test(node.text.slice(node.text.lastIndexOf('/')));
	// A call's result is validated when the module loads.
	if (ts.isCallExpression(node)) return true;
	return false;
}

/**
 * Each form the pattern sandbox refuses in one engine source, as
 * "<name>:<line> <what>". `name` labels the problems; `escapesEngine(specifier)`
 * says whether an import specifier leaves the engine directory.
 */
export function sandboxProblems(ts: TSModule, name: string, text: string, escapesEngine: (specifier: string) => boolean): string[] {
	const problems: string[] = [];
	const source = ts.createSourceFile(name, text, ts.ScriptTarget.ES2022, true, ts.ScriptKind.TS);
	const at = (node: TSNode) => `${name}:${source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1}`;
	for (const statement of source.statements) {
		if (ts.isClassDeclaration(statement)) problems.push(`${at(statement)} declares a class at module scope`);
		const declares = ts.isImportDeclaration(statement) || ts.isExportDeclaration(statement) || ts.isVariableStatement(statement) ||
			ts.isFunctionDeclaration(statement) || ts.isInterfaceDeclaration(statement) || ts.isTypeAliasDeclaration(statement) ||
			ts.isClassDeclaration(statement) || ts.isEnumDeclaration(statement);
		if (!declares) problems.push(`${at(statement)} runs a statement at module scope`);
		if (ts.isEnumDeclaration(statement)) problems.push(`${at(statement)} declares an enum at module scope`);
		if (ts.isFunctionDeclaration(statement) && statement.asteriskToken) problems.push(`${at(statement)} declares a generator at module scope`);
		if (ts.isVariableStatement(statement)) {
			if (!(statement.declarationList.flags & ts.NodeFlags.Const)) problems.push(`${at(statement)} declares let or var at module scope`);
			for (const declaration of statement.declarationList.declarations) {
				const init = declaration.initializer;
				if (init === undefined || ts.isArrowFunction(init) || ts.isFunctionExpression(init)) continue;
				if (!plainData(ts, init)) problems.push(`${at(declaration)} holds something other than a function or plain data at module scope`);
			}
		}
		if (ts.isImportDeclaration(statement) || (ts.isExportDeclaration(statement) && statement.moduleSpecifier)) {
			const specifier = (statement.moduleSpecifier as TSStringLiteral).text;
			if (!specifier.startsWith('.') || escapesEngine(specifier)) problems.push(`${at(statement)} imports ${specifier}, outside the engine`);
			if (!specifier.endsWith('.ts')) problems.push(`${at(statement)} imports ${specifier} without its extension`);
		}
	}
	(function walk(node: TSNode): void {
		if (ts.isAsExpression(node) && ts.isAsExpression(node.expression) && node.expression.type.kind === ts.SyntaxKind.UnknownKeyword) {
			problems.push(`${at(node)} casts through unknown`);
		}
		if (ts.isClassExpression(node)) problems.push(`${at(node)} holds a class expression`);
		ts.forEachChild(node, walk);
	})(source);
	return problems;
}

/** Source with comments and string literals blanked, so prose cannot trip a rule. */
export function codeOnly(text: string): string {
	return text
		.replace(/\/\*[\s\S]*?\*\//g, ' ')
		.replace(/\/\/[^\n]*/g, ' ')
		.replace(/'(?:[^'\\\n]|\\.)*'/g, "''")
		.replace(/"(?:[^"\\\n]|\\.)*"/g, '""');
}

/** What the rules layer and the games may not use, by pattern. */
export const FORBIDDEN_PLATFORM: readonly (readonly [RegExp, string])[] = [
	[/\bBigInt\b|\b\d+n\b/, 'BigInt'],
	[/\bcrypto\b/, 'crypto'],
	[/\bsetTimeout\b|\bsetInterval\b|\bsetImmediate\b|\bqueueMicrotask\b|\brequestAnimationFrame\b/, 'a timer'],
	[/\bDate\b|\bperformance\b/, 'a clock'],
	[/Math\.random/, 'Math.random'],
	[/\bdocument\b|\bwindow\b|\bnavigator\b|\blocalStorage\b|\bfetch\b|\bglobalThis\b/, 'a browser global'],
	[/\bprocess\b|\bBuffer\b|\b__dirname\b/, 'a Node global'],
	[/\bTextEncoder\b|\bTextDecoder\b|\bstructuredClone\b|\bAtomics\b|\bSharedArrayBuffer\b|\bWebAssembly\b/, 'a host API'],
	[/\basync\b|\bawait\b|\bPromise\b/, 'asynchrony'],
	[/\beval\b|\bnew Function\b/, 'dynamic code'],
];

/** The first use of each forbidden platform API in `text`, as [what, the text matched]. */
export function platformUses(text: string): [what: string, found: string][] {
	const code = codeOnly(text);
	const out: [string, string][] = [];
	for (const [pattern, what] of FORBIDDEN_PLATFORM) {
		const found = pattern.exec(code);
		if (found !== null) out.push([what, found[0]]);
	}
	return out;
}

/** Whether the text loads a module dynamically (`require(` or `import(`). */
export function loadsDynamically(text: string): boolean {
	return /\brequire\s*\(|\bimport\s*\(/.test(codeOnly(text));
}
