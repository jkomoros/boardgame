import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import '../../../server/static/src/components/boardgame-status-text.js';

class BoardgameRenderPlayerInfoBlackjack extends LitElement {
  @property({ type: Object })
  state: any = null;

  @property({ type: Number })
  playerIndex = 0;

  @property({ type: Object })
  playerState: any = null;

  private _calculateStatus(playerState: any): string {
    if (playerState?.Busted) {
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
