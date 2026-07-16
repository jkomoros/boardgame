import '../client.js';
import type { ReadinessParticipant, ReadinessView } from '../client.js';

const readiness = document.createElement('boardgame-readiness');
const participants = [
  { key: 0, label: 'Ada', state: 'ready' },
  { key: 1, label: 'Grace', state: 'waiting' },
] as const satisfies readonly ReadinessParticipant[];
readiness.participants = participants;
readiness.label = 'Votes';
readiness.completeLabel = 'All votes cast';
readiness.emptyLabel = 'Voting is closed';
readiness.progressLabel = 'votes cast';
readiness.readyLabel = 'Voted';
readiness.waitingLabel = 'Thinking';
readiness.notRequiredLabel = 'Eliminated';
readiness.view = 'summary' satisfies ReadinessView;

// @ts-expect-error state is a closed vocabulary
readiness.participants = [{ key: 0, label: 'Ada', state: 'thinking' }];
// @ts-expect-error view is a closed vocabulary
readiness.view = 'compact';
