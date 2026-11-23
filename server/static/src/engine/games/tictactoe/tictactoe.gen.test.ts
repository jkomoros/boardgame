// The generated tictactoe (games/tictactoe/, `boardgame-util transpile
// ./examples/tictactoe`) against everything Go recorded for tictactoe: its 4
// conformance documents and its 8 fuzz documents, including their
// rejections. See ../generated.testkit.test.ts for what is compared.
import { generatedGameSuite } from '../generated.testkit.test.ts';
import { tictactoe } from './index.ts';

generatedGameSuite('tictactoe', tictactoe, { corpus: 4, fuzz: 8 });
