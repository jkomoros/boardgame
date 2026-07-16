import { LitElement, css, html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { isBoundMoveAction, type BoundMoveAction } from '../moves/action.js';

export class BoardgameActionButton extends LitElement {
  static override styles = css`
    :host {
      display: inline-block;
    }

    button {
      min-width: 44px;
      min-height: 44px;
      padding: 0.65rem 1rem;
      border: 0;
      border-radius: 999px;
      background: var(--md-sys-color-primary, #2e6b4f);
      color: var(--md-sys-color-on-primary, white);
      font: inherit;
      font-weight: 600;
      cursor: pointer;
    }

    button:disabled {
      cursor: not-allowed;
      opacity: 0.55;
    }

    button:focus-visible {
      outline: 3px solid var(--md-sys-color-secondary, #8b7432);
      outline-offset: 3px;
    }

    #status {
      display: block;
      max-width: 28rem;
      margin-top: 0.25rem;
      color: var(--md-sys-color-error, #ba1a1a);
      font-size: 0.875rem;
    }

    @media (forced-colors: active) {
      button:focus-visible {
        outline: 3px solid Highlight;
      }
    }
  `;

  @property({ attribute: false })
  action: BoundMoveAction<string, object> | null = null;

  #unsubscribe: (() => void) | null = null;

  override connectedCallback(): void {
    super.connectedCallback();
    this.#subscribe();
  }

  protected override updated(changedProperties: Map<PropertyKey, unknown>): void {
    super.updated(changedProperties);
    if (changedProperties.has('action')) {
      this.#subscribe();
    }
  }

  override disconnectedCallback(): void {
    this.#unsubscribe?.();
    this.#unsubscribe = null;
    super.disconnectedCallback();
  }

  readonly #activate = async (): Promise<void> => {
    if (isBoundMoveAction(this.action)) await this.action.activate();
  };

  #subscribe(): void {
    this.#unsubscribe?.();
    this.#unsubscribe = isBoundMoveAction(this.action)
      ? this.action.subscribe(() => this.requestUpdate())
      : null;
  }

  override render() {
    const action = this.action;
    const bound = isBoundMoveAction(action);
    const pending = bound && action.submission.kind === 'pending';
    const disabled = !bound || !action.canActivate;
    const rejection = bound && action.submission.kind === 'rejected' ? action.submission.reason : null;
    const baseReason = bound
      ? action.reason?.message ?? rejection
      : action ? 'Bind required move input with .with(...)' : 'No move action supplied';
    const reason = bound && action.preview.kind === 'failed' && action.preview.retryable
      ? `${baseReason ?? 'Move legality check failed'}. Activate to retry.`
      : baseReason;
    return html`
      <button
        type="button"
        ?disabled=${disabled}
        aria-disabled=${String(disabled)}
        aria-busy=${String(pending)}
        aria-describedby=${reason ? 'status' : nothing}
        title=${reason ?? nothing}
        @click=${this.#activate}>
        <slot></slot>
      </button>
      ${reason ? html`<span id="status" role="status" aria-live="polite">${reason}</span>` : nothing}
    `;
  }
}

customElements.define('boardgame-action-button', BoardgameActionButton);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-action-button': BoardgameActionButton;
  }
}
