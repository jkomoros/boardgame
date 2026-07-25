import { html, css, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';

export type FadingTextTrigger = string | number | boolean | null | undefined;
export type FadingTextAutoMessage = 'diff' | 'diff-up' | 'fixed' | 'new';
export type FadingTextSuppress = 'none' | 'falsey' | 'truthy';

const autoMessages = new Set<FadingTextAutoMessage>(['diff', 'diff-up', 'fixed', 'new']);
const suppressPolicies = new Set<FadingTextSuppress>(['none', 'falsey', 'truthy']);

export class BoardgameFadingText extends BoardgameAnimatableItem {
  static override styles = css`
    #container {
      position: absolute;
      top: 0;
      left: 0;
      height: 100%;
      width: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      visibility: hidden;
      pointer-events: none;
    }

    #container.animating {
      visibility: visible;
    }

    #message {
      font-size: var(--message-font-size, 16px);
    }
  `;

  @property({ type: String })
  message = 'Point Scored';

  @property({ type: Object })
  trigger: FadingTextTrigger = null;

  @property({ type: String })
  suppress: FadingTextSuppress = 'none';

  @property({ type: String })
  autoMessage: FadingTextAutoMessage = 'fixed';

  /** Announce direct callouts; wrappers such as status-text can disable this. */
  @property({ type: Boolean })
  announce = true;

  @property({ type: Boolean, attribute: false })
  protected _visible = false;

  private _previousTriggerValue: FadingTextTrigger;

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    this._validateConfiguration();

    if (changedProperties.has('trigger')) {
      this._triggerChanged(this.trigger, this._previousTriggerValue);
      this._previousTriggerValue = this.trigger;
    }
  }

  animateFade(): void {
    this.finishAllAnimations();          // retrigger = finish prior fade (parity
    this._visible = true;                // with the old generation-counter reset)
    void this.updateComplete.then(() => {
      const message = this.renderRoot.querySelector('#message') as HTMLElement | null;
      if (!message || !this.isConnected) { this._visible = false; return; }
      const anim = this.play(message, [
        { opacity: 1, transform: 'scale(1.0)' },
        { opacity: 0, transform: 'scale(6.0)' },
      ], { easing: 'ease-out' });        // duration defaults to animationLengthMs()
      if (!anim) { this._visible = false; return; }
      anim.finished.catch(() => {}).finally(() => { this._visible = false; });
    });
  }

  private _triggerChanged(newValue: FadingTextTrigger, oldValue: FadingTextTrigger) {
    if (oldValue === undefined) return;

    switch (this.autoMessage) {
      case 'diff':
      case 'diff-up': {
        const newValueAsNumber = this._finiteNumber(newValue);
        const oldValueAsNumber = this._finiteNumber(oldValue);
        if (newValueAsNumber !== null && oldValueAsNumber !== null) {
          const diff = newValueAsNumber - oldValueAsNumber;
          if (this.autoMessage === 'diff-up' && diff < 0) {
            return;
          }
          this.message = (diff > 0) ? '+' + diff : String(diff);
        } else {
          this.message = String(newValue ?? '');
        }
        break;
      }
      case 'new':
        this.message = String(newValue ?? '');
        break;
    }

    switch(this.suppress) {
      case 'falsey':
        if (!newValue) return;
        break;
      case 'truthy':
        if (newValue) return;
        break;
    }

    this.animateFade();
  }

  private _finiteNumber(value: FadingTextTrigger): number | null {
    if (typeof value === 'number') return Number.isFinite(value) ? value : null;
    if (typeof value !== 'string' || value.trim() === '') return null;
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  }

  private _validateConfiguration(): void {
    if (typeof this.message !== 'string') {
      throw new Error('boardgame-fading-text: .message must be a string');
    }
    if (typeof this.trigger === 'number' && !Number.isFinite(this.trigger)) {
      throw new Error('boardgame-fading-text: .trigger numbers must be finite');
    }
    if (this.trigger !== null && this.trigger !== undefined
      && !['string', 'number', 'boolean'].includes(typeof this.trigger)) {
      throw new Error('boardgame-fading-text: .trigger must be a string, number, boolean, null, or undefined');
    }
    if (!autoMessages.has(this.autoMessage)) {
      throw new Error(`boardgame-fading-text: unknown autoMessage "${this.autoMessage}"`);
    }
    if (!suppressPolicies.has(this.suppress)) {
      throw new Error(`boardgame-fading-text: unknown suppress policy "${this.suppress}"`);
    }
  }

  override render() {
    this._validateConfiguration();
    return html`
      <div
        id="container"
        class="${this._visible ? 'animating' : ''}"
        role=${this.announce ? 'status' : nothing}
        aria-live=${this.announce ? 'polite' : nothing}
        aria-atomic=${this.announce ? 'true' : nothing}>
        <div id="message">
          ${this.message}
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-fading-text', BoardgameFadingText);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-fading-text': BoardgameFadingText;
  }
}
