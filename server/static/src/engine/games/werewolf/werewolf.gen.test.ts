// The generated werewolf (games/werewolf/, `boardgame-util transpile
// ./examples/werewolf`) against everything Go recorded for werewolf: its 1
// conformance document and its 8 fuzz documents, including their
// rejections. See ../generated.testkit.test.ts for what is compared.
import { generatedGameSuite } from '../generated.testkit.test.ts';
import { werewolf } from './index.ts';

generatedGameSuite('werewolf', werewolf, { corpus: 1, fuzz: 8 });
