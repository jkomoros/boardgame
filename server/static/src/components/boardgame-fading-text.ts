import { LitElement, html, css, nothing } from 'lit';
import { property } from 'lit/decorators.js';

export type FadingTextTrigger = string | number | boolean | null | undefined;
export type FadingTextAutoMessage = 'diff' | 'diff-up' | 'fixed' | 'new';
export type FadingTextSuppress = 'none' | 'falsey' | 'truthy';

const autoMessages = new Set<FadingTextAutoMessage>(['diff', 'diff-up', 'fixed', 'new']);
const suppressPolicies = new Set<FadingTextSuppress>(['none', 'falsey', 'truthy']);

export class BoardgameFadingText extends LitElement {
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

    .animating #message {
      animation-name: fadetext;
      animation-duration: var(--animation-length, 0.25s);
      animation-timing-function: ease-out;
    }

    @media (prefers-reduced-motion: reduce) {
      .animating #message {
        animation-duration: 1ms;
      }
    }

    @keyframes fadetext {
      from {
        opacity: 1.0;
        transform: scale(1.0);
      }
      to {
        opacity: 0.0;
        transform: scale(6.0);
      }
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
  protected _animating = false;

  private _previousTriggerValue: FadingTextTrigger;
  private _animationGeneration = 0;

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    this._validateConfiguration();

    if (changedProperties.has('trigger')) {
      this._triggerChanged(this.trigger, this._previousTriggerValue);
      this._previousTriggerValue = this.trigger;
    }
  }

  private _animationEnded() {
    this._animating = false;
  }

  animateFade(): void {
    const generation = ++this._animationGeneration;
    this._animating = false;
    void this.updateComplete.then(() => {
      requestAnimationFrame(() => {
        if (generation === this._animationGeneration && this.isConnected) {
          this._animating = true;
        }
      });
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

  private _classes(_animating: boolean): string {
    const classes: string[] = [];
    if (_animating) {
      classes.push('animating');
    }
    return classes.join(' ');
  }

  override render() {
    this._validateConfiguration();
    return html`
      <div
        id="container"
        class="${this._classes(this._animating)}"
        role=${this.announce ? 'status' : nothing}
        aria-live=${this.announce ? 'polite' : nothing}
        aria-atomic=${this.announce ? 'true' : nothing}>
        <div id="message" @animationend=${this._animationEnded} @animationcancel=${this._animationEnded}>
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
