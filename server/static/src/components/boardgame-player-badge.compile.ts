import '../client.js';
import type { PlayerPresentation } from '../client.js';

declare const player: PlayerPresentation;
const badge = document.createElement('boardgame-player-badge');
badge.player = player;
badge.compact = true;

// @ts-expect-error badges require an explicit sanitized presentation
badge.player = 0;
// @ts-expect-error compact mode is boolean
badge.compact = 'true';
