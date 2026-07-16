import { LitElement, css, html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import {
  readinessPresentation,
  type ReadinessParticipant,
} from '../status/readiness.js';

export type ReadinessView = 'list' | 'summary';

/** Accessible public progress for simultaneous, already-sanitized participation. */
export class BoardgameReadiness extends LitElement {
  static override styles = css`
    :host { display: block; }
    #surface {
      display: grid;
      gap: var(--boardgame-readiness-gap, 0.75rem);
      padding: var(--boardgame-readiness-padding, 1rem);
      border: var(--boardgame-readiness-border, 1px solid var(--md-sys-color-outline-variant, #cac4d0));
      border-radius: var(--boardgame-readiness-radius, 0.75rem);
      background: var(--boardgame-readiness-background, transparent);
      color: inherit;
    }
    #heading { margin: 0; font: inherit; font-weight: 700; }
    #summary { margin: 0; }
    progress { width: 100%; accent-color: var(--boardgame-readiness-accent, currentColor); }
    ul { display: grid; gap: 0.5rem; margin: 0; padding: 0; list-style: none; }
    li { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
    .state { display: inline-flex; align-items: center; gap: 0.4rem; white-space: nowrap; }
    .state::before { content: ''; width: 0.65rem; height: 0.65rem; border-radius: 50%; border: 1px solid currentColor; }
    [data-state='ready'] .state::before { background: var(--boardgame-readiness-ready, #2e7d32); }
    [data-state='waiting'] .state::before { background: var(--boardgame-readiness-waiting, #ed6c02); }
    [data-state='not-required'] { color: var(--md-sys-color-on-surface-variant, #49454f); }
    [data-state='not-required'] .state::before { background: transparent; }
    @media (forced-colors: active) {
      #surface { border-color: CanvasText; }
      [data-state='ready'] .state::before { background: CanvasText; }
    }
  `;

  /** Already-sanitized public state. Bind as a property, not an attribute. */
  @property({ attribute: false })
  participants: readonly ReadinessParticipant[] = Object.freeze([]);

  @property({ type: String })
  label = 'Readiness';

  @property({ type: String, attribute: 'complete-label' })
  completeLabel = '';

  @property({ type: String, attribute: 'empty-label' })
  emptyLabel = '';

  /** Phrase after the numeric summary, such as “ready” or “votes cast”. */
  @property({ type: String, attribute: 'progress-label' })
  progressLabel = 'ready';

  @property({ type: String, attribute: 'ready-label' })
  readyLabel = 'Ready';

  @property({ type: String, attribute: 'waiting-label' })
  waitingLabel = 'Waiting';

  @property({ type: String, attribute: 'not-required-label' })
  notRequiredLabel = 'Not required';

  @property({ type: String, reflect: true })
  view: ReadinessView = 'list';

  override render() {
    if (!this.label.trim()) throw new Error('boardgame-readiness: label must be a non-empty visible heading');
    if (this.view !== 'list' && this.view !== 'summary') {
      throw new Error(`boardgame-readiness: view must be "list" or "summary", not ${JSON.stringify(this.view)}`);
    }
    const stateLabels = {
      ready: requiredLabel('readyLabel', this.readyLabel),
      waiting: requiredLabel('waitingLabel', this.waitingLabel),
      'not-required': requiredLabel('notRequiredLabel', this.notRequiredLabel),
    } as const;
    const status = readinessPresentation(this.participants, {
      complete: this.completeLabel,
      empty: this.emptyLabel,
      progress: this.progressLabel,
    });
    return html`
      <section id="surface" part="surface" aria-labelledby="heading" data-complete=${String(status.complete)}>
        <h2 id="heading" part="heading">${this.label.trim()}</h2>
        <p id="summary" part="summary" aria-live="polite" aria-atomic="true">${status.message}</p>
        ${status.requiredCount > 0 ? html`
          <progress
            part="progress"
            max=${status.requiredCount}
            value=${status.readyCount}
            aria-label=${status.message}></progress>
        ` : nothing}
        ${this.view === 'list' ? html`
          <ul part="list">
            ${status.participants.map(participant => html`
              <li part="participant participant-${participant.state}" data-state=${participant.state}>
                <span part="participant-label">${participant.label}</span>
                <span class="state" part="participant-state">${stateLabels[participant.state]}</span>
              </li>
            `)}
          </ul>
        ` : nothing}
      </section>
    `;
  }
}

function requiredLabel(name: string, value: string): string {
  if (typeof value !== 'string' || !value.trim()) {
    throw new Error(`boardgame-readiness: ${name} must be a non-empty string`);
  }
  return value.trim();
}

customElements.define('boardgame-readiness', BoardgameReadiness);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-readiness': BoardgameReadiness;
  }
}
