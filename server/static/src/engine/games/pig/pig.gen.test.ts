// The generated pig (games/pig/, `boardgame-util transpile ./examples/pig`)
// against everything Go recorded for pig: its 3 conformance documents and
// its 8 fuzz documents, including their rejections. See
// ../generated.testkit.test.ts for what is compared.
import { generatedGameSuite } from '../generated.testkit.test.ts';
import { pigGenerated } from './index.ts';

generatedGameSuite('pig', pigGenerated, { corpus: 3, fuzz: 8 });
