import { LitElement, css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { TimerController, type TimerReference } from '../timers/timer-service.js';

export type TimerDisplayFormat = 'seconds' | 'clock';

/** Accessible, selectively-updating presentation for a generated game timer. */
export class BoardgameTimer extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-width: 0;
    }

    #timer {
      display: grid;
      gap: var(--boardgame-timer-gap, 0.375rem);
    }

    #header {
      display: flex;
      align-items: baseline;
      justify-content: space-between;
      gap: 1rem;
    }

    #label {
      font-weight: 600;
    }

    #value {
      font-variant-numeric: tabular-nums;
    }

    progress {
      width: 100%;
      height: var(--boardgame-timer-progress-height, 0.5rem);
      accent-color: var(--boardgame-timer-color, currentColor);
    }

    .visually-hidden {
      position: absolute;
      width: 1px;
      height: 1px;
      padding: 0;
      margin: -1px;
      overflow: hidden;
      clip: rect(0, 0, 0, 0);
      white-space: nowrap;
      border: 0;
    }
  `;

  @property({ type: Object })
  timer: TimerReference | null = null;

  @property({ type: String })
  label = 'Time remaining';

  @property({ type: String })
  format: TimerDisplayFormat = 'seconds';

  @property({ type: Boolean, attribute: 'hide-progress' })
  hideProgress = false;

  @property({ type: Boolean, attribute: 'hide-value' })
  hideValue = false;

  @property({ type: Boolean, attribute: 'show-when-idle' })
  showWhenIdle = false;

  @property({ type: String, attribute: 'expired-label' })
  expiredLabel = 'Time expired';

  private readonly _clock = new TimerController(this, () => this.timer, {
    getCadence: () => this.hideProgress ? 'second' : 'frame',
  });

  override render() {
    this._validateConfiguration();
    const reading = this._clock.reading;
    if (reading.status === 'idle' && !this.showWhenIdle) return null;
    const value = reading.status === 'running'
      ? this._formatSeconds(reading.secondsLeft)
      : reading.status === 'elapsed'
        ? this.expiredLabel.trim()
        : reading.status === 'idle'
          ? 'Not running'
          : 'Timer unavailable';
    return html`
      <section id="timer" part="timer" data-status=${reading.status} aria-labelledby="label">
        <div id="header" part="header">
          <span id="label" part="label">${this.label.trim()}</span>
          ${!this.hideValue ? html`<output id="value" part="value" role="timer">${value}</output>` : null}
        </div>
        ${!this.hideProgress
          ? html`<progress
              part="progress"
              max="1"
              .value=${reading.progress}
              aria-labelledby="label">
            </progress>`
          : null}
        <span class="visually-hidden" role="status" aria-live="polite" aria-atomic="true">
          ${reading.status === 'elapsed' ? this.expiredLabel.trim() : ''}
        </span>
      </section>
    `;
  }

  private _formatSeconds(seconds: number): string {
    if (this.format === 'seconds') return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    return `${minutes}:${String(seconds % 60).padStart(2, '0')}`;
  }

  private _validateConfiguration(): void {
    if (!this.label.trim()) throw new Error('boardgame-timer: label must be non-empty');
    if (this.format !== 'seconds' && this.format !== 'clock') {
      throw new Error(`boardgame-timer: unknown format ${JSON.stringify(this.format)}`);
    }
    if (!this.expiredLabel.trim()) throw new Error('boardgame-timer: expiredLabel must be non-empty');
  }
}

customElements.define('boardgame-timer', BoardgameTimer);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-timer': BoardgameTimer;
  }
}
