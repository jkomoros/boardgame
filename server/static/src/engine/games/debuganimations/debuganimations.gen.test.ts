// The generated debuganimations (games/debuganimations/, `boardgame-util
// transpile ./examples/debuganimations`) against everything Go recorded for
// debuganimations: its 1 conformance document and its 8 fuzz documents,
// including their rejections. See ../generated.testkit.test.ts for what is
// compared.
import { generatedGameSuite } from '../generated.testkit.test.ts';
import { debuganimationsGenerated } from './index.ts';

generatedGameSuite('debuganimations', debuganimationsGenerated, { corpus: 1, fuzz: 8 });
