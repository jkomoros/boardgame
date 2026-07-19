import { LitElement, css, html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import {
  defaultMessageResolver,
  type MessageResolver,
  type MoveChoiceProjectionTypes,
  type ProjectedMoveChoices,
} from '../moves/projected-choices.js';
import './boardgame-action-button.js';

/** Safe, always-available fallback for every projected candidate set. */
export class BoardgameProjectedChoices extends LitElement {
  static override styles = css`
    :host {
      display: block;
      position: sticky;
      bottom: 0;
      z-index: 20;
      max-height: min(50vh, 28rem);
      padding-bottom: env(safe-area-inset-bottom, 0);
      overflow: auto;
      overscroll-behavior: contain;
    }
    #failure, fieldset {
      box-sizing: border-box;
      width: min(100% - 2rem, 48rem);
      margin: 0.75rem auto;
      padding: 1rem;
      border: 1px solid var(--md-sys-color-outline-variant, #ccc4b8);
      border-radius: 0.75rem;
      background: var(--md-sys-color-surface-container, #f3edf7);
      color: var(--md-sys-color-on-surface, #1d1b20);
    }
    #failure {
      border-width: 2px;
      border-color: var(--md-sys-color-error, #ba1a1a);
      background: var(--md-sys-color-error-container, #ffdad6);
      color: var(--md-sys-color-on-error-container, #410002);
    }
    legend { padding-inline: 0.35rem; font-weight: 700; }
    .candidates { display: flex; flex-wrap: wrap; gap: 0.75rem; }
    .empty { margin: 0; }
  `;

  @property({ type: Object, attribute: false })
  choices: ProjectedMoveChoices<MoveChoiceProjectionTypes> | null = null;

  @property({ attribute: false })
  messageResolver: MessageResolver = defaultMessageResolver;

  private resolve(message: { readonly id: string; readonly defaultMessage: string }): string {
    try {
      const resolved = this.messageResolver(message);
      return typeof resolved === 'string' && resolved.trim() ? resolved : message.defaultMessage;
    } catch (error) {
      console.error('[projected-choices] message resolver failed:', error);
      return message.defaultMessage;
    }
  }

  override render() {
    const choices = this.choices;
    if (!choices) return nothing;
    if (choices.status === 'failed') {
      return html`<section id="failure" role="alert" aria-live="assertive">
        ${this.resolve(choices.message!)}
      </section>`;
    }
    return html`<aside aria-label="Available game actions" aria-live="polite" aria-atomic="false">
      ${choices.all().map(set => {
      const prompt = this.resolve(set.message);
      return html`<fieldset data-projected-move=${set.move}>
        <legend>${prompt}</legend>
        ${set.candidates.length ? html`
          <div class="candidates">
            ${set.candidates.map(candidate => {
              const label = this.resolve(candidate.message);
              return html`
              <boardgame-action-button
                data-candidate-id=${candidate.id}
                .action=${candidate.action}
                .label=${`${prompt}: ${label}`}>${label}</boardgame-action-button>
            `;})}
          </div>
        ` : html`<p class="empty" role="status">No choices are available.</p>`}
      </fieldset>`;
    })}</aside>`;
  }
}

customElements.define('boardgame-projected-choices', BoardgameProjectedChoices);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-projected-choices': BoardgameProjectedChoices;
  }
}
