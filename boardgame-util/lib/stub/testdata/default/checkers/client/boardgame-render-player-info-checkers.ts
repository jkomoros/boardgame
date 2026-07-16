import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import type { PlayerState, State } from './_types.js';

class BoardgameRenderPlayerInfoCheckers extends LitElement {
  @property({ type: Object }) state: State | null = null;
  @property({ type: Number }) playerIndex = 0;
  @property({ type: Object }) playerState: PlayerState | null = null;

  override render() {
    return html`<p>Render player summary information here.</p>`;
  }
}

customElements.define('boardgame-render-player-info-checkers', BoardgameRenderPlayerInfoCheckers);
