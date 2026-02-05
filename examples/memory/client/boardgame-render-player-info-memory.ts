import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import '../../../server/static/src/components/boardgame-status-text.js';

class BoardgameRenderPlayerInfoMemory extends LitElement {
  @property({ type: Object })
  state: any = null;

  @property({ type: Number })
  playerIndex = 0;

  @property({ type: Object })
  playerState: any = null;

  override render() {
    return html`
      Won Cards <boardgame-status-text>${this.playerState?.WonCards?.Indexes?.length}</boardgame-status-text>
    `;
  }
}

customElements.define('boardgame-render-player-info-memory', BoardgameRenderPlayerInfoMemory);
