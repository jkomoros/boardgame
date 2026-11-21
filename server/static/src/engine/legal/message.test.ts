// The engine's renderLegalMessage over a table of edge cases, with the text
// each must give. These answers were first established by comparing this
// renderer with the client's former copy (src/legal/renderLegalMessage.ts),
// case by case; that copy is gone and the expected text is written out.
import { test } from 'node:test';
import assert from 'node:assert/strict';

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

// One expected text per case above, in order.
const EXPECTED = [
  '',
  '',
  'ana has 3 (false) missing',
  'who has n (flag) missing',
  'nope 1',
  '',
  'k',
  'no placeholders',
  'xx true',
  'toString constructor',
];

test('the engine renders each edge case to its established text', () => {
  assert.equal(EXPECTED.length, CASES.length);
  CASES.forEach((c, index) => {
    assert.equal(renderLegalMessage(c.message, c.table), EXPECTED[index], c.name);
  });
});
