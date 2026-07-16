import { LitElement, html } from 'lit';
import { TimerController, type TimerReference } from './timer-service.js';

class CustomTimerView extends LitElement {
  timer: TimerReference | null = null;
  readonly clock = new TimerController(this, () => this.timer, { cadence: 'second' });

  override render() {
    return html`${this.clock.reading.secondsLeft}`;
  }
}

const view = new CustomTimerView();
view.timer = { ID: 'turn', IsTimer: true };

// @ts-expect-error readings are service-owned immutable snapshots
view.clock.reading = { timerId: '', status: 'idle', timeLeftMs: 0, durationMs: 0, secondsLeft: 0, progress: 0 };
// @ts-expect-error cadence is a closed performance policy
new TimerController(view, () => view.timer, { cadence: 'minute' });
