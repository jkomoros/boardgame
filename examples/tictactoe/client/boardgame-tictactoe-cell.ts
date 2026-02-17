import { LitElement, html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { MoveNames } from './_move_names.js';

class BoardgameTictactoeCell extends LitElement {
  static override styles = css`
    :host {
      height: 100px;
      width: 100px;
      cursor: pointer;
      font-family: 'Roboto', 'Noto', sans-serif;
      font-size: 34px;
      font-weight: 400;
      letter-spacing: -.01em;
      line-height: 40px;
    }

    .cell {
      height: 100%;
      width: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
    }

    .cell > div {
      text-align: center;
    }
  `;

  @property({ type: Object })
  token: any = null;

  @property({ type: Number })
  index = 0;

  @property({ type: String })
  value = '';

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('token')) {
      this._tokenChanged(this.token);
    }
  }

  private _tokenChanged(newValue: any) {
    if (!newValue) {
      this.value = '';
      return;
    }
    this.value = newValue.Values.Value;
  }

  override render() {
    return html`
      <div
        class="cell"
        propose-move="${MoveNames.PlaceToken}"
        data-arg-slot="${this.index}"
      >
        ${this.value}
      </div>
    `;
  }
}

customElements.define('boardgame-tictactoe-cell', BoardgameTictactoeCell);
