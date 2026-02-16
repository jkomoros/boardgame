import { LitElement, html, css } from 'lit';
import { customElement, property, query } from 'lit/decorators.js';
import './boardgame-fading-text.js';

@customElement('boardgame-status-text')
export class BoardgameStatusText extends LitElement {
  static override styles = css`
    :host {
      position: relative;
      display: inline-block;
    }

    .hidden {
      display: none;
    }
  `;

  // The message to show. If autoMessage is set to something other than
  // 'fixed' this will be set based on trigger.
  @property({ type: String })
  message = '';

  @property({ type: String })
  autoMessage = 'diff-up';

  @query('#content')
  private _contentSlot!: HTMLSlotElement;

  private _observer: MutationObserver | null = null;

  override render() {
    return html`
      <strong>${this.message}</strong>
      <div class="hidden">
        <slot id="content" @slotchange=${this._slotChanged}></slot>
      </div>
      <boardgame-fading-text 
        .trigger=${this.message} 
        .autoMessage=${this.autoMessage} 
        suppress="falsey">
      </boardgame-fading-text>
    `;
  }

  override firstUpdated() {
    this._slotChanged();
  }

  private _textContentChanged(records: MutationRecord[]) {
    const ele = records[records.length - 1].target as HTMLElement;
    let message = ele.textContent || ele.innerText || '';
    message = message.trim();
    this.message = message;
  }

  private _slotChanged() {
    const nodes = this._contentSlot.assignedNodes();
    if (!nodes.length) return;

    for (let i = 0; i < nodes.length; i++) {
      const ele = nodes[i] as HTMLElement;
      let message = ele.textContent || ele.innerText || '';
      // This could happen if it's an empty text node for example.
      message = message.trim();
      // We used to only register these if the message existed, but in many
      // real cases like the BUSTED line in blackjack it starts off as a
      // nil message.
      if (this._observer) {
        this._observer.disconnect();
        this._observer = null;
      }
      this._observer = new MutationObserver(rec => this._textContentChanged(rec));
      this._observer.observe(ele, { characterData: true });
      this.message = message;
      return;
    }
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this._observer) {
      this._observer.disconnect();
      this._observer = null;
    }
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-status-text': BoardgameStatusText;
  }
}
