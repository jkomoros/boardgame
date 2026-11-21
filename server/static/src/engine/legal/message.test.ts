// Proves the engine's renderLegalMessage and the client's
// (src/legal/renderLegalMessage.ts) give the same text. The engine carries
// its own copy because the client's module cannot load in a Common Fabric
// pattern; see message.ts.
import { test } from 'node:test';
import assert from 'node:assert/strict';

import { renderLegalMessage as clientRender } from '../../legal/renderLegalMessage.ts';
import { renderLegalMessage } from './message.ts';

const TABLE = { k: '{who} has {n} ({flag}) {missing}', empty: '', plain: 'no placeholders', twice: '{a}{a} {b}' };

const CASES: { name: string; message: Parameters<typeof renderLegalMessage>[0]; table: Record<string, string> | undefined }[] = [
  { name: 'null', message: null, table: TABLE },
  { name: 'undefined', message: undefined, table: TABLE },
  { name: 'all kinds of binding', message: { template: 'k', bindings: { who: 'ana', n: 3, flag: false } }, table: TABLE },
  { name: 'no bindings', message: { template: 'k' }, table: TABLE },
  { name: 'missing key renders the key', message: { template: 'nope {x}', bindings: { x: 1 } }, table: TABLE },
  { name: 'present but empty body', message: { template: 'empty' }, table: TABLE },
  { name: 'no table', message: { template: 'k', bindings: { who: 'bo' } }, table: undefined },
  { name: 'no placeholders', message: { template: 'plain', bindings: { a: 1 } }, table: TABLE },
  { name: 'repeated placeholder', message: { template: 'twice', bindings: { a: 'x', b: true } }, table: TABLE },
  { name: 'own-property names', message: { template: '{toString} {constructor}' }, table: {} },
];

test('the engine renders a legal message exactly as the client does', () => {
  for (const c of CASES) {
    assert.equal(renderLegalMessage(c.message, c.table), clientRender(c.message, c.table), c.name);
  }
  assert.equal(renderLegalMessage(CASES[2].message, TABLE), 'ana has 3 (false) missing');
});
