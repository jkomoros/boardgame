import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import '../../../server/static/src/components/boardgame-status-text.js';

class BoardgameRenderPlayerInfoPig extends LitElement {
  @property({ type: Object })
  state: any = null;

  @property({ type: Number })
  playerIndex = 0;

  @property({ type: Object })
  playerState: any = null;

  override render() {
    return html`
      <div>Round Score <boardgame-status-text>${this.playerState?.RoundScore}</boardgame-status-text></div>
      <div>Total Score <boardgame-status-text>${this.playerState?.TotalScore}</boardgame-status-text></div>
    `;
  }
}

customElements.define('boardgame-render-player-info-pig', BoardgameRenderPlayerInfoPig);
