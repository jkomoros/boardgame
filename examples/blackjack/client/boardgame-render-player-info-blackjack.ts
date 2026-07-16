import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import '../../src/components/boardgame-status-text.js';
import type { PlayerState, State } from './_types.js';

declare module './_types.js' {
  interface PlayerComputed {
    readonly HandValue?: number;
  }
}

class BoardgameRenderPlayerInfoBlackjack extends LitElement {
  @property({ type: Object })
  state: State | null = null;

  @property({ type: Number })
  playerIndex = 0;

  @property({ type: Object })
  playerState: PlayerState | null = null;

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
      <div><boardgame-status-text>${this._calculateStatus(this.playerState)}</boardgame-status-text></div>
    `;
  }
}

customElements.define('boardgame-render-player-info-blackjack', BoardgameRenderPlayerInfoBlackjack);
