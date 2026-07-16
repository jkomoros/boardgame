import { LitElement, css, html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { isBoundMoveAction, type BoundMoveAction } from '../moves/action.js';

export class BoardgameActionButton extends LitElement {
  static override styles = css`
    :host {
      display: inline-block;
    }

    button {
      width: var(--boardgame-action-width, auto);
      min-width: 44px;
      min-height: 44px;
      padding: 0.65rem 1rem;
      border: 0;
      border-radius: 999px;
      background: var(--boardgame-action-background, var(--md-sys-color-primary, #2e6b4f));
      color: var(--boardgame-action-color, var(--md-sys-color-on-primary, white));
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

    #spinner {
      display: inline-block;
      width: 0.9em;
      height: 0.9em;
      margin-inline-end: 0.45em;
      border: 2px solid currentColor;
      border-inline-end-color: transparent;
      border-radius: 50%;
      vertical-align: -0.1em;
      animation: boardgame-action-spin 700ms linear infinite;
    }

    #spinner[hidden] {
      display: none;
    }

    @keyframes boardgame-action-spin {
      to { transform: rotate(360deg); }
    }

    @media (prefers-reduced-motion: reduce) {
      #spinner { animation-duration: 1.4s; }
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

  /** Required for icon-only controls; ordinary text content names itself. */
  @property({ type: String })
  label = '';

  /** Explanation shown while a higher-level workflow has not produced an action yet. */
  @property({ type: String, attribute: 'unbound-reason' })
  unboundReason = 'No move action supplied';

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
    if (action !== null && !isBoundMoveAction(action)) {
      throw new Error('boardgame-action-button: .action must be a bound move action from move(...), move(...).with(...), or a target candidate');
    }
    if (!this.label.trim() && !this.textContent?.trim()) {
      throw new Error('boardgame-action-button: provide visible text or a non-empty label for an accessible name');
    }
    if (action === null && !this.unboundReason.trim()) {
      throw new Error('boardgame-action-button: unboundReason must explain why no action is available');
    }
    const bound = action !== null;
    const pending = bound && action.submission.kind === 'pending';
    const disabled = !bound || !action.canActivate;
    const rejection = bound && action.submission.kind === 'rejected' ? action.submission.reason : null;
    const baseReason = bound
      ? action.reason?.message ?? rejection
      : this.unboundReason.trim();
    const reason = bound && action.preview.kind === 'failed' && action.preview.retryable
      ? `${baseReason ?? 'Move legality check failed'}. Activate to retry.`
      : baseReason;
    return html`
      <button
        part="button"
        type="button"
        ?disabled=${disabled}
        aria-disabled=${String(disabled)}
        aria-busy=${String(pending)}
        aria-label=${this.label.trim() || nothing}
        aria-describedby=${reason ? 'status' : nothing}
        title=${reason ?? nothing}
        @click=${this.#activate}>
        <span id="spinner" part="spinner" ?hidden=${!pending} aria-hidden="true"></span>
        <span part="label"><slot @slotchange=${this.#slotChanged}></slot></span>
      </button>
      ${reason ? html`<span id="status" part="status" role="status" aria-live="polite">${reason}</span>` : nothing}
    `;
  }

  readonly #slotChanged = (): void => {
    // Dynamic slot changes can add or remove the control's accessible name.
    this.requestUpdate();
  };
}

customElements.define('boardgame-action-button', BoardgameActionButton);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-action-button': BoardgameActionButton;
  }
}
