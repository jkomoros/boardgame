// The generated blackjack (games/blackjack/, `boardgame-util transpile
// ./examples/blackjack`) against everything Go recorded for blackjack: its 7
// conformance documents and its 8 fuzz documents, including their
// rejections. See ../generated.testkit.test.ts for what is compared.
import { generatedGameSuite } from '../generated.testkit.test.ts';
import { blackjack } from './index.ts';

generatedGameSuite('blackjack', blackjack, { corpus: 7, fuzz: 8 });
