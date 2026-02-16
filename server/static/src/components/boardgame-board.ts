import { LitElement, html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';

class BoardgameBoard extends LitElement {
  static override styles = css`
    .cell {
      height: 50px;
      width: 50px;
      position: relative;
      cursor: pointer;
    }

    .cell.even {
      background-color: black;
    }

    .container {
      position: relative;
    }

    ::slotted([boardgame-component]) {
      position: absolute;
      top: 0;
      left: 0;
    }

    .row {
      display: flex;
      flex-direction: row;
    }
  `;

  @property({ type: Number })
  rows = 0;

  @property({ type: Number })
  cols = 0;

  get _cellItems(): string[][] {
    return this._computeCellItems(this.rows, this.cols);
  }

  private _computeCellItems(rows: number, cols: number): string[][] {
    let isOdd = false;
    const result: string[][] = [];
    for (let r = 0; r < rows; r++) {
      const row: string[] = [];
      for (let c = 0; c < cols; c++) {
        row.push(isOdd ? 'odd' : 'even');
        isOdd = !isOdd;
      }
      // each row should alternate which way it starts
      isOdd = !isOdd;
      result.push(row);
    }
    return result;
  }

  private _regionTapped(e: Event) {
    const target = e.target as HTMLElement;
    const r = target.getAttribute('r');
    const c = target.getAttribute('c');
    this.dispatchEvent(new CustomEvent('region-tapped', {
      composed: true,
      bubbles: true,
      detail: { index: [parseInt(r || '0'), parseInt(c || '0')] }
    }));
  }

  override render() {
    return html`
      <div class="container">
        ${repeat(this._cellItems, (row, r) => html`
          <div class="row">
            ${repeat(row, (col, c) => html`
              <div
                class="${col} cell"
                r="${r}"
                c="${c}"
                @click="${this._regionTapped}"
              ></div>
            `)}
          </div>
        `)}
        <slot></slot>
      </div>
    `;
  }
}

customElements.define('boardgame-board', BoardgameBoard);
