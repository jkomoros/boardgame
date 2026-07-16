import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import '../../src/components/boardgame-status-text.js';
import type { PlayerState, State } from './_types.js';

class BoardgameRenderPlayerInfoMemory extends LitElement {
  @property({ type: Object })
  state: State | null = null;

  @property({ type: Number })
  playerIndex = 0;

  @property({ type: Object })
  playerState: PlayerState | null = null;

  override render() {
    return html`
      Won Cards <boardgame-status-text>${this.playerState?.WonCards?.Indexes?.length}</boardgame-status-text>
    `;
  }
}

customElements.define('boardgame-render-player-info-memory', BoardgameRenderPlayerInfoMemory);
