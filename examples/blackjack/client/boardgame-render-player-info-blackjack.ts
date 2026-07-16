import { html } from '../../src/client.js';
import { PlayerInfoRenderer } from './_game_renderer.js';
import type { PlayerState } from './_types.js';

declare module './_types.js' {
  interface PlayerComputed {
    readonly HandValue?: number;
  }
}

class BoardgameRenderPlayerInfoBlackjack extends PlayerInfoRenderer {
  private _calculateStatus(playerState: PlayerState | null): string {
    if (playerState?.Eliminated) {
      return 'Busted';
    }
    if (playerState?.Stood) {
      return 'Stood';
    }
    // Non breakable space so when the first player busts the layout doesn't jump
    return '\xa0';
  }

  override render() {
    return html`
      <div>Score <strong>${this.playerState?.Computed?.HandValue}</strong></div>
      <div><boardgame-status-text .value=${this._calculateStatus(this.playerState)}></boardgame-status-text></div>
    `;
  }
}

customElements.define('boardgame-render-player-info-blackjack', BoardgameRenderPlayerInfoBlackjack);
