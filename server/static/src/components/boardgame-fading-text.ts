import { LitElement, html, css } from 'lit';
import { property, query } from 'lit/decorators.js';

class BoardgameFadingText extends LitElement {
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
  trigger: any = null;

  @property({ type: String })
  suppress = 'none';

  @property({ type: String })
  autoMessage = 'fixed';

  @property({ type: Boolean, attribute: false })
  protected _animating = false;

  @query('#message')
  private _messageElement?: HTMLElement;

  private _boundAnimationEnded?: () => void;
  private _previousTriggerValue?: any;

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);

    this._boundAnimationEnded = () => this._animationEnded();
    if (this._messageElement) {
      this._messageElement.addEventListener('animationend', this._boundAnimationEnded);
    }
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this._boundAnimationEnded && this._messageElement) {
      this._messageElement.removeEventListener('animationend', this._boundAnimationEnded);
    }
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('trigger')) {
      this._triggerChanged(this.trigger, this._previousTriggerValue);
      this._previousTriggerValue = this.trigger;
    }
  }

  private _animationEnded() {
    this._animating = false;
  }

  animateFade(): void {
    this._animating = true;
  }

  private _triggerChanged(newValue: any, oldValue: any) {
    if (oldValue === undefined) return;

    // If people use us directly newValue and oldValue might be a number...
    // but for example boardgame-status-text will pass us strings.
    const newValueAsNumber = parseInt(newValue);
    const oldValueAsNumber = parseInt(oldValue);

    switch (this.autoMessage) {
      case 'diff':
      case 'diff-up':
        if (!isNaN(newValueAsNumber) && !isNaN(oldValueAsNumber)) {
          const diff = newValueAsNumber - oldValueAsNumber;
          if (this.autoMessage === 'diff-up' && diff < 0) {
            // Skip animating
            return;
          }
          this.message = (diff > 0) ? '+' + diff : String(diff);
        } else {
          this.message = newValue;
        }
        break;
      case 'new':
        this.message = newValue;
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

  private _classes(_animating: boolean): string {
    const classes = [];
    if (_animating) {
      classes.push('animating');
    }
    return classes.join(' ');
  }

  override render() {
    return html`
      <div id="container" class="${this._classes(this._animating)}">
        <div id="message">
          ${this.message}
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-fading-text', BoardgameFadingText);
