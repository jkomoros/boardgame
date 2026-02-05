import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';

class BoardgameRenderPlayerInfoTictactoe extends LitElement {
  @property({ type: Object })
  state: any = null;

  @property({ type: Number })
  playerIndex = 0;

  get chipText(): string {
    return this._computeChipText(this.state, this.playerIndex);
  }

  get chipColor(): string {
    return this._computeChipColor(this.state, this.playerIndex);
  }

  private _computeChipColor(state: any, playerIndex: number): string {
    if (state?.Players?.[playerIndex]?.TokenValue === 'X') {
      return 'blue';
    }
    return 'red';
  }

  private _computeChipText(state: any, playerIndex: number): string {
    return state?.Players?.[playerIndex]?.TokenValue || '';
  }

  override render() {
    return html``;
  }
}

customElements.define('boardgame-render-player-info-tictactoe', BoardgameRenderPlayerInfoTictactoe);
