import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import './boardgame-fading-text.js';

export type StatusTextValue = string | number | null | undefined;
export type StatusTextAutoMessage = 'diff' | 'diff-up' | 'fixed' | 'new';
const autoMessages = new Set<StatusTextAutoMessage>(['diff', 'diff-up', 'fixed', 'new']);

@customElement('boardgame-status-text')
export class BoardgameStatusText extends LitElement {
  static override styles = css`
    :host {
      position: relative;
      display: inline-block;
    }

  `;

  /** Current display value. Use a computed string when custom formatting is needed. */
  @property({ attribute: false })
  value: StatusTextValue = null;

  @property({ type: String })
  autoMessage: StatusTextAutoMessage = 'diff-up';

  /** Announce value changes politely by default. */
  @property({ type: Boolean })
  announce = true;

  override render() {
    this.#validateAuthoring();
    const displayValue = this.value ?? '';
    return html`
      <strong
        role=${this.announce ? 'status' : nothing}
        aria-live=${this.announce ? 'polite' : nothing}
        aria-atomic=${this.announce ? 'true' : nothing}>${displayValue}</strong>
      <boardgame-fading-text
        aria-hidden="true"
        .announce=${false}
        .trigger=${displayValue}
        .autoMessage=${this.autoMessage} 
        suppress="falsey">
      </boardgame-fading-text>
      <slot hidden @slotchange=${this.#validateAuthoring}></slot>
    `;
  }

  readonly #validateAuthoring = (): void => {
    if (this.hasAttribute('value') || this.hasAttribute('message')) {
      throw new Error('boardgame-status-text: bind the typed property with .value=${...}; value/message attributes are not supported');
    }
    const legacyMessage = (this as unknown as { message?: unknown }).message;
    if (legacyMessage !== undefined) {
      throw new Error('boardgame-status-text: .message is not supported; bind .value=${...}');
    }
    if ([...this.childNodes].some(node => node.nodeType === Node.ELEMENT_NODE || node.textContent?.trim())) {
      throw new Error('boardgame-status-text: slotted content is not supported; bind .value=${...}');
    }
    if (this.value !== null && this.value !== undefined
      && typeof this.value !== 'string' && typeof this.value !== 'number') {
      throw new Error('boardgame-status-text: .value must be a string, number, null, or undefined');
    }
    if (!autoMessages.has(this.autoMessage)) {
      throw new Error(`boardgame-status-text: unknown autoMessage "${this.autoMessage}"`);
    }
  };
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-status-text': BoardgameStatusText;
  }
}
