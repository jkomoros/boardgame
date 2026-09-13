import { LitElement, css, html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { styleMap } from 'lit/directives/style-map.js';

export interface TrackStep<Key extends string | number = string | number> {
  readonly key: Key;
  readonly label: string;
  readonly color?: string;
}

/** A labelled sequence with one current value; game rules remain with the author. */
export class BoardgameTrack extends LitElement {
  static override styles = css`
    :host { display: block; min-width: 0; }
    #label { margin: 0 0 .4rem; font: inherit; font-weight: 650; }
    #steps {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(min(100%, var(--boardgame-track-step-min-width, 4.75rem)), 1fr));
      overflow: hidden; border: var(--boardgame-track-border, 1px solid currentColor);
      border-radius: var(--boardgame-track-radius, .65rem);
    }
    .step {
      min-width: 0; overflow-wrap: anywhere; text-align: center;
      padding: var(--boardgame-track-step-padding, .55rem .2rem);
      background: var(--step-color, var(--boardgame-track-background, transparent));
      font-size: var(--boardgame-track-font-size, .8rem);
    }
    .step[aria-current] {
      font-weight: 800;
      box-shadow: inset 0 -4px var(--boardgame-track-current-color, currentColor);
    }
    .marker { display: block; height: .75em; font-size: .75em; line-height: 1; }
  `;

  @property({ type: String }) label = '';
  @property({ attribute: false }) steps: readonly TrackStep[] = [];
  @property({ attribute: false }) value: string | number | null = null;

  override render() {
    if (!this.label.trim()) throw new Error('boardgame-track requires a label');
    if (!Array.isArray(this.steps) || this.steps.length > 512) throw new Error('boardgame-track supports up to 512 steps');
    const keys = new Set<string | number>();
    for (const step of this.steps) {
      if (!step || (typeof step.key !== 'string' && typeof step.key !== 'number')
        || (typeof step.key === 'string' ? !step.key.trim() : !Number.isFinite(step.key))
        || keys.has(step.key) || typeof step.label !== 'string' || !step.label.trim()) {
        throw new Error('boardgame-track requires unique keys and labelled steps');
      }
      if (step.color !== undefined && !CSS.supports('color', step.color)) throw new Error('boardgame-track step color must be a CSS color');
      keys.add(step.key);
    }
    return html`<div id="label" part="label">${this.label}</div>
      <div id="steps" part="steps" role="list" aria-labelledby="label">
        ${repeat(this.steps, step => step.key, step => html`<div class="step" part="step"
          role="listitem" aria-current=${step.key === this.value ? 'step' : nothing}
          style=${styleMap({ '--step-color': step.color ?? null })}>
          <span class="marker" aria-hidden="true">${step.key === this.value ? '▼' : nothing}</span>
          ${step.label}</div>`)}
      </div>`;
  }
}
customElements.define('boardgame-track', BoardgameTrack);
declare global { interface HTMLElementTagNameMap { 'boardgame-track': BoardgameTrack; } }
