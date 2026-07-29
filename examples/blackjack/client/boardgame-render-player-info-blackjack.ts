import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';
import type { PlayerState } from './_types.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoBlackjack extends PlayerInfoRenderer {
  private _calculateStatus(playerState: PlayerState | null): string {
    if (playerState?.Eliminated) {
      return 'Busted';
    }
    if (playerState?.Stood) {
      return 'Stood';
    }
    // No non-breaking space any more: boardgame-stat reserves a line of its
    // own, so an empty status cannot make the tile jump when a player busts.
    return '';
  }

  override render() {
    return html`
      <div><boardgame-stat label="Score" .value=${this.playerState?.Computed?.HandValue}></boardgame-stat></div>
      <div><boardgame-stat .value=${this._calculateStatus(this.playerState)}></boardgame-stat></div>
    `;
  }
}
